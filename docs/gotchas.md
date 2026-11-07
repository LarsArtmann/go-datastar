# Gotchas & Operational Detail — go-datastar

Companion to [AGENTS.md](../AGENTS.md): the full text of every gotcha,
module decision, and build detail AGENTS.md summarizes. AGENTS.md keeps the
headline + a pointer; this file keeps the reasoning.

## Module Structure Decisions

Four Go modules in a go.work workspace (rationale and rules:
[ADR 002](adr/002-multi-module-split.md)).

Replace directives: root go.mod replaces `static => ./static`; broadcast and
datastartest go.mod replace `go-datastar => ..` and `static => ../static`. All
resolve locally for `GOWORK=off` builds (CI, Nix, consumers). Root must NEVER
require datastartest (circular dependency; `module_boundary_test.go` enforces
it). Root also never requires broadcast (same boundary, one direction).

Decisions:

- `go.work.sum` is intentionally gitignored (the toolchain regenerates it on
  demand). Per-module `go.sum` files are the reproducibility source of truth
  for consumers; the replace directives make sibling-module checksums
  unnecessary locally, while workspace-mode hashes for EXTERNAL modules
  accumulate in `go.work.sum` without a reproducibility contract — consumers
  never see it.
- Sibling requires use real published versions (not `v0.0.0`) so consumers
  testing without replaces resolve to a real published module. BuildFlow's
  `workspace/pseudo-version-hygiene` preflight FAILs on exactly this — a
  verified false positive for lockstep-tagged multi-module repos (zeroing the
  require would re-poison consumers through MVS; see the 2026-10-03 fleet
  de-poisoning). Documented deliberate non-fix in `.buildflow.yml`.
- The `go` directives are minor form (`go 1.27` across go.mod ×3 and go.work;
  static stays at `go 1.26`) since the 2026-10-03 fleet de-poisoning — a
  patch-pinned floor re-poisons consumers through MVS. CI's setup-go pins
  `1.27.3` (moved from 1.27.1 by `09fe522` for the govulncheck stdlib fixes)
  and the flake `goPkg` is nixpkgs `go_1_27`, both satisfying the minor-form
  directives while keeping toolchain choice deliberate. `.buildflow.yml` sets
  `tool_options.go-version-auto-configure.respect_patch_floor: true` so
  BuildFlow respects the deliberate patch pin.

## CI Detail

- `ci.yml` — per-module matrix (GOWORK=off build/vet/race + mod verify +
  tidy-diff) plus a workspace job (race suite, use-vs-disk, sync idempotency,
  replace audit, JS-version drift test); lint (golangci-lint v2.14.0
  single-sourced via `GOLANGCI_VERSION`, go-installed — 2.13.2 cannot read
  go1.27.2+ export data), erraudit (probe-gated while private), govulncheck.
  Runs ONLY on code-affecting paths — docs-only pushes skip it entirely.
- `actionlint.yml` — workflow YAML validation on EVERY push/PR (the signal
  that still fires when ci.yml skips).
- `coverage.yml` — master-push coverage badge to the orphan `coverage`
  branch (same `paths` filter).
- `nix.yml` — hermetic `nix flake check` on code-affecting paths (promoted
  2026-09-18: `continue-on-error` dropped; a red run is red master).
- `fuzz.yml` — scheduled daily 300s fuzz runs over all six fuzz targets,
  crash artifacts uploaded.
- `codeql.yml` — GitHub CodeQL Go security analysis (SHA-pinned action).
- `renovate.json` — custom manager proposing embedded-DataStar-JS bumps from
  upstream releases into `static/static.go`. One-bot decision settled
  2026-10-01: Dependabot owns ecosystems (gomod ×4 + actions); Renovate is
  scoped to this custom manager only via `enabledManagers: ["regex"]`.
- All workflows declare least-privilege `permissions:` (cue-vet fleet
  policy; added 2026-11-07). Actions are SHA-pinned; nothing is a required
  check (branch protection removed); local gates are the real gate.

## Gotchas

### Serving the JS client

`ScriptTagVersioned(path)` is the recommended pattern (versioned `?v=` cache
key, idempotent — it replaces an existing `v` value rather than
double-appending). The example's index page pins the rendered tag with a
smoke test, and datastartest's `RequireScriptTagVersion` is the
consumer-side assertion; never regress a page back to a bare `datastar.js`
src.

