package datastartest

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateSnapshots rewrites golden snapshot files when
// -datastartest-update=true is passed to the test binary. The flag is
// namespaced so it cannot collide with flags a consumer's test binary
// registers.
//
//nolint:gochecknoglobals // flag registration is inherently package-level
var updateSnapshots = flag.Bool(
	"datastartest-update",
	false,
	"rewrite datastartest golden snapshot files",
)

// Named file modes keep golden files private.
const (
	snapshotDirMode  = os.FileMode(0o750)
	snapshotFileMode = os.FileMode(0o600)
)

// Snapshot compares events against the golden file testdata/<test name>.golden
// (relative to the CALLING test's package directory) and fails the test on any
// difference, printing a [Diff]-style breakdown.
//
// Golden files are committed test fixtures. To (re)generate them after a
// deliberate behavior change, run the test binary with
// -datastartest-update=true, review the diff, and commit it — a snapshot
// change is a behavior change and belongs in the CHANGELOG.
func Snapshot(tb testing.TB, events []Event) {
	tb.Helper()

	path := filepath.Join("testdata", tb.Name()+".golden")
	want := strings.Join(renderEvents(events), "\n") + "\n"

	if *updateSnapshots {
		if err := os.MkdirAll(filepath.Dir(path), snapshotDirMode); err != nil {
			tb.Fatalf("create snapshot dir: %v", err)
		}

		if err := os.WriteFile(path, []byte(want), snapshotFileMode); err != nil {
			tb.Fatalf("write snapshot %s: %v", path, err)
		}

		return
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read snapshot %s (run with -datastartest-update=true to create it): %v", path, err)

		return
	}

	if string(existing) != want {
		tb.Fatalf("snapshot mismatch in %s\n%s", path,
			diffLines(strings.Split(string(existing), "\n"), strings.Split(want, "\n")))
	}
}
