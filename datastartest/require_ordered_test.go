package datastartest_test

import (
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

func elementsEvent(selector, mode, html string) datastartest.Event {
	return datastartest.Event{
		Type: "datastar-patch-elements",
		DataLines: []string{
			"selector " + selector,
			"mode " + mode,
			"elements " + html,
		},
	}
}

func signalsEvent(json string) datastartest.Event {
	return datastartest.Event{
		Type:      "datastar-patch-signals",
		DataLines: []string{"signals " + json},
	}
}

func TestRequireElementsOrdered(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("#feed", "append", "<div>one</div>"),
		signalsEvent(`{"count":1}`),
		elementsEvent("#feed", "append", "<div>two</div>"),
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
		elementsEvent("#feed", "append", "<div>one</div>"),
	})

	if len(tb.fatals) != 1 {
		t.Fatalf("expected exactly one Fatal, got %v", tb.fatals)
	}
}

func TestRequireElementsOrdered_OrderMismatch(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("#feed", "append", "<div>two</div>"),
		elementsEvent("#feed", "append", "<div>one</div>"),
	},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>two</div>"},
	)

	if len(tb.errors) != 1 {
		t.Fatalf("expected exactly one Error (elements[0] mismatch), got %v", tb.errors)
	}
}

func TestRequireElementsOrdered_ExtraElementsEventFails(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireElementsOrdered(tb, []datastartest.Event{
		elementsEvent("#feed", "append", "<div>one</div>"),
		elementsEvent("#feed", "append", "<div>duplicate</div>"),
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
		elementsEvent("#other", "append", "<div>one</div>"),
	},
		datastartest.ElementExpectation{Selector: "#feed", Mode: "append", HTML: "<div>one</div>"},
	)

	if len(tb.errors) != 1 {
		t.Fatalf("expected exactly one Error (selector mismatch), got %v", tb.errors)
	}
}