### gopls false positives

gopls warnings on `encoding/json/v2` are false positives — do not "fix"
them. Its stdlib DB lags the toolchain (flagged json/v2 as "requires
go1.27" long before it was standard; `go vet`/golangci-lint never flag it).
The v1 API would change error types and HTML escaping (wire format). Same
class: the `infertypeargs` hint on `sse.WithBufferSize` in broadcast
(dropping `[sse.Event]` does not compile). Verify any LSP hint with the
compiler before acting on it.

### Shared checkout

The auto-commit daemon commits and pushes dirty files to whatever branch is
checked out, multiple crush sessions share this checkout, and `git town`
(v24) manages syncs. Branch tips can move or lose commits at any moment.
Re-verify with `git log`/`git reflog` before and after every git operation;
stage by explicit path list (never `git add -A`); re-read files before every
edit (mod-time races); quarantine risky work in a `git worktree`. A failed
`git sync` leaves an unfinished run: `git town status`, then
`skip`/`continue` (they abort on missing lineage — fix via
`git config --add git-town.observed-branches <branch>` or
`git config git-town-branch.<branch>.parent master` without moving HEAD;
prune stale lineage after branch deletions). `git town propose` =
one-command branch+push+PR. Session ritual: `git town status`, `git status`,
`gh pr list` at start; clean tree + synced master at end.

The daemon's heuristic commits have also DELETED files it should not touch:
`.buildflow.yml` vanished in `99d4ab5` (2026-11-07) and every documented
skip silently stopped applying. If BuildFlow findings that were previously
skipped reappear, check `git log --follow -- .buildflow.yml` first and
restore from git.

### Recurring F/spacing damage in archived reports — ROOT CAUSE FOUND

`docs/status/archived/2026-08-16_08-47…:19` repeatedly had the space before
an inline-code span stripped (the word "against" glued to the following
backtick span). Root cause identified 2026-11-07: **the dprint markdown
plugin (0.22.1) itself strips that space** — it mis-parses the line's
escaped-backtick construct (`at backtick-hash-backtick` inside a code span),
loses track of the span boundary, and re-joins the following text. Verified
with `dprint check` on the fixed file: dprint wants the damaged form. Any
session running `dprint fmt` / `buildflow format` over the repo re-damaged
the line; treefmt (the canonical format gate) does not run dprint, so CI
never caught it — docs-gates' annotate gate is the detector.

Damage landed via daemon commits `4539949`, `2e05692`, `ff4f51e`, `da3e0e3`,
`1742daf`, `2b4f9b7` (fixes alternated in `c4474a3` and others). Durable fix
(2026-11-07): `docs/status/archived/**` is now in `dprint.json` excludes —
archived reports are point-in-time snapshots that no formatter may touch.
Never quote the damaged byte pattern verbatim in living docs — the F/spacing
detector parses file text and flags the example itself.

### Shared golangci-lint cache ghost findings (verified 2026-09-18)

The session env points `GOLANGCI_LINT_CACHE` at
`/mnt/buildcache/golangci-lint` (shared, multi-GB); it held results for a
concurrent session's worktree (`/tmp/gd-prereview`), and after that worktree
was deleted the cache kept emitting its gochecknoglobals/nolintlint findings
on every full lint — with warnings that the files no longer exist. If lint
reports issues in paths outside the repo, do NOT chase them in the code:
re-run with a fresh cache
(`GOLANGCI_LINT_CACHE=$(mktemp -d) go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ...`)
and trust only that verdict. Purging the shared cache itself is an owner
decision (other sessions rely on it).

### go.work vs global gitignore

`go.work` is committed, but a **global** gitignore (`~/.config/git/ignore`)
can still hide it on some machines. After touching `.gitignore` or creating
module files, run `git check-ignore -v <file>` and `git ls-files <file>` —
`git status` alone lies when a global ignore is in play.

### Formatting ownership

`dprint.json` documents formatting intent for non-Go files but is NOT wired
into treefmt/flake (canonical formatting = treefmt via `nix flake check`).
`static/datastar.js` is a vendored artifact — never reformat it. A prettier
sweep once reformatted the minified bundle and the auto-commit daemon
committed it, breaking the checksum pin (`static/checksum_test.go` — that
test failing means the bundle drifted, NOT that you should update the
constant). Shields: `.prettierignore`, `.codespellrc` (walk-mode glob skip +
`crasher`/`keep-alives` ignore-list; skip patterns MUST be `*/…` wildcards —
directory scans match `./`-prefixed paths, so exact patterns silently never
fire — verified 2026-11-07 with codespell 2.4.3; codespell also rides
treefmt via `programs.codespell` with history-corpus excludes), and dprint
excludes for `.github/DISCUSSION_TEMPLATE/**` (markdown-in-yml GitHub
templates that dprint's YAML parser rejects) and `docs/status/archived/**`
(see the F/spacing root cause above).

### docs-gates ledger markers must be a pure line

The `<!-- check-rows:ledger -->` table exemption only works as its own line
directly above the table header; gluing it onto other text silently cancels
the exemption (found by the 2026-11-07 archive sweep).

### BuildFlow

BuildFlow (repo-external quality pipeline) reads `.buildflow.yml`; every
skip and override in it is a deliberate, documented decision (see the file's
comments — eslint/oxfmt: no first-party JS, only the vendored bundle;
go-auto-upgrade: no samber/lo under the minimal-dependency policy;
`max_file_size: 700` for table-driven test suites). go-structure-linter
stays running but is shaped via `.go-structure-linter.yaml`
(`presets: [flat]` + rule suppressions with reasons). Its erraudit variant
flags `defer x.Close()` blank-ignores that the canonical gate tolerates;
errcheck excludes in `.golangci.yml` cover the concrete Close types
(`*sse.Stream`, `*gzip.Writer`). BuildFlow's result cache keys cover matched
files only — a finding can replay after its cause was fixed elsewhere;
bypass with `BUILDFLOW_NO_RESULT_CACHE=1` when verdicts look stale.

`license-check` must run inside the devshell (`nix develop -c buildflow
...`, closed 2026-10-01): go-licenses' stdlib detection compares package
paths against the GOROOT of the Go that built/ran it; outside the devshell
`GOTOOLCHAIN=auto` switches to the toolchain cache and every stdlib package
errors with "does not have module info" (the ×4 failure loop). The devshell
ships go-licenses + `GOTOOLCHAIN=local` + real GOROOT. Do NOT add
`GOTOOLCHAIN: local` to `.buildflow.yml` env — plain user-shell go is
1.26.7, local would then fail every go1.27.1-directive module build.

`govulncheck` findings of the form "package requires newer Go version
go1.27" mean BuildFlow ran outside the devshell (user-shell go 1.26.7 cannot
analyze go1.27 packages) — run the verification pass inside `nix develop`
like license-check.

branching-flow's slice-index findings ("may panic if out of bounds") are
accepted heuristic false positives on this repo's guarded patterns (spot-
checked 2026-11-07): the diff.go LCS table is allocated (n+1)×(m+1) so
`i+1`/`j+1` are in bounds by construction, storecontract's parallel-slice
range is length-asserted, and sugar.go's `slices.IndexFunc` results are
`i < 0`-guarded before use. Same for its swap-hazard hints on
stringly-typed test-assertion params — branding those types would churn the
public datastartest API for zero caller benefit. vulnix findings target
nixpkgs store derivations (binutils, brotli, …), not this repo's code —
accepted noise until nixpkgs moves.

