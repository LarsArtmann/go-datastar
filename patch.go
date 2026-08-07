package datastar

import "github.com/larsartmann/go-sse"

// Patch is the core interface of go-datastar. Every DataStar protocol message
// (element patches, signal patches, script execution, redirects, etc.)
// implements this interface.
//
// The key design principle: patches are first-class VALUES, not method calls on
// a live connection. A Patch can be constructed, stored in a slice, filtered by
// a predicate, replayed from an EventStore, and broadcast through a
// [sse.Broadcaster] — all without an open HTTP connection.
//
// Call [Patch.Event] to produce the final [sse.Event] that go-sse serializes to
// the wire:
//
//	patch := datastar.NewElementsPatch("<div>hi</div>", datastar.WithSelector("#feed"))
//	broadcaster := sse.NewBroadcaster[sse.Event]()
//	broadcaster.Broadcast(patch.Event())
//
// Or broadcast patches directly for typed filtering:
//
//	patchCaster := sse.NewBroadcaster[datastar.Patch]()
//	patchCaster.Broadcast(patch)
type Patch interface {
	// Event returns the SSE wire-format event for this patch. The returned
	// [sse.Event] contains the event type, data lines, optional event ID, and
	// optional retry duration — everything go-sse needs to serialize the patch.
	Event() sse.Event
}
