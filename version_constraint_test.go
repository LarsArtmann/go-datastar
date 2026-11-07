package datastar_test

import (
	"os"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/static"
	"golang.org/x/mod/modfile"
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
			"CHANGELOG.md never mentions the pinned JS client version %s — "+
				"record the bundle bump in release history (or bump static.Version deliberately)",
			pinned,
		)
	}
}

// TestTransportPinsRecordedInChangelog extends the invisible-bump contract to
// the whole transport pin set the three consumer-facing modules require: the
// go-sse family (transport + test helpers) and the lockstep go-datastar
// sibling requires. For every required version, the CHANGELOG must mention
// the pin under the release-history short-name convention ("go-sse v0.7.0",
// "ssetest v0.5.0", "go-datastar v0.6.3", "go-datastar/static v0.6.1") — a
// pin a release ships that release history never mentions is a bump that
// landed invisibly. Repo-internal by design (consumers have no CHANGELOG.md
// or sibling go.mod files), like the static contract above.
func TestTransportPinsRecordedInChangelog(t *testing.T) {
	t.Parallel()

	changelogBytes, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}

	changelog := string(changelogBytes)

	shortNames := map[string]string{
		"github.com/larsartmann/go-sse":             "go-sse",
		"github.com/larsartmann/go-sse/ssetest":     "ssetest",
		"github.com/larsartmann/go-datastar":        "go-datastar",
		"github.com/larsartmann/go-datastar/static": "go-datastar/static",
	}

	recorded := map[string]bool{}

	for _, goModPath := range []string{"go.mod", "broadcast/go.mod", "datastartest/go.mod"} {
		goModBytes, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("read %s: %v", goModPath, err)
		}

		file, err := modfile.Parse(goModPath, goModBytes, nil)
		if err != nil {
			t.Fatalf("parse %s: %v", goModPath, err)
		}

		for _, require := range file.Require {
			shortName, tracked := shortNames[require.Mod.Path]
			if !tracked {
				continue
			}

			record := shortName + " " + require.Mod.Version
			if recorded[record] {
				continue
			}

			recorded[record] = true

			if !strings.Contains(changelog, record) {
				t.Errorf(
					"CHANGELOG.md never mentions the required pin %q (from %s) — "+
						"record the dependency bump in release history",
					record, goModPath,
				)
			}
		}
	}
}
