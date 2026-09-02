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

| Task                                                                                                                                                                    | Status    | Impact | Effort | Evidence                                                                                       |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------- |
| Investigate the datastartest FOD source-sensitivity (hash moved across tree states with identical go.mod — likely the local `replace => ..` under `proxyVendor`), then correct ADR 004's mechanism claim. | 🔴 `TODO` | High   | 2h     | 20-12 report f.3/e.2; 11-07 report §3; ADR 004                                                  |
| Add an exact-CI-version golangci-lint app to flake.nix (`go run …/golangci-lint/v2/cmd/golangci-lint@v2.12.2`) and make it the pre-push habit — kills the "green locally, red in CI" class (caused the 2026-08-29 red master). | 🔴 `TODO` | High   | 30min  | 20-12 report f.1/e.1; commit `8cc56a7`                                                          |
| Verify tag `v0.3.0`'s flake state in a `git worktree` + `nix flake check`; record the verdict (expected FAIL per the vendorHash finding).                                | 🔴 `TODO` | Medium | 30min  | 20-12 report f.2, d.1; fix on master `1f3cd93`                                                  |
| Migrate the 4 remaining `encoding/json` call sites to `encoding/json/v2` (inbound.go:36, signals.go:86, script_convenience.go:105, benchmark_test.go:92) and modernize `b.N` → `b.Loop()` ×4 (benchmark_test.go) — clears the gopls stdversion/bloop warnings. | 🔴 `TODO` | Medium | 45min  | LSP diagnostics 2026-09-02; 07-38 f.20-f.21; 07-55 f.22-f.23                                     |
| CI hygiene batch: add `go mod verify` step, version-drift detection, and a `go.work` `use`-vs-disk check to ci.yml; extend module_boundary_test to run `go mod tidy` in check mode. | 🔴 `TODO` | Medium | 1h     | 07-38 f.9-f.11; 07-55 f.8-f.10; 07-27 F.11-F.13; 20-12 f.36                                      |
| `docker build example/` (+ compose up) to verify the shipped Dockerfile actually builds — a Dockerfile never built is documentation, not packaging.                     | 🔴 `TODO` | Medium | 30min  | 20-12 report f.11, e.9; `example/Dockerfile`                                                    |
| Re-run `docs/performance.md` benchmarks with default benchtime for stable stats and correct the prose numbers (badge is live-current; prose lags).                      | 🔴 `TODO` | Low    | 30min  | 20-12 report f.21; `docs/performance.md`                                                        |
| Response test-depth batch: parse JSON in `TestErrorResponseFromError` (drop substring asserts), `ErrorResponseFromError` integration + fuzz + benchmark, nil-error/nil-detail cases, `NotificationResponse`/`ErrorResponse` edge cases, concurrent `Response` calls, large multi-line elements, nested outbound signals, query+body precedence. | 🔴 `TODO` | Low    | 2h     | 07-04 report f.11-f.17, f.25-f.27, f.31-f.33; 09-18 f.30-f.39                                    |
| Coverage re-measure before printing numbers into docs (ritual after code changes); lint-cache before/after timing (closes the 12.3 measurement debt).                  | 🔴 `TODO` | Low    | 30min  | 20-12 report f.32; re-verified current 2026-09-02 (98.4%/92.7%)                                 |
| Plan v0.4.0 scope: batch the `[Unreleased]` items (go-sse v0.6.0, goldens, accessors, ReplaceURLQuerystring, guides) into the next release.                             | 🔴 `TODO` | Medium | 30min  | 20-12 report f.41; CHANGELOG `[Unreleased]`                                                     |
| Watch the first runs of the new CI workflows and promote/act — runbook: [docs/ci-watch.md](docs/ci-watch.md) (promote/drop/verify per workflow: `nix.yml`, `fuzz.yml`, `codeql.yml`, Renovate). | 🔴 `TODO` | Low    | 15min  | 20-12 report f.12, f.14-f.16, f.42, f.44; `docs/ci-watch.md` |
| Micro-hygiene batch: heartbeat producer ticker → named constant (`example/main.go`), `http.NewResponseController` in `gzipSSEWriter`, stable `nix run .#bench` headers, PR-template multi-module row. | 🔴 `TODO` | Low    | 45min  | 20-12 report f.34, f.35, f.47; 04-48 f.43; `example/main.go`                                     |
| Upgrade `module_boundary_test.go` from text-scanning to `golang.org/x/mod/modfile` parsing (comment-safe).                                                               | 🔴 `TODO` | Low    | 30min  | 07-55 report f.14, e.2; `module_boundary_test.go`                                               |
| Community files batch: CODEOWNERS (owner must name them), SUPPORT.md, DISCUSSION_TEMPLATE.md — decide which apply to a single-maintainer library.                       | 🔵 `BLOCKED` | Low | 30min | 07-04 report f.23, f.46, f.47; no files in .github/                                             |
| Reply to the 5 CodeRabbit threads on merged PR #3 (courtesy closure; fixes landed in `ed815c7`); human-review the 5 parallel-session commits (`ce3b4bc` set).            | 🔴 `TODO` | Low    | 45min  | 12-24 report b.1, c.6; 17-05 report f.45-f.46                                                    |

