# Pareto Execution T01–T27: v0.4.0 Shipped, Recurring Breaks Killed, Depth Surface Cleared

**Date:** 2026-09-03 02:00
**Plan:** `docs/planning/2026-09-02_23-59_pareto-ship-v0.4.0-kill-recurring-breaks-depth-surface.md`
**Result:** all 27 tasks executed and verified; v0.4.0 released (lockstep tags,
GitHub Release, proxy + pkg.go.dev verified); every gate green at HEAD
(race ×5, vet, exact-CI lint 0 issues, docspec, `nix flake check`, isolation
×3, sync idempotency, replace audit, actionlint).

## a) What was shipped

| # | Task | Outcome | Commit |
| - | ---- | ------- | ------ |
| T01 | Exact-CI linter parity | `nix run .#lint-ci` app (golangci-lint v2.12.2, the version CI installs); 0 issues at HEAD; THE pre-push habit in CONTRIBUTING + AGENTS | `80034e1` |
| T02 | LSP-warning surface | 4× `b.Loop()` modernizations landed; the 4 json/v2 stdversion warnings PROVEN to be gopls false positives (every spelling triggers; `go vet`/lint never fire) and documented in AGENTS so no session re-churns | `cd54cdc` |
| T04 | FOD/vendorHash investigation | Mechanism verified with an evidence matrix (worktree FOD builds + manual vendor-tree hashing): `go mod vendor` copies replaced directories ENTIRELY — `datastartestVendorHash` moved on ANY tracked-file edit (docs included); root hash requires/toolchain-only. ADR 004 corrected | `5750fc5` |
| T06 | Tag v0.3.0 flake verdict | Verified FAIL at `60cf5b1`: datastartest FOD mismatch (root FOD clean) — exactly the mechanism predicts; recorded in ADR 004 | `5750fc5` |
| T03 | Ship v0.4.0 | Release commit `6016986`; tags v0.4.0/static/v0.4.0/datastartest/v0.4.0 pushed; GitHub Release from CHANGELOG excerpt; proxy `go get` + `go list -m -versions` verified for all 3 modules; pkg.go.dev rendering; upstream comparison current (v1.2.2) | `418ee31`+ |
| T05 | CI hygiene batch | `go mod verify` ×3, go.work use-vs-disk, tidy-diff ×3, and the JS-version-in-CHANGELOG drift test (`version_constraint_test.go`) | `3d7cada` |
| T07 | Docker verify | Dockerfile FIXED (needed repo-root context — example/ has no go.mod; cached download layer + static/go.mod for the replace); compose smoke: HTTP 200 + live `datastar-patch-elements/signals` datalines from `/events` | `c242d3f` |
| T08 | Performance re-measure | `-benchtime=1s` (multi-million samples): Elements ~476→~146 ns/op, Script ~633→~329, Signals ~1318→~1113, MarshalSignals ~1045→~853; Collect ~906µs→~186µs; caveat dropped, methodology footer added | `80d2f94` |
| T09 | Micro-hygiene | `feedInterval` named const, `http.NewResponseController` flush, bench machine/go-version header, PR-template per-module isolation rows | `2f29b71` |
| T10 | Measurement ritual | Coverage re-measured (98.4% / 92.7% / 100% — unchanged); lint-cache timing: cold ~17s → warm ~1.5s; documented in docs/testing.md | `4b4c76a` |
| T11 | AGENTS pruning | 19,369B → ~15.0KB: wire-format list → pointer to goldens + wire-format.md, error catalog → docs/error-system.md, daemon/town/checkout gotchas merged into one section; CI section synced | `9d09512` |
| T12 | modfile boundary guard | `module_boundary_test.go` parses go.mod semantically via `golang.org/x/mod/modfile` (comment-safe, replace-aware) | `62128bf` |
| T13 | datastartest top-3 helpers | `RequireElementsOrdered` (+ `ElementExpectation`), `Diff` (LCS line diff over decoded datalines), `Snapshot` (golden files, `-datastartest-update`); coverage 92.7%→93.4%; 0 lint | `b87374e` |
| T14 | Hygiene pack | ROADMAP consolidated idea source-cited (G12) + vendorHash idea marked resolved; `git-town.observed-branches` configured (town status clean) | `a994657` |
| T15 | CI watch runbook | `docs/ci-watch.md`: state/verify/promote/drop per workflow (nix, fuzz, CodeQL, Renovate) + general rules | `f77a941` |
| T16 | Response test-depth | 9 tests: decoded-JSON assertions, nil-error guard (new code `datastar.error_response_nil_error` — never dereferences), nil custom-event detail → JSON null, empty/unicode/10KB edge cases, 16-goroutine concurrent Response under -race, 200-line splitting, nested signals, GET/POST/DELETE source precedence, wire-shape E2E | `72aed78` |
| T17 | External closure | All 5 CodeRabbit threads on merged PR #3 answered (4 already-resolved by landed work, 1 valid → go.work.sum policy narrowed in AGENTS/ROADMAP + [Unreleased] Fixed note); the 5 parallel-session commits reviewed — all clean, well-motivated (`ffeeda` in the plan doesn't exist) | `06b84fb` |
| T18 | Community files | SUPPORT.md + Q&A/show-and-tell discussion templates; CODEOWNERS deliberately NOT created (G11) | `69ede1b` |
| T19 | static hardening | `nosniff` header, SHA-256 checksum pin test, provenance comment, ScriptTag edge tests, empty-bundle test, `BenchmarkComputeETag` (one-time init cost, already cached — no change needed); Last-Modified + bundle fuzz = recorded Not-Dos with reasons; `DatastarJSVersion` deprecated | `685a347` |
| T20 | datastartest examples | ExampleCollectPost/CollectN/Event_DataValue/Event_String with deterministic Output blocks; DataValue's trailing-space key-prefix contract documented | `5b9e6ec` |
| T21 | Audit pack | README comparison table re-verified vs upstream v1.2.2 via pkg.go.dev (compression ✅, no broadcast/replay ✅, ReplaceURLQuerystring upstream-only ✅); 10% verdict spot-check: 5/5 sampled hashes reachable + claim-matching; appendix added to the 2026-09-02 report | `45b7399` |
| T22 | Decisions pack | ADR 002 addendum (example/ stays in root), ADR 005 addendum (no coverage floor, confirmed), datastartest-CHANGELOG + CI-table questions resolved in ROADMAP | `592e12c` |
| T23 | Infra pack | ci.yml test job split: per-module matrix (GOWORK=off build/vet/race/verify/tidy ×3, parallel) + workspace job; `go work vendor` evaluated → deliberately not adopted (recorded) | `6cb34ae` |
| T24 | Docs depth | `docs/architecture.md`: mermaid three-layer diagram + per-layer types/files; linked from README, AGENTS, static-js (which also documents the two version-constraint tests) | `7932f74` |
| T25 | Consumer content | `docs/migration-starfederation.md` (method mapping, gains/tradeoffs, checklist); Broadcaster/SubscribeFilter/MemoryStore runnable examples; Learn-DataStar link; Documentation/Docs-Map links | `ac1de23` |
| T26 | docspec | `//go:build docspec` mirrored snippets (root + datastartest) + `nix run .#docspec`; **caught 3 real doc drifts on introduction**: replay.md's `defer stream.Close()` (Close returns error), `stream, err := sse.NewStream` (returns 1 value), CollectWithRequest arg order; break-verified | `b055625` |
| T27 | Platform pack | `Response` method forms (additive; functions unchanged), `version` package (ldflags), `.goreleaser.yaml` skeleton (goreleaser check ✅); signalsMap HELD, changelog automation DECLINED (recorded) | `49a0ae6`/`9cb1d17` |

