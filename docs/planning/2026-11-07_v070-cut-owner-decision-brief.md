# v0.7.0 Cut — Owner Decision Brief

> **Status:** BRIEF (four decisions await the owner; every executable prep
> step around them is done or in flight). Evidence collected 2026-11-07,
> HEAD `ff4f51e`-era master, all probes run live this session.

**Purpose:** one place with the evidence, options, and a recommendation for
each decision blocking the v0.7.0 cut (TODO_LIST rows: Cut v0.7.0, release
lane, cache policy, push timing). The recommendations are mine; the calls
are the owner's.

## 0. The v0.6.3 incident (new context every decision inherits)

Discovered and verified 2026-11-07 while prepping this cut:

- `v0.6.3` was tagged 2026-11-04 on `09fe522` (toolchain pin 1.27.3) —
  **root module only**. `git ls-remote` shows no `static/v0.6.3`,
  `broadcast/v0.6.3`, `datastartest/v0.6.3`; a live proxy probe returns
  `unknown revision` for all three submodule `@v0.6.3` paths.
- Checklist §2 was skipped: the tagged root `go.mod` still requires
  `static v0.6.1` (siblings at their v0.6.2 pins at the tag). Root-only
  consumption resolves fine — the release is *functional but wrong-shaped*:
  lockstep `go get .../static@v0.6.3` fails by design gap.
- Checklist §2's CHANGELOG promotion was skipped too: the tag shipped with
  its entries under `[Unreleased]`; master had **no `[0.6.3]` section at
  all** until the 2026-11-07 backfill (CHANGELOG now records the incident
  inline in that section).

Nothing needs emergency action — root v0.6.3 consumers are unbroken — but
every decision below is now made against this backdrop, and the v0.7.0 cut
is the fix-forward vehicle for the full checklist (§2 requires, §2.5 nix
gate, §3 four-tag lockstep, §4 proxy + release verification).

## D1 — Remove `ScriptHandlerWith`'s dead version parameter? (breaking)

**Evidence.** `script_handler.go:38-42`:

> The version parameter is unused: it is kept only for signature
> compatibility with the v0.5.x API. Removing it is a breaking change
> pencilled in for v0.7.0 (owner-gated); pass any value — static.Version is
> the conventional choice.

The parameter is blank-identified (`_ string`) — dead in the body, proven
by the doc comment itself.

**Options.**
(a) **Remove at v0.7.0** (recommended). The godoc has pencilled exactly
this since v0.6.2; the repo is pre-1.0 where minor bumps may break with a
migration note; the parameter has no semantic future (a *versioned custom
bundle* would need a real version, not a dead one). Migration is one-line
for callers (`ScriptHandlerWith(b, static.Version)` → `ScriptHandlerWith(b)`),
and it lands in the migration guide.
(b) Keep until v1.0. Zero functional cost — it stays a blank param forever.

**Recommendation: (a) remove at v0.7.0**, with the CHANGELOG `Changed`
entry and a migration-guide bullet written at the cut.

## D2 — Release lane for `ScriptTagVersioned`

**Evidence.** `v0.6.3` is taken (root-only, published 2026-11-04). Re-using
or moving the tag means rewriting a proxy-published version — the fleet
de-poisoning history (AGENTS.md, 2026-10-03) exists precisely because that
poisons consumers. `ScriptTagVersioned` is a new exported API → semver
minor.

**Options.**
(a) **v0.7.0, full lockstep** (recommended). Four annotated tags on one
commit, §2 require bumps (root → `static v0.7.0`; broadcast/datastartest →
root + static), §2.5 nix gate, §3 exact-ref push, §4 verification. The
CHANGELOG `[Unreleased]` (`Added: ScriptTagVersioned`) renames to
`[0.7.0]` at the cut.
(b) v0.6.3 re-tag — **ruled out** (poison).
(c) Root-only v0.7.0 like v0.6.3 — ruled out; it would double the defect.

**Sub-decision (v0.6.3 lockstep repair):** retro-tag
`static/v0.6.3` + `broadcast/v0.6.3` + `datastartest/v0.6.3` on the same
commit `09fe522` (technically clean: static is dependency-free; the sibling
go.mods at that commit require root `v0.6.2`, a real published version, so
MVS resolves), **or** fix forward at v0.7.0 and leave v0.6.3 root-only
forever (documented in CHANGELOG). **Recommendation: fix forward** — no
consumer is known to need submodule `@v0.6.3`; retro-tags add published
surface nobody asked for. Retro-tag only if someone reports the gap.

## D3 — Cache policy: raise ScriptHandler to `immutable`?

