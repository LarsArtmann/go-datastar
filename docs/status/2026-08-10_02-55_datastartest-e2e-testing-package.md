# Status Report: 2026-08-10 02:55 — datastartest E2E Testing Package

> Session focused on building a consumer-facing E2E testing helper package for
> go-datastar. Triggered by the question: "My consumers LOVE to have very high
> test coverage. How could we provide something that would make it easier?"

---

## What Was Done

### Goal

Give consumers a reusable library for E2E testing DataStar handlers without
hand-rolling SSE parsing and DataStar dataline decoding.

### Approach taken

Created a `datastartest/` subpackage with:

- `Event` type with typed accessors (`Selector()`, `Mode()`, `Elements()`, etc.)
- `ReadEvents(io.Reader)` SSE wire-format parser
- `Collect(t, handler)` one-liner test server + request + decode
- Filter helpers (`FilterElements`, `FilterSignals`)
- Assertion helpers (`RequireElements`, `RequireElementsContains`, `RequireSignals`)
- 24 tests + 4 testable examples
- Refactored existing `e2e_test.go` from 261 → 109 lines (dogfooding)

### Files created/modified

| File                           | Lines | Status                                                     |
| ------------------------------ | ----- | ---------------------------------------------------------- |
| `datastartest/doc.go`          | 65    | NEW — package documentation with quick-start example       |
| `datastartest/event.go`        | 122   | NEW — Event type + 12 typed accessors                      |
| `datastartest/reader.go`       | 112   | NEW — SSE wire-format parser                               |
| `datastartest/collect.go`      | 44    | NEW — Collect helper                                       |
| `datastartest/filter.go`       | 21    | NEW — FilterElements/FilterSignals                         |
| `datastartest/assert.go`       | 93    | NEW — Require* assertion helpers                           |
| `datastartest/event_test.go`   | 268   | NEW — 16 tests covering all accessors                      |
| `datastartest/reader_test.go`  | 187   | NEW — 8 tests covering SSE parsing edge cases              |
| `datastartest/example_test.go` | 90    | NEW — 4 testable examples                                  |
| `e2e_test.go`                  | 109   | MODIFIED — refactored from 261 to 109 lines                |
| `AGENTS.md`                    | —     | MODIFIED — added datastartest to file layout + new section |
| `README.md`                    | —     | MODIFIED — added "Testing your handlers" section           |

---

## (a) FULLY DONE

1. ✅ `datastartest/` package with 6 source files (485 lines of library code)
2. ✅ `Event` type with typed accessors for every dataline key:
   - `Selector()`, `Mode()` (with default-to-outer), `Namespace()` (with default-to-html)
   - `Elements()` (multi-line rejoined), `SignalsJSON()` (multi-line rejoined)
   - `UnmarshalSignals(&v)` with proper error wrapping
   - `UseViewTransitions()`, `ViewTransitionSelector()`, `OnlyIfMissing()`
   - `IsElements()`, `IsSignals()` type predicates
3. ✅ `ReadEvents(io.Reader)` — spec-compliant SSE parser handling event/data/id/retry/comment lines, EOF without trailing blank line, no-space-after-colon edge case
4. ✅ `MustReadEvents(t, r)` — Fatal-on-error variant
5. ✅ `Collect(t, handler)` — complete test-server-to-decoded-events pipeline
6. ✅ `FilterElements` / `FilterSignals` — non-mutating slice filters
7. ✅ 5 assertion helpers: `RequireEventCount`, `RequireEventType`, `RequireElements`, `RequireElementsContains`, `RequireSignals`
8. ✅ 24 tests (16 event accessor tests + 8 reader tests) — all pass with `-race`
9. ✅ 4 testable examples with `// Output:` assertions — all pass
10. ✅ Refactored `e2e_test.go` to use `datastartest` (261 → 109 lines, -58%)
11. ✅ `golangci-lint run ./...` — **0 issues**
12. ✅ `go vet ./...` — clean
13. ✅ `go test ./... -race -count=1` — all pass (272 tests total across project)
14. ✅ AGENTS.md updated with package in file layout + new "E2E Testing" section
15. ✅ README.md updated with "Testing your handlers" section

---

## (b) PARTIALLY DONE

1. 🟡 **CHANGELOG.md** — NOT updated. The `[Unreleased]` section is empty. This
   new package is a user-facing feature addition and should be logged. ~~→ done at `7c18089`~~
2. 🟡 **FEATURES.md** — NOT updated. Should have a new "Testing" section listing
   the `datastartest` package as `FULLY_FUNCTIONAL`. ~~→ done at `222353e`~~
