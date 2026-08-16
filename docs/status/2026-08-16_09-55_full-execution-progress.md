# Status: Full Execution Mode — Pareto Plan Execution (2026-08-16 09:55)

**Trigger:** "NOW GET SHIT DONE! The WHOLE TODO LIST!" — executing
`docs/planning/2026-08-16_08-53_pareto-ci-trust-recovery-and-hardening-plan.md`
(T01–T16) with default rulings for the 3 owner questions (recorded in ROADMAP).
**Stopped mid-execution on owner instruction ("write the .md and wait").**

---

## Owner rulings applied as defaults (all recorded in ROADMAP "Open questions")

1. **Q1 Go directive:** YES — bumped to `go 1.26.6` everywhere. Supersedes the
   v0.0.2/v0.0.3 CHANGELOG "lowered to 1.26" ghost.
2. **Q2 erraudit CI job:** probe-gate, don't delete. Job probes
   `go list -m github.com/larsartmann/erraudit@v0.3.0`; skips with a visible
   notice while the repo is private, becomes a hard gate when public.
3. **Q3 DOMAIN_LANGUAGE.md:** create (T15) — **not yet executed**.

## Completed and verified

### T01 — Green master CI (the 1% / 51%) — DONE, PUSHED, CI GREEN
- `go 1.26.6` in 3× go.mod + go.work + ci.yml `go-version` (via `go mod edit`,
  never sed). Clears GO-2026-5972/6089/6090/6218.
- ci.yml erraudit job: probe-gated, `continue-on-error` removed, **latent bug
  fixed** (old command passed 3 package patterns; erraudit v0.3.0 takes ONE
  directory — it could never have passed even with the repo public). Same fix
  applied to AGENTS.md command and `nix run .#erraudit` app.
- flake.nix: `go_1_26.overrideAttrs` pins 1.26.6 source (nixpkgs still ships
  1.26.5; TODO comment marks removal). vendorHash refreshed (1.26.6's
  `go mod vendor` output differs). treefmt-nix's built-in `flakeCheck` disabled
  (it registered an unguarded second treefmt-check whose goimports tried a
  network toolchain download in the sandbox; the guarded `checks.format` with
  goPkg first on PATH is the only one now).
- **Discovery:** erraudit is still PRIVATE on GitHub; the earlier "it's public"
  read was my machine's proxy-cache artifact. CI correctly skips + notices.
- **Gate:** workspace build/vet/race green; GOWORK=off ×3 green; lint 0 issues;
  `go work sync` idempotent; `nix flake check` ALL CHECKS PASSED; hermetic
  govulncheck: **No vulnerabilities found.**
- **Commits:** `bf68063` (auto-committed by daemon, pushed), `affbe30` (ADR 002
  + AGENTS + vendorHash), `f198377` (flake dedupe). CI run 31933895108:
  **success**. Master is green.

### T02 — Hygiene pack — DONE
- `result` symlink trashed. ROADMAP go-directive entry resolved with 1.26.6
  facts. AGENTS erraudit invocation corrected. `nix flake check` green
  (recorded here, not in TODO_LIST — harvest-out pending).

### T03 — Formatter & lint wiring — DONE (uncommitted)
- `dprint.json` removed (guard G7 verdict: wiring it would make hermetic
  `nix flake check` depend on network-fetched WASM plugins; treefmt stays the
  single formatter). CHANGELOG entry added.
- `actionlint` CI job added (v1.7.12, pinned; validated locally: clean).
- `actionlint` added to devShell. erraudit in devShell **attempted and
  reverted**: hermetic build impossible — erraudit's dep tree contains private
  modules (`go-finding`), sandbox build fails. App now documents this and
  go-installs with credentials. devShell builds green.

### T05 — ADR 002 — DONE (committed in `affbe30`)
- `docs/adr/002-multi-module-split.md`: three modules, strict DAG
  (static → root → datastartest), mutual relative replaces with 4 enforced
  rules, lockstep tags, consequences. Linked from AGENTS Module Structure.
  Facts cross-checked against the modularization proposal + live go.mod/tags.

