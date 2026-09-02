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

# Isolation mode (GOWORK=off, per-module — verifies replace directives; run from each module dir):
GOWORK=off GOEXPERIMENT=jsonv2 go test ./...

# Error audit (all modules — erraudit v0.3.0 takes ONE directory per run, never package patterns):
for mod in . ./datastartest ./static; do
  (cd "$mod" && GOEXPERIMENT=jsonv2 erraudit . --type-aware --enforce-go-error-family --no-suppress)
done

# Fuzz smoke tests (30s; corpora are committed regression suites — see CONTRIBUTING.md "Fuzzing"):
GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadSignals$' -fuzztime 30s .
(cd datastartest && GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadEvents$' -fuzztime 30s .)

# CI also enforces (run locally to pre-empt CI failures):
GOEXPERIMENT=jsonv2 go work sync        # go.work must not change after sync (idempotency)
go work use . ./datastartest ./static   # go.work must match this exactly
GOWORK=off go mod tidy -diff            # per module; must print nothing
GOWORK=off go mod verify                # per module; "all modules verified"
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

One flat package; roles in `doc.go`, discoverable via `ls`. Key files:
`patch.go` (Patch interface), `errors.go` (catalog), `response.go` (fluent
builder), `store.go` (replay), `script_handler.go` (JS serving), `example/`.

## Wire-Format Parity Requirements

The exact upstream-parity behaviors (dataline keys/order, splitting, defaults,
elision) are pinned by `wire_golden_test.go` and documented in
[docs/wire-format.md](docs/wire-format.md). A golden change IS a wire-format
change — make it deliberately, against the DataStar SDK, and record it in the
CHANGELOG.

> go-sse's `JoinLines`/`KeyedLines` are deliberately NOT adopted: `KeyedLines`
> normalizes CRLF to LF (upstream splits on `\n` only) and its key convention
> conflicts with the trailing-space dataline constants.

## CI

- `ci.yml` — test (build/vet/race + GOWORK=off ×3 + `go mod verify` +
  go.work use-vs-disk + tidy-diff + sync idempotency + replace audit +
  JS-version-in-CHANGELOG drift test), lint (golangci-lint v2.12.2
  go-installed, analysis cache cached), erraudit (probe-gated while the repo
  is private), govulncheck. Runs ONLY on code-affecting paths (`paths`
  filter) — docs-only pushes skip it entirely.
- `actionlint.yml` — workflow YAML validation on EVERY push/PR (the signal
  that still fires when ci.yml skips).
- `coverage.yml` — master-push coverage badge to the orphan `coverage`
  branch (same `paths` filter).
- `nix.yml` — hermetic `nix flake check` on code-affecting paths (first
  green run 2026-09-03; `continue-on-error` until proven stable).
- `fuzz.yml` — scheduled daily 60s fuzz runs over all four fuzz targets,
  crash artifacts uploaded.
- `codeql.yml` — GitHub CodeQL Go security analysis (SHA-pinned action).
- `renovate.json` — custom manager proposing embedded-DataStar-JS bumps from
  upstream releases into `static/static.go`; coexists with
  `.github/dependabot.yml` (one-bot decision pending, see TODO_LIST).
- Actions are SHA-pinned; nothing is a required check (branch protection
  removed); local gates are the real gate.

## Gotchas

- **gopls `stdversion` warnings on `encoding/json/v2` are false positives — do
  not "fix" them.** gopls (v0.23.0) flags ANY json/v2 symbol as "requires
  go1.27" (its stdlib DB has no GOEXPERIMENT awareness); under
  `GOEXPERIMENT=jsonv2` the package is fully available in go1.26. `go vet` and
  golangci-lint never flag it. Every alternative spelling triggers the same
  warning, and the v1 `encoding/json` API would change error types and HTML
  escaping (wire format). Leave the v2 direct calls.
- **Shared checkout — one section for the whole operational reality:** the
  auto-commit daemon commits and pushes dirty files to whatever branch is
  checked out, multiple crush sessions share this checkout, and `git town`
  (v24) manages syncs. Branch tips can move or lose commits at any moment.
  Re-verify with `git log`/`git reflog` before and after every git operation;
  stage by explicit path list (never `git add -A`); re-read files before
  every edit (mod-time races); quarantine risky work in a `git worktree`.
  A failed `git sync` leaves an unfinished run: `git town status`, then
  `skip`/`continue` (they abort on missing lineage — fix via
  `git config --add git-town.observed-branches <branch>` or
  `git config git-town-branch.<branch>.parent master` without moving HEAD;
  prune stale lineage after branch deletions). `git town propose` =
  one-command branch+push+PR. Session ritual: `git town status`, `git status`,
  `gh pr list` at start; clean tree + synced master at end.
- `go.work` is committed, but a **global** gitignore (`~/.config/git/ignore`)
  can still hide it on some machines. After touching `.gitignore` or creating
  module files, run `git check-ignore -v <file>` and `git ls-files <file>` —
  `git status` alone lies when a global ignore is in play.
- `dprint.json` documents formatting intent for non-Go files but is NOT wired
  into treefmt/flake (canonical formatting = treefmt via `nix flake check`).
- `origin/master` has NO branch protection (owner decision): CI is
  informational, nothing blocks a bad push — run the gates locally first.
- Status reports live in `docs/status/*.md` (indexed by its README) and are
  point-in-time snapshots — excluded from CHANGELOG entries by policy.
  `.md` is the repo's report format (the status-report skill's HTML default
  is overridden by convention).

