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

| Task                                                                                                                                                                                                                                                                                                                                           | Status       | Impact | Effort | Evidence                                                       |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------- |
| Verify the Renovate GitHub App is actually installed (owner: GitHub → Settings → Applications) — zero renovate PRs ever and the installations API 403s with a user token; if absent, the regex-scoped `renovate.json` is dead config and JS-bump proposals never arrive.                                                                       | 🔵 `BLOCKED` | Low    | 5min   | `gh pr list --author renovate[bot]` = 0; 2026-10-01 report §b2 |
| Cut v0.7.0 (owner): the headliner tranche (broadcast options / `Store` seam / heartbeat) plus the invalid-UTF-8 error-response fix, docspec completion, and submodule LICENSE files are complete in `[Unreleased]`. Ride-along decision: remove `ScriptHandlerWith`'s dead version parameter (breaking, owner-gated, documented in its godoc). | 🔵 `BLOCKED` | Medium | 30min  | CHANGELOG [Unreleased]; docs/release-checklist.md §2.5 gate    |
| BuildFlow upstream candidate: align its on-demand go-licenses install env with its run env (the GOROOT mismatch this repo works around via the devshell contract in `.buildflow.yml`). Fleet-level fix — belongs in the BuildFlow repo, not here.                                                                                              | 🔴 `TODO`    | Low    | 15min  | 2026-10-01 license-check diagnosis                             |

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
