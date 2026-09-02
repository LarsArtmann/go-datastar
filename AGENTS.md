# AGENTS.md — go-datastar

DataStar protocol library for Go. Patches as first-class values producing `sse.Event`. Built on go-sse. Single package (`datastar`), flat layout. The `datastartest/` subpackage is a separate Go module for consumer E2E testing.

## Module Structure

Three Go modules in a go.work workspace (rationale and rules: [ADR 002](docs/adr/002-multi-module-split.md)):

| Module       | Path                                              | Purpose                            | Dependencies            |
| ------------ | ------------------------------------------------- | ---------------------------------- | ----------------------- |
| Root         | `github.com/larsartmann/go-datastar`              | Protocol library                   | go-sse, go-error-family |
| static       | `github.com/larsartmann/go-datastar/static`       | Embedded DataStar JS client bundle | zero (stdlib only)      |
| datastartest | `github.com/larsartmann/go-datastar/datastartest` | Consumer E2E test helpers          | go-datastar, go-sse     |

Replace directives: root go.mod replaces `static => ./static`; datastartest
go.mod replaces `go-datastar => ..` and `static => ../static`. All resolve locally
for `GOWORK=off` builds (CI, Nix, consumers). Root must NEVER require
datastartest (circular dependency; `module_boundary_test.go` enforces it).

Decisions: `go.work.sum` is intentionally gitignored (advisory — the toolchain
regenerates it; per-module `go.sum` files are the reproducibility source of
truth, and replaces make sibling checksums unnecessary). Sibling requires use
real published versions (not `v0.0.0`) so consumers testing without replaces
resolve to a real published module. The `go` directive pins the exact patch
release (currently **1.26.7** across go.mod ×3, go.work, CI, and the flake
`overrideAttrs` pin) to clear stdlib CVEs under `GOTOOLCHAIN=local`.

## Commands

```bash
# Workspace mode (default, uses go.work) — covers all three modules:
GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1
GOEXPERIMENT=jsonv2 go vet ./... ./datastartest/... ./static/...
GOEXPERIMENT=jsonv2 golangci-lint run ./... ./datastartest/... ./static/...
# Pre-push lint = EXACT CI parity (flake: `nix run .#lint-ci`):
GOEXPERIMENT=jsonv2 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./... ./datastartest/... ./static/... --timeout 5m

# Isolation mode (GOWORK=off, per-module) — verifies replace directives:
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # root only
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # datastartest (run from datastartest/)
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...                      # static (run from static/)

# Error audit (all modules — erraudit v0.3.0 takes ONE directory per run, never package patterns):
for mod in . ./datastartest ./static; do
  (cd "$mod" && GOEXPERIMENT=jsonv2 erraudit . --type-aware --enforce-go-error-family --no-suppress)
done

