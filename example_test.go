package datastar_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/larsartmann/go-datastar"
)

// ExampleElementsPatch demonstrates the library's keystone: a patch is a value
// you construct without a connection. Its wire format is fully determined by
// the struct fields and options.
func ExampleElementsPatch() {
	patch := datastar.NewElementsPatch("<div>Hello</div>",
		datastar.WithSelector("#feed"),
		datastar.WithModePrepend(),
	)

	fmt.Println(patch.Event().Data)
	// Output:
	// selector #feed
	// mode prepend
	// elements <div>Hello</div>
}

// ExampleSignalsPatch shows pre-encoded JSON signals emitted as a wire event.
func ExampleSignalsPatch() {
	patch := datastar.SignalsPatch{Signals: []byte(`{"count":1}`)}

	fmt.Println(patch.Event().Data)
	// Output:
	// signals {"count":1}
}

// ExampleNewReplaceURLQuerystringPatch replaces the browser URL's query
// string while preserving the request path (upstream-parity convenience).
func ExampleNewReplaceURLQuerystringPatch() {
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/items?filter=old",
		nil,
	)

	patch := datastar.NewReplaceURLQuerystringPatch(req, url.Values{"filter": {"new"}})

	fmt.Println(patch.Event().Data)
	// Output:
	// selector body
	// mode append
	// elements <script data-effect="el.remove()">window.history.replaceState({}, "", "/items?filter=new")</script>
}

// ExampleReadSignals demonstrates extracting signals from an inbound request body.
func ExampleReadSignals() {
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api",
		strings.NewReader(`{"count":2}`),
	)

	var signals struct {
		Count int `json:"count"`
	}
	if err := datastar.ReadSignals(req, &signals); err != nil {
		return
	}

	fmt.Println(signals.Count)
	// Output: 2
}
