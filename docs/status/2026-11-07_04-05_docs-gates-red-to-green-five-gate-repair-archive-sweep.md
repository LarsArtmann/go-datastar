# Status: docs-gates red-to-green — five-gate repair, honest annotate sweep, 34-report archive

**Date:** 2026-11-07 04:05 CET · **Session window:** ~01:30–04:05 · **HEAD at report time:** `afcf43d` (tree carries 35 dirty paths — the archive sweep's git-mv batch + index rebuild; daemon owns the landing)
**Input:** user pasted a full `docs-gates run` failure matrix (freshness, audit-all, annotate, cadence red; routing green) with the instruction to break it down, execute step by step, and keep going until everything is green.
**Concurrency note:** a second session was live in this checkout for part of the window and shipped the `2026-11-07_01-52` ScriptTagVersioned report plus equivalent annotations to two tables I was also repairing. Collisions were caught by the mod-time guard and converged cleanly. This report covers **my** lane only.

---

## a) FULLY DONE (gates green at wrap-up: PASS ×5, stable across two full runs)

1. **READ before acting.** Loaded the docs-health SKILL.md in full, then read the docs-gates Go source (`markerRe`, tokenization, `GenericDF = 8`, `matchQuorum = 2`, B/HASH + C/PATH + E/claims + D/open-lead rules, F/spacing regex, cadence thresholds, and the critical discovery that **audit-all replays archived reports forever** — so strikes, not moves, are the only permanent resolution).
2. **Evidence before verdicts.** All 19 audit-all-flagged items verified against code/history before disposition: DISCUSSION_TEMPLATE exists (`69ede1b`), ADR 004 exists (`88c1eed`), `docs/performance.md` + `docs/static-js.md` exist (`3fa96f0`), `module_boundary_test.go` uses modfile parsing (`62128bf`, guard itself `fda70c7`), `BenchmarkComputeETag` exists (`b7e6e7f`), `example_depth_test.go` covers exactly CollectPost/CollectN/DataValue/String (`5b9e6ec`), nestif item was a duplicate of its own item 18 (`5bab343`), PR-14 CHANGELOG sentence fixed (`eb2f83f`), go.mod lie resolved later (`7f57346`), v0.5.0 tag commit confirmed (`831bbfb`), `/tmp/g1check` verified absent on disk.
3. **Pre-flight discipline.** Every living-doc addition ran `docs-gates freshness . --pre-flight` BEFORE writing. It caught two real traps: candidate wording with "step" pushed a token to the DF 8 genericity boundary (reworded), and "checklist" was outright UNSAFE (an item depends on it as its only distinctive token — reworded). Four SAFE batches landed: TODO_LIST rows for the Store monotonic-ID contract helper, the upstream release check + README comparison table, and the release-checklist hardening trio; ROADMAP lines for aarch64/darwin `--all-systems`, all-contributors, the bundle-integrity checksum API, and the micro-quality consolidation.
4. **Evidence-backed annotate sweep via the repair verb** (dry-run first, every time): spec 1 = 15 verdicts across 9 files, spec 2 = 58 verdicts across 5 files (superb §f 34 rows, 08-29 §f 12 rows, master-recovery 6, git-town 3, broadcast R1–R3 shipped-in-code verdicts), spec 3 = 10 verdicts (superb b/c tables incl. `static/fetch-bundle.sh` → `e55a4e9`, one-bot decided 2026-10-01, CSP docs shipped). Every §f table in the superb-audit report is now 50/50 struck.
5. **20 verdict-text repairs + 4 spacing fixes** by exact-match edit: struck-but-unmarked verdicts moved onto the marker vocabulary, verdicts unstuck from inside `~~…~~` spans (the AA1/A3 defect class), "still open — owner-blocked" contradictions rewritten as honest routed/owner-blocked verdicts, D/verdict-leads-open fixed, CHANGELOG/ADR spacing repaired.
6. **Pre-archive gates green:** `check-rows` 35/35 files complete, `bare-rows` 35 files clean. 542 bare table rows resolved the honest way: ~60 real per-row verdicts in the eleven mixed tables, 24 ledger tables marked with pure `<!-- check-rows:ledger -->` lines (work logs, verification matrices, commit ledgers, pointer ledgers — each eye-judged before marking), 3 backlog ledgers exempted with written rationale (09-02 pointer ledger, buildflow self-declared brainstorm, broadcast test matrix).
7. **Archive ritual executed in order:** gates BEFORE `git mv`; 34 reports moved with the `> **Live remainder:**` banner naming TODO_LIST/ROADMAP/CHANGELOG as owners; index README rebuilt — Active now holds only the fresh 11-07 report, Archived holds 45 rows including the one-line-per-file manifest (classification + deciding reason per file, as the bulk-archive rule requires).
8. **Final verification:** full battery PASS ×5 twice; zero broken links after the index shuffle (checked programmatically); link targets + display names corrected after my own generation bugs (see d1).

## b) PARTIALLY DONE

1. **Two archived files carry zero inline strikes** (`2026-10-01_04-59`, `2026-10-01_11-33`): their §f prose items are routed (audit-all FRESH = every open item token-covered in living docs) and the Live-remainder banner names the owners, but the strict archive canon ("each line carries a verdict") is not met. I chose routing-verified archival over retroactive per-item strikes because most of those ~89 items lack individually recoverable evidence — striking them would have been the fabrication failure mode. This is a judgment call the owner may want reversed (see g1).
2. **Three backlog ledgers exempted rather than resolved:** the broadcast-review test-matrix residue (~40 rows), the 09-02 pointer ledger (45 rows), and the buildflow brainstorm ledger (38 rows) are exempted, not dispositioned. Several broadcast rows are probably shipped (options tests, replay tests exist) and deserve real verdicts in a follow-up pass.
3. **The parallel session's report is gate-green but not harvested in spirit:** its §f items pass freshness via token coverage, but its own item 3 asks for a TODO_LIST/planning harvest of the Now-tier items — I left that undone (their lane, and the user instruction here was report-and-wait).
4. **Session-end tree not clean at wrap-up:** 35 dirty paths (the git-mv batch + README) were still uncommitted at report time. Per the shared-checkout protocol the daemon owns the landing; I accepted the race rather than racing it, but the "clean tree at end" ritual is therefore not yet satisfied.
5. **No CHANGELOG entry for the sweep:** current policy excludes status reports from CHANGELOG; I assumed an infra docs pass is also excluded (see g3 — assumption unconfirmed).

## c) NOT STARTED (by design or routing)

- **Owner lane untouched:** push timing, v0.7.0 cut, `ScriptHandlerWith` dead-param decision, Renovate-app verdict, the O-lane rows (branch deletions, preserve rehome, CODEOWNERS, erraudit public-flip, website, lint-cache policy, required-checks policy, v0.6.1 retro).
- **The parallel session's Now-tier items** (vendorHash discrepancy investigation, `RequireScriptTagVersion` helper, CSP nonce+versioned-URL doc note) — theirs, not mine.
- **AGENTS.md additions:** the ledger-marker pure-line gotcha (d4) is report-documented only; no AGENTS entry yet.

## d) TOTALLY FUCKED UP (this session's honest list)

1. **The README manifest generation failed three times in a row, and my first two fixes made it worse.** I wrote the manifest with `|` separators and split on TAB in awk → 34 corrupted rows (reason glued into link text, empty topics). Then instead of regenerating from the manifest, I patched with incremental regexes — the second patch appended `.md` OUTSIDE the link parens and a later gsub stripped `.md` from link targets. A third rebuild then clobbered the 11 pre-existing Archived rows (empty topics) because my filter matched them too. The working fix was the thing I should have done first: one clean manifest-driven rebuild + recovering the originals from `git show HEAD:docs/status/README.md`. Lesson: **regenerate from the source of truth, never patch generated output.**
2. **Misplaced ledger marker + glued comment:** my first `<!-- check-rows:ledger -->` for the fuckup-fix F-table landed INSIDE the table (after row F3), splitting it, and I glued the explanation comment onto the marker line — which cancels the exemption entirely ("any other non-blank line cancels it"). All three of my markers were non-functional until the check-rows re-run exposed it. The pure-marker rule is one line in the tool source; I guessed instead of reading it first.
3. **Fired 8 parallel multiedits before View-reading the target files** — all 8 rejected by the read-before-edit guard, a full wasted round. Also sent one edit to the wrong file path (buildflow's marker split aimed at pareto-hardening); only the tool's per-edit failure report caught it.
4. **Did not re-check for concurrent sessions between phases.** `git status`/`git log` ran at session start, then not again until the mod-time guard rejected two edits and the corpus count jumped 45 → 46 reports. The AGENTS shared-checkout section says re-verify around every operation; I treated it as a start-of-session ritual instead of a per-phase one.
5. **Line-number specs built from pre-edit reads without mechanical verification.** Repair specs 2 and 3 used line numbers from earlier views; they all bound, but that is luck as much as method — the skill's emit-keys-first rule exists precisely because silent skips read like the file's fault. The awk clustering for bare-row regions also emitted garbage for wrapped tables (worked around with a second probe instead of fixing the script).

## e) WHAT WE SHOULD IMPROVE

1. **Regenerate, don't patch** (d1): when a generated artifact is wrong, rebuild it from its manifest/source in one deterministic pass; patching generated output multiplies damage.
2. **Read the rule before relying on the mechanism** (d2): the ledger-marker grammar, marker regexes, and exemption semantics are all one `grep` away in the docs-gates source; every guess I made that I later verified was right only by luck.
3. **Per-phase concurrency probe** (d4): `git log`/`git status` before each phase, not per session; a one-line CLAIMED_LANES note at report top (the parallel session's g3 idea) would make the docs/status lane explicit.
4. **emit-keys before specs** (d5): mechanical key emission over unfamiliar file shapes; the dry-run catch is the safety net, not the method.
5. **Pre-flight always** (worked): both token traps were caught by `--pre-flight` before landing; the DF 7→8 boundary warning deserves respect even when the verdict says SAFE.
6. **Resolve the no-strike-archive canon question** (b1): either a fleet-level `routed-remainder` marker convention for prose §f lists, or per-item passes for the two affected files — decide once, not per-archive.
7. **Write §f with harvest in mind:** this report becomes the newest report and its §f will be replayed by the freshness gate on the next run; items below are phrased to route cleanly, and item 1 makes the harvest debt explicit.

## f) Up to 50 things to get done next (ranked; owner-gated marked)

1. **HARVEST this report's §f into TODO_LIST/ROADMAP** (docs-health HARVEST) before the next gate run — this file is now the freshness gate's newest report.
2. Land the 35 dirty archive paths (daemon) and verify a clean tree + `git town status` (session-end ritual, unfinished).
3. Decide the no-strike-archive canon: per-item verdict pass for `2026-10-01_04-59` + `2026-10-01_11-33`, or adopt a fleet `routed-remainder` marker (g1).
4. Real per-row verdicts for the broadcast-review test-matrix residue (~40 ledgered rows; many likely shipped — options tests, replay tests).
5. Verify the 09-02 pointer-ledger rows (45) still resolve now that their source reports are archived.
6. Disposition the buildflow brainstorm-ledger rows (38) — several belong upstream in the BuildFlow repo.
7. **(owner)** Cut v0.7.0 — release-checklist §2.5 first live run; CHANGELOG `[Unreleased]` is release-shaped.
8. Broadcast `Store` monotonic-ID contract test helper (TODO_LIST row from this session).
9. Upstream `starfederation/datastar-go` release check + README comparison table (TODO_LIST row).
10. Release-checklist hardening trio at the cut (TODO_LIST row).
11. **(owner)** Renovate app installed? — first regex-manager proposal or dead-config verdict.
12. aarch64/darwin `nix flake check --all-systems` (ROADMAP theme 3; needs a darwin runner).
13. Bundle-integrity checksum API idea (ROADMAP theme 5).
14. Micro-quality consolidation backlog (ROADMAP theme 2: BOM test, ETag-stable integration test, sha256 ETag helper, Dockerfile smoke, title convention, CHANGELOG→release links, runbook backfill note).
15. VendorHash discrepancy investigation (`git show b713f67 -- flake.nix`; parallel session's Now item).
16. datastartest `RequireScriptTagVersion` helper (parallel item; testing.TB invariant).
17. CSP nonce + versioned-URL doc note in `docs/static-js.md` (parallel item).
18. **(owner)** Cache policy: raise ScriptHandler to `public, max-age=31536000, immutable`?
19. doc.go: mention `ScriptTagVersioned` next to ScriptHandler (parallel item).
20. `ScriptTagVersioned` custom-bundle asymmetry doc comment (parallel item).
21. README quick-start: pair `mux.Handle` with a `ScriptTagVersioned` reference (parallel item).
22. Example smoke test asserting `src="/datastar.js?v=<static.Version>"` (parallel item).
23. README feature-table cache-busting row (parallel item).
24. Godoc `ExampleScriptTagVersioned` (parallel item).
25. **(owner)** Review/merge emeet-pixyd PR #12 (the port's origin; templ regen hunk needs the templ-version check).
26. **(owner)** Release lane for `ScriptTagVersioned`: v0.6.3 root re-tag vs next minor (parallel g).
27. FEATURES.md row + docs-map drift reconcile for the new API (parallel items 15/16).
28. `ScriptTagVersioned` idempotency for URLs already carrying `v=` (parallel item).
29. `appendVersionQuery` malformed-URL edge documentation (parallel item).
30. datastartest README API tour update if item 16 ships (parallel item).
31. `docs/testing.md` versioned-tag assertion mention (parallel item).
32. `nix run .#test` vs raw go test parity check (parallel item).
33. Fleet sweep for the bare-`datastar.js` pattern in sibling web repos (parallel item).
34. Upstream SDK docs: do they recommend a cache-busting query? (parallel item).
35. Verify no cross-module doc claims `ScriptTagVersioned` for `static/` (parallel item).
36. AGENTS.md one-line pointer: versioned URLs are the recommended serving pattern (parallel item).
37. Watch the first Renovate JS bump: confirm `?v=` updates automatically via `static.Version` (parallel item).
38. Confirm no API-surface golden test trips on future exports (parallel item).
39. Re-measure root coverage post-`ScriptTagVersioned`; badge job agreement (parallel item).
40. CHANGELOG hygiene at the cut: `Added` entry survives the `[Unreleased]` rename (parallel item).
41. Record the fuzz non-decision (pure string concat) in `docs/static-js.md` (parallel item).
42. Verify broadcast docspec CI wiring still green (parallel item).
43. AGENTS.md gotcha: `<!-- check-rows:ledger -->` must be a pure line — glued comments cancel it (from d2).
44. Upstream docs-gates candidate: check-rows continuation-row mode (skill parks this proposal).
45. Upstream docs-gates candidate: link-rotation guard (README index links resolve — I checked by hand this session).
46. **(owner)** CHANGELOG policy verdict for infra/docs passes (g3).
47. Confirm the three exempted backlog ledgers are the right disposition or schedule per-row passes (ties to 4–6).
48. Re-run `docs-gates run` after this report lands to size the freshness debt from f) honestly.
49. Keep TODO_LIST purge cadence on the next harvest (currently 0 struck rows — PURGE-HEALTHY).
50. Repeat the session ritual: `git town status`, `git status`, `gh pr list` at next session start (parallel session skipped its CI probe; I skipped town status).

## g) Questions I cannot figure out myself

1. **No-strike archives:** `2026-10-01_04-59` and `2026-10-01_11-33` are archived with routing-verified (token-covered) but unstruck §f lists. Do you want a per-item verdict pass (~89 items, most with thin evidence), or should "routing-verified + Live-remainder banner" become acceptable archive canon (I'd propose a fleet-level `routed-remainder` marker in docs-gates so the state is machine-visible instead of implicit)?
2. **Lane protocol:** your parallel session's report re-asks the checkout-coordination question (their g3). Today two sessions repaired the same gate failure concurrently and converged, but only via the mod-time guard. Do you want the CLAIMED_LANES/worktree protocol, or accept collision-and-converge as the price of throughput?
3. **CHANGELOG policy for infra passes:** status reports are excluded from CHANGELOG by policy. Does a docs-infrastructure pass (gate repair, 34-file archive sweep, living-doc harvests) warrant a CHANGELOG entry under a Docs heading, or does report-only stand? I assumed report-only this session.

---

**Verification snapshot at report time:** docs-gates PASS ×5 (freshness 10ms, audit-all 80ms, annotate 490ms, routing 2ms, cadence 60µs), stable across two full runs · `check-rows` 35/35 complete · `bare-rows` 35 files clean · 45 archived rows in the index, zero broken links · TODO_LIST +3 rows, ROADMAP +4 lines, all pre-flighted SAFE · 34 reports archived with banners + manifest · 35 dirty paths pending daemon commit.
