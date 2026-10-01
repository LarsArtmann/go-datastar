package datastar_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// invalidUTF8 is invalid by construction: \xff and \xfe never start a valid
// UTF-8 sequence.
const invalidUTF8 = "bad\xff\xfe bytes"

// The best-effort reporting senders exist so an error or notification UI
// reaches the client even when the surrounding handler is already failing —
// they must never fail on their own payload. json/v2 (unlike v1) rejects
// invalid UTF-8, so every handler-supplied diagnostic string is sanitized to
// U+FFFD instead. Found by FuzzErrorResponseFromError; pinned here for the
// whole class.
func TestBestEffortSendersSurviveInvalidUTF8(t *testing.T) {
	t.Parallel()

	if utf8.ValidString(invalidUTF8) {
		t.Fatal("test bug: input must be invalid UTF-8")
	}

	senders := []struct {
		name string
		send func(stream *sse.Stream) error
	}{
		{
			name: "ErrorResponse",
			send: func(stream *sse.Stream) error {
				return datastar.ErrorResponse(stream, invalidUTF8, invalidUTF8)
			},
		},
		{
			name: "ErrorResponseFromError",
			send: func(stream *sse.Stream) error {
				//nolint:err113 // the invalid-UTF-8 message bytes are the input under test
				return datastar.ErrorResponseFromError(stream, errors.New(invalidUTF8))
			},
		},
		{
			name: "NotificationResponse",
			send: func(stream *sse.Stream) error {
				return datastar.NotificationResponse(stream, invalidUTF8, invalidUTF8)
			},
		},
	}

	for _, sender := range senders {
		t.Run(sender.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"/events",
				nil,
			)
			stream := sse.NewStream(recorder, req)

			defer func() { _ = stream.Close() }()

			if err := sender.send(stream); err != nil {
				t.Fatalf("%s must not fail on invalid UTF-8: %v", sender.name, err)
			}

			body := recorder.Body.String()
			if !utf8.ValidString(body) {
				t.Fatalf("%s sent invalid UTF-8: %q", sender.name, body)
			}

			if !strings.Contains(body, "\ufffd") {
				t.Fatalf("%s must replace invalid bytes with U+FFFD: %q", sender.name, body)
			}
		})
	}
}
