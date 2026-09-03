# Performance

Measured on AMD Ryzen AI Max+ 395 (linux/amd64), Go 1.26.7,
`-benchtime=1s` (2026-09-03; multi-million-iteration samples per row).
Treat numbers as order-of-magnitude; run `nix run .#bench` on your own
hardware before making decisions.

## Patch rendering (`Event()`)

Rendering a patch to its `sse.Event` is what your handler pays per patch:

| Benchmark                   | ns/op | B/op | allocs/op |
| --------------------------- | ----- | ---- | --------- |
| ElementsPatch_Event         | ~146  | 352  | 6         |
| SignalsPatch_Event          | ~1113 | 423  | 14        |
| ScriptPatch_Event           | ~329  | 752  | 17        |
| MarshalSignals (map → JSON) | ~853  | 230  | 11        |

Rendering is allocation-light (single digits to mid-teens per patch) and
far below network costs: a patched kilobyte costs microseconds to build and
milliseconds to ship.

## datastartest (test-time, not production)

| Benchmark                            | ns/op    | allocs/op | What it covers                                      |
| ------------------------------------ | -------- | --------- | --------------------------------------------------- |
| Collect (16 events, full round trip) | ~186,000 | 326       | httptest server + GET + SSE parse + dataline decode |
| ReadEvents (parser floor, 16 frames) | ~17,000  | ~112–123  | wire parsing + decoding only, no HTTP               |

The test helper's overhead is dominated by the HTTP round trip, not the
parser — parser cost is ~9% of the Collect round trip.

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
# or, per module (fixed -benchtime for stable stats):
GOEXPERIMENT=jsonv2 go test -run '^$' -bench . -benchmem -benchtime=1s .
GOEXPERIMENT=jsonv2 go test -run '^$' -bench . -benchmem -benchtime=1s ./datastartest
```

Methodology: `-benchtime=1s` (the b.Loop benchmarks self-scale their
iterations), single run per row, desktop-class zen5 APU under normal desktop
load — never benchmark on a battery profile or while a build runs.