### Submodule LICENSEs

Each published submodule (`broadcast/`, `datastartest/`, `static/`) carries
its own MIT LICENSE (module zips ship only the module dir; go-licenses
flagged license-less modules). Keep them in sync with the root LICENSE.

### No branch protection

`origin/master` has NO branch protection (owner decision): CI is
informational, nothing blocks a bad push — run the gates locally first.

### Status reports

Status reports live in `docs/status/*.md` (indexed by its README) and are
point-in-time snapshots — excluded from CHANGELOG entries by policy. `.md`
is the repo's report format (the status-report skill's HTML default is
overridden by convention).

## Nix / Build Gotchas

- **erraudit + go-finding are private repos** → no hermetic Nix build
  possible. CI probe-gates with `go list -m`; the flake app go-installs
  with credentials.
- **treefmt-nix `flakeCheck = true` registers its OWN unguarded
  `checks.treefmt`** without go on PATH. goimports shells out to `go env`
  per file; without a directive-satisfying `go` first on PATH, the sandbox
  tries a network toolchain download. Keep `flakeCheck = false` + a guarded
  `checks.format` that prepends `goPkg` to `buildInputs`.
- **vendorHash sensitivity (ADR 004).** Root `vendorHash` moves only on
  requires/toolchain changes. `datastartestVendorHash` moves on edits inside
  its MINIMAL src fileset (datastartest/, root *.go + go.mod, static/) —
  which is what makes the FOD converge on one paste. Don't widen that
  fileset (the pre-fix repo-root fileset was a self-referencing
  never-converging FOD).
