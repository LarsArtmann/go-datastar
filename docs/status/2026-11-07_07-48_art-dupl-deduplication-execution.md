# art-dupl deduplication execution — 2026-11-07 07:48

**Scope.** This report covers ONLY this session's run: the art-dupl
deduplication task started from the user's pasted clone report (~07:00).

Verified here, with the artifact named:

- art-dupl actionable clone groups **7 → 3**, re-run twice with the user's
  exact flags (`art-dupl --sort total-tokens -t 1 --suggest-generics --timing
  --rich-text`); final verdict `Detected 299 clone groups, 3 shown` — the 3
  are exactly the baseline-accepted groups.
- Workspace race suite green **×2** (`go test ./... ./broadcast/...
  ./datastartest/... ./static/... -race -count=1`, all 8 packages `ok`,
  ~07:13 and ~07:47). `go vet` green (×4 module patterns).
- CI-parity lint (golangci-lint v2.14.0, pinned go-run): **0 findings in
  every file this session touched** after one fix; the command still exits 1
  with 4 findings that are ALL in `broadcast/broadcaster_test.go` — last
  touched by daemon commit `2f25449` (07:19:52), NOT this session (see b1).
- Vendor hashes: `buildflow -s nix-hash-fix --fix` (cache-bypassed) reported
  `2 fixed, rewrote: flake.nix`, verify `hashMismatches=0`; flake.nix
  committed by the daemon as `af6ffbd` (07:17).
- Hermetic gate: `nix flake check --keep-going` → **all checks passed**
  (~07:48), derivations stamped at rev `78ba2c1` (post-fix tree; includes
  treefmt over the new `options.go`, `docs/duplication-baseline.md`, and the
  AGENTS.md row).

NOT verified here:

- CI runs on `origin/master` for the daemon-pushed commits
  (`33f1067..78ba2c1`) — `gh run list` not run this session (see d3).
- Any fix for the 4 `broadcaster_test.go` lint findings — another session's
  in-flight work, deliberately untouched.
- Fuzz smoke (`FuzzReadEvents`/`FuzzReadSignals`, 30s) not run after the
  parsing-internal refactor (see c1) — e2e, golden, and race suites all green
  over the changed code, so risk is low but the check is real and was skipped.
- The stale gopls "unused: applyOptions" diagnostic never visibly cleared in
  the LS panel; compiler-side it is disproven (vet/build/tests/lint all pass
  and `response.go` calls it). Treated as the documented LS false-positive
  class, not chased.

## a) FULLY DONE

1. **All 7 actionable clone groups read and judged** — every flagged site
   opened and classified harmful vs intentional (4 extract, 3 accept).
2. **Root options-fold unified**: new `options.go` with unexported
   `applyOptions[T any, O ~func(*T)]`; the five identical fold loops in
   `NewElementsPatch` (elements.go), `NewScriptPatch` (script.go),
   `NewSignalsPatch` (signals.go), `NewDispatchCustomEventPatch`
   (script_convenience.go), and `Response.PatchSignals` (response.go) now all
   route through it. The `~func(*T)` constraint accepts the named option
   types without API changes.
3. **`CustomEventName` now delegates to `quotedSuffix`**
   (datastartest/script_accessors.go) — it was a literal re-implementation of
   the marker-cut-then-unquote helper that existed 30 lines above it.
4. **Script open-tag parsing single-sourced**: new `Event.scriptTagParts()`
   shared by `ScriptContent` (event.go) and `ScriptAttributes`
   (script_accessors.go); each had its own copy of CutPrefix + indexTagEnd +
   guards.
5. **Signals guard extracted**: `requireSignalsEvent` mirrors the existing
   `requireElementsPatch` convention; `RequireSignals` /
   `RequireSignalsContain` share it, and both now use `evt.IsSignals()`
   instead of re-spelling the type comparison.
6. **Acceptance made a gate, not prose**: `docs/duplication-baseline.md`
   pins the 3 accepted groups (constructor calls, cross-module option folds,
   separator idiom) with rationale, the canonical re-run command, and the
   expected shown-count of 3; AGENTS.md Docs Map row added.
7. **Vendor hashes landed correctly** — first `nix-hash-fix` run was a stale
   cache replay (d1); the `BUILDFLOW_NO_RESULT_CACHE=1` re-run performed the
   real FOD builds and landed both moved hashes (datastartest + broadcast)
   with verify green; committed `af6ffbd`.
