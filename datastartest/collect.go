package datastartest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Collect starts a test server for the handler, sends a GET request, reads the
// full SSE response, and returns decoded DataStar events.
//
// This is the simplest way to E2E test a synchronous DataStar handler. The
// handler should send all patches and return (closing the stream).
//
// For non-GET requests, custom headers, or request bodies, use
// [CollectWithRequest] or [CollectPost].
//
// For streaming handlers that keep the connection open, use [CollectN].
func Collect(t *testing.T, handler http.Handler) []Event {
	t.Helper()

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET test server: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	return MustReadEvents(t, resp.Body)
}

// CollectWithRequest starts a test server, sends a request with the given
// method, body, and content type, reads the full SSE response, and returns
// decoded DataStar events. Use this for POST/PUT/PATCH handlers that expect
// request bodies.
//
// For the common POST-JSON case, prefer [CollectPost].
func CollectWithRequest(
	t *testing.T,
	handler http.Handler,
	method string,
	body io.Reader,
	contentType string,
) []Event {
	t.Helper()

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL, body)
	if err != nil {
		t.Fatalf("build %s request: %v", method, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s test server: %v", method, err)
	}

	defer func() { _ = resp.Body.Close() }()

	return MustReadEvents(t, resp.Body)
}

// CollectPost is a convenience wrapper around [CollectWithRequest] for POST
// requests with a JSON body — the most common non-GET pattern for DataStar
// handlers (e.g., submitting a form that updates signals).
func CollectPost(t *testing.T, handler http.Handler, jsonBody string) []Event {
	t.Helper()

	return CollectWithRequest(
		t,
		handler,
		http.MethodPost,
		strings.NewReader(jsonBody),
		"application/json",
	)
}

// CollectN starts a test server, sends a GET request, and reads exactly n
// events from the SSE stream before closing the connection. Use this for
// streaming handlers that keep the connection open (e.g., broadcasting through
// a Broadcaster). Unlike [Collect], this does not wait for the handler to
// finish — it returns as soon as n events have been received.
func CollectN(t *testing.T, handler http.Handler, count int) []Event {
	t.Helper()

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET test server: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	events, err := ReadNEvents(resp.Body, count)
	if err != nil {
		t.Fatalf("read %d events: %v", count, err)
	}

	return events
}

// CollectWithTimeout is like [Collect] but enforces a maximum duration. If the
// handler does not close the stream within timeout, the context cancels and
// whatever events were received so far are returned. If no events were received
// before the timeout, the test fails.
//
// Use this for defensive testing against handlers that might hang.
func CollectWithTimeout(t *testing.T, handler http.Handler, timeout time.Duration) []Event {
	t.Helper()

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET test server: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	// A large count so ReadNEvents reads everything; the timeout enforces the deadline.
	const maxEvents = 1 << 30

	events, err := ReadNEvents(resp.Body, maxEvents)
	if err != nil {
		t.Fatalf("read events within %v: %v", timeout, err)
	}

	return events
}

// ReadNEvents reads up to n events from r. Returns as soon as n events have
// been dispatched, without waiting for EOF. This is the streaming-reader
// counterpart to [ReadEvents]: use it with a live SSE connection body that does
// not close on its own (e.g., a handler broadcasting through a Broadcaster).
//
// A scanner error after events have been collected is treated as a clean
// connection close, not a failure.
func ReadNEvents(r io.Reader, count int) ([]Event, error) {
	if count <= 0 {
		return nil, nil
	}

	scanner := newSSEScanner(r)

	var (
		events  []Event
		current Event
		started bool
	)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if started {
				events = append(events, current)
				current = Event{}
				started = false

				if len(events) >= count {
					return events, nil
				}
			}

			continue
		}

		started = true

		applySSELine(&current, line)
	}

	if err := scanner.Err(); err != nil {
		if len(events) > 0 {
			return events, nil
		}

		return nil, errorfamily.WrapTransient(err, CodeSSEScanFailed, "scan SSE stream")
	}

	if started {
		events = append(events, current)
	}

	return events, nil
}
