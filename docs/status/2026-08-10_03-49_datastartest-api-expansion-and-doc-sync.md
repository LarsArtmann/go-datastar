# Status Report: 2026-08-10 03:49 — datastartest API Expansion + Doc Sync

> Continuation session: executed the "next steps" from the 02:55 and 02:57
> status reports. Focused on expanding the `datastartest` API surface with
> script-content extraction, non-GET and streaming Collect variants, test
> hardening, and documentation sync (CHANGELOG, FEATURES, AGENTS, doc.go).

---

## What Was Done

### Goal

Complete all actionable items from the two prior status reports:
`datastartest` API gaps (ScriptContent, CollectPost, CollectN, etc.), test
coverage holes (MustReadEvents, Retry/EventID round-trip, empty handler,
parseSSEField edge cases), documentation drift (CHANGELOG, FEATURES, AGENTS,
doc.go), and static-package consistency tests.

### Approach taken

18 tasks, prioritized by impact/customer-value/effort, executed one at a time
with verification after each. All code changes passed golangci-lint (120+
linters, zero issues), `go vet`, and `go test -race` before moving to the next.

### Files created/modified (this session)

| File | Lines | Status |
| --- | --- | --- |
| `CHANGELOG.md` | +27 | MODIFIED — `[Unreleased]` section with datastartest + static entries |
| `FEATURES.md` | +18 -3 | MODIFIED — new "Consumer Test Helpers" section, fixed stale Script Handler refs, fixed e2e line count |
| `datastartest/event.go` | +48 | MODIFIED — `ScriptContent()`, `DataValue()`, `String()` methods |
| `datastartest/collect.go` | +143 -19 | MODIFIED — `CollectWithRequest`, `CollectPost`, `CollectN`, `readNEvents` |
| `datastartest/doc.go` | +30 -19 | MODIFIED — sections for non-GET, streaming, script patches |
| `datastartest/event_test.go` | +180 | MODIFIED — 12 new tests (ScriptContent x4, DataValue, String, Retry/EventID, EmptyHandler, DatalineConstants x8 subtests) |
| `datastartest/reader_test.go` | +49 | MODIFIED — 2 new edge-case tests (multi-colon, empty data) |
| `datastartest/collect_test.go` | 132 (NEW) | NEW — 5 tests (CollectPost, CollectWithRequest, CollectN streaming, CollectN all, failing reader) |
| `response_test.go` | +47 | MODIFIED — `TestStaticVersionConsistency`, `TestScriptHandler_ServesStaticBytes` |
| `AGENTS.md` | +6 | MODIFIED (uncommitted) — API surface table expanded with new exports |

**Total: 9 files, +655 -19 lines**

---

## (a) FULLY DONE

1. ✅ **CHANGELOG.md `[Unreleased]`** — Added "Added" section with
   `datastartest/` subpackage (full feature description including Collect,
   CollectWithRequest, CollectPost, CollectN, ScriptContent, assertion
   helpers, 24+ tests, 4 examples) and `static/` subpackage (Bytes, Version,
   extraction from script_handler.go). Added "Changed" section for
   script_handler.go refactor.
2. ✅ **FEATURES.md "Consumer Test Helpers (`datastartest/`)" section** — New
   section with 8 feature rows: Collect, CollectWithRequest, CollectPost,
   CollectN, ReadEvents, Event accessors (15+ typed), Assertion helpers,
   Filter helpers. All marked `FULLY_FUNCTIONAL`.
3. ✅ **FEATURES.md stale references fixed** — Script Handler section now
   references `static/` subpackage (was `script_handler.go`). Added
   `static.Bytes()` / `Version` row. E2E test line count fixed (260 -> 109).
4. ✅ **`ScriptContent()` accessor** — Strips `<script ...>` wrapper and
   `</script>` suffix from Elements(), returns inner JS. Handles attributes
   in the opening tag (e.g., `type="module"`, `data-effect="el.remove()"`).
   Returns empty for non-script elements patches.
