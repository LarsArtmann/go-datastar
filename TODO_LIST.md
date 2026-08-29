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

| Task                                                                                                                                                                                                                                        | Status    | Impact   | Effort | Evidence                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | -------- | ------ | ------------------------------------------------------------------------------------------------- |
| Finish or revert the in-flight **go 1.26.7 toolchain bump**: root go.mod says 1.26.7 while go.work, datastartest, static, CI, and the flake pin still say 1.26.6 — workspace `go` commands fail and a half-bump breaks the hermetic build. Complete the G2 set (go.mod ×3 + go.work + ci.yml + flake `overrideAttrs`) or restore 1.26.6. Coordinate with the session that started it. | 🟡 `IN_PROGRESS` | Critical | 30min  | `go.mod:3`; `go work use` error (2026-08-29); plan guard G2 in `docs/planning/archived/2026-08-16_08-53_*` |
| Restore master CI green: commit the pending gofumpt reformat of `datastartest/reader_fuzz_test.go` (already in the working tree; also fixes the `nix flake check` format gate), fix `example/main.go:153` (mnd), `datastartest/reader.go:98` (makezero), and migrate `errors.As` → `errors.AsType` at root `errors_test.go:238,289` + `datastartest/errors_test.go:34,120` (modernize) | 🔴 `TODO` | Critical | 45min  | CI run 33258092947 (lint red on `d032dc5`); `golangci-lint run` 2026-08-29; 12-43 report f-table item 3 |
| Cut **v0.3.0**: [Unreleased] carries the 1.26.6 CVE fixes, hermetic per-module checks, WPT-conformant SSE parser, datastartest request options, coverage badge. Tag all 3 modules in lockstep; verify pkg.go.dev after.                                                                                       | 🔴 `TODO` | High     | 1h     | `CHANGELOG.md` [Unreleased]; 11-07 report F1; `docs/release-checklist.md`                          |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                                                                                                            | 🔵 `BLOCKED` | Medium | 5min   | `git branch -a` (2026-08-29); 12-43 report f-table item 4                                          |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.                                                                                             | 🔵 `BLOCKED` | Medium | 15min  | `git branch -a` (2026-08-29); 12-24 report c.2                                                    |
| CI path filters so docs-only PRs skip test/lint/govulncheck (~2min saved per docs change).                                                                                                                                                   | 🔴 `TODO` | Medium   | 15min  | `.github/workflows/ci.yml`; 12-43 report f-table item 12                                           |
| `docs/status/README.md` index: date + one-liner + outcome per report (30 snapshots, no navigation).                                                                                                                                          | 🔴 `TODO` | Medium   | 30min  | `docs/status/` (30 files); 12-43 report f-table item 15                                            |
| golangci-lint CI caching (1m33s long pole on every push).                                                                                                                                                                                     | 🔴 `TODO` | Low      | 30min  | CI run 33258092947; 12-43 report f-table item 13                                                   |
| PR template: mark CI-dependent checklist boxes as "checked by CI", not pre-checked by agents.                                                                                                                                                 | 🔴 `TODO` | Low      | 10min  | `.github/PULL_REQUEST_TEMPLATE.md`; 12-43 report f-table item 14                                   |
| CONTRIBUTING.md: how to run the fuzz tests (`go test -fuzz=FuzzReadEvents -fuzztime=30s`, per-module).                                                                                                                                        | 🔴 `TODO` | Low      | 15min  | `CONTRIBUTING.md` (no fuzz section); 11-07 report F11                                              |

## Notes

- Harvested 2026-08-29 (docs-health) from `docs/status/2026-08-16_12-43_*`,
  `2026-08-16_12-24_*`, and `2026-08-16_11-07_*` — every open item verified
  against the repo before routing; resolved items were struck in the reports.
- All 19 items from the 2026-08-16 pareto plan (T01–T16 + T04) are complete.
  See `CHANGELOG.md` [Unreleased] section for details.
- Resolved questions (go.work.sum tracking, v0.0.0 sibling requires, go
  directive policy) are documented in `ROADMAP.md` "Resolved questions" and
  `AGENTS.md`.
