package datastar

import (
	"testing"

	"github.com/larsartmann/go-datastar/static"
)

// BenchmarkETagFor documents the one-time ETag cost: ScriptHandlerWith
// hashes the bundle once at handler construction (already cached in the
// closure), so this cost is init-time, never per-request.
func BenchmarkETagFor(b *testing.B) {
	data := static.Bytes()

	for b.Loop() {
		_ = ETagFor(data)
	}
}

// TestAppendVersionQueryIdempotent pins the URL-level idempotency contract:
// versioning an already-versioned URL changes nothing, so rendering pipelines
// can apply the transform at every layer without double-appending.
func TestAppendVersionQueryIdempotent(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"/datastar.js",
		"/js?x=1",
		"/js?v=0.0.1",
		"/js?x=1&v=0.0.1&y=2",
		"/js#main",
		"/js?x=1#main",
	} {
		once := appendVersionQuery(path, "9.9.9")
		twice := appendVersionQuery(once, "9.9.9")

		if once != twice {
			t.Errorf("appendVersionQuery not idempotent for %q: once %q, twice %q", path, once, twice)
		}
	}
}
