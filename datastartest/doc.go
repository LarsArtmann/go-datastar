// Package datastartest provides helpers for E2E testing DataStar handlers.
//
// It solves the two problems that make DataStar handlers hard to test:
//
//  1. Parsing the SSE wire format (event:/data:/id:/retry: lines) into events.
//  2. Decoding DataStar datalines (selector/mode/elements/signals key-value
//     pairs) back into typed, assertable values.
//
// The library's own [e2e_test.go] in the parent package hand-rolls ~260 lines
// of private parsing code for the same purpose. This package exports that logic
// so consumers don't have to reinvent it.
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
// # Streaming handlers
//
// For handlers that keep the connection open (e.g., broadcasting through a
// [Broadcaster]), use [CollectN] to read exactly N events then close:
//
//	events := datastartest.CollectN(t, handler, 3)
//
// # Script patches
//
// Script patches (ExecuteScript, Redirect, ConsoleLog, etc.) produce
// patch-elements events with JS wrapped in <script> tags. Use
// [Event.ScriptContent] to extract the inner JavaScript source:
//
//	events := datastartest.Collect(t, handler)
//	js := events[0].ScriptContent() // "console.log('hello')"
package datastartest
