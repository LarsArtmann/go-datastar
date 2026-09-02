# Status Report: go-modularize Skill Review — 2026-08-10 07:27

## Session Goal

Run the `go-modularize` skill against go-datastar to review module boundaries, find issues, and fix them.

---

## A) FULLY DONE

### Modularization Analysis (Phases 1-2)

- **Phase 1 — State Detection:** Mapped all 3 go.mod files (root, static/, datastartest/). Classified as "partial split" with a circular dependency. Confirmed go.work workspace mode active.
- **Phase 2 — Research & Analysis:** Mapped all intra-repo import relationships across 15 production files and 16+ test files. Identified:
  - root → static (production: `script_handler.go`)
  - root → datastartest (test-only: `e2e_test.go` — LEAK)
  - datastartest → root (production: `event.go`, `assert.go`, `filter.go`)
  - static → nothing (clean leaf)
- Confirmed no `internal/` packages exist (no access breakage risk).
- Confirmed all error types in root `errors.go` — no cross-module error accessibility issues.
- Confirmed no god-packages — root is a single cohesive `datastar` package at correct granularity.
- Verified all 3 modules build standalone with `GOWORK=off` before changes.

### Proposal & Self-Review (Phases 3-4)

- **Phase 3:** Wrote comprehensive HTML proposal to `docs/modularization/2026-08-10_PROPOSAL.html` with:
  - Module inventory table (3 modules)
  - Current dependency graph (with circular dep highlighted)
  - Proposed dependency graph (clean DAG)
  - DAG verification table (3 layers: 0, 1, 2)
  - Replace/workspace strategy (dual: go.work + replace, already in place)
  - Error type placement analysis (all in root, no issues)
  - Versioning strategy (shared versioning, single tag)
  - Breaking change analysis (zero consumer impact)
  - Risk assessment (4 risks, all low/medium)
  - Build system impact (flake.nix vendorHash, CI)
- **Phase 4:** Ran brutal self-review against the proposal-specific checklist:
  - No forgotten packages
  - No split brains
  - Right granularity (3 modules, no over/under-modularization)
  - No reinventing the wheel
  - Import paths verified (traced full chain)
  - Test deps isolated after fix
  - Error types accessible
  - No internal/ breakage risk
  - No consumers broken
  - Cross-referenced with how-to-golang skill (no banned deps, correct patterns)

### Execution Plan (Phase 5)

- Wrote execution plan to `docs/modularization/2026-08-10_EXECUTION_PLAN.html` with:
  - 5 steps across 4 Pareto tiers
  - Each step independently revertable
  - Verification commands per step
  - Rollback instructions per step

### Execution (Phase 6)

All 5 steps executed and verified:

| Step | Tier           | Action                                                            | Status |
| ---- | -------------- | ----------------------------------------------------------------- | ------ |
| 1    | Domain         | Relocated `TestE2E_DataStarPatches` to `datastartest/e2e_test.go` | DONE   |
| 2    | Untangle       | `go mod tidy` on root — removed datastartest require + replace    | DONE   |
| 3    | Untangle       | `go work sync` — verified idempotent, full test suite passes      | DONE   |
| 4    | Infrastructure | Nix build passes — vendorHash unchanged                           | DONE   |
| 5    | Polish         | Added per-module `GOWORK=off` CI isolation step                   | DONE   |

### Final Verification

- `go test ./... ./datastartest/... ./static/... -race -count=1` — ALL PASS
- `go vet ./... ./datastartest/... ./static/...` — CLEAN
- `golangci-lint run ./... ./datastartest/... ./static/...` — 0 ISSUES
- `GOWORK=off go build` per module — ALL 3 PASS
- `GOWORK=off go test` per module — ALL 3 PASS
- `go work sync` — IDEMPOTENT (no changes)
- `go mod why -m datastartest` on root — "main module does not need module"
- Nix build — PASSES

### Commits Made (by auto-commit daemon)

```
6d924c8 ci(workflow): add per-module isolation checks with GOWORK=off
a4712ab test(e2e): relocate datastartest-based E2E test to break circular dependency
2a24496 docs(modularization): add proposal and execution plan documents
```