5. ✅ **`DataValue(key)` accessor** — Generic dataline lookup using the key
   prefix (e.g., `"selector "`). Escape hatch for keys without typed
   accessors.
6. ✅ **`Event.String()` debug method** — Returns
   `Event{type=... id=... retry=... datalines=N}` (with id/retry only if
   non-zero). Useful for test failure messages and logging.
7. ✅ **`CollectWithRequest(t, handler, method, body, contentType)`** — Full
   custom HTTP method support. Sets Content-Type header. Uses
   NewRequestWithContext.
8. ✅ **`CollectPost(t, handler, jsonBody)`** — Thin wrapper over
   CollectWithRequest for POST + application/json. The most common non-GET
   DataStar pattern (form submission with signals).
9. ✅ **`CollectN(t, handler, count)`** — Reads exactly `count` events from
   the SSE stream, then returns. Uses internal `readNEvents` which returns
   as soon as the Nth event is dispatched (after the blank line). Does NOT
   wait for EOF. Handles connection close gracefully (scanner error after
   events collected is treated as clean close, not failure).
10. ✅ **`readNEvents(r, count)` internal helper** — Streaming SSE reader
    that returns early at N events. Buffer/cap constants shared with
    ReadEvents.
11. ✅ **doc.go updated** — New sections: "Non-GET requests" (CollectPost,
    CollectWithRequest), "Streaming handlers" (CollectN), "Script patches"
    (ScriptContent). Old "Non-GET requests" section (manual httptest
    boilerplate) replaced with typed helpers.
12. ✅ **AGENTS.md API surface table expanded** — 6 new rows: CollectPost,
    CollectWithRequest, CollectN, ScriptContent, DataValue, String.
13. ✅ **ScriptContent tests** — 4 tests: ExecuteScript (console.log),
    Redirect (URL extraction), ConsoleLog (console.log prefix), NonScript
    (returns empty for regular elements patch).
14. ✅ **CollectPost test** — POST with JSON body `{"name":"alice"}`, handler
    reads signals via ReadSignals, patches back personalized greeting.
    Full round-trip verification.
15. ✅ **CollectWithRequest test** — PUT with nil body + content-type,
    handler validates method, patches response.
16. ✅ **CollectN tests** — Streaming handler sends 10 events in a loop then
    blocks on `<-r.Context().Done()`. CollectN(t, handler, 3) returns
    exactly 3 events with correct content. Plus CollectN_AllEvents for
    synchronous handler.
17. ✅ **Retry/EventID round-trip test** — Patch with
    WithElementsEventID("evt-42") + WithElementsRetryDuration(5000ms),
    verified through full Collect pipeline (not just string input). Event.ID
    and Event.Retry both correct. String() output includes id and retry.
18. ✅ **Empty handler test** — Handler that sends nothing. Collect returns
    empty slice (len == 0).
19. ✅ **parseSSEField edge case tests** — Multi-colon line
    (`data:{"json":"with:colons"}`) correctly takes everything after first
    colon. Empty data field (`data:` with nothing) produces empty string
    dataline.
20. ✅ **Dataline constants table-driven test** — All 8 dataline key
    constants (selector, mode, namespace, elements, signals, onlyIfMissing,
    useViewTransition, viewTransitionSelector) verified to have trailing
    space and work with DataValue().
21. ✅ **ReadEvents failing reader test** — Custom `failingReader` that
    returns error on Read(). ReadEvents returns error + nil events.
22. ✅ **Static version consistency test** —
    `datastar.DatastarJSVersion == static.Version` assertion.
23. ✅ **ScriptHandler serves static bytes test** — GET ScriptHandler(),
    read full body, byte-compare against `static.Bytes()`. Exact length and
    byte-for-byte match.
24. ✅ **golangci-lint** — 0 issues across all packages.
25. ✅ **go vet** — clean across all packages.
26. ✅ **go test -race -count=1** — all pass (298 test invocations, 166 PASS
    lines including subtests, 4 packages).

---

## (b) PARTIALLY DONE

