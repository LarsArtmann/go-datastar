# Full Execution Status: Pareto Plan T01–T27, v0.3.0, Green Master — and the Honest Debrief

**Date:** 2026-08-29 20:12 CEST
**Scope:** this session's execution run only (plan at
[`docs/planning/2026-08-29_18-32_pareto-green-master-and-v0.3.0-plan.md`](../planning/2026-08-29_18-32_pareto-green-master-and-v0.3.0-plan.md)).
Point-in-time snapshot — verdicts live here, per docs-health policy.
**End state:** master = `8cc56a7`, clean except one foreign untracked file,
all five CI workflows green, v0.3.0 shipped and rendered on pkg.go.dev.

## Verdict at a glance

| Tier           | Tasks                           | Verdict                                                      |
| -------------- | ------------------------------- | ------------------------------------------------------------ |
| 1% (T01–T03)   | toolchain, green master, v0.3.0 | ✅ fully done                                                |
| 4% (T04–T09)   | trust surface                   | ✅ 5/6 done, T04 blocked by design                           |
| 20% (T10–T20)  | CI depth, docs, features        | ✅ done (T20 includes decision spike)                        |
| Rest (T21–T27) | long tail                       | 🟡 5 done, T21/T27 delivered as scoped deferrals, not spikes |

17 commits pushed. Every gate re-verified at HEAD: race ×5 packages, vet,
lint 0 issues ×3 modules, `GOWORK=off` isolation ×3, `go work sync`
idempotent, govulncheck clean, `nix flake check` all-passed, actionlint
clean, and all five CI workflows **success** on the final push (run
33266637847 et al.).

---

## Self-critique first: what I forgot, what I did worse, what to improve

### Forgotten (caught late or by this report)

1. **The exact-CI linter build.** I verified lint locally with the devShell
   golangci-lint, but CI go-installs `v2.12.2` — a build that flags
   `varnamelen` on `s` where mine didn't. Result: one red master push
   (run 33266437052), fixed one commit later. I used exact-version
   `go run` for actionlint but not for golangci-lint. Asymmetric discipline.
2. ~~**`erraudit` local run.** The AGENTS command block includes it; my final~~ done (done — erraudit loop run clean on all three modules (docs-health pass 2026-09-02: 0 violations))
   ~~gate skipped it. The new classified error in `script_accessors.go`~~
   ~~(`datastartest.custom_event_detail_unmarshal_failed`) is unverified by~~
   ~~the audit tool this session.~~
3. ~~**FEATURES.md was never updated** with today's additions (typed~~ done (done — FEATURES.md updated 2026-09-02 (accessors, domain-adapter, guides, CI workflows, compression, ReplaceURLQuerystring, goldens))
   ~~accessors, domain-adapter example, guides, new CI workflows). Last~~
   ~~touched by the morning audit (`011104d`).~~
4. ~~**AGENTS.md CI section went stale the same day I wrote it** — it lists~~ done (done — AGENTS.md CI section synced 2026-09-02 (nix.yml, fuzz.yml, codeql.yml, renovate.json, SHA pinning))
   ~~ci.yml/actionlint.yml/coverage.yml but not `nix.yml`, `fuzz.yml`,~~
   ~~`codeql.yml`, `renovate.json` (added later in T24).~~
5. ~~**No cross-links from README to the seven new guides** (plan sub-task~~ done (done — README 'Documentation' section + AGENTS Docs Map link all seven guides (2026-09-02))
   ~~16.3). The guides exist; the README never points at them.~~
6. **Renovate + Dependabot now run in parallel** — I added `renovate.json`
   without checking that `.github/dependabot.yml` exists. Two dependency
   bots = duplicate PR noise. Confirmed by `ls` during this report.
7. **Dockerfile/compose never built.** `example/Dockerfile` +
   `docker-compose.yml` are written but no `docker build` was run.
8. **The docs-only fast path (05.3) was never demonstrated** — every push
   mixed code and docs, so the skip-CI behavior is configured but unobserved.
9. ~~**Coverage claims in testing.md/performance docs carry the morning's~~ done (verified 2026-09-02 — re-measured 98.4% (root) and 92.7% (datastartest); prose numbers are current)
   ~~numbers** (98.4%/92.7%) — accurate at the audit, not re-measured after~~
   ```300 lines of new code today (badge is current; the prose is stale-ish).~~
   ```
