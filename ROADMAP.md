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
- Domain-adapter example (EventBridge-style) demonstrating the
  Patch-as-value payoff
- datastartest helper-API expansion (consolidated from the 2026-08-10
  reports' ~40 micro-items, e.g. 2026-08-10_07-27 #34–47): RequireElementsOrdered,
  RequireNotScript, FindAllElements, FindScript, EventToSelectorMap,
  ReadAllEvents, Diff, Snapshot, ServeSSE, NewRecorder, RawSSE, Event.LogJSON,
  GoString, fluent Assert API, Ginkgo/Gomega matchers, JSON-aware
  SignalsContain, timeout variants of CollectWithRequest/CollectPost; plus
  internal polish (accessor methods over the public ID/Retry fields,
  tag-attribute parsing beyond quotes, indexTagEnd rename, table-driven
  benchmark shapes). First three shipped 2026-09-03: RequireElementsOrdered,
  Diff, Snapshot (see CHANGELOG [Unreleased]).
- `datastartest.NewResponse`-style helper for test ergonomics
- Response ergonomics: `ErrorResponse`/`NotificationResponse`/
  `ErrorResponseFromError` as `Response` methods; a `signalsMap` type for the
  signals-patch pattern; review `signalKeyMessage` naming
- Compile-checked doc snippets (a tiny docspec test target) so guide examples
  like `BroadcastMany` cannot drift from the API
- `example/README.md` and an `example/docker-compose.yml` for easy local
  runs; benchmark for `Collect` helper overhead
- Community metadata: GitHub Sponsors / funding, contributor list

### 3. CI/CD & Hermeticity

Make quality gates hermetic and reproducible.

Raw ideas:

- Route all lint/audit tools through nix checks so `nix flake check` is the
  single canonical quality gate (golangci-lint, erraudit, govulncheck)
- Nix CI job (cachix/install-nix-action) running `nix flake check` so
  hermetic build regressions surface before merge, not after
- Hermetic `checks.lint` / `checks.vet` / `checks.govulncheck` derivations;
  `flake.nix` `apps.bench` for running benchmarks
- vendorHash fragility under the `gitTracked` fileset — RESOLVED 2026-09-03:
  mechanism verified and corrected in ADR 004 (`go mod vendor` copies replaced
  directories entirely; root hash requires/toolchain-only; datastartest check
  now builds from a minimal fileset so the hash converges)
- Verify the erraudit probe-gate transition once the repo goes public
  (manual trigger or scheduled probe)
- Scheduled fuzz runs in CI (`go test -fuzz` on a cron, corpus committed);
  CodeQL workflow for Go security analysis
- CI matrix / parallel jobs per module for faster feedback
- `go work vendor` support or a flake app for offline module graphs
- `example/` as its own Go module (structural decision — keeps demo deps out
  of the root module graph)
- Coverage-floor policy decision (optional CI gate at a threshold)
- Release automation (goreleaser, changelog-from-release, tag-triggered
  GitHub releases)
- Build-time version variable or `version` package

### 4. Documentation Depth

Move beyond API reference into conceptual and operational docs.

Raw ideas:

- `docs/error-system.md` deep-dive: the full contract, decision rationale,
  why `--enforce-samber-oops` must NOT be used
- ADRs: 003 error classification, 004 nix per-module hermetic checks,
  005 coverage strategy (what the % includes)
- Consumer guides: `docs/replay.md` (EventStore + LastEventID),
  `docs/wire-format.md` (annotated dataline examples),
  `docs/testing.md` (unit/E2E/fuzz/WPT strategy), `docs/performance.md`,
  `docs/migration-guide.md` (for the next minor bump)
- `docs/architecture.md` overview diagram (transport → protocol → domain)
- Website launch (Astro + Starlight pattern)
- Document the DataStar JS version pinning strategy and upgrade process
- SSE heartbeat documentation
- Constraint check tying `static.Version` mentions to the CHANGELOG (pinning
  hygiene)

### 5. Upstream Protocol Tracking

Stay current with the DataStar protocol as it evolves upstream.

Raw ideas:

- Subscribe to upstream `starfederation/datastar` for protocol changes
- Renovate rule for upstream DataStar JS releases
- Protocol version negotiation if DataStar introduces breaking wire changes
- Compat-test matrix: go-datastar × go-sse version combinations (catch
  transport regressions before consumers do)
- Watch go-sse for a Stream-level OnDrop (would reopen the Response
  drop-observability question; through v0.6.0 OnDrop stays Broadcaster-only)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **No CQRS, event bus, or domain opinions:** This is a pure protocol layer.
  Consumers build domain adapters on top (e.g., cqrs-htmx/datastar's
  EventBridge). Mixing domain logic in would violate the separation.
- **No opinionated session/state management:** The library produces `sse.Event`
  values; it does not manage user sessions, authentication, or application state.
- **No bundling beyond DataStar JS:** The embedded client is the DataStar SDK
  only. No CSS frameworks, no JS runtimes, no opinionated frontend stack.

## Resolved questions

- **`go.work.sum` tracking (decided 2026-08-16, Full Execution Mode; refined
  2026-09-03 per CodeRabbit PR #3 thread):**
  intentionally gitignored. `go.work` is force-added for workspace development;
  `go.work.sum` is regenerated by the go toolchain on demand and the committed
  per-module `go.sum` files are the source of truth for reproducibility. The
  replace directives make sibling-module checksums unnecessary (they resolve to
  local paths); workspace-mode hashes for external modules accumulate in
  `go.work.sum` without a reproducibility contract — consumers never see it, so
  committing it would only add diff noise on every dependency update.
  Documented in AGENTS.md.
- **`v0.0.0` vs real versions for sibling requires (decided 2026-08-16):**
  sibling requires use real published versions (not `v0.0.0`). The replace
  directives make versions irrelevant locally, but a consumer testing without
  replaces must resolve to a real published module. `go mod tidy` already
  emits the correct published versions (e.g., v0.2.0). Documented in AGENTS.md.
- **`go` directive policy (decided 2026-08-16, updated 2026-08-29):**
  directives pin the exact patch release — now `go 1.26.7` across go.mod ×3,
  go.work, CI `go-version`, and the flake pin, clearing the four 1.26.6-era
  stdlib CVEs (GO-2026-5972/6089/6090/6218) and superseding the v0.0.2/v0.0.3
  "lowered to `go 1.26`" CHANGELOG ghost. Nix stays hermetic through a
  `go_1_26.overrideAttrs` pin (marked TODO for removal) until nixpkgs ships
  ≥ 1.26.7.