1. 🟡 **MustReadEvents Fatal path is tested indirectly, not directly.** I
   added `TestReadEvents_FailingReader` which tests that `ReadEvents`
   returns an error from a failing reader. But `MustReadEvents` (which
   calls `t.Fatal` on that error) is NOT tested with a `*testing.T` that
   verifies the Fatal behavior. The error path is proven to exist, but the
   Fatal-conversion behavior is assumed, not asserted. This is because
   testing `t.Fatal` requires either a subprocess test or a mock testing.T
   — more complexity than warranted for a one-line wrapper.

2. 🟡 **`ScriptContent()` strips `<script>` tags but doesn't parse
   attributes.** The method returns everything between `<script ...>` and
   `</script>`. It correctly handles attributes (`type="module"`,
   `data-effect="el.remove()"`) by not including them in the output. But
   there's no way to access those attributes separately. A consumer who
   needs to verify `type="speculationrules"` (Prefetch patches) would need
   to use `Elements()` and parse manually.

3. 🟡 **`CollectN` test for streaming handler uses a busy-loop handler.**
   The streaming test handler sends 10 events as fast as possible, then
   blocks on `<-r.Context().Done()`. This works but isn't representative
   of real streaming (time-spaced events). A more realistic test would use
   a ticker or channel-driven handler. The current test verifies the
   mechanics (read N, close, return) but not the timing semantics.

4. 🟡 **AGENTS.md is modified but uncommitted.** The auto-commit daemon
   committed 5 of my changes as separate commits, but the AGENTS.md API
   surface table expansion (6 new rows) was the last edit and hasn't been
   committed yet. The diff is clean and correct, just pending.

---

## (c) NOT STARTED

1. ❌ **`CollectWithTimeout(t, handler, timeout)`** — Listed as P0 item #7
   in the prior status report. I deprioritized this because `CollectN`
   solves the immediate "hung test" problem for streaming handlers, and
   `Collect` (synchronous handlers) closes the stream when the handler
   returns. A timeout would add safety but the urgency dropped after
   CollectN landed. Still valuable for defensive testing.

2. ❌ **`RedirectURL()` accessor** — The prior report suggested extracting
   the URL from redirect script patches. `ScriptContent()` now returns the
   raw JS (`setTimeout(() => window.location.href = "https://...")`), which
   makes a dedicated `RedirectURL()` less critical — consumers can
   `strings.Contains(scriptContent, url)`. But a purpose-built accessor
   would still be cleaner.

3. ❌ **`CustomEventName()` / `CustomEventDetail()` accessors** — Same
   pattern as RedirectURL: the data is inside the JS blob that
   `ScriptContent()` now returns. Structured extraction would require JS
   parsing or regex matching. Not started.

4. ❌ **`RawSSE()` method on Event** — For debugging test failures by
   seeing the exact wire format. Not started; `String()` partially covers
   this need.

5. ❌ **Godoc examples on individual exported functions** — The prior
   report noted that `example_test.go` has package-level examples but
   individual functions (Collect, CollectPost, CollectN, ScriptContent)
   lack function-level `// Example` functions visible on pkg.go.dev.

6. ❌ **Benchmark for `ReadEvents`** — The SSE parser is untested for
   performance. Low priority for a test helper, but the parent package has
   benchmarks so the bar exists.

7. ❌ **Fuzz test for `ReadEvents`** — The parser handles untrusted input
   (SSE wire format from HTTP responses). Fuzzing would harden it. Not
   started.

8. ❌ **README.md "Testing your handlers" section update** — The existing
   section mentions `Collect` and manual `ReadEvents` for non-GET. It does
   NOT mention the new `CollectPost`, `CollectWithRequest`, `CollectN`, or
   `ScriptContent` helpers. The section is stale relative to the expanded
   API.

9. ❌ **`example_test.go` examples for new functions** — The new
   CollectPost, CollectN, ScriptContent, DataValue, String methods have no
   testable `// Output:` examples.

10. ❌ **CI pipeline verification** — The `.github/workflows/ci.yml` was
    not modified. The new tests should work in CI (runs `go test ./...`),
    but the expanded package was never verified in the CI environment.

