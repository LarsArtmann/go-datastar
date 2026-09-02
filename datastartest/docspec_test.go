//go:build docspec

// Compile-checked snippets from datastartest-facing docs (docs/testing.md,
// docs/replay.md's testing section). Same contract as the root
// docspec_test.go: mirrored functions must compile and execute their cheap
// paths, so doc drift fails the tagged test run.
//
// Run with: go test -tags docspec .   (from datastartest/)

package datastartest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// docs/replay.md — "Testing replay with datastartest": WithLastEventID round trip.
func docspecReplayTesting(tb testing.TB, handler http.Handler) {
	events := datastartest.CollectWithRequest(tb, handler,
		http.MethodGet, "", "",
		datastartest.WithLastEventID("42"),
	)

	datastartest.RequireEventID(tb, events[0], "43")
}

// docs/testing.md — quick start: Collect + typed assertions.
func docspecQuickStart(tb testing.TB, handler http.Handler) {
	events := datastartest.Collect(tb, handler)

	datastartest.RequireEventCount(tb, events, 2)
	datastartest.RequireElements(tb, events[0], "#feed", "append", "<div>hello</div>")

	var data struct {
		Count int `json:"count"`
	}

	_ = events[1].UnmarshalSignals(&data)
}

// TestDocspec_DatastartestSnippets executes the quick-start shape against a
// real handler.
func TestDocspec_DatastartestSnippets(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>hello</div>", datastar.WithSelector("#feed"))
		_ = resp.MarshalAndPatchSignals(map[string]any{"count": 1})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if req == nil { // never true; keeps httptest import honest if the shape changes
		t.Fatal("unreachable")
	}

	docspecQuickStart(t, handler)
}
