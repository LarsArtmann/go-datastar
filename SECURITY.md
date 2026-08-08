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
GitHub issue. Instead:

1. Email the maintainer directly
2. Include a description of the vulnerability and steps to reproduce
3. Wait for acknowledgment within 72 hours

You will receive updates on the fix progress and coordination for responsible
disclosure once a patch is ready.

## Scope

This is a **library** — it does not serve HTTP requests directly. Security
concerns are limited to:

- Incorrect wire-format encoding that could confuse client-side parsing
- Input validation gaps in `ReadSignals` (deserialization of untrusted JSON)
- Embedded `datastar.js` bundle integrity