# Fuzz smoke tests (30s; corpora are committed regression suites — see CONTRIBUTING.md "Fuzzing"):
GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadSignals$' -fuzztime 30s .
(cd datastartest && GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadEvents$' -fuzztime 30s .)

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

One flat package; file roles are documented in `doc.go` (design rationale,
quick start, error-system contract) and discoverable via `ls`. Key entry
points: `patch.go` (Patch interface), `errors.go` (error catalog),
`constants.go` (dataline keys), `response.go` (fluent SSE builder),
`store.go` (MemoryStore replay), `script_handler.go` (static JS serving),
`example/` (live-feed demo). Multi-module rationale:
[docs/modularization/README.md](docs/modularization/README.md).

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

## CI

- `ci.yml` — test (build/vet/race + GOWORK=off ×3 + sync idempotency +
  replace audit), lint (golangci-lint v2.12.2 go-installed, analysis cache
  cached), erraudit (probe-gated while the repo is private), govulncheck.
  Runs ONLY on code-affecting paths (`paths` filter) — docs-only pushes skip
  it entirely.
- `actionlint.yml` — workflow YAML validation on EVERY push/PR (the signal
  that still fires when ci.yml skips).
- `coverage.yml` — master-push coverage badge to the orphan `coverage`
  branch (same `paths` filter).
- `nix.yml` — hermetic `nix flake check` on code-affecting paths
  (`continue-on-error` until proven stable; promote or drop after a green
  week).
- `fuzz.yml` — scheduled daily 60s fuzz runs over all four fuzz targets,
  crash artifacts uploaded.
- `codeql.yml` — GitHub CodeQL Go security analysis (SHA-pinned action).
- `renovate.json` — custom manager proposing embedded-DataStar-JS bumps from
  upstream releases into `static/static.go`; coexists with
  `.github/dependabot.yml` (one-bot decision pending, see TODO_LIST).
- Actions are SHA-pinned (`checkout`, `setup-go` verified against their v7
  tags); nothing is a required check (branch protection removed); local
  gates are the real gate.

## Gotchas

- **gopls `stdversion` warnings on `encoding/json/v2` are false positives — do
  not "fix" them.** gopls (v0.23.0) flags ANY json/v2 symbol (`Marshal`,
  `Unmarshal`, `UnmarshalRead`, `MarshalWrite`, …) as "requires go1.27" because
  its stdlib DB records the graduated API version and has no GOEXPERIMENT
  awareness; under `GOEXPERIMENT=jsonv2` the package is fully available in
  go1.26. `go vet` and golangci-lint (the real gates) never flag it. Verified
  2026-09-02: every alternative spelling triggers the same warning, and
  switching to the v1 `encoding/json` API would change error types and default
  HTML escaping (wire format). Leave the v2 direct calls; benchmark `b.N`
  loops are the only genuine modernization (fixed via `b.Loop()`).
- `go.work` is committed, but a **global** gitignore (`~/.config/git/ignore`)
  can still hide it on some machines. After touching `.gitignore` or creating
  module files, run `git check-ignore -v <file>` and `git ls-files <file>` —
  `git status` alone lies when a global ignore is in play.
- `dprint.json` exists in the repo root but is NOT wired into treefmt/flake —
  it documents the project's intent for non-Go files (JSON, YAML, Markdown,
  Dockerfile) and supports editor/external integrations. Canonical formatting
  is treefmt (gofumpt/goimports/golines/nixfmt) via `nix flake check`; wiring
  dprint into the hermetic check would make it depend on network-fetched WASM
  plugins.
- `origin/master` has NO branch protection (removed 2026-08-16, owner
  decision): CI is informational, nothing blocks a bad push — run the gates
  locally first. The auto-commit daemon can commit straight to master.
- Multiple crush sessions share this checkout, plus an auto-commit daemon
  that commits dirty files to whatever branch is checked out. Branch tips
  can move or lose commits at any moment (for example a hard reset by a
  parallel session). Re-verify with `git log` and `git reflog <branch>`
  before and after every git operation, and quarantine work in a
  `git worktree` outside the main checkout. Stage by explicit path list;
  never `git add -A`; re-read files before every edit (mod-time races).
- `git town` (v24) manages syncs. A failed `git sync` leaves an unfinished
  run: check `git town status`, then `git town skip` (finish without the
  failing step; the checkout may end on another branch of the stack) or
  `git town continue` (retry). `skip`/`continue` abort non-interactively
  when a local branch has no configured lineage — fix without moving HEAD
  via `git config --add git-town.observed-branches <branch>` (snapshot
  branches) or `git config git-town-branch.<branch>.parent master` (stack
  children). All lineage/branchtype metadata lives in git config. After any
  branch deletion, prune its stale lineage.
- `git town propose` is the one-command branch+push+PR flow. `gh pr merge
  --merge --delete-branch` from master also deletes the LOCAL PR branch.
- Session-entry ritual: `git town status`, `git status`, `gh pr list` before
  starting work. End-of-session ritual: clean tree, synced master, no
  unfinished git-town run.
