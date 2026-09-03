# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`Response` method forms of the response helpers** — `resp.ErrorResponse(msg, code)`,
  `resp.ErrorResponseFromError(err)`, and `resp.NotificationResponse(msg, kind)`
  mirror the package-level functions on the response's own stream (additive;
  the package functions are unchanged).
- **`version` package** — a build-time `Version` var for binary consumers
  (`-ldflags "-X …/version.Version=vX.Y.Z"`); documented in CONTRIBUTING.
- **`.goreleaser.yaml` skeleton** for tag-triggered example-binary releases
  (check-validated only; the library release path remains the lockstep tags
  per docs/release-checklist.md).

### Added — datastartest

- **Three debug-grade helpers** (the head of the consolidated helper-expansion
  idea, absorbing the requests from the 2026-08-10/08-16/08-29 status
  reports): `RequireElementsOrdered` (ordered elements-stream assertions with
  `ElementExpectation`, tolerant of interleaved signals, strict about extra
  elements events), `Diff` (line-based diff between two event streams with
  decoded datalines), and `Snapshot` (golden-file testing against
  `testdata/<TestName>.golden`, regenerate with `-datastartest-update`).
  datastartest coverage rose from 92.7% to 93.4%.

### Changed — static

- **Embedded JS client bumped to v1.0.3** (upstream 2026-08-27): opt-in CSP
  mode (`data-nonce`, no `unsafe-eval` required), signals are resent when a
  backend request retries after a network error, view-transition support is
  also checked on the document, and a dynamic `multiple` on `<select>` now
  updates the bound signal. Also switches the embedded artifact to the
  canonical minified upstream bundle — the served asset shrinks from a
  beautified ~56 KB to 33.5 KB. Checksum pin updated in the same change;
  wire-format goldens unchanged.
- `ScriptHandler`/`ScriptHandlerWith` now set `X-Content-Type-Options: nosniff`;
  the embedded bundle is pinned by a SHA-256 checksum test
  (`static/checksum_test.go`) and a provenance comment documents the upstream
  source. `Last-Modified` was evaluated and deliberately not set (ETag +
  Cache-Control fully own freshness; see docs/static-js.md).

### Deprecated

- `datastar.DatastarJSVersion` — use `datastar.Version()` or
  `github.com/larsartmann/go-datastar/static.Version` instead. The constant
  remains for compilation compatibility and will be removed in a future
  minor release.

## [0.4.0] - 2026-09-03

### Added

- **`ReplaceURLQuerystring`** — upstream-parity convenience
  (`NewReplaceURLQuerystringPatch` + `Response.ReplaceURLQuerystring`):
  replaces the browser URL's query string with the encoded values while
  preserving the request path, mirroring the upstream SDK's semantics
  (`url.Values.Encode()` key sorting, fragment dropped). Removes a documented
  parity gap from the README's "where the official SDK wins" table.
- `wire_golden_test.go` — `TestPatchWireGoldens` pins the exact SSE wire bytes of every patch family (elements with all options, signals with `onlyIfMissing`, script with attributes/autoRemove, redirect/console-log sugar, custom-event dispatch): event names, data-line keys and ordering, default-value elision, `id:`/`retry:` placement, and the one-`data:`-line-per-source-line splitting for multi-line payloads. A change to any golden is a wire-format change — make it deliberately, against the DataStar SDK, and record it here.

### Added — documentation

- **Consumer guide set under `docs/`**: `replay.md` (MemoryStore +
  Last-Event-ID reconnection), `error-system.md` (the three matching
  dimensions), `wire-format.md` (annotated datalines per patch family),
  `testing.md` (datastartest quick start, fuzzing, coverage story),
  `performance.md` (measured benchmark table), `migration-guide.md`
  (v0.2.0 → v0.3.0), and `static-js.md` (embedded JS pinning and upgrade
  process).
- **ADRs 003–006**: error classification (go-error-family only, never
  samber/oops), hermetic Nix checks and where the line is drawn, coverage
  strategy (badged, not gated), and the formatter decision (treefmt
  canonical, dprint.json as non-Go intent).
- `docs/status/README.md` — navigation index for the point-in-time reports,
  with the archiving policy.

### Added — examples and CI

- **`example/domain-adapter/`** — a miniature EventBridge: domain events →
  `[]Patch` values → broadcaster + MemoryStore, E2E-tested with datastartest.
