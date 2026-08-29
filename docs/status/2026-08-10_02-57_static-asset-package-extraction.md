# Status Report — 2026-08-10 02:57

## Session Goal

Extract the `//go:embed datastar.js` from the root `datastar` package into a dedicated module/package.

---

## What Was Done

The `//go:embed static/datastar.js` directive, the `embeddedDatastarJS []byte` var, and the `DatastarJSVersion` constant were moved out of `script_handler.go` into a new **`static` subpackage** (`static/static.go`). The root package now imports `static` and re-exports its symbols for backward compatibility.

### Files Created

| File                    | Role                                                                                                  |
| ----------------------- | ----------------------------------------------------------------------------------------------------- |
| `static/static.go`      | Dedicated asset package: `//go:embed datastar.js`, `Bytes() []byte`, `Version` const ("1.0.2")        |
| `static/static_test.go` | 4 tests: version sanity, non-empty bytes, header-banner-matches-version guard, shared-slice stability |

### Files Modified

| File                | Change                                                                                                                                                                                                                                                                |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `script_handler.go` | Dropped `_ "embed"` import + `embeddedDatastarJS` var + inline version const. Added `static` import. `ScriptHandler()` now calls `ScriptHandlerWith(static.Bytes(), static.Version)`. `DatastarJSVersion` re-exports `static.Version` as a backward-compatible alias. |
| `AGENTS.md`         | File-layout table: `script_handler.go` row updated; new `static/` row added.                                                                                                                                                                                          |

### Verification (all passed)

| Gate                           | Result                                                    |
| ------------------------------ | --------------------------------------------------------- |
| `go build ./...`               | clean                                                     |
| `go vet ./...`                 | clean                                                     |
| `go test ./... -race -count=1` | all pass (root, datastartest, **static**), example builds |
| `golangci-lint run ./...`      | 0 issues                                                  |
| `go vet ./example/`            | clean                                                     |

---

## a) FULLY DONE

1. **`static` package created** — owns the embed, exposes `Bytes()` + `Version`, documented with package-level godoc.
2. **`script_handler.go` refactored** — consumes `static` package; public API (`ScriptHandler`, `ScriptHandlerWith`, `ScriptTag`, `Version`, `DatastarJSVersion`) unchanged.
3. **Tests written** — 4 tests covering version format, non-empty bytes, header/banner consistency, and shared-slice identity.
4. **AGENTS.md updated** — file-layout table reflects the new architecture.
5. **All quality gates pass** — build, vet, test (race), lint, example vet.

## b) PARTIALLY DONE

Nothing is half-finished. The refactor itself is complete and verified.

## c) NOT STARTED

| # | Item                                  | Why it matters                                                                                                                                                                                   |
| - | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **CHANGELOG.md `[Unreleased]` entry** | Structural refactor (new package); the `[Unreleased]` section exists but was not updated. Clear miss. ~~→ done at `7c18089`~~                                                                    |
| 2 | **FEATURES.md line 66**               | Says "v1.0.2 embedded. ScriptHandler() ... (`script_handler.go`)." The embed now lives in `static/`. File reference is stale. ~~→ done at `222353e` (later updated again for the module split)~~ |
| 3 | **README.md API surface table**       | Does not mention `static.Bytes()` / `static.Version` for consumers who want raw access to the JS bundle without HTTP.                                                                            |
| 4 | **Root `doc.go`**                     | Doesn't mention that the JS client is served from the `static` subpackage. Minor.                                                                                                                |
| 5 | **`.github/workflows/ci.yml` review** | Not checked for embed-path references that may now be stale. ~~→ moot — CI covers all three modules since v0.1.0~~                                                                               |

## d) TOTALLY FUCKED UP

Nothing. The refactor is clean, correct, and fully tested. No breaking changes. No data loss. No broken builds.

The closest thing to a fuckup is the **interpretation ambiguity** (see Questions below): the user said "dedicated module" and I built a "dedicated package." If the user meant a separate Go module (`go.mod`), this is wrong and needs rework. But given the engineering context (a JS asset bundle does not warrant independent versioning), a package was the right call.

