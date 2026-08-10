package datastartest_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// helperHandler runs a handler function that has access to a DataStar Response.
// The function sends patches; the test server collects the SSE output.
func helperHandler(send func(*datastar.Response)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		send(resp)
	})
}

func TestEvent_SelectorAndMode(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hello</div>",
			datastar.WithSelector("#feed"),
			datastar.WithMode(datastar.ElementPatchModeAppend),
		)
	}))

	datastartest.RequireEventCount(t, events, 1)

	evt := events[0]
	if !evt.IsElements() {
		t.Fatalf("expected patch-elements, got %q", evt.Type)
	}

	if got := evt.Selector(); got != "#feed" {
		t.Errorf("Selector: got %q, want %q", got, "#feed")
	}

	if got := evt.Mode(); got != "append" {
		t.Errorf("Mode: got %q, want %q", got, "append")
	}
}

func TestEvent_ModeDefaultsToOuter(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hi</div>")
	}))

	evt := events[0]
	if got := evt.Mode(); got != "outer" {
		t.Errorf("Mode default: got %q, want %q", got, "outer")
	}
}

func TestEvent_NamespaceDefaultsToHTML(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hi</div>")
	}))

	evt := events[0]
	if got := evt.Namespace(); got != "html" {
		t.Errorf("Namespace default: got %q, want %q", got, "html")
	}
}

func TestEvent_MultiLineElements(t *testing.T) {
	t.Parallel()

	html := "<div>line1</div>\n<div>line2</div>\n<div>line3</div>"

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements(html, datastar.WithSelector("#feed"))
	}))

	evt := events[0]
	if got := evt.Elements(); got != html {
		t.Errorf("Elements multi-line:\ngot:  %q\nwant: %q", got, html)
	}
}

func TestEvent_SignalsUnmarshal(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.MarshalAndPatchSignals(map[string]any{
			"count": 42,
			"name":  "alice",
		})
	}))

	datastartest.RequireEventCount(t, events, 1)

	evt := events[0]
	if !evt.IsSignals() {
		t.Fatalf("expected patch-signals, got %q", evt.Type)
	}

	var data struct {
		Count int    `json:"count"`
		Name  string `json:"name"`
	}

	if err := evt.UnmarshalSignals(&data); err != nil {
		t.Fatalf("UnmarshalSignals: %v", err)
	}

	if data.Count != 42 {
		t.Errorf("count: got %d, want 42", data.Count)
	}

	if data.Name != "alice" {
		t.Errorf("name: got %q, want %q", data.Name, "alice")
	}
}

func TestEvent_ExecuteScriptBecomesElements(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.ExecuteScript("console.log('hello')")
	}))

	datastartest.RequireEventCount(t, events, 1)

	evt := events[0]
	if !evt.IsElements() {
		t.Fatalf("ExecuteScript should produce patch-elements, got %q", evt.Type)
	}

	if got := evt.Selector(); got != "body" {
		t.Errorf("selector: got %q, want %q", got, "body")
	}

	if got := evt.Mode(); got != "append" {
		t.Errorf("mode: got %q, want %q", got, "append")
	}

	if !strings.Contains(evt.Elements(), "console.log('hello')") {
		t.Errorf("elements should contain script; got %q", evt.Elements())
	}
}

func TestEvent_RemoveElementBecomesElements(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.RemoveElement("#stale")
	}))

	evt := events[0]
	datastartest.RequireElements(t, evt, "#stale", "remove", "")
}

func TestEvent_MultiplePatchesInOrder(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"step": 1})
		_ = resp.PatchElements("<div>2</div>", datastar.WithSelector("#b"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"step": 2})
	}))

	datastartest.RequireEventCount(t, events, 4)

	if got := events[0].Selector(); got != "#a" {
		t.Errorf("event[0] selector: got %q", got)
	}

	if got := events[2].Selector(); got != "#b" {
		t.Errorf("event[2] selector: got %q", got)
	}

	var sig1 struct {
		Step int `json:"step"`
	}

	if err := events[1].UnmarshalSignals(&sig1); err != nil {
		t.Fatal(err)
	}

	if sig1.Step != 1 {
		t.Errorf("event[1] step: got %d, want 1", sig1.Step)
	}
}

