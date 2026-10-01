package datastartest_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// ExampleCollect demonstrates the simplest way to E2E test a DataStar handler.
//
// datastartest.Collect spins up a test server, sends a GET request, parses the
// SSE response, and returns decoded events with typed accessors.
func ExampleCollect() {
	// Your production handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>hello</div>", datastar.WithSelector("#feed"))
	})

	// In your test: events := datastartest.Collect(t, handler)
	//
	// Here we parse the same wire format manually to demonstrate the decoded shape.
	sseOutput := "event: datastar-patch-elements\n" +
		"data: selector #feed\n" +
		"data: elements <div>hello</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))
	fmt.Println("type:", events[0].Type)
	fmt.Println("selector:", events[0].Selector())
	fmt.Println("elements:", events[0].Elements())

	_ = handler
	// Output:
	// type: datastar-patch-elements
	// selector: #feed
	// elements: <div>hello</div>
}

// ExampleEvent_unmarshalSignals demonstrates decoding signals JSON into a struct.
func ExampleEvent_unmarshalSignals() {
	sseOutput := "event: datastar-patch-signals\n" +
		`data: signals {"count":42,"name":"alice"}` + "\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	var data struct {
		Count int    `json:"count"`
		Name  string `json:"name"`
	}

	_ = events[0].UnmarshalSignals(&data)
	fmt.Printf("count=%d name=%s", data.Count, data.Name)
	// Output: count=42 name=alice
}

// ExampleFilterElements demonstrates filtering events by type.
func ExampleFilterElements() {
	sseOutput := "event: datastar-patch-elements\ndata: elements <div>1</div>\n\n" +
		"event: datastar-patch-signals\ndata: signals {\"x\":1}\n\n" +
		"event: datastar-patch-elements\ndata: elements <div>2</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	elements := datastartest.FilterElements(events)
	signals := datastartest.FilterSignals(events)

	fmt.Printf("%d elements, %d signals", len(elements), len(signals))
	// Output: 2 elements, 1 signals
}

// ExampleRequireElements demonstrates the assertion helpers.
func ExampleRequireElements() {
	// In your test:
	//
	//   events := datastartest.Collect(t, handler)
	//   datastartest.RequireEventCount(t, events, 1)
	//   datastartest.RequireElements(t, events[0], "#feed", "append", "<div>hello</div>")
	//   datastartest.RequireElementsContains(t, events[0], "body", "append", "console.log")
	//
	// These helpers produce clear failure messages showing exactly what mismatched.
	fmt.Println("Assert helpers: RequireElements, RequireElementsContains, RequireSignals")
	// Output: Assert helpers: RequireElements, RequireElementsContains, RequireSignals
}

// ExampleEvent_scriptContent demonstrates extracting JavaScript from a script patch.
// Script patches (ExecuteScript, Redirect, ConsoleLog, etc.) wrap JS in <script> tags
// inside a patch-elements event. ScriptContent strips the wrapper and returns the JS.
func ExampleEvent_scriptContent() {
	sseOutput := "event: datastar-patch-elements\n" +
		"data: selector body\n" +
		"data: mode append\n" +
		"data: elements <script>console.log('hello')</script>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	fmt.Println(events[0].IsScript())
	fmt.Println(events[0].ScriptContent())
	// Output:
	// true
	// console.log('hello')
}

// ExampleFindElement demonstrates finding a specific elements patch by selector
// when a handler sends multiple patches.
func ExampleFindElement() {
	sseOutput := "event: datastar-patch-elements\ndata: selector #header\ndata: elements <h1>Title</h1>\n\n" +
		"event: datastar-patch-signals\ndata: signals {\"count\":1}\n\n" +
		"event: datastar-patch-elements\ndata: selector #body\ndata: elements <p>Content</p>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	evt, ok := datastartest.FindElement(events, "#body")
	fmt.Printf("found=%v elements=%s", ok, evt.Elements())
	// Output: found=true elements=<p>Content</p>
}