10. **Session-entry ritual partial** — `git town status` and `gh pr list`
    were skipped at session start.

### Done worse than it should have been

11. **v0.3.0 was tagged with a stale datastartest `vendorHash`** — the
    lockstep requires bump (in the same release commit) invalidated the hash
    harvested an hour earlier, and my in-flight `nix flake check` (started
    19:14) sat unread in a background job while I tagged. Empirically: the
    FOD at the tag produces `GciltFE7…`, the tag's flake specifies
    `nkJghg…` → **`nix flake check` at tag `v0.3.0` fails**. Master is
    fixed (`1f3cd93`, `G0orQHjx…`), consumers via `go get` are unaffected,
    and re-tagging is forbidden (proxy poisoning) — but the canonical gate
    is red at the tag. See (d).
12. **Module-boundary violation in my own test.** I wrote
    `example/domain-adapter/main_test.go` importing `datastartest` — the
    exact dependency the repo spent an ADR and a regression guard
    eliminating. Caught only because the root `GOWORK=off` isolation run
    failed at session end, not by design review. Lucky catch, not a system.
13. **Background jobs outlived their reading.** Job 087 (`nix flake check`
    after the bench/flake edits) was launched and never drained; its likely
    failure was invisible to me for two hours.
14. **Commit hygiene wobbled at the end**: `nix fmt`'s six-file churn rode
    along in the vendorHash fix commit; the CHANGELOG I staged in
    `1dda530` contains the parallel session's `wire_golden_test.go` bullet,
    and that file is **still untracked on master** — the released CHANGELOG
    references a file the repo doesn't contain. Coordination debt from a
    shared checkout, but my staging decision caused the inconsistency.
15. **Renovate's tag-scheme regex is a guess twice over** — first
    `^client/…`, then `^v?…`, neither verified against actual upstream
    release tags. It is flagged in TODO_LIST, but it shipped unverified.

### What to improve (systemic, not task-level)

- **Linter version parity as a gate**: lint locally with the _exact_ CI
  build (`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2`)
  or add a flake app pinning it — the "green locally, red in CI" class dies
  only with build parity.
- **vendorHash structural fragility needs a real ADR update**: evidence
  today suggests the datastartest FOD hash moved with _tree state_, not just
  go.mod — likely because the local `replace => ..` puts repo source into
  the module graph under `proxyVendor`. ADR 004 currently blames "Go patch
  bumps"; the mechanism may be broader (and explain the daemon-era breaks).
  Investigate; if confirmed, the fix is excluding local replaces from the
  module FOD, not repeated hash dances.
- **Background jobs are promises**: drain or kill before advancing.
- **Ask blocking owner questions at session start**, not at report time —
  T04 sat blocked for the whole run while the answer (below, section g) was
  one question away.
- **Tag only from a quiescent, fully-gated tree** — the release commit was
  gated, but the gate evidence predated the release commit's own go.mod
  change.

---

## a) FULLY DONE

