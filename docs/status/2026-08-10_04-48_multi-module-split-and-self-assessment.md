# Status Report: Multi-Module Split & Self-Assessment

**Date:** 2026-08-10 04:48
**Session goal:** Answer three open questions from prior sessions; execute Q1 (module split)

---

## a) FULLY DONE

### Q1: `datastartest` split into its own Go module

Executed end-to-end with verification across all build modes.

| Step | Status | Verification |
| --- | --- | --- |
| `datastartest/go.mod` created | DONE | `go mod tidy` resolved deps: go-sse direct, go-error-family + go-branded-id indirect |
| `datastartest/go.sum` created | DONE | 6 lines, 3 module hashes |
| Mutual replace directives | DONE | Root go.mod: `datastartest => ./datastartest`; datastartest go.mod: `go-datastar => ..` |
| `go.work` created | DONE | `go 1.26.5` with `use ( . ./datastartest )` |
| Workspace test (both modules, -race) | PASS | 4 packages, 0 failures |
| GOWORK=off root test (-race) | PASS | e2e_test.go resolves datastartest via replace directive |
| GOWORK=off datastartest test (-race) | PASS | Resolves go-datastar via replace directive |
| Workspace vet | PASS | 0 issues |
| Workspace lint | PASS | 0 issues |
| Nix hermeticCheck build | PASS | `nix build .#checks.x86_64-linux.build` exit 0 |
| `flake.nix` updated | DONE | Removed `GOWORK=off` from devShell; all 8 apps now include `./datastartest/...` |
| `AGENTS.md` updated | DONE | Module Structure section, dual-mode Commands section, file layout table |

**What the split achieves:** Consumers can now `go get github.com/larsartmann/go-datastar/datastartest` as an independently versioned module, without pulling the protocol library's test infrastructure into their dependency tree.

### Q2 and Q3: Presented with code examples and PRO/CON analysis

Both questions answered with concrete examples — awaiting user decisions.

---

## b) PARTIALLY DONE

Nothing. Everything I started this session was completed.

---

## c) NOT STARTED (things I should have done but didn't)

These are gaps I identified during the self-assessment:

1. **`.gitignore` still ignores `go.work` and `go.work.sum`** — This is the single biggest oversight. `go.work` is essential for the dev workflow (workspace mode is the default now that `GOWORK=off` was removed from the devShell). A fresh clone will not have `go.work`, so `go test ./...` from root won't find `datastartest`. The replace directives keep `GOWORK=off` builds working, but the workspace dev experience is broken on fresh clone.

2. **`dependabot.yml` not updated** — Only monitors `/` (root go.mod). Needs a second entry for `directory: /datastartest` to get dependency update PRs for the new module.

3. **`flake.nix` hermeticCheck doesn't build or test `datastartest`** — `subPackages = [ "." ]` only builds the root module. The Nix CI passes but never compiles or tests `datastartest/`. This is a false-green: CI looks healthy but doesn't cover the new module.

4. **`CHANGELOG.md` not updated** — No entry for the module split.

5. **`FEATURES.md` not updated** — Still says `datastartest/` subpackage, not separate module.

6. **`README.md` not updated** — Installation instructions don't mention the separate module path for test helpers.

7. **`reader.go:114` stale comment** — Says "Shared by ReadEvents and readNEvents" but `readNEvents` was exported to `ReadNEvents` in the prior session. Pre-existing, not introduced this session, but I touched the file's context and should have caught it.

---

## d) TOTALLY FUCKED UP

### The `go.work` gitignore problem is a structural defect

This is not just "not started" — it's a design error I introduced. Here's the chain:

1. I removed `GOWORK = "off"` from the flake.nix devShell (correct for workspace mode)
2. I wrote AGENTS.md commands that assume workspace mode is the default (correct for the new design)
3. But `go.work` is in `.gitignore`, so it will never be committed
4. A fresh clone has no `go.work` — the default `go test ./...` won't find `datastartest`
5. The flake devShell no longer sets `GOWORK=off`, so it silently falls back to no-workspace mode
6. Result: **the documented default commands don't work on fresh clone**

The fix is either:
- **Option A:** Remove `go.work` and `go.work.sum` from `.gitignore` and commit them. This is the modern Go multi-module pattern (Kubernetes, Helm, cosign all do this). The replace directives + go.work together provide both local dev and CI isolation.
- **Option B:** Keep `go.work` gitignored and set `GOWORK=off` back in the devShell. Document that workspace mode requires `go work init`. This is more conservative but adds friction.

