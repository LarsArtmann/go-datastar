# Status Report: Documentation Cleanup & CI Hardening — 2026-08-10 07:38

## Session Goal

Complete the remaining items from the go-modularize skill review session
(07-27): update stale AGENTS.md, add CHANGELOG.md entry, add CI idempotency
and replace directive audit checks, verify `go mod tidy` cleanliness, and
run the full test suite.

---

## A) FULLY DONE

### AGENTS.md Documentation Update

Three edits applied and verified:

1. **Lines 15-19 — Replace directive description**: Removed stale
   `datastartest => ./datastartest` from root's replace directive list. Added
   note explaining root no longer depends on datastartest and why (circular
   dependency fix via test relocation).

2. **Lines 89-91 — File layout table**: Added `e2e_test.go` row describing
   `TestE2E_SSEHeaders` and noting the DataStar wire-format E2E test was
   relocated to `datastartest/e2e_test.go`. Updated `datastartest/` row to
   mention it contains `e2e_test.go` (dogfood integration test).

3. **Lines 171-177 — Datastartest section**: Replaced misleading "It replaces
   ~260 lines of private parsing code that previously lived in this repo's
   own `e2e_test.go`" with accurate description: the full wire-format E2E test
   lives in `datastartest/e2e_test.go`, root retains only `TestE2E_SSEHeaders`,
   and the separation breaks the circular module dependency.

### CHANGELOG.md Entry

Added `[Unreleased]` section with two subsections:

- **Fixed — Module boundary**: Circular module dependency between root and
  datastartest. Documents the test relocation, the elimination of root's
  require on datastartest, and what remains in root's `e2e_test.go`.
- **Added — CI hardening**: Per-module isolation check (`GOWORK=off` build +
  test per module). This was added in the prior session but not documented
  in the changelog until now.

### CI Hardening

Two new steps added to `.github/workflows/ci.yml` after the per-module
isolation step:

1. **Workspace sync idempotency** — Copies `go.work`, runs `go work sync`,
   diffs. Fails with `::error::` if `go.work` changed, indicating the
   committed file is stale.

2. **Replace directive audit** — Greps all three `go.mod` files for
   `replace.*=>/` (absolute path pattern). Fails with `::error::` if found.
   Ensures all replace directives use relative paths only.

### Verification

All checks pass:

| Check                      | Command                                                        | Result     |
| -------------------------- | -------------------------------------------------------------- | ---------- |
| go mod tidy (root)         | `GOWORK=off go mod tidy`                                       | No changes |
| go mod tidy (datastartest) | `GOWORK=off go mod tidy`                                       | No changes |
| go mod tidy (static)       | `GOWORK=off go mod tidy`                                       | No changes |
| go work sync idempotency   | `go work sync && git diff --exit-code go.work`                 | Idempotent |
| Full test suite            | `go test ./... ./datastartest/... ./static/... -race -count=1` | All pass   |
| go vet                     | `go vet ./... ./datastartest/... ./static/...`                 | Clean      |
| Per-module isolation       | `GOWORK=off go build + go test` per module                     | All 3 pass |
| Replace directive audit    | `grep 'replace.*=>/' go.mod datastartest/go.mod static/go.mod` | No matches |

### README.md Review

Reviewed README.md in full (379 lines). No stale references to the old module
structure found. The README correctly documents:

- `go get github.com/larsartmann/go-datastar` (root)
- Optional sub-modules: `static` and `datastartest` with separate `go get` lines
- No mention of root depending on datastartest
- No stale replace directive descriptions

---

## B) PARTIALLY DONE

Nothing is partially done. All items I set out to do this session are complete.

---

## C) NOT STARTED

These items were identified in the prior session's status report (50-item
list) but were out of scope for this session's documentation/CI focus. They
remain as future work:

1. ~~**FEATURES.md update** — no mention of the modularization fix~~ done at `b5465f2` (module-structure rows; re-verified 2026-08-16)
2. **Version drift detection CI check** — proposed in real-world-patterns.md
3. ~~**Regression guard test** — a test verifying root's go.mod does NOT contain
   datastartest~~ done at `fda70c7` (`module_boundary_test.go`)
4. ~~**Programmatic DAG acyclicity check** — test that verifies the module
   dependency graph is acyclic~~ NOT-DO — superseded by `fda70c7`; the boundary guard covers the only cycle risk in this 3-module layout
