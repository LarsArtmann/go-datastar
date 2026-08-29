package datastar_test

import (
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
)

// TestPatchWireGoldens pins the exact SSE wire bytes every patch family
// produces through Patch.Event() + sse.WriteEvent. These goldens ARE the
// DataStar protocol compatibility contract: event names, data-line keys and
// ordering, default-value elision, id/retry field placement, and the
// one-data-line-per-source-line splitting for multi-line payloads (including
// the empty `data: elements ` line produced by a blank JS line — the space
// after the colon is stripped by SSE parsers, so it round-trips as empty).
//
// A change here is a wire-format change: make it deliberately, against the
// DataStar SDK, and record it as a breaking change in CHANGELOG.md.
func TestPatchWireGoldens(t *testing.T) {
	t.Parallel()

	signalsDefault, err := datastar.NewSignalsPatch(map[string]any{"count": 1})
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	signalsOptions, err := datastar.NewSignalsPatch(map[string]any{"count": 1},
		datastar.WithOnlyIfMissing(true),
		datastar.WithSignalsEventID("sig-1"),
		datastar.WithSignalsRetryDuration(3*time.Second),
	)
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	customDefault, err := datastar.NewDispatchCustomEventPatch(
		"feed:updated", map[string]int{"n": 3})
	if err != nil {
		t.Fatalf("NewDispatchCustomEventPatch: %v", err)
	}

	customOptions, err := datastar.NewDispatchCustomEventPatch("feed:updated", nil,
		datastar.WithCustomEventSelector(".item"),
		datastar.WithCustomEventBubbles(false),
		datastar.WithCustomEventCancelable(false),
		datastar.WithCustomEventComposed(false),
	)
	if err != nil {
		t.Fatalf("NewDispatchCustomEventPatch: %v", err)
	}

	tests := []struct {
		name  string
		patch datastar.Patch
		want  string
	}{
		{
			name:  "elements default outer mode elided",
			patch: datastar.NewElementsPatch("<div>hello</div>"),
			want: `event: datastar-patch-elements
data: elements <div>hello</div>

`,
		},
		{
			name: "elements selector then mode inner",
			patch: datastar.NewElementsPatch("<div>hi</div>",
				datastar.WithSelector("#feed"),
				datastar.WithModeInner(),
			),
			want: `event: datastar-patch-elements
data: selector #feed
data: mode inner
data: elements <div>hi</div>

`,
		},
		{
			name: "elements every option, multi-line HTML, id and retry",
			patch: datastar.NewElementsPatch("<div>a</div>\n<div>b</div>",
				datastar.WithSelector("#feed"),
				datastar.WithModeReplace(),
				datastar.WithNamespaceSVG(),
				datastar.WithViewTransitionsEnabled(),
				datastar.WithViewTransitionSelector("#item"),
				datastar.WithElementsEventID("evt-42"),
				datastar.WithElementsRetryDuration(2*time.Second),
			),
			want: `event: datastar-patch-elements
data: selector #feed
data: mode replace
data: namespace svg
data: useViewTransition true
data: viewTransitionSelector #item
data: elements <div>a</div>
data: elements <div>b</div>
id: evt-42
retry: 2000

`,
		},
		{
			name:  "elements remove sugar emits no elements line",
			patch: datastar.NewRemovePatch("#gone"),
			want: `event: datastar-patch-elements
data: selector #gone
data: mode remove

`,
		},
		{
			name:  "signals default",
			patch: signalsDefault,
			want: `event: datastar-patch-signals
data: signals {"count":1}

`,
		},
		{
			name:  "signals onlyIfMissing, id and retry",
			patch: signalsOptions,
			want: `event: datastar-patch-signals
data: onlyIfMissing true
data: signals {"count":1}
id: sig-1
retry: 3000

`,
		},
		{
			name:  "script default autoRemove wraps with data-effect",
			patch: datastar.NewScriptPatch("console.log('hi')"),
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script data-effect="el.remove()">console.log('hi')</script>

`,
		},
		{
			name: "script attributes, autoRemove off, id and retry",
			patch: datastar.NewScriptPatch("doThing()",
				datastar.WithScriptAutoRemove(false),
				datastar.WithScriptAttributes(`type="module"`),
				datastar.WithScriptEventID("scr-7"),
				datastar.WithScriptRetryDuration(1500*time.Millisecond),
			),
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script type="module">doThing()</script>
id: scr-7
retry: 1500

`,
		},
		{
			name:  "redirect sugar",
			patch: datastar.NewRedirectPatch("/dashboard"),
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script data-effect="el.remove()">setTimeout(() => window.location.href = "/dashboard")</script>

`,
		},
		{
			name:  "console log sugar",
			patch: datastar.NewConsoleLogPatch("done"),
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script data-effect="el.remove()">console.log("done")</script>

`,
		},
		{
			// The JS body is tab-indented; each source line becomes its own
			// `data: elements` line. The blank JS lines produce "data: elements "
			// WITH a trailing space (key prefix) — parsers strip it, round-trip
			// tests pin that.
			name:  "custom event default document target",
			patch: customDefault,
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script data-effect="el.remove()">{
data: elements 	const elements = [document]
data: elements 
data: elements 	const event = new CustomEvent("feed:updated", {
data: elements 		bubbles: true,
data: elements 		cancelable: true,
data: elements 		composed: true,
data: elements 		detail: {"n":3},
data: elements 	});
data: elements 
data: elements 	elements.forEach((element) => {
data: elements 		element.dispatchEvent(event);
data: elements 	});
data: elements }
data: elements 	</script>

`,
		},
		{
			name:  "custom event selector and inverted flags",
			patch: customOptions,
			want: `event: datastar-patch-elements
data: selector body
data: mode append
data: elements <script data-effect="el.remove()">{
data: elements 	const elements = document.querySelectorAll(".item")
data: elements 
data: elements 	const event = new CustomEvent("feed:updated", {
data: elements 		bubbles: false,
data: elements 		cancelable: false,
data: elements 		composed: false,
data: elements 		detail: null,
data: elements 	});
data: elements 
data: elements 	elements.forEach((element) => {
data: elements 		element.dispatchEvent(event);
data: elements 	});
data: elements }
data: elements 	</script>

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := writeEvent(t, tt.patch); got != tt.want {
				t.Errorf("wire mismatch\n--- want ---\n%s--- got ---\n%s", tt.want, got)
			}
		})
	}
}