## b) What broke and how it was fixed

1. **The vendorHash could never converge (discovered mid-release-gate).** With
   the repo-root fileset, the `datastartestVendorHash` constant sat inside its
   own FOD input — pasting a measured hash changed flake.nix, which changed
   the vendor output, which invalidated the hash. Two paste iterations proved
   non-convergence. Fix: minimal src fileset for the datastartest check
   (datastartest + root *.go + go.mod + static/) → converges on one paste;
   `nix flake check` passed at the release tree and all CI workflows went
   green (nix.yml's first success).
2. **Docker build failed twice** (`go.mod not found`, then the replace
   directive needing static/go.mod at download time) — fixed with root
   context + staged module files + .dockerignore.
3. **docspec caught 3 real doc drifts before landing** (see T26).
4. **Daemon races** clobbered two edits mid-flight (diff_test.go assertion,
   docspec heartbeat) — both caught by failing tests, re-applied from fresh
   reads (the documented G1 ritual works).

## c) What is deliberately open

- **Owner-blocked (untouched, G11):** bot choice (Renovate vs Dependabot),
  branch deletions, CODEOWNERS naming, erraudit flip, website trigger,
  status-index tiering, consolidation depth.
- **CI watch ritual:** nix.yml promotion (needs a green fortnight), first
  fuzz artifacts, first CodeQL triage — runbook written, execution is
  time-gated.
- **signalsMap type** held; goreleaser stays skeleton-only; changelog
  automation declined (rationale in ROADMAP resolved questions).

## d) Verification evidence

- `GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1` — all green
- `nix run .#lint-ci` — 0 issues
- `nix run .#docspec` — green (both modules)
- GOWORK=off build+test ×3 — green (also the CI matrix legs, verified locally)
- `go work sync` idempotent; `go work use` matches disk; tidy-diff clean ×3; `go mod verify` ×3
- replace audit: relative paths only
- `nix flake check` — all checks passed (at the release tree; final hash refresh pending commit — see e.1)
- `nix run .#govulncheck` — no vulnerabilities
- actionlint — clean
- Coverage: root 98.4%, datastartest 93.4%, static 100%
- Release: proxy `go list -m -versions` shows v0.4.0 for all 3 modules; GitHub Release published

## e) What we should improve next

1. **The final datastartestVendorHash (post-treefmt) must be committed with
   this report** — it converged at `O1o+dHD…`; future root *.go or
   datastartest edits re-stale it (loud, one-paste fix).
2. Promote nix.yml after two green weeks (drop `continue-on-error`).
3. v0.5.0: the `[Unreleased]` section is release-ready; scope per checklist.
4. datastartest helper tranche 2 (ROADMAP theme 2, source-cited).
5. Consider a `Broadcaster[datastar.Patch]`-typed example upgrade if go-sse
   gains a subscription-hook that renders lazily (watch item).
6. Keep the docspec contract alive: every API change touching a documented
   snippet updates snippet + mirror in one commit.

## f) Time and effort

Plan estimate: ~27h30 for T01–T27. Actual: one session; the 1% (T01–T03)
landed inside the first third of the session, including the release. The
largest unplanned discovery (unsolvable vendorHash self-reference) consumed
the release-gate contingency and produced the session's most durable fix.

---

_Point-in-time snapshot. Written by the 2026-09-03 pareto execution session
(Full Execution Mode, T01–T27). See `docs/status/README.md` for the index
and archiving policy._