## e) WHAT WE SHOULD IMPROVE

### Design Decisions Worth Questioning

1. **`Bytes()` returns the shared embedded slice, not a copy.** Documented as "must not be modified," but a careless caller can corrupt the asset for all subsequent `ScriptHandler` requests. Go's own `embed.FS.ReadFile` returns a copy. Tradeoff: zero-allocation vs. safety. If this is a public API that consumers call directly, a defensive copy may be wiser.

2. **`DatastarJSVersion` redundancy.** Both `datastar.DatastarJSVersion` and `static.Version` exist. The const alias (`const DatastarJSVersion = static.Version`) is backward-compatible, but it is two names for the same value. Consider deprecating the root alias with a doc comment once consumers migrate.

3. **`ScriptHandlerWith(scriptBytes []byte, _ string)`.** The second parameter (`version`) is unused — the `_` discard was pre-existing, but the signature now looks odd when the version is available from `static.Version`. The parameter exists for custom bundles where the caller supplies their own version string, but it's never read. Consider using it (e.g., in a header) or removing it in a future major version.

### Process Gaps

4. **CHANGELOG discipline.** I completed a structural change and did not update the CHANGELOG. This violates the project's own documentation-file table: `CHANGELOG.md` is for change history, and this is a change.

5. **FEATURES.md drift.** I updated AGENTS.md but not FEATURES.md. Both reference the old embed location. Half-updating docs creates split brains.

---

## f) Up to 50 Things to Get Done Next

### Documentation (high priority — stale right now)

1. ~~Add `[Unreleased]` CHANGELOG entry for the `static` package extraction~~ done at `7c18089`
2. ~~Update FEATURES.md line 66: replace `script_handler.go` with `static/` for the embed location~~ done at `222353e`
3. ~~Update FEATURES.md line 67: same file-reference check for HEAD support~~ done at `222353e`
4. Add `static.Bytes()` and `static.Version` to the README API surface table
5. Mention the `static` subpackage in root `doc.go` package documentation
6. Add a CONTRIBUTING.md note about how to update the embedded `datastar.js` bundle
7. Consider adding an ADR entry (`docs/adr/002-static-asset-package.md`) documenting the extraction decision

### Correctness & Safety

8. Decide whether `Bytes()` should return a defensive copy (safety) or shared slice (zero-alloc)
9. ~~Add a consistency test asserting `datastar.DatastarJSVersion == static.Version`~~ done (`TestStaticVersionConsistency`, `response_test.go`)
10. ~~Add a test that `ScriptHandler()` serves the exact bytes from `static.Bytes()` (not a stale copy)~~ done (`TestScriptHandler_ServesStaticBytes`, `response_test.go`)
11. ~~Check `.github/workflows/ci.yml` for embed-path or file-location references that are now stale~~ moot — CI covers all three modules since v0.1.0
12. Fix the pre-existing `erraudit` `silent_swallow` in `example/main.go:87` (noticed during this session, not caused by it)
13. ~~Add a CI guard that verifies `static.Version` matches the version banner in `datastar.js`~~ done — `static_test.go` banner-consistency guard runs in CI
14. Review whether `ScriptHandlerWith`'s unused `_ string` parameter should be removed or used

### Upstream Asset Tracking

15. ~~Check whether upstream DataStar has released a version newer than 1.0.2~~ done — confirmed latest in the v0.0.3 session (T13, 2026-08-08; re-check periodically)
16. If newer exists, update `static/datastar.js` and `static.Version`
17. Add a `go:generate` or flake target to download/verify the upstream bundle
18. Consider pinning the upstream commit SHA in a comment for reproducibility
19. Add a checksum verification step for the downloaded bundle

### API Surface Polish

20. Consider whether `Bytes()` should be renamed to `JavaScript()` or `Bundle()` for clarity
21. Consider whether `Version` should be a function (`Version() string`) to match the root package's `Version()` style
22. Consider deprecating `DatastarJSVersion` root alias with a `// Deprecated:` comment
23. Add godoc `// Example` functions for the `static` package
24. Consider whether the `static` package should export `ETag()` for consumers building custom handlers
25. Evaluate whether `ScriptHandler` should add a `Last-Modified` header for better CDN caching

