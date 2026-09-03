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

| Task                                                                                                                                                                                                                                                                                                   | Status    | Impact | Effort | Evidence                                                                           |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------- | ------ | ------ | ---------------------------------------------------------------------------------- |
| CI watch ritual (monthly or per workflow change): follow [docs/ci-watch.md](docs/ci-watch.md) — promote `nix.yml` (first green 2026-09-03; two green weeks → drop `continue-on-error`), triage first `fuzz.yml` artifacts, review CodeQL alerts, verify Renovate proposals against real upstream tags. | 🔴 `TODO` | Medium | 30min  | `docs/ci-watch.md`; nix.yml first green on the v0.4.0 fix commit                   |
| datastartest helper-API expansion, next tranche (from the consolidated ROADMAP idea): RequireNotScript, FindAllElements, FindScript, EventToSelectorMap, timeout variants of CollectWithRequest/CollectPost.                                                                                           | 🔴 `TODO` | Low    | 2h     | ROADMAP theme 2 (source-cited); first three helpers shipped 2026-09-03 (`b87374e`) |
| v0.5.0 scoping: the `[Unreleased]` section (datastartest top-3 helpers, CI matrix split, static hardening, Response method forms, version pkg, docspec, community files) is release-ready once the CI-watch ritual confirms a green week.                                                              | 🔴 `TODO` | Medium | 30min  | CHANGELOG `[Unreleased]`; release checklist                                        |

## Owner-blocked

| Task                                                                                                                                                       | Status       | Impact | Effort | Evidence                                               |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------ |
| Dependency bots: keep Renovate or Dependabot, disable the other (both active today — duplicate PR churn); also set Renovate dashboard/label preferences.   | 🔵 `BLOCKED` | Medium | 15min  | `.github/dependabot.yml` + `renovate.json` coexist     |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                          | 🔵 `BLOCKED` | Medium | 5min   | PR #3 state: MERGED                                    |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.           | 🔵 `BLOCKED` | Medium | 15min  | branch exists; observed-branches configured 2026-09-03 |
| CODEOWNERS: create with named owners (deliberately NOT created 2026-09-03 — naming is the owner's call; SUPPORT.md + discussion templates landed instead). | 🔵 `BLOCKED` | Low    | 10min  | SUPPORT.md exists; no CODEOWNERS by design             |
| erraudit CI job flips to hard gate when the repo goes public — verify the probe notices the flip on the first push after publication.                      | 🔵 `BLOCKED` | Low    | 5min   | `ci.yml` probe job; AGENTS.md Nix gotchas              |
| Website launch (Astro + Starlight pattern) — deferred by owner decision (T27-of-18-32).                                                                    | 🔵 `BLOCKED` | Low    | —      | ROADMAP theme 4                                        |
| Status-index "Monitoring" tier for never-fully-resolvable reports; per-item marker depth for consolidated ROADMAP ideas.                                   | 🔵 `BLOCKED` | Low    | 15min  | 2026-09-02 report owner questions                      |
| AGENTS.md settle point: currently 15.0KB after the 2026-09-03 prune (target ≤15KB met); prune again only if it regrows past 15KB.                          | 🔵 `BLOCKED` | Low    | 45min  | `wc -c AGENTS.md`                                      |

## Notes

- Rebuilt 2026-09-03 (pareto execution session, T01–T27 of
  `docs/planning/2026-09-02_23-59_pareto-ship-v0.4.0-kill-recurring-breaks-depth-surface.md`):
  all 27 tasks executed and verified. Shipped: v0.4.0 (lockstep tags, GitHub
  Release, proxy + pkg.go.dev verified), exact-CI lint app (`nix run .#lint-ci`),
  LSP-warning surface cleared (b.Loop; json/v2 stdversion documented as gopls
  false positive), FOD/vendorHash mechanism verified + ADR 004 corrected with
  the never-converging self-reference discovery + minimal datastartest fileset,
  tag-v0.3.0 flake verdict recorded, CI hygiene batch (mod verify, use-vs-disk,
  tidy-diff, JS-version drift test), per-module CI matrix, Docker fix + live
  SSE smoke, performance re-measure (1s benchtime), micro-hygiene batch,
  measurement ritual, AGENTS prune to 15.0KB, modfile boundary guard,
  datastartest top-3 helpers (93.4% coverage), hygiene pack, CI watch runbook,
  Response test-depth batch (+ new `error_response_nil_error` code), CodeRabbit
  thread closure (5/5 answered, 1 fix landed), community files, static
  hardening (nosniff, checksum pin, provenance, deprecation), datastartest
  function examples, audit pack (README table re-verified vs upstream v1.2.2;
  10% verdict spot-check 5/5), decision pack (ADRs 002/005 addenda + ROADMAP),
  architecture.md, starfederation migration guide + Broadcaster/filter/replay
  examples, docspec compile-checked snippets (caught 3 real doc drifts),
  platform pack (Response method forms, version pkg, goreleaser skeleton).
- Owner-blocked items were never actioned (G11): the four branch/bot/naming
  decisions, erraudit flip, website trigger, and the 2026-09-02 report's owner
  questions remain open.
- Resolved questions are documented in ROADMAP.md "Resolved questions" and
  AGENTS.md.