- **`example/sse_middleware.go`** — gzip SSE compression middleware with
  per-event flushing (the recommended alternative to in-library
  compression), tested.
- **`example/README.md`, `Dockerfile`, `docker-compose.yml`** — the demo now
  documents its heartbeat keep-alive and runs from one command.
- **`nix.yml`** — the hermetic `nix flake check` gate runs on CI
  (non-required until stable); **`fuzz.yml`** schedules daily 60s fuzz runs
  with crash-artifact upload; **`codeql.yml`** adds Go security scanning;
  **`renovate.json`** proposes embedded-JS-client bumps from upstream
  releases; the lint job caches golangci-lint's analysis cache; docs-only
  changes skip the code gates (paths filters) while `actionlint.yml` still
  validates workflows on every push.
- **`nix run .#bench`** plus committed `datastartest` benchmarks for the
  Collect round trip and the raw parser floor.
- **Typed script accessors in datastartest** (`RedirectURL`,
  `CustomEventName`, `CustomEventDetail`, `UnmarshalCustomEventDetail`,
  `ScriptAttributes`) for intent-level assertions, with the new
  `datastartest.custom_event_detail_unmarshal_failed` code.

### Changed

- Root benchmarks migrated from `for range b.N` to `for b.Loop()` (the
  modern Go benchmark loop with automatic timer handling); benchmark
  semantics and results are unchanged.
- The flake's `vendorHash` comments and AGENTS.md Nix gotcha corrected to the
  verified sensitivity mechanism (ADR 004): root `vendorHash` moves only on
  requires/toolchain changes, while `datastartestVendorHash` additionally
  moves on any root/static source edit (directory-replaced package source is
  copied into the vendor tree). Both hashes refreshed to the current module
  set; ADR 004 gained the evidence matrix and the v0.3.0 tag flake verdict.

### Changed — CI

