// Package datastartest provides helpers for E2E testing DataStar handlers.
//
// It solves the two problems that make DataStar handlers hard to test:
//
//  1. Parsing the SSE wire format (event:/data:/id:/retry: lines) into events.
//  2. Decoding DataStar datalines (selector/mode/elements/signals key-value
//     pairs) back into typed, assertable values.
//
// The SSE parsing is delegated to the shared go-sse/ssetest parser, so
// conformance fixes land once for both consumers and the fuzz corpora stay in
// lockstep. Helper errors are classified per the go-error-family contract:
// stable codes live in errors.go (datastartest.sse_scan_failed,
// datastartest.signals_unmarshal_failed,
// datastartest.custom_event_detail_unmarshal_failed).
//
// # Quick start
//
//	import (
//	    "github.com/larsartmann/go-datastar"
//	    "github.com/larsartmann/go-datastar/datastartest"
//	    "github.com/larsartmann/go-sse"
//	)
//
//	func TestFeedHandler(t *testing.T) {
//	    events := datastartest.Collect(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        stream := sse.NewStream(w, r)
//	        defer func() { _ = stream.Close() }()
//
//	        resp := datastar.NewResponse(stream)
//	        _ = resp.PatchElements("<div>hello</div>", datastar.WithSelector("#feed"))
//	        _ = resp.MarshalAndPatchSignals(map[string]any{"count": 1})
//	    }))
//
//	    datastartest.RequireEventCount(t, events, 2)
//
//	    // Elements patch: typed accessors decode the datalines
//	    el := events[0]
//	    if el.Selector() != "#feed" {
//	        t.Errorf("selector: got %q, want %q", el.Selector(), "#feed")
//	    }
//	    if el.Elements() != "<div>hello</div>" {
//	        t.Errorf("elements: got %q", el.Elements())
//	    }
//
//	    // Signals patch: unmarshal into a struct
//	    var signals struct {
//	        Count int `json:"count"`
//	    }
//	    if err := events[1].UnmarshalSignals(&signals); err != nil {
//	        t.Fatalf("unmarshal signals: %v", err)
//	    }
//	    if signals.Count != 1 {
//	        t.Errorf("count: got %d, want 1", signals.Count)
//	    }
//	}
//
// # Non-GET requests
//
// For POST/PUT/PATCH with request bodies, use [CollectPost] or
// [CollectWithRequest]:
//
//	events := datastartest.CollectPost(t, handler, `{"name":"alice"}`)
//	events := datastartest.CollectWithRequest(t, handler, http.MethodPut, body, "application/json")
//
// # Request options
//
// Every Collect helper accepts [RequestOption]s. Use [WithPath] to target a
// route on a mux (query strings allowed), [WithDatastarSignals] to submit
// inbound signals the way DataStar clients do with GET/DELETE (see
// datastar.ReadSignals), [WithLastEventID] to simulate a reconnecting browser
// for replay testing, and [WithHeader] for any custom header:
//
//	events := datastartest.Collect(t, mux, datastartest.WithPath("/events?filter=alerts"))
//	events := datastartest.Collect(t, h, datastartest.WithDatastarSignals(`{"filter":"alerts"}`))
//	events := datastartest.Collect(t, h, datastartest.WithLastEventID("42"))
//
// # Streaming handlers
//
// For handlers that keep the connection open (e.g., broadcasting through a
// go-sse Broadcaster), use [CollectN] to read exactly N events then close.
// Use [CollectWithTimeout] for a time-bounded read that returns whatever
// events arrived before the deadline:
//
//	events := datastartest.CollectN(t, handler, 3)
//	events := datastartest.CollectWithTimeout(t, handler, 5*time.Second)
//
// # Script patches
//
// Script patches (ExecuteScript, Redirect, ConsoleLog, etc.) produce
// patch-elements events with JS wrapped in <script> tags. Use [Event.IsScript]
// to check, [Event.ScriptContent] to extract the inner JavaScript source, and
// [RequireScript] to assert the exact source in one call:
//
//	events := datastartest.Collect(t, handler)
//	if events[0].IsScript() {
//	    js := events[0].ScriptContent() // "console.log('hello')"
//	}
//	datastartest.RequireScript(t, events[0], "console.log('hello')")
//
// # Finding events
//
// When a handler sends multiple patches, use [FindElement] and [FindSignals]
// to locate specific events by selector or type without indexing by position:
//
//	evt, ok := datastartest.FindElement(events, "#header")
//	sigEvt, ok := datastartest.FindSignals(events)
//
// # Ginkgo and benchmarks
//
// All helpers that take a `t` accept [testing.TB], not *testing.T, so they
// work with *testing.T, *testing.B, and Ginkgo's GinkgoT().
//
// # Debugging
//
// Use [Event.String] and [EventsString] for human-readable representations
// useful in test failure messages:
//
//	t.Fatalf("unexpected events:\n%s", datastartest.EventsString(events))
package datastartest
