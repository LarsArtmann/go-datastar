# Status: v0.7.0 full-execution run — cut prep, contract helpers, docs tranches, archive dispositions, v0.6.1 retro

**Date:** 2026-11-07 ~05:30 CET · **Lane:** pareto-plan M01–M25 full execution (parallel sessions active throughout; their landings incorporated, never reverted)
**Plan:** [2026-11-07_04-12 pareto plan](../planning/2026-11-07_04-12_pareto-v070-consumer-trust-scripttagversioned-surface-plan.md) · **Cut brief:** [2026-11-07_v070-cut-owner-decision-brief](../planning/2026-11-07_v070-cut-owner-decision-brief.md)

## a) TL;DR

Executed the Pareto plan end-to-end: v0.6.3 release-shape defect discovered,
evidenced, and truth-repaired; v0.7.0 cut prep complete (owner brief +
release-note draft + §2.5 rehearsal); the broadcast `Store` contract helper
shipped (`broadcast/storecontract`); the JS bundle bumped v1.0.3 → v1.0.4
through the canonical flow after the upstream check found it stale; the
docs tranches (M08/M10/M11) landed; three archive ledgers dispositioned
(10 + all-pointers + 23 verdicts); v0.6.1 retro written (section e); all
four modules green (race ×4, vet, lint v2.14.0 fresh-cache, nix flake check
all-passed, docspec, fuzz smokes).

## b) Done

1. **v0.6.3 incident (the session's biggest truth find).** v0.6.3 was cut
   2026-11-04 as a **root-only** release: no submodule lockstep tags exist
   (submodule `@v0.6.3` is an unknown revision on the proxy — live-probed),
   checklist §2's require bump was skipped (tagged root still requires
   `static v0.6.1`), and the CHANGELOG promotion was skipped (no `[0.6.3]`
   section existed on master). Root-only consumers are unbroken; lockstep
   consumers cannot resolve the submodules. CHANGELOG now carries a
   backfilled `[0.6.3]` section with the defect recorded inline; TODO_LIST
   and the cut brief carry the evidence.
2. **M02 cut prep.** CHANGELOG `[Unreleased]` coherence repaired;
   migration-guide v0.7.0 skeleton added (truth-discipline placeholders);
   release-note draft written (cut brief Appendix A, LICENSE provenance
   included); `go mod verify` ×4 clean; §2.5 rehearsal on a detached
   worktree — the nix gate correctly went red on the simulated post-bump
   tree (broadcast + datastartest vendor FODs failed; root's unmoved,
   matching the AGENTS movement rules) and the worktree was discarded, no
   tag.
3. **M03 owner brief.** Four decisions (ScriptHandlerWith param, release
   lane, cache policy, push timing) packaged with evidence + recommendations
   in the cut brief; emeet-pixyd PR #12 verified MERGED upstream (its templ
   residual: generated file stamped v0.3.1100 vs CLI v0.3.1114); TODO rows
   now link the brief.
