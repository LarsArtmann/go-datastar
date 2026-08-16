# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.
> Completed items are removed and logged in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |

## Open items

| Task                                                                                              | Status         | Impact   | Effort | Evidence                                                            |
| ------------------------------------------------------------------------------------------------- | -------------- | -------- | ------ | ------------------------------------------------------------------- |
| Decide and apply the `go` directive policy: go.mod/go.work say `go 1.26.5` but the v0.0.2/v0.0.3 CHANGELOG entries claim it was lowered to `go 1.26` — the lowering has never landed at any tag | 🔵 `BLOCKED` (user decision) | Medium | 5min | `go.mod:3`, `datastartest/go.mod:3`, `go.work:1`; `git show v0.0.3:go.mod` still says `1.26.5` |
| Add per-module Nix hermetic checks (`hermeticCheckStatic`, `hermeticCheckDatastartest`) — `flake.nix` still carries the TODO; only root is built by `nix flake check` | 🔴 `TODO` | Medium | 60min | `flake.nix:45-47`                                                   |
| Wire `dprint.json` into treefmt/flake.nix (or remove it) — committed config with no integration; formatter today is treefmt (gofumpt/goimports/golines/nixfmt) only | 🔴 `TODO` | Medium | 30min | `dprint.json`; `flake.nix:86-101` has no dprint                     |
| Add `CollectPost` error-path tests: handler returns 400/500 or non-SSE body — current helpers assume success | 🔴 `TODO` | Medium | 30min | No such test in `datastartest/collect_test.go`                      |
| Add `actionlint` step to CI (workflow YAML currently validated only manually)                     | 🔴 `TODO` | Low      | 15min | `.github/workflows/ci.yml` has no actionlint job                   |
| Add `erraudit` to the nix devShell (golangci-lint + govulncheck are there; erraudit only exists as an app) | 🔴 `TODO` | Low      | 5min  | `flake.nix:106-113`                                                 |
| Write ADR 002: multi-module split + mutual-replace pattern (root replaces static; datastartest replaces both) | 🔴 `TODO` | Low      | 30min | `docs/adr/` has only `001-architecture.md`                          |
| Add CONTRIBUTING.md section on multi-module development (workspace vs `GOWORK=off`, replace directives, per-module tagging) | 🔴 `TODO` | Low      | 15min | `CONTRIBUTING.md` never mentions modules or `go.work`              |
| Verify pkg.go.dev renders docs for latest versions (root v0.2.0, static, datastartest)            | 🔴 `TODO` | Low      | 10min | Badge links there; rendering unconfirmed                           |
| Add coverage badge to README                                                                       | 🔴 `TODO` | Low      | 20min | README has CI/GoRef/GoReport/License badges; no coverage badge     |
| Set up branch protection on master (require CI pass) — requested in three retrospectives          | 🔴 `TODO` | Low      | 5min  | Repo settings; needs `gh` or web UI                                 |
| Trash the stale `result` symlink in repo root (points at the a4712ab Nix store path; HEAD is past it) | 🔴 `TODO` | Low      | 1min  | `result -> /nix/store/...a4712ab...`; gitignored                   |

## Notes

- Decisions awaiting the owner (not tasks): `go.work.sum` git tracking, `v0.0.0`
  vs real versions for sibling requires — see ROADMAP.md "Open questions".
