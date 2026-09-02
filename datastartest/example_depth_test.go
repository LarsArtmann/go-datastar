// Additional function examples: CollectPost, CollectN, DataValue, String.
// They follow the same pattern as the examples in example_test.go: demonstrate
// the decoded shape with deterministic Output blocks.

package datastartest_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// ExampleCollectPost demonstrates testing a POST handler that reads signals
// from the JSON body — the most common mutation pattern.
func ExampleCollectPost() {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		var signals struct {
			Name string `json:"name"`
		}

		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements(
			"<div>Welcome, "+signals.Name+"</div>",
			datastar.WithSelector("#greeting"),
		)
	})

	// In your test: events := datastartest.CollectPost(t, handler, `{"name":"ada"}`)
	sseOutput := "event: datastar-patch-elements\n" +
		"data: selector #greeting\n" +
		"data: elements <div>Welcome, ada</div>\n\n"

	_ = handler

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))
	fmt.Println("selector:", events[0].Selector())
	fmt.Println("elements:", events[0].Elements())

	// Output:
	// selector: #greeting
	// elements: <div>Welcome, ada</div>
}

// ExampleCollectN demonstrates reading an exact number of events from a
// streaming handler — useful when the handler keeps the connection open.
func ExampleCollectN() {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>first</div>")
		_ = resp.PatchElements("<div>second</div>")
		_ = resp.PatchElements("<div>third</div>")
	})

	// In your test: events := datastartest.CollectN(t, handler, 3)
	sseOutput := "event: datastar-patch-elements\n" +
		"data: elements <div>first</div>\n\n" +
		"event: datastar-patch-elements\n" +
		"data: elements <div>second</div>\n\n" +
		"event: datastar-patch-elements\n" +
		"data: elements <div>third</div>\n\n"

	_ = handler

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))
	fmt.Println("count:", len(events))
	fmt.Println("third:", events[2].Elements())

	// Output:
	// count: 3
	// third: <div>third</div>
}

// ExampleEvent_DataValue demonstrates reading a dataline value by key when
// no typed accessor exists for it. The key is the dataline PREFIX including
// its trailing space (matching the datastar dataline-key constants).
func ExampleEvent_DataValue() {
	sseOutput := "event: datastar-patch-elements\n" +
		"data: selector #feed\n" +
		"data: useViewTransition true\n" +
		"data: elements <div>hello</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))
	fmt.Println("viewTransitions:", events[0].DataValue("useViewTransition "))

	// Output:
	// viewTransitions: true
}

// ExampleEvent_String demonstrates the debug representation used in failure
// messages.
func ExampleEvent_String() {
	sseOutput := "event: datastar-patch-elements\n" +
		"data: elements <div>hello</div>\n\n"

	events, _ := datastartest.ReadEvents(strings.NewReader(sseOutput))
	fmt.Println(events[0].String())

	// Output:
	// Event{type=datastar-patch-elements datalines=1}
}
