# ADR 001: Architecture — go-datastar, go-sse, and the DataStar SDK

Date: 2026-08-07

## Context

go-datastar sits between two dependencies with distinct responsibilities:

- **go-sse** (`github.com/larsartmann/go-sse`) — a transport library: SSE wire
  format, single-connection `Stream`, fan-out `Broadcaster[T]`, `EventStore` +
  `Replay`, heartbeat, `SubscribeFilter`. No DataStar-specific types.
- **DataStar SDK** (`github.com/starfederation/datastar-go`) — the reference Go
  implementation from the DataStar project. Defines the wire protocol (event
  types, data-line keys, patch modes) and provides `ServerSentEventGenerator`
  with `PatchElements`, `PatchSignals`, `ExecuteScript`, etc.

go-datastar is **not** a fork of the SDK. It is an independent implementation
that produces the same wire format, built on top of go-sse's transport
primitives rather than the SDK's own SSE writer.

## Decision

### Layered architecture

```
┌─────────────────────────────────────┐
│  Consumer (e.g. cqrs-htmx/datastar) │  domain logic, CQRS wiring
├─────────────────────────────────────┤
│  go-datastar                        │  DataStar protocol: Patch types,
│                                     │  Response API, MemoryStore
├─────────────────────────────────────┤
│  go-sse                             │  SSE transport: Stream, Broadcaster,
│                                     │  EventStore, Replay, Heartbeat
├─────────────────────────────────────┤
│  net/http                           │  Go standard library
└─────────────────────────────────────┘
```

Each layer has a single responsibility and no upward dependencies.

### Patch as the keystone type

```go
type Patch interface { Event() sse.Event }
```

Every DataStar instruction (elements patch, signals patch, script execution,
redirect) is a first-class value implementing `Patch`. The `Event()` method is
called once; the resulting `sse.Event` is fanned out, filtered, replayed, or
sent directly.

This means:

- **Fan-out:** `Broadcaster[Patch]` broadcasts typed patches, not raw events.
- **Filtering:** `SubscribeFilter` predicates can inspect the patch type, not
  just the event name.
- **Replay:** `MemoryStore` stores `sse.Event`s; patches produce deterministic
  events, so replay is consistent.
- **Testing:** Patches are values — construct one, call `Event()`, assert on
  the wire output. No HTTP server needed for unit tests.

### Why not depend on the SDK directly?

1. **The SDK reinvents SSE transport.** Its `ServerSentEventGenerator` writes
   SSE frames directly to `http.ResponseWriter`. go-sse already provides a
   battle-tested `Stream` with mutex-guarded writes, heartbeat, disconnect
   hooks, and flush management. Duplicating this is wasteful.

2. **The SDK has no fan-out or replay story.** DataStar applications need
   fan-out (one broadcast → N connected clients) and reconnection replay.
   go-sse's `Broadcaster[T]` and `EventStore` provide these. Using the SDK
   would mean building fan-out from scratch.

3. **The SDK's API is builder-style, not value-style.** The SDK uses
   `sse.PatchElements(html, opts...)` as a method on a generator object.
   go-datastar uses `NewElementsPatch(html, opts...)` returning a value. Values
   can be stored, composed, filtered, and tested independently of any I/O.

### Wire-format parity with the SDK

go-datastar produces byte-identical SSE output to the SDK for every patch type.
This is verified by the E2E HTTP round-trip test (`e2e_test.go`) and by direct
comparison with the SDK source:

| Patch type    | Event type                | Key behavior                                     |
| ------------- | ------------------------- | ------------------------------------------------ |
| ElementsPatch | `datastar-patch-elements` | `data: selector`, `data: mode`, `data: elements` |
| SignalsPatch  | `datastar-patch-signals`  | `data: signals <json>`                           |
| ScriptPatch   | `datastar-patch-elements` | Wraps in `<script>`, appends to `body`           |
| RedirectPatch | `datastar-patch-elements` | `window.location.href` script                    |

**Critical:** `ScriptPatch` does **not** emit a `datastar-execute-script`
event. The SDK itself uses `PatchElements` to send `<script>` elements (see SDK
`execute-script.go:71`). The only event types in the DataStar protocol are
`datastar-patch-elements` and `datastar-patch-signals`.

## Consequences

- **Consumers need both go-datastar and go-sse.** go-datastar re-exports the
  types most consumers need (`sse.Event`, `sse.Stream` via `Response`), but
  advanced features (broadcaster, replay, filtering) require importing go-sse
  directly.
- **Wire-format drift is a risk.** If the DataStar SDK changes its wire format,
  go-datastar must follow. The E2E test and the SDK comparison comments in
  `script.go` and `elements.go` document the expected format.
- **No SDK version coupling.** go-datastar does not import the SDK. The SDK is
  only used as a reference during development. go-datastar's `go.mod` has no
  dependency on `starfederation/datastar-go`.
