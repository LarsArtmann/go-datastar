# go-datastar

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-datastar.svg)](https://pkg.go.dev/github.com/larsartmann/go-datastar)
[![CI](https://github.com/LarsArtmann/go-datastar/actions/workflows/ci.yml/badge.svg)](https://github.com/LarsArtmann/go-datastar/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-datastar)](https://goreportcard.com/report/github.com/larsartmann/go-datastar)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

DataStar protocol library for Go. Every patch is a **first-class value** that produces an [`sse.Event`](https://pkg.go.dev/github.com/larsartmann/go-sse#Event) — so you can store, queue, filter, replay, and broadcast them through go-sse's transport infrastructure.

Built on [go-sse](https://github.com/LarsArtmann/go-sse).

## What is this?

[DataStar](https://data-star.dev) is a hypermedia framework that uses Server-Sent Events to push DOM patches, signal updates, and script execution from server to browser. **go-datastar** provides the Go server-side protocol vocabulary.

The key design principle: **patches are values, not method calls**. Every patch implements `Patch interface { Event() sse.Event }`, so you can construct one without an open connection, hand it to a `Broadcaster[T]`, persist it in an `EventStore`, or filter it through a `SubscribeFilter`.

## Why not `starfederation/datastar-go`?

The upstream SDK couples patch construction to a live SSE connection (`PatchElements` is a method on `ServerSentEventGenerator`). Patches are not values — you cannot queue, filter, replay, or broadcast them. go-datastar fixes this with a one-method interface that unlocks composition with go-sse.

## Requirements

- **Go 1.26+**
- **`GOEXPERIMENT=jsonv2`** environment variable (required transitively via go-branded-id through go-sse)

```bash
# All go commands need this:
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./... -race -count=1
```

## Install

```bash
go get github.com/larsartmann/go-datastar
```

## Quick start

### Patches as values — broadcast to many connections

This is the core pattern that distinguishes go-datastar from the upstream SDK. Construct a patch without a connection, then broadcast it:

```go
broadcaster := sse.NewBroadcaster[sse.Event]()

patch := datastar.NewElementsPatch("<div>Update</div>",
    datastar.WithSelectorID("feed"),
    datastar.WithModePrepend(),
)

// Every subscriber receives the same sse.Event
broadcaster.Broadcast(patch.Event())
```

### Single-request response — fluent builder

For the common case of sending patches on a single HTTP connection, `Response` wraps a stream and provides fluent methods:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    stream := sse.NewStream(w, r)
    defer func() { _ = stream.Close() }()

    resp := datastar.NewResponse(stream)

    if err := resp.PatchElements("<div>Hello</div>", datastar.WithSelector("#feed")); err != nil {
        log.Printf("patch elements: %v", err)
        return
    }

    if err := resp.MarshalAndPatchSignals(map[string]any{"count": 1}); err != nil {
        log.Printf("patch signals: %v", err)
    }
}
```

### Serve the DataStar JS client

```go
mux.Handle("GET /datastar.js", datastar.ScriptHandler())
```

### Full working example

A complete live-feed application is in [`example/`](example/):

```bash
go run ./example/
# Open http://localhost:8765
```

## Core concept: the Patch interface

```go
type Patch interface {
    Event() sse.Event
}
```

Four types implement this interface. Everything else is a convenience constructor:

| Type                       | What it does                     |
| -------------------------- | -------------------------------- |
| `ElementsPatch`            | Merge HTML elements into the DOM |
| `SignalsPatch`             | Update reactive signals          |
| `ScriptPatch`              | Execute JavaScript on the client |
| `DispatchCustomEventPatch` | Dispatch a custom DOM event      |

## Patch constructors

### Elements

| Constructor                             | Returns                  | Notes                          |
| --------------------------------------- | ------------------------ | ------------------------------ |
| `NewElementsPatch(html, opts...)`       | `ElementsPatch`          | Core element patch             |
| `NewRemovePatch(selector)`              | `ElementsPatch`          | Remove element by CSS selector |
| `NewRemoveByIDPatch(id)`                | `ElementsPatch`          | Remove element by ID           |
| `ElementsFromTempl(component, opts...)` | `(ElementsPatch, error)` | Render a [Templ] component     |
| `ElementsFromGostar(renderer, opts...)` | `(ElementsPatch, error)` | Render a [GoStar] element      |

### Signals

| Constructor                            | Returns                 | Notes                              |
| -------------------------------------- | ----------------------- | ---------------------------------- |
| `NewSignalsPatch(v, opts...)`          | `(SignalsPatch, error)` | Marshal a Go value to signals JSON |
| `NewSignalsIfMissingPatch(v, opts...)` | `(SignalsPatch, error)` | Only set signals that don't exist  |

Pre-encoded JSON? Construct directly: `datastar.SignalsPatch{Signals: []byte("{\"count\":1}")}`.

### Script execution

| Constructor                                 | Returns                             | Notes                         |
| ------------------------------------------- | ----------------------------------- | ----------------------------- |
| `NewScriptPatch(js, opts...)`               | `ScriptPatch`                       | Core script patch             |
| `NewRedirectPatch(url)`                     | `ScriptPatch`                       | Redirect the browser          |
| `NewConsoleLogPatch(msg)`                   | `ScriptPatch`                       | `console.log` on the client   |
| `NewConsoleErrorPatch(err)`                 | `ScriptPatch`                       | `console.error` on the client |
| `NewReplaceURLPatch(url)`                   | `ScriptPatch`                       | `history.replaceState`        |
| `NewPrefetchPatch(urls...)`                 | `ScriptPatch`                       | Speculation-rules prefetch    |
| `NewDispatchCustomEventPatch(name, detail)` | `(DispatchCustomEventPatch, error)` | Dispatch a custom DOM event   |

### Printf-style variants

`WithSelectorf`, `NewRedirectfPatch`, `NewConsoleLogfPatch` — same behavior with `fmt.Sprintf` formatting.

## Options

### Elements

`WithSelector`, `WithSelectorID`, `WithSelectorf`, `WithMode`, `WithNamespace`, `WithViewTransitions`, `WithViewTransitionSelector`, `WithElementsEventID`, `WithElementsRetryDuration`

Mode sugar: `WithModeOuter`, `WithModeInner`, `WithModeRemove`, `WithModeReplace`, `WithModePrepend`, `WithModeAppend`, `WithModeBefore`, `WithModeAfter`

Namespace sugar: `WithNamespaceHTML`, `WithNamespaceSVG`, `WithNamespaceMathML`

### Signals

`WithOnlyIfMissing`, `WithSignalsEventID`, `WithSignalsRetryDuration`

### Script

`WithScriptAutoRemove`, `WithScriptAttributes`, `WithScriptAttributeKVs`, `WithScriptEventID`, `WithScriptRetryDuration`

### Custom events

`WithCustomEventSelector`, `WithCustomEventBubbles`, `WithCustomEventCancelable`, `WithCustomEventComposed`, `WithCustomEventEventID`

## Response builder

`Response` wraps an `sse.Stream` for single-connection patching. Every method returns `error`:

| Method                                    | Description                        |
| ----------------------------------------- | ---------------------------------- |
| `PatchElements(html, opts)`               | Send an `ElementsPatch`            |
| `PatchElementsTempl(c, opts)`             | Render + send a Templ component    |
| `PatchSignals(json, opts)`                | Send pre-encoded JSON signals      |
| `MarshalAndPatchSignals(v, opts)`         | Marshal + send a Go value          |
| `RemoveElement(selector)`                 | Remove element by selector         |
| `RemoveElementByID(id)`                   | Remove element by ID               |
| `ExecuteScript(js, opts)`                 | Send a `ScriptPatch`               |
| `Redirect(url, opts)`                     | Redirect the browser               |
| `ConsoleLog(msg, opts)`                   | `console.log` on the client        |
| `ConsoleError(err, opts)`                 | `console.error` on the client      |
| `DispatchCustomEvent(name, detail, opts)` | Dispatch a custom DOM event        |
| `ReplaceURL(url, opts)`                   | `history.replaceState`             |
| `Prefetch(urls...)`                       | Speculation-rules prefetch         |
| `ApplyPatches(patches...)`                | Send multiple patches in sequence  |
| `Send(evt)`                               | Send a raw `sse.Event`             |
| `Stream()`                                | Access the underlying `sse.Stream` |

Convenience constructors:

- `NewResponse(stream)` — wrap an existing stream
- `NewResponseFromHTTP(w, r)` — create stream + response in one call
- `ErrorResponse(stream, message, code)` — signals patch with error info
- `ErrorResponseFromError(stream, err)` — signals patch with errorfamily metadata extracted from a Go error (code, family, retryable, httpStatus)
- `NotificationResponse(stream, message, kind)` — signals patch with notification

## Inbound: reading signals

```go
var signals struct {
    Email string `json:"email"`
}
if err := datastar.ReadSignals(r, &signals); err != nil {
    log.Printf("read signals: %v", err)
}
```

- `ReadSignals(r, &target)` — extracts signals from `?datastar=` query param (GET/DELETE) or JSON body (all other methods). Returns nil if no signals present.
- `LastEventID(r)` — extracts the last event ID from the `Last-Event-ID` header or `lastEventId` query param (for SSE reconnection replay).

## HTTP helpers

DataStar action attribute strings for use in HTML:

```go
datastar.GetSSE("/api/feed")
datastar.PostSSE("/api/items/%d", id)
datastar.PutSSE("/api/settings")
datastar.PatchSSE("/api/merge")
datastar.DeleteSSE("/api/items/%d", id)
```

## Serving the JS client

| Function                     | Description                                                       |
| ---------------------------- | ----------------------------------------------------------------- |
| `ScriptHandler()`            | Serve the embedded DataStar JS (v1.0.2) with ETag + Cache-Control |
| `ScriptHandlerWith(js, ver)` | Serve a custom JS bundle                                          |
| `ScriptTag(path)`            | HTML `<script type="module">` tag string                          |
| `Version()`                  | Embedded JS client version string                                 |

## Event store (reconnection replay)

`MemoryStore` is an in-memory ring buffer implementing `sse.EventStore`. It keeps the last N events so reconnecting clients can replay missed patches:

```go
store := datastar.NewMemoryStore(128) // keep last 128 events

// In your producer:
store.Append(patch.Event())

// go-sse uses it for automatic replay on reconnection.
```

For multi-instance deployments, implement `sse.EventStore` against a shared backend (Redis, Postgres).

## Error handling

Every error returned by go-datastar is a classified [`*errorfamily.Error`](https://github.com/LarsArtmann/go-error-family) carrying a stable **code**, a behavioral **family**, and structured **context**.

### Three ways to handle errors

```go
// 1. By code (stable string):
if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed { ... }

// 2. By sentinel (errors.Is matches by code+family):
if errors.Is(err, datastar.ErrEventNameRequired) { ... }

// 3. By family (behavioral — retryable? whose fault?):
if errorfamily.Classify(err) == errorfamily.Transient { /* backoff + retry */ }
```

### Families

| Family        | When                                    | Retryable | HTTP Status |
| ------------- | --------------------------------------- | --------- | ----------- |
| Rejection     | Bad or missing caller input             | no        | 400         |
| Transient     | Temporary I/O failure reading body      | yes       | 503         |
| Orchestration | Internal render failure (templ, gostar) | no        | 500         |

### Error codes

| Code                                          | Family        | Retryable |
| --------------------------------------------- | ------------- | --------- |
| `datastar.templ_render_failed`                | Orchestration | no        |
| `datastar.gostar_render_failed`               | Orchestration | no        |
| `datastar.body_read_after_close`              | Rejection     | no        |
| `datastar.body_read_failed`                   | Transient     | yes       |
| `datastar.signals_unmarshal_failed`           | Rejection     | no        |
| `datastar.signals_marshal_failed`             | Rejection     | no        |
| `datastar.custom_event_detail_marshal_failed` | Rejection     | no        |
| `datastar.event_name_required`                | Rejection     | no        |
| `datastar.element_patch_mode_invalid`         | Rejection     | no        |
| `datastar.namespace_invalid`                  | Rejection     | no        |
| `datastar.stream_send_failed`                 | Transient     | yes       |

## Wire format parity

go-datastar reproduces the exact DataStar wire format expected by the DataStar JavaScript client. Mode `outer` is never emitted (default). Namespace `html` is never emitted (default). Retry is emitted only when it deviates from the 1000ms default. Data-line construction order, script wrapping, and signal splitting all match the upstream SDK.

## Companion libraries

- [go-sse](https://github.com/LarsArtmann/go-sse) — SSE transport (Stream, Broadcaster, EventStore, Replay, Heartbeat)
- [go-error-family](https://github.com/LarsArtmann/go-error-family) — structured error classification

## License

MIT

[Templ]: https://templ.guide/
[GoStar]: https://github.com/delaneyj/gostar
