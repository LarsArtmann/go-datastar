# The wire format: what go-datastar puts on the wire

A DataStar "patch" is one SSE event whose `data:` lines carry a small
key-value protocol. This page shows the exact bytes per patch family, as
produced by `Patch.Event()` — the same bytes pinned by `wire_golden_test.go`
and asserted by the testable examples. All keys have a **trailing space**
(`selector `, `elements `, …); that is part of the upstream-compatible
format, not a typo.

Anatomy of one event on the wire:

```
event: datastar-patch-elements      ← the event type
data: selector #feed                ← dataline: key + space + value
data: elements <div>hi</div>
                                    ← blank line ends the event
```

## ElementsPatch

```go
datastar.NewElementsPatch("<div>Hello</div>",
	datastar.WithSelector("#feed"),
	datastar.WithModePrepend(),
).Event().Data
```

```
selector #feed
mode prepend
elements <div>Hello</div>
```

Multi-line element HTML is split on `\n`; **each line becomes its own
`elements` dataline** (parity with upstream). `mode outer` is never emitted —
it is the client default. `namespace html` likewise.

## SignalsPatch

```
signals {"count":1}
```

Multi-line signals JSON emits one `signals` line per source line, all
unconditionally. `onlyIfMissing true` adds `onlyIfMissing true` as its own
dataline.

## ScriptPatch (and every script sugar patch)

Script patches are elements patches appending to `<body>`:

- `selector body` + `mode append` — always (parity item 5)
- `data-effect="el.remove()"` added when `AutoRemove` is nil or true
  (parity item 4)

```go
datastar.NewScriptPatch(`console.log("hi")`).Event().Data
```

```
selector body
mode append
elements <script data-effect="el.remove()">console.log("hi")</script>
```

Sugar variants and their shapes:

| Constructor | Inner JS |
| ----------- | -------- |
| `NewRedirectPatch(u)` | `setTimeout(() => window.location.href = "u")` |
| `NewConsoleLogPatch(msg)` / `NewConsoleErrorPatch(err)` | `console.log(%q)` / `console.error(%q)` — `%q` quoting |
| `NewDispatchCustomEventPatch(name, detail)` | `new CustomEvent("name", {bubbles: true, cancelable: true, composed: true, detail: …})` dispatched on `document` (default) |
| `NewReplaceURLPatch(u)` / `NewReplaceURLQuerystringPatch(r, values)` | `window.history.replaceState({}, "", "…")` |
| `NewPrefetchPatch(urls...)` | speculation-rules JSON, `type="speculationrules"`, no auto-remove |

## Retry and event IDs

`retry: <ms>` is emitted only when the retry duration is `> 0` **and**
differs from the 1000ms default (parity item 3). `id: <id>` is emitted when
an event ID is set. Both lines are written by go-sse before the `data:`
lines.

## Inbound: ReadSignals

- `GET` and `DELETE`: signals come from the `?datastar=` **query parameter**.
- All other methods: signals come from the JSON **request body**.

## Where conformance is enforced

- `wire_golden_test.go` — exact wire bytes for every patch family
- `example_test.go` — testable examples with `// Output:` assertions
- `datastartest/wpt_format_corpus_test.go` — the WHATWG/WPT SSE corpus
- `docs/adr/` parity list in AGENTS.md — the 12 upstream-parity invariants