| Task                                               | Result                                                                                                                                                                                                                                           | Evidence                                                    |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| T01 toolchain settle (adopt 1.26.7, G2 atomic set) | go.mod ×3, go.work, ci.yml ×5, coverage.yml, flake `overrideAttrs` + tarball hash + root vendorHash re-discovered; `go work sync` idempotent                                                                                                     | `b37d11a`; sync idempotency re-run at session end           |
| T02 green master                                   | 4× `errors.As`→`AsType` (type extraction only; `errors.Is` sentinels untouched per G4), mnd heartbeat const, makezero append; race/vet/lint green                                                                                                | `489256b`; local 0 issues; CI green on follow-ups           |
| T03 v0.3.0 release                                 | checklist gate fully executed; CHANGELOG promoted + supersession note; lockstep tags ×3 pushed; GitHub Release with CHANGELOG excerpt; `go get` ×3 resolves from proxy; pkg.go.dev renders v0.3.0 (root page verified)                           | tag `v0.3.0`; release page; `go get` logs; pkg.go.dev fetch |
| T05 CI path filters                                | `paths` allowlists on ci.yml + coverage.yml; `actionlint.yml` runs on every push (docs-only signal)                                                                                                                                              | `5887043`; actionlint workflow green in CI                  |
| T06 AGENTS.md pruning                              | 28,711 → 16,635 bytes; file-layout + datastartest API tables → pointers to `doc.go` / `datastartest/README.md`; resolved protected-master ceremony removed; G9 cross-reference check done (only point-in-time reports referenced pruned content) | `5887043`; `wc -c`                                          |
| T07 status index                                   | `docs/status/README.md`: 25 active rows + archived table + policy + planning/archived pointer                                                                                                                                                    | `12a2de4`                                                   |
| T08 PR template honesty                            | local-run boxes kept with "tick only what you ran"; CI-attest boxes removed with the reason stated                                                                                                                                               | `5887043`                                                   |
| T09 CONTRIBUTING fuzz section                      | 4 targets table, 30s smoke commands, crash-seed policy; both documented commands executed (PASS)                                                                                                                                                 | `5887043`; fuzz smoke runs                                  |
| T10 nix CI job                                     | `nix.yml` (pinned install-nix-action v31, `continue-on-error`, code paths only); **green in CI on its first master run** (32s)                                                                                                                   | `88c1eed`; run 33266637891 success                          |
| T12 lint caching                                   | `actions/cache` (SHA resolved via API, not guessed) for `~/.cache/golangci-lint` keyed on go.sum; rationale comment (setup-go already covers build/module caches)                                                                                | `88c1eed`                                                   |
| T13 ADR 003 error classification                   | context/decision/consequences incl. family table, `error`-interface rationale, go-sse layering                                                                                                                                                   | `cf19bf1`                                                   |
| T15 ADR 005 coverage strategy                      | badged-not-gated decision with soft-floor thresholds                                                                                                                                                                                             | `cf19bf1`                                                   |
| T16 guides A                                       | `docs/replay.md` (APIs verified against go-sse source: `BroadcastMany`, `Stream.LastEventID`), `docs/error-system.md`                                                                                                                            | `3fa96f0`                                                   |
| T17 guides B                                       | `docs/wire-format.md` (annotated datalines, parity notes), `docs/testing.md`                                                                                                                                                                     | `3fa96f0`                                                   |
| T19 ReplaceURLQuerystring                          | `NewReplaceURLQuerystringPatch` + `Response.ReplaceURLQuerystring`; upstream semantics confirmed from the v1.2.2 module source; wire/parity tests, Response table case, testable Example; README gap row removed; ROADMAP idea removed           | `3d3cba0`                                                   |
| T20 SSE compression                                | decision: middleware over in-library (README honesty kept); working `gzipSSEMiddleware` with per-event flush, Vary, Content-Length strip; 2 tests green                                                                                          | `8f190ea`                                                   |
| T22 typed script accessors                         | `RedirectURL`, `CustomEventName`, `CustomEventDetail`, `UnmarshalCustomEventDetail` (new classified code), `ScriptAttributes`; 4 tests green; datastartest README section; AGENTS codes updated                                                  | `671e57c`                                                   |
| T23 domain-adapter example                         | domain events → `Bridge() → []Patch` → broadcaster+MemoryStore; boundary-clean E2E (raw SSE, no datastartest); README pointer                                                                                                                    | `efde465`, `1dda530`                                        |
| T24 CI expansions                                  | `fuzz.yml` (daily 60s ×4 targets, crash artifacts), `codeql.yml` (SHA-pinned v3), `renovate.json` (custom manager for the embedded JS); erraudit transition probe judged redundant (already in ci.yml) and documented                            | `1a72616`                                                   |
| T25 community polish                               | `example/README.md` (heartbeat docs with proxy-timeout rationale), `Dockerfile` (base image fixed to `golang:…`), `docker-compose.yml`, `.github/FUNDING.yml`                                                                                    | `cf7b3f4`                                                   |
| T26 formatter ADR                                  | ADR 006 (treefmt canonical; dprint.json as non-Go intent; hermeticity rationale)                                                                                                                                                                 | `cf19bf1`                                                   |
| Final verification sweep                           | full local gate green (see Verdict at a glance) + all 5 CI workflows success on master                                                                                                                                                           | run 33266637847 et al.                                      |
| Living-doc sync                                    | TODO_LIST rebuilt (open set now: watch items + owner-blocked rows); CHANGELOG `[Unreleased]` consolidated; ROADMAP policy note → 1.26.7; `docs/status/README.md` policy created                                                                  | `1dda530`                                                   |

