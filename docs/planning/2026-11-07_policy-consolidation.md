# Policy Consolidation — Open Owner Questions, One Place

> **Status:** DECISION-READY compilation (2026-11-07). This doc consolidates
> every standing policy question scattered across TODO_LIST owner-blocked
> rows and archived reports into one list, so the owner can answer in one
> sitting. Answers fold back into TODO_LIST/ROADMAP and close the rows.
> The four v0.7.0 cut decisions live separately (and with evidence) in
> [the cut brief](2026-11-07_v070-cut-owner-decision-brief.md).

## The questions

| #  | Question                                                                                                                                                               | Where it came from                                                           | Recommendation                                                                                                                                                           |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Q1 | **Lane protocol for shared checkouts**: adopt a CLAIMED_LANES banner convention (each session's report names its lane up front) or keep accept-collision-and-converge? | 2026-11-07 repair session §d; three sessions duplicated a sweep the same day | Adopt the banner — one line per report, zero new tooling; it turned this session's ScriptTagVersioned overlap into a clean retraction                                    |
| Q2 | **CHANGELOG policy for infra/docs-only passes**: do CI/tooling/docs sessions get an entry, or does report-only stand?                                                  | repeated cut-prep ambiguity                                                  | Keep the current practice: user-visible changes get entries (the [0.6.3] CI-repairs entry shipped one), pure report passes don't — codify it here and stop re-litigating |
| Q3 | **Website trigger**: launch the Astro/Starlight docs site now, at v1.0, or on a demand signal?                                                                         | ROADMAP theme 4; deferred by owner (T27)                                     | Deferred stands; revisit at v1.0 or when a second consumer asks                                                                                                          |
| Q4 | **Status-index "Monitoring" tier** for never-fully-resolvable reports + per-item marker depth for consolidated ROADMAP ideas                                           | 2026-09-02 report owner questions                                            | Skip the tier; the ledger-marker convention (below) covers the same need with zero index churn                                                                           |
| Q5 | **Shared lint-cache policy**: purge the 3.8G `/mnt/buildcache/golangci-lint` or bless fresh-cache (`mktemp -d`) as the documented default for gates                    | 2026-09-18_22-29 g1; AGENTS.md ghost-cache gotcha                            | Bless fresh-cache for gates (checklist §1 already runs it); purge is an owner-only maintenance action                                                                    |
| Q6 | **Required-checks policy**: master has none (branch protection removed) — document "local gates are the gate" canonically, or add lint/nix alerting                    | 2026-09-18_20-45 g3, f26                                                     | Document the contract in one AGENTS.md line + keep `docs/ci-watch.md` as the human ritual; alerting only if a second red-master incident slips through the watch         |
| Q7 | **erraudit public flip**: when the repo goes public, the probe-gated CI erraudit job becomes a hard gate — verify the probe notices on the first push?                 | TODO_LIST row (owner-blocked)                                                | Keep the probe; add one ci-watch checklist line for the flip                                                                                                             |
| Q8 | **Should the auto-commit daemon run on `static/**`?** It committed the 2026-08 bundle mangle; shields catch reformatting but the daemon remains the transport          | buildflow ledger #37                                                         | Exclude `static/**` from the daemon's heuristic (config lives in the fleet daemon, owner-side change); the checksum test stays as the backstop                           |

## Appendix — routed-remainder marker spec (M14 prep, upstream candidate)

The docs-gates battery treats archived files as fully-resolved; two archive
classes still leak noise:

1. **Idea-list archives** whose rows are pointers or brainstorm residue —
   per-item verdict passes are possible but low-value for genuinely-open
   ideas.
2. **Ledger tables** already exempted via `<!-- check-rows:ledger -->` —
   exempt, but the _archive-level_ story ("this file's open rows are
   intentional") is only implicit.

**Proposed marker** (freshness skip class, like `fresh:skip`):

```markdown
<!-- routed-remainder: <one-line reason> -->
```

- Placement: directly above the affected table (or the report's §f heading).
- Effect: audit-all replay skips the marked region's items; check-rows
  treats bare rows in the marked table as LEDGER-EXEMPT (same semantics as
  the existing marker, but with a written reason).
- The reason string is the archive-manifest line — bulk-archive sweeps get
  their manifest for free.

**Files that would take the marker today:** none in this repo (the two
no-strike archives the 2026-11-07 sweep named received their per-item
passes instead, and the three backlog ledgers carry `check-rows:ledger`).
The spec is parked for the docs-gates repo (upstream), owner call on filing.
Related upstream candidates from the same lane: check-rows continuation-row
classification, link-rotation guard (index links resolve). All three are
specs-only; file them in Artmann-Technologies/docs-gates if wanted.
