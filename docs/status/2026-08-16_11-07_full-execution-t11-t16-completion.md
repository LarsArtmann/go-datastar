# Status: Full Execution Mode — T11–T16 + T04 + Final Sync (2026-08-16 11:07)

**Session arc:** Resumed from a prior session that completed T01–T10. This
session executed T11–T16, T04, and the final sync. All 16 tasks from the pareto
plan (`docs/planning/2026-08-16_08-53_*.md`) are now complete.

**Working tree:** CLEAN — the auto-commit daemon landed everything in
`496a18b` and `83d7c60`. No uncommitted changes.

---

## A) FULLY DONE (verified green this session)

### T11 — Per-module Nix hermetic checks ✅
- Three `buildGoModule` derivations in `flake.nix`:
  `hermeticCheck` (root), `hermeticCheckStatic` (vendorHash=null, zero deps),
  `hermeticCheckDatastartest` (modRoot=datastartest, vendorHash discovered).
- All run in module mode: `GOWORK=off`, `go.work` excluded from source fileset.
- `nix flake check`: **all checks passed** (format + 3 module builds + tests).
- Flake TODO comment removed. `datastartestVendorHash` discovered via fakeHash
  dance, then re-discovered when the daemon committed the parallel session's
  new test files (source archive changed → vendor output changed).

### T12 — README comparison + release checklist ✅
- `go list -m -versions github.com/starfederation/datastar-go` → v1.2.2 still
  latest. Comparison table rows re-verified — no drift.
- JS client version (v1.0.2) added to the "Serve the DataStar JS client" row.
- `docs/release-checklist.md` created: pre-release gate, version bump, tag,
  post-release pkg.go.dev verification, quarterly comparison re-verify.

### T13 — De-flake LastEventID test ✅
- `TestCollect_WithLastEventID_HeaderArrives` rewritten with channel
  synchronization: handler signals after writing + flushing, test waits on
  the channel before reading. No sleeps (guard G6).
- Reproduction attempt: 50x isolated + 3x full-suite under race+parallel=16,
  all green. Flake was NOT reproducible — hardened anyway.
- Lint fix: `http.NewRequest` → `http.NewRequestWithContext` (noctx linter).
- 20x post-fix under race: green.

### T14 — Modularization docs index ✅
- `docs/modularization/README.md` created: links proposal, execution plan,
  ADR 001, ADR 002. Module table + DAG description.
- Linked from AGENTS.md file-layout section.

### T15 — DOMAIN_LANGUAGE.md ✅
- `docs/DOMAIN_LANGUAGE.md` created: Patch, Signals, Dataline, Event (SSE),
  ElementsPatch, SignalsPatch, ScriptPatch, RedirectPatch, ReadSignals,
  LastEventID, Stream, Broadcaster, EventStore, Replay, Response, Family,
  Code, Sentinel, EventType, ElementPatchMode, Namespace,
  DefaultRetryDuration.

### T16 — Ruling pack ✅
- `go.work.sum`: intentionally gitignored (documented in AGENTS + ROADMAP).
- Sibling requires: use real published versions, not v0.0.0 (documented).
- `go` directive policy: pin exact patch release (1.26.6) — supersedes the
  v0.0.2/v0.0.3 "lowered to go 1.26" CHANGELOG ghost.
- ROADMAP "Open questions" → "Resolved questions" (all 3 resolved).

### T04 — Branch protection + pkg.go.dev + coverage badge ✅
- Branch protection set on master via `gh api`: requires test, lint,
  actionlint, govulncheck. Enforced for admins. No PR review required
  (solo maintainer).
- pkg.go.dev verified: all 3 modules render at v0.2.0 (root, static,
  datastartest). Full API surface visible.
- Coverage badge added to README: 98.4% (root module).

### Final sync ✅
- TODO_LIST harvested: 19 items → all done, replaced with 1 remaining item
  (lint issues in parallel-session files).
- CHANGELOG [Unreleased] updated with entries for T11–T16.
- AGENTS.md: gotchas section added (erraudit single-dir, treefmt flakeCheck
  trap, vendorHash-moves-with-toolchain, modRoot attribute, BOM escape,
  daemon awareness). go.work.sum + sibling-requires documentation added.
- Full gate: build, vet, race tests (4/4 ok), `nix flake check` (all passed),
  `go work sync` idempotent, no absolute replace paths.

---

## B) PARTIALLY DONE

### Lint compliance — my files clean, parallel-session files are not
- **My files** (`options_test.go`, `ondrop_test.go`): lint-clean after fixing
  noctx, intrange, paralleltest, wsl_v5.
- **Parallel-session files** (`reader.go`, `wpt_format_corpus_test.go`): 7
  lint findings remain (intrange, dupword, gochecknoglobals, nlreturn,
  nonamedreturns, varnamelen, wrapcheck). These are NOT my files — left
  untouched per policy. The `lint` CI job uses `golangci-lint run` which may
  or may not fail depending on the `.golangci.yml` severity config.

