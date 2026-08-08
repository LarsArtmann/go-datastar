package datastar

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strconv"
)

//go:embed static/datastar.js
var embeddedDatastarJS []byte

// DatastarJSVersion is the version of the embedded DataStar JavaScript client.
const DatastarJSVersion = "1.0.2"

func computeETag(data []byte) string {
	h := sha256.Sum256(data)

	return `"` + hex.EncodeToString(h[:16]) + `"`
}

// ScriptHandler returns an [http.Handler] that serves the embedded DataStar
// JavaScript client bundle with correct Content-Type, ETag, and Cache-Control
// headers. Only GET and HEAD requests are allowed; all others return 405.
func ScriptHandler() http.Handler {
	return ScriptHandlerWith(embeddedDatastarJS, DatastarJSVersion)
}

// ScriptHandlerWith returns an [http.Handler] that serves a custom JavaScript
// bundle. Use this to serve a different version of the DataStar client.
func ScriptHandlerWith(scriptBytes []byte, _ string) http.Handler {
	etag := computeETag(scriptBytes)

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		responseWriter.Header().Set("Content-Type", "text/javascript; charset=utf-8")
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
// the given path.
func ScriptTag(path string) string {
	return `<script type="module" src="` + path + `"></script>`
}

// Version returns the version of the embedded DataStar JavaScript client.
func Version() string { return DatastarJSVersion }
