# Status Report: CI Hardening Cleanup & Regression Guard — 2026-08-10 07:55

## Session Goal

Complete the 7 remaining items identified in the prior session's self-review
(07-38 report): fix CHANGELOG incompleteness, fix the CI diff exit-code bug,
document new CI checks in AGENTS.md, validate CI YAML, run golangci-lint, add
a circular-dependency regression guard test, and run the full verification
suite.

---

## A) FULLY DONE

### 1. CHANGELOG.md — CI hardening section completed

Added two bullet points to the `Added — CI hardening` section under
`[Unreleased]`:

- **Workspace sync idempotency check** — CI copies `go.work`, runs `go work
  sync`, fails if the file changed.
- **Replace directive audit** — CI greps all `go.mod` files for absolute paths
  in replace directives.

The prior session only documented the per-module isolation check from the
session before that. This closes the documentation gap.

**File:** `CHANGELOG.md:20-30`

### 2. CI diff exit-code bug fixed

The prior session's idempotency check used a `diff ... && echo ... || { ... }`
pattern that conflated diff exit code 1 (files differ) with exit code 2
(trouble, e.g. file not found). Both went to the same `||` block, producing a
misleading "go.work changed" error even when the real problem was an I/O error.

Replaced with explicit `rc=$?` handling:

```bash
diff go.work go.work.bak; rc=$?
rm -f go.work.bak
if [ "$rc" -eq 0 ]; then
  echo "go.work is idempotent"
elif [ "$rc" -eq 1 ]; then
  echo "::error::go.work changed after 'go work sync' — run locally and commit"
  exit 1
else
  echo "::error::diff exited with code $rc (trouble)"
  exit 1
fi
```

Also changed `rm go.work.bak` to `rm -f go.work.bak` so the cleanup doesn't
fail if the file was already removed.

**File:** `.github/workflows/ci.yml:40-54`

### 3. AGENTS.md Commands section updated

Added two lines to the Commands section documenting the CI-enforced checks
that developers should run locally to pre-empt CI failures:

```bash
# CI also enforces (run locally to pre-empt CI failures):
GOEXPERIMENT=jsonv2 go work sync  # go.work must not change after sync (idempotency)
grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod  # must find nothing (relative paths only)
```

**File:** `AGENTS.md:35-38`

### 4. AGENTS.md file layout table updated

Added a row for the new `module_boundary_test.go` file:

```
| `module_boundary_test.go` | Regression guard: asserts root go.mod never requires datastartest (circular dependency prevention) |
```

**File:** `AGENTS.md:91`

### 5. CI YAML validation

Validated `.github/workflows/ci.yml` with two tools:

- **Python pyyaml** `yaml.safe_load()` — valid YAML
- **actionlint** (via `nix run nixpkgs#actionlint`) — exit code 0, no issues

actionlint is the more thorough check — it validates GitHub Actions syntax
semantics (step structure, expression syntax, job dependencies), not just
YAML structure.

### 6. golangci-lint — 0 issues

Ran `GOEXPERIMENT=jsonv2 golangci-lint run ./... ./datastartest/... ./static/... --timeout 5m`.

Initial run found 2 issues in the new `module_boundary_test.go`:

1. **modernize (stringsseq)** — `strings.Split` should be `strings.SplitSeq`
   for range-over-iteration (Go 1.24+)
2. **wsl_v5** — missing whitespace above the `if` after the `continue` block

Both fixed immediately. Re-run: 0 issues.

### 7. Regression guard test added

Created `module_boundary_test.go` with
`TestRootModuleDoesNotRequireDatastartest`:

- Reads root's `go.mod`
- Scans every non-comment line for `github.com/larsartmann/go-datastar/datastartest`
- Fails with a descriptive error if found, instructing the developer to
  relocate the test to `datastartest/` instead