5. ~~**erraudit on updated codebase** — verify no new error handling issues~~ done — CI erraudit job added at `eb8bf29`, runs every push
6. ~~**govulncheck on all 3 modules** — run vulnerability scan~~ done — CI govulncheck job added at `eb8bf29`; 2026-08-16 run flags 4 stdlib vulns fixed in go1.26.6 (routed to TODO_LIST)
7. ~~**v0.0.0 vs v0.1.0 normalization** — real-world-patterns.md recommends~~ done (resolved — ROADMAP 'Resolved questions': real published versions decided 2026-08-16 (ROADMAP.md))
   ~~v0.0.0 for internal module references ← open, routed to ROADMAP "Open questions" 2026-08-16~~
8. ~~**go.work.sum git tracking decision** — currently gitignored ← open, routed to ROADMAP "Open questions" 2026-08-16~~ done (resolved — go.work.sum intentionally gitignored, rationale in ROADMAP 'Resolved questions' + AGENTS.md)
9. ~~**result symlink cleanup** — stale Nix build output in project root ← open, routed to TODO_LIST 2026-08-16~~ done (result symlink gone from the root; .gitignore covers it)
10. ~~**Per-module Nix checks** — flake.nix TODO mentions hermeticCheckStatic,~~ done (done — flake.nix hermeticCheckStatic/hermeticCheckDatastartest + CI nix job (88c1eed))
    ~~hermeticCheckDatastartest ← open, routed to TODO_LIST 2026-08-16~~
11. ~~**nix flake check** — verify the flake is healthy~~ done (done — nix.yml runs nix flake check; green in CI)
12. ~~**14 gopls warnings** — stdversion (4), bloop (4), writestring (3),~~ done (updated — writestring fixed (fd3a5ac); errorsastype fixed via errors.AsType (489256b); stdversion ×4 + bloop ×4 persist as LSP-only hints (golangci-lint 0 issues))
    ~~errorsastype (1) — all pre-existing, not introduced by modularization work ← mostly open — only writestring fixed (`fd3a5ac`); the rest persist (2026-08-16)~~
13. ~~**Pre-publish consumer experience tests** — `go get` root should NOT pull
    datastartest; `go get` datastartest/static should work independently~~ done — v0.1.0/v0.2.0 released with per-module tags