---

## (d) TOTALLY FUCKED UP

Nothing is broken or regressively damaged. However:

1. ⚠️ **The auto-commit daemon fragmented this session into 5 commits.**
   The daemon committed at intermediate points, creating commits like
   `docs(changelog): document new datastartest/ and static/ subpackages`
   and `feat(datastartest): add streaming, POST, and event inspection
   helpers` separately. This is expected behavior per the project's git
   workflow, but it means the session's work is spread across 5 commits
   rather than 1-2 logical commits. Not a problem, just a structural
   observation.

2. ⚠️ **`collect_test.go` lint iterations took 3 rounds.** The initial
   write had 12 lint issues (err113, intrange, nlreturn x2, varnamelen x5,
   wsl_v5 x3). I fixed them across 3 iterations (multiedit partial
   failures -> full rewrite -> golangci-lint fmt -> manual wsl_v5 fixes).
   This is normal for this project's extremely strict 120+-linter config,
   but it means I burned extra tool calls on formatting that could have
   been avoided by writing lint-clean code the first time. Specifically:
   - I used `w` instead of `writer` for http.ResponseWriter parameters
     (varnamelen)
   - I used `errors.New()` in test code instead of a package-level var
     (err113)
   - I used `for i := 0; i < n; i++` instead of `for i := range n`
     (intrange)
   - I put `return` immediately after a block without a blank line
     (nlreturn)
   - I assigned a variable and used it in an `if` on the next line without
     a blank line (wsl_v5)

3. ⚠️ **`ScriptContent()` uses `strings.Cut` on `>` which would break if
   the script content itself contains `>` before `</script>`.** In
   practice this can't happen because the `<script ...>` opening tag is
   always followed by `>` before any JS content. But if someone constructs
   a synthetic Elements() value with `>` in the opening tag's attribute
   value (e.g., `<script data-foo="a>b">console.log()</script>`), the cut
   would take everything after the first `>`, producing
   `b">console.log()`. This is a theoretical edge case that doesn't occur
   with real DataStar patches, but it's a correctness gap.

4. ⚠️ **`readNEvents` duplicates buffer/cap constants from `ReadEvents`.**
   Both functions create `bufio.Scanner` with the same
   `initialLineCap`/`maxLineBytes`. If these values need to change,
   they're already shared (same constants), but the scanner setup is
   duplicated. A `newSSEScanner(r io.Reader)` helper would DRY this.

---

## (e) WHAT WE SHOULD IMPROVE

### Design improvements

1. **`ScriptContent()` attribute extraction gap.** The method discards the
   opening tag's attributes. Consumers testing Prefetch patches (which use
   `type="speculationrules"`) or custom script attributes need those. A
   `ScriptAttributes() map[string]string` companion method would be
   valuable.

2. **No `IsScript()` predicate.** `ScriptContent()` returns empty for
   non-script elements, but there's no boolean predicate. A consumer can't
   ask "is this event a script patch?" without checking
   `ScriptContent() != ""`, which is fragile (an empty script is valid).
   `IsScript() bool` would check for the `<script` prefix in Elements().

3. **`CollectN` doesn't cancel the request context on return.** When
   `CollectN` reads N events and returns, it defers `resp.Body.Close()`,
   which causes the server to see EOF. But the handler's
   `<-r.Context().Done()` in the test only fires after the body close
   propagates. If the handler is slow to respond to cancellation, the test
   server's `defer srv.Close()` in CollectN blocks until the handler
   exits. This could cause test slowness with poorly-written handlers.

4. **`readNEvents` scanner setup duplication.** Extract a
   `newSSEScanner(r io.Reader) *bufio.Scanner` helper shared by
   `ReadEvents` and `readNEvents`.

5. **No `CollectWithOptions(t, handler, opts)` extensible pattern.** We now
   have Collect, CollectPost, CollectWithRequest, CollectN — four variants.
   A functional-options pattern (`WithMethod`, `WithBody`, `WithCount`,
   `WithTimeout`) would be more extensible than proliferating functions.
   But the current functions are clearer for the 80% case.

