package datastartest_test

import (
	"flag"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

var snapshotEvents = []datastartest.Event{
	elementsEvent("#feed", "append", "<div>snapshot</div>"),
	signalsEvent(`{"total":1}`),
}

// TestSnapshot exercises the golden round trip: with -datastartest-update the
// snapshot is (re)written; a following plain call passes. The golden file it
// writes (testdata/TestSnapshot.golden) is a committed fixture.
func TestSnapshot(t *testing.T) {
	if err := flag.Set("datastartest-update", "true"); err != nil {
		t.Fatalf("set update flag: %v", err)
	}

	datastartest.Snapshot(t, snapshotEvents)

	if err := flag.Set("datastartest-update", "false"); err != nil {
		t.Fatalf("reset update flag: %v", err)
	}

	datastartest.Snapshot(t, snapshotEvents)
}

// TestSnapshot_Mismatch writes its own golden first (update mode), then
// replays a mutated stream against it through a recordingTB — the genuine
// mismatch path.
func TestSnapshot_Mismatch(t *testing.T) {
	if err := flag.Set("datastartest-update", "true"); err != nil {
		t.Fatalf("set update flag: %v", err)
	}

	datastartest.Snapshot(t, snapshotEvents)

	if err := flag.Set("datastartest-update", "false"); err != nil {
		t.Fatalf("reset update flag: %v", err)
	}

	tb := &recordingTB{}
	datastartest.Snapshot(tb, []datastartest.Event{
		elementsEvent("#feed", "append", "<div>changed</div>"),
	})

	if len(tb.fatals) != 1 {
		t.Fatalf("expected exactly one Fatal on snapshot mismatch, got %v", tb.fatals)
	}
}

// TestSnapshot_MissingGoldenFile runs under a fresh test name whose golden
// does not exist, exercising the creation-guidance path.
func TestSnapshot_MissingGoldenFile(t *testing.T) {
	tb := &recordingTB{}
	datastartest.Snapshot(tb, snapshotEvents)

	if len(tb.fatals) != 1 {
		t.Fatalf("expected exactly one Fatal for the missing golden file, got %v", tb.fatals)
	}
}