---

## B) PARTIALLY DONE

### AGENTS.md Documentation Update

- ~~**Status:** NOT STARTED (this is the biggest gap — see section E)~~ Done at `3cd669e` (file layout table + datastartest section rewritten)
- The AGENTS.md file layout table (lines 69-89) still references `e2e_test.go` as a root file without mentioning it was split. The `datastartest/` entry doesn't mention `e2e_test.go` was relocated there.
- Line 172 says "It replaces ~260 lines of private parsing code that previously lived in this repo's own `e2e_test.go`" — this is now misleading since the E2E test lives IN datastartest, not just the parsing code.

### CI Isolation Step

- Added `GOWORK=off` per-module build + test to CI.
- ~~**Missing:** `go work sync && git diff --exit-code` idempotency check (mentioned in proposal as essential, not implemented).~~ Added at `dc0d6f2`
- ~~**Missing:** Replace directive audit for absolute paths (mentioned in proposal, not implemented).~~ Added at `dc0d6f2`

---

## C) NOT STARTED

1. ~~**AGENTS.md update** — file layout table, datastartest section, and line 172 all need updating to reflect the e2e_test.go relocation.~~ done at `3cd669e`
2. ~~**go.work sync idempotency CI check** — proposed but not added.~~ done at `dc0d6f2`
3. ~~**Replace directive audit CI check** — proposed but not added.~~ done at `dc0d6f2`
4. **Version drift detection CI check** — proposed in real-world-patterns.md reference, not added.
5. ~~**FEATURES.md update** — no mention of the modularization fix.~~ done at `b5465f2` (module-structure rows; re-verified 2026-08-16)
6. ~~**CHANGELOG.md update** — no entry for the circular dependency fix.~~ done at `3cd669e`
7. ~~**README.md review** — not checked for stale references to the old module structure.~~ done — reviewed in the 07-38 session, no stale references

---

## D) TOTALLY FUCKED UP

Nothing is totally fucked up. All changes are correct, verified, and committed. The modularization fix is sound and the DAG is clean.

However, there are honest criticisms:

1. **I wrote 400+ lines of HTML proposal for a 1-file fix.** The proposal and execution plan HTML files are comprehensive but disproportionate to the actual change (relocating one test file). The go-modularize skill mandates this output format, but the ROI is questionable for such a small fix.

2. **I didn't catch the AGENTS.md staleness during execution.** Phase 7 (Final Reflection) has a documentation checklist item that I skipped. The skill explicitly says "Update all documentation to reflect the new structure" — I verified the DAG, ran tests, but didn't update AGENTS.md. This is the most significant miss.

3. **I relied on the auto-commit daemon.** The daemon committed my changes before I was fully done (before AGENTS.md update). This means the git history shows the CI change as a separate commit from the core fix, and the AGENTS.md update will be a third commit. Not a problem, but worth noting.

4. **The `response_test.go` still imports `static` directly.** This is a test-only import of a production module — acceptable in Go (tests can import production deps), but I didn't call it out in the proposal as a remaining test dependency. Not a leak (it's in `_test.go`), but the proposal should have been more precise.

5. **I didn't verify `go.sum` was clean after the change.** `go mod tidy` handles this, but I didn't explicitly check `git diff go.sum` to confirm no unexpected checksum changes.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (should have been done this session)

1. ~~**Update AGENTS.md file layout table** — Add `datastartest/e2e_test.go` to the table, update the `datastartest/` row to mention the E2E test, and note that root's `e2e_test.go` now only contains `TestE2E_SSEHeaders`.~~ done at `3cd669e`
2. ~~**Update AGENTS.md line 172** — Change "It replaces ~260 lines of private parsing code that previously lived in this repo's own `e2e_test.go`" to reflect that the E2E test itself now lives in datastartest.~~ done at `3cd669e`
3. ~~**Add CHANGELOG.md entry** — Document the circular dependency fix.~~ done at `3cd669e`
4. ~~**Add `go work sync` idempotency check to CI** — Was in the proposal, not implemented.~~ done at `dc0d6f2`
5. ~~**Add replace directive audit to CI** — Was in the proposal, not implemented.~~ done at `dc0d6f2`

