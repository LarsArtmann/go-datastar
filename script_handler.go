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
func ScriptHandlerWith(js []byte, _ string) http.Handler {
	etag := computeETag(js)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=86400") // 24h

		// Support conditional requests
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(js)))
		_, _ = w.Write(js)
	})
}

// ScriptTag returns an HTML <script> tag that loads the DataStar client from
// the given path.
func ScriptTag(path string) string {
	return `<script type="module" src="` + path + `"></script>`
}

// Version returns the version of the embedded DataStar JavaScript client.
func Version() string { return DatastarJSVersion }
