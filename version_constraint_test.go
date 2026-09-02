package datastar_test

import (
	"os"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/static"
)

// TestStaticVersionRecordedInChangelog enforces the release constraint between
// the embedded DataStar JS client and the project's release history: whenever
// the pinned client version changes, the CHANGELOG must record it, so a
// bundle bump can never land invisibly.
func TestStaticVersionRecordedInChangelog(t *testing.T) {
	t.Parallel()

	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}

	pinned := "v" + static.Version
	if !strings.Contains(string(changelog), pinned) {
		t.Errorf(
			"CHANGELOG.md never mentions the pinned JS client version %s — record the bundle bump in release history (or bump static.Version deliberately)",
			pinned,
		)
	}
}