## b) PARTIALLY DONE

| Item                                       | Done                                                                                                                                           | Missing                                                                                                                                                                                                                        |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| T11 hermetic checks + vendorHash hardening | `apps.bench`, `datastartest/collect_bench_test.go`, ADR 004 with the lint/vet/govulncheck non-sandboxing verdict; full `nix flake check` green | Sandbox `checks.lint/vet/govulncheck` deliberately NOT built (verdict recorded); the vendorHash _sensitivity mechanism_ is likely misdiagnosed in ADR 004 (see improvements) — needs the investigation, then an ADR correction |
| T14 ADR 004                                | written and linked                                                                                                                             | Must be revised once the FOD-sensitivity question is answered (today's evidence contradicts "only Go patch bumps move it")                                                                                                     |
| T21 headless E2E                           | driver decision (chromedp), scope, and deferral rationale recorded in ROADMAP                                                                  | No spike was run, no build-tag scaffold exists — by the plan's letter this is a deferral, not execution                                                                                                                        |
| T27 website spike                          | deferral + trigger conditions ("next release worth a launch") recorded in ROADMAP                                                              | No Astro/Starlight spike, no demo-video decision artifact                                                                                                                                                                      |
| 16.3 guide cross-linking                   | guides internal-consistent, doc-indexed nowhere                                                                                                | README/AGENTS do not link `docs/replay.md` etc. yet                                                                                                                                                                            |
| Coverage prose                             | testing.md/performance.md state numbers with date                                                                                              | Numbers are the morning's; re-measure now that new code landed (badge is live-current; prose lags)                                                                                                                             |
| 12.3 lint caching measurement              | cache wired, rationale documented                                                                                                              | No before/after timing was captured                                                                                                                                                                                            |
| 05.3 docs-only fast-path verification      | configured                                                                                                                                     | Never observed an actual docs-only push skipping CI                                                                                                                                                                            |
| T25 "contributor list"                     | FUNDING.yml added                                                                                                                              | No AUTHORS/contributor file (deliberate: GitHub renders it); counted as a choice, noted here for completeness                                                                                                                  |

## c) NOT STARTED

- T04 branch hygiene — **blocked by design (G7)**: deleting
  `pr/docs-test-consolidation` and rehoming/dropping
  `preserve/status-report-coderabbit-pr3` need owner approval; both branches
  untouched. (Question 1 below.)
- Sandbox `checks.lint` / `checks.vet` / `checks.govulncheck` — consciously
  not built; verdict + revisit-condition in ADR 004.
- Cachix or binary-cache acceleration for the nix CI job — only if the job's
  runtime (28–32s observed, surprisingly fast) ever becomes a problem.
- AUTHORS/contributor file, community health files beyond FUNDING.
- Website/Astro anything (T27) — deferral recorded.
- chromedp/browser test scaffold (T21) — deferral recorded.

## d) TOTALLY FUCKED UP

Severity-ordered; nothing data-destroying, all recoverable, two already
mitigated on master.

1. **`nix flake check` fails at tag `v0.3.0`.** The release commit carried
   the requires bump but the vendorHash harvested _before_ it
   (`nkJghg…`); the FOD at that tree produces `GciltFE7…`. My in-flight
   verification job was left unread while I tagged. Master is fixed
   (`1f3cd93`) and consumers are unaffected (module proxy ≠ flake), but the
   repo's canonical gate is red at the released tag, and tags are
   irreversible. Mitigation on the table: fix-forward (done), document the
   tag's flake state, do **not** re-tag. The embarrassing part is not the
   hash — it is tagging while my own verification was still running.
2. **Red master from a linter build I didn't reproduce** (run
   33266437052). Local lint green, CI lint red (`varnamelen` on `s`),
   fixed one commit later (`8cc56a7`). Violates the _spirit_ of G8 even
   with local "green" in hand: parity of toolchain build matters as much as
   presence of the gate.
3. **Master's CHANGELOG references an uncommitted file.** My `1dda530`
   staged the whole CHANGELOG, including the parallel session's
   `wire_golden_test.go` bullet — the file is still untracked in the shared
   checkout. On master, the changelog advertises a test that isn't in the
   repo. Caused by my staging scope during someone else's in-flight work.
