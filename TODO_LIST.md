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

## Open items

| Task                                                                                                                                                                        | Status       | Impact   | Effort | Evidence                                                                                         |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | -------- | ------ | ------------------------------------------------------------------------------------------------ |
| Watch the first runs of the new CI workflows (`nix.yml` non-required, `fuzz.yml` nightly, `codeql.yml`) and promote/remove based on stability.                               | 🔴 `TODO`    | Medium   | 15min  | `nix.yml` (continue-on-error until proven); 2026-08-29 execution session                          |
| Renovate first-run check: confirm the custom-manager regex maps `static/static.go` Version to upstream tags correctly; adjust `extractVersionTemplate` if tags differ.       | 🔴 `TODO`    | Low      | 10min  | `renovate.json` (tag scheme written v-prefix-tolerant, unverified against upstream releases)      |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                                           | 🔵 `BLOCKED` | Medium   | 5min   | `git branch -a` (2026-08-29); 12-43 report f-table item 4                                         |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.                            | 🔵 `BLOCKED` | Medium   | 15min  | `git branch -a` (2026-08-29); 12-24 report c.2                                                    |
| erraudit CI job flips to hard gate when the repo goes public — no action needed until then; verify the probe notices the flip on the first push after publication.           | 🔵 `BLOCKED` | Low      | 5min   | `ci.yml` probe job; AGENTS.md Nix gotchas                                                         |

## Notes

- Rebuilt 2026-08-29 (Full Execution Mode): the 2026-08-29 pareto plan
  (T01–T27) executed — toolchain settled at 1.26.7, master lint debt cleared,
  v0.3.0 shipped and proxy-verified, docs-only CI paths, status index,
  PR-template honesty guard, fuzz docs, AGENTS pruning (28.7KB → ~17KB),
  nix CI + bench, ADRs 003–006, consumer guide set, ReplaceURLQuerystring,
  compression middleware, typed datastartest accessors, domain-adapter
  example, CI expansions (fuzz/CodeQL/Renovate), community polish. Blocked
  rows remain: branch deletions and the erraudit flip.
- Harvested 2026-08-29 (docs-health) from `docs/status/2026-08-16_12-43_*`,
  `2026-08-16_12-24_*`, and `2026-08-16_11-07_*` — every open item verified
  against the repo before routing; resolved items were struck in the reports.
- All 19 items from the 2026-08-16 pareto plan (T01–T16 + T04) are complete.
  See `CHANGELOG.md` [Unreleased] section for details.
- Resolved questions (go.work.sum tracking, v0.0.0 sibling requires, go
  directive policy) are documented in `ROADMAP.md` "Resolved questions" and
  `AGENTS.md`.
