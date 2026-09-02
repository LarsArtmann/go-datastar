# The error system: classified errors, three ways to match

Every error returned by go-datastar (and go-sse beneath it, and datastartest
above it) is a classified `*errorfamily.Error`: a **stable machine-readable
code**, a behavioral **family**, and **structured context**. This page is the
consumer's contract; the design rationale lives in
[ADR 003](adr/003-error-classification.md), the catalog in `errors.go`.

## Why not string-matching

```go
// Fragile: breaks when the wording changes, locale-blind, untestable.
if strings.Contains(err.Error(), "marshal") { ... }
```

Stable codes and typed families survive refactors, render in logs, and make
recovery logic explicit.

## The three matching dimensions

### 1. By code — "what exactly failed"

```go
if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed {
	// stable string, no message parsing
}
```

Root codes: `datastar.templ_render_failed`, `datastar.gostar_render_failed`,
`datastar.body_read_after_close`, `datastar.body_read_failed`,
`datastar.signals_unmarshal_failed`, `datastar.signals_marshal_failed`,
`datastar.custom_event_detail_marshal_failed`, `datastar.event_name_required`,
`datastar.element_patch_mode_invalid`, `datastar.namespace_invalid`,
`datastar.stream_send_failed`, `datastar.error_response_nil_error`.

datastartest codes: `datastartest.sse_scan_failed`,
`datastartest.signals_unmarshal_failed`,
`datastartest.custom_event_detail_unmarshal_failed`.

go-sse codes (match through the same mechanism): e.g. `sse.send_failed`.

### 2. By sentinel — "is it this known failure"

```go
if errors.Is(err, datastar.ErrEventNameRequired) { ... }
if errors.Is(err, datastar.ErrBodyReadAfterClose) { ... }
```

Sentinel matching traverses wrapping chains. `ErrBodyReadAfterClose` wraps
`http.ErrBodyReadAfterClose`, so the stdlib cause stays matchable too.
`errors.Is` against sentinels matches by **value** (code+family) — context
clones of a sentinel still match.

### 3. By family — "whose fault, and should I retry"

| Family        | When                                                                                                                | Retryable |
| ------------- | ------------------------------------------------------------------------------------------------------------------- | --------- |
| Rejection     | bad or missing caller input (malformed JSON, empty name, invalid mode/namespace, closed body, unmarshallable value) | no        |
| Transient     | temporary I/O failure reading a request body; stream Send failure                                                   | yes       |
| Orchestration | internal render failure producing HTML (templ, gostar)                                                              | no        |

```go
if errorfamily.IsRetryable(err) {
	// back off and retry: Transient family
}
```

## Full contract rules

1. **Every error is classified.** No bare `fmt.Errorf` escapes the library.
2. **`error` is the return type, not `*errorfamily.Error`.** Typed access
   goes through `errorfamily.Code` / `errors.Is` / `errorfamily.IsRetryable`.
3. **Context loss is a bug.** Wrapping errors carry the in-scope values that
   diagnosed the failure (HTTP method, input byte length, value type).
4. **Layered classification.** go-sse classifies transport errors
   (`sse.send_failed`); go-datastar re-tags stream sends as
   `datastar.stream_send_failed` (Transient). `errors.Is` traverses the
   chain, so matching either code works.
5. **No samber/oops in the library.** Libraries classify; applications
   enrich. Wrap these errors with your observability stack at your boundary.

## A complete handler

```go
func feedHandler(w http.ResponseWriter, r *http.Request) {
	stream := sse.NewStream(w, r)
	defer stream.Close()
	resp := datastar.NewResponse(stream)

	var signals FeedSignals
	if err := datastar.ReadSignals(r, &signals); err != nil {
		switch {
		case errors.Is(err, datastar.ErrBodyReadAfterClose):
			// caller misuse; nothing to retry
		case errorfamily.IsRetryable(err):
			// transient body read failure; client may retry
		}
		slog.Error("read signals", "code", errorfamily.Code(err), "err", err)
		return
	}

	if err := resp.PatchElements(renderFeed(signals)); err != nil {
		slog.Error("send patch", "code", errorfamily.Code(err), "err", err)
	}
}
```

Runnable versions of all three patterns live in
`errors_example_test.go` (root) — `go test` executes them, so the examples
cannot rot.
