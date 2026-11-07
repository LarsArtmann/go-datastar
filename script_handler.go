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

func computeETag(data []byte) string {
	h := sha256.Sum256(data)

	return `"` + hex.EncodeToString(h[:16]) + `"`
}

// ScriptHandler returns an [http.Handler] that serves the embedded DataStar
// JavaScript client bundle with correct Content-Type, ETag, and Cache-Control
// headers. Only GET and HEAD requests are allowed; all others return 405.
func ScriptHandler() http.Handler {
	return ScriptHandlerWith(static.Bytes(), static.Version)
}

// ScriptHandlerWith returns an [http.Handler] that serves a custom JavaScript
// bundle. Use this to serve a different version of the DataStar client.
//
// The version parameter is unused: it is kept only for signature
// compatibility with the v0.5.x API. Removing it is a breaking change
// pencilled in for v0.7.0 (owner-gated); pass any value — static.Version is
// the conventional choice.
func ScriptHandlerWith(scriptBytes []byte, _ string) http.Handler {
	etag := computeETag(scriptBytes)

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
// pinned DataStar client version appended as a cache-busting query parameter.
// [ScriptHandler] serves the bundle with Cache-Control max-age=86400, so a
// browser that fetched the bundle skips ETag revalidation for up to 24h; a
// versioned URL turns each release into a distinct cache key and collapses
// that stale window to zero for pages rendered per request. If path already
// contains a query, the version joins with &; a fragment is preserved by
// inserting the parameter before it. The path is otherwise used verbatim;
// callers own URL correctness.
func ScriptTagVersioned(path string) string {
	return appendVersionQuery(path, static.Version)
}

func appendVersionQuery(path string, version string) string {
	head := path
	fragment := ""
	if index := strings.Index(path, "#"); index >= 0 {
		head = path[:index]
		fragment = path[index:]
	}

	separator := "?"
	if strings.Contains(head, "?") {
		separator = "&"
	}

	return head + separator + "v=" + version + fragment
}

// Version returns the version of the embedded DataStar JavaScript client.
// It reads [static.Version] directly; the deprecated [DatastarJSVersion]
// constant is an alias of the same value.
func Version() string { return static.Version }
