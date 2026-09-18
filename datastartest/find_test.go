package datastartest_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

func TestFindAllElements(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("selector #a", "elements <div>1</div>"),
		signalsEvent(`{"x":1}`),
		elementsEvent("selector #b", "elements <div>2</div>"),
		elementsEvent("selector #a", "elements <div>3</div>"),
	}

	matches := datastartest.FindAllElements(events, "#a")
	if len(matches) != 2 {
		t.Fatalf("selector #a matches: got %d, want 2", len(matches))
	}

	if got := matches[0].Elements(); got != "<div>1</div>" {
		t.Errorf("first #a match: got %q, want %q", got, "<div>1</div>")
	}

	if got := matches[1].Elements(); got != "<div>3</div>" {
		t.Errorf("second #a match: got %q, want %q", got, "<div>3</div>")
	}

	if got := datastartest.FindAllElements(events, "#missing"); len(got) != 0 {
		t.Errorf("selector #missing matches: got %d, want 0", len(got))
	}
}

func TestFindAllElements_EmptySelector(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("elements <div>anonymous</div>"),
		elementsEvent("selector #a", "elements <div>1</div>"),
	}

	matches := datastartest.FindAllElements(events, "")
	if len(matches) != 1 || matches[0].Elements() != "<div>anonymous</div>" {
		t.Errorf("empty selector should match only selector-less patches; got %v", matches)
	}
}

func TestFindScript(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		signalsEvent(`{"x":1}`),
		scriptEvent("console.log('first')"),
		scriptEvent("console.log('second')"),
	}

	found, ok := datastartest.FindScript(events)
	if !ok {
		t.Fatal("FindScript: not found, want the first script patch")
	}

	if got := found.ScriptContent(); got != "console.log('first')" {
		t.Errorf("script content: got %q, want %q", got, "console.log('first')")
	}
}

func TestFindScript_None(t *testing.T) {
	t.Parallel()

	found, ok := datastartest.FindScript([]datastartest.Event{
		elementsEvent("elements <div>ok</div>"),
		signalsEvent(`{"x":1}`),
	})
	if ok {
		t.Fatal("FindScript: found, want none")
	}

	if found.Type != "" || found.DataLines != nil {
		t.Errorf("FindScript without a match should return the zero Event; got %+v", found)
	}
}

func TestEventToSelectorMap(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("selector #feed", "elements <div>1</div>"),
		signalsEvent(`{"x":1}`),
		elementsEvent("selector #toast", "elements <div>2</div>"),
		elementsEvent("selector #feed", "elements <div>3</div>"),
	}

	bySelector := datastartest.EventToSelectorMap(events)

	if got := bySelector["#toast"].Elements(); got != "<div>2</div>" {
		t.Errorf("#toast: got %q, want %q", got, "<div>2</div>")
	}

	if got := bySelector["#feed"].Elements(); got != "<div>3</div>" {
		t.Errorf("#feed should map to the last patch; got %q, want %q", got, "<div>3</div>")
	}

	if len(bySelector) != 2 {
		t.Errorf("map size: got %d, want 2 (signals events are ignored)", len(bySelector))
	}
}

func TestEventToSelectorMap_EmptySelectorKey(t *testing.T) {
	t.Parallel()

	events := []datastartest.Event{
		elementsEvent("elements <div>anonymous</div>"),
	}

	bySelector := datastartest.EventToSelectorMap(events)

	if got := bySelector[""].Elements(); got != "<div>anonymous</div>" {
		t.Errorf("selector-less patch should be keyed under \"\"; got %q", got)
	}
}

func TestRequireNotScript(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireNotScript(tb, elementsEvent("elements <div>ok</div>"))
	datastartest.RequireNotScript(tb, signalsEvent(`{"x":1}`))

	if len(tb.fatals) != 0 {
		t.Errorf("non-script events should pass; got %v", tb.fatals)
	}
}

func TestRequireNotScript_Failure(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireNotScript(tb, scriptEvent("console.log('nope')"))

	if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], "script patch with content") {
		t.Errorf("expected script-patch fatal; got %v", tb.fatals)
	}
}
