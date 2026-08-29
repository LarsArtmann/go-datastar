# ADR 002: Multi-module split with mutual replaces

Date: 2026-08-16

## Context

go-datastar is not one Go module but three, living in one repository with a
single `go.work`:

| Module       | Path                                              | Purpose                     | Dependencies            |
| ------------ | ------------------------------------------------- | --------------------------- | ----------------------- |
| root         | `github.com/larsartmann/go-datastar`              | DataStar protocol library   | go-sse, go-error-family |
| static       | `github.com/larsartmann/go-datastar/static`       | Embedded DataStar JS client | none (stdlib only)      |
| datastartest | `github.com/larsartmann/go-datastar/datastartest` | Consumer E2E test helpers   | root, go-sse            |

Before 2026-08-10 the root module required `datastartest` for a single E2E
test file while `datastartest` required root for production code — a circular
module dependency and a test-dep leak: every consumer `go get`-ing root also
pulled in the test-helper module. The
[modularization review](../modularization/2026-08-10_PROPOSAL.html) diagnosed
this (findings FM#1, FM#3) and prescribed relocating `TestE2E_DataStarPatches`
into `datastartest/e2e_test.go`.

## Decision

### Three modules, strict DAG, no merges

```
static ──── (leaf, zero dependencies)
   ▲
   │ production (script_handler.go embeds the JS bundle)
root
   ▲
   │ production (event.go, assert.go, filter.go use datastar types)
datastartest
```

`static → root → datastartest` is a strict DAG. No upward dependencies, no
cycles. The boundaries were judged to be at the right depth: no further
splits (the root `datastar` package is a single cohesive protocol package)
and no merges (static must stay dependency-free; datastartest must never be
a production dependency of anything).

Rationale per module:

- **static is its own module** so that embedding the DataStar JS bundle adds
  zero Go dependencies to whatever consumes it. A consumer who only wants the
  asset can depend on `static` alone.
- **datastartest is its own module** so that test helpers never leak into the
  production dependency graph of root — the original bug, now structurally
  prevented. The dogfood E2E test lives there because it tests the helpers.
- **root stays a single flat package** — a protocol library gains nothing
  from internal package layers.

### Mutual replaces for local development, real versions for consumers

Replace directives make sibling modules resolve locally during development:

| Module       | Replace directives                         |
| ------------ | ------------------------------------------ |
| root         | `static => ./static`                       |
| static       | (none — leaf)                              |
| datastartest | `go-datastar => ..`, `static => ../static` |

Rules enforced by CI and tests:

1. **Relative paths only** — the CI "Replace directive audit" greps for
   `replace.*=>/` (absolute paths) and fails the build. Absolute paths break
   every other checkout.
2. **`go work sync` must be idempotent** — a dedicated CI step diffs go.work
   before/after sync; drift means go.work was hand-edited inconsistently with
   the go.mod files.
3. **Every module must build with `GOWORK=off`** — CI loops
   `. ./datastartest ./static` with `GOWORK=off go build && go test`. This
   proves the replace directives are complete and each module resolves
   standalone, exactly as a consumer's build would.
4. **root must never require datastartest** — `module_boundary_test.go`
   regression-guards the circular dependency at test time, not just review
   time.

### Lockstep versioning with per-module tags

All modules share a version number and are tagged together from the same
commit: `v0.2.0`, `static/v0.2.0`, `datastartest/v0.2.0`. Consumers see real
versions in `require` blocks (the replaces are local-only and never shipped),
so root's `go.mod` carries `static v0.2.0` and datastartest's carries
`go-datastar v0.2.0` + `static v0.2.0`. The require versions are the
compatibility contract; the replaces are the developer convenience.

## Consequences

- **Adding a fourth module** means: new `go.mod`, `use` entry in go.work,
  sibling replaces where needed, a lockstep tag prefix, and a row in the CI
  GOWORK=off loop. The pattern scales mechanically.
- **Version bumps of sibling modules** are two-step: tag the sibling, then
  bump the `require` version in dependents. Forgetting the second step is
  caught by the workspace (`go work sync`) only partially — the GOWORK=off
  build is the real guard, since the workspace masks require-version drift.
- **`v0.0.0` vs real versions in sibling requires** is a standing open
  question (ROADMAP): replaces make the version irrelevant locally, but real
  versions document tested compatibility and keep `go mod tidy` from emitting
  pseudo-versions if a replace is ever removed.
- **The E2E dogfood test runs in datastartest, not root** — root's own
  `e2e_test.go` retains only transport-level checks (go-sse-owned behavior).
  Anyone changing root's wire format runs the datastartest suite to see the
  full round-trip.
