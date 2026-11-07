# AGENTS.md — go-datastar

DataStar protocol library for Go. Patches as first-class values producing `sse.Event`. Built on go-sse. Single package (`datastar`), flat layout. The `broadcast/` submodule adds SSE connection lifecycle (fan-out, replay, hub sharing); `datastartest/` is a separate Go module for consumer E2E testing.

**Full gotcha text, module decisions, CI detail, and docs map: [docs/gotchas.md](docs/gotchas.md).** This file keeps only what every session needs at a glance.

## Module Structure

Four Go modules in a go.work workspace (rationale and rules: [ADR 002](docs/adr/002-multi-module-split.md)):

| Module       | Path                                              | Purpose                            | Dependencies            |
| ------------ | ------------------------------------------------- | ---------------------------------- | ----------------------- |
| Root         | `github.com/larsartmann/go-datastar`              | Protocol library                   | go-sse, go-error-family |
| broadcast    | `github.com/larsartmann/go-datastar/broadcast`    | SSE fan-out, replay, hub sharing   | go-datastar, go-sse     |
| static       | `github.com/larsartmann/go-datastar/static`       | Embedded DataStar JS client bundle | zero (stdlib only)      |
| datastartest | `github.com/larsartmann/go-datastar/datastartest` | Consumer E2E test helpers          | go-datastar, go-sse     |

