# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