### T06 — CONTRIBUTING multi-module section — DONE (uncommitted)
- New "Multi-Module Development" section (workspace vs GOWORK=off, replace
  rules, per-module tags). **Fixed drift:** old text claimed the devShell sets
  `GOWORK=off` (false) and told manual-setup users to export it (wrong for a
  workspace repo). All commands verified green this session.

### T07 — CollectPost error-path tests — DONE (uncommitted)
- `datastartest/collect_error_test.go`: 400/422/500 × non-SSE bodies, 200
  non-SSE, garbage frames, and the documented sharp edge (SSE payload in a 500
  still decodes). Pins current behavior: helpers never gate on status; failure
  mode is zero events, never panic/hang. All pass, race.

### T08 — e2e dogfood expansion — DONE (uncommitted)
- `TestE2E_CollectPostRoundTrip` (ReadSignals → echo signals patch) and
  `TestE2E_CollectNStreaming` (open-stream handler; CollectN(2) returns without
  waiting). Pass, race.

### T09 — Example WithOnDrop integration test — DONE (uncommitted)
- `example/ondrop_test.go`: WithBufferSize(2) + slow subscriber → asserts
  exact dropped events in order, buffer contents, no extra drops. Pass, race.
  Example godoc now points to it.

### T10 — UnmarshalSignals fuzz — DONE (uncommitted)
- `datastartest/event_fuzz_test.go`: 17 seeds (valid/malformed/nested/BOM/NUL);
  invariant = classified `datastartest.signals_unmarshal_failed` code or
  success, never panic. Seeds pass; 10s exploration: 2.27M execs, clean.

## Remaining (in plan order)

| Task | Content | Notes |
| ---- | ------- | ----- |
| T11 | Per-module Nix hermetic checks (`hermeticCheckStatic` with `vendorHash = null`, `hermeticCheckDatastartest`) | remove flake TODO comment |
| T12 | README comparison re-verify (v1.2.2 pin), JS-version row decision, `docs/release-checklist.md` | |
| T13 | De-flake `TestCollect_WithLastEventID_HeaderArrives` (channel-sync, G6: no sleeps) | passed race here; still worth hardening |
| T14 | `docs/modularization/README.md` index + AGENTS link | |
| T15 | `docs/DOMAIN_LANGUAGE.md` (Q3 ruling: create) | |
| T16 | go.work.sum tracking decision, v0.0.0 vs real sibling requires | rulings recorded, mechanics pending |
| T04 | Branch protection (gh api), pkg.go.dev verify, coverage badge | last — after everything else lands |
| — | Final sync: TODO_LIST harvest-out (19 rows → done), CHANGELOG pass, AGENTS notes, full gate | |

## Uncommitted working tree (mine, verified)

`.github/workflows/ci.yml` (actionlint job), `CHANGELOG.md` ([Unreleased]
entries for T01/T03), `CONTRIBUTING.md` (T06), `flake.nix` (T03), dprint.json
deleted, `datastartest/collect_error_test.go`, `datastartest/e2e_test.go` (T08
legs), `datastartest/event_fuzz_test.go`, `example/ondrop_test.go`,
`example/main.go` (godoc note). All tests/lint/flake-check green on this tree.

## Gotchas learned this session (AGENTS candidates)

- erraudit v0.3.0 CLI: ONE directory argument; multi-pattern invocations fail.
- erraudit + go-finding are private → no hermetic Nix build possible; probe-gate
  in CI, go-install in the app.
- treefmt-nix `flakeCheck = true` registers its OWN unguarded checks.treefmt;
  goimports needs a directive-satisfying `go` first on PATH (gotools wrapper
  only APPENDS its go). Keep `flakeCheck = false` + guarded `checks.format`.
- Go patch bumps move `vendorHash` (modules.txt format changed in 1.26.6).
- Parallel session + auto-commit daemon are live: stage by explicit path list,
  re-read files before every edit (hit 3 mod-time races this session).
- BOM in Go source = compile error ("illegal byte order mark"); use `"\xef\xbb\xbf"`.

## Resume instruction

Read this file + the plan, `git status` (expect the uncommitted list above),
run the full gate, then continue with T11 → T12 → T13 → T14 → T15 → T16 → T04
→ final sync. Do not re-do completed tasks.
