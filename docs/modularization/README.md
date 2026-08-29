# Modularization

The go-datastar project uses a multi-module Go workspace layout. This directory
holds the proposal, execution plan, and architecture decisions that shaped it.

## Documents

| Document                                                             | Type           | Description                                                                   |
| -------------------------------------------------------------------- | -------------- | ----------------------------------------------------------------------------- |
| [2026-08-10_PROPOSAL.html](2026-08-10_PROPOSAL.html)                 | Proposal       | Original case for splitting go-datastar into independently versioned modules (executed — shipped in v0.1.0) |
| [2026-08-10_EXECUTION_PLAN.html](2026-08-10_EXECUTION_PLAN.html)     | Execution plan | Step-by-step migration from single module to multi-module workspace (fully executed; outcome recorded in ADR 002) |
| [../adr/001-architecture.md](../adr/001-architecture.md)             | ADR 001        | Architecture: go-datastar, go-sse, and the DataStar SDK (layer separation)    |
| [../adr/002-multi-module-split.md](../adr/002-multi-module-split.md) | ADR 002        | Multi-module split: three modules, strict DAG, mutual replaces, lockstep tags |

## Module structure

| Module       | Path                                              | Purpose                            | Dependencies            |
| ------------ | ------------------------------------------------- | ---------------------------------- | ----------------------- |
| Root         | `github.com/larsartmann/go-datastar`              | Protocol library                   | go-sse, go-error-family |
| static       | `github.com/larsartmann/go-datastar/static`       | Embedded DataStar JS client bundle | zero (stdlib only)      |
| datastartest | `github.com/larsartmann/go-datastar/datastartest` | Consumer E2E test helpers          | go-datastar, go-sse     |

The dependency graph is a strict DAG: `static → root → datastartest`. All three
modules tag in lockstep. See ADR 002 for the full rules and consequences.

See also: [CONTRIBUTING.md](../../CONTRIBUTING.md) "Multi-Module Development" section
for workspace-vs-GOWORK=off commands and replace rules.
