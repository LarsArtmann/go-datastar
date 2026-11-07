package datastar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
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
		{
			"absolute URL", "https://cdn.example.com/datastar.js",
			`<script type="module" src="https://cdn.example.com/datastar.js"></script>`,
		},
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

// TestScriptTagVersioned_EdgeCases pins the cache-busting contract: the result
// always carries the pinned client version as the v query parameter (?v=, or
// replacing an existing v value so the pinned version can never be shadowed by
// a stale one), inserted before any fragment so it lands in the URL query.
// Other existing query parameters are preserved and join with &. The rest is
// verbatim, like ScriptTag: no normalization, no escaping.
func TestScriptTagVersioned_EdgeCases(t *testing.T) {
	t.Parallel()

	version := datastar.Version()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"plain path", "/static/datastar.js", "/static/datastar.js?v=" + version},
		{"existing query", "/js?x=1", "/js?x=1&v=" + version},
		{"fragment", "/js#main", "/js?v=" + version + "#main"},
		{"query and fragment", "/js?x=1#main", "/js?x=1&v=" + version + "#main"},
		{"empty path", "", "?v=" + version},
		{
			"absolute URL",
			"https://cdn.example.com/datastar.js",
			"https://cdn.example.com/datastar.js?v=" + version,
		},
		{
			"stale v is replaced",
			"/js?v=0.0.1",
			"/js?v=" + version,
		},
		{
			"v replaced among other params",
			"/js?x=1&v=0.0.1&y=2",
			"/js?x=1&v=" + version + "&y=2",
		},
		{
			"v replaced before fragment",
			"/js?v=0.0.1#main",
			"/js?v=" + version + "#main",
		},
		{
			"empty v value is replaced",
			"/js?v=",
			"/js?v=" + version,
		},
		{
			"similar key is not matched",
			"/js?verbose=1",
			"/js?verbose=1&v=" + version,
		},
		{
			"v inside another value is not matched",
			"/js?callback=v=1",
			"/js?callback=v=1&v=" + version,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := datastar.ScriptTagVersioned(tt.path); got != tt.want {
				t.Errorf("ScriptTagVersioned(%q): got %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestScriptTagVersioned_Idempotent pins the idempotency contract: applying
// the cache-busting transform to an already-versioned path changes nothing,
// so rendering pipelines can call it at every layer without double-appending.
func TestScriptTagVersioned_Idempotent(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"/datastar.js",
		"/js?x=1",
		"/js?v=0.0.1",
		"/js?x=1&v=0.0.1&y=2",
		"/js#main",
	} {
		once := datastar.ScriptTagVersioned(path)
		twice := datastar.ScriptTagVersioned(once)

		if once != twice {
			t.Errorf("ScriptTagVersioned not idempotent for %q: once %q, twice %q", path, once, twice)
		}
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
