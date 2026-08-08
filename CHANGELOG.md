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

### Fixed

- **Broken godoc example on `Response`** (`response.go`) — called non-existent
  `resp.Close()` (should be `stream.Close()`) and `resp.PatchSignals(map)` (wrong
  signature; should be `MarshalAndPatchSignals`). Same bug class as the README
  fix in v0.0.1, but the godoc copy was never synced.

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
