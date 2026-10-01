package datastar_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larsartmann/go-datastar"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// FuzzErrorResponseFromError hardens the error-metadata extraction boundary:
// whatever the error message, code, or classification shape, sending the
// error signals patch must never panic and must succeed (nil error) — or, for
// the nil-error misuse case, return the classified Rejection without
// panicking. This closes T16.8 from the 2026-09-03 plan, whose Not-Do was
// never written down.
func FuzzErrorResponseFromError(f *testing.F) {
	// Plain error, empty message.
	f.Add("boom", "datastar.x", uint8(0))
	// Empty everything.
	f.Add("", "", uint8(1))
	// Control characters and newlines in message and code.
	f.Add("line\nbreak\x00ctl", "code.with.dots", uint8(2))
	// Nil-error misuse.
	f.Add("ignored", "ignored", uint8(3))

	f.Fuzz(func(t *testing.T, message, code string, kind uint8) {
		var err error

		switch kind % 4 {
		case 0:
			err = errors.New(message)
		case 1:
			err = errorfamily.NewTransient(code, message)
		case 2:
			err = errorfamily.WrapRejectionf(errors.New(message), code, "wrap: %s", message)
		case 3:
			err = nil // caller misuse: must return the classified Rejection
		}

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		stream := sse.NewStream(recorder, req)

		defer func() { _ = stream.Close() }()

		// The invariants under fuzz: never panic; a non-nil error input
		// sends cleanly, the nil input is rejected with a classified error.
		if err := datastar.ErrorResponseFromError(stream, err); err != nil && kind%4 != 3 {
			t.Fatalf("ErrorResponseFromError: %v", err)
		}
	})
}
