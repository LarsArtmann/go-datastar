# Domain Language — go-datastar

The ubiquitous language for the go-datastar protocol library. These terms shape
all code, documentation, and discussions. When a term appears in code, it means
exactly what it means here — no more, no less.

## Core Concepts

### Patch

A first-class value that produces an `sse.Event`. Every DataStar protocol
message is a patch. Patches are storable, filterable, replayable, and
broadcastable. This is the keystone abstraction: it decouples producing a
protocol message from delivering it.

```go
type Patch interface { Event() sse.Event }
```

### Signals

The client-side reactive state managed by the DataStar JavaScript client.
Signals are JSON objects. The server patches them via `SignalsPatch`; the client
sends them to the server via `ReadSignals` (inbound). Signals are the data
binding between server-side state changes and client-side reactivity.

### Dataline

A key-value pair in an SSE event's `data:` field. go-datastar emits datalines
with trailing-space keys: `"selector #feed"`, `"elements <div>"`,
`"signals {...}"`, `"script ..."`. Each dataline is one line in the SSE data
block. The key identifies the field; the space after the key separates it from
the value.

### Event (SSE)

A Server-Sent Events message: an `event:` type, one or more `data:` lines, an
optional `id:` and `retry:`. go-datastar patches produce `sse.Event` values that
the go-sse transport writes to the wire. The DataStar JS client interprets the
`event:` type to decide how to apply the datalines.

## Patch Types

### ElementsPatch

A patch that merges HTML elements into the DOM at a selector, using an
`ElementPatchMode` (append, prepend, replace, outer, inner, before, after,
remove). Emits `event: datastar-patch-elements`.

### SignalsPatch

A patch that updates the client's reactive signals. Emits
`event: datastar-patch-signals`. The signals JSON is split across multiple
`data: signals ...` datalines (one per line).

### ScriptPatch

A patch that executes JavaScript in the client's browser. Emits
`event: datastar-execute-script`. Always uses `selector: body` and
`mode: append` (matching upstream). Convenience wrappers: `RedirectPatch`,
`ConsoleLog`, `ConsoleError`, `DispatchCustomEvent`, `ReplaceURL`, `Prefetch`.

### RedirectPatch

A convenience `ScriptPatch` that navigates the browser to a URL. The redirect
is a client-side `window.location` change, not an HTTP redirect.

## Inbound

### ReadSignals

Extracts inbound signals from an HTTP request. GET and DELETE read from the
`?datastar=` query parameter; POST, PUT, and PATCH read from the JSON request
body. This is the server-side counterpart to the client's signals — the bridge
between what the browser knows and what the handler receives.

### LastEventID

The SSE reconnection cursor. The client sends it as the `Last-Event-ID` header
(or `lastEventId` query parameter) on reconnect. The server uses it to replay
missed events from an `EventStore`. Extracted from a request via
`datastar.LastEventID(r)`.

## Transport (go-sse)

### Stream

An SSE connection over an HTTP response. go-datastar writes patches to a
`Stream`; the stream frames them as SSE wire-format events. Owned by go-sse,
not go-datastar.

### Broadcaster

Fans out one `sse.Event` to every connected subscriber. The key enabler for
live feeds, dashboards, and multi-client updates. A patch stored as a value can
be broadcast to N connections — impossible with connection-bound methods.

### EventStore

Persists events for replay on reconnect. `MemoryStore` is the in-memory
implementation. When a client reconnects with a `LastEventID`, `sse.Replay`
re-sends everything after that ID from the store.

### Replay

Re-sending missed events to a reconnecting client. The server reads the
`LastEventID`, queries the `EventStore` for events after it, and sends them
before live updates resume.

## Response

A fluent builder for composing SSE output. `NewResponse(stream)` wraps a
go-sse stream; methods like `PatchElements`, `PatchSignals`, `ExecuteScript`
marshal and send patches in sequence. The `Response` is the ergonomic API over
raw patch construction — it handles marshaling and stream-send errors.

## Error System

### Family

A behavioral classification of an error: **Rejection** (bad caller input, not
retryable), **Transient** (temporary I/O failure, retryable), or
**Orchestration** (internal render failure, not retryable). Families drive
retry decisions: Transient → backoff and retry; others → fail fast.

### Code

A stable, machine-readable string identifying an error (e.g.,
`datastar.signals_marshal_failed`). Codes never change between releases;
callers match on them via `errorfamily.Code(err)`. Contrast with error
messages, which are human-readable and may change.

### Sentinel

A pre-declared error value matched via `errors.Is` (e.g.,
`ErrEventNameRequired`). Sentinels carry a code and family but no context;
`WithContext` returns a clone so shared sentinels never leak caller-specific
context.

## Wire-Format Constants

### EventType

The SSE `event:` field value. go-datastar emits three types:
`datastar-patch-elements`, `datastar-patch-signals`, `datastar-execute-script`.

### ElementPatchMode

How an `ElementsPatch` merges into the DOM: `outer` (default, never emitted),
`inner`, `replace`, `append`, `prepend`, `before`, `after`, `remove`.

### Namespace

The DOM namespace for an `ElementsPatch`: `html` (default, never emitted),
`svg`, `math`. Only non-default namespaces appear on the wire.

### DefaultRetryDuration

The SSE retry interval the DataStar client uses by default: 1000ms. Emitted on
the wire only when the retry value differs from this default.