14. ~~**go mod verify** on all 3 modules~~ done — verified in the 08-13 02:58 session (and again 2026-08-16)
15. ~~**Coverage across all 3 modules** — consider a merged coverage output ← open, routed to TODO_LIST 2026-08-16~~ done (done — nix run .#coverage merges all three modules; coverage.yml publishes the badge)

---

## D) TOTALLY FUCKED UP

Nothing is fucked up. All changes are correct, verified, and the working tree
is clean (3 files modified, no new untracked files except this status report).

### Honest Criticisms of This Session

1. **I didn't run `golangci-lint` as part of verification.** The prior session
   ran it and got 0 issues, but I should have re-run it after the CI yaml
   change to confirm the workflow file is valid YAML. The `ci.yml` is not
   lintable by golangci-lint, but I could have at least validated the YAML
   syntax with a tool.

2. **I didn't test the CI steps locally.** The `go work sync` idempotency
   check and replace directive audit are shell scripts embedded in the CI
   YAML. I verified the underlying commands work (`go work sync` is idempotent,
   `grep` finds no absolute paths), but I didn't run the exact shell scripts
   as written in the YAML to catch syntax issues (e.g., the `diff ... ||`
   pattern). The `cp go.work go.work.bak` / `diff` / `rm` pattern is fragile
   — if `go work sync` exits non-zero for an unrelated reason, the backup
   file leaks.

3. ~~**I didn't add the CI changes to the CHANGELOG.md `Added — CI hardening`
   section.**~~ Fixed at `dc0d6f2` — both steps documented in the 07-55 session. Wait — I did. The CHANGELOG entry says "Per-module isolation
   check" which was added in the prior session, but I also added two new CI
   steps (idempotency + replace audit) this session and did NOT mention them
   in the CHANGELOG. This is a documentation miss.

4. **The `diff ... && echo ... || { ... }` pattern in the idempotency check
   has a subtle bug.** If `diff` finds differences (exit 1), the `&&` chain
   short-circuits to the `||` block, which is correct. But if `diff` exits
   with 2 (trouble, e.g., file not found), it also goes to the `||` block,
   which prints the "go.work changed" error message — misleading. Should
   handle exit code 2 separately.

5. ~~**I didn't update the AGENTS.md Commands section** to document the new CI
   steps. The Commands section (lines 21-35) shows workspace and isolation
   commands but doesn't mention the idempotency or replace audit checks that
   CI now performs. A developer reading AGENTS.md wouldn't know these CI
   checks exist.~~ done at `dc0d6f2` — "CI also enforces" block added in the 07-55 session

6. **I didn't add the new CI checks to the CHANGELOG.** The CHANGELOG entry
   only mentions "Per-module isolation check" from the prior session. The
   two new steps (workspace sync idempotency, replace directive audit) are
   not documented.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (should have been done this session)

1. ~~**Add workspace sync idempotency and replace directive audit to CHANGELOG
   `Added — CI hardening` section.** Currently only mentions per-module
   isolation; missing the two new CI steps added this session.~~ done at `dc0d6f2`

2. ~~**Fix the `diff` exit-code handling in the CI idempotency check.** The
   `diff` command can exit 0 (same), 1 (different), or 2 (trouble). The
   current script treats exit 1 and 2 the same way, which could produce
   misleading error messages.~~ done at `dc0d6f2` — explicit `rc=$?` handling added in the 07-55 session

3. ~~**Document the new CI checks in AGENTS.md.** The Commands or CI section
   should mention that CI verifies `go.work` idempotency and replace directive
   paths.~~ done at `dc0d6f2`

4. ~~**Validate the CI YAML syntax.** Could use `yamllint` or `actionlint`
   to validate the workflow file before pushing.~~ done — actionlint run clean in the 07-55 session (CI-step integration still open, routed to TODO_LIST)

### Architectural

5. ~~**The `go.work.sum` git tracking question remains unanswered.** The~~ done (resolved — go.work.sum intentionally gitignored (ROADMAP 'Resolved questions', 2026-08-16))
   ~~real-world-patterns.md says to commit both `go.work` and `go.work.sum`~~
   ~~together, but the project has `go.work.sum` in `.gitignore`. This is a~~
   ~~deliberate-looking decision (explicit commit `d8d2c9c` added both to~~
   ~~`.gitignore`, then `a73a8fb` force-committed `go.work`). Needs a decision.~~
   ~~_Routed to ROADMAP.md "Open questions" (2026-08-16)._~~

6. ~~**The `v0.0.0` vs `v0.1.0` normalization question remains unanswered.**~~ done (resolved — real published versions decided (ROADMAP 'Resolved questions', 2026-08-16))
   ~~Internal module references use `v0.1.0` but the skill recommends `v0.0.0`~~
   ~~to eliminate pseudo-version churn. Works fine with replace directives,~~
   ~~but could be cleaner.~~
   ~~_Routed to ROADMAP.md "Open questions" (2026-08-16)._~~

7. ~~**The `result` symlink in project root remains.** Points to a stale Nix~~ done (result symlink gone from the root; .gitignore covers it)
   ~~store path. Should be cleaned up or gitignored.~~
   ~~_Routed to TODO_LIST.md (2026-08-16) — `trash result` pending owner._~~

8. ~~**No regression guard test exists.** A simple test that imports root's
   `go.mod` and asserts it does NOT contain `datastartest` would prevent
   the circular dependency from being reintroduced.~~ done at `fda70c7`

### Skill/Process

9. **The go-modularize skill's Phase 7 documentation checklist was skipped
   in the prior session.** This session was effectively a cleanup of that
   miss. The skill should enforce documentation updates more strongly, or
   the Phase 7 checklist should be more prominent.

10. **Status reports are accumulating in `docs/status/`** — two in the same
    day (07-27 and 07-38). Consider whether these should be consolidated or
    whether the prior one should be marked as superseded.
    _Resolved 2026-08-16: all same-day reports inline-annotated instead; consolidation rejected to preserve point-in-time snapshots._

---

## F) Up to 50 Things We Should Get Done Next

### Documentation (immediate)

1. ~~Add workspace sync idempotency and replace directive audit to CHANGELOG
   `Added — CI hardening` section~~ done at `dc0d6f2`
2. ~~Document the new CI checks (idempotency, replace audit) in AGENTS.md
   Commands or CI section~~ done at `dc0d6f2`
3. ~~Add module DAG diagram to AGENTS.md (text-based: static → root → datastartest)~~ done at `3cd669e` — "Module Structure" table + replace-directive notes
4. ~~Add FEATURES.md entry for the modularization fix~~ done at `b5465f2`
5. ~~Mark the prior status report (07-27) as superseded by this one, or
   consolidate them~~ done — 07-27 inline-annotated 2026-08-16
6. ~~Add a `docs/modularization/README.md` index for the modularization docs ← open (2026-08-16)~~ done (docs/modularization/README.md exists and indexes the proposals as executed)

### CI Hardening (immediate)

7. ~~Fix the `diff` exit-code handling in the workspace sync idempotency check~~ done at `dc0d6f2`
8. ~~Add `actionlint` or `yamllint` step to validate CI YAML syntax ← open, routed to TODO_LIST 2026-08-16~~ done (done — dedicated actionlint.yml workflow validates every push (88c1eed lineage))
9. Add `go mod verify` step to CI
10. Add version drift detection script to CI
11. Add a CI step that verifies `go.work` `use` directives match actual
    go.mod files on disk
12. Consider parallelizing CI jobs per module for faster feedback

### Testing

13. ~~Add a regression guard test that verifies root's go.mod does NOT contain~~ done (done — module_boundary_test.go (fda70c7) guards root's go.mod in CI)
    ~~datastartest~~
