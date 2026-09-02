// Package version exposes the build-time version of the go-datastar example
// binary (and any consumer binary that links this package and sets it via
// -ldflags). Library consumers do not need it: module versions come from the
// Go module system (go list -m).
package version

// Version is overridden at build time:
//
//	go build -ldflags "-X github.com/larsartmann/go-datastar/version.Version=v0.4.0" ./example
//
// The default is "dev" for plain `go build`/`go run`.
//
//nolint:gochecknoglobals // ldflags injection requires a package-level var
var Version = "dev"