- The ci.yml test job is split into a per-module matrix (root, datastartest,
  static — build/vet/race/verify/tidy each in GOWORK=off isolation, in
  parallel) plus a workspace job (workspace race suite, go.work use-vs-disk,
  sync idempotency, replace audit). Feedback time drops from one serial job
  to the slowest parallel leg. `go work vendor` was evaluated and deliberately
  not adopted: the hermetic path already vendors per module through
  buildGoModule, and a committed workspace vendor/ tree would duplicate
  replaced source (ADR 004's split-brain warning).

### Changed — dependencies

- **go-sse bumped from v0.5.1 to v0.6.0** in the root module, and
  `go-sse/ssetest` from v0.2.0 to v0.3.0 in `datastartest` (root go.mod,
  datastartest/go.mod). No library API changes were required. Two
  `encoding/json` call sites in example and test code moved to
  `encoding/json/v2` (`json.UnmarshalRead`) for consistency with the
  `GOEXPERIMENT=jsonv2` build requirement.

### Fixed

- **`go.work.sum` policy wording narrowed** (CodeRabbit PR #3 thread): the
  per-module `go.sum` files are the reproducibility source of truth for
  consumers, and the replace directives make sibling checksums unnecessary —
  but workspace-mode hashes for external modules do accumulate in
  `go.work.sum`, so calling it wholesale "advisory" was too broad. AGENTS.md
  and the ROADMAP decision now state the scoped policy; the released v0.3.0
  note stands as written (append-only history).
- **The datastartest hermetic Nix check could never converge** — with the
  repo-root fileset, the `datastartestVendorHash` constant sat inside its own
  fixed-output derivation's input (`go mod vendor` copies replaced module
  directories entirely, flake.nix included), so every pasted hash changed the
  input and re-invalidated itself. The check now builds from a minimal src
  fileset (datastartest, root `*.go` + `go.mod`, `static/`); the hash
  converges on one paste and moves only on genuine module-set changes. ADR
  004 documents the full mechanism, the evidence matrix, and the v0.3.0 tag
  flake verdict.
- **The v0.3.0 tag's Nix flake vendorHash was stale** — the release commit
  bumped the lockstep `requires` after the hash was harvested, so
  `nix flake check` at the tag fails while master is fixed (`1f3cd93`).
  Consumers via `go get` are unaffected (module proxy ≠ flake); re-tagging
  is forbidden (proxy poisoning), so the tag's flake state stands as
  documented. Recorded here for honest release history.
- **`varnamelen` lint fix** — a short stream parameter name that passed the
  devShell linter but failed the CI build (`8cc56a7`); motivates running the
  exact CI linter build locally.

## [0.3.0] - 2026-08-29

### Changed

- **Go directives pinned to 1.26.7 across all modules** (go.mod ×3, go.work,
  CI `go-version` ×6) and the Nix flake `go_1_26` pin re-pointed at the
  go1.26.7 source tarball with re-discovered vendor hashes (the patch bump
  moves the module-set hash). Consumers must use Go ≥ 1.26.7.

### Added — datastartest

- **Request options on every `Collect*` helper** — `WithPath` (target a mux
  route, query strings allowed), `WithHeader`, `WithLastEventID` (simulates a
  reconnecting browser for replay testing), and `WithDatastarSignals`
  (submits the `?datastar=` query parameter the way DataStar clients do with
  GET/DELETE). Previously every helper hard-requested `GET /`, making
  multi-route handlers, GET-signal submissions, and reconnection replay
  untestable through the helpers.
- **`RequireScript`** — one-call assertion on a script patch's inner
  JavaScript source (strips the `<script>` wrapper).
- **`RequireEventID`** — event-ID assertion for replay/reconnection tests.
- **`datastartest/README.md`** — consumer-facing quick start and API tour.

### Changed — datastartest

- **All public helpers now accept `testing.TB`** instead of `*testing.T`, so
  they work with `*testing.T`, `*testing.B`, and Ginkgo's `GinkgoT()`.
  Backward compatible for existing `*testing.T` callers.
- Test coverage raised from 82.2% to 92.7% (assertion failure paths,
  `RequireSignals`, `MustReadNEvents`, option plumbing, and a full
  EventStore replay dogfood E2E via `WithLastEventID`; re-measured 2026-08-29).
- **Response-body Close errors now surface as test errors.** The `Collect*`
  helpers previously ignored `resp.Body.Close()` failures with `_ =`; they now
  report them via `tb.Errorf` instead of silently discarding them. Clears the
  erraudit audit (0 violations with `--enforce-go-error-family`).

### Fixed

- **`NewDispatchCustomEventPatch` godoc corrected** — the constructor comment
  claimed the detail value was marshaled lazily when `Event()` is called; it
  has been marshaled in the constructor (with a classified error on failure)
  since v0.0.3. The doc now matches the code.

### Fixed — datastartest

- **SSE wire parser brought to WHATWG HTML § 9.2.6 conformance**, synced from
  go-sse's `ssetest`. Six deviations corrected, all behavioral (no API
  signature changes): lone CR is now a line terminator (§ 9.2.5), an
  incomplete final frame at EOF is discarded, exactly one leading UTF-8 BOM
  is stripped, an `id:` value containing NUL is ignored, the last event ID is
  sticky connection state (an event reports the most recent `id:` value at
  dispatch; empty `id:` resets it), and the `retry:` reconnection time is
  likewise sticky (updated even by dataless frames, never reset by invalid
  values, parsed at 64-bit width).
  _Superseded refinement:_ the parser is no longer a synchronized duplicate —
  `datastartest` now delegates parsing directly to the shared
  `go-sse/ssetest` package (v0.2.0), so conformance fixes land once for both
  consumers.
- **The official Web Platform Tests `eventsource/format-*` corpus is now
  transcribed as Go tests** (`wpt_format_corpus_test.go`: 15 WPT vectors,
  3 spec § 9.2.6 example streams, 8 Chromium `event_source_parser_test.cc`
  cases, each with its upstream citation), re-run through 1–4096 byte chunked
  readers (`chunk_boundary_test.go`) to prove TCP-chunking independence.
- **Conformance fuzz corpus ported from go-sse/ssetest** — 51 committed
  regression seeds under `testdata/fuzz/FuzzReadEvents/`, including the
  `"0data: hello\n\n"` crasher and the trailing-LF terminator regression,
  keeping the two modules' fuzz corpora in lockstep.

### Changed — dependencies and example

- **go-sse bumped from v0.5.0 to v0.5.1** in the root module and
  `datastartest`, with `datastartest` additionally depending on
  `go-sse/ssetest` v0.2.0 for the shared SSE test parser.
