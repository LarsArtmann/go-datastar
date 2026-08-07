# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `MemoryStore` type: ring buffer implementing `sse.EventStore` for reconnection replay
- `NewMemoryStore(capacity)` and `DefaultMemoryStoreCapacity` (128)
- E2E HTTP round-trip test verifying wire-format parity with the DataStar SDK
- Nix flake `go-sse-src` input for hermetic builds with local `replace` directive
- Response method tests covering all 11 methods previously at 0% coverage
- ADR 001 documenting the go-datastar/go-sse/SDK layered architecture

### Changed

- `flake.nix` `vendorHash` computed and set (was `lib.fakeHash`)
- `postPatch` in Nix build copies go-sse source to resolve local `replace` directive

### Fixed

- `example/main.go` rewritten with pure DataStar attributes (zero JavaScript)

## [0.1.0] - 2026-01-01

### Added

- Initial release
