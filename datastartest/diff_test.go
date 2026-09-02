package datastartest_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

func TestDiff_IdenticalStreams(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("selector #feed", "mode append", "elements <div>one</div>"),
		signalsEvent(`{"count":1}`),
	}

	if diff := datastartest.Diff(events, events); diff != "" {
		t.Fatalf("expected empty diff for identical streams, got:\n%s", diff)
	}
}

func TestDiff_BothEmpty(t *testing.T) {
	t.Parallel()

	if diff := datastartest.Diff(nil, nil); diff != "" {
		t.Fatalf("expected empty diff for two empty streams, got:\n%s", diff)
	}
}

func TestDiff_ShowsChangedLine(t *testing.T) {
	t.Parallel()

	want := []datastartest.Event{elementsEvent("selector #feed", "mode append", "elements <div>one</div>")}
	got := []datastartest.Event{elementsEvent("selector #feed", "mode append", "elements <div>one edited</div>")}

	diff := datastartest.Diff(want, got)

	if !strings.Contains(diff, "-   elements <div>one</div>") {
		t.Fatalf("diff missing want-side line:\n%s", diff)
	}

	if !strings.Contains(diff, "+   elements <div>one edited</div>") {
		t.Fatalf("diff missing got-side line:\n%s", diff)
	}
}

func TestDiff_ShowsMissingAndExtraEvents(t *testing.T) {
	t.Parallel()

	want := []datastartest.Event{
		elementsEvent("selector #a", "mode append", "elements <div>a</div>"),
		elementsEvent("selector #b", "mode append", "elements <div>b</div>"),
	}
	got := []datastartest.Event{elementsEvent("selector #a", "mode append", "elements <div>a</div>")}

	diff := datastartest.Diff(want, got)

	if !strings.Contains(diff, "Event{type=datastar-patch-elements}") ||
		!strings.Contains(diff, "- ") {
		t.Fatalf("diff should mark the missing event with -:\n%s", diff)
	}
}
