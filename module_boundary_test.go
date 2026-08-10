package datastar_test

import (
	"os"
	"strings"
	"testing"
)

// TestRootModuleDoesNotRequireDatastartest is a regression guard against the
// circular module dependency that existed when root's go.mod required
// datastartest for a single E2E test file. If this test fails, someone added
// datastartest back to go.mod — relocate the test instead.
func TestRootModuleDoesNotRequireDatastartest(t *testing.T) {
	t.Parallel()

	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	const datastartestPath = "github.com/larsartmann/go-datastar/datastartest"

	for line := range strings.SplitSeq(string(goMod), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") {
			continue
		}

		if strings.Contains(line, datastartestPath) {
			t.Errorf("root go.mod references %q — root must never require datastartest "+
				"(circular dependency). Move the test to datastartest/ instead.",
				datastartestPath)
		}
	}
}
