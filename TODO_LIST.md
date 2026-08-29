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

| Task                                                                                                                                                                       | Status    | Impact | Effort | Evidence                   |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | -------------------------- |
| De-flake remaining lint issues in parallel-session files (`reader.go` intrange, `wpt_format_corpus_test.go` dupword, `reader.go` gochecknoglobals/nlreturn/nonamedreturns) | 🔴 `TODO` | Low    | 10min  | `golangci-lint run` output |

## Notes

- All 19 items from the 2026-08-16 pareto plan (T01–T16 + T04) are complete.
  See `CHANGELOG.md` [Unreleased] section for details.
- Resolved questions (go.work.sum tracking, v0.0.0 sibling requires, go directive
  policy) are documented in `ROADMAP.md` "Resolved questions" and `AGENTS.md`.