3. 🟡 **Test coverage of `datastartest` itself** — Good but not exhaustive:
   - `Collect` is tested transitively via every event test, but has no dedicated
     error-path test (e.g., handler that panics, handler that returns 500).
   - `MustReadEvents` (the Fatal variant) has no test — only `ReadEvents` is tested.
   - No test for a response with zero events (empty SSE stream from a handler
     that sends nothing).
   - No test for the `Retry` field being correctly parsed from a real handler
     (only tested via string input in reader_test.go).
4. 🟡 **Assertion helpers coverage** — `RequireSignals`, `RequireEventType` are
   implemented but only exercised incidentally, not in dedicated focused tests.

---

## (c) NOT STARTED

1. ❌ **`go.mod` — no new module needed**, but the subpackage creates a new
   import path (`github.com/larsartmann/go-datastar/datastartest`). This works
   fine with Go modules (subpackage in same module), but it's worth confirming
   pkg.go.dev renders it correctly.
2. ❌ **CI pipeline verification** — `.github/workflows/ci.yml` was not modified.
   It should already work (runs `go test ./...`), but the new subpackage was
   never verified in the CI environment. The `GOEXPERIMENT=jsonv2` env var is
   already set globally in CI, so this should be fine. ~~→ done — CI has run the package since v0.1.0 (all-module test job)~~
3. ❌ **Integration test with cqrs-htmx/datastar** — the domain-layer consumer
   was not tested to confirm the package API works for their EventBridge pattern.
4. ❌ **Benchmark** — no `benchmark_test.go` in `datastartest/`. The SSE parser
   is untested for performance. For a test helper this is low priority, but the
   parent package has benchmarks so the bar is set. ~~→ done — `BenchmarkReadEvents` (~131 MB/s)~~

---

## (d) TOTALLY FUCKED UP

Nothing is broken or regressively damaged. However:

1. ⚠️ **erraudit failure is pre-existing, not caused by this session.** The
   `erraudit ./...` command fails because erraudit cannot parse
   `encoding/json/v2` (needs `GOEXPERIMENT=jsonv2` which erraudit doesn't
   support). This affects ALL files using jsonv2, including pre-existing ones
   (`inbound.go`, `signals.go`, `script_convenience.go`). My new
   `datastartest/event.go` adds one more file to this list. The error is a
   tooling limitation, not a code bug.
2. ⚠️ **`example_test.go` examples are partially illustrative.** `ExampleCollect`
   and `ExampleRequireElements` don't actually call `Collect` or `RequireElements`
   (because they need `*testing.T`); they demonstrate the decoded output shape
   via `ReadEvents` on string input instead. This is idiomatic for Go examples
   but worth noting — a newcomer might be confused why the example doesn't match
   the function name exactly.
3. ⚠️ **`e2e_test.go` no longer tests `emitAllPatches` as a named function.**
   The refactor inlined the patches directly into the handler closure. This is
   cleaner but lost the named helper. Not a problem, just a structural change.

---

## (e) WHAT WE SHOULD IMPROVE

### Design improvements

1. **No `ScriptContent()` accessor.** Script patches produce `<script>` HTML
   that gets decoded via `Elements()`, but consumers will frequently want just
   the JS source extracted. Currently they'd need to parse the `<script>` tags
   themselves. A `ScriptContent()` method that strips the wrapper would be
   valuable.
2. **No `RedirectURL()` accessor.** Same pattern — redirect patches embed the
   URL inside JS inside `<script>` inside Elements. Extracting the URL requires
   regex or string parsing.
