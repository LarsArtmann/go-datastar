# Status Report — 2026-09-02 23:55 — Docs-Health Audit Over ALL 2026-0* Files: Annotate, Harvest, Living-Doc Sync

- **Session scope:** Full docs-health AUDIT (BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE
  decision) over every `**/2026-0*` file: 26 active status reports, 1 active plan,
  4 archived statuses, 2 archived plans, 2 modularization HTMLs — plus all six
  living docs (README, AGENTS, FEATURES, TODO_LIST, ROADMAP, CHANGELOG).
- **Report time state:** master at `1409b0a` (auto-commit daemon swept this
  session's work in batches `73025c0`, `907cb63`, `1409b0a`); `ROADMAP.md` +
  `TODO_LIST.md` dirty at report time (final TODO row edit), expected to be
  swept. Tree otherwise clean; branch up to date with origin at session start.
- **Format note:** `.md` per repo convention (status-report HTML default
  overridden by AGENTS.md policy).

---

## a) FULLY DONE

### 1. Inventory + skill load — all 31 `2026-0*` files accounted for

26 active `docs/status/*.md`, 1 active `docs/planning/*.md`, 4 archived
statuses, 2 archived plans, 2 `docs/modularization/*.html` (content not
re-verified — see b.2). The docs-health SKILL.md and 6 of its reference docs
(doc-ownership, harvest-guide, verify-checklist, resolving-items,
annotation-placement, health-report-format) read before any edit; the two
batch-annotation scripts' interfaces dry-run-verified before first real use.

### 2. VERIFY — concrete claims checked against code, not trusted

- **Coverage re-measured, not asserted:** root **98.4%**, datastartest
  **92.7%** (`go test -cover -count=1` at HEAD). The 20-12 report's Forgotten
  #9 suspected prose drift — disproven; the printed numbers are current.
- **erraudit loop run clean on all three modules** (0 violations each) —
  closes the 20-12 report's Forgotten #2 / f.6 audit gap.
- **External state verified:** pkg.go.dev renders **v0.3.0 for all three
  modules** (fetched root, `static@v0.3.0`, `datastartest@v0.3.0` — both
  sub-modules "Latest", Published Aug 29); `gh release list` shows
  v0.0.1→v0.3.0 (Latest); `gh release view v0.3.0` shows the 8.9KB
  CHANGELOG-excerpt body; `go run ./example/` verified serving on :8765;
  go-sse v0.6.0 source grepped — **OnDrop remains Broadcaster-only** (closes
  the 02-58 watch item); `git config git-town.observed-branches` absent
  (confirms 20-12 f.20 still open); both `.github/dependabot.yml` AND
  `renovate.json` still coexist (20-12 Q2 still open); `wire_golden_test.go`
  committed (`a0c0aea`) with its CHANGELOG bullet now accurate.
- **Full quality gate green at HEAD:** race ×5 packages, vet ×3 modules,
  golangci-lint 0 issues ×3 modules, `GOWORK=off` build ×3,
  `go work sync` idempotent, replace-directive audit clean, all internal
  markdown links resolve (README/AGENTS/TODO_LIST/ROADMAP/FEATURES +
  docs/status index + docs/testing.md relatives).
- **Cross-file consistency:** no completed items in TODO_LIST, no
  TODO↔CHANGELOG or TODO↔ROADMAP duplication, status vocabulary correct,
  dprint-vs-treefmt documentation matches ADR 006.

### 3. ANNOTATE — every numbered item in every historical file given a verdict

Using the skill's `annotate-prose.py` / `annotate-rows.py` (mandatory dry-runs,
atomic writes, shape-checked read-backs), roughly **370 inline verdicts**
(`~~struck~~ done at <hash>` / `done — <verified evidence>` /
`**Won't implement — <reason>**`) across:

- All fourteen 08-07/08-08 reports (typed-error-system, both retrospectives,
  deep-review, docs-health-audit, pareto-hardening, fuckup-fix, 09-18, 09-36)
- All ten 08-10 reports (02-55 → 07-55)
- Both 08-13 go-sse reports
- The 08-16 trio (11-07, 12-24, 12-43) — closing their release/CI/guides
  residue
