package broadcast

import (
	"time"

	datastar "github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// Store is the injection seam for custom replay stores. It pairs the append
// side used by [Broadcaster.Broadcast], [Broadcaster.BroadcastMany] and
// [Broadcaster.BroadcastEvent] with the replay side ([sse.EventStore]) used
// when a client reconnects with a Last-Event-ID.
//
// [*datastar.MemoryStore] implements it. Multi-instance deployments bring
// their own shared store (Redis, Postgres, ...) — this module deliberately
// ships no backends, per the root module's non-goals.
//
// Implementations must be safe for concurrent use. The store is owned by the
// caller: the broadcaster appends to and replays from it but never closes or
// resets it. Replay is keyed on numeric [sse.EventID]s (see
// [datastar.MemoryStore]); stores that receive ID-less events should assign
// their own monotonic IDs, or reconnection replay degenerates to
// "everything after zero".
type Store interface {
	sse.EventStore

	// Append stores an event for later replay. It is called BEFORE the
	// event is fan-out to subscribers, so a client reconnecting
	// mid-broadcast replays the event instead of missing it.
	Append(evt sse.Event)
}

// Option configures a [Broadcaster] at construction time. Options apply in
// order; where several options set the same knob, the last one wins.
type Option func(*config)

// config is the accumulated constructor state. Zero values keep the
// respective defaults: go-sse's buffer size, no replay store, and the
// default heartbeat interval.
type config struct {
	bufferSize        int
	store             Store
	heartbeatInterval time.Duration
}

// WithBufferSize sets the per-subscriber channel capacity. Slow clients
// whose buffer fills silently miss events, so size it for your worst
// broadcast burst. size <= 0 keeps the go-sse default.
func WithBufferSize(size int) Option {
	return func(c *config) {
		if size > 0 {
			c.bufferSize = size
		}
	}
}

// WithReplayCapacity enables reconnection replay on an in-memory ring buffer
// retaining the last capacity events. It is sugar for
// WithStore(datastar.NewMemoryStore(capacity)); capacity <= 0 uses
// [datastar.DefaultMemoryStoreCapacity].
func WithReplayCapacity(capacity int) Option {
	return func(c *config) {
		c.store = datastar.NewMemoryStore(capacity)
	}
}

// WithStore injects a custom replay store ([Store]). Use it to share replay
// state across instances (the single-process limitation
// [datastar.MemoryStore] documents) or to observe the event stream. Passing
// nil disables replay, matching [NewBroadcaster] with no replay option.
func WithStore(store Store) Option {
	return func(c *config) {
		c.store = store
	}
}

// WithHeartbeatInterval overrides the per-connection SSE comment-ping
// interval (the default keeps idle streams alive through reverse proxies
// and firewalls). interval <= 0 keeps the default.
func WithHeartbeatInterval(interval time.Duration) Option {
	return func(c *config) {
		if interval > 0 {
			c.heartbeatInterval = interval
		}
	}
}

// newConfig applies opts in order and returns the accumulated state.
func newConfig(opts ...Option) config {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