// ExampleFindSignals demonstrates finding the first signals event in a stream.
func ExampleFindSignals() {
	sseOutput := "event: datastar-patch-elements\ndata: elements <div>1</div>\n\n" +
		"event: datastar-patch-signals\ndata: signals {\"step\":2}\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	evt, ok := datastartest.FindSignals(events)
	fmt.Printf("found=%v type=%s", ok, evt.Type)
	// Output: found=true type=datastar-patch-signals
}

// ExampleEventsString demonstrates the multi-event debug representation.
// Useful for logging when a test assertion fails on a specific event.
func ExampleEventsString() {
	sseOutput := "event: datastar-patch-elements\ndata: elements <div>1</div>\n\n" +
		"event: datastar-patch-signals\ndata: signals {\"x\":1}\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	fmt.Println(datastartest.EventsString(events))
	// Output:
	// Event{type=datastar-patch-elements datalines=1}
	// Event{type=datastar-patch-signals datalines=1}
}

// ExampleFindScript demonstrates finding the first script patch in a stream.
// Script patches (ExecuteScript, Redirect, ConsoleLog, ...) wrap JavaScript
// in <script> tags inside patch-elements events; FindScript returns the
// first one, (Event, bool)-style like FindSignals. Use [datastartest.RequireScript]
// when the exact JavaScript content must match.
func ExampleFindScript() {
	sseOutput := "event: datastar-patch-elements\ndata: selector #feed\ndata: elements <div>1</div>\n\n" +
		"event: datastar-patch-elements\ndata: selector body\ndata: elements <script>console.log('done')</script>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	evt, ok := datastartest.FindScript(events)
	fmt.Printf("found=%v js=%s", ok, evt.ScriptContent())
	// Output: found=true js=console.log('done')
}

// ExampleFindAllElements demonstrates the plural counterpart of FindElement:
// every elements patch for one selector, in stream order — for handlers that
// legitimately patch the same target more than once (feed appends, progress).
// Script-bearing patches are elements patches too and match when their
// selector matches.
func ExampleFindAllElements() {
	sseOutput := "event: datastar-patch-elements\ndata: selector #feed\ndata: elements <div>1</div>\n\n" +
		"event: datastar-patch-signals\ndata: signals {\"x\":1}\n\n" +
		"event: datastar-patch-elements\ndata: selector #feed\ndata: elements <div>2</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	for _, evt := range datastartest.FindAllElements(events, "#feed") {
		fmt.Println(evt.Elements())
	}
	// Output:
	// <div>1</div>
	// <div>2</div>
}

// ExampleEventToSelectorMap demonstrates O(1) lookup by selector. When
// several events share a selector, the last patch wins — later patches
// overwrite earlier DOM state, so the map reflects the DOM the client ends
// with. Use FindAllElements when you need every occurrence instead.
func ExampleEventToSelectorMap() {
	sseOutput := "event: datastar-patch-elements\ndata: selector #title\ndata: elements <h1>Draft</h1>\n\n" +
		"event: datastar-patch-elements\ndata: selector #title\ndata: elements <h1>Final</h1>\n\n" +
		"event: datastar-patch-elements\ndata: selector #body\ndata: elements <p>Text</p>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	bySelector := datastartest.EventToSelectorMap(events)
	fmt.Printf("selectors=%d title=%s", len(bySelector), bySelector["#title"].Elements())
	// Output: selectors=2 title=<h1>Final</h1>
}

// ExampleRequireNotScript demonstrates pinning a handler that must NOT
// respond with JavaScript — the negative counterpart of RequireScript, for
// pure-DOM endpoints where a script patch would be a regression.
//
// In your test:
//
//	events := datastartest.Collect(t, handler)
//	datastartest.RequireNotScript(t, events[0])
//
// A script patch reaching such an endpoint fails with a message showing the
// JavaScript that regressed. Under the hood the assertion checks IsScript.
func ExampleRequireNotScript() {
	sseOutput := "event: datastar-patch-elements\ndata: selector #feed\ndata: elements <div>hello</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))

	fmt.Println(events[0].IsScript())
	// Output: false
}
