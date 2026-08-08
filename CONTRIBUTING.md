# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Prerequisites

This project requires **Go 1.26+** with the **`GOEXPERIMENT=jsonv2`** flag enabled
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

The Nix dev shell automatically sets `GOEXPERIMENT=jsonv2`, `GOWORK=off`, and
`GOTOOLCHAIN=local`.

### Manual setup (no Nix)

```bash
# All Go commands require these environment variables:
export GOEXPERIMENT=jsonv2
export GOWORK=off

# Run the full test suite:
go test ./... -race -count=1

# Lint:
golangci-lint run ./...

# Vet:
go vet ./...
```

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
