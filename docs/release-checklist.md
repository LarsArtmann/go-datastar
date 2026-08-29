# Release Checklist

Steps to cut a new go-datastar release. All three modules tag in lockstep
(root, `static`, `datastartest`) — see [ADR 002](../adr/002-multi-module-split.md).

## 1. Pre-release verification

- [ ] `GOTOOLCHAIN=go1.26.7 GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1` — all green
- [ ] `GOTOOLCHAIN=go1.26.7 GOEXPERIMENT=jsonv2 go vet ./... ./datastartest/... ./static/...` — clean
- [ ] `GOTOOLCHAIN=go1.26.7 GOEXPERIMENT=jsonv2 golangci-lint run ./... ./datastartest/... ./static/...` — 0 issues
- [ ] `nix flake check` — all checks passed
- [ ] `GOTOOLCHAIN=go1.26.7 GOEXPERIMENT=jsonv2 go work sync` — go.work unchanged (idempotency)
- [ ] Per-module isolation (`GOWORK=off`) build + test for all 3 modules
- [ ] `grep -rn 'replace.*=>/' go.mod datastartest/go.mod static/go.mod` — finds nothing (relative paths only)
- [ ] `nix run .#govulncheck` — No vulnerabilities found

## 2. CHANGELOG and version

- [ ] Review `CHANGELOG.md` `[Unreleased]` section; promote to a versioned heading
- [ ] Bump versions in all 3 `go.mod` files via `go mod edit -version` (never sed)
- [ ] Verify all 3 modules have the same version (lockstep)

## 3. Tag and push

- [ ] `git tag -a vX.Y.Z -m "Release vX.Y.Z"` on the root module
- [ ] `git tag -a static/vX.Y.Z -m "Release static vX.Y.Z"`
- [ ] `git tag -a datastartest/vX.Y.Z -m "Release datastartest vX.Y.Z"`
- [ ] `git push --tags` (or `--force-with-lease` only if re-tagging)
- [ ] Watch CI: `gh run watch --exit-status`

## 4. Post-release verification

- [ ] `pkg.go.dev` renders all 3 modules:
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar`
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar/static`
      - `https://pkg.go.dev/github.com/larsartmann/go-datastar/datastartest`
- [ ] `go get github.com/larsartmann/go-datastar@latest` works from a clean module cache
- [ ] Create GitHub Release from the tag with CHANGELOG excerpt

## 5. Comparison re-verify (quarterly or after upstream release)

- [ ] `go list -m -versions github.com/starfederation/datastar-go` — check for new releases
- [ ] If upstream released a new version, re-verify every row in the README comparison table against the new `pkg.go.dev` docs
- [ ] Update the pinned footnote: `_Compared against [datastar-go vX.Y.Z]..._`
- [ ] Verify the embedded JS client version (`static/static.go:Version`) matches the latest DataStar client release