func TestEvent_RedirectPatch(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.Redirect("https://example.com")
	}))

	evt := events[0]
	if !strings.Contains(evt.Elements(), "https://example.com") {
		t.Errorf("redirect elements should contain URL; got %q", evt.Elements())
	}
}

func TestEvent_OnlyIfMissingFlag(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		patch, err := datastar.NewSignalsPatch(
			map[string]any{"x": 1},
			datastar.WithOnlyIfMissing(true),
		)
		if err != nil {
			return
		}

		_ = resp.ApplyPatches(patch)
	}))

	evt := events[0]
	if !evt.OnlyIfMissing() {
		t.Error("OnlyIfMissing should be true")
	}
}

func TestEvent_ViewTransitions(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>vt</div>",
			datastar.WithSelector("#feed"),
			datastar.WithViewTransitions(true),
			datastar.WithViewTransitionSelector("#scope"),
		)
	}))

	evt := events[0]
	if !evt.UseViewTransitions() {
		t.Error("UseViewTransitions should be true")
	}

	if got := evt.ViewTransitionSelector(); got != "#scope" {
		t.Errorf("ViewTransitionSelector: got %q, want %q", got, "#scope")
	}
}

func TestFilter_Elements(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>")
		_ = resp.MarshalAndPatchSignals(map[string]any{"x": 1})
		_ = resp.PatchElements("<div>2</div>")
	}))

	els := datastartest.FilterElements(events)
	if len(els) != 2 {
		t.Fatalf("FilterElements: got %d, want 2", len(els))
	}

	sigs := datastartest.FilterSignals(events)
	if len(sigs) != 1 {
		t.Fatalf("FilterSignals: got %d, want 1", len(sigs))
	}
}

func TestEvent_ScriptContent_ExecuteScript(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.ExecuteScript("console.log('test')")
	}))

	evt := events[0]

	got := evt.ScriptContent()
	if !strings.Contains(got, "console.log('test')") {
		t.Errorf("ScriptContent should contain JS; got %q", got)
	}
}

func TestEvent_ScriptContent_Redirect(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.Redirect("https://example.com")
	}))

	evt := events[0]

	got := evt.ScriptContent()
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("ScriptContent for redirect should contain URL; got %q", got)
	}
}

func TestEvent_ScriptContent_ConsoleLog(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.ConsoleLog("hello world")
	}))

	evt := events[0]

	got := evt.ScriptContent()
	if !strings.Contains(got, "console.log") {
		t.Errorf("ScriptContent for ConsoleLog should contain console.log; got %q", got)
	}
}

func TestEvent_ScriptContent_NonScriptReturnsEmpty(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>not a script</div>", datastar.WithSelector("#feed"))
	}))

	evt := events[0]
	if got := evt.ScriptContent(); got != "" {
		t.Errorf("ScriptContent for non-script elements should be empty; got %q", got)
	}
}

func TestEvent_DataValue(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hi</div>",
			datastar.WithSelector("#feed"),
			datastar.WithMode(datastar.ElementPatchModeAppend),
		)
	}))

	evt := events[0]
	if got := evt.DataValue("selector "); got != "#feed" {
		t.Errorf("DataValue selector: got %q, want %q", got, "#feed")
	}

	if got := evt.DataValue("mode "); got != "append" {
		t.Errorf("DataValue mode: got %q, want %q", got, "append")
	}

	if got := evt.DataValue("nonexistent "); got != "" {
		t.Errorf("DataValue for missing key should be empty; got %q", got)
	}
}

func TestEvent_String(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hi</div>", datastar.WithSelector("#feed"))
	}))

	got := events[0].String()
	if !strings.Contains(got, "datastar-patch-elements") {
		t.Errorf("String should contain event type; got %q", got)
	}

	if !strings.Contains(got, "datalines=") {
		t.Errorf("String should contain dataline count; got %q", got)
	}
}

