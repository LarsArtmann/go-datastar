package datastartest_test

import (
	"net/http"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

// The Collect* helpers do not gate on the HTTP status code or the response
// content type: whatever body the handler wrote is parsed as SSE, and a body
// without dispatchable data lines simply yields zero events. These tests pin
// that failure mode so a future change to it is deliberate.
//
// If a handler under test misbehaves, the symptom through these helpers is an
// empty (or truncated) event slice — never a panic, never a hang.

func TestCollectPost_ErrorStatusWithNonSSEBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
	}{
		{
			name:        "400 with plain text",
			status:      http.StatusBadRequest,
			contentType: "text/plain; charset=utf-8",
			body:        "Bad Request\n",
		},
		{
			name:        "422 with JSON error",
			status:      http.StatusUnprocessableEntity,
			contentType: "application/json",
			body:        `{"error":"invalid signals"}`,
		},
		{
			name:        "500 with plain text",
			status:      http.StatusInternalServerError,
			contentType: "text/plain; charset=utf-8",
			body:        "Internal Server Error\n",
		},
		{
			name:        "500 with empty body",
			status:      http.StatusInternalServerError,
			contentType: "text/plain; charset=utf-8",
			body:        "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", test.contentType)
				writer.WriteHeader(test.status)
				_, _ = writer.Write([]byte(test.body))
			})

			events := datastartest.CollectPost(t, handler, `{"name":"x"}`)

			if len(events) != 0 {
				t.Fatalf("expected zero events from a %d response, got %d: %s",
					test.status, len(events), datastartest.EventsString(events))
			}
		})
	}
}

func TestCollectPost_SuccessStatusWithNonSSEBody(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte("<html><body>not an SSE stream</body></html>\n"))
	})

	events := datastartest.CollectPost(t, handler, `{}`)

	if len(events) != 0 {
		t.Fatalf("expected zero events from a non-SSE 200 body, got %d: %s",
			len(events), datastartest.EventsString(events))
	}
}

func TestCollectPost_GarbageFramesDoNotDispatch(t *testing.T) {
	t.Parallel()

	// Lines without a "field: value" shape are ignored by the SSE parser;
	// a "data:" line without a blank line terminator still dispatches at EOF.
	// Both behaviors below are pinned so changes surface here first.
	body := "this is not sse\n:no colon no dispatch\n\nrandom trailing text"

	handler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte(body))
	})

	events := datastartest.CollectPost(t, handler, `{}`)

	if len(events) != 0 {
		t.Fatalf("expected garbage lines to yield zero events, got %d: %s",
			len(events), datastartest.EventsString(events))
	}
}

func TestCollectPost_SSEPayloadInErrorResponseStillDecodes(t *testing.T) {
	t.Parallel()

	// Documented sharp edge: the parser never inspects the status code, so a
	// body that happens to contain a well-formed SSE frame decodes even from
	// a 500 response. Consumers must not treat "some events" as "success".
	handler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write(
			[]byte("event: datastar-patch-signals\ndata: signals {\"ok\":false}\n\n"),
		)
	})

	events := datastartest.CollectPost(t, handler, `{}`)

	datastartest.RequireEventCount(t, events, 1)
	datastartest.RequireSignals(t, events[0], `{"ok":false}`)
}
