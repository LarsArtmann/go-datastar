# Release Checklist

Steps to cut a new go-datastar release. All four modules tag in lockstep
(root, `static`, `datastartest`, `broadcast`) — see
[ADR 002](../adr/002-multi-module-split.md).

Run everything from the repo root under the pinned toolchain
(`GOTOOLCHAIN=go1.27.1`). `GOEXPERIMENT=jsonv2` is no longer needed under the
1.27.1 floor — v0.6.1 un-gated `encoding/json/v2`.

> **Gates are stateful.** §2 edits go.mod files, which invalidates §1's nix
> vendor-hash check — missing that re-run caused both documented red-master
> incidents (v0.6.0's permanent red prep run; 2026-09-29's unnoticed stale
> hash). A §1 green is void the moment §2 runs: §2.5 re-runs the nix gate on
> the post-bump tree, before anything is tagged.

## 1. Pre-release verification

- [ ] `GOTOOLCHAIN=go1.27.1 go test ./... ./broadcast/... ./datastartest/... ./static/... -race -count=1` — all green
- [ ] `GOTOOLCHAIN=go1.27.1 go vet ./... ./broadcast/... ./datastartest/... ./static/...` — clean
- [ ] `GOTOOLCHAIN=go1.27.1 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./... ./broadcast/... ./datastartest/... ./static/... --timeout 5m` — 0 issues (version pinned = CI parity)
- [ ] Fresh-cache lint — the shared `GOLANGCI_LINT_CACHE` can replay ghost findings from other sessions' deleted worktrees: re-run the same command with `GOLANGCI_LINT_CACHE=$(mktemp -d)` and trust only that verdict — 0 issues. _(owner-policy pending: decide whether this replaces or supplements the shared-cache run)_
- [ ] erraudit loop with CI flags, all four modules — "No violations found" ×4:
      `for mod in . ./broadcast ./datastartest ./static; do (cd "$mod" && GOEXPERIMENT=jsonv2 erraudit . --type-aware --enforce-go-error-family --severity-threshold error); done`
      _(owner-policy pending: the CI erraudit leg is probe-gated off while the repo is private; this row is the substitute — the `--no-suppress` audit-mode variant in AGENTS.md intentionally lists a few documented-tolerated patterns on top)_
- [ ] `nix flake check` — all checks passed
- [ ] `go work sync` — go.work unchanged (idempotency)
- [ ] Per-module isolation (`GOWORK=off`) build + test for all 4 modules
- [ ] Per-module `GOWORK=off go mod tidy -diff` — prints nothing (×4)
- [ ] `grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod broadcast/go.mod` — finds nothing (relative paths only)
- [ ] `nix run .#govulncheck` — No vulnerabilities found

## 2. CHANGELOG and version

- [ ] Review `CHANGELOG.md` `[Unreleased]` section; promote to a versioned heading
- [ ] Bump the sibling requires with `go mod edit -require` (there is no
      `-version` flag; `static` has no dependencies to bump):
      - root: `go mod edit -require=github.com/larsartmann/go-datastar/static@vX.Y.Z`
      - broadcast: `go mod edit -require=github.com/larsartmann/go-datastar@vX.Y.Z`
      - datastartest: `go mod edit -require=github.com/larsartmann/go-datastar@vX.Y.Z -require=github.com/larsartmann/go-datastar/static@vX.Y.Z`
- [ ] Verify lockstep: the new version appears in every sibling require and nothing else moved — `grep -n 'go-datastar' go.mod broadcast/go.mod datastartest/go.mod`

## 2.5. Post-bump nix gate (mandatory — never skip to §3)

- [ ] `nix flake check --keep-going` — `--keep-going` collects ALL moved vendor
      hashes; plain `flake check` (and CI) aborts on the first mismatch, so the
      error line is never the complete list
- [ ] Paste EVERY moved `vendorHash` (root / broadcast / datastartest) into
      `flake.nix` — one pass, same commit as the require bumps
- [ ] Re-run `nix flake check` — all checks passed on the post-bump tree

## 3. Tag and push

- [ ] Create four **annotated** tags on the SAME commit:
      `git tag -a vX.Y.Z -m "Release vX.Y.Z"` ·
      `git tag -a static/vX.Y.Z -m "Release static vX.Y.Z"` ·
      `git tag -a datastartest/vX.Y.Z -m "Release datastartest vX.Y.Z"` ·
      `git tag -a broadcast/vX.Y.Z -m "Release broadcast vX.Y.Z"`
- [ ] Tag hygiene: `git for-each-ref refs/tags --format='%(refname:short) %(objecttype)' | grep -v ' tag$'` — prints nothing (every tag annotated); all four `rev-parse '<tag>^{commit}'` values are identical; the tagged tree carries the new go.mod versions and the versioned CHANGELOG heading
- [ ] Push the exact ref list — never `git push --tags` (it would push any stray local tag):
      `git push origin vX.Y.Z static/vX.Y.Z datastartest/vX.Y.Z broadcast/vX.Y.Z`
- [ ] Watch CI: `gh run watch --exit-status`

## 4. Post-release verification

- [ ] Proxy state per module: `curl https://proxy.golang.org/github.com/larsartmann/go-datastar/@v/vX.Y.Z.info` (repeat with `/broadcast/`, `/static/`, `/datastartest/` before `/@v/`) — `Version` is the new tag and `Origin.Hash` is the tagged commit
- [ ] `go list -m -versions github.com/larsartmann/go-datastar` (and the three submodule paths) — new version listed ×4
- [ ] Consumer smoke from a clean cache: fresh directory, `GOMODCACHE=$(mktemp -d) go mod init smoke && GOMODCACHE=$(mktemp -d) go get` all four modules `@vX.Y.Z` (sum-DB verified), then build and run. Check the API with `go doc` BEFORE writing the program — guessing shapes cost a failed compile during the v0.6.0 release
- [ ] `pkg.go.dev` renders all 4 modules:
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar`
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar/broadcast`
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar/static`
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar/datastartest`
- [ ] GitHub Releases ×4 with real CHANGELOG excerpts (root = full notes; submodule releases = accurate short notes). Create the **root** release LAST — GitHub marks the newest-created release "Latest", and v0.6.0 accidentally landed it on `datastartest/v0.6.0`. Recover with `gh release edit vX.Y.Z --latest` if it lands wrong

## 5. Comparison re-verify (quarterly or after upstream release)

- [ ] `go list -m -versions github.com/starfederation/datastar-go` — check for new releases
- [ ] If upstream released a new version, re-verify every row in the README comparison table against the new `pkg.go.dev` docs
- [ ] Update the pinned footnote: `_Compared against [datastar-go vX.Y.Z]..._`
- [ ] Verify the embedded JS client version (`static/static.go:Version`) matches the latest DataStar client release
