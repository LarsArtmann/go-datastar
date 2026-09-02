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

`vendorHash` broke in practice several times: the Go 1.26.5 → 1.26.6 patch
bump changed the `go mod vendor` `modules.txt` format; the 1.26.7 bump moved
the module-set hash again; dependency bumps (go-sse v0.6.0) moved both hashes;
and the v0.3.0 tag shipped with a stale `datastartestVendorHash` (see the
verified verdict below). All failures were caught immediately by
`nix flake check` — the failure mode is fail-visible, never fail-silent.

### 2026-09-02 correction: the vendorHash sensitivity mechanism (verified)

An earlier draft of this ADR attributed hash movement to Go patch bumps
(module-set changes) only. That is wrong for `datastartestVendorHash`. A
controlled experiment (2026-09-02, worktree, nix FOD builds + manual
`go mod vendor` tree hashes) established the real mechanism:

**`go mod vendor` copies the ENTIRE directory of a directory-replaced module
into the vendor tree — not just package source.** `datastartest` replaces
`go-datastar => ..` and `static => ../static`, so every tracked file under the
repo root and `static/` (docs, AGENTS.md, CHANGELOG.md, CI config, everything)
lands in `datastartest`'s vendor tree — and its FOD hash moves on ANY edit to
any such file, even with byte-identical go.mod/go.sum. (Proven twice: a root
package comment moved the tree hash, and a docs/AGENTS/CHANGELOG-only edit
moved the nix FOD from `xc54T9…` to `Qp5algG…`.) The root module imports no
replaced package (`go-datastar/static` is never imported by root code), so
root's `vendorHash` is INSENSITIVE to repo files; it moves only when the
module set (go.mod/go.sum requires) or the toolchain's `modules.txt` format
changes.

Evidence matrix (vendor-tree content hash unless noted):

| State | go.mod/go.sum | Source toggle | datastartest hash | Moved? |
| ----- | ------------- | ------------- | ----------------- | ------ |
| replaces active | unchanged | — (baseline) | `d4dd09ac…` | — |
| replaces active | unchanged | comment in root `response.go` | `fde17bae…` | **YES** |
| replaces active | unchanged | comment in `static/static.go` | FOD `ZEoxiszp…` (nix) | **YES** |
| replaces removed | + published v0.3.0 entries | — | `16aafda9…` | — |
| replaces removed | unchanged | comment in root `response.go` | `16aafda9…` | **NO** |

Nix-level cross-check: with replaces active, a root-source comment alone moved
the datastartest FOD from `xc54T9…` to `Hsut8cL…`, while the root FOD stayed
at `dgqHjh3…` across BOTH root- and static-source toggles.

Tag `v0.3.0` flake verdict (T06, verified 2026-09-02 via `git worktree` of the
tag): `nix flake check` at `60cf5b1` FAILS — the datastartest go-modules FOD
reports `specified: nkJghgIG… / got: 2o8l28pR…`, while the root FOD builds
clean. Exactly the mechanism above predicts: the datastartest hash had been
harvested before late source/requires changes; the root hash was current.

### Durable-mitigation decision (2026-09-03: corrected by the release gate)

The first decision draft said "accept the dance". The v0.4.0 release gate
DISPROVED it: with flake.nix inside the FOD's src fileset, the vendorHash
constant sits inside its own FOD input, so pasting a freshly measured hash
changes flake.nix, which changes the vendor output, which invalidates the
hash — an unsolvable self-reference (verified: two paste-and-rebuild
iterations, `PV2V+… → Gsz4kq…`, never converge). **The dance is unwinnable
for `datastartestVendorHash` as long as metadata files enter the vendor
tree.**

The durable fix (implemented in flake.nix, verified converging on the first
paste): give the datastartest hermetic check a MINIMAL src fileset —
datastartest itself, the replaced modules' package sources (root `*.go` +
root `go.mod`, `static/`) — so root-level metadata (flake.nix, *.md, CI
config) never enters `go mod vendor`'s output. The FOD then only moves on
genuine module-set changes (requires/toolchain), like the root hash, and the
paste-once dance converges.

Rejected alternatives, for the record: dropping the replaces inside the
hermetic check would test the PUBLISHED library instead of the working tree
(false-green class worse than a loud mismatch); committing a vendor/ tree
would duplicate root source (split-brain); go.mod/go.sum-only filesets are
impossible because `go mod vendor` must read the replaced directories.

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
