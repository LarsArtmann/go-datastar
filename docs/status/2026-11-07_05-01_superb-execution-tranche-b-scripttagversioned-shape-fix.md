# Status: "Make This Repo More Superb" — Tranche B Executed, Unreleased-API Shape Fixed, Pin Contract Landed

**Date:** 2026-11-07 05:01 CET
**Session scope:** Full-execution of the non-owner-gated, non-docs-lane tasks of
the 2026-11-07_04-12 pareto plan (M05/M07/M09/M13/M18/M24 + parts of M02
prep), in a shared checkout with an actively-converging parallel session.
**Starting state:** tree clean at `9f7637e`; race/vet/lint green;
docs-gates audit-all red (2 unharvested items, owned + fixed by the parallel
session mid-session).
**Ending state:** race suite ×8 packages green, vet green, pinned lint green,
erraudit ×4 clean, docs-gates PASS ×5; root coverage 98.9%.

---

## TL;DR

Executing the live plan's code+test lane surfaced a **real unreleased-API
defect**: `ScriptTagVersioned` returned a bare URL while its name, godoc,
CHANGELOG entry, README table, and `example/main.go` all treated it as a full
`<script>` tag — the example page silently loaded **no DataStar client at
all**. Because the API is unreleased, the shape was corrected now (last
chance before v0.7.0 locks it): it returns the tag, replaces an existing
`v` value (idempotent), and is pinned by table tests, a godoc example, and an
example smoke test. Landed alongside: the datastartest
`RequireScriptTagVersion` + `RequireETagForBundle` helpers (with a new
exported `datastar.ETagFor` so no algorithm duplicates), the CHANGELOG
transport-pin contract test (go-sse v0.7.0 / ssetest v0.5.0 / sibling pins
are now recorded — they were all invisible), the M24 micro-tranche (BOM
test, ETag-stable integration test, Dockerfile repair), the ROADMAP
go-directive drift fix, and M13 gate verdicts. The honest cost: I built a
duplicate of the parallel session's `broadcast/storecontract` package while
it was landing — the lane-claim failure the 04:33 report documented one hour
earlier, repeated by me despite reading it (§d1).

## a) FULLY DONE

1. **`ScriptTagVersioned` shape fix (unreleased-window).** Returned
   URL → full tag (`script_handler.go`), matching `ScriptTag`'s contract and
   every existing doc. `appendVersionQuery` now **replaces** an existing
   `v` parameter value (idempotent, `f(f(x)) == f(x)`, pinned by
   `TestAppendVersionQueryIdempotent`) and never matches `v` inside other
   params' keys/values (11 edge cases pinned). The stale-JS-window feature
   now actually guarantees the pinned version.
2. **Example page bug fixed + pinned.** `example/main.go` spliced the bare
   URL into `<head>` as if it were a tag — the live-feed demo loaded no
   client. The smoke test (`example/index_test.go`) asserts the exact
   `ScriptTagVersioned` output appears; godoc `ExampleScriptTagVersioned`
   added (its literal Output acted as the tripwire that caught the shape
   defect).
3. **datastartest consumer assertions.** `RequireScriptTagVersion(html,
   path)` and `RequireETagForBundle(etag, bundle)` (testing.TB invariant,
   failure-path tests, `*testing.B` acceptance), documented in
   datastartest README ("Versioned-cache assertions") and docs/testing.md.
4. **`datastar.ETagFor` exported** (replaces private `computeETag`; one
   name, one algorithm) so the datastartest ETag helper asserts handler
   truth instead of duplicating the digest.
5. **CHANGELOG transport-pin contract (M07).** `TestTransportPinsRecordedInChangelog`
   walks go.mod ×3 and requires every go-sse/ssetest/go-datastar/static pin
   to appear in CHANGELOG under the short-name convention; the `[Unreleased]
   ### Dependencies` section records the previously invisible pin set
   (go-sse v0.7.0, ssetest v0.5.0, error-family v0.12.0, static v0.6.1) and
   the post-0.6.3 go-sse bump (`ba90d86`).
6. **M24 micro-tranche.** `TestBytes_NoBOM` (static), the ETag-stable
   integration test (body==bundle, stable ETag, 304 round-trip), and the
   `example/Dockerfile` repair (`golang:1.26.7-alpine` → `1.27-alpine` —
   the 1.26.7 image cannot build `go 1.27`-directive modules — plus the
   dead `GOEXPERIMENT=jsonv2` the 2026-10-11 sweep missed).
7. **Docs truth.** ROADMAP "go directive policy" corrected to the
   minor-form reality (it still claimed patch-pinned `1.27.1`, contradicting
   AGENTS.md and go.mod); testing.md coverage date-stamped (root 98.9%,
   datastartest 95.6%, static 100%, broadcast 95.4% measured) and its
   measurement snippet dropped the `GOEXPERIMENT` flag; AGENTS.md gained the
   versioned-serving gotcha (M18b; 21.2KB, in band); migration-guide
   v0.7.0 skeleton updated to the final shapes (M02.4).