- **Example adds a heartbeat mechanism** to the SSE event handler and ships a
  rebuilt example binary, demonstrating a keep-alive pattern on top of
  go-sse's broadcaster.

### Security — CI and toolchain

- **Go directives pinned to 1.26.6 across all modules** (go.mod ×3, go.work,
  CI `go-version`) — clears the four stdlib vulnerabilities that kept the
  govulncheck job red on master (GO-2026-5972, GO-2026-6089, GO-2026-6090,
  GO-2026-6218; all "Fixed in go1.26.6"). The Nix flake pins
  `go_1_26` to 1.26.6 via `overrideAttrs` until nixpkgs ships it, so
  `nix flake check` stays hermetic and green alongside CI. _Superseded
  within this release by the 1.26.7 pin in "Changed" above — see that entry
  for the final toolchain state._
- **erraudit CI job probe-gated and un-red-X'd** — the job now probes
  `go list -m github.com/larsartmann/erraudit@v0.3.0` first: while the
  erraudit repository is private (as of 2026-08-16) the job skips with a
  visible notice instead of failing at install; once the module resolves
  publicly, the audit runs as a hard gate over each module in turn. Also
  fixes the latent broken invocation: erraudit accepts one directory
  argument, not three package patterns — the old command could never have
  passed even with the repo public.

### Changed — developer tooling

- **`dprint.json` kept for non-Go formatting** — it stays in the repo to
  document the project's intent for non-treefmt consumers and editor
  integrations (JSON, YAML, Markdown, Dockerfile). It is deliberately not
  wired into treefmt or `nix flake check`: doing so would make the hermetic
  build depend on network-fetched WASM plugins. treefmt (gofumpt, goimports,
  golines, nixfmt) remains the single canonical formatter for the Nix
  hermetic build.
- **`actionlint` CI job added** — workflow YAML is now linted on every push
  (pinned v1.7.12), so a bad workflow edit can no longer silently redden
  master.
- **`erraudit` app fixed and `actionlint` added to the devShell** —
  `nix run .#erraudit` now audits each module with the correct
  single-directory invocation (the tool rejects package patterns; the old
  app command could never have passed). erraudit itself stays out of the
  hermetic devShell: its dependency tree contains private modules (e.g.
  go-finding), which a sandboxed Nix build cannot fetch, so the app
  go-installs it instead and requires local GitHub credentials.
- **Per-module Nix hermetic checks** — `nix flake check` now builds and tests
  all three Go modules in isolation (`checks.build`, `checks.buildStatic`,
  `checks.buildDatastartest`), mirroring the CI `GOWORK=off` per-module legs.
  The `static` module uses `vendorHash = null` (zero deps); the `datastartest`
  module uses `modRoot` so its sibling replaces resolve inside the sandbox.
- **Release checklist** — `docs/release-checklist.md` codifies the pre-release
  gate, version bump, tag, post-release `pkg.go.dev` verification, and quarterly
  comparison-table re-verify against upstream datastar-go.
- **Domain language** — `docs/DOMAIN_LANGUAGE.md` defines the ubiquitous
  language (Patch, Signals, Dataline, Replay, Family, Code, etc.) for
  consistent naming and documentation.
- **Modularization docs index** — `docs/modularization/README.md` links the
  proposal, execution plan, and ADRs 001/002. Linked from the AGENTS file
  layout section.
- **`go` directive policy: pin the exact patch release.** The v0.0.2/v0.0.3
  decision to "lower to `go 1.26`" (avoiding patch versions for consumer
  compatibility) is superseded: 1.26.6 clears four stdlib CVEs
  (GO-2026-5972/6089/6090/6218), and `GOTOOLCHAIN=local` in hermetic builds
  forbids auto-downloading a newer toolchain. Consumers must use Go ≥ 1.26.6.
- **`go.work.sum` intentionally gitignored** — `go.work` is force-added for
  workspace development; `go.work.sum` is advisory (regenerated by the go
  toolchain on demand) and the committed per-module `go.sum` files are the
  source of truth for reproducibility. The replace directives make
  sibling-module checksums unnecessary (they resolve to local paths), so
  committing `go.work.sum` would only add diff noise on every dependency
  update.
- **Sibling requires use real published versions** (not `v0.0.0`) — the
  replace directives make versions irrelevant locally, but a consumer testing
  without replaces must resolve to a real published module.
