# go-datastar/broadcast

Optional connection-lifecycle module for
[go-datastar](https://github.com/LarsArtmann/go-datastar): fan-out, SSE
serving, reconnection replay, and cross-transport hub sharing for DataStar
patches.

The root go-datastar module is a pure protocol vocabulary — patches are
values, but it manages no connections. This module is the glue that turns
those values into a live SSE endpoint.

## Install

```bash
go get github.com/larsartmann/go-datastar/broadcast
```

## Quick start

```go
import (
    "net/http"

    datastar "github.com/larsartmann/go-datastar"
    "github.com/larsartmann/go-datastar/broadcast"
)

broadcaster := broadcast.NewBroadcasterWithReplay(128)
mux.Handle("GET /events", broadcaster)

broadcaster.Broadcast(datastar.NewElementsPatch("<div>Update</div>",
    datastar.WithSelectorID("feed"),
    datastar.WithModePrepend(),
))
```

## API

| Function                             | Description                                                                                                |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| `NewBroadcaster(opts...)`            | Fan-out SSE patches; options compose buffer, replay, store, heartbeat                                      |
| `NewBroadcasterWithBufferSize(n)`    | Sugar for `NewBroadcaster(WithBufferSize(n))`                                                              |
| `NewBroadcasterWithReplay(n)`        | Sugar for `NewBroadcaster(WithReplayCapacity(n))` — ring-buffer replay on reconnect (Last-Event-ID)        |
| `NewBroadcasterFromHub(hub)`         | Wrap an existing `*sse.Broadcaster[sse.Event]` hub                                                         |
| `Broadcaster.Hub()`                  | Access/share the embedded go-sse hub                                                                       |
| `Broadcaster.Broadcast(patch)`       | Send one patch to all clients                                                                              |
| `Broadcaster.BroadcastMany(...)`     | Send multiple patches                                                                                      |
| `Broadcaster.BroadcastEvent(evt)`    | Send a raw `sse.Event`                                                                                     |
| `storecontract.AssertReplayOrdering` | Prove a custom store replays verbatim, ordered, sliceable by Last-Event-ID (`broadcast/storecontract` pkg) |
| `storecontract.AssertIDAssignment`   | Prove a custom store assigns monotonic numeric IDs to ID-less events (`broadcast/storecontract` pkg)       |

### Options

Buffer size, replay, and heartbeat are orthogonal knobs on one constructor:

```go
b := broadcast.NewBroadcaster(
    broadcast.WithBufferSize(64),      // per-subscriber channel capacity
    broadcast.WithReplayCapacity(256), // in-memory ring-buffer replay
)
```

`WithStore(store)` injects a custom replay store (`broadcast.Store` =
`sse.EventStore` + `Append`) so multi-instance deployments can share replay
state in Redis/Postgres — the module ships no backends by design; the store
is owned and closed by you. `WithStore(nil)` disables replay, and a later
option overwrites an earlier one. `WithHeartbeatInterval(d)` overrides the
per-connection comment-ping interval (default 15s); `WithHeartbeatDisabled()`
turns the ping off entirely — an absolute switch that wins over any interval
regardless of order (use it when the transport below already keeps the
connection alive).

Hub-shared broadcasters take options too: `NewBroadcasterFromHub(hub, opts...)`
applies heartbeat and replay-store options normally; only the buffer size is
hub-owned (size the hub itself with `sse.WithBufferSize` at construction).

`Broadcaster` implements `http.Handler` and embeds `*sse.Broadcaster[sse.Event]`,
so `Subscribe`, `SubscribeFilter`, `Health`, `Shutdown`, `Close`, `OnSubscribe`,
and `OnUnsubscribe` are promoted. Each connection gets a heartbeat (15s by
default, `WithHeartbeatInterval` to change); replay
subscribes before replaying so events broadcast in the race window arrive as
idempotent duplicates instead of being lost.

Domain-event → patch mapping (an EventBridge) stays a consumer concern — see
the root module's non-goals. A working miniature lives in
[`example/domain-adapter/`](../example/domain-adapter/).

### Custom-store contract testing

The broadcaster appends events to your store verbatim and never assigns
IDs, so a custom store must keep replay keyed on monotonic numeric IDs —
the contract [`broadcast.Store`](options.go) documents. The
`broadcast/storecontract` package ships two `testing.TB` assertions so your
store's own tests can prove it (they work with `*testing.T`, `*testing.B`,
and Ginkgo's `GinkgoT()`):

```go
import (
    "testing"

    "github.com/larsartmann/go-datastar/broadcast/storecontract"
    "github.com/larsartmann/go-sse"
)

func TestMyRedisStoreReplay(t *testing.T) {
    store := newRedisStore() // your broadcast.Store implementation

    storecontract.AssertReplayOrdering(t, store, []sse.Event{
        {ID: sse.NewEventID("1"), Event: "feed", Data: "one"},
        {ID: sse.NewEventID("2"), Event: "feed", Data: "two"},
    })

    // The ID-less slice of the contract: your store assigns the IDs.
    storecontract.AssertIDAssignment(t, store, 3)
}
```

`AssertReplayOrdering` checks the Last-Event-ID behavior (verbatim replay,
ascending numeric order, exact suffix slicing); `AssertIDAssignment` checks
that ID-less events come back with assigned, strictly increasing numeric
IDs — the degeneration the contract warns about. Note the shipped
[`datastar.MemoryStore`](../store.go) deliberately passes only the first:
its contract is caller-assigned IDs (what `WithReplayCapacity` consumers
do), so it is the wrong reference for stores that receive ID-less events.

## License

MIT