4. **I reintroduced the exact dependency this repo banned.** The
   domain-adapter test importing `datastartest` would have grown the
   forbidden root→datastartest edge on the next `go mod tidy`. Caught by
   the isolation run at session end, not by my own review of the module
   rules I had just re-parsed while pruning AGENTS.md.
5. **Two dependency bots configured against each other** (renovate.json
   added; dependabot.yml pre-existing) — guaranteed duplicate PR churn until
   one is switched off. Unambiguous oversight; needs a one-line decision
   (Question 2).

## e) WHAT WE SHOULD IMPROVE

1. **Gate on the CI linter build.** Add a flake app / devShell alias that
   runs golangci-lint at exactly `v2.12.2` via `go run`, and make that the
   pre-push habit. Kills the entire "green locally, red remotely" class.
2. **Investigate the vendorHash sensitivity properly, then correct ADR
   004.** Today's evidence (hash changed across tree states with identical
   go.mod/go.sum) suggests the local `replace` directives pull repo source
   into the module FOD under `proxyVendor`. If confirmed: vendorHash moves
   on _any_ root/static/datastartest source change — the "broke twice under
   the daemon" incidents and today's break share one root cause, and the
   durable fix is restructuring the FOD inputs (e.g., minimal src fileset of
   go.mod/go.sum only), not hash dances.
3. **Drain-or-kill background jobs before advancing.** Two-hour-blind spots
   come from unread `nix flake check` runs.
4. **Owner questions asked at the start of execution**, not discovered in
   section (g) of a report. T04 sat blocked all session by choice of
   ordering.
5. **Sync living docs in the same commit as the change they describe** —
   AGENTS' CI section is already behind (missing nix/fuzz/codeql/renovate),
   FEATURES missed today's features, README misses the guides. The
   docs-health discipline exists; it wasn't applied per-commit.
6. **Doc snippets should be machine-checked.** Seven guides hand-verified
   at best; a tiny `docs/examples` compile target (or `//go:build doc` test
   file) would keep `BroadcastMany`-class drift impossible.
7. **Stage atomic, coordinate with the daemon**: commit the parallel
   session's CHANGELOG bullet only together with its file, or not at all.
8. **Dependency-bot policy**: exactly one of Renovate/Dependabot.
9. **Verify container builds** if a Dockerfile ships — a Dockerfile that
   was never built is documentation, not packaging.
10. **Re-measure coverage after code changes** before printing numbers into
    docs (the badge does this right; prose should follow the badge).

## f) Up to 50 things to do next (prioritized)