- **`TestCollect_WithLastEventID_HeaderArrives` de-flaked** — the handler now
  signals via a channel after writing and flushing the SSE event, eliminating
  the theoretical race where EOF arrives before the event data under extreme
  parallel load. No sleeps (channel-synchronized).
- **Live coverage badge** — README's coverage badge is now generated by CI
  (`coverage` workflow) from the same root + datastartest + static module set
  as `nix run .#coverage`, not hard-coded. The workflow publishes a
  shields.io endpoint JSON to an orphan `coverage` branch via the built-in
  `GITHUB_TOKEN` (no external service or secret), and the badge links to the
  coverage workflow runs.

## [0.2.0] - 2026-08-13

### Changed — go-sse v0.5.0

- **go-sse bumped from v0.4.0 to v0.5.0** across the root module and
  `datastartest`. go-sse v0.5.0 ships its own `go-error-family` classification
  (codes like `sse.send_failed`), drop observability (`WithOnDrop`), and
  multi-line SSE data helpers (`JoinLines`/`KeyedLines`). go-datastar's
  `wrapStreamError` wraps go-sse's classified errors as
  `datastar.stream_send_failed` (Transient). Because `errorfamily.Classify`
  returns the outermost family, go-datastar's classification wins — which is
  correct since Send failures are transient I/O errors. `errors.Is` still
  traverses the chain, so callers matching go-sse codes work transparently.
- **`datastartest` errors now use `go-error-family`** — Three `fmt.Errorf` call
  sites (SSE scanner errors in `reader.go` and `collect.go`, signals unmarshaler
  in `event.go`) now return `errorfamily`-classified errors with stable codes
  (`datastartest.sse_scan_failed`, `datastartest.signals_unmarshal_failed`).
  `go-error-family` promoted from indirect to direct in `datastartest/go.mod`.
  Consumers can now distinguish test-helper failures by code instead of
  message-matching.
- **Example adopts `WithOnDrop`** — `example/main.go` now registers a drop
  callback on the broadcaster, logging when subscriber buffers overflow. This
  demonstrates go-sse v0.5.0's drop observability feature in a real serving
  context.
- **`JoinLines`/`KeyedLines` not adopted** — go-sse v0.5.0's `splitLines`
  normalizes CRLF to LF, while the upstream DataStar SDK splits on `\n` only
  (wire-format parity items 6-7). Its key convention (`key + " "`) also conflicts
  with go-datastar's trailing-space dataline constants (`"selector "`,
  `"elements "`). Revisit if upstream adopts CRLF normalization. Documented in
  `AGENTS.md`.

### Added — Tooling and supply-chain hardening

- **CI GitHub Actions pinned to commit SHAs** — `actions/checkout` and
  `actions/setup-go` are pinned to specific v7 commit SHAs (verified against the
  `v7` tags) to harden the supply chain against tag-mutation attacks. Version
  comments preserved for reviewer readability.
- **`dprint.json` formatter config** — Centralized formatting for JSON, YAML,
  Markdown, and Dockerfile across editors and CI. Excludes `vendor/`, `.git/`,
  and `CHANGELOG.md`.

### Fixed — Module boundary

- **Circular module dependency between root and datastartest** — Root's go.mod
  no longer requires `datastartest`. `TestE2E_DataStarPatches` was relocated from
  root's `e2e_test.go` to `datastartest/e2e_test.go`, placing the dogfood test
  alongside the helpers it exercises. Root's `e2e_test.go` retains only
  `TestE2E_SSEHeaders` (transport header verification owned by go-sse). This
  eliminates the test-dep leak where a production module (root) depended on a
  test-only module (datastartest) for a single test file.

### Added — CI hardening

- **Per-module isolation check** — CI now runs `GOWORK=off go build` and
  `GOWORK=off go test` in each module directory (`.`, `./datastartest`,
  `./static`) to verify replace directives are sufficient for standalone
  builds.
- **Workspace sync idempotency check** — CI copies `go.work`, runs `go work sync`,
  and fails if the file changed, catching stale workspace state before merge.
- **Replace directive audit** — CI greps all `go.mod` files for absolute paths in
  replace directives, enforcing relative-path-only conventions.

## [0.1.0] - 2026-08-10

This release splits the repository into three independently versioned Go modules:
`go-datastar` (root protocol library), `go-datastar/static` (embedded JS client,
zero dependencies), and `go-datastar/datastartest` (consumer E2E test helpers).
A committed `go.work` workspace ties them together for local development.

