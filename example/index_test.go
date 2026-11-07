package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	datastar "github.com/larsartmann/go-datastar"
)

// TestIndexHandlerCarriesVersionedScriptTag is the smoke test for the
// cache-busting serving pattern the example teaches: the rendered page must
// reference the script exactly as [datastar.ScriptTagVersioned] produces it,
// with the pinned client version in the URL, so the example can never regress
// to a bare datastar.js src (the stale-browser-cache class the versioned tag
// exists to kill).
func TestIndexHandlerCarriesVersionedScriptTag(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/",
		nil,
	)
	indexHandler(rec, request)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}

	want := datastar.ScriptTagVersioned("/datastar.js")
	if !strings.Contains(rec.Body.String(), want) {
		t.Errorf("rendered index missing versioned script tag %q", want)
	}
}