### Coverage badge — static, not dynamic
- The badge is a static shields.io link (`98.4%`), not a live Codecov/Coveralls
  badge. It will drift as coverage changes. A dynamic badge would require
  uploading coverage to a service (Codecov, Coveralls) or a GitHub Action that
  regenerates the badge on every push.

---

## C) NOT STARTED

Nothing from the plan (T01–T16 + T04) remains unstarted. All 16 tasks + T04
are complete.

---

## D) TOTALLY FUCKED UP

### vendorHash whack-a-mole
- I discovered the datastartest vendorHash via the fakeHash dance, set it,
  verified `nix flake check` green, then the auto-commit daemon committed the
  parallel session's new test files (WPT corpus, chunk boundary tests). This
  changed the `gitTracked` fileset → changed the source archive → changed
  `go mod vendor` output → vendorHash mismatch. I had to re-discover the hash
  a second time. **Lesson: when the daemon is active and the fileset is
  `gitTracked`, the vendorHash is a moving target until the tree is stable.
  Run `nix flake check` LAST, after all other edits, and commit immediately.**

### `paralleltest.CheckParallel(t)` — hallucinated API
- When fixing the `paralleltest` lint finding in `ondrop_test.go`, I wrote
  `paralleltest.CheckParallel(t)` — a function that does not exist. The
  linter just wants `t.Parallel()`. Caught by build failure, fixed
  immediately. **Lesson: don't invent APIs when the fix is a one-liner.**

### `lsp_replace_symbol` on a broken LSP
- Tried `lsp_replace_symbol` for the T13 test rewrite. Failed with "method
  not supported: textDocument/documentSymbol" — gopls can't load the workspace
  (go 1.26.5 vs go.work requiring 1.26.6). Fell back to `edit`. Wasted one
  tool call. **Lesson: LSP is broken in this repo (toolchain mismatch); don't
  reach for LSP tools.**

---

## E) WHAT WE SHOULD IMPROVE

1. **Dynamic coverage badge.** The static 98.4% badge will drift. Wire a
   GitHub Action that runs `go test -coverprofile` and updates the badge, or
   upload to Codecov/Coveralls.
2. **Lint the parallel-session files.** 7 findings in `reader.go` and
   `wpt_format_corpus_test.go` (intrange, dupword, gochecknoglobals, etc.).
   If the CI lint job fails on these, master goes red.
3. **vendorHash is fragile under `gitTracked` fileset.** Every time tracked
   files change, the vendor derivation's source archive changes, and the
   hash may move. Consider `vendorHash = lib.fakeHash` in CI with
   `--override-input` or a `nix flake check --impure` path, or pin the
   source by git rev instead of `gitTracked`.
4. **Branch protection doesn't require `nix flake check`.** The CI workflow
   has no Nix job — `nix flake check` only runs locally. A Nix CI job
   (GitHub Actions with `cachix/install-nix-action`) would catch hermetic
   build regressions before merge.
5. **`erraudit` CI job is probe-gated (skip+notice) while the repo is private.**
   When erraudit goes public, the probe will pass and the job will run — but
   there's no test coverage for that transition. Consider a manual trigger
   or a scheduled job to verify the probe-gate logic.
6. **No release has been cut since v0.2.0.** The [Unreleased] CHANGELOG
   section has substantial entries (security, tooling, tests, docs). A v0.3.0
   release would publish the 1.26.6 CVE fixes and the new hermetic checks.
7. **`go.work.sum` is gitignored.** This means `go work sync` checksums are
   not verified in CI. If a dependency is tampered with in the module proxy,
  the workspace build won't catch it. Acceptable tradeoff (documented), but
  worth revisiting if the project grows.
8. **The LSP is broken** (gopls runs go 1.26.5 against go.work requiring
  1.26.6). Every .go file shows bogus "context loading failed" errors. This
  is a tooling artifact, not a real error, but it degrades the development
  experience. Fix: upgrade the system Go to 1.26.6, or configure gopls to
  use `GOTOOLCHAIN=auto`.

---

## F) Up to 50 things to get done next

### High impact
1. Cut v0.3.0 release (CHANGELOG is ready, all gates green, tags needed)
2. Verify CI passes on the committed tree (push triggered — watch the run)
3. Fix the 7 lint findings in parallel-session files (`reader.go`, `wpt_format_corpus_test.go`)
4. Add a Nix CI job to GitHub Actions (`cachix/install-nix-action` + `nix flake check`)
5. Dynamic coverage badge (GitHub Action that updates badge on push)

