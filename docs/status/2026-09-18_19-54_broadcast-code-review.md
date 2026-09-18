# Code Review: broadcast submodule

> Point-in-time review of `broadcast/` (full-code-review pass), 2026-09-18.
> Every public file visited; go-sse v0.6.0 internals cross-checked where the
> module depends on transport semantics. Verified by the full local gate.

## Scope

| File                  | Lines (review → now) | Role                                    |
| --------------------- | -------------------- | --------------------------------------- |
| `broadcaster.go`      | 163 → 183            | Patch fan-out hub + `http.Handler`      |
| `broadcaster_test.go` | 427 → 471            | 13 tests, race-covered                  |
| `doc.go`              | 21                   | Package docs                            |
| `README.md`           | 63                   | Module README                           |
| `go.mod` / `go.sum`   | —                    | Requires go-datastar v0.5.0, go-sse v0.6.0 |

## Verdict

Sound architecture: thin wrapper over `*sse.Broadcaster[sse.Event]`, patches as
values, hub shared via `Hub()`/`NewBroadcasterFromHub`. One real correctness
bug found (replay race window), two test bugs, and the canonical lint gate was
red on committed master (9 findings, all in this module). Everything fixed on
the spot; repo-wide CI-parity lint now exits 0.

## Fixed on the spot

| #  | Severity | Location                                  | Issue                                                                                                                               | Fix                                                                                                      |
| -- | -------- | ----------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| F1 | BUG      | `Broadcast`/`BroadcastEvent`              | Event appended to replay store AFTER hub fan-out: a reconnect racing between the two replays nothing and receives nothing — event permanently lost for that client | Append to store BEFORE fan-out; reconnecting mid-broadcast now replays the event instead of missing it    |
| F2 | Smell    | `BroadcastMany`                           | Looped per-patch `Broadcast`: N lock passes, batch interleavable by concurrent broadcasters                                         | Single-pass atomic `BroadcastMany` on the hub; all events appended first, then one fan-out pass           |
| F3 | Split brain | `Broadcast` vs `BroadcastEvent`        | Identical fan-out+store logic duplicated in two methods                                                                             | `Broadcast` now delegates to `BroadcastEvent(patch.Event())`                                              |
| F4 | Test bug | `TestBroadcasterBroadcastEvent`           | Never called `BroadcastEvent` — tested `Broadcast` under the wrong name                                                             | Rewritten: raw `sse.Event` in, `event: raw` / `data: payload` asserted on the wire                        |
| F5 | Test gap | `TestBroadcasterBroadcastMany`            | Zero delivery assertions (only SubscriberCount after disconnect)                                                                    | Asserts both patch families (`datastar-patch-signals`, `datastar-patch-elements`) reach the wire          |
| F6 | Flake risk | `readFirstResponse` helper              | Single `Read` of the first chunk — body split across TCP segments breaks assertions; channel-based, error-swallowing                | Replaced by `startReader`/`readBody`: drain-until-EOF (deterministic), typed result, `t.Fatalf` on test goroutine |
| F7 | Gate red | whole module                              | CI-parity `golangci-lint` failed on committed master with 9 findings (mnd, noctx ×2, varnamelen ×5, makezero) — unnoticed because nothing is a required check | Fixed all 9: `defaultHeartbeatInterval` const, `httptest.NewRequestWithContext`, `events`/`recorder`/`countMu` renames; makezero resolved by F6 rewrite |
| F8 | Docs     | `ServeHTTP`                               | Replay-failure path silently returned with no documented semantics                                                                  | Documented: connection ends; a well-behaved client reconnects with Last-Event-ID and retries              |
| F9 | Docs     | `Broadcast`/`BroadcastEvent`/`BroadcastMany` | Store/fan-out ordering invariant undocumented                                                                                   | Doc comments now state append-BEFORE-fan-out and the single-pass batch atomicity                          |

F1 detail: subscribe-before-replay already closed one half of the gap
(duplicates instead of loss after subscribe). The other half — events that left
the hub before entering the store — is closed by appending first. Duplicates
remain possible and are documented as harmless. No deterministic unit test can
pin F1 without injection seams between hub fan-out and store append; the
invariant is enforced by construction and documented in F9.

## Investigated, no action needed

- **Heartbeat goroutine concurrency**: `Stream.Send` and `Heartbeat` serialize
  on the stream mutex (go-sse `stream.go`); the goroutine exits on ctx-done or
  write error. No leak, no race. Race suite confirms.
- **gopls `infertypeargs` hint on `broadcaster.go:48` is a false positive**:
  dropping the explicit `[sse.Event]` from `sse.WithBufferSize` fails to
  compile (`cannot infer T`). Matches the AGENTS.md "LSP output can lie" rule;
  verified with the compiler before accepting.
- **erraudit `--no-suppress` finding** (`_ = stream.Close()`): documented
  tolerated pattern — `Stream.Close` always returns nil; errcheck excludes
  cover `*sse.Stream`. Canonical gate green.
- **Drop-on-full backpressure**: intentional go-sse semantics, documented in
  README and fan-out docs.

## Recommendations (routed to TODO_LIST, not fixed)

- **R1 — pluggable replay store (injection seam only)**: `store` is hardwired
  to `*datastar.MemoryStore`; `sse.Replay` already accepts any
  `sse.EventStore`. A `NewBroadcasterWithStore` would let consumers inject
  their own store for multi-instance deployments (the exact limitation
  `store.go` documents). No backend implementations ship in this repo —
  Redis/Postgres adapters stay consumer-side, per the module's non-goals.
- **R2 — constructor matrix gap**: buffer-size × replay are orthogonal axes
  with only 3 of 4 combos constructible. Functional options (or one combined
  constructor) would close it without breaking v0.x callers.
- **R3 — heartbeat configurability**: interval is now the
  `defaultHeartbeatInterval` const (was a magic number); expose as an option
  only if a consumer asks (YAGNI).
- **R4 — test file size**: 471 lines for 13 cohesive tests; under control, but
  split by theme (lifecycle / delivery / replay) if it keeps growing past
  ~600.

## Verification (all green at report time)

- `GOEXPERIMENT=jsonv2 go test ./... ./broadcast/... ./datastartest/... ./static/... -race -count=1` — ok
- broadcast `-race -count=5` — ok; `GOWORK=off` isolation — ok
- `golangci-lint v2.12.2 run ./... ./broadcast/... ./datastartest/... ./static/...` (fresh cache) — **0 issues**
- `go vet` (workspace) — clean; `gofmt -l` — clean
- `GOWORK=off go mod tidy -diff` — empty; `go mod verify` — all verified
- `erraudit` — sole finding is the documented tolerated `Close` blank-ignore