### Added — Multi-module architecture

- **`static/` is now a separate Go module** (`github.com/larsartmann/go-datastar/static`).
  Zero dependencies, embeds the DataStar JS client bundle. Consumers can import
  just the JS bytes without pulling the protocol library's dependency tree.
- **`datastartest/` is now a separate Go module** (`github.com/larsartmann/go-datastar/datastartest`).
  Consumer E2E test helpers with their own dependency surface (go-sse, go-datastar).
- **Committed `go.work`** — workspace mode is now the default. `go.work` removed
  from `.gitignore`; fresh clones work immediately without manual `go work init`.
- **CI covers all three modules** — test, vet, lint, erraudit, govulncheck all run
  against `./... ./datastartest/... ./static/...` in workspace mode.
- **Dependabot monitors all three modules** — separate gomod entries for `/`,
  `/datastartest`, and `/static`.

### Added — datastartest API

- **`datastartest/` package** — Consumer-facing E2E test helpers for
  DataStar handlers. Provides `Collect(t, handler)` (one-liner test server +
  GET + decode), `ReadEvents(io.Reader)` (SSE wire-format parser), `Event` type
  with typed accessors (`Selector()`, `Mode()`, `Elements()`, `SignalsJSON()`,
  `UnmarshalSignals()`, etc.), filter helpers (`FilterElements`,
  `FilterSignals`), and assertion helpers (`RequireElements`,
  `RequireElementsContains`, `RequireSignals`, `RequireEventCount`). Includes
  `ScriptContent()` (strip `<script>` wrapper), `CollectWithRequest` /
  `CollectPost` (non-GET requests), and `CollectN` (streaming handlers). The
  library's own `e2e_test.go` was refactored from 261 to 109 lines using this
  package.
- **`datastartest/` expanded API** — `CollectWithTimeout` (defensive
  time-bounded GET), `ReadNEvents` (exported streaming reader),
  `Event.IsScript()` (script-patch predicate), `FindElement` /
  `FindSignals` (search by selector/type), `EventsString` (multi-event debug
  dump), `RequireSignalsContain` (signal key existence assertion),
  `UnmarshalSignals` error now includes JSON preview. Fuzz test
  (`FuzzReadEvents`, 9-seed corpus) and benchmark (`BenchmarkReadEvents`,
  ~131 MB/s) added. 60+ tests, 8 testable examples.
- **`static/` subpackage** — Dedicated asset module owning the embedded
  DataStar JavaScript client bundle. Exports `Bytes() []byte` and
  `Version` const (`"1.0.2"`). Extracted from `script_handler.go` so the
  protocol package no longer owns the raw bytes. Public API unchanged
  (`ScriptHandler`, `ScriptHandlerWith`, `ScriptTag`, `Version`,
  `DatastarJSVersion` all still work via re-exports).

### Changed

- **`script_handler.go` refactored** — Now imports the `static` subpackage.
  `ScriptHandler()` delegates to `ScriptHandlerWith(static.Bytes(),
  static.Version)`. `DatastarJSVersion` is a backward-compatible const alias
  for `static.Version`. No behavioral change.
- **`ScriptContent()` robustness** — Now uses quote-aware tag-end detection
  instead of `strings.Cut(s, ">")`, correctly handling `>` characters inside
  quoted attribute values (e.g., `<script data-x="a>b">`).
- **`ReadNEvents` early return** — `count <= 0` now returns immediately with
  nil instead of reading one event before checking the threshold.
- **`RequireSignalsContain` doc corrected** — was documented as checking
  "top-level property" but actually matches at any nesting level. Doc now
  reflects the substring-matching behavior honestly.

## [0.0.3] - 2026-08-08

### Added

- **HEAD request support in `ScriptHandler`** — HEAD now returns `200 OK` with
  headers (Content-Type, ETag, Cache-Control, Content-Length) but **no message
  body**, complying with RFC 7231 §4.3.2. Previously HEAD wrote the full JS body.
- **Testable examples** (`example_test.go`) — three `Example` functions with
  `// Output:` assertions verifying wire format for `ElementsPatch`,
  `SignalsPatch`, and `ReadSignals`. Compile-checked by `go test`.
