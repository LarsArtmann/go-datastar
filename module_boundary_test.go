package datastar_test

import (
	"os"
	"testing"

	"golang.org/x/mod/modfile"
)

// TestRootModuleDoesNotRequireDatastartest is a regression guard against the
// circular module dependency that existed when root's go.mod required
// datastartest for a single E2E test file. It parses go.mod semantically
// (require blocks, replace targets, and comments), so a mention inside a
// comment or a replace target can never false-positive, and a real require
// can never slip through a textual edge case. If this test fails, someone
// added datastartest back to go.mod — relocate the test instead.
func TestRootModuleDoesNotRequireDatastartest(t *testing.T) {
	t.Parallel()

	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	file, err := modfile.Parse("go.mod", goMod, nil)
	if err != nil {
		t.Fatalf("parse go.mod: %v", err)
	}

	const datastartestPath = "github.com/larsartmann/go-datastar/datastartest"

	for _, require := range file.Require {
		if require.Mod.Path == datastartestPath {
			t.Errorf("root go.mod requires %q — root must never require datastartest "+
				"(circular dependency). Move the test to datastartest/ instead.",
				datastartestPath)
		}
	}
}
