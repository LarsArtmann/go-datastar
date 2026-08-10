package datastar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// TestE2E_DataStarPatches verifies the complete SSE wire format produced by
// go-datastar through a real HTTP server and client — using the datastartest
// helper package to parse and decode the SSE response.
//
// This test also serves as a dogfood integration test for datastartest itself:
// it exercises Collect, RequireElements, RequireElementsContains,
// UnmarshalSignals, and the typed accessors against real wire output.
func TestE2E_DataStarPatches(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		// 1. Patch elements with selector + mode
		_ = resp.PatchElements("<div>hello</div>",
			datastar.WithSelector("#feed"),
			datastar.WithMode(datastar.ElementPatchModeAppend),
		)

		// 2. Patch signals
		_ = resp.MarshalAndPatchSignals(map[string]any{"count": 1})

		// 3. Execute script (emits patch-elements, NOT execute-script)
		_ = resp.ExecuteScript("console.log('hi')")

		// 4. Remove element
		_ = resp.RemoveElement("#stale")
	}))

	datastartest.RequireEventCount(t, events, 4)

	// 1. Elements patch: exact selector, mode, and HTML
	datastartest.RequireElements(t, events[0], "#feed", "append", "<div>hello</div>")

	// 2. Signals patch: decode JSON into a typed struct
	datastartest.RequireEventType(t, events[1], "datastar-patch-signals")

	var signals struct {
		Count int `json:"count"`
	}

	if err := events[1].UnmarshalSignals(&signals); err != nil {
		t.Fatalf("unmarshal signals: %v", err)
	}

	if signals.Count != 1 {
		t.Errorf("signals count: got %d, want 1", signals.Count)
	}

	// 3. ExecuteScript → patch-elements with selector=body, mode=append,
	//    script wrapped in <script> tags
	datastartest.RequireElementsContains(t, events[2], "body", "append", "console.log('hi')")

	// 4. RemoveElement → patch-elements with selector=#stale, mode=remove
	datastartest.RequireElements(t, events[3], "#stale", "remove", "")
}

// TestE2E_SSEHeaders verifies the transport headers that go-sse sets on SSE
// responses. These are owned by go-sse (NewStream), not go-datastar, but are
// critical for the DataStar client's connection management.
func TestE2E_SSEHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	wantHeaders := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	for header, want := range wantHeaders {
		if got := resp.Header.Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}
