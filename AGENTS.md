# AGENTS.md — go-datastar

DataStar protocol library for Go. Patches as first-class values producing `sse.Event`. Built on go-sse. Single package (`datastar`), flat layout. The `datastartest/` subpackage is a separate Go module for consumer E2E testing.

## Module Structure

Three Go modules in a go.work workspace (rationale and rules: [ADR 002](docs/adr/002-multi-module-split.md)):

| Module | Path | Purpose | Dependencies |
| --- | --- | --- | --- |
| Root | `github.com/larsartmann/go-datastar` | Protocol library | go-sse, go-error-family |
| static | `github.com/larsartmann/go-datastar/static` | Embedded DataStar JS client bundle | zero (stdlib only) |
| datastartest | `github.com/larsartmann/go-datastar/datastartest` | Consumer E2E test helpers | go-datastar, go-sse |

Replace directives: root go.mod replaces `static => ./static`; datastartest
go.mod replaces `go-datastar => ..` and `static => ../static`. All resolve locally
for `GOWORK=off` builds (CI, Nix, consumers). Root no longer depends on
datastartest — the E2E test that used datastartest helpers was relocated to
`datastartest/e2e_test.go` to break a circular module dependency.

## Commands

```bash
# Workspace mode (default, uses go.work) — covers all three modules:
GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1
GOEXPERIMENT=jsonv2 go vet ./... ./datastartest/... ./static/...
GOEXPERIMENT=jsonv2 golangci-lint run ./... ./datastartest/... ./static/...

# Isolation mode (GOWORK=off, per-module) — verifies replace directives:
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # root only
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # datastartest (run from datastartest/)
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # static (run from static/)

# Error audit (all modules — erraudit v0.3.0 takes ONE directory per run, never package patterns):
for mod in . ./datastartest ./static; do
  (cd "$mod" && GOEXPERIMENT=jsonv2 erraudit . --type-aware --enforce-go-error-family --no-suppress)
done

# CI also enforces (run locally to pre-empt CI failures):
GOEXPERIMENT=jsonv2 go work sync  # go.work must not change after sync (idempotency)
grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod  # must find nothing (relative paths only)
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

See also [docs/modularization/README.md](docs/modularization/README.md) for the
proposal, execution plan, and ADRs behind the multi-module split.

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
| `store.go`               | `MemoryStore` — in-memory `sse.EventStore` ring buffer for reconnection replay             |
| `script_handler.go`      | ScriptHandler, ScriptTag, Version (HTTP serving of the `static` asset bundle)              |
| `static/`                | **Separate Go module** (zero deps). `//go:embed datastar.js`, `Bytes()`, `Version`         |
| `response.go`            | Response (fluent SSE builder), ErrorResponse, ErrorResponseFromError, NotificationResponse |
| `doc.go`                 | Package documentation (design rationale, quick start, error-system contract)               |
| `example/`               | Live-feed demo app (broadcaster + MemoryStore + ScriptHandler), zero client JS              |
| `example_test.go`        | Testable examples (Example functions with `// Output:` assertions)                         |
| `inbound_fuzz_test.go`   | Fuzz test for ReadSignals (10-seed corpus, regression-guarded)                             |
| `benchmark_test.go`      | Benchmarks for patch `Event()` generation + `FuzzMarshalSignalsRoundtrip`                  |
| `coverage_test.go`       | Option-application, construction error branches, stream-send failure paths                 |
| `errors_example_test.go` | Example functions showing all three error-handling patterns                |
| `e2e_test.go`             | `TestE2E_SSEHeaders` — transport header verification (go-sse owned). The full DataStar wire-format E2E test was relocated to `datastartest/e2e_test.go` |
| `module_boundary_test.go` | Regression guard: asserts root go.mod never requires datastartest (circular dependency prevention) |
| `datastartest/`           | **Separate Go module** (`go.work` workspace). Consumer E2E test helpers: SSE parsing, DataStar decoding, Collect, CollectPost, CollectN, CollectWithTimeout, FindElement, FindSignals, assertions, fuzz test. Also contains `e2e_test.go` (dogfood integration test) |

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

> **Note:** go-sse v0.5.0 added `JoinLines`/`KeyedLines` helpers for multi-line
> SSE data. go-datastar does NOT adopt them because `KeyedLines` normalizes
> CRLF to LF (items 6-7 split on `\n` only, matching upstream), and its key
> convention (`key + " "`) conflicts with go-datastar's trailing-space dataline
> constants (item 9). Revisit if upstream adopts CRLF normalization.

## Gotchas

- `go.work` is committed, but a **global** gitignore (`~/.config/git/ignore`)
  can still hide it on some machines. After touching `.gitignore` or creating
  module files, run `git check-ignore -v <file>` and `git ls-files <file>` —
  `git status` alone lies when a global ignore is in play.
