# Status Report: 2026-08-16 07:52 — README Comparison Against the Official DataStar Go SDK

**Session goal:** The user felt the README needed an explicit comparison to the
official Go DataStar package — what makes go-datastar different, what makes it
better. Replace the terse three-line "Why not `starfederation/datastar-go`?"
paragraph with a full, honest comparison section.

**Scope:** Documentation only. One file changed (`README.md`). No code, no
dependencies, no go.mod changes.

**Commits this session:**

- This report + `README.md` (user explicitly requested the commit)

---

## a) FULLY DONE

### 1. Loaded the verification gate before writing claims

This task is exactly the inbound case for the `verify-external-claims` skill:
claims about an external library's API were about to be encoded into project
documentation. The skill was loaded and its checklist applied — every claim in
the comparison section is backed by a primary source (pkg.go.dev, the upstream
README), or by go-datastar's own code for self-claims.

### 2. Discovered the official SDK moved modules

Key finding that invalidated the old README text:

- Old assumption: the official SDK is `github.com/starfederation/datastar/sdk/go`.
  That path is **stale** — frozen at v0.21.4, published **Dec 23, 2024**,
  flagged "not the latest version of its module" by pkg.go.dev.
- Current reality (verified via pkg.go.dev search + package page): the official
  SDK is now **`github.com/starfederation/datastar-go`**, latest **v1.2.2**
  (published Jun 2, 2026), MIT, imported by 64 packages.

