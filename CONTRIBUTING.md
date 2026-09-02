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
nix run .#lint                 # run golangci-lint (nixpkgs version)
nix run .#lint-ci              # EXACT CI linter parity — THE pre-push gate
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

# Lint (dev-shell / nixpkgs version):
golangci-lint run ./... ./datastartest/... ./static/...

# Pre-push, ALWAYS use the exact-CI lint (same version CI installs):
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 \
  run ./... ./datastartest/... ./static/... --timeout 5m
# (or: nix run .#lint-ci)

# Doc-snippet compile check (guides under docs/ + example/README.md):
nix run .#docspec
# (or: go test -tags docspec -run TestDocspec ./... ./datastartest/...)

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

## Fuzzing

Four fuzz targets guard the parsers and serializers. Their seed corpora are
committed, so every regular `go test` run replays them as regression cases;
`-fuzz` explores beyond the seeds.

| Target                        | Module         | What it shakes out                                                                     |
| ----------------------------- | -------------- | -------------------------------------------------------------------------------------- |
| `FuzzReadSignals`             | root           | `ReadSignals` request-body parsing (malformed JSON, closed bodies)                     |
| `FuzzMarshalSignalsRoundtrip` | root           | signals marshal → unmarshal roundtrip stability                                        |
| `FuzzReadEvents`              | `datastartest` | SSE wire-format parser conformance (51-seed corpus in `testdata/fuzz/FuzzReadEvents/`) |
| `FuzzUnmarshalSignals`        | `datastartest` | dataline signals decoding                                                              |

Quick smoke (30 seconds, per module — fuzzing runs one target at a time):

```bash
GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadSignals$' -fuzztime 30s .
(cd datastartest && GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz '^FuzzReadEvents$' -fuzztime 30s .)
```

On a crash, Go writes the failing input to `testdata/fuzz/<FuzzName>/` as a
new seed — commit it so the regression is replayed by every future `go test`
run. The seeds are intentionally portable across checkouts; do not gitignore
them.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.

## Doc snippets must compile (docspec)

Guide snippets under `docs/` and in `example/README.md` are mirrored as
compile-checked functions in `docspec_test.go` (root) and
`datastartest/docspec_test.go`, behind the `docspec` build tag so they stay
out of the default test run. The contract:

- When you change a public API that a guide snippet uses, update the snippet
  AND its mirrored function in the same commit.
- Run `nix run .#docspec` (or `go test -tags docspec -run TestDocspec
  ./... ./datastartest/...`) before pushing doc or API changes.
- New guide snippet? Add a mirrored function — compile-only is the floor;
  executing the cheap path is better (semantic drift fails the run too).
