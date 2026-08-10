package datastartest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Collect starts a test server for the handler, sends a GET request, reads the
// full SSE response, and returns decoded DataStar events.
//
// This is the simplest way to E2E test a synchronous DataStar handler. The
// handler should send all patches and return (closing the stream).
//
// For non-GET requests, custom headers, or request bodies, use httptest.NewServer
// directly with [MustReadEvents]:
//
//	srv := httptest.NewServer(handler)
//	defer srv.Close()
//	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, body)
//	resp, err := http.DefaultClient.Do(req)
//	defer resp.Body.Close()
//	events := datastartest.MustReadEvents(t, resp.Body)
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