- Both 08-29 reports (17-05 audit report, 20-12 execution debrief incl. 16 of
  its 50 f-table rows)
- The 18-32 plan: **26 of 27 task rows struck with commit hashes** +
  resolution appendix pointing at the 20-12 execution report (T04 left bare =
  still owner-blocked)

Six marker-semantics violations I introduced myself (see d.1) were caught and
reverted to bare-with-routing-notes. Open items stay bare throughout —
absence of a marker IS the open signal.

### 4. HARVEST — the open residue is now in living docs, deduplicated

- **TODO_LIST.md rebuilt:** 15 rows in two sections (Verified next-up /
  Owner-blocked), every row evidence-cited (report + code path). New rows:
  FOD/vendorHash investigation + ADR 004 correction, exact-CI golangci-lint
  app, tag-worktree flake check, json/v2 + `b.Loop()` gopls cleanup, CI
  hygiene batch (go mod verify / drift / use-match), docker build,
  performance benchtime re-run, Response test-depth batch, v0.4.0 scoping,
  CI-workflow first-run watch (incl. the Renovate regex check that the
  previous TODO rebuild had dropped), micro-hygiene batch, modfile upgrade of
  module_boundary_test, community-files batch, CodeRabbit replies + human
  review, and the Renovate-vs-Dependabot owner decision (previously missing).
- **ROADMAP.md updated:** shipped idea removed (typed script accessors —
  `671e57c`); new raw ideas added: consolidated datastartest helper-API
  expansion + internal polish (absorbs ~40 repeated micro-items across four
  08-10 reports), Response ergonomics (ErrorResponse-as-methods, signalsMap),
  compile-checked doc snippets, CI per-module matrix, `go work vendor`,
  example-module question, coverage-floor policy, compat-test matrix,
  Stream-OnDrop watch, `static.Version`↔CHANGELOG constraint check. The
  one-bot decision was routed to TODO_LIST only (single-home rule).

### 5. BUILD — living docs brought current

- **FEATURES.md:** +8 features across 3 sections — ReplaceURLQuerystring,
  typed script accessors, wire-format goldens, new **Examples** section
  (live-feed demo, domain-adapter, gzip middleware, Docker packaging), and 5
  new Build & CI rows (nix.yml, fuzz.yml, codeql.yml, Renovate, docs guides).
- **README.md:** `NewReplaceURLQuerystringPatch` + `Response.ReplaceURLQuerystring`
  rows, a **Documentation** section linking all seven consumer guides,
  `static.Bytes()/static.Version` row, wire-golden note in the parity section.
- **AGENTS.md:** CI section synced (nix/fuzz/codeql/renovate/SHA-pinning) and
  Docs Map extended with the seven guides.
- **CHANGELOG.md:** `[Unreleased]` gained the go-sse v0.6.0/ssetest v0.3.0
  bump entry (daemon commit `2e8593d` was unrecorded) and a **Fixed** section
  recording the tag-flake vendorHash incident and the varnamelen linter-parity
  fix (20-12 f.17).
- **datastartest/doc.go:** stale "parent hand-rolls parsing" paragraph
  rewritten (ssetest delegation + the three error codes) — pkg.go.dev renders
  it on next release. **example/README.md:** new "Drop observability" section
  documenting the Broadcaster-only OnDrop design (closes 02-58 f.10).

### 6. ARCHIVE decision — zero moves, honestly