### Performance

26. Add a benchmark for `computeETag` with the full 56 KB `datastar.js` payload
27. Evaluate pre-computing the ETag at init time (currently computed per `ScriptHandlerWith` call)
28. Consider gzip/Content-Encoding support for the served bundle
29. Consider whether `ScriptHandler` should support HTTP Range requests

### Testing Hardening

30. Add a fuzz test for `ScriptHandlerWith` with random byte inputs (ETag stability)
31. Add a test for `ScriptHandler` with empty/zero-length custom bytes (edge case)
32. Dogfood the `static` package in `e2e_test.go` (currently only tests `ScriptHandler` HTTP behavior)
33. Add a race-detector-specific test that calls `Bytes()` concurrently from many goroutines
34. Add a test verifying the `If-None-Match` conditional request with the static-package ETag

### Architecture

35. Consider whether the `static` package should eventually own JS bundling/minification
36. Evaluate whether `static` should be split further (e.g., `static/asset` + `static/version`)
37. Consider whether the `static` package needs any build tags (e.g., `//go:build !wasm`)
38. Evaluate whether `static` should support multiple assets (maps, icons) or stay single-file
39. Consider whether the root package should re-export `Bytes()` directly (not just via `ScriptHandler`)

### DevOps & CI

40. Add a flake.nix target for updating/regenerating the embedded bundle
41. Add a CI step that fails if `static/datastar.js` and `static.Version` are out of sync
42. Review whether the auto-commit message (`44147a2`) accurately describes the change for future readers
43. Add the `static` package to any coverage tracking/thresholds in CI
44. Consider whether `erraudit` should be run on `./static/...` specifically (currently covered by `./...`) — moot: CI erraudit scans all three modules

### Misc

45. Consider whether the package name `static` could shadow anything in consumer code
46. Add a `static/doc.go` if the package grows beyond one file
47. Evaluate whether the `Cache-Control: max-age=86400` (24h) is appropriate for a versioned, ETagged asset
48. Consider whether `ScriptHandler` should set `X-Content-Type-Options: nosniff`
49. Add a test for `ScriptTag()` with edge-case paths (empty, with query params, with fragments)
50. Evaluate whether the `static` package should have its own version separate from the module version

---

## g) Questions (cannot figure out myself)

### Q1: ~~"Module" or "Package"?~~ Resolved — became a separate Go **module** at v0.1.0 (`github.com/larsartmann/go-datastar/static`, tagged `static/v0.1.0`).

You said "dedicated module." In Go, **module** = a `go.mod` unit with independent versioning. **Package** = a directory of `.go` files within a module. I interpreted your request as **package** and created `static/` within the existing module. If you actually wanted a **separate Go module** (own `go.mod`, independent versioning, consumers importing two modules), that is a fundamentally different architecture and I need to rework this.

### Q2: Should `Bytes()` return a defensive copy?

`Bytes()` currently returns the shared embedded slice (zero allocation). A careless or malicious caller could `modify` the returned slice and corrupt the asset for all subsequent `ScriptHandler` requests. Go's `embed.FS.ReadFile` returns a fresh copy. Should I switch to a defensive copy (safe, allocates 56 KB per call) or keep the shared slice (fast, documented "do not modify")?

### Q3: Should the `static` package own the JS lifecycle (download, minify, version-pin)?

Right now `static/` ships a hand-pinned `datastar.js` blob. Should it eventually own the full lifecycle — `go:generate` to download from the upstream `starfederation/datastar` releases, minify, checksum-verify, and write `Version` from the downloaded release tag? Or should it always remain a hand-maintained static file?

---

## Session Summary

The refactor is **complete and verified**. The `//go:embed` is now a dedicated `static` package with its own tests. Public API is unchanged. All quality gates pass. The main gaps are **documentation drift** (CHANGELOG, FEATURES.md) and the **module-vs-package interpretation ambiguity** that needs user confirmation.
