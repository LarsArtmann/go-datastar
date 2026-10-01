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

- More example applications (toasts, progress bars, signal merge modes)
- Playground or example repo link for interactive exploration
- Comparison table vs upstream SDK in README
- `broadcast.Broadcaster` + `SubscribeFilter` usage examples (the fan-out
  module moved in from cqrs-htmx/datastar 2026-09-17 — surface its
  hub-sharing and replay recipes from this repo's docs)
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
  benchmark shapes). Shipped so far: tranche 1 (RequireElementsOrdered,
  Diff, Snapshot) in v0.5.0 and tranche 2 (RequireNotScript, FindScript,
  FindAllElements, EventToSelectorMap, CollectPostWithTimeout,
  CollectWithRequestWithTimeout) in v0.6.0. Natural tranche-3 heads:
  JSON-aware `RequireSignalsContain`, `RequireRedirect`/`RequireHeader`
  assertion wrappers, `ServeSSE`/`NewRecorder` handler-less synthesis
- `datastartest.NewResponse`-style helper for test ergonomics
- Response ergonomics: a `signalsMap` type for the signals-patch pattern
  (held pending a concrete consumer example); review `signalKeyMessage` naming
- `example/README.md` and an `example/docker-compose.yml` for easy local
  runs; benchmark for `Collect` helper overhead
- Worked multi-instance replay example: a minimal Redis `sse.EventStore`
  implementation in `example/` (the broadcast README promises "bring your
  own" store but shows nothing)
- Broadcast subscriber-level metrics (connection counts, dropped-event
  counters) behind an optional interface
- Community metadata: GitHub Sponsors / funding, contributor list

### 3. CI/CD & Hermeticity

Make quality gates hermetic and reproducible.

Raw ideas:

- Route all lint/audit tools through nix checks so `nix flake check` is the
  single canonical quality gate (golangci-lint, erraudit, govulncheck);
  erraudit `--format sarif` output for GitHub code scanning
- Hermetic `checks.lint` / `checks.vet` / `checks.govulncheck` derivations;
  `flake.nix` `apps.bench` for running benchmarks
- Verify the erraudit probe-gate transition once the repo goes public
  (manual trigger or scheduled probe)
- `go work vendor` support or a flake app for offline module graphs
- Coverage-floor policy decision (optional CI gate at a threshold); per-module
  coverage badges (the single badge mixes example code)
- goreleaser: decide the skeleton's fate (run `build --snapshot` dry-run or
  delete it)
- Scheduled upstream-drift alarm beyond Renovate's custom manager (proxy
  `@latest` == newest pushed tag; embedded-JS `gh api releases/latest` check)
- Wire `nix run .#docspec` into CI so doc drift fails remotely, not just
  locally

### 4. Documentation Depth

Move beyond API reference into conceptual and operational docs.

Raw ideas:

- SSE heartbeat documentation outside `example/README.md` (the only place it
  is documented today)
- `docs/migration-guide.md` refresh per minor release (currently covers up to
  the 1.26.7 era; needs the 1.27.1/v0.6.x note)
- Website launch (Astro + Starlight pattern)
- CSP-mode (`data-nonce`, no `unsafe-eval`) documentation for ScriptHandler
  consumers — upstream v1.0.3 capability, undocumented here
- `docs/architecture.md`: add the lockstep release-train diagram
- CONTRIBUTING.md: release-prep walkthrough pointing at the checklist,
  mentioning the auto-commit daemon

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

- **Dependency bots: one-bot policy (decided 2026-10-01):** Dependabot is the
  sole ecosystem bot (gomod ×4 modules + GitHub Actions); Renovate stays,
  scoped by `enabledManagers: ["regex"]` to the embedded-DataStar-JS custom
  manager — the one update class Dependabot cannot express. Duplicate
  gomod/actions PR churn ends. If the JS proposals ever prove not worth it,
  the fallback is theme 5's scheduled drift alarm, not hand-bumping.
- **Changelog automation (decided 2026-09-03):** manual keep-a-changelog
  stays. Evaluated changie and GitHub-native auto-notes: with ~monthly
  releases, an append-only hand-maintained file plus the release-checklist
  gate is less machinery than a bot config, and the append-only policy (G4)
  is easier to enforce by eye. Revisit if release cadence reaches weekly.
- **datastartest CHANGELOG placement (decided 2026-09-03):** datastartest has
  NO module-level CHANGELOG file; its changes are recorded in the root
  CHANGELOG under "Added — datastartest" sections. Rationale: the lockstep
  release train (ADR 002) means one release = one history; a second file
  would duplicate every entry and drift. Revisit only if datastartest ever
  releases independently of root.
- **AGENTS CI summary table (decided 2026-09-03):** no separate summary table
  — the CI section's bullets already carry per-workflow state (active /
  probe-gated / watch) and the promotion triggers live in
  [docs/ci-watch.md](docs/ci-watch.md). A table would duplicate both and push
  AGENTS.md over its 15KB budget.
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
- **`go` directive policy (decided 2026-08-16, updated 2026-09-19):**
  directives pin the exact patch release — now `go 1.27.1` across go.mod ×3,
  go.work, CI `go-version`, and the flake `goPkg` (nixpkgs `go_1_27`),
  keeping the 1.26.6-era stdlib CVEs (GO-2026-5972/6089/6090/6218) cleared
  and un-gating `encoding/json/v2` — the 1.26.7 directives could not build
  under a 1.27 toolchain (jsonv2 language-version gate).