After full annotation, **no active report qualifies for archiving**: every one
retains routed-open residue (ROADMAP-grade ideas, owner-blocked decisions, or
CI-watch items). Per the `docs/status/README.md` policy ("archive only when
EVERY item is resolved"), moving them would fake resolution. The prior 5-file
archive set (08-16) was spot-verified as correctly archived.

### 7. Health report emitted inline (per skill: not written to a file)

Accuracy **9.75/10** (one Low: README option-table rows verified by sampling,
not exhaustively), Fitness **10.0/10** — against the 2026-08-29 baseline of
9.0/10.0.

---

## b) PARTIALLY DONE

1. **Annotation completeness vs routing depth.** ~370 items closed, but the
   large 08-10 files still carry 20–39 bare items each (04-25: 39, 03-49: 31,
   06-00: 30). These are deliberately-open micro-ideas, but most are covered
   only by ONE consolidated ROADMAP line, not individually cross-referenced —
   a reader of the bare item must trust the consolidation.
2. **Agent-verdict reliance.** Four subagents classified ~300 items in the
   older reports; I spot-checked evidence but did not independently re-verify
   every DONE citation. Fabrication risk was reduced (strict prompt, UNCERTAIN
   allowed, spot checks passed) but not eliminated by my own eyes.
3. **AGENTS.md grew during a sync again** (16.6KB → 17,969B with the CI/Docs
   Map additions) — above the 5–15KB target, under the 30KB flag line. The
   17-05 report flagged the same pattern; I repeated it because the sync was
   load-bearing.
4. **Agent 5 stalled** (returned a one-line "Now checking…" instead of its
   report); I covered its four files myself, at the cost of the parallelism.
5. **docs/performance.md prose numbers** were NOT re-verified against fresh
   benches (only coverage was); the benchtime re-run is routed TODO_LIST.
6. **README comparison table vs upstream datastar-go v1.2.2** — not re-verified
   this session (quarterly standing check; left per release checklist).
7. **actionlint output was not captured** in the gate batch (the fallback
   chain swallowed it) — the workflow YAML validation is therefore unverified
   this session despite the command being issued.
8. **Hand-edited table rows** (README, FEATURES, TODO_LIST) are pipe-aligned
   by hand, not by a formatter; cosmetic padding drift possible.
9. **The two modularization HTML files** were inventoried but their content
   was not re-verified (the README already marks them executed).

## c) NOT STARTED

1. `git worktree` of tag `v0.3.0` + `nix flake check` → record the expected
   FAIL verdict (TODO_LIST).
2. `docker build example/` + compose up (TODO_LIST).
3. Exact-CI golangci-lint app in flake.nix (TODO_LIST, High).
4. FOD/vendorHash sensitivity investigation + ADR 004 correction (TODO_LIST,
   High).
5. gopls cleanup: 4 `encoding/json` → `encoding/json/v2` call sites + 4
   `b.N` → `b.Loop()` (TODO_LIST).
6. CI hygiene batch: `go mod verify`, version-drift detection, `go.work`
   use-vs-disk check, tidy-check mode for module_boundary_test (TODO_LIST).
7. Renovate-vs-Dependabot one-bot decision — owner (TODO_LIST, BLOCKED).
8. Branch deletions (`pr/docs-test-consolidation`,
   `preserve/status-report-coderabbit-pr3`) — owner (TODO_LIST, BLOCKED).
9. CodeRabbit thread replies on PR #3 + human review of the 5 parallel
   commits (TODO_LIST).
10. Community files batch: CODEOWNERS / SUPPORT.md / DISCUSSION_TEMPLATE
    (TODO_LIST, owner).
11. Status-index row for THIS report (added immediately after writing it —
    see the index).
12. dprint-style realignment of the hand-edited tables.
13. AGENTS.md pruning pass back toward ≤15KB.

## d) TOTALLY FUCKED UP

1. **I violated the strike=done-only rule I was enforcing.** Annotating the
   12-24 report, I routed five open items (CodeRabbit replies, preserve-branch
   rehoming, coverage-floor, branch cleanup, commit review) plus one g-question
   through the script's `v` kind, producing `~~item~~ done (still open — …)` —
   a reader scanning for markers sees "done". Exactly the G6 hazard the 18-32
   plan's guard table names. Caught within minutes by re-reading the diff;
   fixed via a scripted revert to bare text + explicit "_Still open — routed…_"
   notes. Root cause: batching open items into a done-marker run instead of
   leaving them untouched.
2. **My first link checker reported 31 false BROKEN links** — I ran it from the
   repo root against relative links whose base is `docs/status/`. Second run
   from the correct bases cleared everything. One wasted cycle on a bug in my
   own verification tool, not the docs.
3. **A heredoc patch failed its exact-match assertion** (fuckup-fix c-table:
   row whitespace differed from my transcription) after the annotate call in
   the same bash invocation had already succeeded — partial-failure inside a
   batched call; recovered with a re-grepped precise patch. The lesson is the
   repo's own: never retry with guessed text.
4. **Two mod-time races with my own annotate scripts** — edit-tool calls
   rejected because the script had rewritten the file seconds earlier (07-38
   G-notes, 03-25 Q2). The "re-read before every edit" rule, bitten twice in
   one session.
5. **One atomic-refusal round trip on 02-55 c4** — the script correctly refused
   an item whose continuation line already carried an inline `→ done` marker.
   I should have grepped for pre-existing markers before batching specs.
6. **Agent 5's truncated output went undetected until I read its result** — a
   completion sentinel in the prompt would have made the stall obvious sooner.
7. **Spot-check, not full re-verification, of ~300 agent-classified verdicts**
   (b.2) — efficient, but the accuracy claim for those files rests partly on
   the agents' evidence citations.

## e) WHAT WE SHOULD IMPROVE

1. **The annotation script needs an explicit "open/routed" kind** (or the
   operator must dry-run-scan for `done (still open` before applying) — the
   `v` kind's `done (<value>)` shape makes open-item routing look resolved.
   This is the highest-leverage fix; it converts my d.1 from "careful human
   catch" into "impossible by construction".
2. **Grep for pre-existing inline markers before batch runs** (pattern
   `→ done|done —|done at`) so atomic refusals are predicted, not hit.
3. **Sequence script-rewrites and edit-tool calls**: script → re-view → edit.
   The mod-time race is self-inflicted and fully predictable.
4. **Run link checkers from the document's directory** (or resolve bases
   explicitly) — never from the repo root against relative links.
