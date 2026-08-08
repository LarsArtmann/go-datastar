# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open items

| Task                                                                                       | Status       | Impact | Effort | Evidence                                                           |
| ------------------------------------------------------------------------------------------ | ------------ | ------ | ------ | ------------------------------------------------------------------ |
| GitHub repo polish: set topics (`datastar`, `sse`, `go`, `hypermedia`), disable empty wiki | 🔵 `BLOCKED` | Low    | 10min  | Requires `gh` CLI access; topics not set, wiki enabled but empty   |
| Address `nestif` complexity in `ReadSignals` (complexity 6)                                | 🔴 `TODO`    | Low    | 30min  | `inbound.go` — `ReadSignals` has nested conditionals               |
| Verify pkg.go.dev docs are rendered for latest version                                     | 🔴 `TODO`    | Low    | 10min  | Badge links there but rendering unconfirmed                        |
| Add coverage badge to README                                                               | 🔴 `TODO`    | Low    | 20min  | README has CI/GoRef/GoReport/License badges; no coverage badge     |
| Tag v0.0.3 release                                                                         | 🔵 `BLOCKED` | High   | 30min  | All code changes stable; waiting for user release cadence decision |

## Completed this session (2026-08-08)

All items below were resolved during the Pareto hardening session. See CHANGELOG `[Unreleased]` for details.

| Task                                                                                                                               | Task ID |
| ---------------------------------------------------------------------------------------------------------------------------------- | ------- |
| Fix CONTRIBUTING.md — add `GOEXPERIMENT=jsonv2`, `GOWORK=off`, nix workflow                                                        | T01     |
| Fix `WithScriptAttributeKVs` doc/code mismatch — corrected doc to match silent-drop behavior                                       | T03     |
| Update AGENTS.md file layout table — added `example_test.go`, `inbound_fuzz_test.go`, `coverage_test.go`, `errors_example_test.go` | T01     |
| Add HEAD/RFC 7231 compliance to AGENTS.md wire-format parity requirements (requirement #12)                                        | T01     |
| Update `doc.go` package comment to mention classified errors                                                                       | T01     |
| Add error codes table (11 codes) to README with families and retryability                                                          | T07     |
| Fix `DispatchCustomEventPatch.Event()` silent error swallowing — detail now marshaled in constructor                               | T06     |
| Add `erraudit` to CI (`--severity-threshold error`)                                                                                | T05     |
| Add `govulncheck` to CI                                                                                                            | T05     |
| Pin `golangci-lint` version in CI to v2.12.2                                                                                       | T05     |
| Add golangci-lint / erraudit / govulncheck as nix apps                                                                             | T10     |
| Add benchmark tests for patch `Event()` generation                                                                                 | T08     |
| Add fuzz test for `MarshalSignals` round-trip                                                                                      | T08     |
| Add `input_preview` (first 200 bytes) to `CodeSignalsUnmarshalFailed` error context                                                | T06     |
| Add `ErrorResponseFromError` helper using `errorfamily.HTTPStatus`                                                                 | T06     |
| Add `errorfamily.WrapOncef` at `ReadSignals` boundary to prevent double-classification                                             | T06     |
| Add `errors.As(err, &target)` test for `*errorfamily.Error` on every error path                                                    | T06     |
| Document error-code naming convention in `errors.go`                                                                               | T06     |
| Add `SECURITY.md` and `CODE_OF_CONDUCT.md`                                                                                         | T04     |
| Create issue templates (bug, feature) and PR template                                                                              | T04     |
| Add Dependabot config (gomod + github-actions)                                                                                     | T11     |
| Add `errors_example_test.go` showing all three error-handling patterns                                                             | T07     |
| Clean CHANGELOG `[Unreleased]` — removed internal noise entries                                                                    | T02     |
| Fix vague annotations in typed-error-system status report with commit hashes                                                       | T02     |
| Verify `go-error-family` v0.10.0 is the latest release (confirmed latest)                                                          | T13     |
| Verify DataStar JS v1.0.2 is the latest release (confirmed latest)                                                                 | T13     |
| Upgrade CI Actions versions: `checkout@v4→v5`, `setup-go@v5→v6` (8 references)                                                     | F1      |
| Add `TestErrorResponseFromError` — Rejection, Transient, non-errorfamily paths                                                     | F2      |
| Fix `ErrorResponseFromError` doc — Classify defaults to Transient (503), not Rejection (400)                                       | F2      |
| Lower go.mod from `go 1.26.5` to `go 1.26` (matching v0.0.2 CHANGELOG claim)                                                       | F3      |
| Add CHANGELOG entry for CI Actions upgrade (`checkout@v5`, `setup-go@v6`)                                                          | F4      |
| Annotate prior status report (06-52) with resolution pointer                                                                       | F5      |
| Run `actionlint` on ci.yml — 0 violations, YAML valid                                                                             | F6      |
| Verify checkout@v5/setup-go@v6 breaking changes — none affect this CI config                                                        | F7      |
