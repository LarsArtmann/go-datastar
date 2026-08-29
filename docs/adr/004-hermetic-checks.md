# ADR 004: Hermetic per-module checks in the Nix flake — and where the line is

Date: 2026-08-29

## Context

The repo's declared canonical gate is `nix flake check`: sandboxed,
network-free, reproducible. CI (GitHub runners) is fast but hermetic only by
convention (pinned toolchains, SHA-pinned actions). Two build strategies were
compared:

1. **`gitTracked` source + pinned `vendorHash`** (chosen) — the flake copies
   the git-tracked working tree, builds a go-modules fixed-output derivation
   per module, and fails LOUDLY on hash mismatch.
2. **`gitTracked`/`gitTrackedWith` source with computed hashes or git-rev
   pinning** — removes the vendorHash constant but either weakens purity or
   re-downloads on every commit.

`vendorHash` DID break twice in practice (the Go 1.26.5 → 1.26.6 patch bump
changed the `go mod vendor` `modules.txt` format; the 1.26.7 bump moved the
module-set hash again). Both failures were caught immediately by
`nix flake check` — the failure mode is fail-visible, never fail-silent.

## Decision

**Sandboxed `checks.*` (enforced by `nix flake check`):**

- `checks.build` — root module: `GOWORK=off`, `GOEXPERIMENT=jsonv2`, full
  test suite, vendored deps.
- `checks.buildStatic` — `static` module (`vendorHash = null`, zero deps).
- `checks.buildDatastartest` — `datastartest` module via `modRoot` so the
  sibling replaces (`=> ..`, `=> ../static`) resolve inside the sandbox.
- `checks.format` — treefmt (gofumpt/goimports/golines/nixfmt) with `goPkg`
  prepended to `buildInputs` (treefmt-nix's built-in `flakeCheck` would
  register an UNGUARDED `checks.treefmt` without go on PATH; goimports
  shells out to `go env` per file and must not attempt a toolchain download).

**Flake apps (hermetic toolchain, non-sandboxed):** test, test-race, build,
vet, lint, govulncheck, erraudit (go-installed; private deps), coverage, and
`bench` (root + datastartest benchmarks, `-benchmem`).

**Deliberately NOT sandboxed derivations: lint / vet / govulncheck.**
Verdict, after evaluation:

- golangci-lint from nixpkgs is compiled against nixpkgs' Go, but the repo
  pins the exact toolchain (go 1.26.7) via `GOTOOLCHAIN=local`; running a
  foreign-toolchain golangci-lint over the directive-pinned module is the
  kind of version skew the pin exists to prevent.
- Making them hermetic would mean vendoring golangci-lint and govulncheck
  themselves (three more vendorHash pins that move with every Go patch bump —
  the exact fragility documented above) for a gate CI already runs with the
  precisely pinned toolchain.

Until nixpkgs ships a Go ≥ 1.26.7 that matches the directive, CI owns
lint/vet/govulncheck and Nix owns build/test/format.

## Consequences

- `nix flake check` stays fast (4 checks) and green-tolerant of nixpkgs
  golangci-lint availability.
- Every Go patch bump requires the fakeHash dance (AGENTS.md, Nix gotchas) —
  accepted because the failure is loud and the fix is mechanical.
- Benchmarks have a one-command entry point (`nix run .#bench`) and a
  committed `datastartest/collect_bench_test.go` guarding the helper's
  end-to-end cost.
- Revisit this ADR when nixpkgs' `go_1_26` override can be dropped (see the
  TODO in `flake.nix`): at that point golangci-lint matching the directive
  may become buildable hermetically, and lint can move into `checks.*`.
