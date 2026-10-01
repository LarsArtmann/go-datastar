# Migration guide

Two hops are documented: [v0.2.0 → v0.3.0](#v020--v030) and
[v0.5.0 → v0.6.x](#v050--v06x). No exported API was removed in either;
all changes are additive or toolchain-level.

## Go toolchain (current truth)

**Go ≥ 1.27.1 is required** (floor raised in v0.6.1 across go.mod ×3,
go.work, CI, and the flake). The 1.26.7-era directives could not build under
a 1.27 toolchain (`encoding/json/v2` language-version gate).

**`GOEXPERIMENT=jsonv2` is no longer needed** — v0.6.1 un-gated
`encoding/json/v2`, so the experiment flag survives only as a harmless no-op
in this repo's own CI environment. If your build still sets it, you can drop
it.

Historically: v0.3.0 pinned `go 1.26.7` (stdlib-CVE hardening:
GO-2026-5972/6089/6090/6218) and required `GOEXPERIMENT=jsonv2`; both claims
were true then and false for every consumer on v0.6.1+.

## v0.2.0 → v0.3.0

### Dependency changes

| Module       | Change                                                             |
| ------------ | ------------------------------------------------------------------ |
| root         | go-sse v0.5.0 → **v0.5.1**                                         |
| datastartest | go-sse v0.5.1 + **go-sse/ssetest v0.2.0** (shared SSE test parser) |

### datastartest: what's new (all additive)

- **Request options on every `Collect*` helper** — `WithPath` (mux routes,
  query strings), `WithHeader`, `WithLastEventID` (reconnection replay),
  `WithDatastarSignals` (GET/DELETE signal submission). Previously every
  helper hard-requested `GET /`.
- **`RequireScript`, `RequireEventID`** assertions.
- **All helpers accept `testing.TB`** — works with `*testing.B` and Ginkgo's
  `GinkgoT()`; existing `*testing.T` callers unchanged.
- **SSE parser now delegates to go-sse/ssetest** — one parser implementation,
  WPT-conformant (lone CR, BOM, sticky ids/retry, EOF handling). If you
  relied on the old parser's non-conformant edge behavior, this is the one
  behavioral change to check.

### Root: what's new

- **`Response.ReplaceURLQuerystring`** (+ `NewReplaceURLQuerystringPatch`) —
  upstream-parity query-string replacement.
- Live coverage badge (`coverage.yml`).
- Example app heartbeat keep-alive pattern (`example/`).

## v0.5.0 → v0.6.x

### New optional `broadcast/` submodule (v0.6.0)

`github.com/larsartmann/go-datastar/broadcast` is the connection-lifecycle
layer the root module deliberately omits:

- `Broadcaster` — an `http.Handler` embedding `*sse.Broadcaster[sse.Event]`
- patch-level fan-out: `Broadcast`/`BroadcastMany`/`BroadcastEvent`
- reconnection replay: `NewBroadcasterWithReplay` (ring-buffer `MemoryStore`
  + `Last-Event-ID`, subscribe-before-replay ordering)
- a 15s per-connection heartbeat
- cross-transport hub sharing: `Hub()`/`NewBroadcasterFromHub`

Adopting it is opt-in: `go get
github.com/larsartmann/go-datastar/broadcast` and mount the `Broadcaster` as
a handler. Events are appended to the replay store **before** fan-out (a
v0.6.0 fix closing a reconnect race window where an event could be missed
permanently), so reconnecting clients may see documented, harmless
duplicates instead of gaps. Domain-event → patch mapping (EventBridge)
remains a consumer concern — see `example/domain-adapter` for the pattern.

### datastartest: helper tranche 2 (v0.6.0)

- `RequireNotScript` — pin handlers that must NOT respond with JavaScript.
- `FindScript` — first script patch, `(Event, bool)` like `FindSignals`.
- `FindAllElements` — every elements patch for a selector, in stream order.
- `EventToSelectorMap` — O(1) selector lookup; last patch wins (final DOM
  state).
- `CollectPostWithTimeout` / `CollectWithRequestWithTimeout` — deadline-bound
  variants of the non-GET collectors; events received before the deadline are
  returned, receiving none fails the test.

### Toolchain and dependencies (v0.6.1)

- Go floor **1.27.1** (see the toolchain section above).
- go-sse bumped to v0.6.1 (root, broadcast, datastartest) and
  go-sse/ssetest to v0.4.0 (datastartest); `go-sse/sseparse` v0.1.0 enters
  the graph transitively. datastartest's scan-error re-wrap preserves the
  underlying cause — no consumer action needed.

### Repo-side (no consumer action)

- The nix gate was promoted from `continue-on-error` to a real check — a red
  nix run counts as red master for this repo's CI only.

## Migrating from the official SDK?

The [README comparison table](../README.md) covers when each wins; the API
shapes map closely (`ServerSentEventGenerator` methods ↔ `Response` methods,
patch constructors ↔ generator calls). The structural difference: go-datastar
patches are **values** (`Patch.Event() sse.Event`) — store, filter, replay,
and broadcast them instead of writing through a generator.
