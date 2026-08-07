# go-datastar

DataStar protocol library for Go — patches as first-class values built on [go-sse](https://github.com/LarsArtmann/go-sse).

## What is this?

[DataStar](https://data-star.dev) is a hypermedia framework that uses Server-Sent Events to push DOM patches, signal updates, and script execution from server to browser. **go-datastar** provides the Go server-side protocol vocabulary.

The key design principle: **patches are values, not method calls**. Every patch implements `Patch interface { Event() sse.Event }`, so you can store, queue, filter, replay, and broadcast them using go-sse's `Broadcaster[T]`, `EventStore`, `SubscribeFilter`, and `Shutdown` infrastructure.

## Why not `starfederation/datastar-go`?

The upstream SDK couples patch construction to a live SSE connection (`PatchElements` is a method on `ServerSentEventGenerator`). Patches are not values — you cannot queue, filter, replay, or broadcast them. go-datastar fixes this with a one-method interface that unlocks composition with go-sse.

## Install

```bash
go get github.com/larsartmann/go-datastar
```

## Quick start

### Single-request response

```go
func handler(w http.ResponseWriter, r *http.Request) {
    stream := sse.NewStream(w, r)
    defer stream.Close()

    resp := datastar.NewResponse(stream)
    resp.PatchElements("<div>Hello</div>", datastar.WithSelector("#feed"))
    resp.PatchSignals(map[string]any{"count": 1})
}
```

### Broadcast to multiple connections

```go
broadcaster := sse.NewBroadcaster[sse.Event]()

// Create a patch as a value
patch := datastar.NewElementsPatch("<div>Update</div>",
    datastar.WithSelector("#feed"),
    datastar.WithMode(datastar.ElementPatchModeInner),
)

// Broadcast — every subscriber receives the same sse.Event
broadcaster.Broadcast(patch.Event())
```

### Serve the DataStar JS client

```go
mux.Handle("GET /datastar.js", datastar.ScriptHandler())
```

## API overview

### Patches

| Type                       | Constructor                                 | Description                      |
| -------------------------- | ------------------------------------------- | -------------------------------- |
| `ElementsPatch`            | `NewElementsPatch(html, opts...)`           | Patch HTML elements into the DOM |
| `SignalsPatch`             | `NewSignalsPatch(v, opts...)`               | Patch reactive signals           |
| `ScriptPatch`              | `NewScriptPatch(js, opts...)`               | Execute JavaScript               |
| `RedirectPatch`            | `NewRedirectPatch(url)`                     | Redirect the browser             |
| `ConsoleLogPatch`          | `NewConsoleLogPatch(msg)`                   | console.log on the client        |
| `DispatchCustomEventPatch` | `NewDispatchCustomEventPatch(name, detail)` | Dispatch a custom DOM event      |
| `RemovePatch`              | `NewRemovePatch(selector)`                  | Remove an element from the DOM   |

### Options

Elements: `WithSelector`, `WithMode`, `WithNamespace`, `WithViewTransitions`, `WithElementsEventID`, `WithElementsRetryDuration`

Signals: `WithOnlyIfMissing`, `WithSignalsEventID`, `WithSignalsRetryDuration`

Script: `WithScriptAutoRemove`, `WithScriptAttributes`, `WithScriptEventID`

### Inbound

- `ReadSignals(r, &signals)` — extract signals from GET query param or POST body
- `LastEventID(r)` — extract last event ID from header or query param

### HTTP helpers

- `GetSSE`, `PostSSE`, `PutSSE`, `PatchSSE`, `DeleteSSE` — DataStar action attribute strings
- `ScriptHandler()` — serve the embedded DataStar JS client

## Wire format parity

go-datastar reproduces the exact DataStar wire format expected by the DataStar JavaScript client. All data-line construction order, mode/namespace gating, retry logic, and script wrapping match the upstream SDK behavior.

## Companion libraries

- [go-sse](https://github.com/LarsArtmann/go-sse) — SSE transport (Stream, Broadcaster, Replay, Heartbeat)
- [go-error-family](https://github.com/LarsArtmann/go-error-family) — structured error wrapping
- [go-branded-id](https://github.com/LarsArtmann/go-branded-id) — phantom-type branded IDs

## License

MIT
