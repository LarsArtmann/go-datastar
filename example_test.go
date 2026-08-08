package datastar_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// ExampleElementsPatch demonstrates the library's keystone: a patch is a value
// you construct without a connection. Its wire format is fully determined by
// the struct fields and options.
func ExampleElementsPatch() {
	patch := datastar.NewElementsPatch("<div>Hello</div>",
		datastar.WithSelector("#feed"),
		datastar.WithModePrepend(),
	)

	fmt.Println(patch.Event().Data)
	// Output:
	// selector #feed
	// mode prepend
	// elements <div>Hello</div>
}

// ExampleSignalsPatch shows pre-encoded JSON signals emitted as a wire event.
func ExampleSignalsPatch() {
	patch := datastar.SignalsPatch{Signals: []byte(`{"count":1}`)}

	fmt.Println(patch.Event().Data)
	// Output:
	// signals {"count":1}
}

// ExampleResponse demonstrates the single-connection fluent builder. This is a
// compile-checked regression guard: it exercises every method documented in the
// Response godoc, so a rename or signature change fails CI here rather than
// silently shipping stale documentation.
func ExampleResponse() {
	var buf mockFlushWriter

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	stream := sse.NewStream(&buf, req)
	defer func() { _ = stream.Close() }()

	resp := datastar.NewResponse(stream)

	_ = resp.PatchElements("<div>Hello</div>", datastar.WithSelector("#feed"))
	_ = resp.MarshalAndPatchSignals(map[string]any{"count": 1})
}

// ExampleReadSignals demonstrates extracting signals from an inbound request body.
func ExampleReadSignals() {
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(`{"count":2}`))

	var signals struct {
		Count int `json:"count"`
	}
	if err := datastar.ReadSignals(req, &signals); err != nil {
		return
	}

	fmt.Println(signals.Count)
	// Output: 2
}
