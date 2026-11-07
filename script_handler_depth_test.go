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

// TestScriptTagVersioned_EdgeCases pins the cache-busting contract: the
// result is a full <script> tag whose src always carries the pinned client
// version as the v query parameter (?v=, or replacing an existing v value so
// the pinned version can never be shadowed by a stale one), inserted before
// any fragment so it lands in the URL query. Other existing query parameters
// are preserved and join with &. The rest is verbatim, like ScriptTag: no
// normalization, no escaping.
func TestScriptTagVersioned_EdgeCases(t *testing.T) {
	t.Parallel()

	version := datastar.Version()

	const (
		open     = `<script type="module" src="`
		closeTag = `"></script>`
	)

	tests := []struct {
		name string
		path string
		want string
	}{
		{"plain path", "/static/datastar.js", open + "/static/datastar.js?v=" + version + closeTag},
		{"existing query", "/js?x=1", open + "/js?x=1&v=" + version + closeTag},
		{"fragment", "/js#main", open + "/js?v=" + version + "#main" + closeTag},
		{"query and fragment", "/js?x=1#main", open + "/js?x=1&v=" + version + "#main" + closeTag},
		{"empty path", "", open + "?v=" + version + closeTag},
		{
			"absolute URL",
			"https://cdn.example.com/datastar.js",
			open + "https://cdn.example.com/datastar.js?v=" + version + closeTag,
		},
		{
			"stale v is replaced",
			"/js?v=0.0.1",
			open + "/js?v=" + version + closeTag,
		},
		{
			"v replaced among other params",
			"/js?x=1&v=0.0.1&y=2",
			open + "/js?x=1&v=" + version + "&y=2" + closeTag,
		},
		{
			"v replaced before fragment",
			"/js?v=0.0.1#main",
			open + "/js?v=" + version + "#main" + closeTag,
		},
		{
			"empty v value is replaced",
			"/js?v=",
			open + "/js?v=" + version + closeTag,
		},
		{
			"similar key is not matched",
			"/js?verbose=1",
			open + "/js?verbose=1&v=" + version + closeTag,
		},
		{
			"v inside another value is not matched",
			"/js?callback=v=1",
			open + "/js?callback=v=1&v=" + version + closeTag,
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

// TestScriptTagVersioned_ExactlyOneVersionParam pins the shape that makes the
// transform safe in rendering pipelines: however the input path is
// pre-versioned, the rendered tag carries the pinned version exactly once —
// never a double append.
func TestScriptTagVersioned_ExactlyOneVersionParam(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"/datastar.js",
		"/js?x=1",
		"/js?v=0.0.1",
		"/js?x=1&v=0.0.1&y=2",
	} {
		tag := datastar.ScriptTagVersioned(path)

		if got := strings.Count(tag, "v="+datastar.Version()); got != 1 {
			t.Errorf(
				"ScriptTagVersioned(%q): pinned version appears %d times, want 1: %q",
				path,
				got,
				tag,
			)
		}
	}
}

// TestScriptHandlerWith_EmptyBundle pins the degenerate-but-valid case: an
// empty bundle still serves with 200, correct headers, and an empty body.
func TestScriptHandlerWith_EmptyBundle(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandlerWith(nil)

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

// TestScriptHandler_ETagStableIntegration pins the caching contract end to
// end: the embedded handler serves the bundle bytes verbatim, the ETag is
// the exported digest of those bytes, the ETag is stable across requests
// (a per-request digest would defeat conditional-request caching), and a
// matching If-None-Match yields 304 without a body.
func TestScriptHandler_ETagStableIntegration(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()
	wantETag := datastar.ETagFor(static.Bytes())

	get := func() (int, string, string) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/datastar.js",
			nil,
		)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		return rec.Code, rec.Body.String(), rec.Header().Get("ETag")
	}

	firstCode, firstBody, firstETag := get()
	secondCode, _, secondETag := get()

	if firstCode != http.StatusOK || secondCode != http.StatusOK {
		t.Fatalf("status: got %d/%d, want 200/200", firstCode, secondCode)
	}

	if firstETag == "" || firstETag != secondETag {
		t.Fatalf("ETag not stable across requests: %q vs %q", firstETag, secondETag)
	}

	if firstETag != wantETag {
		t.Errorf("served ETag %q != ETagFor(bundle) %q", firstETag, wantETag)
	}

	if firstBody != string(static.Bytes()) {
		t.Errorf(
			"served body differs from the embedded bundle (len %d vs %d)",
			len(firstBody),
			len(static.Bytes()),
		)
	}

	conditional := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/datastar.js",
		nil,
	)
	conditional.Header.Set("If-None-Match", firstETag)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, conditional)

	if rec.Code != http.StatusNotModified {
		t.Errorf("conditional request status: got %d, want 304", rec.Code)
	}
}
