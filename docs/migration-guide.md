# Migration guide: v0.2.0 → v0.3.0

v0.3.0 is the first release since v0.2.0 (2026-08-16). No exported API was
removed; the changes are additive plus toolchain requirements.

## Go toolchain: 1.26.7 required

All three modules now pin `go 1.26.7` (root, `static`, `datastartest`,
`go.work`, CI). Consumers must build with Go ≥ 1.26.7. This completes the
1.26.6 stdlib-CVE hardening (GO-2026-5972/6089/6090/6218) on the current
patch.

`GOEXPERIMENT=jsonv2` remains required for all go commands (transitive via
go-branded-id through go-sse).

## Dependency changes

| Module       | Change                                                             |
| ------------ | ------------------------------------------------------------------ |
| root         | go-sse v0.5.0 → **v0.5.1**                                         |
| datastartest | go-sse v0.5.1 + **go-sse/ssetest v0.2.0** (shared SSE test parser) |

## datastartest: what's new (all additive)

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

## Root: what's new

- **`Response.ReplaceURLQuerystring`** (+ `NewReplaceURLQuerystringPatch`) —
  upstream-parity query-string replacement.
- Live coverage badge (`coverage.yml`).
- Example app heartbeat keep-alive pattern (`example/`).

## Migrating from the official SDK?

The [README comparison table](../README.md) covers when each wins; the API
shapes map closely (`ServerSentEventGenerator` methods ↔ `Response` methods,
patch constructors ↔ generator calls). The structural difference: go-datastar
patches are **values** (`Patch.Event() sse.Event`) — store, filter, replay,
and broadcast them instead of writing through a generator.
