package datastartest_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

// sliceStore is a minimal sse.EventStore for the replay dogfood test: it
// returns every stored event whose ID sorts strictly after the request's
// Last-Event-ID, mirroring what a real replay store does on reconnection.
type sliceStore []sse.Event

func (s sliceStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	var after []sse.Event

	for _, evt := range s {
		if evt.ID.Get() > lastID.Get() {
			after = append(after, evt)
		}
	}

	return after, nil
}

// TestE2E_DataStarPatches verifies the complete SSE wire format produced by
// go-datastar through a real HTTP server and client — using the datastartest
// helper package to parse and decode the SSE response.
//
// This test also serves as a dogfood integration test for datastartest itself:
// it exercises Collect, RequireElements, RequireElementsContains,
// UnmarshalSignals, and the typed accessors against real wire output.
//
// It lives in the datastartest module (not the root module) to avoid a circular
// module dependency: root must never depend on datastartest in its go.mod, and
// datastartest already depends on root for production code.
func TestE2E_DataStarPatches(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(
		t,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		}),
	)

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

// TestE2E_ReplayWithLastEventID dogfoods the full reconnection story: the
// handler replays missed events from an EventStore based on the Last-Event-ID
// header, and the test drives the reconnect with [WithLastEventID] — no
// browser required.
func TestE2E_ReplayWithLastEventID(t *testing.T) {
	t.Parallel()

	store := sliceStore{
		{
			Event: string(datastar.EventTypePatchElements),
			Data:  "selector #feed\nelements <div>1</div>",
			ID:    sse.NewEventID("1"),
		},
		{
			Event: string(datastar.EventTypePatchElements),
			Data:  "selector #feed\nelements <div>2</div>",
			ID:    sse.NewEventID("2"),
		},
		{
			Event: string(datastar.EventTypePatchElements),
			Data:  "selector #feed\nelements <div>3</div>",
			ID:    sse.NewEventID("3"),
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		if _, err := sse.Replay(stream, store, datastar.LastEventID(r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>live</div>", datastar.WithSelector("#feed"))
	})

	t.Run("fresh connection replays everything", func(t *testing.T) {
		t.Parallel()

		events := datastartest.Collect(t, handler)
		datastartest.RequireEventCount(t, events, 4) // 3 replayed + 1 live

		datastartest.RequireEventID(t, events[0], "1")
		datastartest.RequireEventID(t, events[2], "3")
		datastartest.RequireElements(t, events[3], "#feed", "outer", "<div>live</div>")
	})

	t.Run("reconnect from ID 2 replays only missed events", func(t *testing.T) {
		t.Parallel()

		events := datastartest.Collect(t, handler, datastartest.WithLastEventID("2"))
		datastartest.RequireEventCount(t, events, 2) // event 3 + live

		datastartest.RequireEventID(t, events[0], "3")
		datastartest.RequireElements(t, events[1], "#feed", "outer", "<div>live</div>")
	})
}

// TestE2E_CollectPostRoundTrip dogfoods [datastartest.CollectPost]: the
// handler reads inbound signals from the POST body via
// [datastar.ReadSignals] and answers with a signals patch derived from them —
// the canonical form-submit flow without a browser.
func TestE2E_CollectPostRoundTrip(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var inbound struct {
			Email string `json:"email"`
		}

		if err := datastar.ReadSignals(r, &inbound); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.MarshalAndPatchSignals(map[string]any{
			"echo": inbound.Email,
			"len":  len(inbound.Email),
		})
	})

	events := datastartest.CollectPost(t, handler, `{"email":"lars@example.com"}`)
	datastartest.RequireEventCount(t, events, 1)

	var signals struct {
		Echo string `json:"echo"`
		Len  int    `json:"len"`
	}

	if err := events[0].UnmarshalSignals(&signals); err != nil {
		t.Fatalf("unmarshal signals: %v", err)
	}

	if signals.Echo != "lars@example.com" || signals.Len != 16 {
		t.Errorf("round-trip signals: got (%q, %d), want (%q, %d)",
			signals.Echo, signals.Len, "lars@example.com", 16)
	}
}

// TestE2E_CollectNStreaming dogfoods [datastartest.CollectN] against a
// handler that keeps the connection open after sending its events: CollectN
// must return as soon as the requested count arrives, without waiting for the
// handler to finish.
func TestE2E_CollectNStreaming(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		for i := 1; i <= 3; i++ {
			_ = resp.PatchElements(
				fmt.Sprintf("<div>tick %d</div>", i),
				datastar.WithSelector("#ticks"),
			)
		}

		// Hold the connection open like a broadcaster-backed handler would;
		// CollectN(2) below must not block on this.
		<-r.Context().Done()
	})

	events := datastartest.CollectN(t, handler, 2)
	datastartest.RequireEventCount(t, events, 2)

	datastartest.RequireElements(t, events[0], "#ticks", "outer", "<div>tick 1</div>")
	datastartest.RequireElements(t, events[1], "#ticks", "outer", "<div>tick 2</div>")
}
