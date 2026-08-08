# AGENTS.md — go-datastar

DataStar protocol library for Go. Patches as first-class values producing `sse.Event`. Built on go-sse. Single package (`datastar`), flat layout.

## Commands

```bash
GOWORK=off GOEXPERIMENT=jsonv2 go test ./... -race -count=1   # tests
GOWORK=off GOEXPERIMENT=jsonv2 go vet ./...                    # vet
GOWORK=off GOEXPERIMENT=jsonv2 golangci-lint run ./...         # lint
erraudit ./... --type-aware --enforce-go-error-family --no-suppress  # error audit
```

**`GOEXPERIMENT=jsonv2` is required** (transitively via go-branded-id through go-sse).

Note: do **not** pass `--enforce-samber-oops` to erraudit. This is a library, and
the go-error-family contract is that libraries classify via go-error-family only
(never samber/oops). See [Error System](#error-system) below.

## Dependencies

- `github.com/larsartmann/go-sse` — SSE transport (Stream, Broadcaster, EventStore, Replay)
- `github.com/larsartmann/go-error-family` — structured error classification (Family, Code, Context). Direct dependency; every error this library returns is a classified `*errorfamily.Error`.
- Transitive: `go-branded-id` (via go-sse)

## Architecture

Three layers:

| Layer     | Repo                    | Role                                                                                  |
| --------- | ----------------------- | ------------------------------------------------------------------------------------- |
| Transport | go-sse                  | Stream, Broadcaster[T], Replay, EventID, Heartbeat, SubscribeFilter, Shutdown, Health |
| Protocol  | go-datastar (this repo) | Patch interface, ElementsPatch, SignalsPatch, ScriptPatch, RedirectPatch, etc.        |
| Domain    | cqrs-htmx/datastar      | EventBridge (CQRS event → Patch), thin re-exports                                     |

### Patch interface — the keystone

```go
type Patch interface { Event() sse.Event }
```

Every DataStar protocol message is a value that produces an `sse.Event`. This makes patches storable, filterable, replayable, and broadcastable — none of which the upstream SDK (`starfederation/datastar-go`) supports.

### File layout

| File                     | Role                                                                                       |
| ------------------------ | ------------------------------------------------------------------------------------------ |
| `patch.go`               | `Patch` interface                                                                          |
| `errors.go`              | Error catalog: stable codes, sentinel errors, family mapping                               |
| `constants.go`           | EventType, ElementPatchMode, Namespace, dataline keys, DefaultRetryDuration                |
| `elements.go`            | `ElementsPatch` struct + `Event()` + options                                               |
| `signals.go`             | `SignalsPatch` struct + `Event()` + marshal helpers                                        |
| `script.go`              | `ScriptPatch` struct + `Event()` + options                                                 |
| `script_convenience.go`  | Redirect, ConsoleLog/Error, DispatchCustomEvent, ReplaceURL, Prefetch                      |
| `sugar.go`               | Mode helpers, RemovePatch, validation, namespace helpers                                   |
| `adapters.go`            | ElementsFromTempl, ElementsFromGostar                                                      |
| `http.go`                | GetSSE/PostSSE/PutSSE/PatchSSE/DeleteSSE                                                   |
| `inbound.go`             | ReadSignals, LastEventID                                                                   |
| `script_handler.go`      | Embedded datastar.js, ScriptHandler, ScriptTag, Version                                    |
| `response.go`            | Response (fluent SSE builder), ErrorResponse, ErrorResponseFromError, NotificationResponse |
| `example_test.go`        | Testable examples (Example functions with `// Output:` assertions)                         |
| `inbound_fuzz_test.go`   | Fuzz test for ReadSignals (10-seed corpus, regression-guarded)                             |
| `coverage_test.go`       | Option-application, construction error branches, stream-send failure paths                 |
| `errors_example_test.go` | Example functions showing all three error-handling patterns                                |

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
12. HEAD requests to ScriptHandler return `200 OK` with headers but no message body (RFC 7231 §4.3.2)

## Error System

Every error returned by go-datastar is a classified `*errorfamily.Error` carrying
a stable machine-readable **code**, a behavioral **family**, and structured
**context**. The catalog lives in `errors.go`.

### Three strongly typed ways to handle errors

```go
// 1. By code (stable string, no string-matching on messages):
if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed { ... }

// 2. By sentinel (errors.Is matches by code+family, so context clones match too):
if errors.Is(err, datastar.ErrEventNameRequired) { ... }

// 3. By family (behavioral: retryable? whose fault?):
if errorfamily.IsRetryable(err) { /* backoff + retry */ }
```

### Family assignments

| Family        | When                                                                                                                            | Retryable |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------- | --------- |
| Rejection     | Bad/missing caller input (malformed JSON, empty name, unrecognized mode/namespace, body closed by misuse, unmarshallable value) | no        |
| Transient     | Temporary I/O failure reading the request body                                                                                  | yes       |
| Orchestration | Internal render failure producing HTML output (templ, gostar)                                                                   | no        |

### Codes

`datastar.templ_render_failed`, `datastar.gostar_render_failed`,
`datastar.body_read_after_close`, `datastar.body_read_failed`,
`datastar.signals_unmarshal_failed`, `datastar.signals_marshal_failed`,
`datastar.custom_event_detail_marshal_failed`, `datastar.event_name_required`,
`datastar.element_patch_mode_invalid`, `datastar.namespace_invalid`,
`datastar.stream_send_failed`.

### Sentinels

- `ErrBodyReadAfterClose` (wraps `http.ErrBodyReadAfterClose`, preserving the cause)
- `ErrEventNameRequired`

### Design decisions

1. **Library contract: go-error-family only, never samber/oops.** Per the
   go-error-family README, libraries classify but never presume the app's
   observability stack. Applications enrich with oops; this library does not.
2. **Return `error` interface, not `*errorfamily.Error`.** Idiomatic Go and
   consistent with go-sse (the direct dependency). Typed access is via
   `errorfamily.Code` / `errors.Is` / `errors.As`. erraudit's `generic_return`
   warnings on these signatures are accepted by design.
3. **Sentinels stay context-pristine.** `WithContext` returns a clone, so shared
   sentinels never leak caller-specific context.
4. **Context loss is a bug.** Wrapping errors include relevant in-scope values
   (HTTP method, input byte length, value type) so diagnosis needs no re-run.

## What This Library Is NOT

No CQRS, no event bus, no domain opinions. It is a pure protocol layer. Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's EventBridge).