| #      | Thing                                                                                                                                                                                                                      | Why / notes                          |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| 1      | Add exact-CI-version golangci-lint app (`go run …@v2.12.2`) to flake + use pre-push                                                                                                                                        | kills linter-parity reds             |
| 2      | Verify tag `v0.3.0` flake state via a `git worktree` of the tag + `nix flake check`; record verdict                                                                                                                        | confirms (d)1 empirically            |
| 3      | Investigate datastartest FOD source-sensitivity; correct or confirm ADR 004 mechanism                                                                                                                                      | stops repeat vendorHash breaks       |
| 4      | Decide Renovate vs Dependabot; disable the loser                                                                                                                                                                           | duplicate-bot churn (Q2)             |
| ~~5~~  | ~~Commit or strip the `wire_golden_test.go` CHANGELOG bullet~~ done — wire_golden_test.go committed (a0c0aea); CHANGELOG bullet now accurate                                                                               | ~~master consistency (Q3)~~          |
| ~~6~~  | ~~Run `erraudit` loop locally over the new code (`script_accessors.go` etc.)~~ done — erraudit loop clean on all three modules (docs-health 2026-09-02)                                                                    | ~~audit gap from today~~             |
| ~~7~~  | ~~Link the 7 guides from README (docs section) + AGENTS docs map~~ done — README Documentation section + AGENTS Docs Map link the seven guides (2026-09-02)                                                                | ~~plan 16.3 leftover~~               |
| ~~8~~  | ~~Update FEATURES.md: typed accessors, domain-adapter, guides, new CI, compression middleware~~ done — FEATURES.md updated 2026-09-02 (accessors, domain-adapter, guides, CI, compression, ReplaceURLQuerystring, goldens) | ~~FEATURES drifted today~~           |
| ~~9~~  | ~~Sync AGENTS CI section: nix.yml, fuzz.yml, codeql.yml, renovate.json~~ done — AGENTS.md CI section synced (nix.yml, fuzz.yml, codeql.yml, renovate.json, SHA pinning)                                                    | ~~AGENTS stale same-day~~            |
| ~~10~~ | ~~Re-measure coverage (root/datastartest) post-changes; correct prose numbers~~ done — verified 2026-09-02 — 98.4% root / 92.7% datastartest re-measured; prose numbers stand                                              | ~~badge vs prose drift~~             |
| 11     | `docker build example/` (and compose up) to verify packaging actually builds                                                                                                                                               | unverified Dockerfile                |
| 12     | Watch/promote `nix.yml`: after a week green, drop `continue-on-error`                                                                                                                                                      | TODO_LIST watch item                 |
| 13     | Observe one docs-only push skipping CI (and actionlint still running)                                                                                                                                                      | 05.3 leftover                        |
| 14     | Verify Renovate regex against real upstream tags on first run                                                                                                                                                              | TODO_LIST item                       |
| 15     | Watch `fuzz.yml` nightly run; confirm artifacts upload path correctness                                                                                                                                                    | new workflow                         |
| 16     | Check CodeQL first analysis results for surprises                                                                                                                                                                          | new workflow                         |
| ~~17~~ | ~~Record CHANGELOG `[Unreleased]` Fixed: tag-flake vendorHash note + varnamelen fix~~ done — CHANGELOG [Unreleased] 'Fixed' records the tag-flake vendorHash note + varnamelen fix (2026-09-02)                            | ~~honest release history~~           |
| 18     | Owner: branch deletions + lineage prune (`git config` cleanup afterwards)                                                                                                                                                  | T04, blocked (Q1)                    |
| 19     | Owner: rehome 11-37 report from `preserve/…` as a PR (or drop)                                                                                                                                                             | T04 companion                        |
| 20     | Configure `git-town.observed-branches` for `preserve/…` while deletion is blocked                                                                                                                                          | stops town aborts                    |
| 21     | Fix `docs/performance.md` numbers with a default-benchtime re-run for stable stats                                                                                                                                         | 100x sample was small                |
| 22     | Add compile-checked snippets for guides (small `internal/docspec` test)                                                                                                                                                    | doc drift class                      |
| ~~23~~ | ~~Verify `static/v0.3.0` and `datastartest/v0.3.0` pkg.go.dev pages render~~ done — verified 2026-09-02 — static/v0.3.0 and datastartest/v0.3.0 render on pkg.go.dev (Latest)                                              | ~~only root verified visually~~      |
| ~~24~~ | ~~Re-read the v0.3.0 GitHub Release page formatting~~ done — gh release view v0.3.0 shows the 8.9KB CHANGELOG-excerpt body rendered                                                                                        | ~~created blind from excerpt~~       |
| 25     | Decide on `datastartest/v0.3.0` consumer announcement (CHANGELOG excerpt tweet/RSS?) — optional                                                                                                                            | visibility                           |
| 26     | Add `docs/replay.md` backlog/live duplicate-window example test (dedupe pattern)                                                                                                                                           | guide honesty upgrade                |
| ~~27~~ | ~~Extend `wire-format.md` with the golden-test file reference once `wire_golden_test.go` lands~~ done — docs/wire-format.md 'Where conformance is enforced' cites wire_golden_test.go                                      | ~~pending (Q3)~~                     |
| ~~28~~ | ~~Consider `AnimatedEventsString`-style debug helpers… (or) prune: not needed~~ **Won't implement — skip-worthy by its own text — debug-helper churn without a consumer ask.**                                             | ~~skip-worthy, listed to resist it~~ |
| 29     | Add Renovate `:label`/schedule tuning after first run                                                                                                                                                                      | noise control                        |
| 30     | Confirm dependabot/renovate paths don't fight over `go.mod` (lockfile overlap)                                                                                                                                             | part of #4                           |
| 31     | Add `checks.govulncheck` sandbox derivation IF the ADR-004 investigation shows it's cheap                                                                                                                                  | revisit trigger from ADR             |
| 32     | Benchmark CI: record lint job duration pre/post cache (close 12.3)                                                                                                                                                         | measurement debt                     |
| ~~33~~ | ~~Add `example/domain-adapter` to the example README table of ports/routes~~ **Won't implement — the example README describes domain-adapter in prose; a ports/routes table does not exist to extend.**                    | ~~small doc polish~~                 |
| 34     | Heartbeat interval in `example/main.go`: make the 2s producer ticker a named constant too (mnd ignored-list covers `2` today; fragile)                                                                                     | latent mnd regression                |
| 35     | `gzipSSEWriter`: consider `http.NewResponseController` for flush-depth correctness on Go 1.2x                                                                                                                              | correctness hardening                |
| 36     | Add a test asserting the root module go.mod never gains `datastartest` after a `go mod tidy` (extend module_boundary_test to run tidy in check mode)                                                                       | near-miss today                      |
| 37     | Ensure `renovate.json` `:disableDependencyDashboard` is the owner's preference                                                                                                                                             | config taste                         |
| ~~38~~ | ~~Post-release: watch `go get .../datastartest@latest` resolves (not just @v0.3.0)~~ done — v0.3.0 is the latest tag; proxy-verified at release                                                                            | ~~proxy latest-tagging~~             |
| ~~39~~ | ~~Re-run `docs-health VERIFY` on the seven guides after the next code change~~ done — docs-health VERIFY re-ran over the guides 2026-09-02                                                                                 | ~~keep guides honest~~               |
| ~~40~~ | ~~Add the report row + archive ritual to `docs/status/README.md` (done for this report; keep it per-report)~~ done — the report row + archive ritual is codified in docs/status/README.md                                  | ~~policy adherence~~                 |
| 41     | Plan v0.4.0 scope: batch the post-0.3.0 `[Unreleased]` items (accessors, ReplaceURLQuerystring, guides-linked fixes)                                                                                                       | release cadence                      |
| 42     | Evaluate `actions/upload-artifact` v5 / current major when touching fuzz.yml next                                                                                                                                          | pin freshness                        |
| 43     | Consider adding `constraintlint`-style check that `static.Version` matches a CHANGELOG mention                                                                                                                             | pinning hygiene                      |
| 44     | Review `nix.yml` 32s runtime — confirm it actually built (not vacuously skipped)                                                                                                                                           | trust the green                      |
| 45     | Add CI job summary table to AGENTS CI section after promotion decisions settle                                                                                                                                             | doc single-source                    |
| ~~46~~ | ~~Prune `docs/planning/2026-08-29_18-32_*.md` task statuses (mark executed) or archive it per policy~~ done — the plan's task rows are struck with commit hashes + a resolution appendix (2026-09-02)                      | ~~plan lifecycle~~                   |
| 47     | Make `nix run .#bench` output stable headers for future comparisons                                                                                                                                                        | measurement hygiene                  |
| 48     | Sweep TODO_LIST watch items after first CI week; close or convert                                                                                                                                                          | keep list honest                     |
| 49     | Decide favicon/branding assets if website deferral flips to go                                                                                                                                                             | T27 trigger prep                     |
| 50     | Post-owner-answers: execute T04 + erraudit-flip watch + close blocked rows                                                                                                                                                 | unblocks the list                    |

## g) Questions for the owner (that I cannot answer myself)

1. **Branches (T04):** Do I now have approval to delete `pr/docs-test-consolidation`
   (local + remote, merged via PR #3) — and for `preserve/status-report-coderabbit-pr3`,
   do you want the 11-37 report rehomed to master as a PR first, or the branch dropped outright?

2. **Dependency bots:** The repo now has both Dependabot (pre-existing) and
   Renovate (added today for the embedded JS client). Which one should own
   dependency updates — keep Renovate and remove `dependabot.yml`, or keep
   Dependabot and drop `renovate.json`?

3. **Parallel-session file (`wire_golden_test.go`):** It is still untracked,
   but my pushed CHANGELOG already describes it. Should I commit that file
   as-is (it passes the full gate when I last tested the tree), strip its
   CHANGELOG bullet until it lands, or leave it for the other session to
   finish and commit?

---

_Point-in-time report — see [`docs/status/README.md`](README.md) for the
index and archiving policy._