I should have caught this before declaring done.

### The duplicated AGENTS.md paragraph

My initial edit produced a duplicated "Note: do not pass --enforce-samber-oops" paragraph in AGENTS.md. I caught and fixed it, but it shouldn't have happened — the multiedit should have been structured to avoid the overlap. Sloppy editing.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Run `git status` BEFORE declaring done** — I declared Q1 done without checking whether the files were actually tracked. The auto-commit daemon committed them, but `go.work` was silently eaten by `.gitignore`. A `git ls-files go.work` check would have caught this.

2. **Test fresh-clone behavior, not just working-tree behavior** — All my verification was in the working tree where `go.work` already existed. I never simulated a fresh clone. The `nix build` test is closest to fresh-clone but it uses `GOWORK=off` implicitly via the Nix sandbox, so it passed.

3. **Read `.gitignore` before creating files it might affect** — I created `go.work` without checking whether it was ignored. A 5-second `grep go.work .gitignore` would have caught this.

4. **The auto-commit daemon makes "is it committed?" checks unreliable** — It committed my changes before I could verify tracking status. I need to check `git ls-files` explicitly, not `git status`.

### Architecture improvements (from the module split experience)

5. **The mutual-replace pattern is correct but adds maintenance burden** — Two go.mod files with replace directives pointing at each other. Any version bump in a shared dependency (go-sse, go-error-family) must be done in both modules. `go work sync` helps but must be run manually.

6. **`go.sum` duplication** — Both modules maintain independent go.sum files for the same underlying deps. This is correct (each module is independent) but means checksum mismatches can drift between modules.

---

## f) Up to 50 things to do next

### P0 — Critical (broken on fresh clone)

1. ~~**Remove `go.work` and `go.work.sum` from `.gitignore`** — Commit the workspace file so fresh clones work~~ done at `a73a8fb` (force-added past the global gitignore too)
2. ~~**Verify fresh-clone behavior** — `git clone` to a temp dir, `go test ./... ./datastartest/...` without any manual `go work init`~~ done — CI exercises the committed workspace
3. ~~**Add dependabot entry for `datastartest/`** — `directory: /datastartest` in `.github/dependabot.yml`~~ done in the v0.1.0 session (plus `/static`)

### P0 — CI gap (false green)

4. **Update `flake.nix` hermeticCheck to build and test `datastartest`** — Either add a second `buildGoModule` or restructure the fileset/subPackages ← still open (`flake.nix` TODO comment)
5. ~~**Add CI workflow for `datastartest/`** if using GitHub Actions (check `.github/workflows/`)~~ done — CI test job covers all modules

### P1 — Documentation sync

6. ~~**Update `CHANGELOG.md`** with module split entry under `[Unreleased]`~~ done (v0.1.0 section)
7. ~~**Update `FEATURES.md`** — Change "subpackage" to "separate module" in Consumer Test Helpers section~~ done in the v0.1.0 session
8. ~~**Update `README.md`** — Add installation instructions for `datastartest` as separate module~~ done in the v0.1.0 session
9. ~~**Fix `reader.go:114` stale comment** — `readNEvents` → `ReadNEvents`~~ done in the v0.1.0 session

### P1 — Awaiting user decisions (Q2 + Q3)

10. ~~**Q2 decision: Freeze or consolidate Collect variants** — I recommend freeze; user needs to confirm~~ resolved (2026-08-16) — per-helper variadic options (`WithPath`, `WithHeader`, `WithLastEventID`, `WithDatastarSignals`) landed in CHANGELOG `[Unreleased]`
11. ~~**Q3 decision: `RequireSignalsContain` substring vs JSON parsing** — I recommend fix-doc (Option A); user needs to confirm~~ resolved — doc fixed (v0.1.0)
12. ~~**If Q3 = Option A: Fix `RequireSignalsContain` doc comment** at `datastartest/assert.go:99-101`~~ done (v0.1.0)
13. ~~**If Q3 = Option C: Add `RequireSignalsHasKey`** with `map[string]any` top-level lookup~~ **Won't implement** — Option A chosen

### P1 — Code quality (from prior session, still open)

