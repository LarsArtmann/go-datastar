package datastar

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/larsartmann/go-datastar/static"
)

// DatastarJSVersion is the version of the embedded DataStar JavaScript client.
// It re-exports [static.Version] for backward compatibility with the root API.
//
// Deprecated: use [Version] (which returns the same string) or
// github.com/larsartmann/go-datastar/static.Version directly. This constant
// exists only so existing consumers keep compiling; it will be removed in the
// next minor release after the migration window.
const DatastarJSVersion = static.Version

// ETagFor returns the ETag header value the script handlers serve for the
// given bundle: a quoted, truncated SHA-256 digest, computed once at handler
// construction. It is exported so consumer tests (via
// github.com/larsartmann/go-datastar/datastartest) can assert the caching
// setup around their own [ScriptHandlerWith] bundles without duplicating
// the digest algorithm.
func ETagFor(bundle []byte) string {
	h := sha256.Sum256(bundle)

	return `"` + hex.EncodeToString(h[:16]) + `"`
}

// ScriptHandler returns an [http.Handler] that serves the embedded DataStar
// JavaScript client bundle with correct Content-Type, ETag, and Cache-Control
// headers. Only GET and HEAD requests are allowed; all others return 405.
func ScriptHandler() http.Handler {
	return ScriptHandlerWith(static.Bytes())
}

// ScriptHandlerWith returns an [http.Handler] that serves a custom JavaScript
// bundle. Use this to serve a different version of the DataStar client. The
// bundle's content identity is the ETag [ETagFor] computes, so serve it under
// a URL you version yourself — [ScriptTagVersioned] pins the embedded
// bundle's version, not a custom bundle's.
func ScriptHandlerWith(scriptBytes []byte) http.Handler {
	etag := ETagFor(scriptBytes)

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		responseWriter.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
		responseWriter.Header().Set("ETag", etag)
		responseWriter.Header().Set("Cache-Control", "public, max-age=86400") // 24h

		// Support conditional requests
		if match := request.Header.Get("If-None-Match"); match == etag {
			responseWriter.WriteHeader(http.StatusNotModified)

			return
		}

		responseWriter.Header().Set("Content-Length", strconv.Itoa(len(scriptBytes)))

		// HEAD requests receive the same headers (including Content-Length) but no
		// message body, per RFC 7231 §4.3.2.
		if request.Method == http.MethodHead {
			responseWriter.WriteHeader(http.StatusOK)

			return
		}

		if _, err := responseWriter.Write(scriptBytes); err != nil {
			return // client disconnected or write failed; nothing more to send
		}
	})
}

// ScriptTag returns an HTML <script> tag that loads the DataStar client from
// the given path. The path is used verbatim: no normalization, no escaping;
// callers own URL correctness. Prefer [ScriptTagVersioned] so a browser cache
// cannot serve a stale client after an upgrade.
func ScriptTag(path string) string {
	return `<script type="module" src="` + path + `"></script>`
}

// ScriptTagVersioned returns an HTML <script> tag like [ScriptTag], with the
// pinned DataStar client version as a cache-busting query parameter in the
// src URL. [ScriptHandler] serves the bundle with Cache-Control
// max-age=86400, so a browser that fetched the bundle skips ETag revalidation
// for up to 24h; a versioned URL turns each release into a distinct cache key
// and collapses that stale window to zero for pages rendered per request. If
// the path already carries a v parameter, its value is replaced with the
// pinned version — the result always carries exactly the pinned version, so
// the versioned URL is idempotent; any other existing query joins with &, and
// a fragment is preserved by inserting the parameter before it. The path is
// otherwise used verbatim — no normalization, no escaping — so a malformed
// path yields a malformed src, and callers own URL correctness. The version
// is always the embedded [static.Version]: pair this tag with [ScriptHandler]
// or a serving setup for `static.Bytes`; a custom bundle served via
// [ScriptHandlerWith] carries a different content identity than the URL
// advertises, so version the URL yourself in that setup.
func ScriptTagVersioned(path string) string {
	return ScriptTag(appendVersionQuery(path, static.Version))
}

func appendVersionQuery(path string, version string) string {
	head := path
	fragment := ""

	if index := strings.Index(path, "#"); index >= 0 {
		head = path[:index]
		fragment = path[index:]
	}

	queryStart := strings.Index(head, "?")
	if queryStart < 0 {
		return head + "?v=" + version + fragment
	}

	query := head[queryStart+1:]
	for start := 0; ; {
		rest := query[start:]
		end := strings.IndexByte(rest, '&')

		param := rest
		if end >= 0 {
			param = rest[:end]
		}

		if key, _, hasValue := strings.Cut(param, "="); hasValue && key == "v" {
			valueStart := queryStart + 1 + start + len("v=")

			return head[:valueStart] + version + head[valueStart+len(param)-len("v="):] + fragment
		}

		if end < 0 {
			break
		}

		start += end + 1
	}

	return head + "&v=" + version + fragment
}

// Version returns the version of the embedded DataStar JavaScript client.
// It reads [static.Version] directly; the deprecated [DatastarJSVersion]
// constant is an alias of the same value.
func Version() string { return static.Version }
