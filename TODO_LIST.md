# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.
> Completed items are removed and logged in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |

## Verified next-up

| Task                                                                                                                                                                                                                      | Status    | Impact | Effort | Evidence                                                                        |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------- |
| Harden `docs/release-checklist.md` with the v0.6.0 lessons: add §2.5 "re-run `nix flake check` after require bumps, paste EVERY moved hash"; replace the phantom `go mod edit -version` (line 21, flag does not exist) with the real sibling-require procedure; pin explicit tag-push refs instead of `git push --tags`; add "verify GitHub Latest = root release" step. | 🔴 `TODO` | High   | 30min  | `2026-09-18_22-29` b1 + f1–f4 (★ items); checklist verified stale 2026-10-01    |
| Run the erraudit loop over all four modules (AGENTS Commands block) — skipped at v0.5.0 AND v0.6.0; reconcile any findings with `.golangci.yml` excludes.                                                                   | 🔴 `TODO` | High   | 15min  | `2026-09-18_22-29` b2, f33; `2026-09-03_12-26` b9                               |
| AGENTS.md prune: pruned to ~18.0KB by the 2026-10-01 pass (from 19,008B) — still above the 15KB target; further cuts would touch load-bearing gotchas (shared-checkout reality, wire-format rules). Decide the settle point or prune deeper. | 🔴 `TODO` | Low    | 30min  | `wc -c AGENTS.md`; `2026-09-18_20-45` f27                                                        |
| datastartest tranche-2 polish: godoc `Example*` functions for the six new helpers (`RequireNotScript`, `FindScript`, `FindAllElements`, `EventToSelectorMap`, `CollectPostWithTimeout`, `CollectWithRequestWithTimeout`) + a doc paragraph on `FindAllElements` noting script patches participate (it has no doc comment at all today). | 🔴 `TODO` | Medium | 45min  | `datastartest/search.go` (FindAllElements has no godoc); `2026-09-18_21-02` b5/e7, f9/f10 |
| `docs/migration-guide.md`: add the v0.5.0→v0.6.x note AND fix the stale "Go toolchain: 1.26.7 required" claim (floor is 1.27.1 since v0.6.1, and `GOEXPERIMENT=jsonv2` is no longer needed).                               | 🔴 `TODO` | Medium | 20min  | `docs/migration-guide.md:6`; CHANGELOG [0.6.1]; `2026-09-18_22-29` f16           |
| `datastartest/README.md`: note which helpers arrived in v0.6.0 (tranche 2), matching the per-release precedent.                                                                                                             | 🔴 `TODO` | Low    | 10min  | `2026-09-18_22-29` f17; grep finds no v0.6.0 mention in the README              |
| docspec mirroring for `docs/wire-format.md` + `docs/migration-guide.md` snippets (the guarantee is partial: replay/error-system/testing are mirrored, these two are not); fix the `docs/testing.md` quick-start snippet divergence (handler lacks `WithModeAppend` that its own assertion expects). | 🔴 `TODO` | Medium | 1h     | `docspec_test.go` (mirrors listed); `2026-09-03_12-26` b2/b3, f23/f24            |
| `docs/static-js.md`: document CSP mode (`data-nonce`, no `unsafe-eval`) for ScriptHandler consumers + the canonical-minified-only policy; add `static/fetch-bundle.sh` (download, sha256, provenance comment in one step).   | 🔴 `TODO` | Medium | 45min  | `2026-09-03_15-44` c5/c13, f8/f9/f19; no CSP/policy text in static-js.md today  |
| Re-measure datastartest coverage post-tranche-2 (last recorded 93.4%, 2026-09-03) and record the number in the next CHANGELOG entry.                                                                                         | 🔴 `TODO` | Low    | 15min  | `2026-09-18_21-02` f6, `2026-09-18_22-29` f20                                    |
| Check upstream `starfederation/datastar-go` for a release newer than v1.2.2 and refresh the README comparison table if so (checklist §5 cadence).                                                                           | 🔴 `TODO` | Low    | 15min  | README "current release" footnote; `2026-09-18_22-29` f18                       |
| Broadcast API ergonomics tranche (from the 2026-09-18 review): `NewBroadcasterWithStore(sse.EventStore)` injection seam only — consumer-supplied stores, NO backend implementations in this repo (store.go already directs multi-instance users to bring their own); close the buffer-size × replay constructor-matrix gap, optional heartbeat interval.     | 🔴 `TODO` | Medium | 1h     | `docs/status/2026-09-18_19-54_broadcast-code-review.md` (R1–R3)                 |

