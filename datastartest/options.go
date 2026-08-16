package datastartest

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/larsartmann/go-datastar"
)

// RequestOption customizes the HTTP request that a Collect* helper sends.
// Options compose: WithPath selects the route, WithHeader and
// WithLastEventID add headers, WithDatastarSignals sets the inbound-signals
// query parameter.
type RequestOption func(*requestConfig)

// requestConfig accumulates the applied [RequestOption]s for one request.
type requestConfig struct {
	path    string
	query   url.Values
	headers http.Header
}

// WithPath targets the request at the given path instead of "/". Use this for
// handlers mounted under a route (e.g., a mux serving "/events") or for
// query-parameter-driven handlers. The path may include its own query string:
//
//	WithPath("/submit?filter=alerts")
func WithPath(path string) RequestOption {
	return func(cfg *requestConfig) {
		cfg.path = path
	}
}

// WithHeader adds a request header. Multiple calls with the same key append
// multiple values.
func WithHeader(key, value string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}

		cfg.headers.Add(key, value)
	}
}

// WithLastEventID sets the Last-Event-ID request header, simulating a browser
// reconnecting after a dropped connection. Handlers that replay missed events
// (e.g., via go-sse's Replay) respond with everything after the given event ID.
// Use this to E2E test reconnection replay without a real browser.
func WithLastEventID(id string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}

		cfg.headers.Set("Last-Event-ID", id)
	}
}

// WithDatastarSignals sends the given JSON as the "datastar" query parameter —
// the way DataStar clients submit signals with GET and DELETE requests (see
// [datastar.ReadSignals]). Use this to test handlers that read inbound signals
// from the query string instead of a request body.
func WithDatastarSignals(signalsJSON string) RequestOption {
	return func(cfg *requestConfig) {
		cfg.setQueryParam(datastar.DatastarKey, signalsJSON)
	}
}

// setQueryParam records a query parameter, replacing any previous value.
func (c *requestConfig) setQueryParam(key, value string) {
	if c.query == nil {
		c.query = make(url.Values)
	}

	c.query.Set(key, value)
}

// targetPath builds the request path from the configured base path and query
// parameters, defaulting to "/" when no path was set.
func (c *requestConfig) targetPath() string {
	path := c.path
	if path == "" {
		path = "/"
	}

	if len(c.query) == 0 {
		return path
	}

	if strings.Contains(path, "?") {
		return path + "&" + c.query.Encode()
	}

	return path + "?" + c.query.Encode()
}

// applyRequestOptions folds opts into a fresh config.
func applyRequestOptions(opts []RequestOption) requestConfig {
	var cfg requestConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
