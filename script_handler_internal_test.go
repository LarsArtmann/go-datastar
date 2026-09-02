package datastar

import (
	"testing"

	"github.com/larsartmann/go-datastar/static"
)

// BenchmarkComputeETag documents the one-time ETag cost: ScriptHandlerWith
// hashes the bundle once at handler construction (already cached in the
// closure), so this cost is init-time, never per-request.
func BenchmarkComputeETag(b *testing.B) {
	data := static.Bytes()

	for b.Loop() {
		_ = computeETag(data)
	}
}