3. **No `CustomEventName()` / `CustomEventDetail()` accessors.** DispatchCustomEvent
   patches bury the event name and detail inside a JS blob. These are impossible
   to decode structurally from datalines alone (they're embedded in script).
4. **`Collect` only does GET.** For a "make E2E testing easy" package, POST/PUT
   with request bodies is a very common need. A `CollectWithRequest(t, handler, req)`
   or `CollectPost(t, handler, body, contentType)` variant would reduce boilerplate
   significantly. The doc comments point consumers to `httptest.NewServer` +
   `ReadEvents`, but that's exactly the boilerplate we're trying to eliminate.
5. **No `CollectN(t, handler, n)` for streaming handlers.** If a handler keeps
   the connection open (like the example's `/events` endpoint), `Collect` hangs
   forever. A variant that reads exactly N events then closes would handle this.
6. **`Event.DataLines` is exported but raw.** Consumers might reach for it when
   typed accessors don't cover a field. A `DataValue(key string) string` generic
   accessor would be safer and more discoverable.
7. **No `RawSSE()` method on Event.** For debugging test failures, seeing the
   exact wire format that produced an event would be valuable.
8. **Error message on `UnmarshalSignals` could include the JSON.** Currently
   returns `"unmarshal signals JSON: <json error>"` but doesn't show which JSON
   failed. Adding the payload (or a preview) would speed up debugging.

### Test improvements

9. **No table-driven test for all dataline keys.** A single test iterating
   every `*DatalineKey` constant against its accessor would catch regressions
   if constants change.
10. **No test for `Retry` field round-trip.** The reader parses it from string
    input, but no end-to-end test sends a patch with `WithElementsRetryDuration`
    and verifies `Event.Retry` on the decoded side.
11. **No test for `EventID` field round-trip.** Same gap — parsed in reader
    tests, but not tested through the full Collect pipeline.
12. **`MustReadEvents` has zero test coverage.** Only `ReadEvents` is tested.
13. **No test for concurrent Collect calls.** With `t.Parallel()`, multiple
    test servers run simultaneously. This works (proven by all tests passing
    with `-race`), but there's no explicit test for it.

### Documentation improvements

14. **No godoc examples on individual functions.** The `example_test.go` file
    has package-level examples, but individual functions like `Collect`,
    `ReadEvents`, `RequireElements` would benefit from function-level godoc
    examples visible on pkg.go.dev.
15. **README testing section could be more prominent.** It's placed after
    EventStore and before Error Handling — could get lost. A callout in the
    Quick Start section would increase discoverability.

### Polish

16. **`datastartest/reader.go` has two unexported constants** (`maxLineBytes`,
    `initialLineCap`) with magic numbers. These are fine but could be config
    options on a `Reader` struct if we want configurability.
17. **`Collect` uses `http.DefaultClient`** which has no timeout. For a test
    helper this is usually fine (httptest.Server is local), but if a handler
    hangs, the test hangs. An explicit timeout would be safer.

---

## (f) Up to 50 things we should get done next

### High impact (P0)

1. ~~**Update CHANGELOG.md** `[Unreleased]` section with `datastartest` package~~ done at `7c18089`
2. ~~**Update FEATURES.md** with new "Testing" section~~ done at `222353e`
3. ~~**Add `ScriptContent()` accessor** — strip `<script>` wrapper, return JS source~~ done at `c1ca7ce` (hardened at `5cf2f38`)
4. ~~**Add `CollectPost(t, handler, body)` variant** — eliminate POST boilerplate~~ done at `c1ca7ce`
5. ~~**Add `CollectWithRequest(t, handler, req)` variant** — full custom request control~~ done at `c1ca7ce`
6. ~~**Add `CollectN(t, handler, n)` variant** — for streaming handlers (read N events, close)~~ done at `9f9b7ba`
7. ~~**Add `CollectWithTimeout(t, handler, timeout)` variant** — prevent hung-test hangs~~ done at `22a9589`
8. ~~**Test `MustReadEvents`** — at least one test exercising the Fatal path~~ done — `TestMustReadEvents_FailingReader` + `TestMustReadNEvents` (`datastartest/assert_test.go`, 2026-08-16)
9. ~~**Test `Retry` field round-trip** through full Collect pipeline~~ done (`TestEvent_RetryEventIDRoundTrip`)
10. ~~**Test `EventID` field round-trip** through full Collect pipeline~~ done (same test)
11. ~~**Test `Collect` with empty handler** (zero events returned)~~ done (`TestEvent_EmptyHandler`)

### Medium impact (P1)

12. **Add `RedirectURL()` accessor** — extract URL from redirect script patch
13. **Add `CustomEventName()` accessor** — extract event name from dispatch patch
14. **Add `CustomEventDetail()` accessor** — extract JSON detail from dispatch patch
15. ~~**Add `DataValue(key string) string`** — generic dataline lookup fallback~~ done at `c1ca7ce`
16. **Add `RawSSE()` method on Event** — for debugging test failures
17. ~~**Add `Event.String()` method** — readable debug representation~~ done at `c1ca7ce`
18. ~~**Table-driven test for all dataline constants** — guard against key changes~~ done (`TestDatalineConstants_TableDriven`)
19. ~~**Dedicated test for each `Require*` helper** — including failure message quality~~ mostly done — failure paths covered in `assert_test.go` (2026-08-16)
20. ~~**Test concurrent Collect calls explicitly**~~ done (`TestEvent_ConcurrentCollect`)
21. ~~**Add godoc examples on individual exported functions**~~ done (`ExampleCollect`, `ExampleFindElement`, etc.)
22. **Promote testing section in README** — link from Quick Start
23. **Add `Reader` struct with configurable max line size** — for edge cases
24. ~~**Improve `UnmarshalSignals` error** — include JSON payload in error message~~ done at `9f9b7ba` era (v0.1.0)
25. ~~**Test the `parseSSEField` edge case: line with multiple colons** (e.g., `data: {"a":"b"}`)~~ done (`TestParseSSEField_MultiColon`)
26. ~~**Test the `parseSSEField` edge case: empty data field** (`data:` with nothing after)~~ done (`TestParseSSEField_EmptyData`)

### Lower impact (P2)

27. **Add `FindElement(t, events, selector)` helper** — search by selector
28. **Add `FindSignals(t, events)` helper** — return first signals event
29. **Add `RequireElementsOrdered(t, events, selectors...)` — assert ordering
30. **Add `SignalsContain(t, evt, key)` helper** — check signal key exists
31. **Add `ElementsMatch(t, evt, selector, mode, html)` — alias for RequireElements
32. **Add benchmark for ReadEvents** — establish perf baseline
33. **Add fuzz test for ReadEvents** — SSE parser is a boundary, should be fuzzed
34. **Add `ServeSSE(handler)` returning `(*httptest.Server, func())`** — lower-level than Collect
35. **Consider `datastartest.NewRecorder()`** — like httptest.NewRecorder but for SSE
36. **Verify pkg.go.dev rendering** — check the subpackage appears correctly
37. **Consider a testify-like fluent API** — `datastartest.Assert(t, events).HasElements(2).First().SelectorIs("#feed")`
38. **Add Ginkgo/Gomega matchers** — if the project uses BDD (see bdd-testing skill)
39. **Test with real browser (Playwright)** — true E2E beyond wire format
40. **Document the relationship to go-sse's EventStore** — datastartest reads, EventStore stores
41. **Add `CollectWithOptions(t, handler, opts)`** — extensible config pattern
42. **Consider `Snapshot(t, events)` helper** — golden-file testing for SSE output
43. **Add `Diff(expected, actual []Event) string`** — readable diff of event sequences
44. **Add `EventsString(events) string`** — human-readable multi-event debug dump
45. **Test handler that sends patches from a goroutine** — concurrency within handler
46. **Test handler that calls `stream.Close()` with an error**
47. **Add CI check that `datastartest/` is included in coverage reports**
48. **Consider whether `datastartest` should be a separate Go module** — for consumers who don't want test deps in their go.sum
49. **Add versioning note** — if datastartest evolves, how does it version vs the core package?
50. **Review all doc comments for godoc rendering** — ensure formatting is correct on pkg.go.dev

---

## (g) Questions I CANNOT figure out myself

### 1. ~~Should `datastartest` be a separate Go module (`go.mod`)?~~ Resolved — separate module since v0.1.0 (`github.com/larsartmann/go-datastartest/datastartest`, tagged `datastartest/v0.1.0`).

The parent package `go-datastar` has zero test-only dependencies today — it's
purely stdlib + go-sse + go-error-family. Adding `datastartest` as a subpackage
keeps it in the same module, so consumers who `go get go-datastar` pull in the
test helper code transitively (though they won't compile it unless they import
it). Making it a separate module (`go-datastar/datastartest go.mod`) would let
consumers opt-in to the test dependency explicitly. This is a packaging/
versioning decision with tradeoffs I can't resolve without knowing your
preference.

### 2. ~~Should we support the streaming/broadcaster use case?~~ Resolved — `CollectN` and `CollectWithTimeout` shipped (`9f9b7ba`, `22a9589`).

The current `Collect` assumes a **synchronous handler** — send patches, return,
close stream. But the example app (`example/main.go`) and the Broadcaster
pattern keep connections open indefinitely. Should `datastartest` provide
helpers for testing streaming handlers (e.g., `CollectN` that reads N events
then closes, or a `Subscription` type that reads events in real-time)? This
significantly expands the API surface and I don't know if your consumers need it.

### 3. Is the `example_test.go` pattern (illustrative, not executable against real API) acceptable?

Go examples must have `// Output:` assertions to be run by `go test`, but
functions like `Collect` require a `*testing.T` which examples can't provide.
I worked around this by demonstrating the _decoded output shape_ via
`ReadEvents` on string input, rather than calling `Collect` directly. The
example is titled `ExampleCollect` but doesn't call `Collect`. Is this
acceptable, or should I restructure to have only genuinely-executable examples
that call the real API?

---

## Metrics Summary

| Metric                  | Value                      |
| ----------------------- | -------------------------- |
| New source files        | 6 (485 lines)              |
| New test files          | 3 (545 lines)              |
| New tests               | 24 (all pass with `-race`) |
| New testable examples   | 4 (all pass)               |
| `e2e_test.go` reduction | 261 → 109 lines (-58%)     |
| Total project tests     | 272 (all pass)             |
| `golangci-lint` issues  | 0                          |
| `go vet` issues         | 0                          |
| New dependencies        | 0                          |
