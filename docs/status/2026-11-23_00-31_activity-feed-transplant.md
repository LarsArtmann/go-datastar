# Status Report — 2026-11-23 00:31 — Activity-Feed Example Transplanted from go-sse

Session scope: receive go-sse's DataStar activity-feed example (owner
decision g1, see go-sse's 2026-11-22 23:48 report) as
`example/activity-feed/` — rewritten on this library's Patch API, keeping
go-sse underneath as the transport. Final gate state: the example package
passes `gofmt`/`go vet`/`golangci-lint` (v2.15.0 local; CI pin v2.14.0)/
`erraudit` (`--severity-threshold error`, exit 0)/`go test -race` (10 tests,
4.2s); workspace `go test ./example/... -race` green for all three example
packages; `nix flake check` convergence for the two sibling vendor hashes
handled (see §b).

**Coverage delta (measured this session):** none applicable — the example
package is outside the coverage gate; its 10 tests pass under `-race`.

## a) FULLY DONE

| # | Work                                                                                                                                                                                                                                                                                                                                                                                                   | Evidence                                                                                                                                                                                                                                                                                                                                                 |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | `example/activity-feed/` created (5 Go files + 2 test files + `VERIFY.md` + ported `static/app.js`, `static/styles.css`): Patch-API producer (`ElementsPatch`/`SignalsPatch` with `WithElementsEventID`), ring-buffer `EventStore`, `SubscribeFilter` alerts view, subscriber-count/replay/total signals, heartbeat, CORS, graceful shutdown, `ScriptHandler` + `ScriptTagVersioned` for the pinned JS | `example/activity-feed/`; event names identical to the old hand-written ones (`datastar-patch-elements`/`-signals`), so the ported UI needed zero client changes beyond branding/JS-serving                                                                                                                                                              |
| 2 | Filter predicate re-derived for the Patch API: category rides in the elements HTML (`feed-item--<category>` CSS marker) instead of a custom dataline the Patch API cannot emit                                                                                                                                                                                                                         | `producer.go` (`feedItemCategoryClass`), `server.go` (`subscribeForRequest`), pinned by `TestFilterPredicate_AlertsOnly`                                                                                                                                                                                                                                 |
| 3 | All gates green on the new package: `gofmt -l` empty, `go vet` clean, `golangci-lint run ./example/activity-feed/...` 0 issues, `erraudit … --severity-threshold error` exit 0, `go test -race -count=1` ok 4.165s                                                                                                                                                                                     | gate runs 2026-11-23 01:1x–01:3x; lint fixed 13→0 issues (varnamelen renames `wg`→`waitGroup`, `s`→`server`, `ch`→`events`, `n`→`replayed`, `w`→`responseWriter`; wsl_v5 blank lines); erraudit fixed by wrapping the three signal-builder errors with `errorfamily.Wrapf(err, Rejection, datastar.CodeSignalsMarshalFailed, …)` instead of `fmt.Errorf` |
| 4 | Docs wired: root README "Full working examples" section, `example/README.md` re-titled to describe all three demos (and its stale "MemoryStore replay" claim about the minimal example — never implemented — corrected), `.gitignore` binary pattern, CHANGELOG `[Unreleased]` Added row                                                                                                               | `README.md`, `example/README.md`, `.gitignore`, `CHANGELOG.md`                                                                                                                                                                                                                                                                                           |
| 5 | `nix fmt` run twice (tree-formatter over Go/JS/Markdown); two files of PRE-EXISTING format drift (`static/static.go`, `static/static_test.go`, introduced in f909928 without a formatter run) were brought to canonical form                                                                                                                                                                           | `git diff` reviewed before keeping; daemon committed the sweep in `249faf9`                                                                                                                                                                                                                                                                              |
| 6 | Pre-existing red found and fixed: `TestTransportPinsRecordedInChangelog` failed because go.mod pins `go-datastar/static v0.8.0` while CHANGELOG never mentioned it (bump predates this session; masked until the vendor hashes let the test phase run)                                                                                                                                                 | `nix log` of the failing root drv; CHANGELOG `[Unreleased]` Changed row added; full root suite green after (see §b for the final flake state)                                                                                                                                                                                                            |
| 7 | `broadcastVendorHash` + `datastartestVendorHash` updated for the tree change (their FODs vendor the replaced root tree docs-included, so ANY tracked-file edit moves them — the movement rule documented in the flake comment)                                                                                                                                                                         | `flake.nix`; verified by a `--keep-going` run reporting zero hash mismatches after the last content edit; final exit-coded check in flight at report time (§b)                                                                                                                                                                                           |

## b) PARTIALLY DONE

- **Final `nix flake check` (exit-coded).** Hash mismatches are resolved
  (zero reported after the last content edit), but the full exit-coded check
  was still running when this report was written (DNS-flaky narinfo lookups
  stretch it). What remains: `nix flake check; echo EXIT=$?` → expect 0; if
  the CHANGELOG pin line or any later edit moves the sibling hashes again,
  re-capture and update once more.

## c) NOT STARTED

- Nothing in scope. The go-sse-side removal and reference updates live in
  go-sse's 2026-11-23 00:31 report.

## d) TOTALLY FUCKED UP

1. **Created the example as `example/activityfeed/` first**, then discovered
   `domain-adapter/` sets the dash convention and renamed to
   `activity-feed/` — check sibling conventions before writing files.
2. **Assumed the root `module_boundary_test.go` constraint might not exist**
   and only found it mid-design: the first plan had the example using the
   `broadcast/` module (which would have failed the boundary test and the
   hermetic build). The boundary rule ("root NEVER requires broadcast or
   datastartest") is in AGENTS.md and should have been read before designing.
3. **First flake-check failure was misread as expected hash churn**; the real
   state was hashes moved AND a pre-existing test red hidden behind them.
   `--keep-going` + `nix log` separated the two.

## e) WHAT WE SHOULD IMPROVE

| IMP | Improvement                                                                                                                                                                                                                     | Priority |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| 1   | Read AGENTS.md boundary/module rules before choosing imports for new packages (the broadcast detour would have failed CI)                                                                                                       | High     |
| 2   | When vendor hashes move, run `nix flake check --keep-going` and read test-phase logs in the SAME pass — FOD failures mask test failures behind them                                                                             | High     |
| 3   | The `TestTransportPinsRecordedInChangelog` gap shows dependency bumps can land without changelog lines between releases; the weekly flake-update PR could run the root test suite as a cheap tripwire (it already may — verify) | Low      |

## f) Up to 50 things we should get done next

**P0 — direct follow-ups of this session**

1. Land the exit-coded green `nix flake check` (§b) — the only open gate.
2. Owner: browser-run `example/activity-feed/VERIFY.md` (11 sections,
   including the new ScriptHandler bundle check §11).
3. Cut the tag per the normal release ritual when ready — CHANGELOG
   `[Unreleased]` already carries the Added row (example) and the static-pin
   Changed row.

**P1 — small, bounded**

4. Decide whether the minimal `example/` (raw sse.Broadcaster, no replay)
   should adopt the ring-buffer replay from `activity-feed/` or link to it —
   today its README over-claims are fixed, but the feature gap remains.
5. `example/README.md` mentions `ScriptHandler` serving "the pinned DataStar
   JS" for the minimal demo — verify the `?version=` cache-buster claim once
   in a browser while running VERIFY.md.

**P2 — maintenance / external**

6. go-sse's `KeyedLines`/`SendKeyed` lost their only living example consumer
   (the transplant stopped using them); if this repo's docs cite them as the
   canonical DataStar path, re-point to the Patch API (go-datastar owns the
   wire format now).
