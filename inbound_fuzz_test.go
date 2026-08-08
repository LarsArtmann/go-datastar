package datastar_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/larsartmann/go-datastar"
)

// FuzzReadSignals hardens the inbound parsing boundary against malformed input.
// ReadSignals accepts untrusted JSON from a request body (POST/PUT/…) or the
// datastar query parameter (GET/DELETE). The invariant: it must never panic for
// any input — it returns nil (no signals) or a classified error.
//
// The seed corpus runs as ordinary regression cases under `go test`; use
// `go test -fuzz=FuzzReadSignals` to explore the input space.
func FuzzReadSignals(f *testing.F) {
	// Valid payloads.
	f.Add(http.MethodPost, []byte(`{"count":1}`))
	f.Add(http.MethodPost, []byte(`{"deep":{"a":{"b":{"c":1}}}}`))
	f.Add(http.MethodGet, []byte(`{"x":"y"}`))

	// Malformed / edge-case payloads.
	f.Add(http.MethodPost, []byte(`{`))            // truncated object
	f.Add(http.MethodPost, []byte(`{"k":`))        // truncated value
	f.Add(http.MethodPost, []byte(``))             // empty body → nil
	f.Add(http.MethodPost, []byte(`null`))         // JSON null
	f.Add(http.MethodPost, []byte(`[]`))           // not an object
	f.Add(http.MethodPost, []byte(`"\u0000"`))     // control character
	f.Add(http.MethodPost, []byte(`{"k":"`+string(rune(0x80))+`"}`)) // invalid UTF-8

	f.Fuzz(func(t *testing.T, method string, body []byte) {
		var req *http.Request

		if method == http.MethodGet || method == http.MethodDelete {
			// Query-param path: signals travel as ?datastar=<json>.
			rawURL := "/api?datastar=" + url.QueryEscape(string(body))
			req = httptest.NewRequest(method, rawURL, nil)
		} else {
			req = httptest.NewRequest(method, "/api", bytes.NewReader(body))
		}

		var target map[string]any
		// The only invariant under fuzz: never panic.
		_ = datastar.ReadSignals(req, &target)
	})
}
