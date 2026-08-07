# AGENTS.md — go-datastar

DataStar protocol library for Go. Patches as first-class values producing `sse.Event`. Built on go-sse. Single package (`datastar`), flat layout.

## Commands

```bash
GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # tests
GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...                    # vet
GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...         # lint
```

**`GOEXPERIMENT=jsonv2` is required** (transitively via go-branded-id through go-sse).

## Dependencies

- `github.com/larsartmann/go-sse` — SSE transport (Stream, Broadcaster, EventStore, Replay)
- Transitive: `go-branded-id`, `go-error-family` (via go-sse)

## Architecture

Three layers:

| Layer | Repo | Role |
|-------|------|------|
| Transport | go-sse | Stream, Broadcaster[T], Replay, EventID, Heartbeat, SubscribeFilter, Shutdown, Health |
| Protocol | go-datastar (this repo) | Patch interface, ElementsPatch, SignalsPatch, ScriptPatch, RedirectPatch, etc. |
| Domain | cqrs-htmx/datastar | EventBridge (CQRS event → Patch), thin re-exports |

### Patch interface — the keystone

```go
type Patch interface { Event() sse.Event }
```

Every DataStar protocol message is a value that produces an `sse.Event`. This makes patches storable, filterable, replayable, and broadcastable — none of which the upstream SDK (`starfederation/datastar-go`) supports.

### File layout

| File | Role |
|------|------|
| `patch.go` | `Patch` interface |
| `constants.go` | EventType, ElementPatchMode, Namespace, dataline keys, DefaultRetryDuration |
| `elements.go` | `ElementsPatch` struct + `Event()` + options |
| `signals.go` | `SignalsPatch` struct + `Event()` + marshal helpers |
| `script.go` | `ScriptPatch` struct + `Event()` + options |
| `script_convenience.go` | Redirect, ConsoleLog/Error, DispatchCustomEvent, ReplaceURL, Prefetch |
| `sugar.go` | Mode helpers, RemovePatch, validation, namespace helpers |
| `adapters.go` | ElementsFromTempl, ElementsFromGostar |
| `http.go` | GetSSE/PostSSE/PutSSE/PatchSSE/DeleteSSE |
| `inbound.go` | ReadSignals, LastEventID |
| `script_handler.go` | Embedded datastar.js, ScriptHandler, ScriptTag, Version |
| `response.go` | Response (fluent SSE builder), ErrorResponse, NotificationResponse |

## Wire-Format Parity Requirements

These behaviors reproduce the upstream SDK exactly:

1. Mode `outer` is never emitted (default)
2. Namespace `html` is never emitted (default)
3. Retry emitted when `> 0 && != DefaultRetryDuration (1000ms)`
4. AutoRemove `*bool`: nil and true both add `data-effect="el.remove()"`
5. ExecuteScript always uses `selector: body` + `mode: append`
6. Elements split on `\n`; each line gets `data: elements ...`
7. Signals split on `\n`; every line emitted unconditionally
8. ReadSignals: GET/DELETE from `?datastar=` query; others from JSON body
9. Dataline keys have trailing space: `"selector "`, `"elements "`, etc.
10. ConsoleLog/Error use `%q` for JS string quoting
11. DispatchCustomEvent defaults: bubbles/cancelable/composed=true, selector=document

## What This Library Is NOT

No CQRS, no event bus, no domain opinions. It is a pure protocol layer. Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's EventBridge).