4. **M04 Store contract helper (the plan's one new package).**
   `broadcast/storecontract` ships `AssertReplayOrdering` +
   `AssertIDAssignment` (testing.TB, negative tests via a failure-capturing
   TB, README section, docspec mirror, CHANGELOG entry). MemoryStore is
   documented as deliberately passing only the ordering slice (its contract
   is caller-assigned IDs). A parallel session's overlapping
   `broadcasttest` package was retracted the same day — storecontract is
   the single helper; the filter-interaction clause split into its own
   TODO row.
5. **M06/M07 upstream truth.** Go SDK check: v1.2.2 current (README
   footnote date-stamped; the repo publishes tags, not release objects).
   JS bundle: pin found **one release stale** (v1.0.3 vs upstream v1.0.4 of
   2026-09-21) — bumped through `fetch-bundle.sh` semantics (verbatim
   download, sha256 `727844ad…` pinned, `static.Version` → v1.0.4,
   checksum test updated, wire-format goldens verified green, CHANGELOG
   recorded). M07's transport-pin contract test landed by a parallel
   session and immediately caught 5 unrecorded pins — all now recorded
   (CHANGELOG [Unreleased] Dependencies + [0.6.3]).
6. **M08/M10/M11 docs tranches.** doc.go, README quick-start pairing,
   CSP+versioned-URL composition note, Renovate delivery-unverified note,
   v1.0.4 scope line, FEATURES Script Serving section (3 rows), script
   handler godoc asymmetry + malformed-URL sentences; the fuzz
   non-decision and testing.md/datastartest-README items were already
   shipped by parallel sessions and verified.
7. **M12 vendorHash mystery solved + gotcha re-verified live.** b713f67 was
   a sibling-require bump (v0.6.2 → v0.6.3) + deps sweep with all three
   hashes re-derived in-commit — correct process, no discrepancy. The
   AGENTS gotcha was then **re-verified live**: this session's comment-only
   root edits + the bundle bytes moved broadcastVendorHash exactly as
   documented, and `buildflow -s nix-hash-fix --fix` landed the exact
   `got:` values (zero manual pastes); AGENTS carries the verification.
8. **M13 gate verdicts.** Broadcast docspec CI wiring confirmed
   (ci.yml:101); nix parity green (flake check all-passed post-hash-fix);
   golden-test scope confirmed wire-format-only (no API-surface golden
   test exists to trip); Renovate probe re-confirmed 403 + zero PRs.
9. **M15/M16/M17 archive dispositions.** Broadcast matrix: 10 evidence
   verdicts (shipped/not-do/routed). 09-02 pointer ledger: every pointer
   verified resolving into archived/ (scripted check) — integrity stamped
   in the ledger preamble. BuildFlow ledger: 23 verdicts (strikes with
   evidence, NOT-DOs with rationale, upstream routings); fuzz smokes ran
   PASS; `erraudit nolint-audit` result recorded (1 needed, 0 stale);
   `.gitattributes` added for the vendored bundle; CONTRIBUTING gained the
   bundle-upgrade section.
10. **M17/M18/M22/M24 residue.** AGENTS truth pass (lint pin v2.14.0 +
    GOLANGCI_VERSION single-source, toolchain 1.27.3, `nix flake check` in
    Commands, result-cache masking gotcha, nix-hash-fix verification);
    M24's BOM test / ETag-stable integration test / RequireETagForBundle
    sha256 helper verified shipped (parallel sessions); policy doc written
    ([2026-11-07_policy-consolidation](../planning/2026-11-07_policy-consolidation.md))
    consolidating the 8 owner questions + the routed-remainder marker spec.

## c) NOT DONE (owner-gated, prepped and waiting)

1. The v0.7.0 cut itself — four decisions in the brief (D1–D4); every
   executable prep step is done, including the §2.5 rehearsal.
2. Upstream filings (BuildFlow go-licenses env alignment, docs-gates
   marker/continuation/link-rotation candidates) — specs exist; filing
   needs the owner's nod on the docs-gates candidates and a BuildFlow
   source-level verification pass before issues go up.
3. Retro-tagging the three missing submodule v0.6.3 tags — recommendation
   is fix-forward at v0.7.0 (brief D2).

## d) TOTALLY FUCKED UP

1. **Trusted a plan-internal PR number.** The plan said "PR #12 (emeet-pixyd
   templ check)" — this repo's PR #12 is an unrelated closed dependabot
   bump. Nearly reviewed the wrong repository's PR; the TODO row now says
   explicitly which repo the PR lives in. Lesson: plan items referencing
   external identifiers need the repo-qualified name at plan time.
2. **Scheiß parallel-lane awareness, again.** I built the storecontract
   package while a parallel session built broadcasttest — the overlap cost
   both sessions a convergence step (theirs retracted). The CLAIMED_LANES
   banner (Q1 in the policy doc) exists precisely for this; it is not
   adopted yet. Second incident of the day class (three sessions duplicated
   the docs-gates sweep this morning).