8. **M13 gate verdicts (verify-only).** Broadcast docspec CI wiring:
   confirmed (`ci.yml:101` covers `./broadcast/...`). API-surface golden
   test: none exists — future exports cannot trip one (non-issue
   confirmed). Renovate: installations API still 403s with a user token,
   zero PRs — the owner check stands. Upstream datastar-go `releases/latest`
   404s (no GitHub releases; tags only — the §5 checklist trio should probe
   tags, not releases).
9. **Session-end battery.** race ×8 green, vet green, pinned lint green
   (fresh cache), erraudit ×4 clean, docs-gates PASS ×5, tree committed by
   the daemon.

## b) PARTIALLY DONE

1. **The G7/M12 nix pass** — `nix flake check --keep-going` correctly caught
   that my root-source edit moved the vendor hashes (stale-vendor compile
   failure) and treefmt corrected golines drift; the buildflow hash-fix loop
   was still running at report time. The vendorHash movement is
   **expected** (root non-test `.go` edited, per the AGENTS gotcha) — this
   is the recorded confirmation of that rule, not a discrepancy.
2. **Tranche A docs** (README quick-start, doc.go mention, FEATURES row,
   static-js CSP note) — left to the parallel session, which had README.md
   dirty mid-flight at report time. My `ScriptTagVersioned` shape fix
   changes what those docs should say (full tag, replace semantics).

## c) NOT STARTED

1. The Store-contract helper's `SubscribeFilter` + injected-store
   replay-filter interaction extension (the shipped helpers cover ordering +
   ID assignment).
2. Archive-ledger residue passes, owner-blocked lane processing, the v0.7.0
   cut itself (owner-gated; the owner decision brief exists at
   `docs/planning/2026-11-07_v070-cut-owner-decision-brief.md`).

## d) TOTALLY FUCKED UP

1. **Rebuilt a package the parallel session was shipping.** I designed
   `broadcast/broadcasttest` while their `broadcast/storecontract` was
   landing mid-session; a test run exposed the collision. I discarded mine
   (theirs is better: `AssertIDAssignment` covers the ID-less path honestly)
   — but the 04:33 report documented this exact failure mode an hour
   earlier, and I pressed on anyway after reading it. The mod-time probe
   must come BEFORE the first write, not after the first `go test`.
2. **Replaced an existing regression test instead of appending** —
   `TestScriptHandler_NoSniffHeader` was briefly deleted by my own edit
   replacing its body; caught on re-read and restored. Appending to a file I
   had just restructured without re-reading the exact target region is the
   root cause.
3. **A broken comment line slipped into script_handler.go** (one `//`
   missing → invalid syntax), caught by the next compile. Multi-edits that
   rewrite whole comment blocks need an immediate `go build` between edits,
   not after the batch.
4. **Stayed on the corrupted shared build cache for two probe cycles**
   (missing fact files under the toolchain cache) instead of switching to a
   session-local `GOCACHE` on first sight.

## e) WHAT WE SHOULD IMPROVE

1. **Lane claim before ANY new file/package**: `git status` + newest-report
   check + one `ls` of the target directory immediately before creating
   anything. Cost of skipping tonight: one wasted package build (mine).
2. **The godoc example as tripwire worked** — a literal-Output example
   pinned to `static.Version` forces the version bump commit to update it
   in the same breath as the CHANGELOG. Keep this pattern for
   version-coupled surfaces.
3. **Unreleased APIs are free to fix — audit them at the cut.** The
   shape defect existed only because name/doc/CHANGELOG all said "tag"
   while tests pinned "URL": the tests had enshrined the divergence. When
   tests and docs disagree, the docs usually describe the intent.
4. **Formatter-regression opt-outs beat re-fixes** (04:33 report §e6,
   confirmed again): treefmt/golines disagree with plain gofmt on long
   Fatalf calls — always run `nix fmt`, never bare `gofmt -w`, on files the
   treefmt gate will re-check.

## f) Up to 25 things we should get done next (pre-flighted; each routes to a living row or is anchored by one)

1. Finish the Store-contract helper's SubscribeFilter + injected-store replay-filter interaction extension (the TODO row's full scope; the shipped helpers cover ordering + ID assignment only).
2. Verify the example Dockerfile rebuild after the golang:1.27 fix on a docker-capable host (no docker binary on this machine; repair is inspection-verified only).
3. At the v0.7.0 cut: execute release-checklist §2 require bumps, §2.5 nix gate, §3 four-tag lockstep, §4 proxy verification; decide the ScriptHandlerWith param removal and the cache policy.
4. Fold M06's README vs-upstream table into the tranche-A session's README work (upstream `releases/latest` 404s — probe tags, not releases; see §a8).
5. Tranche A docs should describe the final `ScriptTagVersioned` shape (full tag, replace semantics) — see §b2.

## g) Questions for the owner

1. None new — the standing rows (v0.7.0 cut, cache policy, Renovate app
   install, ScriptHandlerWith param) are consolidated in the
   2026-11-07_v070-cut owner decision brief.
