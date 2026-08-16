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
// Options customize the request: [WithPath] targets a route other than "/",
// [WithDatastarSignals] submits inbound signals via the query parameter, and
// [WithLastEventID] simulates a reconnecting client for replay testing.
//
// For non-GET requests or request bodies, use [CollectWithRequest] or
// [CollectPost]. For streaming handlers that keep the connection open, use
// [CollectN].
func Collect(tb testing.TB, handler http.Handler, opts ...RequestOption) []Event {
	tb.Helper()

	resp := doRequest(tb, handler, http.MethodGet, nil, "", context.Background(), opts)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tb.Errorf("close response body: %v", err)
		}
	}()

	return MustReadEvents(tb, resp.Body)
}

// CollectWithRequest starts a test server, sends a request with the given
// method, body, and content type, reads the full SSE response, and returns
// decoded DataStar events. Use this for POST/PUT/PATCH handlers that expect
// request bodies.
//
// For the common POST-JSON case, prefer [CollectPost].
func CollectWithRequest(
	tb testing.TB,
	handler http.Handler,
	method string,
	body io.Reader,
	contentType string,
	opts ...RequestOption,
) []Event {
	tb.Helper()

	resp := doRequest(tb, handler, method, body, contentType, context.Background(), opts)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tb.Errorf("close response body: %v", err)
		}
	}()

	return MustReadEvents(tb, resp.Body)
}

// CollectPost is a convenience wrapper around [CollectWithRequest] for POST
// requests with a JSON body — the most common non-GET pattern for DataStar
// handlers (e.g., submitting a form that updates signals).
func CollectPost(
	tb testing.TB,
	handler http.Handler,
	jsonBody string,
	opts ...RequestOption,
) []Event {
	tb.Helper()

	return CollectWithRequest(
		tb,
		handler,
		http.MethodPost,
		strings.NewReader(jsonBody),
		"application/json",
		opts...,
	)
}

// CollectN starts a test server, sends a GET request, and reads exactly count
// events from the SSE stream before closing the connection. Use this for
// streaming handlers that keep the connection open (e.g., broadcasting through
// a Broadcaster). Unlike [Collect], this does not wait for the handler to
// finish — it returns as soon as count events have been received.
func CollectN(tb testing.TB, handler http.Handler, count int, opts ...RequestOption) []Event {
	tb.Helper()

	if count <= 0 {
		return nil
	}

	resp := doRequest(tb, handler, http.MethodGet, nil, "", context.Background(), opts)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tb.Errorf("close response body: %v", err)
		}
	}()

	events, err := ReadNEvents(resp.Body, count)
	if err != nil {
		tb.Fatalf("read %d events: %v", count, err)
	}

	return events
}

// CollectWithTimeout is like [Collect] but enforces a maximum duration. If the
// handler does not close the stream within timeout, the context cancels and
// whatever events were received so far are returned. If no events were received
// before the timeout, the test fails.
//
// Use this for defensive testing against handlers that might hang.
func CollectWithTimeout(
	tb testing.TB,
	handler http.Handler,
	timeout time.Duration,
	opts ...RequestOption,
) []Event {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp := doRequest(tb, handler, http.MethodGet, nil, "", ctx, opts)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tb.Errorf("close response body: %v", err)
		}
	}()

	// A large count so ReadNEvents reads everything; the timeout enforces the deadline.
	const maxEvents = 1 << 30

	events, err := ReadNEvents(resp.Body, maxEvents)
	if err != nil {
		tb.Fatalf("read events within %v: %v", timeout, err)
	}

	return events
}

// doRequest starts an httptest server for handler, builds a request from the
// method, body, content type, and options, sends it, and returns the response.
// The server is closed via tb.Cleanup; closing the response body is the
// caller's responsibility.
func doRequest(
	tb testing.TB,
	handler http.Handler,
	method string,
	body io.Reader,
	contentType string,
	ctx context.Context,
	opts []RequestOption,
) *http.Response {
	tb.Helper()

	srv := httptest.NewServer(handler)
	tb.Cleanup(srv.Close)

	cfg := applyRequestOptions(opts)

	target := srv.URL + cfg.targetPath()

	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		tb.Fatalf("build %s request to %s: %v", method, target, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, values := range cfg.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tb.Fatalf("%s test server: %v", method, err)
	}

	return resp
}

// ReadNEvents reads up to count events from r. Returns as soon as count events
// have been dispatched, without waiting for EOF. This is the streaming-reader
// counterpart to [ReadEvents]: use it with a live SSE connection body that does
// not close on its own (e.g., a handler broadcasting through a Broadcaster).
//
// Wire-format semantics are identical to [ReadEvents] (spec § 9.2.6), except
// that reading stops early: a frame still pending when count is reached is
// naturally discarded, as is a frame pending at EOF.
//
// A scanner error after events have been collected is treated as a clean
// connection close, not a failure.
func ReadNEvents(r io.Reader, count int) ([]Event, error) {
	if count <= 0 {
		return nil, nil
	}

	parser := streamParser{}
	scanner := newSSEScanner(r)

	for scanner.Scan() {
		parser.acceptLine(scanner.Text())

		if len(parser.events) >= count {
			return parser.events, nil
		}
	}

	if err := scanner.Err(); err != nil {
		if len(parser.events) > 0 {
			return parser.events, nil
		}

		return nil, errorfamily.WrapTransient(err, CodeSSEScanFailed, "scan SSE stream")
	}

	return parser.events, nil
}