6. **`String()` format is not standardized.** It uses
   `Event{type=... datalines=N}` which is Go-struct-like but not
   structured (no JSON, no parseable format). For debugging this is fine,
   but if consumers want to log events structurally, a `LogJSON()` method
   might be useful.

### Test improvements

7. **No test for `CollectN` with zero count.** `CollectN(t, handler, 0)`
   should return immediately with empty slice. Untested edge case.

8. **No test for `CollectN` when handler sends fewer than N events.** If
   the handler sends 2 events and closes, but `CollectN(t, handler, 5)`
   was called, `readNEvents` falls through to the EOF path and returns
   what it has. This works (proven by `CollectN_AllEvents` test) but
   isn't explicitly tested for the "asked for more than available" case.

9. **No test for `CollectPost` with malformed JSON body.** What happens
   when the handler receives invalid JSON? The test handler returns 400,
   but the test doesn't verify error handling — `CollectPost` would get
   non-SSE response body and `MustReadEvents` would try to parse it. This
   is an error-path gap.

10. **No test for `ScriptContent()` with DispatchCustomEvent.** The
    DispatchCustomEvent patch generates complex JS with `new CustomEvent`,
    `querySelectorAll`, `dispatchEvent`. `ScriptContent()` should extract
    this, but it's untested.

11. **No test for `ScriptContent()` with Prefetch (type="speculationrules").**
    Prefetch patches use `WithScriptAttributes(\`type="speculationrules"\`)`
    and `WithScriptAutoRemove(false)`. The `<script type="speculationrules">`
    tag has attributes that `ScriptContent()` must skip past. Untested.

12. **No concurrent Collect test.** Multiple `Collect` calls in parallel
    (with `t.Parallel()`) work (proven by all tests using t.Parallel()), but
    there's no explicit test asserting concurrent safety.

13. **No test for `DataValue` with multi-line keys.** `DataValue` uses
    `firstValue` which returns only the first match. For multi-line fields
    (elements, signals), this returns the first line only. The typed
    accessors (`Elements()`, `SignalsJSON()`) use `allValues` + join. This
    difference isn't documented or tested for `DataValue`.

### Documentation improvements

14. **README.md "Testing your handlers" section is stale.** Still shows
    only `Collect` and manual `ReadEvents` for non-GET. Needs
    CollectPost, CollectN, ScriptContent examples.

15. **No godoc examples for CollectPost, CollectN, ScriptContent,
    DataValue, String.** The `example_test.go` file predates these
    functions.

16. **CHANGELOG `[Unreleased]` could be split into more granular items.**
    The datastartest entry is a paragraph. Keep-a-Changelog style prefers
    bullet points. Minor formatting preference.

17. **`doc.go` mentions `Broadcaster` without linking.** The streaming
    section says "broadcasting through a [Broadcaster]" but doesn't use
    godoc link syntax. Should be `[sse.Broadcaster]` or similar.

### Polish

18. **`collect.go` imports grew from 4 to 7.** The file now imports
    `bufio`, `fmt`, `io`, `strings` in addition to the original
    `context`, `net/http`, `net/http/httptest`, `testing`. This is fine
    but the file grew from 44 to 171 lines. Consider splitting
    `readNEvents` into a `streaming.go` file.

19. **`event.go` is now 170 lines.** The `ScriptContent`, `DataValue`, and
    `String` methods are general-purpose accessors/debugging. They could
    live in an `accessors.go` file, leaving `event.go` focused on core
    Event type + dataline methods. Minor file organization.

20. **The `failingReader` in `collect_test.go` is exported-capable but
    test-only.** It's a useful mock that could be shared across test
    packages. Consider moving to a `testhelpers.go` file (still internal
    to `_test.go`).

---

## (f) Up to 50 things we should get done next

### High impact (P0)

