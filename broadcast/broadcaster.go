package broadcast

import (
	"net/http"
	"time"

	datastar "github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// defaultHeartbeatInterval is the per-connection SSE comment-ping interval.
// It keeps idle streams alive through reverse proxies and firewalls (15s,
// documented in the README).
const defaultHeartbeatInterval = 15 * time.Second

// Broadcaster fans out DataStar patches to all connected SSE clients. It embeds
// [sse.Broadcaster[sse.Event]] and implements [http.Handler].
//
// The embedded hub is the canonical shareable object: use [Broadcaster.Hub] to
// access it directly (Subscribe, SubscribeFilter, Health, Shutdown, Close,
// OnSubscribe, OnUnsubscribe are promoted) or to share one fan-out hub with
// another transport adapter via [NewBroadcasterFromHub].
//
// Mount it at your SSE endpoint:
//
//	broadcaster := broadcast.NewBroadcaster()
//	mux.Handle("GET /events", broadcaster)
//
// To push updates to all clients, call Broadcast:
//
//	broadcaster.Broadcast(datastar.NewElementsPatch(renderTodo(todo),
//		datastar.WithSelectorID("list")))
//
// For reconnection replay, use NewBroadcasterWithReplay. When a client
// reconnects with a Last-Event-ID header, missed events are replayed from an
// in-memory ring buffer before the live event stream resumes.
type Broadcaster struct {
	*sse.Broadcaster[sse.Event]

	store *datastar.MemoryStore
}

// NewBroadcaster creates a DataStar patch broadcaster with default settings
// and no replay support.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{Broadcaster: sse.NewBroadcaster[sse.Event]()}
}

// NewBroadcasterWithBufferSize creates a broadcaster with a custom subscriber
// buffer size and no replay support.
func NewBroadcasterWithBufferSize(size int) *Broadcaster {
	return &Broadcaster{
		Broadcaster: sse.NewBroadcaster[sse.Event](sse.WithBufferSize[sse.Event](size)),
	}
}

// NewBroadcasterWithReplay creates a broadcaster that retains the last capacity
// events in an in-memory ring buffer for reconnection replay. When a client
// reconnects with a Last-Event-ID header, missed events are replayed before the
// live stream resumes.
func NewBroadcasterWithReplay(capacity int) *Broadcaster {
	return &Broadcaster{
		Broadcaster: sse.NewBroadcaster[sse.Event](),
		store:       datastar.NewMemoryStore(capacity),
	}
}

// NewBroadcasterFromHub wraps an existing [*sse.Broadcaster] in a
// [*Broadcaster], enabling cross-transport fan-out hub sharing. Use this when
// you want two SSE flavors (or another transport adapter) to distribute from
// the same hub.
//
// The returned broadcaster has no replay store; [BroadcastEvent] on it reaches
// all hub subscribers, but nothing is retained for reconnection replay.
func NewBroadcasterFromHub(hub *sse.Broadcaster[sse.Event]) *Broadcaster {
	return &Broadcaster{Broadcaster: hub}
}

// Hub returns the embedded [*sse.Broadcaster] — the canonical fan-out hub.
// Use it to share one hub across transport adapters (via
// [NewBroadcasterFromHub]) or to access go-sse features directly
// (SubscribeFilter, Health, Shutdown, configurable buffer size).
func (b *Broadcaster) Hub() *sse.Broadcaster[sse.Event] {
	return b.Broadcaster
}

// Broadcast sends a patch to all connected clients. The patch's Event() is
// computed once and the resulting [sse.Event] is fan-out to all subscribers.
// With replay enabled, the event is appended to the store BEFORE the fan-out,
// so a client reconnecting mid-broadcast replays it instead of missing it.
// Slow clients whose channel buffer is full silently miss the event.
//
// Broadcast shadows the embedded hub's Broadcast(sse.Event) — to send a raw
// event, use [Broadcaster.BroadcastEvent].
func (b *Broadcaster) Broadcast(patch datastar.Patch) {
	b.BroadcastEvent(patch.Event())
}

// BroadcastMany sends multiple patches to all connected clients. All events
// are appended to the replay store first, then fan-out happens in a single
// locked hub pass, so the batch is atomic with respect to concurrent
// broadcasters. It shadows the embedded hub's BroadcastMany([]sse.Event); to
// send raw events, broadcast via [Broadcaster.Hub].
func (b *Broadcaster) BroadcastMany(patches ...datastar.Patch) {
	if len(patches) == 0 {
		return
	}

	evts := make([]sse.Event, 0, len(patches))
	for _, p := range patches {
		evts = append(evts, p.Event())
	}

	if b.store != nil {
		for _, evt := range evts {
			b.store.Append(evt)
		}
	}

	b.Broadcaster.BroadcastMany(evts...)
}

// BroadcastEvent sends a raw [sse.Event] to all connected clients. With replay
// enabled, the event is appended to the store BEFORE the fan-out, so a client
// reconnecting mid-broadcast replays it instead of missing it.
func (b *Broadcaster) BroadcastEvent(evt sse.Event) {
	if b.store != nil {
		b.store.Append(evt)
	}

	b.Broadcaster.Broadcast(evt)
}

// SubscriberCount returns the number of currently connected SSE clients.
func (b *Broadcaster) SubscriberCount() int {
	return b.Health().SubscriberCount
}

// ServeHTTP handles a DataStar SSE connection. It creates an [sse.Stream],
// subscribes to the broadcaster FIRST, then replays missed events from the
// store (if replay is enabled and the client sends a Last-Event-ID), and
// forwards events to the client until the request is cancelled or the
// connection breaks.
//
// Subscribe-before-replay is deliberate: events broadcast between the replay
// snapshot and the subscribe would otherwise be silently missed. With the
// subscription established first, those events race into the channel and are
// delivered as (harmless, idempotent) duplicates alongside the replayed ones.
// A replay failure (store error or client write error) ends the connection; a
// well-behaved client reconnects with its Last-Event-ID and retries.
func (b *Broadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stream := sse.NewStream(w, r)
	defer func() { _ = stream.Close() }()

	events := b.Subscribe()
	defer b.Unsubscribe(events)

	if b.store != nil {
		if lastID := stream.LastEventID(); !lastID.IsZero() {
			if _, err := sse.Replay(stream, b.store, lastID); err != nil {
				return
			}
		}
	}

	go stream.Heartbeat(r.Context(), defaultHeartbeatInterval)

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}

			if err := stream.Send(evt); err != nil {
				return
			}
		}
	}
}
