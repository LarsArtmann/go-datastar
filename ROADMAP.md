# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to TODO_LIST.md.

## Themes

### 1. Error System Maturity

The typed error system (go-error-family classification) is functional but
has room to grow in ergonomics and depth.

Raw ideas:

- Evaluate whether returning `*errorfamily.Error` (or domain-specific wrappers
  like `RenderError`, `SignalsError`) instead of bare `error` gives consumers
  enough value to justify the coupling
- Consider a typed `type Code string` for compile-time safety instead of untyped
  string constants
- Explore `errorfamily.WrapOnce` at API boundaries to prevent double-classification
  when callers pre-classify errors
- Retry ergonomics: example or helper showing how to retry `Transient`
  (`CodeBodyReadFailed`) errors with backoff
- Snapshot-test error messages for stable wire output across versions

### 2. Developer Experience & Onboarding

Lower the barrier for new users discovering and adopting go-datastar.

Raw ideas:

- "Migrating from starfederation/datastar-go" guide — what changes, why patches
  as values matters
- Architecture diagram (D2 or mermaid) showing the three-layer architecture
  (go-sse → go-datastar → domain adapter)
- More example applications (toasts, progress bars, signal merge modes)
- Playground or example repo link for interactive exploration
- Comparison table vs upstream SDK in README
- `Broadcaster[datastar.Patch]` typed-filtering example
- `SubscribeFilter` usage example
- Headless-browser E2E test (chromedp or Playwright) exercising the real
  DataStar JS client — the current E2E stops at wire-format verification
- Typed script-patch accessors in `datastartest` (`RedirectURL`,
  `CustomEventName`/`CustomEventDetail`, `ScriptAttributes`) — structured
  extraction instead of `strings.Contains` on `ScriptContent()`
- Extract the generic SSE parser (`datastartest/reader.go`, `event.go`) into
  a reusable `ssetest` package, usable beyond DataStar

### 3. CI/CD & Hermeticity

Make quality gates hermetic and reproducible.

Raw ideas:

- Route all lint/audit tools through nix checks so `nix flake check` is the
  single canonical quality gate (golangci-lint, erraudit, govulncheck)
- Release automation (goreleaser, changelog-from-release, tag-triggered
  GitHub releases)
- Build-time version variable or `version` package
- Scheduled fuzz testing in CI (`go test -fuzz` on a cron)

### 4. Documentation Depth

Move beyond API reference into conceptual and operational docs.

Raw ideas:

- `docs/error-system.md` deep-dive: the full contract, decision rationale,
  why `--enforce-samber-oops` must NOT be used
- Website launch (Astro + Starlight pattern)
- Document the DataStar JS version pinning strategy and upgrade process
- SSE heartbeat documentation

### 5. Upstream Protocol Tracking

Stay current with the DataStar protocol as it evolves upstream.

Raw ideas:

- Subscribe to upstream `starfederation/datastar` for protocol changes
- Renovate rule for upstream DataStar JS releases
- Protocol version negotiation if DataStar introduces breaking wire changes

## Non-goals

Things we are deliberately NOT pursuing and why:

- **No CQRS, event bus, or domain opinions:** This is a pure protocol layer.
  Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's
  EventBridge). Mixing domain logic in would violate the separation.
- **No opinionated session/state management:** The library produces `sse.Event`
  values; it does not manage user sessions, authentication, or application state.
- **No bundling beyond DataStar JS:** The embedded client is the DataStar SDK
  only. No CSS frameworks, no JS runtimes, no opinionated frontend stack.

## Open questions

Decisions awaiting the owner (not tasks — do not action without a ruling):

- **`go.work.sum` tracking:** gitignored today while `go.work` is committed.
  Committing both strengthens checksum verification for workspace-local
  replaces at the cost of diff noise on every dependency update.
- **`v0.0.0` vs real versions for sibling requires:** the replace directives
  make the version irrelevant locally, but `go mod tidy` can emit
  pseudo-versions if replaces are ever removed.
- **`go` directive policy:** go.mod/go.work say `go 1.26.5`; the v0.0.2/v0.0.3
  CHANGELOG entries claim a lowering to `go 1.26` that never landed. Either
  apply the lowering or accept `1.26.5` and stop claiming otherwise.
