package datastartest

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

// All Require* helpers accept [testing.TB] rather than *testing.T, so they work
// with *testing.T, *testing.B, and Ginkgo's GinkgoT().

// RequireEventCount fails the test unless events has exactly want events.
func RequireEventCount(tb testing.TB, events []Event, want int) {
	tb.Helper()

	if len(events) != want {
		tb.Fatalf("event count: got %d, want %d", len(events), want)
	}
}

// RequireEventType fails the test unless the event type matches want.
func RequireEventType(tb testing.TB, evt Event, want string) {
	tb.Helper()

	if evt.Type != want {
		tb.Fatalf("event type: got %q, want %q", evt.Type, want)
	}
}

// RequireElements fails the test unless evt is a patch-elements event with the
// exact selector, mode, and HTML content. Use [RequireElementsContains] when
// you need a substring match on the HTML (e.g., for script patches).
func RequireElements(tb testing.TB, evt Event, wantSelector, wantMode, wantHTML string) {
	tb.Helper()

	requireElementsPatch(tb, evt, wantSelector, wantMode)

	if got := evt.Elements(); got != wantHTML {
		tb.Errorf("elements: got %q, want %q", got, wantHTML)
	}
}

// requireElementsPatch fails the test unless evt is a patch-elements event
// with the exact selector and mode. The elements comparison is left to the
// caller so exact and substring variants can share it.
func requireElementsPatch(tb testing.TB, evt Event, wantSelector, wantMode string) {
	tb.Helper()

	if !evt.IsElements() {
		tb.Fatalf("expected patch-elements event, got %q", evt.Type)

		return
	}

	if got := evt.Selector(); got != wantSelector {
		tb.Errorf("selector: got %q, want %q", got, wantSelector)
	}

	if got := evt.Mode(); got != wantMode {
		tb.Errorf("mode: got %q, want %q", got, wantMode)
	}
}

// RequireElementsContains fails the test unless evt is a patch-elements event
// with the exact selector and mode, and Elements() contains wantHTMLContains
// as a substring. Useful for script patches where the HTML includes wrapper
// elements (e.g., <script>) around the content you want to verify.
func RequireElementsContains(
	tb testing.TB,
	evt Event,
	wantSelector, wantMode, wantHTMLContains string,
) {
	tb.Helper()

	requireElementsPatch(tb, evt, wantSelector, wantMode)

	if got := evt.Elements(); !strings.Contains(got, wantHTMLContains) {
		tb.Errorf("elements should contain %q; got %q", wantHTMLContains, got)
	}
}

// RequireSignals fails the test unless evt is a patch-signals event whose
// JSON payload equals wantJSON exactly.
func RequireSignals(tb testing.TB, evt Event, wantJSON string) {
	tb.Helper()

	requireSignalsEvent(tb, evt)

	if got := string(evt.SignalsJSON()); got != wantJSON {
		tb.Errorf("signals JSON: got %q, want %q", got, wantJSON)
	}
}

// requireSignalsEvent fails the test unless evt is a patch-signals event,
// mirroring [requireElementsPatch] for the signals family.
func requireSignalsEvent(tb testing.TB, evt Event) {
	tb.Helper()

	if !evt.IsSignals() {
		tb.Fatalf("expected patch-signals event, got %q", evt.Type)

		return
	}
}

// RequireSignalsContain fails the test unless evt is a patch-signals event
// whose JSON payload contains key at any nesting level. This is a convenience
// for checking individual signal keys without decoding the full payload.
func RequireSignalsContain(tb testing.TB, evt Event, key string) {
	tb.Helper()

	requireSignalsEvent(tb, evt)

	jsonStr := string(evt.SignalsJSON())
	needle := `"` + key + `":`

	if !strings.Contains(jsonStr, needle) {
		tb.Errorf("signals JSON should contain key %q; got %s", key, jsonStr)
	}
}

// RequireScript fails the test unless evt is a script-bearing patch-elements
// event whose inner JavaScript source exactly equals wantJS. Use
// [RequireElementsContains] when a substring match on the full HTML suffices.
func RequireScript(tb testing.TB, evt Event, wantJS string) {
	tb.Helper()

	if !evt.IsScript() {
		tb.Fatalf("expected script-bearing patch-elements event, got %q", evt.Type)

		return
	}

	if got := evt.ScriptContent(); got != wantJS {
		tb.Errorf("script content: got %q, want %q", got, wantJS)
	}
}

// RequireEventID fails the test unless the SSE event ID matches want. Use this
// to verify that replayable handlers assign the expected event IDs (e.g., when
// testing reconnection replay driven by [WithLastEventID]).
func RequireEventID(tb testing.TB, evt Event, want string) {
	tb.Helper()

	if evt.ID != want {
		tb.Fatalf("event ID: got %q, want %q", evt.ID, want)
	}
}

// RequireNotScript fails the test if evt is a script-bearing patch-elements
// event. Use it to pin handlers that must NOT respond with JavaScript (e.g.,
// a pure-DOM endpoint where a script patch would be a regression). This is
// the negative counterpart of [RequireScript].
func RequireNotScript(tb testing.TB, evt Event) {
	tb.Helper()

	if !evt.IsScript() {
		return
	}

	tb.Fatalf("expected non-script event, got script patch with content %q", evt.ScriptContent())
}

// RequireScriptTagVersion fails the test unless html embeds the script tag
// for scriptPath exactly as [datastar.ScriptTagVersioned] renders it: a
// <script type="module"> element whose src carries the pinned client version
// as a cache-busting query parameter. This is the consumer-side assertion for
// the versioned-cache contract — a page referencing a bare datastar.js src,
// a stale version, or an unversioned URL fails here, because any of those let
// a browser cache serve an old client against a new server.
func RequireScriptTagVersion(tb testing.TB, html string, scriptPath string) {
	tb.Helper()

	want := datastar.ScriptTagVersioned(scriptPath)
	if !strings.Contains(html, want) {
		tb.Fatalf(
			"script tag: page does not embed %q — a missing or stale cache-busting src risks old-client/new-server protocol skew",
			want,
		)
	}
}

// RequireETagForBundle fails the test unless gotETag equals the ETag the
// script handlers serve for bundle: [datastar.ETagFor] of the exact bytes
// the handler was constructed with. Use it to prove a serving setup (a
// wrapper handler, a proxy in tests, a custom bundle) carries the caching
// identity of the bundle it serves — a mismatched ETag silently disables
// conditional-request reuse or serves stale bytes as fresh.
func RequireETagForBundle(tb testing.TB, gotETag string, bundle []byte) {
	tb.Helper()

	want := datastar.ETagFor(bundle)
	if gotETag != want {
		tb.Fatalf(
			"ETag: got %s, want %s (the digest datastar.ETagFor computes for the served bundle)",
			gotETag,
			want,
		)
	}
}
