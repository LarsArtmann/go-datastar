package datastartest

import (
	"strings"
	"testing"
)

// ElementExpectation describes one expected patch-elements event for
// [RequireElementsOrdered].
type ElementExpectation struct {
	Selector string
	Mode     string
	HTML     string
}

// RequireElementsOrdered fails the test unless the patch-elements events in
// events — taken in order — match want exactly: same count, same order, and
// each event's selector, mode, and HTML equal to the expectation's. Events of
// other types (signals, comments, keep-alives) are ignored; additional
// elements events beyond want are a failure, so duplicate patches cannot slip
// through. Pair with [Diff] for a readable failure output.
func RequireElementsOrdered(tb testing.TB, events []Event, want ...ElementExpectation) {
	tb.Helper()

	var got []Event

	for _, evt := range events {
		if evt.IsElements() {
			got = append(got, evt)
		}
	}

	if len(got) != len(want) {
		tb.Fatalf("patch-elements events: got %d, want %d\ngot:\n%s\nwant:\n%s",
			len(got), len(want),
			renderEvents(got), renderExpectations(want))

		return
	}

	for i, expectation := range want {
		requireElementsPatch(tb, got[i], expectation.Selector, expectation.Mode)

		if gotHTML := got[i].Elements(); gotHTML != expectation.HTML {
			tb.Errorf("elements[%d]: got %q, want %q", i, gotHTML, expectation.HTML)
		}
	}
}

// renderExpectations renders expectations in the same shape renderEvents uses
// for events, so count-mismatch failures read as a direct comparison.
func renderExpectations(want []ElementExpectation) string {
	var b strings.Builder

	for i, expectation := range want {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString("Event{type=datastar-patch-elements}\n")
		b.WriteString("  selector " + expectation.Selector + "\n")
		b.WriteString("  mode " + expectation.Mode + "\n")
		b.WriteString("  elements " + expectation.HTML)
	}

	return b.String()
}
