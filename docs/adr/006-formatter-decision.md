# ADR 006: treefmt is canonical; dprint documents non-Go intent

Date: 2026-08-29

## Context

Two formatter configurations coexist in the repo:

- **treefmt-nix** (via `flake.nix`): gofumpt, goimports, golines, nixfmt.
  Enforced hermetically by `nix flake check` (`checks.format`, with `goPkg`
  prepended to `buildInputs` so goimports' per-file `go env` call never
  triggers a sandboxed toolchain download).
- **`dprint.json`** (repo root): rules for JSON, YAML, Markdown, and
  Dockerfile — file types treefmt does not cover here.

Wiring dprint into the flake was evaluated and rejected: dprint's plugins are
network-fetched WASM artifacts, which would break the hermetic Nix build
(no network in the sandbox) or force a large, fragile pinning effort.

## Decision

- **treefmt is the single canonical, enforced formatter.** `nix flake check`
  is the formatting gate; commit only treefmt-formatted Go and Nix code.
- **`dprint.json` stays in the repo** as a declaration of intent for non-Go
  files and a ready-made editor/external integration (dprint CLI, IDE
  plugins). It is deliberately NOT wired into treefmt or CI.
- Markdown/YAML/JSON formatting is therefore best-effort by humans and
  agents, not machine-gated. Reviewers may ask for `dprint fmt` on
  heavily-edited docs, but CI will not fail on it.

## Consequences

- Formatting enforcement stays hermetic and fast.
- `dprint.json` must not drift from reality: if a file type it covers becomes
  contentious, revisit this ADR (either invest in pinning dprint's WASM
  plugins hermetically or remove the file to avoid a false promise).