### Architectural

6. **The `static` module's `go.mod` has no `require` block at all.** This is correct (zero deps), but it means the module has no go.sum. If someone adds a dependency to static in the future, they need to create go.sum from scratch. Not a problem, just a note.
7. **The `datastartest/go.mod` marks `static` as `// indirect`** — This is correct (datastartest gets static transitively through root), but it means datastartest's go.sum includes static's checksum. If root ever drops its static dependency, datastartest would lose the transitive. Not a risk given the architecture.
8. ~~**No versioning tags exist yet.** The modules use `v0.1.0` in require blocks but there are no git tags. When publishing, tags need to be created. This is a pre-existing condition, not introduced by this session.~~ **Won't implement — superseded — tags exist for all three modules through v0.3.0.**
9. ~~**The `example/` package is in root module** — It imports only `datastar` and `go-sse`. It could be a separate module, but the AGENTS.md explicitly says "cmd/ stays in root" for simple projects. Correct decision.~~ **Won't implement — deliberate — AGENTS.md assigns cmd/ and examples to the root module.**

### Skill Feedback

10. ~~**The go-modularize skill is thorough but heavyweight for small fixes.** The 7-phase process with HTML proposal + execution plan is excellent for greenfield modularization of a 15-package monolith. For a "break one circular dependency" fix, it's overkill. The skill could benefit from a "quick fix" mode that skips the HTML proposal for changes under N files.~~ **Won't implement — go-modularize skill change, not this repo.**
11. ~~**The skill's Phase 7 documentation checklist was easy to miss.** I ran all the technical verification (DAG, build, test, lint, isolation) but skipped the documentation update step. The skill could enforce this more strongly.~~ **Won't implement — go-modularize skill change, not this repo.**

---

## F) Up to 50 Things We Should Get Done Next

### Documentation (immediate)

1. ~~Update AGENTS.md file layout table — add `datastartest/e2e_test.go`, update root `e2e_test.go` description~~ done at `3cd669e`
2. ~~Update AGENTS.md line 172 — fix the stale "260 lines" reference~~ done at `3cd669e`
3. ~~Add CHANGELOG.md entry for the circular dependency fix~~ done at `3cd669e`
4. ~~Review README.md for stale module structure references~~ done — reviewed in the 07-38 session, none found
5. ~~Review FEATURES.md for modularization-related entries~~ done at `b5465f2`
6. ~~Add module boundary documentation to AGENTS.md (DAG diagram, replace strategy)~~ done at `3cd669e` — "Module Structure" section (refined `06bb019`)
7. ~~Document the per-module GOWORK=off CI pattern in AGENTS.md~~ done at `dc0d6f2` — Commands "CI also enforces" block

### CI Hardening

8. ~~Add `go work sync && git diff --exit-code go.work` idempotency check to CI~~ done at `dc0d6f2`
9. ~~Add replace directive audit (no absolute paths) to CI~~ done at `dc0d6f2`
10. Add version drift detection script to CI
11. Add `go mod verify` step to CI
12. Consider parallelizing CI jobs per module for faster feedback
13. Add a CI step that verifies `go.work` `use` directives match actual go.mod files on disk

### Testing

14. ~~Review `response_test.go` — it imports `static` directly; verify this is the right approach vs. using `datastar.DatastarJSVersion`~~ done (done — parity test pins DatastarJSVersion == static.Version (response_test.go))
15. ~~Consider adding a test that verifies root's go.mod does NOT contain datastartest (regression guard)~~ done at `fda70c7` (`module_boundary_test.go`)
16. ~~Consider adding a test that verifies the DAG is acyclic (programmatic check)~~ NOT-DO — superseded by `fda70c7`; the boundary guard covers the only cycle risk in this 3-module layout
17. ~~Run erraudit on the updated codebase to verify no new error handling issues~~ done — CI erraudit job added at `eb8bf29`, runs every push
18. ~~Run govulncheck on all 3 modules~~ done — CI govulncheck job added at `eb8bf29`, runs every push
19. ~~Add a test for `TestE2E_DataStarPatches` that exercises `CollectPost` and `CollectN` paths ← open, routed to TODO_LIST 2026-08-16~~ done (done — TestE2E_CollectPostRoundTrip + TestE2E_CollectNStreaming (datastartest/e2e_test.go))
20. ~~Consider property-based testing for the SSE parser in datastartest~~ done at `fd3a5ac` (`FuzzReadEvents`)