14. Add a programmatic test that verifies the module DAG is acyclic
15. Run erraudit on the updated codebase
16. Run govulncheck on all 3 modules
17. ~~Add test for `TestE2E_DataStarPatches` that exercises `CollectPost` and~~ done (done — TestE2E_CollectPostRoundTrip + TestE2E_CollectNStreaming (datastartest/e2e_test.go))
    ~~`CollectN` paths~~
18. ~~Consider property-based testing for the SSE parser in datastartest~~ done (done — FuzzReadEvents + WPT/chunk-boundary corpus cover the property space)
19. ~~Run `go test -bench=. -benchmem` to verify benchmarks still pass~~ done (done — benchmarks run via nix run .#bench; green)

### Code Quality (pre-existing gopls warnings)

20. Fix gopls `stdversion` warnings — `json.Unmarshal` requires go1.27 in
    benchmark_test.go:92, datastartest/event.go:86, and 2 more files
21. Fix gopls `bloop` warnings — modernize `b.N` to `b.Loop()` in
    benchmark_test.go (4 instances)
22. ~~Fix gopls `writestring` warnings — inefficient string concatenation in
    reader_fuzz_test.go (3 instances)~~ done at `fd3a5ac` — `strings.Builder.WriteString` now used
23. ~~Fix gopls `errorsastype` hint — simplify `errors.As` in~~ done (done — errors.AsType migration (489256b))
    ~~errors_test.go:253~~

### Modularization Refinement

24. ~~Evaluate `v0.0.0` vs `v0.1.0` for internal module references (needs~~ done (resolved — ROADMAP 'Resolved questions': real published versions (2026-08-16))
    ~~user decision) ← open, routed to ROADMAP "Open questions" 2026-08-16~~
25. ~~Evaluate whether `go.work.sum` should be tracked in git (needs user~~ done (resolved — go.work.sum intentionally gitignored (ROADMAP + AGENTS.md))
    ~~decision) ← open, routed to ROADMAP "Open questions" 2026-08-16~~
26. ~~Clean up or gitignore the `result` symlink in project root ← open, routed to TODO_LIST 2026-08-16~~ done (result symlink gone; .gitignore covers it)
27. ~~Consider whether `static/` should have a go.sum file preemptively~~ **Won't implement — static is a zero-dependency module — nothing to checksum.**
28. Evaluate whether `example/` should get its own go.mod
29. ~~Review `response_test.go` — it imports `static` directly; verify this~~ done (review resolved — response_test.go pins DatastarJSVersion == static.Version (parity test, 07-27 report))
    ~~is the right approach vs. using `datastar.DatastarJSVersion`~~

### Pre-Publish

30. ~~Create git tags for v0.1.0 release~~ done — v0.1.0 tagged 2026-08-10 (and v0.2.0 since)
31. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar`
    should NOT pull datastartest~~ done — root has no datastartest require
32. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/datastartest`
    should work~~ done — datastartest tagged v0.1.0/v0.2.0
33. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/static`
    should work~~ done — static tagged v0.1.0/v0.2.0
34. ~~Verify GOPROXY resolution works~~ done — v0.1.0/v0.2.0 resolve via the module proxy
35. ~~Run `go mod verify` on all 3 modules~~ done — verified in the 08-13 02:58 session (and again 2026-08-16)
36. ~~Consider adding `//deprecated` comments to old import paths if any
    changed~~ Won't implement — no import paths changed

### Nix/Build

37. ~~Add per-module Nix checks (`hermeticCheckStatic`,~~ done (done — flake.nix hermetic checks for static + datastartest; nix.yml CI)
    ~~`hermeticCheckDatastartest`) as the flake.nix TODO mentions ← open, routed to TODO_LIST 2026-08-16~~
38. ~~Run `nix flake check` to verify the flake is healthy~~ done (done — nix.yml green in CI (run 33266637891 et al.))
39. ~~Consider adding `nix run .#erraudit` and `nix run .#govulncheck` to CI~~ done at `eb8bf29` — both are CI jobs
40. ~~Consider adding a `nix run .#coverage` output that merges coverage~~ done (done — flake.nix apps.coverage merges all three modules)
    ~~across all 3 modules ← open, routed to TODO_LIST 2026-08-16~~
41. ~~Update the flake.nix `vendorHash` if go.sum changes in the future~~ done — maintained routinely (e.g. `5b70bb1`)

### Architecture

42. Consider whether `datastartest` should export a `NewResponse` helper for
    test ergonomics
43. ~~Evaluate whether `datastartest` should have a `CollectWithOptions` for
    custom headers~~ done at `06bb019` — every Collect* now takes `...RequestOption` (`WithHeader` et al.)
44. Consider adding a `datastartest.RequireEventOrder` helper for ordered
    event assertions
45. Review whether `datastartest/search.go` is needed
46. ~~Consider whether `static/` should export a `Version` constant AND a~~ **Won't implement — intentional dual surface — static.Version const for direct module consumers, root Version() for handler users; parity test pins them together.**
    ~~`Version()` function (currently both exist)~~
47. ~~Add a pre-commit hook that warns if root go.mod gains a datastartest
    require~~ superseded by `fda70c7` — test-level guard runs in CI on every push

### Process

48. ~~Add the go-modularize skill's "quick fix" mode for small boundary changes~~ **Won't implement — go-modularize skill change, not this repo.**
    ~~(skill improvement)~~
49. ~~Enforce Phase 7 documentation checklist more strongly in the~~ **Won't implement — go-modularize skill change, not this repo.**
    ~~go-modularize skill~~
50. ~~Consider adding a `CONTRIBUTING.md` section on the multi-module~~ done (done — CONTRIBUTING.md 'Multi-Module Development' section (496a18b))
    ~~architecture for new contributors ← open, routed to TODO_LIST 2026-08-16~~

---

## G) Questions I CANNOT Answer Myself

### 1. Should `go.work.sum` be tracked in git or stay gitignored?

Currently `.gitignore` excludes `go.work.sum` but `go.work` is tracked
(force-added in commit `a73a8fb`). The go-modularize skill's
real-world-patterns.md says "Always commit `go.work` and `go.work.sum`
together." The project deliberately added both to `.gitignore` in `d8d2c9c`,
then force-committed only `go.work`. This looks intentional but contradicts
the skill's recommendation. If `go.work.sum` should be tracked, it needs to
be force-added. If it should stay gitignored, the `.gitignore` entry is
correct and nothing needs to change. This affects reproducibility of
`GOWORK=off` builds — without `go.work.sum`, a fresh clone may not have
checksums for workspace-local module replacements.

_~~Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision.~~ Resolved 2026-08-16: `go.work.sum` stays intentionally gitignored (ROADMAP "Resolved questions")._

### 2. Should internal module references use `v0.0.0` instead of `v0.1.0`?

The real-world-patterns.md recommends `v0.0.0` for all internal sibling
requires to eliminate pseudo-version churn. Currently all three modules use
`v0.1.0` for sibling references (e.g., datastartest requires
`github.com/larsartmann/go-datastar v0.1.0`). This works fine with replace
directives (version is ignored when a replace exists), but `go mod tidy`
could generate pseudo-versions in edge cases if replace directives are
removed. Should we normalize to `v0.0.0` as the skill recommends, or keep
`v0.1.0` since it works and is already in place? This is a style/robustness
decision that doesn't affect current behavior.

_~~Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision.~~ Resolved 2026-08-16: siblings keep real published versions (ROADMAP "Resolved questions")._

### 3. Is the `result` symlink in the project root intentional or stale?

The project root contains a symlink: `result -> /nix/store/s2g7l4i6...`.
This is a Nix build output symlink, but it points to a specific store path
with a hash that doesn't match the current build output. It may be stale
from a previous `nix build` run. I cannot determine if this is expected
behavior from the Nix build system (some setups leave `result` symlinks) or
if it should be cleaned up. If it should be cleaned up, it should also be
added to `.gitignore` to prevent future occurrences. If it's intentional
(e.g., used by a Nix script), it should be documented.

_~~Routed to TODO_LIST.md (2026-08-16) — `trash result` pending owner; `.gitignore` already covers it.~~ Resolved: symlink gone from the root; `.gitignore` covers future ones._
