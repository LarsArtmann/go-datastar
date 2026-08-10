package datastartest_test

import (
	"net/http"
	"strings"
	"testing"

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