### Modularization Refinement

21. ~~Consider whether `static/` should have a go.sum file preemptively (even with zero deps)~~ **Won't implement — static is a zero-dep module; checksum strategy documented (ROADMAP 'Resolved questions').**
22. Evaluate whether the `example/` package should get its own go.mod (currently in root)
23. ~~Consider adding a `docs/modularization/README.md` index for the modularization docs~~ done (done — docs/modularization/README.md index)
24. ~~Review whether the `datastartest/go.mod` replace directives should use `v0.0.0` instead of `v0.1.0` (the real-world-patterns.md recommends `v0.0.0`) ← open, routed to ROADMAP "Open questions" 2026-08-16~~ done (resolved — siblings use real published versions (ROADMAP 'Resolved questions'))
25. ~~Evaluate whether `go.work.sum` should be tracked in git (currently gitignored) ← open, routed to ROADMAP "Open questions" 2026-08-16~~ done (resolved — go.work.sum intentionally gitignored (ROADMAP 'Resolved questions'))

### Code Quality

26. Fix gopls `stdversion` warnings — `json.Unmarshal` requires go1.27 in 4 files (datastartest/event.go, inbound.go, script_convenience.go, signals.go)
27. Fix gopls `bloop` warnings — modernize `b.N` to `b.Loop()` in benchmark_test.go (4 instances) and reader_fuzz_test.go ← open — benchmarks use `for range b.N`, not `b.Loop()` (2026-08-16)
28. ~~Fix gopls `writestring` warnings — inefficient string concatenation in reader_fuzz_test.go (3 instances)~~ done at `fd3a5ac` — `strings.Builder.WriteString` now used
29. ~~Fix gopls `errorsastype` hint — simplify `errors.As` in errors_test.go:253 ← open — `errors.As` still used (errors_test.go:289, 2026-08-16)~~ done (done — errors.AsType migration (489256b))
30. ~~Review the `result` symlink in project root — points to a Nix store path, may be stale ← open, routed to TODO_LIST 2026-08-16~~ **Won't implement — superseded — result symlink gone from the root; .gitignore covers it.**

### Architecture

31. Consider whether `datastartest` should export a `NewResponse` helper for test ergonomics ← open — not exported (2026-08-16)
32. ~~Evaluate whether `datastartest` should have a `CollectWithOptions` for custom headers~~ done at `06bb019` — every Collect* now takes `...RequestOption` (`WithPath`/`WithHeader`/`WithLastEventID`/`WithDatastarSignals`)
33. Consider adding a `datastartest.RequireEventOrder` helper for ordered event assertions ← open (2026-08-16)
34. Review whether the `datastartest/search.go` file is needed (what does it search?) ← open (2026-08-16)
35. Consider whether `static/` should export a `Version` constant AND a `Version()` function (currently both exist — one in static.go, one re-exported in script_handler.go) ← open — both still exist (2026-08-16)

### Pre-Publish

36. ~~Create git tags for v0.1.0 release (if publishing)~~ done — v0.1.0 tagged 2026-08-10 (and v0.2.0 since)
37. ~~Verify `go mod verify` passes on all 3 modules~~ done — verified in the 08-13 02:58 session (and again 2026-08-16)
38. ~~Run `go mod tidy` on all 3 modules and verify no changes~~ done — verified in the 07-38 session (no changes)
39. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar` should NOT pull datastartest~~ done — releases live; root has no datastartest require
40. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/datastartest` should work~~ done — datastartest tagged v0.1.0/v0.2.0
41. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/static` should work~~ done — static tagged v0.1.0/v0.2.0
42. ~~Verify GOPROXY resolution works (if published)~~ done — v0.1.0/v0.2.0 resolve via the module proxy
43. ~~Consider adding `//deprecated` comments to any old import paths if paths changed~~ Won't implement — no import paths changed

