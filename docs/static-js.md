# The pinned DataStar JS bundle (`static/`)

The `static` module embeds the official DataStar client JavaScript so your
binary serves a versioned, tested bundle with **zero client-side build
tooling**.

```go
import "github.com/larsartmann/go-datastar/static"

static.Version // e.g. "1.0.3"
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
2. Fetch the bundle and its checksum with `static/fetch-bundle.sh <version>`
   (downloads `bundles/datastar.js` at the tag, prints the sha256 and the
   provenance line), then replace `static/datastar.js` and update
   `static.Version` to match.
3. Run the full gate (the wire-format golden tests and the WPT corpus
   attribute changes loudly):
   ```bash
   GOEXPERIMENT=jsonv2 go test ./... ./datastartest/... ./static/... -race -count=1
   nix flake check
   ```
4. If goldens change, treat that as a protocol change: compare against the
   upstream SDK behavior and record the delta in the CHANGELOG — a golden
   change is deliberate, never incidental.

## CSP mode (`data-nonce`, client v1.0.3+)

The pinned client supports Content-Security-Policy deployments without
`unsafe-eval`. It is **consumer opt-in** — nothing on the Go side changes:

1. Your server renders a per-response nonce on the root element:
   `<html data-nonce="YOUR-NONCE">`. The nonce must be nonempty (an empty
   attribute makes the client throw) and should be a fresh random value per
   response, same value your CSP `script-src 'nonce-…'` uses.
2. On load the client reads and removes the attribute, creates a
   `trustedTypes` policy named `datastar` (passthrough for the HTML/JS it
   generates), stamps its injected `<script>` elements with the nonce, and
   compiles expressions via nonce-carrying script elements instead of
   `new Function(...)`.

Go-served bundles need no special handling: `ScriptHandler` serves the file
like any static asset; the CSP dance happens in your page template.

## Canonical minified bundle only

`static/datastar.js` is the **canonical upstream minified bundle**, kept
byte-identical to the tag (the provenance comment lives in `static.go`, not
in the bundle, so the file never gets a local header). The repo deliberately
keeps no beautified variant: a second copy would drift, defeat the checksum
pin (`static/checksum_test.go` fails on ANY byte change), and the 2026-08-29
incident showed reformatting sweeps find exactly such files. Formatter
shields: `.prettierignore`, `.codespellrc`, dprint excludes. If you need
readable client source, read it at the upstream tag, never beautify in-repo.

## v1.0.3 scope note

The v1.0.2 → v1.0.3 bump (shipped in v0.5.0) changed **client runtime
behavior only** — CSP mode (above), signals resent when a backend request
retries after a network error, view-transition support checked on the
document as well, and a dynamic `multiple` on `<select>` updating the bound
signal — plus the artifact switch from a beautified ~56 KB file to the
canonical minified 33.5 KB bundle. The wire format was untouched: all
wire-format goldens stayed green, so no Go protocol changes accompanied it.

## Renovate

`renovate.json`'s custom manager (scoped via `enabledManagers: ["regex"]` —
Dependabot owns every other ecosystem, owner decision 2026-10-01) proposes
`Version` bumps when `starfederation/datastar` cuts a release. The proposal
only touches the version string: the bundle download, checksum update, and
golden re-run above still land manually. Delivery is unverified until the
first upstream release after the 2026-08-29 onboarding (see
`docs/ci-watch.md`).

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