- Status reports live in `docs/status/*.md` (indexed by its README) and are
  point-in-time snapshots — excluded from CHANGELOG entries by policy.
  `.md` is the repo's report format (the status-report skill's HTML default
  is overridden by convention).

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
`datastartest.signals_unmarshal_failed`,
`datastartest.custom_event_detail_unmarshal_failed`.

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

## Nix / Build Gotchas

- **erraudit v0.3.0 CLI takes ONE directory arg.** Multi-pattern invocations
  (e.g., `erraudit ./... ./datastartest/...`) silently fail. Loop per-module:
  `(cd $mod && erraudit . --type-aware --enforce-go-error-family)`.
- **erraudit + go-finding are private repos** → no hermetic Nix build possible.
  CI probe-gates with `go list -m`; the flake app go-installs with credentials.
- **treefmt-nix `flakeCheck = true` registers its OWN unguarded `checks.treefmt`**
  without go on PATH. goimports shells out to `go env` per file; without a
  directive-satisfying `go` first on PATH, the sandbox tries a network
  toolchain download. Keep `flakeCheck = false` + a guarded `checks.format`
  that prepends `goPkg` to `buildInputs`.
- **vendorHash sensitivity (verified 2026-09-02, ADR 004 correction).** Root
  `vendorHash` moves only on requires (go.mod/go.sum) or toolchain
  `modules.txt` changes. `datastartestVendorHash` moves on ANY edit to ANY
  tracked file under the repo root or static/ — `go mod vendor` copies the
  directory-replaced module directories entirely (docs included). Re-discover
  via the fakeHash dance; refresh at the release gate, not ad hoc.
- **`buildGoModule` `modRoot` attribute** points into the repo source for
  submodule builds. The vendor + main derivations both `cd "$modRoot"` — no
  manual `postPatch` cd hacks needed. Available in nixpkgs at the locked rev.
- **BOM in Go source = compile error** ("illegal byte order mark"). Use the
  escape `"\xef\xbb\xbf"` for UTF-8 BOM in test seed data.
- **Auto-commit daemon is active.** It commits and pushes the working tree
  automatically. Check `git log` before assuming what's committed.

## E2E Testing for Consumers: `datastartest/`

`datastartest` is a separate module of reusable E2E helpers (SSE parsing,
DataStar dataline decoding, `Collect*` helpers, assertions). Quick start,
request options, and the full API tour: [datastartest/README.md](datastartest/README.md).
The wire-format E2E test (`TestE2E_DataStarPatches`) lives in
`datastartest/e2e_test.go`; root's `e2e_test.go` retains only the
go-sse-owned header checks. Root must never require datastartest.

**Invariant:** all public helpers accept `testing.TB` (not `*testing.T`), so
they work with `*testing.T`, `*testing.B`, and Ginkgo's `GinkgoT()`. Keep this
when adding helpers.

## Docs Map

| Path                          | Content                                                     |
| ----------------------------- | ----------------------------------------------------------- |
| `doc.go`                      | Package docs: design rationale, quick start, error contract |
| `docs/adr/`                   | Architecture decision records (multi-module, …)             |
| `docs/release-checklist.md`   | Pre-release gate, versioning, lockstep tags, verification   |
| `docs/replay.md`              | MemoryStore + Last-Event-ID reconnection guide               |
| `docs/error-system.md`        | The three error-matching dimensions, codes + families        |
| `docs/wire-format.md`         | Annotated datalines per patch family (+ golden tests)        |
| `docs/testing.md`             | datastartest quick start, fuzzing, coverage story            |
| `docs/performance.md`         | Measured benchmark table                                     |
| `docs/migration-guide.md`     | v0.2.0 → v0.3.0 upgrade guide                                |
| `docs/static-js.md`           | Embedded JS pinning + upgrade process                        |
| `docs/status/` (+ its README) | Point-in-time status reports and audits (index + policy)    |
| `docs/planning/`              | Pareto plans; `archived/` holds executed plans              |
| `CONTRIBUTING.md`             | Dev setup, workspace rules, fuzzing                         |
| `CHANGELOG.md`                | Released history (append-only) + `[Unreleased]` draft       |