1. ~~**Update README.md "Testing your handlers" section** — add CollectPost,
   CollectN, ScriptContent examples. The section is stale right now.~~ done in the 04-25 session (README rewritten with subsections)
2. ~~**Add `IsScript() bool` predicate** — check if Elements() starts with
   `<script`. Complements ScriptContent().~~ done at `9f9b7ba`
3. ~~**Add `CollectN(t, handler, 0)` edge-case test** — should return empty
   immediately.~~ done (`TestCollectN_ZeroCount`)
4. ~~**Add `CollectN` "fewer than N" test** — handler sends fewer events
   than requested.~~ done (`TestCollectN_FewerThanRequested`)
5. ~~**Add `ScriptContent()` DispatchCustomEvent test** — complex JS blob
   extraction.~~ done (`TestEvent_ScriptContent_DispatchCustomEvent`)
6. ~~**Add `ScriptContent()` Prefetch test** — `type="speculationrules"`
   attribute handling.~~ done (`TestEvent_ScriptContent_Prefetch`)
7. ~~**Extract `newSSEScanner` helper** — DRY the scanner setup between
   ReadEvents and readNEvents.~~ done in the 04-25 session
8. ~~**Add `CollectWithTimeout(t, handler, timeout)`** — defensive testing
   against hung handlers. Uses context.WithTimeout.~~ done at `22a9589`
9. ~~**Add godoc examples for CollectPost, CollectN, ScriptContent** —
   pkg.go.dev visibility.~~ partially done — examples exist for Collect, ScriptContent, Find*, Filter*, Require*; CollectPost/CollectN still lack dedicated examples
10. **Add `CollectPost` malformed-body error-path test** — handler returns
    400, CollectPost behavior on non-SSE response.

### Medium impact (P1)

11. **Add `RedirectURL() string` accessor** — regex/string extraction from
    ScriptContent() for the URL in `window.location.href = "..."`.
12. **Add `ScriptAttributes() map[string]string`** — parse opening tag
    attributes.
13. **Add `RawSSE() string` on Event** — reconstruct the wire format for
    debugging.
14. ~~**Add fuzz test for `ReadEvents`** — SSE parser is a boundary.~~ done (`FuzzReadEvents`, 9-seed corpus)
15. ~~**Add benchmark for `ReadEvents`** — establish perf baseline.~~ done (`BenchmarkReadEvents`, ~131 MB/s)
16. **Add `CollectPost` error handling test** — non-200 response code.
17. ~~**Test `DataValue` with multi-line keys** — document/test that it
    returns only the first match.~~ done (`TestEvent_DataValue_MultiLineReturnsFirst`)
18. ~~**Add concurrent Collect test** — explicit `t.Parallel()` with multiple
    servers.~~ done (`TestEvent_ConcurrentCollect`)
19. ~~**Fix `ScriptContent()` `>` in attribute value edge case** — use a more
    robust tag-end detection (find `>` only outside quoted attribute values).~~ done at `5cf2f38` (`indexTagEnd`)
20. **Split `collect.go` into `collect.go` + `streaming.go`** — separate
    readNEvents/CollectN from the synchronous Collect variants.
21. **Add `CollectWithOptions(t, handler, opts...)`** — functional options
    pattern for extensibility.
22. **Add `FindElement(events, selector) (Event, bool)`** — search by
    selector across events.
23. **Add `FindSignals(events) (Event, bool)`** — return first signals
    event.
24. **Add `RequireElementsOrdered(t, events, selectors...)`** — assert
    event ordering.
25. **Add `SignalsContain(t, evt, key)` — check signal key exists.
26. **Add `EventsString(events) string`** — human-readable multi-event dump.
27. **Improve `UnmarshalSignals` error message** — include JSON payload
    preview in error.
28. **Add `Diff(expected, actual []Event) string`** — readable diff of event
    sequences.
29. **Add `ServeSSE(handler) (*httptest.Server, func())`** — lower-level
    than Collect for custom request logic.
30. **Document `readNEvents` graceful-close behavior** — the scanner error
    after events collected is silently ignored.

### Lower impact (P2)

