package datastartest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

// recordingTB captures Fatal/Errorf calls instead of failing the test run.
// It satisfies testing.TB by embedding the interface; methods the Require*
// helpers do not call are never reached. All helpers accept testing.TB, so
// this also proves Ginkgo's GinkgoT() and *testing.B work with them.
type recordingTB struct {
	testing.TB

	fatals []string
	errors []string
}

func (r *recordingTB) Helper() {}

func (r *recordingTB) Fatal(args ...any) {
	r.fatals = append(r.fatals, fmt.Sprint(args...))
}

func (r *recordingTB) Fatalf(format string, args ...any) {
	r.fatals = append(r.fatals, fmt.Sprintf(format, args...))
}

func (r *recordingTB) Error(args ...any) {
	r.errors = append(r.errors, fmt.Sprint(args...))
}

func (r *recordingTB) Errorf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func TestRequireEventCount_Failures(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireEventCount(tb, nil, 2)

	if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], "got 0, want 2") {
		t.Errorf("expected count mismatch fatal; got %v", tb.fatals)
	}
}

func TestRequireEventType_Failure(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireEventType(tb, datastartest.Event{Type: "other"}, "datastar-patch-signals")

	if len(tb.fatals) != 1 {
		t.Errorf("expected type mismatch fatal; got %v", tb.fatals)
	}
}

func elementsEvent(datalines ...string) datastartest.Event {
	return datastartest.Event{
		Type:      "datastar-patch-elements",
		DataLines: datalines,
	}
}

func signalsEvent(json string) datastartest.Event {
	return datastartest.Event{
		Type:      "datastar-patch-signals",
		DataLines: []string{"signals " + json},
	}
}

func TestRequireElements_FailurePaths(t *testing.T) {
	t.Parallel()

	t.Run("not an elements event", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireElements(tb, signalsEvent(`{}`), "#feed", "outer", "<div/>")

		if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], "patch-elements") {
			t.Errorf("expected wrong-type fatal; got %v", tb.fatals)
		}
	})

	t.Run("field mismatches report all three", func(t *testing.T) {
		t.Parallel()

		evt := elementsEvent("selector #feed", "mode inner", "elements <div/>")
		tb := &recordingTB{}
		datastartest.RequireElements(tb, evt, "#other", "outer", "<span/>")

		if len(tb.errors) != 3 {
			t.Errorf("expected 3 field errors (selector, mode, elements); got %v", tb.errors)
		}
	})
}

func TestRequireElementsContains_FailurePaths(t *testing.T) {
	t.Parallel()

	t.Run("not an elements event", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireElementsContains(
			tb,
			datastartest.Event{Type: "message"},
			"#f",
			"outer",
			"x",
		)

		if len(tb.fatals) != 1 {
			t.Errorf("expected wrong-type fatal; got %v", tb.fatals)
		}
	})

	t.Run("substring missing", func(t *testing.T) {
		t.Parallel()

		evt := elementsEvent("selector #feed", "elements <div>hello</div>")
		tb := &recordingTB{}
		datastartest.RequireElementsContains(tb, evt, "#feed", "outer", "goodbye")

		if len(tb.errors) != 1 || !strings.Contains(tb.errors[0], "goodbye") {
			t.Errorf("expected substring error; got %v", tb.errors)
		}
	})
}

func TestRequireSignals_ExactJSON(t *testing.T) {
	t.Parallel()

	evt := signalsEvent(`{"count":1}`)
	datastartest.RequireSignals(t, evt, `{"count":1}`)
}

func TestRequireSignals_FailurePaths(t *testing.T) {
	t.Parallel()

	t.Run("not a signals event", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireSignals(tb, elementsEvent(), `{}`)

		if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], "patch-signals") {
			t.Errorf("expected wrong-type fatal; got %v", tb.fatals)
		}
	})

	t.Run("JSON mismatch", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireSignals(tb, signalsEvent(`{"count":1}`), `{"count":2}`)

		if len(tb.errors) != 1 || !strings.Contains(tb.errors[0], "signals JSON") {
			t.Errorf("expected JSON mismatch error; got %v", tb.errors)
		}
	})
}

func TestRequireSignalsContain_Failure(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.RequireSignalsContain(tb, signalsEvent(`{"count":1}`), "missing")

	if len(tb.errors) != 1 || !strings.Contains(tb.errors[0], `"missing"`) {
		t.Errorf("expected key-missing error; got %v", tb.errors)
	}
}

func scriptEvent(js string) datastartest.Event {
	return elementsEvent(
		"selector body",
		"mode append",
		"elements <script type=\"module\">"+js+"</script>",
	)
}

func TestRequireScript(t *testing.T) {
	t.Parallel()

	datastartest.RequireScript(t, scriptEvent("console.log('hi')"), "console.log('hi')")
}

func TestRequireScript_FailurePaths(t *testing.T) {
	t.Parallel()

	t.Run("not a script event", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireScript(tb, elementsEvent("elements <div/>"), "x")

		if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], "script-bearing") {
			t.Errorf("expected wrong-type fatal; got %v", tb.fatals)
		}
	})

	t.Run("content mismatch", func(t *testing.T) {
		t.Parallel()

		tb := &recordingTB{}
		datastartest.RequireScript(tb, scriptEvent("console.log('a')"), "console.log('b')")

		if len(tb.errors) != 1 || !strings.Contains(tb.errors[0], "script content") {
			t.Errorf("expected content mismatch error; got %v", tb.errors)
		}
	})
}

func TestRequireEventID(t *testing.T) {
	t.Parallel()

	datastartest.RequireEventID(t, datastartest.Event{ID: "42"}, "42")

	tb := &recordingTB{}
	datastartest.RequireEventID(tb, datastartest.Event{ID: "41"}, "42")

	if len(tb.fatals) != 1 || !strings.Contains(tb.fatals[0], `"41"`) {
		t.Errorf("expected ID mismatch fatal; got %v", tb.fatals)
	}
}

func TestMustReadNEvents(t *testing.T) {
	t.Parallel()

	const wire = "event: datastar-patch-elements\n" +
		"data: elements <div>1</div>\n\n" +
		"event: datastar-patch-signals\n" +
		"data: signals {}\n\n"

	events := datastartest.MustReadNEvents(t, strings.NewReader(wire), 2)
	datastartest.RequireEventCount(t, events, 2)
}

func TestMustReadNEvents_FailingReader(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.MustReadNEvents(tb, failingReader{}, 2)

	if len(tb.fatals) != 1 {
		t.Errorf("expected fatal from failing reader; got %v", tb.fatals)
	}
}

func TestMustReadEvents_FailingReader(t *testing.T) {
	t.Parallel()

	tb := &recordingTB{}
	datastartest.MustReadEvents(tb, failingReader{})

	if len(tb.fatals) != 1 {
		t.Errorf("expected fatal from failing reader; got %v", tb.fatals)
	}
}

// TestHelpers_AcceptTestingB proves the helpers work with *testing.B, not just
// *testing.T — the reason every public helper takes testing.TB.
func TestHelpers_AcceptTestingB(t *testing.T) {
	t.Parallel()

	b := &testing.B{}
	datastartest.RequireEventCount(b, []datastartest.Event{{Type: "x"}}, 1)
	datastartest.RequireEventType(b, datastartest.Event{Type: "x"}, "x")
	datastartest.RequireEventID(b, datastartest.Event{ID: "1"}, "1")

	if b.Failed() {
		t.Error("helper calls on *testing.B should not fail")
	}
}
