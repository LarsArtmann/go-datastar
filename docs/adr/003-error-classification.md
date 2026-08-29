# ADR 003: Error classification via go-error-family, never samber/oops

Date: 2026-08-29

## Context

Every error go-datastar returns crosses a library boundary: server frameworks,
middleware, and application code all receive it. Three failure-handling styles
competed:

1. **Bare errors** (`fmt.Errorf` + sentinels) — callers string-match messages
   or assert on concrete types; both break silently across refactors.
2. **Enrichment stacks** (samber/oops, pkg/errors) — great ergonomics, but the
   library would dictate the application's observability stack and drag a
   heavy dependency into every consumer.
3. **Classification** (go-error-family) — a small structured contract: a
   stable machine-readable **code**, a behavioral **family** (is this
   retryable? whose fault is it?), and typed **context**. No opinions about
   tracing, logging, or telemetry.

The dependency set also matters: go-sse (the transport layer beneath this
library) classifies its errors with go-error-family since v0.5.0, so the
contract already exists one layer down.

## Decision

- **Every error returned by go-datastar is a classified `*errorfamily.Error`**
  carrying a stable code from the catalog in `errors.go` (e.g.
  `datastar.signals_unmarshal_failed`), a family (Rejection / Transient /
  Orchestration), and structured context (HTTP method, input byte length,
  value type — whatever diagnosed the failure without a re-run).
- **The public API returns `error`, not `*errorfamily.Error`.** Idiomatic Go,
  consistent with go-sse. Typed access goes through `errorfamily.Code(err)`,
  `errors.Is(err, datastar.Err...)`, `errorfamily.IsRetryable(err)`.
  erraudit's `generic_return` warnings on these signatures are accepted by
  design (routed 2026-08-08, re-affirmed here).
- **Libraries classify; they never enrich.** samber/oops (and every other
  enrichment stack) is forbidden in this module. Applications are free to wrap
  the classified errors with oops at their own boundary.
- **Sentinels stay context-pristine.** `WithContext` returns a clone; shared
  sentinels never accumulate caller-specific context.
- **Layering with go-sse:** `wrapStreamError` re-tags go-sse Send failures as
  `datastar.stream_send_failed` (Transient). `errorfamily.Classify` returns
  the outermost family, so go-datastar's classification wins; `errors.Is`
  still traverses the chain, so callers matching go-sse codes keep working.

| Family        | Assigned to                                                       | Retryable |
| ------------- | ----------------------------------------------------------------- | --------- |
| Rejection     | Malformed JSON, empty names, invalid mode/namespace, closed body  | no        |
| Transient     | Temporary I/O failure reading request bodies; stream Send failure | yes       |
| Orchestration | Internal render failure (templ, gostar)                           | no        |

## Consequences

- Consumers get three stable matching dimensions (code / sentinel / family)
  that survive refactors; no message string-matching.
- The library stays dependency-light and observability-agnostic.
- When Go 1.26's `errors.AsType` is preferred over `errors.As` for type
  extraction, migrations are mechanical; sentinel matches via `errors.Is`
  remain the documented pattern for value matching (see
  `errors_example_test.go` for all three handling patterns).
- If go-error-family ever grows enrichment ambitions, this library pins to the
  classification subset only.