func TestEvent_RetryEventIDRoundTrip(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		patch := datastar.NewElementsPatch("<div>retry test</div>",
			datastar.WithSelector("#feed"),
			datastar.WithElementsEventID("evt-42"),
			datastar.WithElementsRetryDuration(5000*time.Millisecond),
		)
		_ = resp.ApplyPatches(patch)
	}))

	evt := events[0]
	if evt.ID != "evt-42" {
		t.Errorf("Event ID: got %q, want %q", evt.ID, "evt-42")
	}

	if evt.Retry != 5000 {
		t.Errorf("Retry: got %d, want 5000", evt.Retry)
	}

	desc := evt.String()
	if !strings.Contains(desc, "id=evt-42") {
		t.Errorf("String should contain event ID; got %q", desc)
	}

	if !strings.Contains(desc, "retry=5000") {
		t.Errorf("String should contain retry; got %q", desc)
	}
}

func TestEvent_EmptyHandler(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		// Send nothing
	}))

	if len(events) != 0 {
		t.Errorf("empty handler: got %d events, want 0", len(events))
	}
}

func TestDatalineConstants_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dataline string
	}{
		{"selector", datastar.SelectorDatalineKey},
		{"mode", datastar.ModeDatalineKey},
		{"namespace", datastar.NamespaceDatalineKey},
		{"elements", datastar.ElementsDatalineKey},
		{"signals", datastar.SignalsDatalineKey},
		{"onlyIfMissing", datastar.OnlyIfMissingDatalineKey},
		{"useViewTransition", datastar.UseViewTransitionDatalineKey},
		{"viewTransitionSelector", datastar.ViewTransitionSelectorDatalineKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !strings.HasSuffix(tt.dataline, " ") {
				t.Errorf("dataline key %q should have trailing space", tt.dataline)
			}

			evt := datastartest.Event{
				Type:      "datastar-patch-elements",
				DataLines: []string{tt.dataline + "test-value"},
			}

			if got := evt.DataValue(tt.dataline); got != "test-value" {
				t.Errorf("DataValue(%q): got %q, want %q", tt.dataline, got, "test-value")
			}
		})
	}
}

func TestEvent_IsScript_ExecuteScript(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.ExecuteScript("console.log('test')")
	}))

	if !events[0].IsScript() {
		t.Errorf("ExecuteScript event should be a script; got %q", events[0].Elements())
	}
}

func TestEvent_IsScript_RegularElements(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>not a script</div>", datastar.WithSelector("#feed"))
	}))

	if events[0].IsScript() {
		t.Errorf("regular elements event should not be a script; got %q", events[0].Elements())
	}
}

func TestEvent_ScriptContent_DispatchCustomEvent(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.DispatchCustomEvent("item-added", map[string]any{"id": 1})
	}))

	evt := events[0]
	if !evt.IsScript() {
		t.Fatalf("DispatchCustomEvent should produce a script patch; got %q", evt.Elements())
	}

	got := evt.ScriptContent()
	if !strings.Contains(got, "CustomEvent") {
		t.Errorf("ScriptContent for DispatchCustomEvent should contain CustomEvent; got %q", got)
	}

	if !strings.Contains(got, "item-added") {
		t.Errorf("ScriptContent should contain event name; got %q", got)
	}
}

func TestEvent_ScriptContent_Prefetch(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.Prefetch("https://example.com/page1")
	}))

	evt := events[0]
	if !evt.IsScript() {
		t.Fatalf("Prefetch should produce a script patch; got %q", evt.Elements())
	}

	got := evt.ScriptContent()
	if !strings.Contains(got, "example.com") {
		t.Errorf("ScriptContent for Prefetch should contain URL; got %q", got)
	}
}