- **Fuzz test for `ReadSignals`** (`inbound_fuzz_test.go`) — 10-seed corpus
  covering valid payloads, truncated JSON, null, arrays, control characters,
  invalid UTF-8. 1.2M+ executions, 0 failures. Seeds run as regression cases.
- **Error-handling examples** (`errors_example_test.go`) — three `Example`
  functions showing all three typed error-handling patterns (by code, by
  sentinel, by family).
- **`ErrorResponseFromError`** — new helper that sends a signals patch with
  errorfamily metadata (code, family, retryable, httpStatus) extracted from a
  Go error via `errorfamily.HTTPStatus`, `errorfamily.Code`, etc.
- **CI hardening** — erraudit (`--severity-threshold error`) and govulncheck
  jobs added to CI. golangci-lint pinned to v2.12.2 (was `@latest`).
- **Community files** — `SECURITY.md`, `CODE_OF_CONDUCT.md`, issue templates,
  and PR template added.
- **Error codes table in README** — all 11 error codes with their families and
  retryability, visible in the main documentation.

### Changed

- **CI Actions upgraded** — `actions/checkout` v4→v5 and `actions/setup-go`
  v5→v6 across all 4 CI jobs (test, lint, erraudit, govulncheck).
- **`WithScriptAttributeKVs` doc corrected** — the doc comment incorrectly
  claimed it "returns an error via the patch if the argument count is odd"; the
  code silently drops the unpaired key. Doc now matches the code.
- **`DispatchCustomEventPatch` no longer silently swallows marshal errors** —
  the detail value is now marshaled in `NewDispatchCustomEventPatch`, which
  returns a classified error instead of emitting `null` in `Event()`.
- **`ReadSignals` error context enriched** — unmarshal failures now include
  `input_preview` (first 200 bytes of the offending input) in addition to
  `input_bytes`.
- **`WrapOncef` replaces `Wrapf`** at the `ReadSignals` boundary to prevent
  double-classification when a caller's error flows through.
- **Error-code naming convention documented** — `_failed`, `_invalid`,
  `_required`, `_after_close` suffix rules defined in `errors.go`.

### Fixed

- **Broken godoc example on `Response`** (`response.go`) — called non-existent
  `resp.Close()` (should be `stream.Close()`) and `resp.PatchSignals(map)` (wrong
  signature; should be `MarshalAndPatchSignals`). Same bug class as the README
  fix in v0.0.1, but the godoc copy was never synced.
- **CONTRIBUTING.md missing required environment variables** — `GOEXPERIMENT=jsonv2`
  and `GOWORK=off` are now documented with both Nix and manual workflow sections.
- **AGENTS.md stale file layout** — `example_test.go`, `inbound_fuzz_test.go`,
  and `coverage_test.go` rows added. HEAD/RFC 7231 compliance added as
  wire-format parity requirement #12.
- **go.mod `go` directive lowered** from `1.26.5` to `1.26` — the v0.0.2
  changelog claimed this was done but the change was never applied to go.mod.
  Patch versions in go.mod are unusual and reduce consumer compatibility.
- **`ErrorResponseFromError` doc corrected** — the comment claimed
  non-errorfamily errors default to Rejection (400); in reality Classify
  defaults to Transient (503, fail-open for retry).

## [0.0.2] - 2026-08-07

### Fixed

- **CI lint job now passes** — golangci-lint installed via `go install` (v2)
  instead of the pre-built binary action, which was compiled with Go 1.24 and
  could not analyze Go 1.26 code
- Removed deprecated `stable: false` input from `actions/setup-go@v5`
- Replaced `golangci/golangci-lint-action@v6` with direct `go install` to ensure
  golangci-lint is built with the same Go toolchain as the project

### Changed

- Lowered `go.mod` from `go 1.26.5` to `go 1.26` (patch versions in go.mod are
  unusual and reduce consumer compatibility)
- Added `stream.Send` to wrapcheck ignore-sigs in `.golangci.yml` — the Response
  methods are thin pass-throughs by design
- Added `p`, `r`, `id` to varnamelen ignore-names in `.golangci.yml` — idiomatic
  Go short names for patch, request, and identifier

## [0.0.1] - 2026-08-07