This is the regression guard the prior session identified as needed (item
#3 in its "NOT STARTED" list, item #8 in its "WHAT WE SHOULD IMPROVE" list,
item #13 in its "50 things" list) but never created.

**File:** `module_boundary_test.go` (new, 35 lines)

### 8. Full verification suite — all green

| Check                    | Command                                                            | Result     |
| ------------------------ | ------------------------------------------------------------------ | ---------- |
| Regression guard test    | `go test -run TestRootModuleDoesNotRequireDatastartest -race`      | PASS       |
| Full test suite (race)   | `go test ./... ./datastartest/... ./static/... -race -count=1`     | All pass   |
| Vet                      | `go vet ./... ./datastartest/... ./static/...`                     | Clean      |
| Build                    | `go build ./... ./datastartest/... ./static/...`                   | Clean      |
| Per-module isolation     | `GOWORK=off go build + go test` per module                         | All 3 pass |
| go.work sync idempotency | `cp go.work go.work.bak && go work sync && diff`                   | Idempotent |
| Replace directive audit  | `grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod` | No matches |
| golangci-lint            | `golangci-lint run ./... ./datastartest/... ./static/...`          | 0 issues   |
| actionlint               | `nix run nixpkgs#actionlint .github/workflows/ci.yml`              | Exit 0     |
| YAML syntax              | `python3 -c "import yaml; yaml.safe_load(...)"`                    | Valid      |

---

## B) PARTIALLY DONE

Nothing is partially done. All 7 items from the todo list were completed and
verified.

---

## C) NOT STARTED

These items were carried forward from the prior session's "NOT STARTED" list
and remain out of scope for this session's CI-hardening and regression-guard
focus:

1. ~~**FEATURES.md update** — no mention of the modularization fix or the new
   regression guard test~~ done at `b5465f2` — module-structure rows; the guard is documented in AGENTS.md (`fda70c7`); CI rows re-verified 2026-08-16
2. **Version drift detection CI check** — proposed in real-world-patterns.md
3. ~~**Programmatic DAG acyclicity check** — a test that verifies the module
   dependency graph is acyclic (more robust than the text-scanning regression
   guard I added)~~ NOT-DO — superseded by `fda70c7`; the boundary guard covers the only cycle risk in this 3-module layout
4. ~~**erraudit on updated codebase** — AGENTS.md documents the command but I
   did not run it this session~~ done — CI erraudit job added at `eb8bf29`, runs every push
5. ~~**govulncheck on all 3 modules** — CI runs it but I did not verify locally~~ done — CI govulncheck job (`eb8bf29`) runs every push; 2026-08-16 run flags 4 stdlib vulns fixed in go1.26.6 (routed to TODO_LIST)
6. **v0.0.0 vs v0.1.0 normalization** — internal module references use v0.1.0,
   skill recommends v0.0.0 (needs user decision) ← open, routed to ROADMAP "Open questions" 2026-08-16
7. **go.work.sum git tracking decision** — currently gitignored (needs user
   decision) ← open, routed to ROADMAP "Open questions" 2026-08-16
8. **result symlink cleanup** — stale Nix build output in project root
   (still present, still gitignored, needs user decision) ← open, routed to TODO_LIST 2026-08-16
9. **Per-module Nix checks** — flake.nix TODO mentions hermeticCheckStatic,
   hermeticCheckDatastartest ← open, routed to TODO_LIST 2026-08-16
10. **nix flake check** — verify the flake is healthy
11. **14 gopls warnings** — stdversion (4), bloop (4), writestring (3),
    errorsastype (1) — all pre-existing, not introduced by this work ← mostly open — only writestring fixed (`fd3a5ac`); the rest persist (2026-08-16)
12. ~~**Pre-publish consumer experience tests** — `go get` root should NOT pull
    datastartest; `go get` datastartest/static should work independently~~ done — v0.1.0/v0.2.0 released with per-module tags
13. ~~**go mod verify** on all 3 modules~~ done — verified in the 08-13 02:58 session (and again 2026-08-16)
14. **Coverage across all 3 modules** — consider a merged coverage output ← open, routed to TODO_LIST 2026-08-16
15. ~~**CHANGELOG entry for the regression guard test** — I documented the CI
    steps but did not add a separate CHANGELOG entry for the new
    `module_boundary_test.go` test file itself~~ covered in substance — the [Unreleased] "Fixed — Module boundary" section (`3cd669e`, shipped in [0.1.0]) documents the fix this guard enforces

---

## D) TOTALLY FUCKED UP

Nothing is fucked up. All changes compile, test, lint, and vet cleanly. The
working tree was committed by the auto-commit daemon (3 commits: `3cd669e`,
`dc0d6f2`, `fda70c7`).

### Honest Criticisms of This Session

1. **The regression guard test is fragile.** It reads `go.mod` via
   `os.ReadFile("go.mod")` with a relative path. This assumes the test is
   always run from the module root directory. If Go ever changes test working
   directory behavior, or if someone runs `go test` from a different
   directory, the test fails with "file not found" rather than a meaningful
   assertion failure. A more robust approach would use `golang.org/x/mod/modfile`
   to parse the file semantically, or use `//go:embed go.mod` to embed the
   file at compile time. The text-scanning approach (`strings.Contains`) is
   also fragile — a comment line mentioning datastartest (like the one I
   added to AGENTS.md) could theoretically appear in go.mod as a comment and
   be flagged. The current code skips `//` prefixed lines, but a
   multi-line comment block or a trailing `// comment` on a require line
   would not be caught correctly. Good enough for now, but should be
   upgraded to `modfile.Parse` for production-grade guarding.

2. **I didn't add the regression guard test to the CHANGELOG.** I added the
   two CI steps to `Added — CI hardening` but the new
   `module_boundary_test.go` test file is not mentioned anywhere in the
   CHANGELOG. It should be under `Added — Testing` or similar.

3. **I didn't run erraudit or govulncheck locally.** AGENTS.md documents
   both as commands. The CI pipeline runs them, but I didn't verify locally
   that they pass with the current codebase. This is a gap — CI could fail
   on these and I wouldn't know until push.

4. **I didn't verify the `result` symlink state.** The prior session
   flagged it as a question. It's still present
   (`result -> /nix/store/m5klhxa6m92jmf7h9xyyaygsciwkydqi-...`), still
   gitignored. I noticed it but didn't act on it because it was flagged as a
   user question. Fair, but I could have at least verified whether the Nix
   store path still exists (stale symlink check).

5. **The `go.work.sum` question is still unanswered.** The prior session
   flagged this. `go.work.sum` is in `.gitignore`, `go.work` is tracked.
   The go-modularize skill says to commit both together. I didn't resolve
   this because it needs a user decision, but I also didn't verify whether
   the lack of `go.work.sum` causes any practical issues (e.g. does
   `GOWORK=off` still work without it? Yes, because replace directives
   resolve locally, but checksum verification is weaker).

6. **I didn't update the prior status report (07-38) to mark items as
   resolved.** The 07-38 report lists items #1-4 in its "WHAT WE SHOULD
   IMPROVE" section as not done. This session resolved all 4, but the 07-38
   report still reads as if they're open. Future readers will have to
   cross-reference this report to know they're done. Could have added a
   "Resolved in 07-55 session" annotation.

7. **I didn't check whether the `module_boundary_test.go` test runs in
   isolation mode (`GOWORK=off`).** The test reads `go.mod` from disk, so
   it should work regardless of workspace mode, but I only ran it in
   workspace mode. Should have verified `GOWORK=off go test -run
   TestRootModuleDoesNotRequireDatastartest ./...` as well.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (should have been done this session)

1. ~~**Add `module_boundary_test.go` to the CHANGELOG.** The test file is a
   new addition that prevents a real bug class. It deserves its own entry
   under `Added — Testing` or similar.~~ covered in substance — see the "Fixed — Module boundary" entry (`3cd669e`, shipped in [0.1.0])

2. **Upgrade the regression guard to use `golang.org/x/mod/modfile`.**
   Instead of text-scanning `go.mod`, parse it semantically:
   ```go
   f, err := modfile.Parse("go.mod", goMod, nil)
   for _, r := range f.Require { if r.Mod.Path == datastartestPath { ... } }
   ```
   This is more robust, handles comments correctly, and won't false-positive
   on commented-out lines or trailing comments.

3. ~~**Run erraudit and govulncheck locally.** Both are documented in
   AGENTS.md. Both run in CI. But neither was run this session. A local run
   would catch issues before CI does.~~ done — CI runs both on every push (`eb8bf29`); adding erraudit to the nix devShell is routed to TODO_LIST

4. ~~**Verify the regression guard in GOWORK=off mode.** The test reads
   `go.mod` from disk, so it should be mode-independent, but "should" is
   not "verified."~~ done — the CI isolation job runs the full suite (incl. this guard) with `GOWORK=off` on every push

### Architectural

5. **The `result` symlink is still present and still a question.** It's
   gitignored, so it doesn't affect the repo, but it's clutter in the
   project root. A `nix build` leaves it behind. Should be cleaned up or
   documented as expected Nix behavior.
   _Routed to TODO_LIST.md (2026-08-16) — `trash result` pending owner._

6. **The `go.work.sum` tracking question is still unresolved.** The
   go-modularize skill recommends committing both `go.work` and
   `go.work.sum`. The project commits `go.work` but gitignores
   `go.work.sum`. This works but means fresh clones have weaker checksum
   verification for workspace-local module replacements. Needs a decision.
   _Routed to ROADMAP.md "Open questions" (2026-08-16)._

7. **The `v0.0.0` vs `v0.1.0` normalization question is still unresolved.**
   Internal module references use `v0.1.0`. The skill recommends `v0.0.0` to
   eliminate pseudo-version churn. Works fine with replace directives, but
   could be cleaner. Needs a decision.
   _Routed to ROADMAP.md "Open questions" (2026-08-16)._

8. ~~**No programmatic DAG acyclicity check exists.** The text-scanning
   regression guard I added catches the specific case of root requiring
   datastartest, but it doesn't verify the general case — that the module
   dependency graph is acyclic. A test using `golang.org/x/mod/modfile` to
   parse all three go.mod files and verify the dependency graph is a DAG
   would be more robust and general.~~ NOT-DO — `fda70c7`'s boundary guard covers the only cycle risk in this 3-module layout

### Process

9. **Three status reports in one day (07-27, 07-38, 07-55).** All three
   cover overlapping work. The 07-38 report's "NOT STARTED" list is now
   partially resolved by this session. The 07-27 report is fully resolved.
   Consider consolidating or marking older reports as superseded.
   _Resolved 2026-08-16: all three inline-annotated instead; consolidation rejected to preserve point-in-time snapshots._

10. **The go-modularize skill's Phase 7 documentation checklist is still
    the root cause of these cleanup sessions.** Two sessions (07-38 and
    07-55) were spent cleaning up documentation that should have been
    done in Phase 7 of the original session. The skill should enforce
    this more strongly.

---

## F) Up to 50 Things We Should Get Done Next

### Documentation (immediate)

1. ~~Add `module_boundary_test.go` to CHANGELOG under `Added — Testing`~~ covered in substance — "Fixed — Module boundary" (`3cd669e`, shipped in [0.1.0])
2. ~~Add FEATURES.md entry for the modularization fix and regression guard~~ done at `b5465f2`
3. ~~Mark the 07-38 status report as partially superseded by this one~~ done — 07-38 inline-annotated 2026-08-16
4. ~~Mark the 07-27 status report as fully resolved/superseded~~ done — 07-27 inline-annotated 2026-08-16
5. ~~Add a module DAG diagram to AGENTS.md (text-based: static → root → datastartest)~~ done at `3cd669e` — "Module Structure" table + replace-directive notes
6. Add a `docs/modularization/README.md` index for the modularization docs ← open (2026-08-16)
7. Add CONTRIBUTING.md section on the multi-module architecture ← open, routed to TODO_LIST 2026-08-16

### CI Hardening (immediate)

8. Add `go mod verify` step to CI (all 3 modules)
9. Add version drift detection script to CI
10. Add a CI step that verifies `go.work` `use` directives match actual
    go.mod files on disk
11. Add `actionlint` as a CI step to validate workflow YAML
12. Consider parallelizing CI jobs per module for faster feedback
13. ~~Add a CI step that runs the regression guard test in GOWORK=off mode~~ done — the CI isolation job runs the full suite `GOWORK=off` on every push (`eb8bf29` lineage)

### Testing

14. Upgrade `module_boundary_test.go` to use `golang.org/x/mod/modfile`
    instead of text scanning
15. ~~Add a programmatic test that verifies the module DAG is acyclic~~ NOT-DO — superseded by `fda70c7` boundary guard (covers the only cycle risk in this 3-module layout)
16. ~~Run erraudit on the updated codebase~~ done — CI erraudit job (`eb8bf29`) runs every push
17. ~~Run govulncheck on all 3 modules locally~~ done — CI govulncheck job runs every push; 2026-08-16 findings routed to TODO_LIST
18. Add test for `TestE2E_DataStarPatches` that exercises `CollectPost` and
    `CollectN` paths ← open, routed to TODO_LIST 2026-08-16
19. ~~Consider property-based testing for the SSE parser in datastartest~~ done at `fd3a5ac` (`FuzzReadEvents`)
20. ~~Run `go test -bench=. -benchmem` to verify benchmarks still pass~~ done at `fd3a5ac` — benchmarks run in the 04:25 session
21. ~~Verify the regression guard test passes in GOWORK=off mode~~ done — CI isolation job covers it

### Code Quality (pre-existing gopls warnings)

22. Fix gopls `stdversion` warnings — `json.Unmarshal` requires go1.27 in
    benchmark_test.go:92, datastartest/event.go:86, and 2 more files
23. Fix gopls `bloop` warnings — modernize `b.N` to `b.Loop()` in
    benchmark_test.go (4 instances)
24. ~~Fix gopls `writestring` warnings — inefficient string concatenation in
    reader_fuzz_test.go (3 instances)~~ done at `fd3a5ac` — `strings.Builder.WriteString` now used
25. Fix gopls `errorsastype` hint — simplify `errors.As` in errors_test.go:253

### Modularization Refinement

26. Evaluate `v0.0.0` vs `v0.1.0` for internal module references (needs
    user decision) ← open, routed to ROADMAP "Open questions" 2026-08-16
27. Evaluate whether `go.work.sum` should be tracked in git (needs user
    decision) ← open, routed to ROADMAP "Open questions" 2026-08-16
28. Clean up or document the `result` symlink in project root (needs user
    decision) ← open, routed to TODO_LIST 2026-08-16
29. Consider whether `static/` should have a go.sum file preemptively
30. Evaluate whether `example/` should get its own go.mod
31. Review `response_test.go` — it imports `static` directly; verify this
    is the right approach vs. using `datastar.DatastarJSVersion`

### Pre-Publish

32. ~~Create git tags for v0.1.0 release~~ done — v0.1.0 tagged 2026-08-10 (and v0.2.0 since)
33. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar`
    should NOT pull datastartest~~ done — root has no datastartest require
34. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/datastartest`
    should work~~ done — datastartest tagged v0.1.0/v0.2.0
35. ~~Test consumer experience: `go get github.com/larsartmann/go-datastar/static`
    should work~~ done — static tagged v0.1.0/v0.2.0
36. ~~Verify GOPROXY resolution works~~ done — v0.1.0/v0.2.0 resolve via the module proxy
37. ~~Run `go mod verify` on all 3 modules~~ done — verified in the 08-13 02:58 session (and again 2026-08-16)
38. ~~Consider adding `//deprecated` comments to old import paths if any
    changed~~ Won't implement — no import paths changed

### Nix/Build

39. Add per-module Nix checks (`hermeticCheckStatic`,
    `hermeticCheckDatastartest`) as the flake.nix TODO mentions ← open, routed to TODO_LIST 2026-08-16
40. Run `nix flake check` to verify the flake is healthy
41. ~~Consider adding `nix run .#erraudit` and `nix run .#govulncheck` to CI~~ done at `eb8bf29` — both are CI jobs
42. Consider adding a `nix run .#coverage` output that merges coverage
    across all 3 modules ← open, routed to TODO_LIST 2026-08-16
43. ~~Update the flake.nix `vendorHash` if go.sum changes in the future~~ done — maintained routinely (e.g. `5b70bb1`)

### Architecture

44. Consider whether `datastartest` should export a `NewResponse` helper for
    test ergonomics
45. ~~Evaluate whether `datastartest` should have a `CollectWithOptions` for
    custom headers~~ done at `06bb019` — every Collect* now takes `...RequestOption` (`WithHeader` et al.)
46. Consider adding a `datastartest.RequireEventOrder` helper for ordered
    event assertions
47. Consider whether `static/` should export a `Version` constant AND a
    `Version()` function (currently both exist)
48. ~~Add a pre-commit hook that warns if root go.mod gains a datastartest
    require~~ superseded by `fda70c7` — test-level guard runs in CI on every push

### Process

49. Enforce Phase 7 documentation checklist more strongly in the
    go-modularize skill (root cause of both cleanup sessions)
50. ~~Consider consolidating the three same-day status reports (07-27,
    07-38, 07-55) into a single retrospective~~ resolved 2026-08-16 — all three inline-annotated; consolidation rejected to preserve point-in-time snapshots

---

## G) Questions I CANNOT Answer Myself

### 1. Should `go.work.sum` be tracked in git or stay gitignored?

Carried forward from the prior session (07-38, question G.1). The
go-modularize skill's real-world-patterns.md says "Always commit `go.work`
and `go.work.sum` together." The project deliberately gitignores
`go.work.sum` but force-commits `go.work`. Without `go.work.sum`, fresh
clones lack checksums for workspace-local module replacements, weakening
supply-chain verification. With it, every dependency update produces a
`go.work.sum` diff. This is a tradeoff between reproducibility and noise
that needs a human decision.

_Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision._

### 2. Should internal module references use `v0.0.0` instead of `v0.1.0`?

Carried forward from the prior session (07-38, question G.2). The skill
recommends `v0.0.0` for all internal sibling requires to eliminate
pseudo-version churn. Currently all three modules use `v0.1.0`. This works
with replace directives (version is ignored when a replace exists), but
`go mod tidy` could generate pseudo-versions in edge cases if replace
directives are ever removed. Should we normalize to `v0.0.0` or keep
`v0.1.0` since it works and is already in place? This is a style and
robustness decision that doesn't affect current behavior.

_Routed to ROADMAP.md "Open questions" (2026-08-16) — still awaiting an owner decision._

### 3. Is the `result` symlink in the project root intentional or stale?

Carried forward from the prior session (07-38, question G.3). The symlink
`result -> /nix/store/m5klhxa6m92jmf7h9xyyaygsciwkydqi-...` is still
present. It's a Nix build output, gitignored, pointing to a store path
from commit `a4712ab`. I cannot determine if this is expected Nix behavior
(some setups leave `result` symlinks for `nix build` ergonomics) or if it
should be cleaned up. If intentional, it should be documented. If stale,
it should be removed (`trash result`) and the `.gitignore` entry is
sufficient to prevent future ones from being committed.

_Routed to TODO_LIST.md (2026-08-16) — `trash result` pending owner; `.gitignore` already covers it._