## Error System

Every error returned by go-datastar is a classified `*errorfamily.Error` carrying
a stable machine-readable **code**, a behavioral **family**, and structured
**context**. The catalog lives in `errors.go`; the full matching guide (by
code / sentinel / family, with the family table and every code) is
[docs/error-system.md](docs/error-system.md) and the README.

### Sentinels

- `ErrBodyReadAfterClose` (wraps `http.ErrBodyReadAfterClose`, preserving the cause)
- `ErrEventNameRequired`

### Design decisions

1. **Library contract: go-error-family only, never samber/oops.** Libraries
   classify but never presume the app's observability stack.
2. **Return `error` interface, not `*errorfamily.Error`.** Idiomatic Go;
   typed access via `errorfamily.Code` / `errors.Is` / `errors.As`. erraudit's
   `generic_return` warnings on these signatures are accepted by design.
3. **Sentinels stay context-pristine.** `WithContext` returns a clone, so shared
   sentinels never leak caller-specific context.
4. **Context loss is a bug.** Wrapping errors include relevant in-scope values
   (HTTP method, input byte length, value type) so diagnosis needs no re-run.
5. **Layered composition with go-sse v0.5+.** The transport classifies its own
   errors; `wrapStreamError` wraps Send failures as
   `datastar.stream_send_failed` (Transient). `errorfamily.Classify` returns
   the outermost family, so go-datastar's classification wins; `errors.Is`
   traverses the chain, so callers matching go-sse codes work transparently.

## What This Library Is NOT

No CQRS, no event bus, no domain opinions. It is a pure protocol layer. Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's EventBridge).

## Nix / Build Gotchas

- **erraudit + go-finding are private repos** → no hermetic Nix build possible.
  CI probe-gates with `go list -m`; the flake app go-installs with credentials.
- **treefmt-nix `flakeCheck = true` registers its OWN unguarded `checks.treefmt`**
  without go on PATH. goimports shells out to `go env` per file; without a
  directive-satisfying `go` first on PATH, the sandbox tries a network
  toolchain download. Keep `flakeCheck = false` + a guarded `checks.format`
  that prepends `goPkg` to `buildInputs`.
- **vendorHash sensitivity (verified 2026-09-02/03, ADR 004 correction).**
  Root `vendorHash` moves only on requires (go.mod/go.sum) or toolchain
  `modules.txt` changes. `datastartestVendorHash` used to move on ANY edit to
  any tracked file under the repo root or static/ — and with flake.nix in the
  FOD input the paste-dance could NEVER converge (unsolvable self-reference).
  FIXED: the datastartest check now uses a MINIMAL src fileset (datastartest,
  root *.go + go.mod, static/), so metadata edits don't touch it; the FOD
  converges on one paste. Don't widen that fileset.
- **`buildGoModule` `modRoot`** builds a submodule in place (vendor + main
  derivations both `cd "$modRoot"`).
- **BOM in Go source = compile error.** Use the escape `"\xef\xbb\xbf"` in
  test seed data.

## E2E Testing for Consumers: `datastartest/`

`datastartest` is a separate module of reusable E2E helpers (SSE parsing,
dataline decoding, `Collect*` helpers, assertions). Full API tour:
[datastartest/README.md](datastartest/README.md). The wire-format E2E test
lives in `datastartest/e2e_test.go`; root's retains only go-sse-owned header
checks. Root must never require datastartest.

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