31. **Consider `datastartest.NewRecorder()`** — like httptest.NewRecorder
    but for SSE.
32. **Add `Snapshot(t, events)` helper** — golden-file testing for SSE
    output.
33. **Add Ginkgo/Gomega matchers** — if BDD consumers need it.
34. **Verify pkg.go.dev rendering** — check subpackage visibility.
35. **Consider testify-like fluent API** — `Assert(t, events).HasElements(2)`.
36. **Add `CustomEventName()` / `CustomEventDetail()` accessors** —
    structured extraction from DispatchCustomEvent JS.
37. **Add CI check for datastartest coverage** — ensure new tests don't
    regress coverage.
38. ~~**Consider separate Go module for datastartest** — opt-in test dep.~~ done — separate module since v0.1.0
39. ~~**Add versioning note** — how datastartest versions vs core.~~ done — independent versioning (`datastartest/v0.1.0`)
40. **Review all doc comments for godoc rendering** — formatting check.
41. **Add `Event.LogJSON() string`** — structured JSON representation for
    logging.
42. **Add `Event.ID` accessor method** — currently a public field, could be
    method for interface consistency.
43. **Consider `Event.Retry` accessor method** — same as above.
44. ~~**Add `parseSSEField` edge case: line with only a colon** (`:` alone).~~ done (`TestParseSSEField_ColonOnlyLine`)
45. **Add `parseSSEField` edge case: UTF-8 BOM handling** (spec edge case).
46. ~~**Add `ReadEvents` test with CRLF line endings** — Windows
    compatibility.~~ done (`TestReadEvents_CRLFLineEndings`)
47. **Add `ReadEvents` test with very long lines** — near maxLineBytes
    limit.
48. ~~**Add `ReadEvents` test exceeding maxLineBytes** — verify
    scanner.Err() fires.~~ done (`TestReadEvents_ExceedsMaxLineSize`)
49. **Consider configurable maxLineBytes** — via Reader struct or option.
50. **Add `CONTRIBUTING.md` note about datastartest** — how to use it when
    contributing new patch types.

---

## (g) Questions I CANNOT figure out myself

### Q1: ~~Should `datastartest` be a separate Go module (`go.mod`)?~~ Resolved — separate module since v0.1.0.

The parent package `go-datastar` has zero test-only dependencies today.
Adding `datastartest` as a subpackage keeps it in the same module, so
consumers who `go get go-datastar` pull in the test helper code transitively
(though they won't compile it unless they import it). Making it a separate
module would let consumers opt-in to the test dependency explicitly. This
is a packaging/versioning decision with tradeoffs I can't resolve without
knowing your preference. **This question was asked in the prior report and
remains unanswered.**

### Q2: ~~Should we add a `CollectWithOptions(t, handler, opts...)` pattern to replace the growing number of Collect variants?~~ Resolved (2026-08-16) — every `Collect*` helper gained variadic request options (`WithPath`, `WithHeader`, `WithLastEventID`, `WithDatastarSignals`) instead of one consolidating function; see CHANGELOG `[Unreleased]`.

### Q3: ~~Is the `ScriptContent()` approach (return raw JS string) the right abstraction level, or should it parse into typed structures?~~ Resolved — raw-JS approach kept; doc honesty fixed. Typed accessors remain a ROADMAP idea.

---

## Metrics Summary

| Metric | Value |
| --- | --- |
| Files changed | 9 (+1 uncommitted AGENTS.md) |
| Lines added | +655 |
| Lines removed | -19 |
| New exported functions | 6 (`ScriptContent`, `DataValue`, `String`, `CollectWithRequest`, `CollectPost`, `CollectN`) |
| New tests | 19 (including 8 table-driven subtests) |
| Total datastartest tests | 40 test functions (including subtests: 48) |
| Total project test invocations | 298 (all pass with `-race`) |
| `golangci-lint` issues | 0 |
| `go vet` issues | 0 |
| New dependencies | 0 |
| Auto-commits by daemon | 5 |
| Tasks completed | 18 / 18 |