## Owner-blocked

| Task                                                                                                                                                                    | Status       | Impact | Effort | Evidence                                                                                       |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------- |
| Dependency bots: keep Renovate or Dependabot, disable the other (both active today — duplicate PR churn); also set Renovate dashboard/label preferences.                | 🔵 `BLOCKED` | Medium | 15min  | 20-12 report f.4, f.29-f.30, f.37, Q2; `.github/dependabot.yml` + `renovate.json` coexist       |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                                       | 🔵 `BLOCKED` | Medium | 5min   | `git branch -a`; 12-43 report f.4; 20-12 f.18                                                    |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.                        | 🔵 `BLOCKED` | Medium | 15min  | `git branch -a`; 12-24 report c.2; 20-12 f.19                                                    |
| Configure `git-town.observed-branches` for `preserve/…` while deletion is blocked (stops town aborts).                                                                   | 🔴 `TODO`    | Low    | 5min   | 20-12 report f.20; config absent (verified 2026-09-02)                                          |
| erraudit CI job flips to hard gate when the repo goes public — verify the probe notices the flip on the first push after publication.                                   | 🔵 `BLOCKED` | Low    | 5min   | `ci.yml` probe job; AGENTS.md Nix gotchas                                                       |

## Notes

- Rebuilt 2026-09-02 (docs-health AUDIT pass): every `docs/status/2026-0*` and
  `docs/planning/2026-0*` file re-annotated item-by-item against the code; the
  open residue below is the deduped harvest. Prior harvest (2026-08-29) items
  shipped as T01–T27: toolchain 1.26.7, green master, v0.3.0 (proxy-verified),
  docs-only CI paths, status index, PR-template guard, fuzz docs, AGENTS
  pruning, nix CI + bench, ADRs 003–006, consumer guide set,
  ReplaceURLQuerystring, compression middleware, typed datastartest accessors,
  domain-adapter example, CI expansions (fuzz/CodeQL/Renovate), community
  polish — plus the post-daemon commit: go-sse v0.6.0, encoding/json/v2
  migration, assertion dedup, wire-format golden tests.
- Resolved questions (go.work.sum tracking, sibling requires, go directive
  policy) are documented in ROADMAP.md "Resolved questions" and AGENTS.md.
- datastartest helper-API micro-ideas (RequireElementsOrdered, Diff, Snapshot,
  ServeSSE, Recorder, RawSSE, fluent Assert, Ginkgo matchers, …) are
  consolidated as ONE raw idea in ROADMAP.md Theme 2 — not tracked per-item.