8. **One lint regression in my own code fixed** (`nonamedreturns` on
   `scriptTagParts`) — committed by daemon as `78ba2c1`; final lint run
   confirms 0 findings in session files.

Net code effect: ~45 duplicated lines removed across 8 files, ~41 helper
lines added. The win is single-sourcing (four concepts went from N-places-to-
change to 1), not line count — stated plainly because the diff is roughly
LOC-neutral.

## b) PARTIALLY DONE

1. **CI-parity lint gate is red at repo level** — 4 findings
   (`varnamelen` ×1, `wsl_v5` ×3) in `broadcast/broadcaster_test.go:469-512`,
   introduced by daemon commit `2f25449` (07:19:52, another session's edit —
   it appeared between my two lint runs and I confirmed via `git log -1` on
   the file). Root cause within that edit unknown (not mine to diagnose);
   severity: makes the pinned lint command exit 1, and since nothing is a
   required check the signal is informational only. Effort **S** for the
   owning session. Mitigation: none needed for consumers; local gates are
   the real gate.
2. **LSP diagnostics stale all session** — panel kept showing
   `unused: applyOptions` plus 6-7 pre-existing `wrapcheck` warnings the
   pinned linter does not report. Compiler-verdict green throughout; the
   panel vs gate divergence is unexplained beyond "known false-positive
   class". Effort S (one lsp_restart + one config diff read).

## c) NOT STARTED

1. **Fuzz smoke after parsing-internal refactor** — `ScriptContent` /
   `scriptTagParts` sit under `FuzzReadEvents`' coverage; the documented 30s
   smoke (AGENTS.md "Fuzz smoke tests") was not run. Priority: High for
   belt-and-braces, effort S. Not started because I verified via e2e/golden/
   race and only registered the gap during this self-review.
2. **docs-health HARVEST of §f below** into TODO_LIST/ROADMAP — not run;
   user instruction was report-then-wait.
3. **Master CI watch for the pushed range** (`33f1067..78ba2c1`) —
   `gh run list --branch master --limit 5` not run (also d3). Priority High
   (AGENTS.md session ritual), effort S.
4. **Mechanical enforcement of the duplication baseline** — baseline doc is
   committed; nothing runs `art-dupl` and compares shown-count to 3
   (see f4). Planned, not started, awaiting owner decision (g2).

## d) TOTALLY FUCKED UP

1. **Nearly shipped a stale buildflow verdict as green.** The first
   `buildflow -s nix-hash-fix --fix` returned `1 success … 100% hit rate` in
   19 ms with `0 files / 0 B hashed` — a result-cache replay keyed on
   flake.nix (unchanged), blind to the root `.go` edits that were the actual
   cause of the hash movement. Trusting it would have left master with
   moved-hash FODs and a red nix CI (the daemon pushes everything; nothing
   blocks it). Caught only by reading the output skeptically — the AGENTS.md
   gotcha ("a finding can replay after its cause was fixed elsewhere")
   saved it, not process. Mitigation applied: `BUILDFLOW_NO_RESULT_CACHE=1`
   re-run, 3m15s, hashes landed. Root cause is a buildflow cache-key gap
   (vendored-module source not in the key) — upstream candidate (f5).
2. **Wrote code style against the repo's lint config.** `scriptTagParts`
   shipped with named returns; `nonamedreturns` (enabled in this repo)
   caught it only in the CI-parity run — AFTER the 3-minute hash-fix job had
   already validated the tree. Got lucky: the fix touched datastartest's own
   source, which feeds no vendorHash, so no second FOD pass was needed
   (proven by the flake check at `78ba2c1`). The process fault stands: I
   never read `.golangci.yml` before writing Go in this repo, and I ordered
   the expensive gate before the cheap one.
3. **Skipped the documented session-start CI probe.** AGENTS.md: probe
   master CI health before trusting a clean local tree. I dove into the
   pasted art-dupl output instead. Did not bite this time; the report at
   07-04 says CI was 5/5 green then, but I did not verify even that —
   remembered context, not evidence.
4. **Did not think of the fuzz smoke until this report.** The refactor moved
   real parsing logic (`ScriptContent` internals) and I treated
   unit/e2e/golden/race as sufficient. They probably are — but the repo
   documents a fuzz smoke for exactly this surface and it cost 30 seconds.
   Registered as c1; the miss is on me.

## e) WHAT WE SHOULD IMPROVE

1. **Gate order: cheap-and-critical first.** Run the pinned lint (and vet)
   BEFORE any FOD/hash work; a post-hash code edit risks a second 3-minute
   FOD pass. Cost this time: zero (luck); the rule costs nothing.
