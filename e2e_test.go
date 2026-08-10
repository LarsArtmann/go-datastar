package datastar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-sse"
)

// TestE2E_SSEHeaders verifies the transport headers that go-sse sets on SSE
// responses. These are owned by go-sse (NewStream), not go-datastar, but are
// critical for the DataStar client's connection management.
//
// TestE2E_DataStarPatches (the full wire-format E2E test that used datastartest
// helpers) was relocated to datastartest/e2e_test.go to break the circular
// module dependency between root and datastartest.
func TestE2E_SSEHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	wantHeaders := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	for header, want := range wantHeaders {
		if got := resp.Header.Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}
