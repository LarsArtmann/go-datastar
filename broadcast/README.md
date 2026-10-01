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

| Function                          | Description                                               |
| --------------------------------- | --------------------------------------------------------- |
| `NewBroadcaster(opts...)`         | Fan-out SSE patches; options compose buffer, replay, store, heartbeat |
| `NewBroadcasterWithBufferSize(n)` | Sugar for `NewBroadcaster(WithBufferSize(n))`             |
| `NewBroadcasterWithReplay(n)`     | Sugar for `NewBroadcaster(WithReplayCapacity(n))` — ring-buffer replay on reconnect (Last-Event-ID) |
| `NewBroadcasterFromHub(hub)`      | Wrap an existing `*sse.Broadcaster[sse.Event]` hub        |
| `Broadcaster.Hub()`               | Access/share the embedded go-sse hub                      |
| `Broadcaster.Broadcast(patch)`    | Send one patch to all clients                             |
| `Broadcaster.BroadcastMany(...)`  | Send multiple patches                                     |
| `Broadcaster.BroadcastEvent(evt)` | Send a raw `sse.Event`                                    |

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
per-connection comment-ping interval (default 15s).

`Broadcaster` implements `http.Handler` and embeds `*sse.Broadcaster[sse.Event]`,
so `Subscribe`, `SubscribeFilter`, `Health`, `Shutdown`, `Close`, `OnSubscribe`,
and `OnUnsubscribe` are promoted. Each connection gets a heartbeat (15s by
default, `WithHeartbeatInterval` to change); replay
subscribes before replaying so events broadcast in the race window arrive as
idempotent duplicates instead of being lost.

Domain-event → patch mapping (an EventBridge) stays a consumer concern — see
the root module's non-goals. A working miniature lives in
[`example/domain-adapter/`](../example/domain-adapter/).

## License

MIT
