# Security Policy

## Supported Versions

This project is in early development (pre-v1.0). Only the latest release
receives security fixes.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 0.0.x | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not** open a public
GitHub issue or pull request. Instead, use GitHub's private vulnerability
reporting:

1. Go to this repository's **Security** tab
2. Click **Report a vulnerability**
   (direct link: `https://github.com/LarsArtmann/go-datastar/security/advisories/new`)
3. Include a description of the vulnerability and steps to reproduce

Reports are private by default. Expect acknowledgment within 72 hours and
coordination on responsible disclosure once a patch is ready — including a
CVE request through GitHub's advisory system where appropriate.

## Scope

This is a **library** — it does not serve HTTP requests directly. Security
concerns are limited to:

- Incorrect wire-format encoding that could confuse client-side parsing
- Input validation gaps in `ReadSignals` (deserialization of untrusted JSON)
- Embedded `datastar.js` bundle integrity

Out of scope: vulnerabilities in consumer applications built on this
library, the upstream DataStar JavaScript client (report those to
starfederation/datastar), and generic Go runtime issues (report those
upstream).

## Security Practices

- The library performs **no logging** of signal payloads, request bodies, or
  headers — nothing sensitive leaks through go-datastar itself.
- Treat every value read by `ReadSignals` as client-controlled input: it is
  untrusted JSON from the request body, and trust decisions about what those
  values may influence belong to the consumer.
- The embedded `datastar.js` bundle is checksum-pinned
  (`static/checksum_test.go`); a drifted bundle fails CI instead of shipping
  silently.
- go-datastar adds no authentication or authorization layer — keep the SSE
  endpoints that stream its patches protected like any other HTTP route.
