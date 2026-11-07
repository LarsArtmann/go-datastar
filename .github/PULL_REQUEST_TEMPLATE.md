## Summary

Brief description of what this PR changes and why.

## Type of change

- [ ] Bug fix (non-breaking)
- [ ] New feature (non-breaking)
- [ ] Breaking change
- [ ] Documentation
- [ ] Refactor / cleanup

## Checklist

Run these locally before opening the PR — tick only what you actually ran:

- [ ] `GOWORK=off go test ./... -race -count=1` passes
- [ ] Per-module isolation passes from `broadcast/`, `datastartest/`, and `static/`:
      `GOWORK=off go test ./... -count=1` in each
- [ ] `nix run .#lint-ci` passes (exact CI parity, 0 issues)
- [ ] `go vet ./...` passes
- [ ] CHANGELOG updated (if user-visible change)
- [ ] Tests added for new functionality

CI re-verifies build, vet, race tests, golangci-lint, govulncheck,
actionlint, and erraudit on every code change — those boxes no longer exist
because agents kept pre-checking them without running anything.
