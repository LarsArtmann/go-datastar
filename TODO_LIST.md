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

## High Impact

| Task                                                                                              | Status    | Impact | Effort | Evidence                                                                                    |
| ------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------- |
| Fix CONTRIBUTING.md — add `GOEXPERIMENT=jsonv2`, `GOWORK=off`, nix workflow (`nix flake check`)   | 🔴 `TODO` | High   | 10min  | `CONTRIBUTING.md` verified missing all three; `go test` fails without `GOEXPERIMENT=jsonv2` |
| Fix `WithScriptAttributeKVs` doc/code mismatch — doc says "errors on odd args" but code silently truncates | 🔴 `TODO` | High   | 30min  | `script.go:58-76` — loop `i+1 < len(kvs)` drops trailing element; no error path             |
| Update AGENTS.md file layout table — add `example_test.go`, `inbound_fuzz_test.go` rows           | 🔴 `TODO` | High   | 10min  | AGENTS.md layout table verified missing both files                                          |
| Add HEAD/RFC 7231 compliance to AGENTS.md wire-format parity requirements (requirement #12)       | 🔴 `TODO` | High   | 10min  | AGENTS.md §"Wire-Format Parity" has 11 items, HEAD not mentioned                            |
| Update `doc.go` package comment to mention classified errors                                      | 🔴 `TODO` | High   | 15min  | `doc.go` has no mention of error system; README and AGENTS.md both document it              |

## Medium Impact

| Task                                                                                              | Status    | Impact | Effort | Evidence                                                                                    |
| ------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------- |
| Add error codes table (9 codes) to README                                                          | 🔴 `TODO` | Med    | 20min  | README documents 3 families but not the 9 codes; codes listed in `errors.go`                |
| Fix `DispatchCustomEventPatch.Event()` silent error swallowing — marshal failure sets `null`      | 🔴 `TODO` | Med    | 30min  | `script_convenience.go:117` — `json.Marshal` failure swallowed                             |
| Add `erraudit` to CI                                                                               | 🔴 `TODO` | Med    | 30min  | `.github/workflows/ci.yml` — no erraudit step; `flake.nix` has lint/coverage apps but no erraudit check |
| Add `govulncheck` to CI                                                                            | 🔴 `TODO` | Med    | 30min  | `.github/workflows/ci.yml` — no govulncheck; available in devShell but not CI               |
| Pin `golangci-lint` version in CI (currently `@latest`)                                            | 🔴 `TODO` | Med    | 15min  | `ci.yml:50` — installs `@latest`, non-reproducible                                          |
| Upgrade GitHub Actions: `checkout@v4`→`v5`, `setup-go@v5`→`v6`                                     | 🔴 `TODO` | Med    | 15min  | `ci.yml` — uses v4/v5, Node 20 deprecation warnings                                         |
| Add golangci-lint / erraudit / govulncheck as nix checks                                           | 🔴 `TODO` | Med    | 1h     | `flake.nix` — only `checks.format` and `checks.build` exist                                 |
| Add benchmark tests for patch `Event()` generation (ElementsPatch, SignalsPatch, ScriptPatch)      | 🔴 `TODO` | Med    | 1h     | No `Benchmark*` functions in any `*_test.go` file                                           |
| Add fuzz test for `MarshalSignals` (untrusted Go value → JSON)                                    | 🔴 `TODO` | Med    | 30min  | `inbound_fuzz_test.go` covers ReadSignals; MarshalSignals has no fuzz                       |
| Add `input_preview` (first ~200 bytes) to `CodeSignalsUnmarshalFailed` error context              | 🔴 `TODO` | Med    | 30min  | `errors.go` / `inbound.go` — currently only `input_bytes`; preview more diagnostic          |
| Integrate `errorfamily.HTTPStatus(err)` into `ErrorResponse` so handlers pick status from family  | 🔴 `TODO` | Med    | 1h     | `response.go` — `ErrorResponse` does not use family→HTTPStatus mapping                      |
| Add `SECURITY.md` and `CODE_OF_CONDUCT.md`                                                         | 🔴 `TODO` | Med    | 20min  | Neither file exists                                                                          |
| Create issue templates (bug, feature) and PR template                                              | 🔴 `TODO` | Med    | 30min  | `.github/ISSUE_TEMPLATE/` and `PULL_REQUEST_TEMPLATE.md` do not exist                       |
| Add Dependabot or Renovate config                                                                  | 🔴 `TODO` | Med    | 20min  | No `.github/dependabot.yml` or `renovate.json`                                              |
| Add `//nolint` comments with rationale on accepted `generic_return` / `silent_swallow` sites       | 🔴 `TODO` | Med    | 20min  | 4 `generic_return` + 1 `silent_swallow` warnings accepted by design (`AGENTS.md`)           |

## Low Impact

| Task                                                                                              | Status    | Impact | Effort | Evidence                                                                                    |
| ------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------- |
| GitHub repo polish: set topics (`datastar`, `sse`, `go`, `hypermedia`), disable empty wiki        | 🔴 `TODO` | Low    | 10min  | `gh repo edit` — topics not set, wiki enabled but empty                                      |
| Add `erraudit` to `flake.nix` devShell                                                             | 🔴 `TODO` | Low    | 10min  | `flake.nix` devShell — has golangci-lint, govulncheck, gopls, templ; no erraudit            |
| Add markdown formatter (e.g., `denofmt`) to treefmt                                                | 🔴 `TODO` | Low    | 20min  | `flake.nix` treefmt — formats Go + Nix only                                                  |
| Add `errors_example_test.go` showing all three error-handling patterns (code, sentinel, family)    | 🔴 `TODO` | Low    | 30min  | `errors_test.go` has contract tests but no compileable usage examples                        |
| Address `nestif` complexity in `ReadSignals` (complexity 6)                                        | 🔴 `TODO` | Low    | 30min  | `inbound.go` — `ReadSignals` has nested conditionals                                         |
| Verify `go-error-family` v0.10.0 is the latest release                                             | 🔴 `TODO` | Low    | 10min  | `go.mod` — pinned at v0.10.0                                                                |
| Verify pkg.go.dev docs are rendered for latest version                                             | 🔴 `TODO` | Low    | 10min  | Badge links there but rendering unconfirmed                                                  |
| Add coverage badge to README                                                                       | 🔴 `TODO` | Low    | 20min  | README has CI/GoRef/GoReport/License badges; no coverage badge                               |
