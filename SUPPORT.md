# Support

Thanks for using go-datastar!

## Questions and help

- **GitHub Discussions** (recommended): open a Q&A discussion at
  <https://github.com/LarsArtmann/go-datastar/discussions> — questions,
  integration help, and show-and-tell all live there.
- **Bug reports**: open a GitHub Issue with a minimal reproduction (a handler
  snippet plus the observed vs. expected SSE output). The `datastartest`
  helpers can capture both in a few lines — see
  [docs/testing.md](docs/testing.md).

## Before opening an issue

1. Check the [README](README.md) quick start and the guides under `docs/`
   (replay, error-system, wire-format, testing, migration).
2. Search existing issues and discussions — wire-format questions in
   particular are usually already answered in
   [docs/wire-format.md](docs/wire-format.md).
3. Confirm your environment: Go 1.26.7+, `GOEXPERIMENT=jsonv2` (required),
   and the version(s) of go-datastar you depend on.

## What to include in a bug report

- go-datastar version (`go list -m github.com/larsartmann/go-datastar`)
- Go version (`go version`) and the exact command you ran
- A minimal handler + test reproducing the behavior
- For wire-format surprises: the raw SSE bytes (curl -N output)

## Response expectations

This is a single-maintainer OSS project. There is no SLA; triage happens in
batch. Urgent security matters: see [SECURITY.md](SECURITY.md).
