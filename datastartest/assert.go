package datastartest

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

// RequireEventCount fails the test unless events has exactly want events.
func RequireEventCount(t *testing.T, events []Event, want int) {
	t.Helper()

	if len(events) != want {
		t.Fatalf("event count: got %d, want %d", len(events), want)
	}
}

// RequireEventType fails the test unless the event type matches want.
func RequireEventType(t *testing.T, evt Event, want string) {
	t.Helper()

	if evt.Type != want {
		t.Fatalf("event type: got %q, want %q", evt.Type, want)
	}
}

// RequireElements fails the test unless evt is a patch-elements event with the
// exact selector, mode, and HTML content. Use [RequireElementsContains] when
// you need a substring match on the HTML (e.g., for script patches).
func RequireElements(t *testing.T, evt Event, wantSelector, wantMode, wantHTML string) {
	t.Helper()

	if !evt.IsElements() {
		t.Fatalf("expected patch-elements event, got %q", evt.Type)

		return
	}

	if got := evt.Selector(); got != wantSelector {
		t.Errorf("selector: got %q, want %q", got, wantSelector)
	}

	if got := evt.Mode(); got != wantMode {
		t.Errorf("mode: got %q, want %q", got, wantMode)
	}

	if got := evt.Elements(); got != wantHTML {
		t.Errorf("elements: got %q, want %q", got, wantHTML)
	}
}

// RequireElementsContains fails the test unless evt is a patch-elements event
// with the exact selector and mode, and Elements() contains wantHTMLContains
// as a substring. Useful for script patches where the HTML includes wrapper
// elements (e.g., <script>) around the content you want to verify.
func RequireElementsContains(
	t *testing.T,
	evt Event,
	wantSelector, wantMode, wantHTMLContains string,
) {
	t.Helper()

	if !evt.IsElements() {
		t.Fatalf("expected patch-elements event, got %q", evt.Type)

		return
	}

	if got := evt.Selector(); got != wantSelector {
		t.Errorf("selector: got %q, want %q", got, wantSelector)
	}

	if got := evt.Mode(); got != wantMode {
		t.Errorf("mode: got %q, want %q", got, wantMode)
	}

	if got := evt.Elements(); !strings.Contains(got, wantHTMLContains) {
		t.Errorf("elements should contain %q; got %q", wantHTMLContains, got)
	}
}

// RequireSignals fails the test unless evt is a patch-signals event whose
// JSON payload equals wantJSON exactly.
func RequireSignals(t *testing.T, evt Event, wantJSON string) {
	t.Helper()

	if evt.Type != string(datastar.EventTypePatchSignals) {
		t.Fatalf("expected patch-signals event, got %q", evt.Type)

		return
	}

	if got := string(evt.SignalsJSON()); got != wantJSON {
		t.Errorf("signals JSON: got %q, want %q", got, wantJSON)
	}
}

// RequireSignalsContain fails the test unless evt is a patch-signals event
// whose JSON payload contains key at any nesting level. This is a convenience
// for checking individual signal keys without decoding the full payload.
func RequireSignalsContain(t *testing.T, evt Event, key string) {
	t.Helper()

	if evt.Type != string(datastar.EventTypePatchSignals) {
		t.Fatalf("expected patch-signals event, got %q", evt.Type)

		return
	}

	jsonStr := string(evt.SignalsJSON())
	needle := `"` + key + `":`

	if !strings.Contains(jsonStr, needle) {
		t.Errorf("signals JSON should contain key %q; got %s", key, jsonStr)
	}
}
