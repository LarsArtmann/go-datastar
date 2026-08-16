package datastartest

import (
	"net/url"
	"testing"
)

func TestRequestConfigTargetPath(t *testing.T) {
	t.Parallel()

	encodedSignals := url.QueryEscape(`{"a":1}`)

	tests := []struct {
		name string
		cfg  requestConfig
		want string
	}{
		{
			name: "empty defaults to root",
			cfg:  requestConfig{},
			want: "/",
		},
		{
			name: "plain path",
			cfg:  requestConfig{path: "/events"},
			want: "/events",
		},
		{
			name: "path with own query",
			cfg:  requestConfig{path: "/events?filter=alerts"},
			want: "/events?filter=alerts",
		},
		{
			name: "query params only",
			cfg:  requestConfig{query: map[string][]string{"datastar": {`{"a":1}`}}},
			want: "/?datastar=" + encodedSignals,
		},
		{
			name: "path with own query plus query params merges with ampersand",
			cfg: requestConfig{
				path:  "/events?filter=alerts",
				query: map[string][]string{"datastar": {`{"a":1}`}},
			},
			want: "/events?filter=alerts&datastar=" + encodedSignals,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.cfg.targetPath(); got != tt.want {
				t.Errorf("targetPath(): got %q, want %q", got, tt.want)
			}
		})
	}
}
