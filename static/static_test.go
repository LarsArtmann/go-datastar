package static_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/static"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	if got := static.Version; got == "" {
		t.Fatal("Version is empty; expected a semver string")
	}

	if !strings.Contains(static.Version, ".") {
		t.Errorf("Version %q does not look like semver", static.Version)
	}
}

func TestBytes_NonEmpty(t *testing.T) {
	t.Parallel()

	b := static.Bytes()
	if len(b) == 0 {
		t.Fatal("Bytes() returned an empty slice; the embedded bundle is missing")
	}
}

func TestBytes_HeaderMatchesVersion(t *testing.T) {
	t.Parallel()

	// The DataStar client bundle begins with a version banner such as
	// "// Datastar v1.0.2". This guards against the embedded asset drifting
	// from the declared [static.Version].
	header := string(static.Bytes()[:min(len(static.Bytes()), 64)])

	if !strings.HasPrefix(header, "// Datastar v") {
		t.Fatalf("bundle header %q does not start with the DataStar banner", header)
	}

	if !strings.Contains(header, static.Version) {
		t.Errorf("bundle header %q does not contain declared version %q", header, static.Version)
	}
}

func TestBytes_StableAcrossCalls(t *testing.T) {
	t.Parallel()

	first, second := static.Bytes(), static.Bytes()

	if &first[0] != &second[0] {
		t.Error("Bytes() returned copies rather than the shared embedded value")
	}
}