- `dprint.json` exists in the repo root but is NOT wired into treefmt/flake —
  canonical formatting is treefmt (gofumpt/goimports/golines/nixfmt) via
  `nix flake check`.

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

datastartest codes: `datastartest.sse_scan_failed`,
`datastartest.signals_unmarshal_failed`.

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
5. **Layered composition with go-sse v0.5+.** Since go-sse v0.5.0, the
   transport also classifies its errors via go-error-family (codes like
   `sse.send_failed`). `wrapStreamError` wraps Send errors as
   `datastar.stream_send_failed` (Transient). Because `errorfamily.Classify`
   returns the outermost family, go-datastar's classification wins — correct,
   since Send failures are transient I/O errors. `errors.Is` traverses the
   chain, so callers matching go-sse codes work transparently through the wrap.

## What This Library Is NOT

No CQRS, no event bus, no domain opinions. It is a pure protocol layer. Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's EventBridge).

## E2E Testing for Consumers: `datastartest/`

The `datastartest` subpackage gives consumers reusable helpers for E2E testing
their DataStar handlers without hand-rolling SSE parsing or DataStar dataline
decoding. The full wire-format E2E test (`TestE2E_DataStarPatches`) lives in
`datastartest/e2e_test.go` — it dogfoods the helpers against real HTTP server
output. Root's `e2e_test.go` retains only `TestE2E_SSEHeaders` (transport-level
header checks owned by go-sse). This separation breaks what was otherwise a
circular module dependency: root must never require datastartest in its go.mod.

### API surface

| Export | Purpose |
| --- | --- |
| `Collect(t, handler, opts...)` | Spin up httptest.Server, GET, parse SSE, return decoded events |
| `CollectPost(t, handler, jsonBody, opts...)` | POST with JSON body, parse SSE, return decoded events |
| `CollectWithRequest(t, handler, method, body, ct, opts...)` | Custom method/body/content-type, parse SSE |
| `CollectN(t, handler, count, opts...)` | Read exactly N events (streaming handlers), then close |
| `CollectWithTimeout(t, handler, timeout, opts...)` | GET with deadline; returns events received before timeout |
| `WithPath(path)` | Target a route (query allowed) instead of "/" — mux-friendly |
| `WithDatastarSignals(json)` | Send `?datastar=` query param (GET/DELETE signal submission) |
| `WithLastEventID(id)` | Send Last-Event-ID header — replay/reconnection testing |
| `WithHeader(key, value)` | Any custom request header |
| `ReadEvents(io.Reader)` | Parse SSE wire format from any reader |
| `ReadNEvents(io.Reader, count)` | Streaming SSE reader; returns at N events or clean close |
| `MustReadEvents(t, io.Reader)` | ReadEvents with t.Fatal on error |
| `Event.IsElements()` / `.IsSignals()` / `.IsScript()` | Type predicates |
| `Event.Selector()` / `.Mode()` / `.Elements()` | Typed dataline accessors |
| `Event.ScriptContent()` | Strip `<script>` wrapper, return inner JS source |
| `Event.SignalsJSON()` / `.UnmarshalSignals(&v)` | Decode signals JSON |
| `Event.DataValue(key)` | Generic dataline lookup (escape hatch) |
| `Event.String()` / `EventsString(events)` | Human-readable debug representation |
| `FindElement(events, selector)` / `FindSignals(events)` | Search by selector/type |
| `FilterElements(events)` / `FilterSignals(events)` | Filter by event type |
| `RequireElements(t, evt, sel, mode, html)` | One-liner element assertion |
| `RequireElementsContains(t, evt, sel, mode, htmlSubstr)` | Substring match (scripts) |
| `RequireSignals(t, evt, json)` | Exact signals JSON assertion |
| `RequireSignalsContain(t, evt, key)` | Check signal key exists |
| `RequireScript(t, evt, js)` | Exact script-content assertion |
| `RequireEventID(t, evt, id)` | Event-ID assertion (replay tests) |
| `RequireEventCount(t, events, n)` | Event count assertion |
| `CodeSSEScanFailed` | Error code for SSE scanner I/O failures (`datastartest.sse_scan_failed`) |
| `CodeSignalsUnmarshalFailed` | Error code for signals JSON decode failures (`datastartest.signals_unmarshal_failed`) |

**All public helpers accept `testing.TB`** (not `*testing.T`), so they work with
`*testing.T`, `*testing.B`, and Ginkgo's `GinkgoT()`. Keep this invariant when
adding helpers.

### Consumer usage

```go
import (
    "github.com/larsartmann/go-datastar"
    "github.com/larsartmann/go-datastar/datastartest"
)

func TestFeedHandler(t *testing.T) {
    events := datastartest.Collect(t, myHandler)
    datastartest.RequireEventCount(t, events, 2)

    datastartest.RequireElements(t, events[0], "#feed", "append", "<div>hello</div>")

    var data struct{ Count int `json:"count"` }
    events[1].UnmarshalSignals(&data) // data.Count == 1
}
```