14. **Add `CollectPost` error-path tests** — handler returns 400/500, non-SSE body
15. **Add `CollectWithTimeout(timeout=0)` test** — immediate deadline edge case
16. **Replace `1<<30` magic number** in `datastartest/collect.go:150` with a cleaner `ReadAllEvents` or `math.MaxInt32`
17. **Rename `indexTagEnd` to `indexScriptTagEnd`** at `datastartest/event.go:190` — contract is narrower than the name implies
18. **Split `collect.go` (215 lines)** into `collect.go` + `streaming.go` if it grows further
19. **Split `event.go` (242 lines)** into `event.go` + `accessors.go` if it grows further

### P1 — Multi-module infrastructure

20. ~~**Run `go work sync`** and verify go.work is stable~~ done — idempotent; CI-enforced
21. ~~**Add a CI check that go.work and replace directives are in sync** — prevents drift~~ done at `dc0d6f2` (workspace-sync idempotency + replace audit)
22. ~~**Verify `erraudit` works with workspace mode** — `erraudit ./... ./datastartest/...`~~ done — CI erraudit scans all modules
23. ~~**Verify `govulncheck` works with workspace mode** — `govulncheck ./... ./datastartest/...`~~ done — CI govulncheck scans all modules
24. ~~**Update `.golangci.yml`** if it has path-specific config that doesn't cover `datastartest/`~~ done — lint runs on all modules, 0 issues

### P2 — Nice to have

25. **Add table-driven benchmark** with multiple input shapes (not just 20-event stream)
26. **Add `ReadAllEvents` function** — reads until EOF or context cancel, cleaner than `ReadNEvents(1<<30)`
27. **Add `indexTagEnd` support for unquoted HTML5 attributes** — `<script type=module>`
28. **Add `CollectWithOptions` if user decides to consolidate** (Q2)
29. **Add `RequireSignalsHasKey` if user decides Option C** (Q3)
30. **Document the mutual-replace pattern** in a short ADR (`docs/adr/`)
31. **Add `go work vendor` support** if offline builds are needed
32. ~~**Consider `datastartest` versioning strategy** — does it version-lock with the library or independently?~~ resolved — independent versioning (`datastartest/v0.1.0`, `v0.2.0`)
33. **Add integration test that imports `datastartest` as an external consumer would** — `GOWORK=off go get` in a temp module
34. **Review whether `datastartest` should depend on go-sse directly** — currently does (for `sse.Event` type in filter.go)
35. ~~**Add `doc.go` package-level example** showing the most common test pattern as the first thing in godoc~~ done (`datastartest/doc.go` quick start)
36. **Consider `datastartest` as a standalone repo** in the future if it grows beyond DataStar-specific helpers
37. ~~**Add versioned releases for `datastartest`** — tag `datastartest/v0.1.0` separately from root module~~ done (also `static/v0.1.0`)
38. **Audit `datastartest/go.sum` against root `go.sum`** for checksum consistency
39. **Add `nix flake check` to CI** if not already present
40. **Consider `go-releaser` config** for multi-module tagging
41. **Add CONTRIBUTING.md note** about the dual-module structure
42. **Review whether `example/` should be its own module** — currently in root, may pull in test deps
43. **Add `datastartest` to the PULL_REQUEST_TEMPLATE.md** checklist
44. **Consider semantic import versioning** if datastartest reaches v1
45. **Add a dependency graph visualization** (D2) to docs showing module relationships
46. **Lint the go.work file** — `go work edit -fmt`
47. **Add `make vendor` equivalent** via flake.nix app for offline development
48. **Review test coverage per module** — ensure datastartest has its own coverage report
49. **Add a `CHANGELOG.md` entry for datastartest separately** if it versions independently
50. **Consider extracting SSE parser** (`reader.go`, `event.go`) into a generic `ssetest` package reusable beyond DataStar

---

## g) Questions I CANNOT figure out myself

### Q1: ~~Should `go.work` be committed or stay gitignored?~~ Resolved — committed (`a73a8fb`; the v0.1.0 session also handled the global-gitignore trap).

### Q2: ~~Should the Nix hermeticCheck build and test `datastartest`, or is that a separate CI concern?~~ Resolved by decision — Nix stays root-only for now (TODO comment in `flake.nix`); GitHub Actions covers all three modules. Per-module Nix checks remain a TODO_LIST item.

### Q3: ~~When `datastartest` gets its first tagged release, should it version-lock with the library or version independently?~~ Resolved — independent versioning (`datastartest/v0.1.0`, `datastartest/v0.2.0` tagged alongside root releases).
