# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Prerequisites

This project requires **Go 1.26.6+** with the **`GOEXPERIMENT=jsonv2`** flag enabled
(transitively via `go-branded-id` through `go-sse`).

If you use [Nix](https://nixos.org/), all of this is handled automatically by the
dev shell — see the Nix workflow below.

## Development Setup

### Quick start (Nix)

```bash
nix develop                    # enter dev shell (Go, golangci-lint, govulncheck, templ, gopls)
nix flake check                # run all checks (format + build)
nix run .#test                 # run tests
nix run .#test-race            # run tests with -race
nix run .#lint                 # run golangci-lint
nix run .#coverage             # generate coverage report
nix run .#build                # build the library
nix run .#vet                  # run go vet
```

The Nix dev shell automatically sets `GOEXPERIMENT=jsonv2` and
`GOTOOLCHAIN=local`, and pins the exact Go patch release.

### Manual setup (no Nix)

```bash
# All Go commands require this environment variable:
export GOEXPERIMENT=jsonv2

# Run the full test suite (workspace mode — see next section):
go test ./... ./datastartest/... ./static/... -race -count=1

# Lint:
golangci-lint run ./... ./datastartest/... ./static/...

# Vet:
go vet ./... ./datastartest/... ./static/...
```

## Multi-Module Development

This repository is a Go workspace with **three modules** (rationale and rules:
[ADR 002](docs/adr/002-multi-module-split.md)):

| Module          | Import path                                       | Role                        |
| --------------- | ------------------------------------------------- | --------------------------- |
| root            | `github.com/larsartmann/go-datastar`              | protocol library            |
| `static/`       | `github.com/larsartmann/go-datastar/static`       | embedded DataStar JS bundle |
| `datastartest/` | `github.com/larsartmann/go-datastar/datastartest` | consumer E2E test helpers   |

### Workspace mode (default) vs per-module isolation (`GOWORK=off`)

Day-to-day development uses the workspace (`go.work`) so all three modules
resolve each other locally:

```bash
GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1
```

Before committing, also verify each module standalone — this is what CI does,
and it is what proves the sibling `replace` directives are complete:

```bash
for mod in . ./datastartest ./static; do
  (cd "$mod" && GOWORK=off GOEXPERIMENT=jsonv2 go build ./... && GOWORK=off go test ./... -count=1)
done
```

### Replace directives: the rules

1. Siblings resolve locally through relative replaces (root:
   `static => ./static`; datastartest: `go-datastar => ..`,
   `static => ../static`). **Never use absolute paths** — CI fails the build
   on `replace.*=>/` because they break every other checkout.
2. `go work sync` must be idempotent: run it, then confirm
   `git diff --exit-code go.work`.
3. Root must never require `datastartest` — test helpers are not production
   dependencies. `module_boundary_test.go` enforces this at test time.

### Releasing and per-module tags

All three modules version in lockstep from one commit: tag `vX.Y.Z` (root),
`static/vX.Y.Z`, and `datastartest/vX.Y.Z` together. Consumers see real
versions in `require` blocks; the replaces are local-only conveniences and
never ship.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