Root replaces `static => ./static`; broadcast/datastartest replace `go-datastar => ..` and `static => ../static`. Root NEVER requires datastartest or broadcast (`module_boundary_test.go` enforces it). Sibling requires use REAL published versions — never `v0.0.0` — so consumers resolve without replaces (BuildFlow's pseudo-version-hygiene preflight FAIL on this is a documented false positive). Decisions detail: [docs/gotchas.md](docs/gotchas.md#module-structure-decisions).

## Session-Start CI Health Probe

Before trusting a clean local tree, probe master CI health in every repo the session touches: `gh run list --branch master --limit 5`. Master has repeatedly sat red while local trees looked clean; the fix is usually already known by the time you look.

## Commands

```bash
# Workspace mode (default, uses go.work) — covers all four modules:
go test ./... ./broadcast/... ./datastartest/... ./static/... -race -count=1
go vet ./... ./broadcast/... ./datastartest/... ./static/...
# Pre-push lint = EXACT CI parity (flake: `nix run .#lint-ci`); GOLANGCI_VERSION
# in ci.yml is the single source for the pin (v2.14.0):
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./... ./broadcast/... ./datastartest/... ./static/... --timeout 5m

# Hermetic gate (builds + tests all four modules + format; run before pushing
# go.mod/go.sum edits — ANY such edit can move vendor hashes):
nix flake check

# Isolation mode (GOWORK=off, per-module — verifies replace directives; run from each module dir):
GOWORK=off go test ./...

# Error audit (erraudit takes ONE directory per run, never package patterns; the
# CI gate uses `--severity-threshold error` WITHOUT --no-suppress — that ×4 is the green target):
for mod in . ./broadcast ./datastartest ./static; do
  (cd "$mod" && erraudit . --type-aware --enforce-go-error-family --no-suppress)
done

# Fuzz smoke tests (30s; nightly matrix runs all six targets at 300s — CONTRIBUTING.md):
go test -run '^$' -fuzz '^FuzzReadSignals$' -fuzztime 30s .
(cd datastartest && go test -run '^$' -fuzz '^FuzzReadEvents$' -fuzztime 30s .)

# CI also enforces (run locally to preempt CI failures):
go work sync        # go.work must not change after sync (idempotency)
go work use . ./broadcast ./datastartest ./static   # go.work must match this exactly
GOWORK=off go mod tidy -diff            # per module; must print nothing
GOWORK=off go mod verify                # per module; "all modules verified"
grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod  # must find nothing
```

`GOEXPERIMENT=jsonv2` is a no-op under the 1.27 toolchain — do not re-add it. Do **not** pass `--enforce-samber-oops` to erraudit (libraries classify via go-error-family only, never samber/oops).

## Dependencies

- `github.com/larsartmann/go-sse` — SSE transport (Stream, Broadcaster, EventStore, Replay)
- `github.com/larsartmann/go-error-family` — structured error classification; every error this library returns is a classified `*errorfamily.Error`
- Transitive: `go-branded-id` (via go-sse)

## Architecture

Three layers:

| Layer     | Repo                    | Role                                                                                  |
| --------- | ----------------------- | ------------------------------------------------------------------------------------- |
| Transport | go-sse                  | Stream, Broadcaster[T], Replay, EventID, Heartbeat, SubscribeFilter, Shutdown, Health |
| Protocol  | go-datastar (this repo) | Patch interface, ElementsPatch, SignalsPatch, ScriptPatch, RedirectPatch, etc.        |
| Domain    | cqrs-htmx/datastar      | EventBridge (CQRS event → Patch), thin re-exports                                     |

The keystone: `type Patch interface { Event() sse.Event }` — every protocol message is a value producing an `sse.Event`, making patches storable, filterable, replayable, broadcastable (none of which the upstream SDK supports). One flat package; roles in `doc.go`; key files: `patch.go`, `errors.go`, `response.go`, `store.go`, `script_handler.go`, `example/`.

## Wire-Format Parity

Upstream-parity behaviors are pinned by `wire_golden_test.go` and documented in [docs/wire-format.md](docs/wire-format.md). A golden change IS a wire-format change — make it deliberately, against the DataStar SDK, record it in the CHANGELOG. go-sse's `JoinLines`/`KeyedLines` are deliberately NOT adopted (CRLF normalization + key-convention conflict).

## CI

Per-module matrix + workspace job + lint/erraudit/govulncheck (`ci.yml`, code-affecting paths only); `actionlint.yml` every push; `coverage.yml`; `nix.yml` (hermetic flake check); `fuzz.yml` (daily 300s); `codeql.yml`; Dependabot owns ecosystems, Renovate only the embedded-JS custom manager. Nothing is a required check — local gates are the real gate. Full workflow detail: [docs/gotchas.md](docs/gotchas.md#ci-detail).

## Gotchas (headlines — full text in [docs/gotchas.md](docs/gotchas.md))

- **`ScriptTagVersioned(path)` is the recommended JS-serving pattern** — never regress to a bare `datastar.js` src.
- **gopls warnings on `encoding/json/v2` are false positives** — do not "fix" them; verify any LSP hint with the compiler first.
- **Shared checkout:** auto-commit daemon commits/pushes dirty files; multiple sessions share this checkout; `git town` v24. Re-verify `git log` before/after git ops; stage by explicit paths; re-read files before edits. The daemon's heuristic has DELETED `.buildflow.yml` once (`99d4ab5`) — if documented skips vanish, `git log --follow -- .buildflow.yml`.
- **Recurring F/spacing damage in archived reports — root cause found:** the dprint markdown plugin strips the space itself when it mis-parses escaped backticks. `docs/status/archived/**` is now excluded in `dprint.json` (2026-11-07); docs-gates' annotate gate is the detector.
- **Shared golangci-lint cache can replay ghost findings** — paths outside the repo: re-run with `GOLANGCI_LINT_CACHE=$(mktemp -d)`.
- **`static/datastar.js` is vendored — never reformat it** (`static/checksum_test.go` failing = bundle drifted, not a constant to update). Shields: `.prettierignore`, `.codespellrc`, dprint excludes.
- **BuildFlow:** every skip in `.buildflow.yml` is deliberate (oxfmt/eslint: only JS is the vendored bundle; go-auto-upgrade: no samber/lo policy; `max_file_size: 700` for table-driven tests). go-structure-linter is shaped via `.go-structure-linter.yaml` (`presets: [flat]`). Run inside `nix develop -c` for license-check and govulncheck. Stale verdicts: `BUILDFLOW_NO_RESULT_CACHE=1`.
- **docs-gates ledger markers must be a pure line** directly above the table header.
- **No branch protection on master** — run the gates locally; CI is informational.
- **Status reports** (`docs/status/`) are point-in-time snapshots, excluded from CHANGELOG; `.md` format overrides the status-report skill's HTML default.

## Error System

Every error is a classified `*errorfamily.Error` (stable code, behavioral family, structured context); catalog in `errors.go`, full matching guide in [docs/error-system.md](docs/error-system.md). Sentinels: `ErrBodyReadAfterClose`, `ErrEventNameRequired`. Design decisions: go-error-family only (never samber/oops); return `error`, not `*errorfamily.Error`; sentinels stay context-pristine (`WithContext` clones); context loss is a bug; go-sse Send failures wrap as `datastar.stream_send_failed` (Transient) — go-datastar's classification wins, `errors.Is` still traverses to go-sse codes.

## What This Library Is NOT

No CQRS, no event bus, no domain opinions. Pure protocol layer; consumers build domain adapters (e.g., cqrs-htmx/datastar's EventBridge).

## E2E Testing: `datastartest/`

Separate module of consumer E2E helpers (SSE parsing, dataline decoding, `Collect*`, assertions); full API tour: [datastartest/README.md](datastartest/README.md). Root never requires it. **Invariant:** all public helpers accept `testing.TB` (works with `*testing.T`, `*testing.B`, Ginkgo's `GinkgoT()`) — keep it.

## Nix / Build (headlines — full text in [docs/gotchas.md](docs/gotchas.md#nix--build-gotchas))

- erraudit + go-finding are private → no hermetic build; CI probe-gates.
- treefmt must stay `flakeCheck = false` with a guarded `checks.format` (goimports needs go on PATH).
- `broadcastVendorHash` moves on ANY root non-test `.go`/LICENSE edit, not just require bumps; `nix flake check --keep-going` collects ALL moved hashes; `buildflow -s nix-hash-fix --fix` lands them.
- ANY `go.mod`/`go.sum` edit in ANY module can move vendor hashes — `nix flake check` before pushing.
- BOM in Go source = compile error; use `"\xef\xbb\xbf"` in test seeds.
