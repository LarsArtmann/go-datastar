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
| `NewBroadcaster()`                | Fan-out SSE patches to all clients, no replay             |
| `NewBroadcasterWithBufferSize(n)` | Fan-out with a custom subscriber buffer size              |
| `NewBroadcasterWithReplay(n)`     | Fan-out + ring-buffer replay on reconnect (Last-Event-ID) |
| `NewBroadcasterFromHub(hub)`      | Wrap an existing `*sse.Broadcaster[sse.Event]` hub        |
| `Broadcaster.Hub()`               | Access/share the embedded go-sse hub                      |
| `Broadcaster.Broadcast(patch)`    | Send one patch to all clients                             |
| `Broadcaster.BroadcastMany(...)`  | Send multiple patches                                     |
| `Broadcaster.BroadcastEvent(evt)` | Send a raw `sse.Event`                                    |

`Broadcaster` implements `http.Handler` and embeds `*sse.Broadcaster[sse.Event]`,
so `Subscribe`, `SubscribeFilter`, `Health`, `Shutdown`, `Close`, `OnSubscribe`,
and `OnUnsubscribe` are promoted. Each connection gets a 15s heartbeat; replay
subscribes before replaying so events broadcast in the race window arrive as
idempotent duplicates instead of being lost.

Domain-event → patch mapping (an EventBridge) stays a consumer concern — see
the root module's non-goals. A working miniature lives in
[`example/domain-adapter/`](../example/domain-adapter/).

## License

MIT
