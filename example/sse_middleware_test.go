package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

func TestGzipSSEMiddleware(t *testing.T) {
	t.Parallel()

	handler := gzipSSEMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		if err := resp.PatchElements("<div>compressed hello</div>"); err != nil {
			t.Errorf("PatchElements: %v", err)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q; want gzip", got)
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("body is not gzip: %v", err)
	}

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}

	if !strings.Contains(string(decompressed), "compressed hello") {
		t.Errorf("decompressed stream missing patch; got:\n%s", decompressed)
	}
}

func TestGzipSSEMiddleware_PassthroughWithoutAcceptEncoding(t *testing.T) {
	t.Parallel()

	handler := gzipSSEMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		if err := resp.PatchElements("<div>plain</div>"); err != nil {
			t.Errorf("PatchElements: %v", err)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q; want unset", got)
	}

	if !strings.Contains(rec.Body.String(), "plain") {
		t.Errorf("body should be plain SSE; got:\n%s", rec.Body.String())
	}
}