- **`broadcastVendorHash` tracks ROOT SOURCE EDITS** (verified 2026-10-01 ×2
  in one session; re-verified live 2026-11-07 — comment-only root `.go` edits
  plus the static bundle byte bump moved it, and `buildflow -s nix-hash-fix
  --fix` landed the exact `got:` values with zero manual pastes): broadcast
  vendors root + static through its directory replaces, so ANY change to
  root's non-test `.go` files or LICENSE that `go mod vendor` copies moves
  it — not just require bumps. After touching root source, expect one more
  `--keep-going` pass + paste. Test files (`*_test.go`, `docspec`) do NOT
  move it.
- **nix CI reports only the FIRST hash mismatch** — a require bump can move
  BOTH submodule vendor hashes while the workflow names one and exits.
  Reproduce with a full local `nix flake check --keep-going` (the
  `--keep-going` flag collects EVERY moved hash in one pass; a plain check
  stops at the first) and paste every moved hash in the same commit — or
  let `buildflow -s nix-hash-fix --fix` do it (verified 2026-11-07: it
  landed both moved hashes exactly).
- **ANY `go.mod`/`go.sum` edit in ANY module moves vendor hashes — and the
  nix path filter once hid it.** `datastartestSrc` includes the whole
  `datastartest/` dir, so even replace-directive drops there move
  `datastartestVendorHash` (2026-09-29). `nix.yml`'s paths filter now
  matches `**/go.mod` + `**/go.sum`, but the lesson stands: run
  `nix flake check` locally before pushing go.mod edits in ANY module. Root
  metadata (`.md`, CI yml) stays outside both filesets and never moves a
  hash.
- **`buildGoModule` `modRoot`** builds a submodule in place (vendor + main
  derivations both `cd "$modRoot"`).
- **BOM in Go source = compile error.** Use the escape `"\xef\xbb\xbf"` in
  test seed data.

## Docs Map

| Path                               | Content                                                     |
| ---------------------------------- | ----------------------------------------------------------- |
| `doc.go`                           | Package docs: design rationale, quick start, error contract |
| `docs/adr/`                        | Architecture decision records (multi-module, …)             |
| `docs/release-checklist.md`        | Pre-release gate, versioning, lockstep tags, verification   |
| `docs/replay.md`                   | MemoryStore + Last-Event-ID reconnection guide              |
| `docs/error-system.md`             | The three error-matching dimensions, codes + families       |
| `docs/wire-format.md`              | Annotated datalines per patch family (+ golden tests)       |
| `docs/architecture.md`             | Three-layer architecture diagram + file map                 |
| `docs/migration-starfederation.md` | Upstream SDK → go-datastar migration mapping                |
| `docs/testing.md`                  | datastartest quick start, fuzzing, coverage story           |
| `docs/ci-watch.md`                 | Promote/drop/verify runbook for the non-required workflows  |
| `docs/duplication-baseline.md`     | art-dupl accepted-clone baseline (audit re-verify gate)     |
| `docs/performance.md`              | Measured benchmark table                                    |
| `docs/version.md`                  | The three version concepts (module / JS client / ldflags)   |
| `docs/migration-guide.md`          | v0.2.0 → v0.3.0 upgrade guide                               |
| `docs/static-js.md`                | Embedded JS pinning + upgrade process                       |
| `docs/gotchas.md`                  | This file: full gotcha text, module decisions, docs map     |
| `docs/status/` (+ its README)      | Point-in-time status reports and audits (index + policy)    |
| `docs/planning/`                   | Pareto plans; `archived/` holds executed plans              |
| `CONTRIBUTING.md`                  | Dev setup, workspace rules, fuzzing                         |
| `CHANGELOG.md`                     | Released history (append-only) + `[Unreleased]` draft       |
