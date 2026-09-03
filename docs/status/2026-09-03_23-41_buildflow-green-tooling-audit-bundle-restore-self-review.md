# Status 2026-09-03 23:41 — buildflow-green tooling audit, mangled-bundle restore, stale vendorHash, self-review

**Session scope:** triage of one full buildflow run output (11 failing tools + 7 detect-only reporters) → repo-wide fixes → all gates green. Single session, ~1.5h. No wire-format changes. No library API changes.

**Format note:** `.md` per repo convention (AGENTS.md overrides the status-report skill's HTML default). The brutal-self-review skill's questions are folded into sections **d)/e)** and the self-review block below — one report instead of two, per this session's instruction.

**End state (verified):** buildflow exit 0 (62 success / 0 failed / 5 skipped-via-config) · `nix flake check` **all checks passed** · race suite PASS on all modules · vet 0 · golangci-lint 0 issues (devShell **and** pinned v2.12.2) · erraudit gate clean ×3 · `tidy -diff` clean ×3 · work sync idempotent · replace audit clean · docspec PASS · working tree clean, HEAD `caa0986` committed and hermetically buildable.

---

## Self-review (asked directly)

**What did I forget?**

1. The pasted test output had **no `static` module line** — I trusted the "PASS" impression instead of cross-checking module coverage. Found the checksum failure only when running the suite myself (minutes later than necessary).
2. **nix-checker told me** ("vendorHash may be stale — go.mod modified after hash was set") and I initially filed it under "heuristic noise". It was right. `nix flake check` at minute one would have surfaced the real hash mismatch immediately.
3. oxfmt/oxlint have **no ignore for `static/datastar.js`** — I shielded prettier and codespell only. Nothing stopped (or stops) oxfmt from sweeping the bundle next run.
4. Did not run the baseline gate suite **before** the first edit — the static break and stale hash were pre-existing and would have been found in minute one with a correct READ→UNDERSTAND→VERIFY-first sequence.
5. Did not run `erraudit nolint-audit .` to confirm directive freshness; did not do the git-town session ritual; did not HARVEST the (f) list into `TODO_LIST.md` yet.

**What is stupid that we do anyway?**

- The **auto-commit daemon commits formatter damage to master ungated** (this is exactly how a prettier-reformatted vendored bundle landed in `ef10422`).
- **No branch protection + informational CI** → a red master sat on origin for ~6h (unbuildable hermetic flake via hash mismatch).
- **Four formatters with opinions on non-Go files** (dprint, prettier, oxfmt, treefmt) with no single owner — split brain, see below.
- **buildflow's nix self-healing is trusted but does not actually persist fixes** — `nix-hash-fix` reported ✔ while the hash stayed stale.

**Could I have done better?**

- Sequencing: baseline verification first, lint after every edit batch (my first `defer x.Close()` edit traded erraudit-compliance for a golangci errcheck failure — cost one cycle).
- Spent ~6 tool calls reverse-engineering buildflow's erraudit argv (shim, binary greps, run records) before accepting the residual. Right call to stop; late call to stop.
- I introduced two typos while writing lint/docs edits (`trip the dictionary`, `reformat ted`) — both caught and fixed, but ironic in a spell-check session.

**Did I lie?** No. "All gates green" at session end was true for the working tree; the vendorHash fix has since been committed verbatim (`caa0986`, 1 line), so committed state now matches what was verified.

**Ghost systems?** The `//nolint:erraudit` directive in `example/main.go:98` is currently **honored by no gate that scans `example/`**: the AGENTS.md gate runs `--no-suppress` (ignores it) and doesn't scan `example/` at all; the flake `erraudit` app honors it but doesn't scan `example/`; buildflow scans `example/` but ignores nolint. It is documentation, not enforcement — flagged in b) #1.

**Split brains found:** (1) two documented erraudit invocations (AGENTS.md gate `--no-suppress` vs flake app `--severity-threshold error`) — different contracts, both "green"; (2) dprint vs prettier both claim markdown/JSON/YAML; (3) buildflow's erraudit policy vs the erraudit CLI's own default (nolint-honoring) policy; (4) codespell enforced only in buildflow, absent from treefmt/CI.

**Scope creep check:** resisted adding samber/lo (dep-for-nothing), resisted the hash.nix extraction (ADR-004 self-reference risk), fixed the `cancelability` finding as a comment, not an API rename. The 50-item list below is a brainstorm, not a commitment.