5. **Subagent prompts need a completion sentinel** ("END OF REPORT — N items
   classified") so a stalled agent is detectable without reading prose.
6. **Record the spot-check ratio** when relying on subagent classifications
   (this session: ~10% sampled). Make it a stated input to the health score.
7. **AGENTS.md pruning is now overdue** — every docs session has added more
   than it removed; next docs touch should remove or condense
   resolved-incident gotchas (target ≤15KB).
8. **Consolidated ROADMAP ideas should cite their source items** (report +
   item IDs) so the bare residue in old reports is auditable against the
   consolidation.
9. **The archive bar may be too binary** — ~20 active reports now hold only
   routed ideas and owner-blocked questions; a "monitoring" tier in the index
   would separate "nothing left to do here" from "owner input pending"
   without faking resolution (owner call — see g.1).
10. **Gate batch hygiene:** the actionlint output silently vanished in the
    combined gate command; capture per-check exit codes explicitly (the
    17-05 report's "never read exit codes through a pipe" lesson, new
    variant).

## f) Up to 50 things to get done next

Items 1–15 are the verified TODO_LIST (harvested this session); 16–50 are
session-specific follow-ups and smaller residue, ranked roughly by value.

| #  | Thing                                                                                                                     | Source            |
| -- | ------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| 1  | Investigate datastartest FOD source-sensitivity; correct ADR 004                                                          | 20-12 f.3/e.2     |
| 2  | Add exact-CI golangci-lint app (`go run …@v2.12.2`) to flake; make it the pre-push habit                                   | 20-12 f.1/e.1     |
| 3  | Verify tag `v0.3.0` flake state via worktree; record verdict                                                               | 20-12 f.2         |
| 4  | Migrate 4 json call sites to encoding/json/v2; `b.N`→`b.Loop()` ×4 (clears gopls stdversion/bloop)                         | 07-38/07-55 f-items |
| 5  | CI hygiene batch: go mod verify + version-drift + go.work use-match + tidy-check test                                      | 07-38 f.9-f.11 et al. |
| 6  | `docker build example/` + compose up                                                                                       | 20-12 f.11        |
| 7  | Performance.md benchtime re-run; correct prose numbers                                                                     | 20-12 f.21        |
| 8  | Response test-depth batch (JSON-parse asserts, integration/fuzz/bench, edge cases, concurrency, large patches, query+body) | 07-04 f.11-f.33   |
| 9  | Plan v0.4.0 scope from `[Unreleased]` (go-sse v0.6.0, goldens, accessors, ReplaceURLQuerystring, guides)                   | 20-12 f.41        |
| 10 | Watch first runs: fuzz.yml nightly, CodeQL results, nix.yml promotion, Renovate regex vs real upstream tags                 | 20-12 f.12-f.16   |
| 11 | Micro-hygiene: heartbeat ticker constant, http.NewResponseController in gzipSSEWriter, stable bench headers, PR-template module row | 20-12 f.34/f.35/f.47 |
| 12 | module_boundary_test → golang.org/x/mod/modfile parsing                                                                    | 07-55 f.14        |
| 13 | Community files: CODEOWNERS / SUPPORT.md / DISCUSSION_TEMPLATE (owner)                                                     | 07-04 f.23/f.46   |
| 14 | Reply to 5 CodeRabbit threads on PR #3; human-review the 5 parallel commits                                                | 12-24 b.1/c.6     |
| 15 | Coverage/lint-cache measurement ritual after code changes (close 12.3 debt)                                                | 20-12 f.32        |
| 16 | Configure `git-town.observed-branches` for `preserve/…` while deletion is blocked                                           | 20-12 f.20        |
| 17 | Owner: Renovate vs Dependabot — disable the loser; set dashboard/label prefs                                                | 20-12 f.4/f.29/f.37 |
| 18 | Owner: branch deletions + lineage prune                                                                                    | 20-12 f.18/f.19   |
| 19 | Owner: erraudit hard-gate flip verification once the repo goes public                                                       | 11-07 §5          |
| 20 | AGENTS.md pruning pass → ≤15KB (drop resolved-incident gotchas)                                                             | 17-05 e.5/f.50    |
| 21 | Re-align the hand-edited markdown tables (README/FEATURES/TODO_LIST)                                                       | this session b.8  |
| 22 | Add this report's row to docs/status/README.md (done at write time)                                                        | policy            |
| 23 | Verify actionlint locally with captured exit code (this session's gate gap)                                                | this session b.7  |
| 24 | Cross-reference consolidated ROADMAP ideas back to source report item IDs                                                   | this session e.8  |
| 25 | Spot-check a sample of agent-classified verdicts (record the ratio)                                                        | this session b.2  |
| 26 | Decide "monitoring" tier for the status index (owner; see g.1)                                                              | this session e.9  |
| 27 | Re-verify README comparison table vs upstream datastar-go (quarterly)                                                       | release checklist |
| 28 | Evaluate `checks.govulncheck` hermetic derivation if the ADR-004 investigation shows it's cheap                              | 20-12 f.31        |
| 29 | Compile-checked doc snippets (docspec test target)                                                                          | 20-12 f.22/e.6    |
| 30 | datastartest helper-API expansion (consolidated ROADMAP idea — pick the top 3 helpers first: RequireElementsOrdered, Diff, Snapshot) | 08-10 residue |
| 31 | Response ergonomics idea: ErrorResponse-family as Response methods; signalsMap type                                          | 07-04 f.11/f.17   |
| 32 | `example/` own-go.mod structural decision                                                                                   | 07-38 f.28        |
| 33 | Coverage-floor policy decision (optional CI gate)                                                                           | 12-24 c.3         |
| 34 | static fuzz targets in fuzz.yml (bundle bytes are parser-adjacent)                                                          | 02-57 f.30        |
| 35 | TestScriptHandlerWith with an empty custom bundle                                                                           | 02-57 f.31        |
| 36 | static provenance comment (upstream SHA) + bundle checksum test                                                             | 02-57 f.18/f.19   |
| 37 | Evaluate Last-Modified + nosniff headers for ScriptHandler                                                                  | 02-57 f.25/f.48   |
| 38 | computeETag benchmark + cache-per-call evaluation                                                                            | 02-57 f.26/f.27   |
| 39 | `DatastarJSVersion` alias deprecation comment                                                                               | 02-57 f.22        |
| 40 | ScriptTag edge-case tests (empty, query, fragment)                                                                          | 02-57 f.49        |
| 41 | datastartest doc.go/examples gap: CollectPost/CollectN/DataValue/String examples                                             | 03-49 c.9         |
| 42 | datastartest CHANGELOG decision (root CHANGELOG sections vs module file)                                                    | 04-48 f.49        |
| 43 | `go work vendor` support / flake app for offline graphs                                                                      | 04-48 f.47        |
| 44 | CI per-module matrix / parallel jobs for faster feedback                                                                    | 05-07 f.43        |
| 45 | Constraint check tying static.Version mentions to the CHANGELOG                                                             | 20-12 f.43        |
| 46 | CI job summary table in AGENTS after promotion decisions settle                                                             | 20-12 f.45        |
| 47 | Re-run docs-health VERIFY on the seven guides after the next code change                                                    | 20-12 f.39        |
| 48 | Bench-marking: stable `nix run .#bench` output headers for comparisons                                                      | 20-12 f.47        |
| 49 | Favicon/branding assets if the website deferral flips to go                                                                  | 20-12 f.49        |
| 50 | Sweep the TODO_LIST watch rows after the first CI week; close or convert                                                    | 20-12 f.48        |

## g) Questions I cannot answer myself

1. **Status-index tiers:** ~20 active reports now hold only routed-idea and
   owner-blocked residue, so they can never archive under the current
   "EVERY item resolved" rule. Do you want a third tier in
   `docs/status/README.md` (e.g. "Monitoring — residue routed to
   TODO_LIST/ROADMAP; no open in-repo work") so they leave the active table
   without faking full resolution — or should the strict two-tier policy
   stand and the active table just keep growing?
2. **Consolidation depth:** the ~200 bare micro-items in the 08-10
   datastartest reports (missing helpers, file splits, test edges) are
   covered by ONE consolidated ROADMAP idea, not per-item annotations. Is
   that sufficient disposition for you, or do you want each bare item
   struck with a "consolidated → ROADMAP Theme 2" marker so no bare line
   ever reads as forgotten work?
3. **AGENTS.md size policy:** the CI-section sync pushed it to 17,969B
   (target ≤15KB, flag line 30KB). Prune now — cutting or condensing the
   git-town landing-ceremony material and the older multi-module gotchas —
   or accept the size until the next planned docs touch so this session's
   additions settle first?

---

## Appendix: 2026-09-03 execution-pass resolutions (pareto plan T01–T27)

Section b caveats closed by the 2026-09-03 execution session:

2. **Agent-verdict reliance — spot-check performed:** 10% sample (5 of ~50
   classified items in `2026-08-10_07-27`): `fd3a5ac` (FuzzReadEvents exists,
   diff touches reader_fuzz_test.go), `eb8bf29` (ci.yml modified, erraudit +
   govulncheck jobs), `dc0d6f2` (ci.yml, go.work sync hardening), `3cd669e`
   (AGENTS.md ×2, circular-dep fix), `06bb019` (datastartest request options +
   README). **5/5 pass** — every hash reachable and its diff matches the claim.
5. **performance.md numbers re-verified** with `-benchtime=1s`
   (2026-09-03, multi-million samples): tables + methodology updated.
6. **README comparison table re-verified against upstream v1.2.2** via
   pkg.go.dev (compression built-in ✅, no broadcast/replay ✅,
   ReplaceURLQuerystring upstream-only ✅ — README's wins-list already updated
   when go-datastar added it).
7. **actionlint captured and green** (nix shell run, 2026-09-03) and on every
   push via the actionlint workflow (green on the v0.4.0 release commit).
3. **AGENTS.md now 14.8KB** (≤15KB target met by the pruning pass).
1./8./9. remain as stated (consolidation depth is owner-questioned; table
padding is cosmetic; HTML files remain inventory-only).

Section c items 1–4 (tag verdict, docker, exact-CI lint app, FOD
investigation) were ALL executed 2026-09-03: the tag verdict and the FOD
mechanism (with the never-converging self-reference discovery) are in ADR
004; `nix run .#lint-ci` is the documented pre-push gate; docker builds and
serves live SSE.

_Point-in-time snapshot. Written by the 2026-09-02 docs-health session
(AUDIT over all `2026-0*` files: VERIFY + ANNOTATE + ARCHIVE decision +
HARVEST + living-doc BUILD). See `docs/status/README.md` for the index and
archiving policy._
