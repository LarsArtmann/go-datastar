package datastartest

import (
	"io"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse/ssetest"
)

// ReadEvents parses the SSE wire format from r and returns all decoded events.
// It reads until EOF, so the source must close or end the stream (e.g., an HTTP
// response body from a handler that sends all patches and returns).
//
// Parsing is delegated to [github.com/larsartmann/go-sse/ssetest], which
// implements the WHATWG HTML Living Standard § 9.2.6 event-stream
// interpretation (conformance pinned by the Web Platform Tests vectors):
//
//   - Lines end with CR, LF, or CRLF (§ 9.2.5 end-of-line).
//   - Exactly one leading UTF-8 BOM is stripped; a mid-stream BOM is data.
//   - Field names are case-sensitive; unknown fields and ":" comments are ignored.
//   - A single leading space after the ":" separator is stripped (never a tab).
//   - Only frames with at least one data: line dispatch; a "data:" line with an
//     empty value still dispatches an event with an empty payload.
//   - The last event ID is sticky: an id: field updates the buffer, the buffer
//     persists across frames, and each dispatched event reports the buffer's
//     value at dispatch time. An id: value containing U+0000 NULL is ignored.
//   - The retry: value must be all ASCII digits (leading zeros allowed); an
//     invalid value is ignored without resetting a previous one. Like the last
//     event ID, the reconnection time is connection-level state: it persists
//     across frames, and each dispatched event reports the value in effect.
//   - An incomplete final frame (no blank line before EOF) is discarded, per
//     "Once the end of the file is reached, any pending data must be discarded".
//
// DataStar datalines are preserved individually in [Event.DataLines] with their
// key prefixes intact (e.g., "selector #feed"), so typed accessors like
// [Event.Selector] and [Event.Elements] can decode them.
func ReadEvents(r io.Reader) ([]Event, error) {
	events, err := ssetest.ReadEvents(r)
	if err != nil {
		return nil, rewrapScanError(err)
	}

	return toEvents(events), nil
}

// ReadNEvents reads up to count events from r. Returns as soon as count events
// have been dispatched, without waiting for EOF. This is the streaming-reader
// counterpart to [ReadEvents]: use it with a live SSE connection body that does
// not close on its own (e.g., a handler broadcasting through a Broadcaster).
//
// Wire-format semantics are identical to [ReadEvents] (spec § 9.2.6), except
// that reading stops early: a frame still pending when count is reached is
// naturally discarded, as is a frame pending at EOF.
//
// A scanner error after events have been collected is treated as a clean
// connection close, not a failure.
func ReadNEvents(r io.Reader, count int) ([]Event, error) {
	events, err := ssetest.ReadNEvents(r, count)
	if err != nil {
		return nil, rewrapScanError(err)
	}

	return toEvents(events), nil
}

// MustReadEvents is like [ReadEvents] but calls t.Fatal on error. Accepts
// [testing.TB], so it works with *testing.T, *testing.B, and GinkgoT().
func MustReadEvents(tb testing.TB, r io.Reader) []Event {
	tb.Helper()

	events, err := ReadEvents(r)
	if err != nil {
		tb.Fatalf("read SSE events: %v", err)
	}

	return events
}

// MustReadNEvents is like [ReadNEvents] but calls t.Fatal on error.
// Use this with streaming SSE connections that do not close on their own.
// Accepts [testing.TB], so it works with *testing.T, *testing.B, and GinkgoT().
func MustReadNEvents(tb testing.TB, r io.Reader, count int) []Event {
	tb.Helper()

	events, err := ReadNEvents(r, count)
	if err != nil {
		tb.Fatalf("read %d SSE events: %v", count, err)
	}

	return events
}

// toEvents converts the ssetest event slice into the DataStar event slice.
// The two types are field-identical (Type, DataLines, ID, Retry), so the
// conversion is a straight copy.
func toEvents(events []ssetest.Event) []Event {
	out := make([]Event, 0, len(events))

	for _, evt := range events {
		out = append(out, Event{
			Type:      evt.Type,
			DataLines: evt.DataLines,
			ID:        evt.ID,
			Retry:     evt.Retry,
		})
	}

	return out
}

// rewrapScanError re-tags a scan error with the DataStar-specific code so
// consumers can classify errors via [errorfamily.Code] without depending on
// ssetest's code namespace. The underlying cause is preserved in the chain.
func rewrapScanError(err error) error {
	return errorfamily.WrapTransient(err, CodeSSEScanFailed, "scan SSE stream")
}
