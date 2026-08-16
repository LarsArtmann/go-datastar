package datastartest_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
	errorfamily "github.com/larsartmann/go-error-family"
)

// FuzzUnmarshalSignals hardens the classified-error path of
// [datastartest.Event.UnmarshalSignals]. The invariant: for any payload the
// method either decodes successfully or returns an error carrying the stable
// datastartest.signals_unmarshal_failed code — and it never panics.
//
// Seeds cover valid JSON (objects, arrays, scalars), truncated payloads,
// wrong types, deep nesting, BOMs, and NUL bytes. Run with
// `go test -fuzz=FuzzUnmarshalSignals` to explore the input space.
func FuzzUnmarshalSignals(f *testing.F) {
	// Valid JSON objects — the happy path for signals patches.
	f.Add(`{"count":1}`)
	f.Add(`{"user":{"name":"lars","tags":["a","b"]}}`)
	f.Add(`{}`)

	// Valid JSON that is not an object — decodes into any, still success.
	f.Add(`[1,2,3]`)
	f.Add(`42`)
	f.Add(`"text"`)
	f.Add(`true`)
	f.Add(`null`)

	// Malformed payloads — must fail with the classified code.
	f.Add(``)
	f.Add(`{`)
	f.Add(`{"a":`)
	f.Add(`{"a":1,}`)
	f.Add(`nul`)
	f.Add(`{"a":"unterminated`)

	// Structural edge cases.
	f.Add(`{"duplicate":1,"duplicate":2}`)
	f.Add(`{"deep":` + strings.Repeat(`[`, 32) + strings.Repeat(`]`, 32) + `}`)

	// Byte-level garbage.
	f.Add("\x00\x01\x02")
	f.Add("\xef\xbb\xbf" + `{"bom":"prefix"}`)
	f.Add(`{"nul":"` + "\x00" + `"}`)

	f.Fuzz(func(t *testing.T, payload string) {
		event := datastartest.Event{
			Type: "datastar-patch-signals",
			DataLines: []string{
				"signals " + payload,
			},
		}

		var target any

		err := event.UnmarshalSignals(&target)
		if err == nil {
			return
		}

		if code := errorfamily.Code(err); code != datastartest.CodeSignalsUnmarshalFailed {
			t.Errorf("unmarshal(%q) error must carry %s, got code %q: %v",
				payload, datastartest.CodeSignalsUnmarshalFailed, code, err)
		}
	})
}
