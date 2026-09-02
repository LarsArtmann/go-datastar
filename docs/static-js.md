# The pinned DataStar JS bundle (`static/`)

The `static` module embeds the official DataStar client JavaScript so your
binary serves a versioned, tested bundle with **zero client-side build
tooling**.

```go
import "github.com/larsartmann/go-datastar/static"

static.Version // e.g. "1.0.2"
static.Bytes() // the raw datastar.js contents (//go:embed)
```

Serve it via the root module's `ScriptHandler` (sets content-type and
caching headers) or `ScriptTag` (renders the `<script src>` tag), or stream
it through your own asset pipeline.

## Why pinning matters

The wire format this library speaks is defined by the DataStar client JS.
The bundle and the Go protocol layer must move together: an embedded bundle
pins the client behavior your handlers were tested against. Serving whatever
CDN a browser page points at reintroduces drift.

## Upgrade process

1. Check the [DataStar releases](https://github.com/starfederation/datastar)
   and the JS client changelog.
2. Replace `static/datastar.js` with the new release and update
   `static.Version` to match the client version.
3. Run the full gate (the wire-format golden tests and the WPT corpus
   attribute changes loudly):
   ```bash
   GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1
   nix flake check
   ```
4. If goldens change, treat that as a protocol change: compare against the
   upstream SDK behavior and record the delta in the CHANGELOG — a golden
   change is deliberate, never incidental.

## Renovate

Automated bump proposals for the bundle are on the roadmap (a Renovate rule
watching the upstream client); until then upgrades are manual per the steps
above.

## Hardening decisions

- **`X-Content-Type-Options: nosniff`** is set by `ScriptHandler`/`ScriptHandlerWith`
  (correct-by-default for a fixed Content-Type asset).
- **`Last-Modified` is deliberately NOT set.** Freshness is fully owned by the
  ETag + `Cache-Control: public, max-age=86400`, and the bundle is immutable
  per release; a `Last-Modified` header would add a second freshness axis
  (a fixed epoch per bundle) that must be maintained on every bump for no
  revalidation benefit over If-None-Match.
- **The bundle checksum is pinned by `static/checksum_test.go`** — a bundle
  replacement must update the constant in the same commit, so asset drift can
  never land silently.
- **No fuzz target for the bundle:** `Bytes()` is an embedded, trusted asset,
  not parsed input; fuzzing belongs at the protocol boundaries
  (`FuzzReadSignals`, `FuzzReadEvents`).

## Constraint check

The pinned client version is enforced twice by tests: the bundle header must
contain `static.Version` (`static/static_test.go`), and the CHANGELOG must
mention the pinned version so a bundle bump can never land without release
history (`version_constraint_test.go`, runs in CI's test job).
