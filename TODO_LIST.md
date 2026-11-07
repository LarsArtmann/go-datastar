# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.
> Completed items are removed and logged in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |

## Verified next-up

| Task                                                                                                                                                                                                                                                                                                                                                                                                                               | Status       | Impact | Effort | Evidence                                                       |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------- |
| Verify the Renovate GitHub App is actually installed (owner: GitHub → Settings → Applications) — zero renovate PRs ever and the installations API 403s with a user token; if absent, the regex-scoped `renovate.json` is dead config and JS-bump proposals never arrive.                                                                                                                                                           | 🔵 `BLOCKED` | Low    | 5min   | `gh pr list --author renovate[bot]` = 0; 2026-10-01 report §b2 |
| Cut v0.7.0 (owner): the headliner tranche (broadcast options / `Store` seam / heartbeat) plus the invalid-UTF-8 error-response fix, docspec completion, and submodule LICENSE files are complete in `[Unreleased]`. Ride-along decision: remove `ScriptHandlerWith`'s dead version parameter (breaking, owner-gated, documented in its godoc).                                                                                     | 🔵 `BLOCKED` | Medium | 30min  | CHANGELOG [Unreleased]; docs/release-checklist.md §2.5 gate    |
| BuildFlow upstream candidate: align its on-demand go-licenses install env with its run env (the GOROOT mismatch this repo works around via the devshell contract in `.buildflow.yml`). Fleet-level fix — belongs in the BuildFlow repo, not here.                                                                                                                                                                                  | 🔴 `TODO`    | Low    | 15min  | 2026-10-01 license-check diagnosis                             |
| Release-checklist hardening trio (2026-09-03 superb audit): §5 upstream-release check (`gh api …/releases/latest` vs `static.Version`), §4 tri-tag `gh release create` with `--latest=false`, and a `gh release list` completeness assertion — at the v0.7.0 cut.                                                                                                                                                                  | 🔴 `TODO`    | Low    | 15min  | docs/release-checklist.md; superb-audit report §f              |
| Extend the JS-version CHANGELOG contract beyond the static bundle: `TestStaticVersionRecordedInChangelog` pins `static.Version` only; the same "a version bump can never land invisibly" contract should cover the go-sse/ssetest pins datastartest consumes (assert the CHANGELOG records the pin a release ships). Routed from the 2026-10-11 go-sse hygiene batch.                                                              | 🔴 `TODO`    | Low    | 30min  | `version_constraint_test.go`; go-sse 2026-10-11 batch          |
| Broadcast `Store` contract test helper: a reusable assertion helper so consumer `Store` implementations can prove the monotonic-ID contract that `broadcast/options.go` documents (a store emitting non-monotonic IDs degenerates Last-Event-ID replay). Low impact, Medium effort. Routed from the 2026-10-01 class-fix report §f10.                                                                                              | 🔴 `TODO`    | Low    | Medium | `broadcast/options.go`; 2026-10-01 class-fix report §f10       |
| Upstream release check + README comparison table: diff the pinned static bundle against the latest starfederation/datastar-go release and seed the README "vs upstream SDK" comparison table (ROADMAP theme 2 idea). Low impact, 15min. Routed from the 2026-10-01 one-bot report §f13.                                                                                                                                            | 🔴 `TODO`    | Low    | 15min  | `static/static.go` pin; 2026-10-01 one-bot report §f13         |
| Owner check: private vulnerability reporting cannot be verified via API (`private_vulnerability_reporting_enabled` absent; `/vulnerability-reports` 404s) — confirm Settings → Code security → Private vulnerability reporting is ON. Also: open dependabot PR #23 (codeql-action bump) fails CodeQL's own `analyze` check (not the nix-hash class) — close-and-let-renovate/dependabot-recreate or investigate CodeQL separately. | 🔵 `BLOCKED` | Low    | 5min   | API probe 2026-10-11; `gh pr view 23`                          |
| ScriptTagVersioned follow-up tranche A (docs): CSP + versioned-URL note in `docs/static-js.md`; `doc.go` mention next to ScriptHandler; AGENTS.md one-liner (versioned URLs are the recommended serving pattern); README quick-start pairing `mux.Handle` with `ScriptTagVersioned`; FEATURES.md row + docs-map reconcile; `appendVersionQuery` malformed-URL edge doc comment; datastartest README tour (after tranche B's helper ships); `docs/testing.md` versioned-tag mention; record the fuzz non-decision in `docs/static-js.md`; verify no cross-module doc claims `ScriptTagVersioned` for `static/`; upstream docs cache-busting research. Harvested from the 2026-11-07 sweep report §f. | 🔴 `TODO` | Low | Medium | 2026-11-07 sweep report §f17/19/21/27/29-31/34-36/41 |
| ScriptTagVersioned follow-up tranche B (code+test): datastartest `RequireScriptTagVersion` helper (testing.TB invariant); `appendVersionQuery` idempotency for URLs already carrying `v=` (it double-appends today); example smoke test asserting `src="/datastar.js?v=<static.Version>"`; godoc `ExampleScriptTagVersioned`; `nix run .#test` vs raw go test parity check; confirm no API-surface golden test trips on future exports; re-measure root coverage post-`ScriptTagVersioned`; verify broadcast docspec CI wiring. Harvested from the 2026-11-07 sweep report §f. | 🔴 `TODO` | Medium | Medium | `script_handler.go:101`; 2026-11-07 sweep report §f16/22/24/28/32/38/39/42 |
| Archive-ledger residue passes: real per-row verdicts for the broadcast-review test-matrix (~40 ledgered rows; many likely shipped), verify the 09-02 pointer-ledger rows (45) still resolve post-archive, disposition the buildflow brainstorm-ledger rows (38; several belong upstream in BuildFlow). Harvested from the 2026-11-07 sweep report §f. | 🔴 `TODO` | Low | Medium | `docs/status/archived/2026-09-18_20-45*`; sweep report §f4-6 |
| VendorHash discrepancy investigation: `git show b713f67 -- flake.nix` (the parallel session's unexplained hash move, green-but-unverified). | 🔴 `TODO` | Low | 15min | sweep report §f15; commit b713f67 |
| Watch the first Renovate JS bump: confirm the `?v=` query updates automatically via `static.Version` when the custom manager proposes a bundle bump. | 🔴 `TODO` | Low | 5min | sweep report §f37; `renovate.json` regex manager |

## Owner-blocked

| Task                                                                                                                                                                                                    | Status       | Impact | Effort | Evidence                                               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------ |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                                                                       | 🔵 `BLOCKED` | Medium | 5min   | PR #3 state: MERGED                                    |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.                                                        | 🔵 `BLOCKED` | Medium | 15min  | branch exists; observed-branches configured 2026-09-03 |
| CODEOWNERS: create with named owners (deliberately NOT created 2026-09-03 — naming is the owner's call; SUPPORT.md + discussion templates landed instead).                                              | 🔵 `BLOCKED` | Low    | 10min  | SUPPORT.md exists; no CODEOWNERS by design             |
| erraudit CI job flips to hard gate when the repo goes public — verify the probe notices the flip on the first push after publication.                                                                   | 🔵 `BLOCKED` | Low    | 5min   | `ci.yml` probe job; AGENTS.md Nix gotchas              |
| Website launch (Astro + Starlight pattern) — deferred by owner decision (T27-of-18-32).                                                                                                                 | 🔵 `BLOCKED` | Low    | —      | ROADMAP theme 4                                        |
| Status-index "Monitoring" tier for never-fully-resolvable reports; per-item marker depth for consolidated ROADMAP ideas.                                                                                | 🔵 `BLOCKED` | Low    | 15min  | 2026-09-02 report owner questions                      |
| Shared lint-cache policy: purge the 3.8G `/mnt/buildcache/golangci-lint` or bless `GOLANGCI_LINT_CACHE=$(mktemp -d)` as the documented default for gates.                                               | 🔵 `BLOCKED` | Low    | 5min   | `2026-09-18_22-29` g1; AGENTS.md cache-ghosting gotcha |
| Required-checks policy: master has none — document the "local gates are the gate" contract in one canonical place, or make lint/nix alerting.                                                           | 🔵 `BLOCKED` | Low    | 15min  | `2026-09-18_20-45` g3, f26                             |
| v0.6.1 retro status report: the 2026-09-29 session (toolchain 1.27.1, go-sse v0.6.1, replace drops, lint-pin repair) shipped without one — write retroactively from CHANGELOG + git evidence, or waive. | 🔵 `BLOCKED` | Low    | 30min  | `2026-10-01_04-37` report g2; tags exist, no report    |
| No-strike-archive canon decision: per-item verdict pass for already-archived idea lists vs a fleet `routed-remainder` marker convention — decide once, not per-archive. | 🔵 `BLOCKED` | Low | 15min | 2026-11-07 sweep report §f3 |
| Review/merge emeet-pixyd PR #12 (the ScriptTagVersioned port's origin; the templ regen hunk needs the templ-version check first). | 🔵 `BLOCKED` | Low | 15min | PR #12; 2026-11-07 sweep report §f25 |
| Release lane for `ScriptTagVersioned`: v0.6.3 root re-tag vs next minor — owner call. | 🔵 `BLOCKED` | Low | 5min | 2026-11-07 sweep report §f26 |
| Cache policy: raise ScriptHandler to `public, max-age=31536000, immutable` now that URLs are versioned? Owner-gated (see static-js.md:106). | 🔵 `BLOCKED` | Low | 5min | `docs/static-js.md`; 2026-11-07 sweep report §f18 |

## Notes

- **Harvested 2026-10-01 (pareto plan execution, both concurrent sessions):**
  all 22 next-up rows closed — release-checklist hardening, migration-guide
  truth pass, erraudit ×4 + govulncheck at HEAD (clean), the buildflow
  license-check loop (devshell GOROOT contract), tranche-2 godoc examples +
  coverage 95.5%, static-js CSP docs + fetch-bundle.sh, docspec completion
  (wire-format + migration-guide mirrored; testing.md divergence fixed),
  dependabot PR #14 merged, the broadcast options/Store/heartbeat tranche
  (T07), API hygiene (Version() re-point, version pkg test), fuzz smokes +
  T16.8 closure (which found and fixed the invalid-UTF-8 error-response
  bug), gzip/docs polish, the AGENTS `--keep-going` note, and submodule
  LICENSE files (plus the root LICENSE holder-name typo fix). Evidence:
  CHANGELOG `[Unreleased]`, commits `803972c`…`9a3ece8`, and the archived
  plan `docs/planning/archived/2026-10-01_05-05_*.md`.
- **AGENTS.md settle point (T12, decided 2026-10-01):** the file settles in
  the 15–30KB band (19.2KB today). The prior 18KB prune already removed the
  discretionary content; what remains is load-bearing operational reality
  (shared-checkout protocol, wire-format rules, nix hash traps, gate
  contracts) that concurrent sessions actively consume. The ≤15KB trigger is
  retired — only reconsider if a whole gotcha class becomes obsolete (e.g.,
  the repo goes public).
- Broadcast tranche row unchanged from the 2026-09-18 routing (R1–R3,
  owner-gated for the next minor per that session's scope answer).
- Resolved questions live in ROADMAP.md "Resolved questions"; release-gate
  policy questions (erraudit mandatory? fresh-cache lint default?) are folded
  into rows above pending the owner's call.
