# Performance

Measured on AMD Ryzen AI Max+ 395 (linux/amd64), Go 1.26.7, `-benchtime 100x`
(2026-08-29). Treat numbers as order-of-magnitude; run `nix run .#bench` on
your own hardware before making decisions.

## Patch rendering (`Event()`)

Rendering a patch to its `sse.Event` is what your handler pays per patch:

| Benchmark                   | ns/op | B/op | allocs/op |
| --------------------------- | ----- | ---- | --------- |
| ElementsPatch_Event         | ~476  | 352  | 6         |
| SignalsPatch_Event          | ~1318 | 478  | 14        |
| ScriptPatch_Event           | ~633  | 752  | 17        |
| MarshalSignals (map → JSON) | ~1045 | 287  | 11        |

Rendering is allocation-light (single digits to mid-teens per patch) and
far below network costs: a patched kilobyte costs microseconds to build and
milliseconds to ship.

## datastartest (test-time, not production)

| Benchmark                            | ns/op      | allocs/op | What it covers                                      |
| ------------------------------------ | ---------- | --------- | --------------------------------------------------- |
| Collect (16 events, full round trip) | ~906,000   | 339       | httptest server + GET + SSE parse + dataline decode |
| ReadEvents (parser floor, 16 frames) | ~10–12,000 | ~112      | wire parsing + decoding only, no HTTP               |

The test helper's overhead is dominated by the HTTP round trip, not the
parser — parser cost is ~1% of the Collect round trip.

## Where the real costs live

- **Broadcast fan-out** is go-sse's `Broadcaster` (per-subscriber channels);
  see the go-sse docs and the `example/` app's OnDrop logging.
- **SSE through proxies**: without compression, SSE streams are text; put a
  compressing reverse proxy in front (or go-sse's compression options) for
  high-volume text feeds — see the honest-compression note on the README.
- **Heartbeats** (`Stream.Heartbeat(ctx, interval)`) keep idle connections
  from being reaped; see `example/main.go` for the wiring.

## Reproducing

```bash
nix run .#bench
# or, per module:
GOEXPERIMENT=jsonv2 go test -run '^$' -bench . -benchmem ./...
```
