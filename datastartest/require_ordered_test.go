package datastartest_test

import (
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

func TestRequireElementsOrdered(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("selector #feed", "mode append", "elements <div>one</div>"),
		signalsEvent(`{"count":1}`),
		elementsEvent("selector #feed", "mode append", "elements <div>two</div>"),
	}

	datastartest.RequireElementsOrdered(t, events,
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>two</div>"},
	)
}

func TestRequireElementsOrdered_CountMismatch(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("selector #feed", "mode append", "elements <div>one</div>"),
	})

	if len(tb.fatals) != 1 {
		t.Fatalf("expected exactly one Fatal, got %v", tb.fatals)
	}
}

func TestRequireElementsOrdered_OrderMismatch(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("selector #feed", "mode append", "elements <div>two</div>"),
		elementsEvent("selector #feed", "mode append", "elements <div>one</div>"),
	},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>two</div>"},
	)

	if len(tb.errors) != 2 {
		t.Fatalf("expected both order mismatches reported, got %v", tb.errors)
	}
}

func TestRequireElementsOrdered_ExtraElementsEventFails(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("selector #feed", "mode append", "elements <div>one</div>"),
		elementsEvent("selector #feed", "mode append", "elements <div>duplicate</div>"),
	},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
	)

	if len(tb.fatals) != 1 {
		t.Fatalf("expected exactly one Fatal for the extra elements event, got %v", tb.fatals)
	}
}

func TestRequireElementsOrdered_WrongSelectorFails(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("selector #other", "mode append", "elements <div>one</div>"),
	},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
	)

	if len(tb.errors) != 1 {
		t.Fatalf("expected exactly one Error (selector mismatch), got %v", tb.errors)
	}
}
