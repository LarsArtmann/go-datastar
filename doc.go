// Package datastar implements the DataStar protocol layer on top of go-sse.
//
// DataStar is a hypermedia framework that uses Server-Sent Events (SSE) to push
// DOM patches, signal updates, and script execution commands from the server to
// the browser. This library provides the Go server-side vocabulary for that
// protocol.
//
// # Patches are values
//
// The core design principle is that patches are first-class values, not method
// calls on a live connection. Every patch implements the Patch interface:
//
//	type Patch interface {
//	    Event() sse.Event
//	}
//
// This means patches can be stored, queued, filtered, replayed, and broadcast
// using go-sse's Broadcaster[T], EventStore, SubscribeFilter, and Shutdown
// infrastructure — none of which is possible with the upstream
// starfederation/datastar-go SDK, where patches are methods bound to a live SSE
// generator.
//
// # Quick start
//
//	import (
//	    "github.com/larsartmann/go-datastar"
//	    "github.com/larsartmann/go-sse"
//	)
//
//	// Broadcast patches to multiple connections
//	broadcaster := sse.NewBroadcaster[sse.Event]()
//
//	// Create a patch as a value
//	patch := datastar.NewElementsPatch("<div>Hello</div>",
//	    datastar.WithSelector("#feed"),
//	    datastar.WithMode(datastar.ElementPatchModeInner),
//	)
//
//	// Broadcast it — every subscriber receives the same sse.Event
//	broadcaster.Broadcast(patch.Event())
//
// # Wire format parity
//
// go-datastar reproduces the exact DataStar wire format expected by the
// DataStar JavaScript client. The data-line construction order, mode/namespace
// gating, retry logic, and script wrapping all match the upstream SDK behavior.
//
// # Modules
//
// The embedded DataStar JavaScript client lives in the separate zero-dependency
// module github.com/larsartmann/go-datastar/static (static.Bytes,
// static.Version), served ready-to-use via ScriptHandler. The E2E test helpers
// for consumers live in github.com/larsartmann/go-datastar/datastartest.
//
// # Classified errors
//
// Every error returned by this library is a classified *errorfamily.Error
// carrying a stable machine-readable code, a behavioral family (Rejection,
// Transient, Orchestration), and structured context. Consumers can match by
// code, sentinel, or family:
//
//	// By stable code:
//	if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed { ... }
//
//	// By sentinel (errors.Is matches by code+family):
//	if errors.Is(err, datastar.ErrEventNameRequired) { ... }
//
//	// By behavioral family (retryable? whose fault?):
//	if errorfamily.IsRetryable(err) { /* backoff + retry */ }
//
// See the Error System section in AGENTS.md for the full catalog.
//
// # Companion to go-sse
//
// go-datastar depends on go-sse for the SSE transport layer (Stream,
// Broadcaster, EventStore, Replay). go-sse owns the wire format; go-datastar
// owns the protocol vocabulary. This separation keeps each library focused and
// composable.
package datastar
