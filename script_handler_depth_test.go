package datastar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/static"
)

// TestScriptTag_EdgeCases pins the pure-concatenation contract: whatever path
// is passed lands verbatim in the src attribute — no normalization, no
// escaping. Callers own URL correctness.
func TestScriptTag_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty path", "", `<script type="module" src=""></script>`},
		{"path with query", "/js?x=1", `<script type="module" src="/js?x=1"></script>`},
		{"path with fragment", "/js#main", `<script type="module" src="/js#main"></script>`},
		{"absolute URL", "https://cdn.example.com/datastar.js", `<script type="module" src="https://cdn.example.com/datastar.js"></script>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := datastar.ScriptTag(tt.path); got != tt.want {
				t.Errorf("ScriptTag(%q): got %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestScriptHandlerWith_EmptyBundle pins the degenerate-but-valid case: an
// empty bundle still serves with 200, correct headers, and an empty body.
func TestScriptHandlerWith_EmptyBundle(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandlerWith(nil, "0.0.0-empty")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/datastar.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}

	if body := rec.Body.String(); body != "" {
		t.Errorf("body should be empty; got %q", body)
	}

	if cl := rec.Header().Get("Content-Length"); cl != "0" {
		t.Errorf("Content-Length: got %q, want 0", cl)
	}

	if etag := rec.Header().Get("ETag"); etag == "" {
		t.Error("ETag should still be set for the empty bundle")
	}

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("Content-Type: got %q", ct)
	}

	if nosniff := rec.Header().Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Errorf("X-Content-Type-Options: got %q, want nosniff", nosniff)
	}
}

// TestScriptHandler_NoSniffHeader pins the nosniff header on the embedded
// handler (the check that motivated adding it).
func TestScriptHandler_NoSniffHeader(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/datastar.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nosniff := rec.Header().Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Errorf("X-Content-Type-Options: got %q, want nosniff", nosniff)
	}
}