First public release. DataStar protocol library for Go — patches as first-class
values producing `sse.Event`, built on [go-sse](https://github.com/LarsArtmann/go-sse).

### Added

- **Core `Patch` interface** — `Patch interface { Event() sse.Event }`. Every
  protocol message is a value that can be stored, queued, filtered, replayed, and
  broadcast through go-sse's `Broadcaster[T]`, `EventStore`, `SubscribeFilter`.
- **Four patch types**: `ElementsPatch`, `SignalsPatch`, `ScriptPatch`,
  `DispatchCustomEventPatch` — each with functional-option constructors.
- **Convenience patch constructors**: `NewRemovePatch`, `NewRemoveByIDPatch`,
  `NewRedirectPatch`, `NewConsoleLogPatch`, `NewConsoleErrorPatch`,
  `NewReplaceURLPatch`, `NewPrefetchPatch`, `NewSignalsIfMissingPatch`, plus
  printf-style variants (`WithSelectorf`, `NewRedirectfPatch`, `NewConsoleLogfPatch`).
- **`Response` fluent builder** — wraps `sse.Stream` with 16 methods for
  single-connection patching (`PatchElements`, `MarshalAndPatchSignals`,
  `ExecuteScript`, `Redirect`, `ConsoleLog`, `ConsoleError`,
  `DispatchCustomEvent`, `ReplaceURL`, `Prefetch`, `RemoveElement`,
  `RemoveElementByID`, `ApplyPatches`, `Send`, `Stream`, etc.). Plus
  `NewResponseFromHTTP`, `ErrorResponse`, `NotificationResponse` helpers.
- **Template engine adapters**: `ElementsFromTempl` (Templ) and
  `ElementsFromGostar` (GoStar) — render components to HTML without imposing a
  dependency on consumers who prefer a different engine.
- **`MemoryStore`** — in-memory ring buffer implementing `sse.EventStore` for
  SSE reconnection replay. `NewMemoryStore(capacity)` with
  `DefaultMemoryStoreCapacity` (128).
- **Embedded DataStar JS client** (v1.0.2) — `ScriptHandler()` serves the bundle
  with ETag and Cache-Control headers. Also `ScriptHandlerWith` for custom
  bundles, `ScriptTag(path)`, `Version()`.
- **Inbound helpers**: `ReadSignals(r, &target)` extracts signals from
  `?datastar=` query param (GET/DELETE) or JSON body (all other methods);
  `LastEventID(r)` extracts the reconnection event ID.
- **HTTP action helpers**: `GetSSE`, `PostSSE`, `PutSSE`, `PatchSSE`, `DeleteSSE`
  generate DataStar `@get`/`@post`/etc. attribute strings.
- **Sugar helpers**: mode helpers (`WithModeInner`, `WithModePrepend`, ...),
  namespace helpers (`WithNamespaceSVG`, `WithNamespaceMathML`), selector
  helpers (`WithSelectorID`), validation helpers (`ElementPatchModeFromString`,
  `NamespaceFromString`).
- **Typed error system** built on
  [go-error-family](https://github.com/LarsArtmann/go-error-family): every error
  carries a stable code, a behavioral family (Rejection / Transient /
  Orchestration), and structured context. Two sentinel errors
  (`ErrBodyReadAfterClose`, `ErrEventNameRequired`) and nine stable codes.
- **Wire-format parity** with the upstream DataStar SDK — mode `outer` and
  namespace `html` never emitted, retry gating, script wrapping, signal
  splitting, dataline key trailing spaces.
- **Complete test suite** including E2E HTTP round-trip test verifying
  wire-format parity and response method tests covering all builder methods.
- **Example application** (`example/`): live-feed demo using pure DataStar
  attributes with zero JavaScript, broadcasting patches through go-sse.
- **Nix flake** for hermetic builds with `buildGoModule`, treefmt formatting,
  and dev shell.

### Changed

- Removed local `replace` directive — the module now resolves `go-sse v0.4.0`
  and `go-error-family v0.10.0` from the Go module proxy.

[Unreleased]: https://github.com/LarsArtmann/go-datastar/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/LarsArtmann/go-datastar/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/LarsArtmann/go-datastar/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/LarsArtmann/go-datastar/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/LarsArtmann/go-datastar/compare/v0.0.3...v0.1.0
[0.0.3]: https://github.com/LarsArtmann/go-datastar/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/LarsArtmann/go-datastar/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/LarsArtmann/go-datastar/releases/tag/v0.0.1