## Owner-blocked

| Task                                                                                                                                                       | Status       | Impact | Effort | Evidence                                               |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------ |
| Dependency bots: keep Renovate or Dependabot, disable the other (both active today — duplicate PR churn); also set Renovate dashboard/label preferences.   | 🔵 `BLOCKED` | Medium | 15min  | `.github/dependabot.yml` + `renovate.json` coexist     |
| Triage the remaining dependabot PR (#14, codeql-action analyze SHA bump); #16 (actions group) merged 2026-09-29, x/mod PRs closed via the group.           | 🔵 `BLOCKED` | Low    | 5min   | `gh pr list`; depends on the one-bot decision          |
| Delete merged branch `pr/docs-test-consolidation` (local + remote; PR #3 merged). Irreversible — needs owner nod.                                          | 🔵 `BLOCKED` | Medium | 5min   | PR #3 state: MERGED                                    |
| Rehome or drop `preserve/status-report-coderabbit-pr3` (sole copy of the 11-37 report with the 50-item table). PR it or delete — owner decision.           | 🔵 `BLOCKED` | Medium | 15min  | branch exists; observed-branches configured 2026-09-03 |
| CODEOWNERS: create with named owners (deliberately NOT created 2026-09-03 — naming is the owner's call; SUPPORT.md + discussion templates landed instead). | 🔵 `BLOCKED` | Low    | 10min  | SUPPORT.md exists; no CODEOWNERS by design             |
| erraudit CI job flips to hard gate when the repo goes public — verify the probe notices the flip on the first push after publication.                      | 🔵 `BLOCKED` | Low    | 5min   | `ci.yml` probe job; AGENTS.md Nix gotchas              |
| Website launch (Astro + Starlight pattern) — deferred by owner decision (T27-of-18-32).                                                                    | 🔵 `BLOCKED` | Low    | —      | ROADMAP theme 4                                        |
| Status-index "Monitoring" tier for never-fully-resolvable reports; per-item marker depth for consolidated ROADMAP ideas.                                   | 🔵 `BLOCKED` | Low    | 15min  | 2026-09-02 report owner questions                      |
| Shared lint-cache policy: purge the 3.8G `/mnt/buildcache/golangci-lint` or bless `GOLANGCI_LINT_CACHE=$(mktemp -d)` as the documented default for gates.  | 🔵 `BLOCKED` | Low    | 5min   | `2026-09-18_22-29` g1; AGENTS.md cache-ghosting gotcha |
| Required-checks policy: master has none — document the "local gates are the gate" contract in one canonical place, or make lint/nix alerting.              | 🔵 `BLOCKED` | Low    | 15min  | `2026-09-18_20-45` g3, f26                             |

## Notes

- Harvested + rebuilt by the 2026-10-01 docs-health pass over the 2026-09-03
  and 2026-09-18 report batches (v0.4.0 → v0.6.1 era). Items closed since
  their source reports were annotated inline there; 8 fully-resolved files
  archived (see `docs/status/README.md`).
- Broadcast tranche row unchanged from the 2026-09-18 routing (R1–R3,
  owner-gated for the next minor per that session's scope answer).
- The prior 2026-09-03 rebuild note is history: see CHANGELOG [0.4.0]–[0.6.1]
  and the archived T01–T27 reports for what shipped.
- Resolved questions live in ROADMAP.md "Resolved questions"; release-gate
  policy questions (erraudit mandatory? fresh-cache lint default?) are folded
  into rows above pending the owner's call.
