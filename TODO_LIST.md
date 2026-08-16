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
| Bump Go toolchain/directives `1.26.5` → `1.26.6` — clears the 4 stdlib vulns (GO-2026-5972, GO-2026-6089, GO-2026-6090, GO-2026-6218) that turn the govulncheck CI job red on master; folds into the BLOCKED `go` directive policy below | 🔴 `TODO` | High | 5min | CI run 31931262532, govulncheck job: all four "Fixed in go1.26.6" |
| Decide and apply the `go` directive policy: go.mod/go.work say `go 1.26.5` but the v0.0.2/v0.0.3 CHANGELOG entries claim it was lowered to `go 1.26` — the lowering has never landed at any tag | 🔵 `BLOCKED` (user decision) | Medium | 5min | `go.mod:3`, `datastartest/go.mod:3`, `go.work:1`; `git show v0.0.3:go.mod` still says `1.26.5` |
| Silence the erraudit CI job's install failure while the erraudit repo stays private (remove job or gate it) — the red X is noise the `continue-on-error` comment already accepts | 🔴 `TODO` | Low | 5min | CI run 31931262532 `Install erraudit` step; `ci.yml` job comment |
| Add per-module Nix hermetic checks (`hermeticCheckStatic`, `hermeticCheckDatastartest`) — `flake.nix` still carries the TODO; only root is built by `nix flake check` | 🔴 `TODO` | Medium | 60min | `flake.nix:45-47`                                                   |
| Wire `dprint.json` into treefmt/flake.nix (or remove it) — committed config with no integration; formatter today is treefmt (gofumpt/goimports/golines/nixfmt) only | 🔴 `TODO` | Medium | 30min | `dprint.json`; `flake.nix:86-101` has no dprint                     |
| Add `CollectPost` error-path tests: handler returns 400/500 or non-SSE body — current helpers assume success | 🔴 `TODO` | Medium | 30min | No such test in `datastartest/collect_test.go`                      |
| Exercise `CollectPost` and `CollectN` in the `datastartest/e2e_test.go` dogfood test — E2E currently covers only `Collect` + `WithLastEventID` | 🔴 `TODO` | Medium | 20min | `datastartest/e2e_test.go:43,138,149` — Collect-only                |
| Add integration test for `example/main.go`'s `WithOnDrop` drop callback (fill subscriber buffer, assert drop fires) — callback currently unreachable by any test | 🔴 `TODO` | Medium | 30min | `example/main.go`; `7d6e423` added the callback untested            |
| Add fuzz test for `datastartest.UnmarshalSignals` error paths (malformed JSON) — fuzz covers the reader only | 🔴 `TODO` | Low | 20min | `datastartest/reader_fuzz_test.go`; no UnmarshalSignals fuzz        |
| Re-verify the README comparison table when datastar-go cuts its next release and update the pinned v1.2.2 footnote; optionally add the re-check to the release checklist | 🔴 `TODO` | Low | 10min | `README.md` "Compared against datastar-go v1.2.2"                   |
| Decide whether to mention the embedded JS client version (v1.0.2) in the README comparison's "Serve the DataStar JS client" row | 🔴 `TODO` | Low | 5min | `README.md` ScriptHandler row; `static/static.go:11`                |
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
