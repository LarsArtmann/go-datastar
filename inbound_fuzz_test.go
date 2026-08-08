package datastar_test

import (
	"bytes"
	"context"
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
// The useQuery bool selects the query-param path (GET) vs the body path (POST),
// exercising both code paths in ReadSignals. The seed corpus runs as ordinary
// regression cases under `go test`; use `go test -fuzz=FuzzReadSignals` to
// explore the input space.
func FuzzReadSignals(f *testing.F) {
	// Valid payloads — body path.
	f.Add(false, []byte(`{"count":1}`))
	f.Add(false, []byte(`{"deep":{"a":{"b":{"c":1}}}}`))
	f.Add(false, []byte(`null`))
	f.Add(false, []byte(`[]`))

	// Valid payload — query path.
	f.Add(true, []byte(`{"x":"y"}`))

	// Malformed / edge-case payloads.
	f.Add(false, []byte(`{`))                              // truncated object
	f.Add(false, []byte(`{"k":`))                          // truncated value
	f.Add(false, []byte(``))                               // empty body → nil
	f.Add(false, []byte(`"\u0000"`))                       // control character
	f.Add(false, []byte(`{"k":"`+string(rune(0x80))+`"}`)) // invalid UTF-8

	f.Fuzz(func(t *testing.T, useQuery bool, body []byte) {
		var req *http.Request

		if useQuery {
			rawURL := "/api?datastar=" + url.QueryEscape(string(body))
			req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
		} else {
			req = httptest.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/api",
				bytes.NewReader(body),
			)
		}

		var target map[string]any
		// The only invariant under fuzz: never panic.
		_ = datastar.ReadSignals(req, &target)
	})
}
