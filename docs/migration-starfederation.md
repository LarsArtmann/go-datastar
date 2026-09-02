# Migrating from starfederation/datastar-go

This guide maps the upstream SDK (`github.com/starfederation/datastar-go`,
compared against v1.2.2) to go-datastar. The wire format is identical — both
libraries speak the same DataStar protocol — so the migration is about the
programming model, not output.

## The core difference: patches are values

In the upstream SDK, every patch is a method on a `ServerSentEventGenerator`
bound to a live `http.ResponseWriter`. You cannot construct a patch without
an open connection, and the bytes hit the wire immediately:

```go
// upstream: bound to the connection
sse := datastar.NewSSE(w, r)
err := sse.PatchElements("<div>Update</div>")
```

In go-datastar, a patch is a value you can build anywhere and route anywhere:

```go
// go-datastar: build first, send later (or never send — broadcast instead)
patch := datastar.NewElementsPatch("<div>Update</div>",
	datastar.WithSelectorID("feed"),
	datastar.WithModePrepend(),
)

evt := patch.Event() // an sse.Event — store it, filter it, broadcast it
```

Why that matters: broadcast to many connections, replay on reconnect,
per-subscriber filtering, and testable handlers. The
[example/domain-adapter/](../example/domain-adapter/) shows the payoff: a
domain event bridge that produces `[]Patch` values decoupled from any
connection.

## Method mapping

| upstream `ServerSentEventGenerator` | go-datastar equivalent |
| ----------------------------------- | ---------------------- |
| `NewSSE(w, r)` | `sse.NewStream(w, r)` + `datastar.NewResponse(stream)` |
| `PatchElements(html, opts...)` | `resp.PatchElements(html, opts...)` or `NewElementsPatch` + `Send(evt)` |
| `PatchSignals(signalData, opts...)` | `resp.MarshalAndPatchSignals(v, opts...)` or `NewSignalsPatch` |
| `ExecuteScript(script, opts...)` | `resp.ExecuteScript(script, opts...)` or `NewScriptPatch` |
| `Redirect(url, opts...)` | `resp.Redirect(url, opts...)` or `NewRedirectPatch(url)` |
| `ConsoleLog(msg)` / `ConsoleError(err)` | `resp.ConsoleLog(msg)` / `resp.ConsoleError(err)` |
| `DispatchCustomEvent(name, detail)` | `resp.DispatchCustomEvent(name, detail)` |
| `ReplaceURL(url)` / `ReplaceURLQuerystring(r, vals)` | `resp.ReplaceURL(url)` / `resp.ReplaceURLQuerystring(r, vals)` |
| `RemoveElement(sel)` / `RemoveElementByID(id)` | `resp.RemoveElement(sel)` / `resp.RemoveElementByID(id)` |
| `MergeFragments` / fragment options | elements patches with `WithMode*`, `WithSelector*`, `WithViewTransitions*` |

Option mapping is mostly name-preserving upstream→here: `WithSelectorID`,
`WithMode*`, `WithViewTransitions*`, retry-duration options, and custom-event
flags (`WithCustomEventBubbles/Cancelable/Composed/EventID`) all exist here.
Printf-style variants (`WithSelectorf`, `NewRedirectfPatch`,
`NewConsoleLogfPatch`) are go-datastar additions.

## What you gain

| Capability | How |
| ---------- | --- |
| Broadcast one patch to N connections | `sse.NewBroadcaster[sse.Event]()` + `broadcaster.Broadcast(patch.Event())` |
| Reconnection replay | `datastar.NewMemoryStore(capacity)` (any `sse.EventStore`) + `datastar.LastEventID(r)` |
| Per-subscriber filtering | `broadcaster.SubscribeFilter(func(evt sse.Event) bool { ... })` |
| Classified errors | every returned error carries a stable code, family, retryability (see [error-system.md](error-system.md)) |
| E2E test helpers | the `datastartest` module: `Collect`, typed assertions, fuzz corpora (see [testing.md](testing.md)) |
| Serving the JS client | `datastar.ScriptHandler()` (embedded, ETag + cache headers) |

## What you give up

- **Built-in compression** — upstream ships gzip/Brotli/Zstd/Deflate
  negotiation. go-datastar leaves compression to the HTTP layer; a working
  gzip middleware ships in [example/sse_middleware.go](../example/sse_middleware.go),
  and a compressing reverse proxy is the production answer.
- **Lower Go requirement** — upstream works on Go 1.24+; go-datastar requires
  Go 1.26.7+ and `GOEXPERIMENT=jsonv2` (transitive via go-sse).
- **First-party cadence** — upstream tracks DataStar client releases day one;
  here the embedded bundle bumps through Renovate proposals with a review gate.

## Migration checklist

1. Replace the import: `starfederation/datastar-go/datastar` →
   `github.com/larsartmann/go-datastar` (+ `github.com/larsartmann/go-sse`).
2. Replace `NewSSE(w, r)` with `sse.NewStream(w, r)` + `datastar.NewResponse(stream)`
   (or `datastar.NewResponseFromHTTP(w, r)` in one call).
3. Method calls mostly keep their names on `*Response`; constructor-style code
   (`New*Patch`) gains you the value semantics.
4. Swap option constructors where renamed (see table above) — most are identical.
5. Run your handler tests through `datastartest.Collect` and assert on decoded
   events; if you had golden SSE bytes, the wire format is unchanged and they
   should pass as-is.