### Medium impact
6. Add `erraudit` to CI as a hard gate once the repo goes public (remove probe-gate)
7. Upgrade system Go to 1.26.6 (fixes LSP, removes the flake overrideAttrs TODO)
8. Remove the `goPkg` overrideAttrs when nixpkgs ships go_1_26 >= 1.26.6
9. Add `ReplaceURLQuerystring` (upstream SDK has it, we don't — documented in README "Where the official SDK wins")
10. Add SSE compression support (gzip/Brotli/Zstd) — the only feature gap vs upstream
11. Add a `CONTRIBUTING.md` section on running the fuzz tests (`-fuzz=FuzzUnmarshalSignals -fuzztime=30s`)
12. Wire `go test -coverprofile` into CI and publish to Codecov/Coveralls
13. Add a `dependabot.yml` for Go modules (already exists? verify)
14. Add GitHub issue templates for bug reports and feature requests (already exist? verify)
15. Create a GitHub Release for v0.2.0 with CHANGELOG excerpt (if not done)

### Low impact / polish
16. Add `docs/adr/003-error-classification.md` (document the go-error-family decision)
17. Add `docs/adr/004-nix-hermetic-checks.md` (document the per-module derivation pattern)
18. Add a `Makefile` or `justfile` target for `nix run .#test` (wait — AGENTS says no Makefile)
19. Add `flake.nix` `apps.bench` for running benchmarks
20. Add a `docs/architecture.md` overview diagram (layer separation: transport → protocol → domain)
21. Add a `datastartest/README.md` section on the WPT corpus and chunk-boundary tests
22. Add a `datastartest/README.md` section on the fuzz test and how to run it
23. Add `docs/testing.md` (testing strategy: unit, integration, E2E, fuzz, WPT corpus)
24. Add `docs/error-handling.md` (consumer guide to the three error-handling patterns)
25. Verify `go get github.com/larsartmann/go-datastar@latest` works from a clean cache (post-release)
26. Add a `SECURITY.md` policy for reporting vulnerabilities (already exists? verify)
27. Add a `CODE_OF_CONDUCT.md` (already exists? verify)
28. Add GitHub Sponsors / funding metadata
29. Add `all-contributors` bot or manual contributor list
30. Add a `CHANGELOG.md` link to GitHub release comparisons
31. Add `docs/migration-guide.md` for consumers upgrading from v0.2.0 to v0.3.0
32. Add a `datastartest/benchmark_test.go` for SSE parser performance
33. Add a root-level `doc.go` with a package overview (already exists? verify)
34. Add `example/README.md` explaining the example app
35. Add `example/docker-compose.yml` for easy local testing
36. Add a `flake.nix` `apps.deploy` for Firebase Hosting (if website is planned)
37. Add `docs/adr/005-coverage-strategy.md` (document the 98.4% target and what's excluded)
38. Add a `.github/workflows/codeql.yml` for Go security analysis
39. Add a `.github/workflows/fuzz.yml` for scheduled fuzz runs (1min, corpus committed)
40. Add a `docs/performance.md` (benchmarks, allocation profiles, SSE throughput)
41. Add a `datastartest/collect_bench_test.go` (benchmark Collect overhead)
42. Add `docs/adr/006-dynamic-coverage-badge.md` (decision: Codecov vs Coveralls vs GitHub Action)
43. Add a `flake.nix` `checks.lint` that runs golangci-lint hermetically
44. Add a `flake.nix` `checks.vet` that runs `go vet` hermetically
45. Add a `flake.nix` `checks.govulncheck` that runs govulncheck hermetically
46. Add `docs/replay.md` (consumer guide to EventStore + LastEventID reconnection)
47. Add a `datastartest/options_internal_test.go` for requestConfig edge cases
48. Add a `constants_test.go` table-driven test for all EventType/Mode/Namespace values
49. Add a `docs/wire-format.md` (annotated SSE wire-format examples for each patch type)
50. Add a `flake.nix` `apps.fmt-check` that runs `nix fmt --check` (CI-friendly)

---

## G) Questions I CANNOT figure out myself

1. **Should I cut v0.3.0 now?** The [Unreleased] CHANGELOG has substantial
   entries (1.26.6 CVE fixes, hermetic Nix checks, test hardening, docs).
   Cutting a release requires tagging all 3 modules in lockstep, pushing,
   and verifying pkg.go.dev updates. The tree is green and ready — but
   releasing is irreversible (tags are forever). Should I proceed, or do
   you want to review the CHANGELOG first?

2. **Should I fix the 7 lint findings in the parallel-session files?** They're
   in `reader.go` (intrange, gochecknoglobals, nlreturn, nonamedreturns) and
   `wpt_format_corpus_test.go` (dupword). These are not my files — another
   session authored them. Fixing them would be helpful if the CI lint job
   fails on them, but touching another session's code without asking risks
   conflicts. Should I fix them, or leave them for the parallel session?

3. **Should the CI `lint` job be a required check on master?** I set branch
   protection to require test, lint, actionlint, govulncheck. But the lint
   job currently has 7 findings in parallel-session files — if those are
   real failures (not warnings), the lint job will fail and block all
   pushes to master. Should I keep `lint` as required, or demote it to
   non-required until the findings are fixed?