**Evidence.** `script_handler.go:55` —
`Cache-Control: public, max-age=86400` (24h). `ScriptTagVersioned`'s godoc
(`script_handler.go:88-96`) sells versioned URLs as the stale-window fix.
`docs/static-js.md` documents the current policy.

**Options.**
(a) **Keep 86400 on ScriptHandler; document the versioned+immutable
pairing** (recommended). Raising the *handler default* to
`max-age=31536000, immutable` would give bare-URL consumers (`ScriptTag`,
direct mux mounting — the pre-versioning majority pattern) a **one-year
stale window with no cache key**: an upgrade would not reach browsers for
a year. The safe immutable story is per-consumer: serve versioned URLs,
then (optionally) override the header at the mux. A handler option (e.g.
`ScriptHandlerWithCacheControl`) can be added later on demand — additive,
no decision needed now.
(b) Raise the default to immutable. Only safe *after* bare-URL usage is
extinct — unknowable for a public library.

**Recommendation: (a)** — default unchanged; add a short
"versioned URL + immutable override" recipe to `docs/static-js.md` at the
cut (or now; it is docs-only).

## D4 — Push/publish timing

The cut is a sequence (§1 gates → §2 requires → §2.5 nix → §3 tags → push →
§4 verify → §5 upstream check), estimated 60–90 min end-to-end once D1–D3
are answered. **Recommendation:** run it in one session with the owner
available for the two §3 moments that want eyes (tag push, release-create
order — root LAST so it owns "Latest"). No other timing constraint exists;
the longer `[Unreleased]` sits, the more the release-note diff grows, but
nothing rots.

## Appendix A — v0.7.0 release-note draft (CHANGELOG excerpt, cut-ready)

> ## v0.7.0
>
> ### Added
>
> - **`ScriptTagVersioned(path)`** — cache-busting `<script>` tag: appends
>   the pinned client version as `?v=` (`&v=` when the path already has a
>   query; fragments preserved). Each bundle release becomes a distinct
>   browser cache key, collapsing the 24h stale-JS window that
>   `Cache-Control: public, max-age=86400` otherwise leaves open after an
>   upgrade. `ScriptTag` keeps its verbatim-concatenation contract;
>   `ScriptHandler` is unchanged.
>
> ### Changed
>
> - *(D1 pending)* `ScriptHandlerWith` drops its dead version parameter —
>   `ScriptHandlerWith(bundle, static.Version)` becomes
>   `ScriptHandlerWith(bundle)`. Blank since v0.5.x; no behavior change.
>
> ### Release-shape note
>
> - v0.6.3 was a root-only cut (submodule lockstep tags were never
>   created; its CHANGELOG section was backfilled 2026-11-07). v0.7.0 is
>   the first full four-module lockstep release since v0.6.2 — update all
>   four modules together.
>
> **Licensing:** all four modules (root, `broadcast`, `datastartest`,
> `static`) ship MIT `LICENSE` files; holder of record is Lars Artmann
> (the 2026 root-LICENSE holder-name typo was corrected in the 2026-10-01
> tranche; all four headers verified identical on 2026-11-07).

## Appendix B — emeet-pixyd PR #12 (resolved this session)

PR #12 ("fix(web): add version query to datastar.js for cache busting",
Alex-Craftsman) is **MERGED** in `LarsArtmann/emeet-pixyd` — the TODO row's
"review/merge" action is done. Residual: the generated
`templates_templ.go` carries `// templ: version: v0.3.1100` while the local
CLI is v0.3.1114 — either regenerate under the repo's pinned templ or pin
templ at v0.3.1100; one 5-min check in that repo (its flake decides which
side is canonical). This repo's own PR #12 (an unrelated closed dependabot
nix bump) is NOT this PR — TODO_LIST now says so explicitly.

## Appendix C — §2.5 rehearsal (live run, 2026-11-07)

Simulated §2 on a detached worktree (`/tmp/gd-rehearsal`, requires bumped
to a hypothetical v0.7.0 lockstep set — the §2 lockstep grep verified root
→ static, broadcast/datastartest → root (+ static for datastartest)), then
ran the §2.5 gate. **The gate correctly went red on the post-bump tree**:
broadcast's and datastartest's `go-modules` FODs failed (vendor FOD
builder exit 1), while root's vendor derivation did not move — its
external requires are untouched by a sibling-require bump, matching the
AGENTS.md rule that root `vendorHash` moves only on external-require or
toolchain changes. The rehearsal's failure was partly artificial (§2's
`go mod edit` alone leaves go.sum lagging; a real cut pairs it with a
tidy), but the gate contract is proven: §2.5 fires red before anything can
reach §3. Worktree discarded, no tag created.