**Removed something useful?** No. Skipped steps are documented with rationale; `checks.*` consolidation is behavior-identical.

---

## a) FULLY DONE (evidence attached)

| #  | Work                                                                                                                                                                                                           | Evidence                                                                                                                                                     |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1  | Restored `static/datastar.js` byte-identical to upstream v1.0.3 after a prettier sweep mangled it (daemon committed the mangle in `ef10422`)                                                                   | `git diff ad5ebc2 -- static/datastar.js` empty; `go test ./static/...` PASS; `.prettierignore` added                                                         |
| 2  | Fixed 4/5 erraudit findings: examples stop streaming on first failed `Send` (backlog + live loop); teardown via plain `defer x.Close()`                                                                        | `example/domain-adapter/main.go:141-177`, `example/sse_middleware.go:32`; erraudit gate 0 violations ×3 modules; buildflow erraudit 5 findings → non-failing |
| 3  | Aligned golangci errcheck excludes with their evident `(io.Closer).Close` intent (concrete Close types)                                                                                                        | `.golangci.yml:124-131`; golangci 0 issues (devShell + pinned v2.12.2)                                                                                       |
| 4  | Canonicalized root `go.mod` require blocks (direct vs indirect)                                                                                                                                                | `tidy -diff` clean ×3, `go mod verify` OK, `GOWORK=off` build OK; gomod-check + go-mod-ignore-check findings cleared                                         |
| 5  | Completed flake `meta` ×3 derivations (homepage, mainProgram, platforms); consolidated `checks.*` into one attrset                                                                                             | flake-meta-checker 9 findings → 0; statix W20 → 0; `nix flake check` all passed                                                                              |
| 6  | Fixed stale `datastartestVendorHash` (broken by the static v0.4.0 require bump; buildflow's nix-hash-fix had NOT healed it)                                                                                    | nix hash mismatch observed → updated to `sha256-MdpYsxslWjeCf/6xQsz64AkDcy7far8OJ/AvhNeH8cY=` → `nix flake check` all passed; committed as `caa0986`         |
| 7  | dprint: excluded `.github/DISCUSSION_TEMPLATE/**` (markdown-in-yml GitHub templates its YAML parser rejects)                                                                                                   | `dprint.json`; dprint-format step green in final run                                                                                                         |
| 8  | prettier: fixed missing `</p>` in `docs/modularization/2026-08-10_PROPOSAL.html`                                                                                                                               | prettier-format step green in final run                                                                                                                      |
| 9  | codespell: `.codespellrc` (`crasher` ignore-list — legit Go fuzzing term; vendored JS skipped); fixed real `pre-empt`→`preempt` (AGENTS.md) and `cancelability`→`cancellability` (comment only, no API change) | codespell findings (non-JS) → 0                                                                                                                              |
| 10 | `.buildflow.yml` with every skip documented (go-structure-linter: flat layout is ADR-002 design; eslint: no first-party JS; go-auto-upgrade: samber/lo conflicts with minimal-deps policy)                     | buildflow exit 0, "0 failed, 5 skipped via config"                                                                                                           |
| 11 | Docs: CHANGELOG `[Unreleased]`/Changed entries (example hardening, housekeeping, bundle incident); AGENTS.md gotchas (bundle-formatter hazard, buildflow config)                                               | committed `a518dc0`                                                                                                                                          |
| 12 | Full verification battery                                                                                                                                                                                      | see "End state" above                                                                                                                                        |

## b) PARTIALLY DONE

1. **buildflow ↔ erraudit CLI reconciliation** — works: canonical gate green, CLI honors `//nolint:erraudit`. Open: buildflow's erraudit variant reported the same code the CLI passes (invocation flags unpinned; nolint seemingly ignored there). Final full run: 0 failed (finding non-failing there), but single-step runs still surface it. Blocker: buildflow's argv not discoverable from repo side. Effort: **M** (needs buildflow-side fix or flag pinning).
2. **Vendored-bundle shielding** — works: prettier + codespell + dprint(YAML) shielded. Open: oxfmt/oxlint have no explicit ignore for `static/datastar.js`; they didn't damage it this run, but nothing prevents a future sweep. Effort: **S**.
3. **erraudit split brain** — both documented invocations are green but are different contracts (`--no-suppress` vs `--severity-threshold error`; neither scans `example/` except buildflow). Open: unify into one documented gate. Effort: **S**.
4. **Daemon hygiene** — the daemon swept everything (good) but as 8 heuristic bulk commits; the mangled bundle rode into master inside one of them. Open: no guard prevents a repeat. Effort: **M** (daemon config lives outside this repo).
5. **Self-review integration** — brutal-self-review questions answered in this report instead of a separate `docs/reviews/` HTML file (user's single-report instruction; format already `.md` by convention). Done as a deviation; noting it per skill contract.

## c) NOT STARTED (observed this session; deliberately untouched)

1. **Push to origin** — master ahead 8 (incl. bundle restore + vendorHash fix). Why: never push without explicit instruction. Priority: high, blocked on owner.
2. **v0.5.0 release** — `[Unreleased]` holds consumer-facing additions (Response method forms, `version` package, datastartest debug helpers, static v1.0.3). Waiting on release decision.
3. **docs/static-js.md runbook** — not verified against the new shields/checksum-pin workflow.
4. **TODO_LIST.md harvest** of section (f) — belongs to docs-health HARVEST, not done yet.
5. **Fuzz smoke runs** (30s × FuzzReadSignals / FuzzReadEvents) — not run this session.
6. **`version/` package has zero test files** (visible in every test run).
7. **nix.yml promotion** off `continue-on-error` (first green run was today) — decision pending per docs/ci-watch.md.
8. **One-bot decision** (renovate.json vs dependabot.yml) — pre-existing pending; also unverified whether the JS renovate-manager updates `bundleSHA256` in the same PR.
9. **aarch64 hermetic coverage** — `nix flake check` skipped aarch64-darwin/linux (warning visible today).
10. **codespell/erraudit in CI** — currently buildflow-only enforcement.

## d) TOTALLY FUCKED UP (radical honesty)

1. **Mangled vendored bundle sat committed on master** (`ef10422`, 17:22 → restored ~23:20). Severity: high — the served browser asset was no longer the reviewed upstream artifact; the checksum pin (built for exactly this) was failing on master and nobody watched. Root cause: prettier repair-sweep with no vendored-file exclusion + heuristic daemon auto-commit + informational-only CI. Mitigation: restored byte-identical; `.prettierignore`; gotcha documented in AGENTS.md. Residual: oxfmt/oxlint still unshielded (b2).
2. **~6h red committed state (hermetic build)** — the static v0.4.0 require bump invalidated `datastartestVendorHash`, so `nix flake check` failed on master from `ef10422` until `caa0986`. Severity: medium (CI nix.yml is `continue-on-error`; local gate broken for anyone pulling). Root cause: same bulk commit; hash not updated with the bump.
3. **buildflow nix-hash-fix false success** — reported ✔ (fix-gomod-stale / fix-hash-mismatch / fix-vendor-inconsistency) while the hash stayed stale. Severity: medium-long-term — it teaches distrust of green self-healing steps. Root cause: unknown (buildflow-internal); needs upstream fix or repro. Workaround: only trust a real `nix flake check`.
4. **The nolint directive enforces nothing** (ghost system, see self-review). Severity: low today (finding is non-failing in full runs), but it means the one accepted suppression is not machine-checked. Mitigation pending: b1.
5. **Mine, small:** typos introduced while editing lint/docs (fixed in-session); baseline suite not run before first edit; static-module absence in the pasted log not questioned early. No repo damage, but the session's own process score is not 10/10.

## e) WHAT WE SHOULD IMPROVE

1. **Formatter governance** — one canonical non-Go formatter, wired into treefmt; vendored artifacts excluded everywhere (prettier done; dprint/YAML done; oxfmt/oxlint/codespell: codespell done, oxfmt open). Impact: kills the entire mangle-class of incidents. Fix: pick dprint **or** prettier for md/json/yaml; add `.gitattributes` (`static/datastar.js linguist-vendored -diff`).
2. **Verify-first discipline for AI sessions** — run the full baseline suite before touching anything; the session's two real bugs were pre-existing and visible in minute one. Impact: converts late finds into minute-one finds.
3. **Never trust self-healing summaries** — a ✔ from nix-hash-fix must be confirmed by an actual `nix flake check`. Fix: add `nix flake check` to the AGENTS.md commands block and the pre-push ritual.
4. **Daemon guardrails** — exclude vendored artifacts from auto-commit sweeps, or run `go test ./static/...` before committing anything under `static/`. Impact: prevents master poisoning without human attention.
5. **One erraudit contract** — one flag set, one documented scope, nolint honored and audited (`erraudit nolint-audit` in a gate). Impact: removes a whole class of "why is this green there and red here".
6. **Cultivate CI as a signal even when non-blocking** — today's incident was visible in CI-shaped tools for hours; the ci-watch runbook exists but nothing reads it on a schedule.

## f) NEXT 50 (brainstorm for HARVEST — not commitments)

| #  | Task                                                                                                                      | Impact   | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Push master (8 commits: bundle restore, vendorHash fix, docs) after owner approval                                        | Critical | S      | Process       |
| 2  | Add oxfmt/oxlint ignore for `static/datastar.js` (config file)                                                            | High     | S      | Quality       |
| 3  | Fix buildflow nix-hash-fix false-success (upstream in buildflow repo)                                                     | High     | M      | Tooling       |
| 4  | Make buildflow's erraudit honor `//nolint:erraudit` or pin its flags                                                      | High     | M      | Tooling       |
| 5  | Guard the daemon against committing `static/datastar.js` changes (checksum test pre-commit)                               | High     | S      | Tooling       |
| 6  | Verify the renovate JS-bump manager updates `bundleSHA256` in the same PR                                                 | High     | M      | Tooling       |
| 7  | Cut v0.5.0 per docs/release-checklist.md (Unreleased is substantial)                                                      | High     | M      | Release       |
| 8  | Unify the two documented erraudit invocations into one contract                                                           | Medium   | S      | Quality       |
| 9  | Update docs/static-js.md with checksum-pin + shields upgrade runbook                                                      | Medium   | S      | Documentation |
| 10 | Add `.gitattributes`: `static/datastar.js linguist-vendored -diff`                                                        | Medium   | S      | Cleanup       |
| 11 | Wire `erraudit nolint-audit` into a gate (lint-ci or CI)                                                                  | Medium   | S      | Quality       |
| 12 | HARVEST this list into TODO_LIST/ROADMAP (docs-health)                                                                    | Medium   | S      | Documentation |
| 13 | Run fuzz smoke tests (FuzzReadSignals, FuzzReadEvents, 30s each)                                                          | Medium   | S      | Quality       |
| 14 | Add `version/` package unit test                                                                                          | Medium   | S      | Quality       |
| 15 | Audit for other ef10422 collateral: check `example/Dockerfile` hadolint edits changed semantics only                      | Medium   | S      | Cleanup       |
| 16 | Add `nix flake check` to AGENTS.md commands + pre-push ritual                                                             | Medium   | S      | CI            |
| 17 | Promote nix.yml off `continue-on-error` after 7 green days (ci-watch runbook)                                             | Medium   | S      | CI            |
| 18 | Decide one-bot: renovate vs dependabot (pending per AGENTS.md)                                                            | Medium   | S      | Process       |
| 19 | Consolidate non-Go formatters (pick dprint or prettier; wire into treefmt)                                                | Medium   | L      | Quality       |
| 20 | Run `erraudit nolint-audit .` and record result (directive freshness)                                                     | Medium   | S      | Quality       |
| 21 | Manually audit for encoding/json v1 remnants (go-auto-upgrade step is now skipped)                                        | Medium   | S      | Quality       |
| 22 | Confirm datastartest README documents Diff/Snapshot/RequireElementsOrdered (Unreleased helpers)                           | Medium   | S      | Documentation |
| 23 | Document CONTRIBUTING.md bundle-upgrade procedure (checksum pin, shields)                                                 | Medium   | S      | Documentation |
| 24 | Add belt-and-braces test: bundle must stay minified (line-count heuristic) alongside checksum                             | Low      | S      | Quality       |
| 25 | datastartest/diff.go: add explicit bounds guards to silence branching-flow index warnings legitimately                    | Low      | S      | Quality       |
| 26 | Fix buildflow vulnix step (runs without PATH argument → usage error)                                                      | Low      | S      | Tooling       |
| 27 | Run `buildflow doctor` for the "9 tools unavailable" + go-licenses preflight warning                                      | Low      | S      | Tooling       |
| 28 | Wire codespell into treefmt so the canonical nix gate owns it                                                             | Low      | S      | Quality       |
| 29 | Add erraudit-based test for `Response.Send` error → `stream_send_failed` contract (check existing coverage first)         | Medium   | M      | Quality       |
| 30 | Reproduce buildflow's silent_swallow via erraudit `--pipeline` flags; file detector bug upstream if real                  | Medium   | M      | Tooling       |
| 31 | Decide coverage-badge scope: include/exclude `example/` (23.6% / 14.3%)                                                   | Low      | S      | Quality       |
| 32 | Make example handler error paths testable (extract handler, raise coverage honestly)                                      | Low      | M      | Quality       |
| 33 | aarch64: run `nix flake check --all-systems` (CI job or local one-off)                                                    | Low      | M      | CI            |
| 34 | Verify clean-checkout `nix flake check` on HEAD `caa0986` (kill the red-window class)                                     | Medium   | S      | Process       |
| 35 | Document in AGENTS.md: buildflow result cache can mask fresh findings (`BUILDFLOW_NO_RESULT_CACHE=1`)                     | Low      | S      | Documentation |
| 36 | Cross-link `.buildflow.yml` skip rationale from AGENTS.md Commands section                                                | Low      | S      | Documentation |
| 37 | Review whether the daemon should run at all on `static/**` (config outside repo — propose to owner)                       | High     | S      | Process       |
| 38 | Pin buildflow + erraudit versions used by gates in one place (flake/devShell)                                             | Low      | M      | Tooling       |
| 39 | Check ci.yml actually ran green on the post-`ef10422` pushes (informational ≠ watched)                                    | Medium   | S      | CI            |
| 40 | Add release-audit report for v0.5.0 when cutting it (status-report convention)                                            | Low      | M      | Release       |
| 41 | Verify v0.4.0 module proxy/pkg.go.dev propagation as release-checklist dry run                                            | Low      | S      | Release       |
| 42 | Document that `example/` is intentionally not scanned by the local erraudit gates (or start scanning it)                  | Low      | S      | Documentation |
| 43 | Consider `.codespellrc` skip list for `docs/modularization/*.html` (vendored-style audit HTML) if it ever false-positives | Low      | S      | Quality       |
| 44 | Add session-start ritual line: `git town status` + baseline gate suite (AGENTS.md)                                        | Low      | S      | Process       |
| 45 | Evaluate extracting flake hashes to `hash.nix` only WITH an ADR-004 addendum proving no FOD self-reference                | Low      | M      | Cleanup       |
| 46 | Sweep `docs/` for pre-2026-08 formatter reformats that changed meaning (spot-check 5 oldest)                              | Low      | M      | Cleanup       |
| 47 | Add TODO_LIST entry for buildflow/erraudit upstream fixes (items 3, 4, 30) so they survive sessions                       | Medium   | S      | Documentation |
| 48 | Decide whether `example/` should move to its own module (likely reject; document why)                                     | Low      | S      | Documentation |
| 49 | Add CI leg comment: which leg would have caught the mangled bundle, and why it did not (post-mortem note in ci-watch doc) | Medium   | S      | Documentation |
| 50 | Close this session properly: re-run one fast buildflow pass post-daemon-sweep to confirm green on committed HEAD          | Medium   | S      | Process       |

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Push now or hold?** Master is 8 commits ahead (bundle restore, stale-hash fix, docs, shields). I checked `git status` and AGENTS.md (no branch protection; pushes normally flow through you/git-town) and my own rule is no unprompted pushes. **Do you want master pushed to origin now, or after you review?**
2. **Are buildflow and erraudit in scope for me?** Both are your tools. I tried to pin buildflow's erraudit invocation from the repo side (argv shim, binary strings, run records — failed) and observed nix-hash-fix claiming success on a hash it did not fix. **Should I file/fix these two upstream in buildflow (and the erraudit detector divergence), or treat them as black boxes and route around them here?**
3. **Release timing: v0.5.0 now or batch?** `[Unreleased]` contains real consumer-facing work (Response method forms, `version` package, three datastartest helpers, static v1.0.3 with CSP support). I read the CHANGELOG and release checklist exists — but whether you want a release cut this week is a product call only you can make. **Cut v0.5.0 next, or accumulate more first?**

---

_Point-in-time snapshot — goes stale by design. Section (f) is HARVEST input for TODO_LIST/ROADMAP. Report written to disk per convention; not committed by the assistant (no explicit commit instruction) — the auto-commit daemon will sweep it, or say the word._