The rewritten section links to the current repo and pins the comparison to
v1.2.2 explicitly ("_Compared against datastar-go v1.2.2, the current
release._") so future readers know the reference point.

### 3. Verified the upstream API surface (datastar-go v1.2.2)

Confirmed from the pkg.go.dev package index:

- `ServerSentEventGenerator` remains connection-bound: `NewSSE(w, r, opts...)`
  requires the `http.ResponseWriter`; every patch (`PatchElements`,
  `MarshalAndPatchSignals`, `ExecuteScript`, ...) is a method on it. The core
  architectural claim stands.
- **Compression is built in and substantial**: `WithBrotli`, `WithGzip`,
  `WithZstd`, `WithDeflate` options plus `ClientPriority` / `ServerPriority` /
  `Forced` strategies via `WithCompression`. Listed honestly as an upstream win.
- **`ReplaceURLQuerystring(r, values, opts)`** exists upstream; go-datastar
  only has `ReplaceURL`. Listed honestly as an upstream win.
- Templ + GoStar rendering (`PatchElementTempl`, `PatchElementGostar`),
  printf-style variants (`Redirectf`, `PatchElementf`, `RemoveElementf`,
  `ConsoleLogf`), `GetSSE`/`PostSSE`/... helpers, `ReadSignals` — parity rows.
- **No** broadcaster, event store, replay, subscribe-filter, test helpers, or
  JS-bundle serving anywhere in the package index. The differentiator rows are
  accurate negatives, not assumptions.
- Upstream README states **Go 1.24+** with no experiments. Stated in the
  constraints row (go-datastar requires Go 1.26+ + `GOEXPERIMENT=jsonv2`).

### 4. Verified go-datastar's own claims against its code

Every self-claim in the table maps to real, existing surface: `Patch`
interface, go-sse `Broadcaster`/`EventStore`/`SubscribeFilter`, `MemoryStore`,
`LastEventID`, `ScriptHandler()` (ETag + Cache-Control), the `datastartest`
module, and the errorfamily code/family/retryability catalog from `errors.go`.

### 5. Wrote the section

`README.md` now contains (replacing the old "Why not" paragraph):

- **Framing paragraph** — official SDK is a fine default; go-datastar exists
  because of one architectural decision.
- **"The core difference: patches are values"** — side-by-side code: upstream
  (construct-and-write in one call, requires connection) vs go-datastar
  (construct anywhere → `Event()` → broadcast/store). Plus the punchline: same
  wire format, the difference is what you can do before the network.
- **11-row feature table** — patch model, offline construction, broadcast,
  replay, filtering, error handling, test helpers, JS client serving,
  compression, Templ/GoStar, printf variants.
- **"Where the official SDK wins"** — four honest items: built-in compression,
  `ReplaceURLQuerystring`, lower environment constraints, first-party cadence.
- **"When to choose which"** — official SDK for single-connection handlers;
  go-datastar when patches are application state (broadcast, replay,
  filtering, classified errors, E2E-tested handlers).

### 6. Checked for dangling references

Grepped for "Why not" across the repo — the only remaining hits are unrelated
(an ADR subsection title, two old status-report table headers). No doc links
into the removed heading.

---

## b) PARTIALLY DONE

### 1. Upstream verification depth: docs-level, not source-level

Claims were verified against the pkg.go.dev v1.2.2 index and the upstream
README — not by reading their full source or writing a compile probe. For a
feature-comparison table this is sufficient (the API index is generated from
source), but e.g. the "no JS bundle serving" negative is an absence claim
backed by the package index + README, not an exhaustive source grep of the
repo's other packages (`cmd/generate`, `cmd/testserver` were not inspected).

### 2. No CI/doc gate run

Doc-only change; build/test/lint gates not run this session (no code touched).
Markdown itself was not linted/formatted — the repo has dprint.json (see prior
report: possibly an orphan), and it was not invoked.

---

## c) NOT STARTED

1. **Comparison freshness maintenance** — the section pins v1.2.2 but nothing
   re-checks it when datastar-go releases. A CI or release-checklist step
   ("re-verify README comparison against latest datastar-go") does not exist.
2. **Version-badge style footnotes for JS client version** — the table says
   "embedded zero-dep `static` module" without noting the bundled client
   version (v1.0.2 per README's ScriptHandler row); the section is silent on
   whether upstream bundles the same client version.
3. **datastar-go ecosystem scan** — pkg.go.dev search surfaced other third-party
   wrappers (dsx, gomponents-datastar, fluent-datastar, datastargostrictcsp).
   No positioning against those; the comparison covers only the official SDK.

---

## d) TOTALLY FUCKED UP

Nothing this session. No code touched, no commits made without instruction
(last session's lesson held — this commit is explicitly user-requested), no
claims written without a primary source. One near-miss worth recording: the
first fetch of the old module path (`.../sdk/go`) 404'd on the badge URL and
could have been misread as "the SDK is dead" — the pkg.go.dev search page is
what revealed the module had simply moved. Chasing the 404 instead of
searching would have produced a fabricated claim.

---

## e) WHAT WE SHOULD IMPROVE

1. **Pin comparison sections to verified versions in-repo.** The
   "_Compared against v1.2.2_" footnote pattern should be the standard for any
   README that contrasts a moving-target dependency — otherwise the table
   silently rots. Consider a release-checklist item to re-verify.
2. **Old external references in docs age badly.** The pre-existing section
   linked a module path that has been dead since the upstream repo split
   (Dec 2024 → mid-2026 with no one noticing). When a doc references an
   external package, prefer the canonical repo + a version note over a deep
   path.
3. **Absence claims deserve a note on how they were checked.** "Not built in"
   rows are the most falsifiable claims in the table. They were checked
   against the generated package index — strong but not exhaustive; the
   section could say so in one footnote if we want to be maximally honest.

---

## f) Next Steps (bounded — doc session)

1. ~~Re-verify the comparison table when datastar-go cuts its next release
   (upstream cadence appears roughly monthly); update the footnote version
   ← open, routed to TODO_LIST 2026-08-16~~ done at `83d7c60` — re-verified
   2026-08-16 (T12): v1.2.2 still latest; standing re-check lives in
   `docs/release-checklist.md`
2. ~~Decide whether to mention the JS client version parity (v1.0.2 embedded
   vs. whatever upstream ships) in the table's "Serve the DataStar JS client"
   row ← open, routed to TODO_LIST 2026-08-16~~ done at `83d7c60` — the row
   now cites "embedded zero-dep `static` module (JS client v1.0.2)"
3. ~~Optionally add the README comparison re-check to the release checklist
   (docs/ or CHANGELOG process) so it cannot rot silently
   ← open, routed to TODO_LIST 2026-08-16~~ done at `83d7c60` — quarterly
   re-verify step codified in `docs/release-checklist.md`
4. ~~The unrelated working-tree changes under `datastartest/` (modified
   `assert.go`, `collect.go`, `reader.go`, `e2e_test.go`, plus untracked
   `options*.go` / `assert_test.go`) belong to another session and were left
   untouched on purpose — whoever owns that work should report on it.~~ done —
   committed as the request-options session: `06bb019` (options, script/ID
   assertions, TB support) and `b0482db` (dataless-frame fix)