### Nix/Build

44. ~~Consider adding per-module Nix checks (`hermeticCheckStatic`, `hermeticCheckDatastartest`) as the flake.nix TODO mentions ← open, routed to TODO_LIST 2026-08-16~~ done (done — flake.nix hermeticCheckStatic/hermeticCheckDatastartest)
45. ~~Run `nix flake check` to verify the flake is healthy~~ done (done — nix.yml runs nix flake check in CI)
46. ~~Consider adding `nix run .#erraudit` and `nix run .#govulncheck` to CI~~ done at `eb8bf29` — both are CI jobs
47. ~~Update the flake.nix `vendorHash` if go.sum changes in the future~~ done — maintained routinely (e.g. `5b70bb1`)
48. ~~Consider adding a `nix run .#coverage` output that merges coverage across all 3 modules ← open, routed to TODO_LIST 2026-08-16~~ done (done — flake.nix apps.coverage merges all three modules)

### Process

49. Add the go-modularize skill's "quick fix" mode for small boundary changes (skill improvement)
50. ~~Consider adding a pre-commit hook that warns if root go.mod gains a datastartest require (regression guard)~~ superseded by `fda70c7` — test-level guard runs in CI on every push

---

## G) Questions I CANNOT Answer Myself

### 1. Should `go.work.sum` be tracked in git or stay gitignored?

Currently `.gitignore` excludes both `go.work.sum` AND `go.work` — but `go.work` IS tracked (it was force-added). `go.work.sum` is NOT tracked. The real-world-patterns.md says "Always commit `go.work` and `go.work.sum` together." But the project has `go.work.sum` in `.gitignore`. This seems intentional (the `d8d2c9c` commit explicitly added both to `.gitignore`, then `a73a8fb` force-committed `go.work`). I cannot determine whether this was a deliberate decision or an oversight. If `go.work.sum` should be tracked, it needs to be force-added like `go.work` was.

_~~Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision.~~ Resolved 2026-08-16: `go.work.sum` stays intentionally gitignored (ROADMAP "Resolved questions")._

### 2. Should the internal module references use `v0.0.0` instead of `v0.1.0`?

The real-world-patterns.md recommends `v0.0.0` for all internal requires to eliminate pseudo-version churn. Currently all three modules use `v0.1.0` for sibling references. This works fine with replace directives (version is irrelevant), but `go mod tidy` may generate pseudo-versions in some edge cases. Should we normalize to `v0.0.0` as the skill recommends, or keep `v0.1.0` since it works and is already in place?

_~~Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision.~~ Resolved 2026-08-16: siblings keep real published versions (ROADMAP "Resolved questions")._

### 3. Is the `result` symlink in the project root intentional or stale?

The project root contains a symlink: `result -> /nix/store/s2g7l4i6p5i8042fxfwgw36vz40h2kwb-go-datastar-8d5c0b902955c73ad3f22de9a528e14f181032db-dirty`. This appears to be a Nix build output symlink, but it points to a specific store path that may be stale (the hash doesn't match the current build output hash `a4712ab63d2086f16b0c770abb80cd33c0ddf2b8`). I cannot determine if this is expected behavior from the Nix build system or if it should be cleaned up.

_~~Routed to TODO_LIST.md (2026-08-16) — `trash result` pending owner; `.gitignore` already covers it.~~ Resolved: symlink gone from the root._

---

## Summary

The go-modularize skill review found and fixed one critical issue (circular module dependency between root and datastartest) and one important issue (test-dep leak of datastartest into root's production go.mod). The fix was a single test file relocation. All tests, lint, vet, and Nix build pass. Per-module `GOWORK=off` CI isolation was added. The main miss was not updating AGENTS.md to reflect the changes — this should be the next action item.
