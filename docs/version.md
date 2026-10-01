# Versions

Three different "versions" live in this repo. They are independent and move
on different schedules:

| What                  | Where                                    | How it changes                                                                 |
| --------------------- | ---------------------------------------- | ------------------------------------------------------------------------------ |
| Module versions       | git tags (`v0.x.y`, lockstep ×4 modules) | Release process — see [release-checklist.md](release-checklist.md)              |
| DataStar JS client    | `static.Version` (currently v1.0.3)     | Upstream DataStar release — see [static-js.md](static-js.md)                    |
| Example binary        | `version.Version` (default `"dev"`)     | Build-time `-ldflags` injection                                                |

## Module versions

The four Go modules (`.`, `broadcast/`, `datastartest/`, `static/`) are
tagged in lockstep. Consumers see versions through the Go module system:

```bash
go list -m -versions github.com/larsartmann/go-datastar
```

## The JS client version

`datastar.Version()` returns the version of the embedded DataStar JavaScript
client (`static.Version`). It moves when the vendored bundle is refreshed —
independently of module versions. The bundle itself is checksum-pinned by
`static/checksum_test.go`.

## The example binary version

`version.Version` demonstrates the ldflags-injection pattern for consumer
binaries. Plain builds report `dev`; release builds stamp it:

```bash
go build -ldflags "-X github.com/larsartmann/go-datastar/version.Version=v0.6.1" ./example
```

Library consumers do not need this package — module versions come from the
module system, not a string.