func TestEvent_ScriptContent_AttributeWithGreaterThan(t *testing.T) {
	t.Parallel()

	evt := datastartest.Event{
		Type: "datastar-patch-elements",
		DataLines: []string{
			"elements <script type=\"speculationrules\" data-x=\"a>b\">console.log('inner')</script>",
		},
	}

	got := evt.ScriptContent()
	if !strings.Contains(got, "console.log('inner')") {
		t.Errorf("ScriptContent should handle > in attribute value; got %q", got)
	}

	if strings.Contains(got, "b\">") {
		t.Errorf("ScriptContent should not include attribute fragment; got %q", got)
	}
}

func TestFindElement(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"step": 1})
		_ = resp.PatchElements("<div>2</div>", datastar.WithSelector("#b"))
	}))

	evt, ok := datastartest.FindElement(events, "#b")
	if !ok {
		t.Fatal("FindElement should find #b")
	}

	if got := evt.Elements(); got != "<div>2</div>" {
		t.Errorf("FindElement #b elements: got %q, want %q", got, "<div>2</div>")
	}

	if _, ok := datastartest.FindElement(events, "#nonexistent"); ok {
		t.Error("FindElement should return false for missing selector")
	}
}

func TestFindSignals(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"step": 1})
	}))

	evt, ok := datastartest.FindSignals(events)
	if !ok {
		t.Fatal("FindSignals should find a signals event")
	}

	if !evt.IsSignals() {
		t.Errorf("FindSignals should return a signals event; got type %q", evt.Type)
	}
}

func TestFindSignals_NotFound(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
	}))

	if _, ok := datastartest.FindSignals(events); ok {
		t.Error("FindSignals should return false when no signals events exist")
	}
}

func TestRequireSignalsContain(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.MarshalAndPatchSignals(map[string]any{"count": 42, "name": "alice"})
	}))

	datastartest.RequireSignalsContain(t, events[0], "count")
	datastartest.RequireSignalsContain(t, events[0], "name")
}

func TestEventsString(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"x": 1})
	}))

	got := datastartest.EventsString(events)
	if !strings.Contains(got, "datastar-patch-elements") {
		t.Errorf("EventsString should contain event type; got %q", got)
	}

	if !strings.Contains(got, "datastar-patch-signals") {
		t.Errorf("EventsString should contain second event type; got %q", got)
	}
}

func TestEventsString_Empty(t *testing.T) {
	t.Parallel()

	got := datastartest.EventsString(nil)
	if got != "(no events)" {
		t.Errorf("EventsString(nil): got %q, want %q", got, "(no events)")
	}
}

func TestEvent_DataValue_MultiLineReturnsFirst(t *testing.T) {
	t.Parallel()

	// DataValue returns only the first match for multi-line keys.
	// The typed accessors (Elements(), SignalsJSON()) use allValues + join.
	evt := datastartest.Event{
		Type: "datastar-patch-elements",
		DataLines: []string{
			"elements <div>first</div>",
			"elements <div>second</div>",
			"elements <div>third</div>",
		},
	}

	got := evt.DataValue("elements ")
	if got != "<div>first</div>" {
		t.Errorf("DataValue multi-line should return first match; got %q, want %q", got, "<div>first</div>")
	}

	// Elements() rejoins all lines.
	all := evt.Elements()
	want := "<div>first</div>\n<div>second</div>\n<div>third</div>"

	if all != want {
		t.Errorf("Elements multi-line: got %q, want %q", all, want)
	}
}

func TestEvent_ConcurrentCollect(t *testing.T) {
	t.Parallel()

	// Multiple Collect calls in parallel should not race.
	const goroutines = 8

	done := make(chan struct{})

	for range goroutines {
		go func() {
			defer func() { done <- struct{}{} }()

			handler := helperHandler(func(resp *datastar.Response) {
				_ = resp.PatchElements("<div>concurrent</div>", datastar.WithSelector("#feed"))
			})

			events := datastartest.Collect(t, handler)
			if len(events) != 1 {
				t.Errorf("concurrent Collect: got %d events, want 1", len(events))
			}
		}()
	}

	for range goroutines {
		<-done
	}
}
