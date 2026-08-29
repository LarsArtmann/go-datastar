# ADR 005: Coverage strategy — measured, badged, not gated

Date: 2026-08-29

## Context

Coverage is reported per module with different shapes:

| Module       | Coverage (2026-08-29) | Shape                                                     |
| ------------ | --------------------- | --------------------------------------------------------- |
| root         | 98.4%                 | Protocol code; remaining % is defensive I/O paths         |
| datastartest | 92.7%                 | Test helpers; uncovered % is assertion-failure formatting |
| static       | n/a                   | Zero-stdlib-plus-embed bundle; no logic to cover          |
| example      | excluded              | Demo app; deliberately not a coverage target              |

The `coverage.yml` workflow publishes a live badge (shields.io endpoint JSON
on the orphan `coverage` branch) computed from all three modules, with
threshold-based color. The question: should coverage also be a **hard CI
gate** (fail below N%)?

## Decision

- **Measured and badged, not gated.** No minimum-coverage check fails CI.
- The badge colors encode a **soft floor**: ≥ 90% bright green, ≥ 75% yellow,
  ≥ 50% orange, below red — a regression is visible at a glance on the README
  without manufacturing failures.
- Coverage claims in living docs must be **measured, not asserted** (the
  docs-health rule): state the number and the measurement date.
- The hard gate for behavior remains the test suite itself (race across all
  modules) — not a coverage proxy.

## Consequences

- No coverage theater: a hard gate incentivizes meaningless tests to cross a
  threshold; this repo's most valuable suites (wire-format parity, WPT
  conformance, fuzz corpora) exist because behavior demanded them, not because
  a number did.
- A silent coverage collapse is still caught socially: the badge changes
  color, and the docs-health VERIFY pass re-measures on every audit.
- If a future maintainer wants a hard floor, the thresholds in
  `coverage.yml` are the natural place to enforce them.
