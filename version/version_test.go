package version

import "testing"

// TestVersionDefault pins the default value: plain `go build`/`go run`
// produce "dev" until -ldflags overrides it (see the package doc for the
// injection path). A consumer binary that injects a version skips this
// package's tests entirely, so the assertion never fights a real override.
func TestVersionDefault(t *testing.T) {
	t.Parallel()

	if got := Version; got != "dev" {
		t.Errorf("default Version = %q, want %q", got, "dev")
	}
}
