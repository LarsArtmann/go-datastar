# Status: Execution Session — §1 gap closing, broadcast heartbeat API, codespell gate wiring

**Scope line — what this report VERIFIED (with which command/artifact):**

- Session-start ritual ran: `git town status` (clean), `git status` (clean at start), `gh pr list`
  (#29 open, dependabot), `gh run list --branch master --limit 5` (actionlint green ×4).
- Prior-session claims verified before reuse: the 06-45 report's "erraudit ×4 CI-mode ran" claim was
  re-executed live, not trusted; M09.1's landed test was run, not assumed; ledger row 32's
  "done" disposition was read from source before acting on 06-36's pointer to it.
- Every gate number below is a live run at this session's HEAD, not an inherited claim.

## a) FULLY DONE

1. **Release-checklist §1 legs closed (the gap the 06-36 self-review owned).**
   erraudit CI-mode ×4 (`--severity-threshold error`, per-module, GOWORK=off): 0 violations,
   exit 0 ×4 — and re-run post-changes. `nix run .#govulncheck`: "No vulnerabilities found"
   across all four modules.
2. **Broadcast replay-filter interaction test** (TODO row 26, carved out of the storecontract
   row): `TestBroadcasterReplayFilterInteraction` — one predicate serves both the live
   promoted `SubscribeFilter` and reconnect `sse.ReplayFiltered` against an injected store,
   pinning MemoryStore's EventsAfter + post-filter fallback. -race PASS. Row struck.
3. **Broadcast API follow-ups** (TODO row 27): `WithHeartbeatDisabled()` (absolute switch —
   go-sse's `time.NewTicker` panics on non-positive durations, so a sentinel interval was
   never viable) and `NewBroadcasterFromHub(hub, opts...)` (variadic, source-compatible;
   heartbeat + replay store apply, buffer size documented hub-owned). Decisive tests:
   `TestBroadcasterHeartbeatDisabled` (enabled twin at 10 ms proves the window) and
   `TestNewBroadcasterFromHubWithOptions` (heartbeat + replay through a shared hub). README
   prose + API table + CHANGELOG [Unreleased] updated. Row struck. -race PASS.
4. **M09.1 verified**: `TestIndexHandlerCarriesVersionedScriptTag` (example/index_test.go)
   was landed by a parallel session; run at this HEAD — PASS.
5. **M10.3 actually executed** (the 06-36 "checked in my head" admission): greps ran for
   real — zero doc references attribute `ScriptTagVersioned` to `static/`; all references
   correctly name the root package (README:157, migration-guide "(root)", FEATURES → root
   script_handler.go); broadcast/README has zero mentions.
6. **M16 dispositioned as a source spot-check** (rows 1, 5, 13, 20, 29 of the 09-02 pointer
   ledger): every sampled source item carries an inline verdict matching today's code truth;
   the ledger is a pointer index (dispositions live at the sources), so the full 45-row
   re-derivation was deliberately not performed — evidence stamped into the ledger comment.
7. **Dockerfile smoke via native simulation**: no container runner exists (podman/docker/
   nerdctl/buildah absent; flake ships no container output), so the Dockerfile's exact
   build sequence was simulated (minimal context → `GOWORK=off go mod download` →
   `go build ./example`), and the built binary served HTTP 200 on `/` with
   `src="/datastar.js?v=1.0.4"` — the `compose up` behavior, verified without a daemon.
8. **PR template repaired** (06-45 §c.1): dead `GOEXPERIMENT=jsonv2` removed from both
   commands, `broadcast/` added to the per-module isolation list, lint line now names
   `nix run .#lint-ci` (exact CI parity) instead of a bare unpinned `golangci-lint run`.
9. **static-js.md upgrade ritual** (06-45 §c.2): step 3 added — refresh version-coupled
   surfaces (`ExampleScriptTagVersioned`'s literal Output moves with `static.Version`);
   pre-flighted SAFE before writing.
10. **Fuzz-target count reconciled** (06-45 §c.3): code has six `func Fuzz*` targets;
    docs/testing.md said "Four" — corrected to six with the accurate scope phrase.
11. **Codespell gate decision landed** (buildflow ledger row 28): wired into treefmt as a
    custom formatter (`command = pkgs.codespell`, `--write-changes`, includes for
    go/md/nix/yml/yaml/json/sh/templ, excludes for CHANGELOG + status/planning docs +
    vendored bundle) + devshell package. The investigation found and fixed a REAL shield
    bug: `.codespellrc` walk-mode skip never matched (directory scans match `./`-prefixed
    paths, so the exact `static/datastar.js` pattern let 32 bundle hits through; `*/…`
    wildcard = 0, verified with codespell 2.4.3). `keep-alives` plural added to the ignore
    list (verified false positive in SSE context). Corpus verdict: zero genuine typos on
    live surfaces — the gate wired in already green.
12. **Long-tail decisions recorded** (TODO row 29): go-licenses stays OUT of ci.yml
    (devshell GOROOT constraint per AGENTS; buildflow owns it locally); datastartest
    coverage chase declined (95.6% healthy; remainder is defensive error paths; coverage
    not gated per ADR 005).
13. **buildflow doctor run** (06-36 §f-18): 68 pass / 5 warn / 30 fail — all 30 failures
    are uninstalled non-Go ecosystem tools a Go-only repo never invokes; warnings
    informational. Nothing repo-specific broken.
14. **F/spacing third occurrence handled per the watch rule**: line 19 of the archived
    08-47 report re-fixed; git log -S shows the damage class landed via daemon commits
    `4539949`, `2e05692`, `ff4f51e`, `da3e0e3` — evidence recorded in AGENTS (a parallel
    session's markdown pass is the source; hunting it needs owner coordination). New rule
    recorded: never quote the damaged byte pattern verbatim in living docs — the F/spacing
    detector flags the example itself (it caught my own AGENTS draft within one battery run).
15. **Wrap battery, all green at HEAD**: race ×4 (8 pkgs), vet, pinned lint v2.14.0
    fresh-cache ×4 = 0 issues, erraudit CI-mode ×4 = 0 violations, govulncheck clean, fuzz
    smokes ×2 PASS, CI-parity set (work-sync idempotency, use-match, replace audit,
    tidy-diff ×4, mod verify ×4), isolation builds ×4, `nix flake check --keep-going`
    ALL PASSED (vendorHashes unmoved, as the AGENTS rules predicted), docs-gates ×5 PASS.

## b) PARTIALLY DONE

1. **M16** — a 5-row source spot-check, not 45 verdicts; deliberate (the archive contract
   already carries per-row dispositions at the sources), recorded as such in the ledger stamp.
2. **Dockerfile smoke** — the build sequence and serving behavior are proven natively; a true
   image build remains env-limited (no runner on this machine).
3. **The daemon-commit tag-history concern** (g1) — no curated commits were made this session
   either; everything rides the daemon, so the v0.7.0 tag would still point at auto-commit
   history. Owner decision pending.

## c) NOT STARTED

1. Owner-gated items remain untouched: v0.7.0 cut (D1–D4), policy Q1–Q8, g1–g3 questions,
   PR #29 disposition, branch deletions, retro-tag vs fix-forward, upstream filings
   (BuildFlow ×4, docs-gates ×3 — each still needs its source pass in the target repo).
2. docs-gates battery wired into a repo gate entry point (06-36 §f-30) — infra work, left
   to its filer's lane.
3. Example/godoc polish tranche (TODO row 28): `Version()` godoc Example, store-injection
   demo handler, flake ldflags wiring.
4. Upstream SDK docs cache-busting check (TODO tranche A row's last open leg) + fleet sweep
   for bare-`datastar.js` serving.

## d) TOTALLY FUCKED UP

1. **Wired a treefmt option that does not exist.** `programs.codespell.enable` isn't in this
   treefmt-nix — `nix flake check` failed at eval with a loud option error. Corrected to the
   custom-formatter shape (`settings.formatter.codespell`), then hit a second cycle because
   treefmt requires explicit `includes`. Two flake-check cycles spent learning the right
   shape; one `nix eval`/check BEFORE writing would have cost one.
2. **My new test code failed the lint it must pass** — golines long lines ×2, varnamelen
   single-letter `n`, wsl_v5 missing blank lines above `go func()` ×2. The battery caught
   all of it; four fix cycles instead of writing formatter-clean the first time.
3. **Created the exact F/spacing false positive I was documenting** — quoting the damaged
   byte pattern verbatim in the new AGENTS gotcha tripped the annotate gate one run later.
   The rule is now in AGENTS: describe the damage class, never byte-quote it.
4. **One wrong-file edit attempt** (broadcaster.go content aimed at broadcaster_test.go) —
   refused by the mod-time guard, zero damage, but the file mixup happened at all.
5. **Two hook blocks for the same `rg -rn` trap** — the rule was in the handoff notes and
   still got hit twice this session.

## e) WHAT WE SHOULD IMPROVE

1. **Verify option existence before wiring**: `nix eval` the option path (or grep the
   treefmt-nix program list) before writing `programs.<x>.enable` — loud failures are
   cheap, but two eval cycles for one wiring is two too many.
2. **Write test files formatter-clean**: <120-char lines, blank line before goroutine
   launches, no single-letter names — the linter suite (golines/varnamelen/wsl_v5) has
   stable, known preferences; pre-empting them saves the whole fix cycle.
3. **The docs-gates detector parses file text, not intent**: any doc that quotes a damaged
   pattern verbatim self-trips. Examples must be paraphrased or marked up so the literal
   bytes never appear.
4. **Verify parallel sessions' claims by running, not reading**: the erraudit re-run proved
   the 06-45 claim true, but the class ("trust by reading") is the same one that produced
   M10.3's fake verification. Run the command; cite the run.
5. **Session-start ritual first, always**: this session ran it before anything else and the
   parallel-session coordination (three landed reports, one live formatter pass) was
   visible from minute one.

## f) Up to 50 things we should get done next

> Brainstorm, not commitment; owner-gated marked 👤. Verify against code before executing.

1. 👤 Cut v0.7.0 per the decision brief (D1–D4 answers pending; §2.5 rehearsal done
   2026-11-07; release-checklist §1 now fully green — the cut is unblocked technically).
2. 👤 Answer g1 (tag history: curated commits vs daemon soup — the concern is now two
   sessions old and every new session adds daemon commits).
3. 👤 Answer g2 (MemoryStore ID-assignment policy — affects storecontract doc permanence).
4. 👤 Answer g3 (which upstream filings to make; each needs a source pass first).
5. 👤 Policy consolidation Q1–Q8 (lanes, CHANGELOG policy, website trigger, Monitoring tier,
   lint cache, required checks, erraudit flip, daemon-on-static).
6. 👤 PR #29 (dependabot codeql bump): merge, stack, or dismiss.
7. 👤 Identify/stop the parallel session's markdown formatter pass that keeps damaging the
   archived 08-47 report (evidence: daemon commits 4539949, 2e05692, ff4f51e, da3e0e3).
8. 👤 Keep or revert the codespell treefmt wiring (landed green this session; owner's gate
   shape, their call).
9. Wire the docs-gates battery into a repo gate entry point (06-36 §f-30; flake input +
   check, ~20 lines).
10. Example/godoc polish tranche (TODO row 28): `Version()` godoc Example, store-injection
    demo handler, flake ldflags stamp.
11. Upstream SDK docs cache-busting recommendation check (tranche A's last leg), then
    strike the tranche A row.
12. Fleet sweep for bare-`datastar.js` serving (ROADMAP theme 2; emeat-pixyd already fixed
    upstream via merged PR).
13. Consider `@query`/QUERY-verb support per the 07-04 report §f-6 (verify upstream
    semantics first; generator + helper + goldens if it lands).
14. Persist the verified client-capability inventory (07-04 §f-1/2/5): docs section +
    plugin-boundary gotcha + bundle plugin-name golden test.
15. Re-verify 2–3 inherited "done" claims from today's reports against source (07-04 §f-23;
    process-discipline claims are the class most often falsely asserted).
16. Probe golangci-lint versions newer than the pinned 2.14.0 for go1.27 export-data health
    (07-04 §f-27).
17. Verify static bundle provenance end-to-end: `static/fetch-bundle.sh 1.0.4` checksum vs
    committed bytes (07-04 §f-29).
18. LICENSE byte-identity check across root + 3 submodules (07-04 §f-30).
19. docs/error-system.md code table vs errors.go catalog parity spot-check (07-04 §f-33).
20. Refresh docs/performance.md benchmark numbers at current HEAD (07-04 §f-36).
21. Sweep ROADMAP for silently-completed items (07-04 §f-39).
22. Backlog hygiene: strike TODO rows older than one week with evidence or re-rank.
23. Archive the executed M01–M25 Pareto plan if still unarchived (07-04 §f-41).
24. M16 full 45-row re-derivation IF the owner wants the ledger self-contained rather than
    pointer-indexed (this session deliberately skipped it; rationale in the ledger stamp).
25. Consider `datastartest` replay-filter helper exposure (this session's interaction test
    could graduate into a reusable `AssertReplayFilterComposition` if consumers need it).
26. Heartbeat: consider exposing effective-interval introspection (Health or a getter) if
    consumers ask; today the config is write-only.
27. Renovate: demonstrate the custom manager sees upstream releases (07-04 §f-9; zero PRs
    ever; Settings API 403s from here).
28. Verify `go.work` gitignore status on this machine (`git check-ignore` ritual; 07-04
    §f-22).
29. aarch64/darwin `nix flake check --all-systems` (needs a darwin runner; ROADMAP theme 3).
30. Bundle-integrity checksum API idea (ROADMAP theme 5).

## g) Questions I cannot figure out myself (max 3)

1. **The recurring formatter pass**: which parallel session/tool keeps stripping spaces
   around inline code in `docs/status/archived/` (four daemon commits cited in §a.14)? It
   re-damages a fixed line and burns battery cycles every time — can you identify and stop
   or fix it? (This session only fixed + documented; the tool is outside my reach.)
2. **Codespell in the canonical gate**: keep the new treefmt wiring (fixes typos on `nix
   fmt`, gate-owns the class, wired in green), or prefer on-demand-only? My recommendation:
   keep — the shield-bug fix alone justified the pass.
3. **Tag history (g1, re-asked)**: for the upcoming v0.7.0 tag, do you want curated commits
   (a squash/rewrite of the daemon soup before tagging) or is pointing the tag at
   auto-commit history acceptable? Every session without an answer adds more soup.

---

_Awaiting instructions._
