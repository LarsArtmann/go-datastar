# datastartest

Deep-testing helpers for [go-datastar](https://github.com/LarsArtmann/go-datastar) consumers.

Testing a DataStar handler should not require hand-rolling an SSE parser or
decoding DataStar datalines by hand. `datastartest` gives you one-liners that
spin up a real HTTP server, drive your handler, and hand back typed,
assertable events.

```bash
go get github.com/larsartmann/go-datastar/datastartest
```

## Quick start

```go
import (
    "net/http"
    "testing"

    "github.com/larsartmann/go-datastar"
    "github.com/larsartmann/go-datastar/datastartest"
    "github.com/larsartmann/go-sse"
)

func TestFeedHandler(t *testing.T) {
    t.Parallel()

    events := datastartest.Collect(t, myHandler)
    datastartest.RequireEventCount(t, events, 2)

    datastartest.RequireElements(t, events[0], "#feed", "append", "<div>hello</div>")

    var signals struct {
        Count int `json:"count"`
    }
    if err := events[1].UnmarshalSignals(&signals); err != nil {
        t.Fatalf("unmarshal signals: %v", err)
    }
}
```

## Collecting events

| Helper                                       | Use when                                                        |
| -------------------------------------------- | --------------------------------------------------------------- |
| `Collect(t, handler, opts...)`                | Handler sends patches and returns (GET)                        |
| `CollectPost(t, handler, jsonBody, opts...)`  | POST with a JSON body (form submissions)                       |
| `CollectWithRequest(t, h, method, body, ct, opts...)` | Any method/body/content-type                            |
| `CollectN(t, handler, count, opts...)`        | Streaming handler; reads exactly N events, then closes         |
| `CollectWithTimeout(t, handler, timeout, opts...)` | Defensive read with a deadline; returns partial events    |
| `ReadEvents(r)` / `ReadNEvents(r, n)`         | Parse SSE from any `io.Reader` yourself                        |

## Request options

Every `Collect*` helper accepts options:

```go
// Target a route on a mux (query strings allowed):
datastartest.Collect(t, mux, datastartest.WithPath("/events?filter=alerts"))

// Submit inbound signals the way DataStar clients do with GET/DELETE
// (read them in the handler with datastar.ReadSignals):
datastartest.Collect(t, handler, datastartest.WithDatastarSignals(`{"filter":"alerts"}`))

// Simulate a reconnecting browser for replay testing:
events := datastartest.Collect(t, handler, datastartest.WithLastEventID("42"))

// Any custom header:
datastartest.Collect(t, handler, datastartest.WithHeader("X-Trace", "abc"))
```

## Assertions

```go
datastartest.RequireEventCount(t, events, 2)
datastartest.RequireEventType(t, events[0], "datastar-patch-signals")
datastartest.RequireElements(t, events[0], "#feed", "append", "<div>hello</div>")
datastartest.RequireElementsContains(t, events[0], "body", "append", "console.log('hi')")
datastartest.RequireSignals(t, events[1], `{"count":1}`)
datastartest.RequireSignalsContain(t, events[1], "count")
datastartest.RequireScript(t, events[2], "console.log('hi')")
datastartest.RequireEventID(t, events[0], "42")
```

All helpers accept [`testing.TB`](https://pkg.go.dev/testing#TB), so they work
with `*testing.T`, `*testing.B`, and Ginkgo's `GinkgoT()`.

## Finding events without index math

```go
evt, ok := datastartest.FindElement(events, "#header")
sigEvt, ok := datastartest.FindSignals(events)
elements := datastartest.FilterElements(events)
```

## Debugging failures

```go
t.Fatalf("unexpected events:\n%s", datastartest.EventsString(events))
```

The package is a separate Go module with a stable, tagged API
(`datastartest/v0.x.y`). See [pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/go-datastar/datastartest)
for the complete API.

## Conformance

The SSE parser underneath `datastartest` is pinned to the official browser
suites:

- The WPT `eventsource/format-*` corpus, the WHATWG HTML standard § 9.2.6
  examples, and Chromium's `event_source_parser_test.cc` cases are transcribed
  into executable tests (`wpt_format_corpus_test.go`), each with its upstream
  citation.
- `chunk_boundary_test.go` re-runs the entire corpus through readers that
  deliver the stream in 1–4096 byte chunks, proving parse results are
  independent of TCP chunking.
- `testdata/fuzz/FuzzReadEvents/` carries committed regression seeds,
  including the `"0data: hello\n\n"` crasher (a substring match that is a
  different field name and must dispatch nothing) and the trailing-LF
  terminator regression, ported from
  [go-sse/ssetest](https://github.com/LarsArtmann/go-sse/tree/master/ssetest)
  — the two modules deliberately share one parser implementation, so their
  fuzz corpora stay in lockstep.