## e) v0.6.1 retro (the M23 deliverable, backfilled from evidence)

The 2026-09-29 v0.6.1 train (toolchain floor 1.27.1, go-sse v0.6.1 +
ssetest v0.4.0 pairing bump, datastartest replace drops, lint-pin repair)
shipped without a session report; this is the retrospective from CHANGELOG,
tags, and git evidence:

- **What shipped:** the Go floor raise across go.mod ×3 + go.work + CI +
  flake (the 1.26.7 directives could not build under a 1.27 toolchain —
  the `encoding/json/v2` language gate the old CI pin masked); the go-sse
  v0.6.1/ssetest v0.4.0 pairing (datastartest unaffected by ssetest's
  breaking scan-error change — its own `sse_scan_failed` code plus a
  re-wrap preserving the cause); all three submodule vendor hashes
  re-derived; lockstep tags ×4 with proxy verification.
- **What it cost:** the vendor-hash re-derivation loop (three moved hashes,
  one `--keep-going` pass — the incident that produced AGENTS' "nix CI
  reports only the FIRST hash mismatch" gotcha), and the lint-pin repair
  (CI's v2.12.2 buildir panic vs the devshell's green — the incident behind
  the version-pinning rule).
- **What it should have done:** a session report (this retro exists only
  because the docs-health lane kept flagging the gap) and the release-shape
  discipline v0.6.3 later skipped — the checklist's §2/§2.5/§3/§4 steps
  existed by then but v0.6.1's cut predates the hardened checklist.
- **Verdict:** the release itself was healthy (proxy + tags verified
  2026-11-07 during the v0.6.3 investigation); the process debt was
  documentation-shaped and is now paid.

## f) Up to 50 things to get done next (ranked; owner-gated marked)

1. **(owner) Answer the cut brief's four decisions** (D1–D4) and run the
   v0.7.0 cut with the checklist — prep is complete.
2. **(owner) Answer the policy doc's eight questions** (Q1–Q8) — each folds
   into TODO_LIST/ROADMAP or a one-line convention on answer.
3. **(owner) Renovate app installed?** — the v1.0.4 miss strengthens the
   dead-config hypothesis; one Settings check settles it.
4. **(owner) Retro-tag or fix-forward the submodule v0.6.3 tags** — brief
   D2 sub-decision; fix-forward recommended.
5. **(owner) BuildFlow upstream filings** — go-licenses env alignment
   (TODO row), erraudit variant flags, vulnix PATH, silent_swallow repro:
   file after a source-level verification pass in ~/projects/BuildFlow.
6. **(owner) docs-gates upstream candidates** — routed-remainder marker,
   continuation-row mode, link-rotation guard (spec in the policy doc
   appendix); file in Artmann-Technologies/docs-gates if wanted.
7. Broadcaster replay-filter interaction test (carved-out TODO row).
8. **(owner)** Website trigger, Monitoring tier, lint-cache purge — policy
   doc Q3/Q4/Q5.
9. Broadcast API follow-ups row (heartbeat true-off, hub options gap) —
   unchanged from TODO_LIST.
10. Example/godoc polish tranche row — unchanged from TODO_LIST.
11. **(owner)** PR #29 (dependabot codeql bump) review — open on master.
12. **(owner)** Private vulnerability reporting Settings check + PR #23
    disposition (TODO row unchanged).
13. Micro-residue: datastartest diff.go bounds guards; example handler
    testability; buildflow doctor run — bare buildflow-ledger rows 25/27/32.
14. Watch the first Renovate JS bump row (now carries the v1.0.4 evidence).
15. CODEOWNERS, branch deletions, v0.6.1-report waiver — owner-blocked rows
    unchanged.

_(16–50: intentionally not enumerated — everything actionable from this
session's universe is in the fourteen items above plus the linked TODO
rows; minting filler items would be noise, not harvest.)_

<!-- check-rows:ledger -->
