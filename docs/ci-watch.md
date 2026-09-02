# CI Watch Runbook

Four workflows run non-required (nothing is a required check — branch
protection is removed by owner decision). Unwatched, they rot: this is the
promote / drop / verify checklist for each. Review monthly or after any
workflow change.

## nix.yml — hermetic `nix flake check`

State: `continue-on-error` until proven stable (first green run 2026-09-03).

- **Promote** (drop `continue-on-error`, make the check blocking in spirit —
  it still cannot gate without branch protection, so "promote" = treat red as
  a master-breaking incident): after 2 consecutive green weeks on master AND
  at least one release tagged from a tree where it passed.
- **Verify:** `nix flake check` locally at the same commit; a red run with a
  green local run usually means a nixpkgs drift (new `nixpkgs-unstable` rev
  changed a derivation input) — pin or rebase the flake.lock.
- **Drop** (remove the workflow): if it stays red for a month with no
  hermeticity value — but per [ADR 004](adr/004-hermetic-checks.md) the flake
  IS the canonical gate, so dropping CI's copy means owning the check locally.

## fuzz.yml — scheduled 60s fuzz runs

State: daily cron over all four fuzz targets, crash artifacts uploaded.

- **Verify after first runs:** check the artifacts — a crash artifact is a
  REAL bug: reproduce it with the committed corpus + `go test -run '^$' -fuzz
  '<Target>' -fuzztime 30s`, fix, and commit the failing input as a corpus
  regression seed (see CONTRIBUTING "Fuzzing").
- **Promote:** extend `-fuzztime` (60s → 300s) once runs are stable and no
  crashes for two weeks.
- **Drop:** only if fuzzing moves to a required local gate (it shouldn't —
  scheduled runs cover what local smoke runs can't).

## codeql.yml — Go security analysis

State: SHA-pinned, default query suite.

- **Verify after first runs:** triage every alert. A true positive is
  master-breaking: fix immediately, note the fix in the CHANGELOG. A false
  positive gets a `// codeql[...]` suppression inline with a reason comment —
  never a blanket config exclusion without a written rationale.
- **Promote:** none needed (CodeQL is already advisory-complete); revisit the
  query suite (extended) after the first clean month.

## renovate.json — embedded-JS-client bumps

State: custom manager proposing `static/static.go` Version bumps from
upstream DataStar releases; coexists with dependabot.yml (one-bot decision
pending — see TODO_LIST).

- **Verify each PR:** the proposed version must match a real upstream release
  tag; the bundle change must be re-verified against the wire-format goldens
  (`go test ./... -run TestPatchWireGoldens`) because the JS client and the
  server protocol must stay in lockstep; run the E2E suite.
- **Promote:** label/schedule tuning after the first successful merge;
  replace dependabot.yml once the one-bot decision lands (owner-blocked).
- **Drop:** if the one-bot decision keeps Renovate — delete the OTHER bot
  instead. Never drop both.

## General rules

- A red non-required workflow is treated as red master: fix or consciously
  accept with a written reason in the status report.
- Workflow changes always run `actionlint` locally (it validates on every
  push anyway) and, for ci.yml changes, the full local gate from AGENTS.md.
- Any workflow added in the future starts here: state, verify checklist,
  promote criteria, drop criteria — or it will rot.