2. **Read the repo linter config before writing code in a repo you haven't
   linted in.** One `view .golangci.yml` beats an edit-fix-commit cycle.
3. **buildflow cache-key gap is upstream-fixable** — nix-hash-fix diagnosis
   should hash the vendored modules' source or always bypass the result
   cache. File it (f5) rather than re-documenting the workaround.
4. **The baseline should be enforced, not just written** — the go-etag
   pattern (committed accepted-group count as a gate) is proven; a fifth
   group appearing should fail something (f4, needs owner decision because
   art-dupl is a private tool — probe-gate pattern already exists for
   erraudit).
5. **LS-vs-gate diagnostics divergence deserves one doc line** — the panel
   showed wrapcheck/unused noise all session while the real gate reported
   none of it; AGENTS.md documents the json/v2 class but not this. Small doc
   addition, prevents future sessions "fixing" non-findings (f8).
6. **Noticed in buildflow preflight, unfixed:** AGENTS.md is 352 lines
   against a 220 max (warns on every run); formatter identity drifted
   (buf/cargo/templ upgraded) so cached format verdicts need one
   `buildflow format` re-verify; dprint fetches 4 remote plugins (cold-run
   hang risk). None block anything; all are one-command fixes (f11-f13).

## f) Next tasks

| #   | Task                                                                                                      | Impact   | Effort | Category      |
| --- | --------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1   | Fix the 4 lint findings in broadcast/broadcaster_test.go (varnamelen, wsl_v5 ×3, commit 2f25449)          | High     | S      | Bug           |
| 2   | Run the 30s fuzz smoke (FuzzReadEvents + FuzzReadSignals) over the refactor                               | High     | S      | Quality       |
| 3   | `gh run list --branch master --limit 5` — verify CI over pushed range 33f1067..78ba2c1                    | High     | S      | Ops           |
| 4   | Decide + wire mechanical baseline gate: art-dupl shown-count must equal docs/duplication-baseline.md (3)  | Medium   | S      | Quality       |
| 5   | File buildflow issue: nix-hash-fix result cache replays stale verdicts when only vendored source changed  | Medium   | S      | Upstream      |
| 6   | docs-health HARVEST this report's §f into TODO_LIST/ROADMAP                                               | Medium   | S      | Documentation |
| 7   | Share `"<script"` prefix as a named const between IsScript and scriptTagParts (2 literal sites)           | Low      | S      | Cleanup       |
| 8   | Add one AGENTS.md line on LS-vs-gate diagnostics divergence (wrapcheck/unused panel noise)                | Low      | S      | Documentation |
| 9   | `buildflow format` once to re-verify cached format verdicts under the drifted formatter identities        | Low      | S      | Quality       |
| 10  | Point dprint.json at local .wasm plugins (kill remote fetch / cold-run hang risk)                         | Low      | S      | Tooling       |
| 11  | Trim AGENTS.md under the 220-line warn threshold (move detail into docs/, keep the map row links)         | Low      | M      | Documentation |
| 12  | Raise or scope buildflow's 1m performance budget for nix-hash-fix runs (3m15s is expected, not a trend)   | Low      | S      | Tooling       |
| 13  | lsp_restart + confirm the unused/wrapcheck panel warnings clear or document them as the known class       | Low      | S      | Tooling       |

Stopped at 13 deliberately: the instruction was to report only this
session's run — every item above was noticed in it. Padding to 50 would mean
inventing unrelated backlog, which HARVEST would rightly reject as noise.

## g) Questions

1. **Who owns the broadcaster_test.go lint findings?** They arrived via
   daemon commit `2f25449` (07:19:52) while this session ran; I can fix them
   in ~10 minutes on your signal, but another session may be mid-edit there
   (shared checkout) — say the word and I'll take it, or leave it to them?
2. **Should the duplication baseline become a mechanical gate now?** It
   needs your call because art-dupl is a private tool: CI already
   probe-gates private binaries (erraudit pattern), so wiring is feasible —
   want it in ci.yml, as a pre-push local gate, or manual-audit-only for now?
3. **Export policy for `applyOptions`?** It stays unexported by design (not
   protocol API), which permanently leaves the two module-local option folds
   (broadcast + datastartest) as the last accepted cross-module clone group.
   Accept as documented, or do you want a tiny shared internal module to
   kill that group too?

---

*Point-in-time snapshot; report-then-wait per instruction. Auto-commit daemon
will pick this file up; no manual commit (harness contract).*
