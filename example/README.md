# example/ — the live-feed demo

`go run .` in this directory starts a live-feed dashboard on `:8765` built
entirely from go-datastar patches: broadcaster fan-out, MemoryStore replay
on reconnect, and the pinned DataStar JS served by `ScriptHandler`. Zero
custom client-side JavaScript.

`domain-adapter/` is a second, smaller demo: a domain layer speaking its own
events, bridged into patches — see the README section "Patches as values in
a domain architecture" in the repository root.

## Heartbeat keep-alive

The event handler starts a heartbeat goroutine per connection:

```go
go stream.Heartbeat(request.Context(), heartbeatInterval) // 15s
```

`Stream.Heartbeat` (go-sse) writes SSE comment frames on an interval so
intermediate proxies and load balancers do not reap idle connections. The
context ends the heartbeat when the client disconnects or the handler
returns. Typical intervals: 15–30s — shorter wastes bytes, longer risks
proxy idle timeouts (nginx's default proxy_read_timeout is 60s).

## Drop observability

The broadcaster registers go-sse's `WithOnDrop` callback, logging events that
are dropped when a subscriber's buffer is full (`ondrop_test.go` exercises
it). This is deliberately a **broadcaster-level** concern: go-sse's `OnDrop`
lives on `Broadcaster`, not `Stream`, so `Response` (which wraps a stream)
cannot expose it. Route lossy-feed logging through your broadcaster, not
through per-connection responses.

## docker-compose

`docker-compose.yml` builds and runs the demo; the Dockerfile pattern is
a plain multi-stage Go build (no Node toolchain — the JS bundle is embedded).

## Compression

`gzipSSEMiddleware` (see `sse_middleware.go`) shows the recommended way to
compress SSE through the HTTP layer — per-event flushing included.
