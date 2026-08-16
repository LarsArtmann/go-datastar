package datastartest_test

// This file transcribes the official Web Platform Tests (WPT) SSE wire-format
// corpus into executable Go tests — the same conformance suite that pins
// go-sse's ssetest reader. The two parsers are deliberately duplicated
// (independent single-dependency modules), so the corpus must pin both.
//
// WPT tests are browser-oriented: they spin a server (resources/message.py)
// and drive EventSource. Each test encodes a precise (wire bytes → expected
// events) pair, which is what we transcribe here.
//
// message.py semantics used below:
//   - default: body = message + "\n\n" (terminates the final frame)
//   - newline=none: body = message + "\n" (no terminating blank line)
//
// Upstream corpus: https://github.com/web-platform-tests/wpt/tree/master/eventsource
// Spec: https://html.spec.whatwg.org/multipage/server-sent-events.html

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

// wantEvent is the observable surface of one dispatched SSE event: the type,
// the joined data payload, the last event ID in effect, and the reconnection
// time in effect.
type wantEvent struct {
	Type  string
	Data  string
	ID    string
	Retry uint
}

// sseCase is one wire-format conformance vector.
type sseCase struct {
	name string
	url  string // upstream source this vector is transcribed (or derived) from
	wire string // exact bytes on the wire, after message.py terminator expansion
	want []wantEvent
}

// wptCorpus holds the transcribed WPT eventsource/format-*.any.js vectors.
var wptCorpus = []sseCase{
	{
		name: "format-field-data",
		url:  "wpt:eventsource/format-field-data.any.js",
		wire: "data:" + "\n\n" +
			"data" + "\n" +
			"data" + "\n\n" +
			"data:test" + "\n\n",
		// A data line with an empty value still dispatches; two bare "data"
		// lines dispatch an event whose payload is a single newline.
		want: []wantEvent{
			{Data: ""},
			{Data: "\n"},
			{Data: "test"},
		},
	},
	{
		name: "format-field-event",
		url:  "wpt:eventsource/format-field-event.any.js",
		wire: "event: x" + "\n" + "data:x" + "\n\n",
		want: []wantEvent{{Type: "x", Data: "x"}},
	},
	{
		name: "format-field-event-empty",
		url:  "wpt:eventsource/format-field-event-empty.any.js",
		wire: "event: " + "\n" + "data:data" + "\n\n",
		// An empty event type fires onmessage, i.e. type "".
		want: []wantEvent{{Data: "data"}},
	},
	{
		// An id field whose value contains U+0000 NULL anywhere is ignored
		// entirely (spec § 9.2.6); the last event ID stays "".
		name: "format-field-id-null",
		url:  "wpt:eventsource/format-field-id-null.any.js",
		wire: "id:\x00\x00" + "\n" + "data:1" + "\n\n" +
			"id:x\x00" + "\n" + "data:2" + "\n\n" +
			"id:\x00x" + "\n" + "data:3" + "\n\n" +
			"id:x\x00x" + "\n" + "data:4" + "\n\n" +
			"id: \x00" + "\n" + "data:5" + "\n\n",
		want: []wantEvent{
			{Data: "1"},
			{Data: "2"},
			{Data: "3"},
			{Data: "4"},
			{Data: "5"},
		},
	},
	{
		// Field names are case-sensitive; unknown fields and no-colon lines
		// with unknown names are ignored; only the first colon separates; a
		// single leading space after the colon is stripped (never a tab).
		name: "format-field-parsing",
		url:  "wpt:eventsource/format-field-parsing.any.js",
		wire: "data:\x00" + "\n" +
			"data:  2" + "\r" +
			"Data:1" + "\n" +
			"data\x00:2" + "\n" +
			"data:1" + "\r" +
			"\x00data:4" + "\n" +
			"da-ta:3" + "\r" +
			"data_5" + "\n" +
			"data:3" + "\r" +
			"data:" + "\r\n" +
			" data:32" + "\n" +
			"data:4" + "\n" +
			"\n",
		want: []wantEvent{{Data: "\x00\n 2\n1\n3\n\n4"}},
	},
	{
		name: "format-field-retry",
		url:  "wpt:eventsource/format-field-retry.any.js",
		// Leading zeros are allowed: the value is a decimal integer.
		wire: "retry:03000" + "\n" + "data:x" + "\n\n",
		want: []wantEvent{{Data: "x", Retry: 3000}},
	},
	{
		name: "format-field-retry-bogus",
		url:  "wpt:eventsource/format-field-retry-bogus.any.js",
		// A bogus retry does not reset a previously valid one, and a retry
		// line takes effect even in a frame that never dispatches.
		wire: "retry:3000" + "\n" + "retry:1000x" + "\n\n" + "data:x" + "\n\n",
		want: []wantEvent{{Data: "x", Retry: 3000}},
	},
	{
		name: "format-field-retry-empty",
		url:  "wpt:eventsource/format-field-retry-empty.any.js",
		// Derived from the upstream file, which sends a bare "retry" line and
		// asserts the reconnection time is unchanged: an empty value is not
		// all-digits, so it is ignored. Pinned here with a prior valid value.
		wire: "retry:2000" + "\n\n" + "retry" + "\n\n" + "data:x" + "\n\n",
		want: []wantEvent{{Data: "x", Retry: 2000}},
	},
	{
		name: "format-field-unknown",
		url:  "wpt:eventsource/format-field-unknown.any.js",
		// Unknown field names (with or without a colon) are ignored.
		wire: "data:1" + "\n" + "foo" + "\n" + "foo: bar" + "\n" + "data:2" + "\n\n",
		want: []wantEvent{{Data: "1\n2"}},
	},
	{
		name: "format-newlines",
		url:  "wpt:eventsource/format-newlines.any.js",
		// CR, LF, and CRLF are all line terminators; a bare "data" line with
		// no colon is an empty data line; "\r\r" is two terminators, the
		// second of which is the blank line that dispatches the frame.
		wire: "data:alpha" + "\r\n" +
			"data" + "\n" +
			"data:beta" + "\r\r" +
			"\n",
		want: []wantEvent{{Data: "alpha\n\nbeta"}},
	},
	{
		name: "format-leading-space",
		url:  "wpt:eventsource/format-leading-space.any.js",
		// Only one leading space is stripped: a tab survives, "data: " is an
		// empty data line.
		wire: "data:\talpha" + "\r" +
			"data: " + "\n" +
			"data:beta" + "\n\n",
		want: []wantEvent{{Data: "\talpha\n\nbeta"}},
	},
	{
		name: "format-comments",
		url:  "wpt:eventsource/format-comments.any.js",
		// Comment lines are ignored; note the ":<CR><LF>" comment: the CRLF is
		// one terminator, not a comment line followed by a blank dispatch line.
		wire: "data:1" + "\r" +
			":\x00" + "\n" +
			":" + "\r\n" +
			"data:2" + "\n" +
			":" + strings.Repeat("x", 2048) + "\r" +
			"data:3" + "\n" +
			":data:fail" + "\r" +
			":" + strings.Repeat("x", 2048) + "\n" +
			"data:4" + "\n" +
			"\n",
		want: []wantEvent{{Data: "1\n2\n3\n4"}},
	},
	{
		name: "format-null-character",
		url:  "wpt:eventsource/format-null-character.any.js",
		// NUL is valid data; only in id values is it rejected.
		wire: "data:\x00" + "\n\n",
		want: []wantEvent{{Data: "\x00"}},
	},
	{
		name: "format-bom",
		url:  "wpt:eventsource/format-bom.any.js",
		// Exactly one leading UTF-8 BOM is stripped. A mid-stream BOM is NOT:
		// it poisons the field name, so that line is ignored.
		wire: "\xEF\xBB\xBF" + "data:1" + "\n\n" +
			"\xEF\xBB\xBF" + "data:2" + "\n\n" +
			"data:3" + "\n\n",
		want: []wantEvent{
			{Data: "1"},
			{Data: "3"},
		},
	},
	{
		name: "format-bom-2",
		url:  "wpt:eventsource/format-bom-2.any.js",
		// Two leading BOMs: only the first is stripped; the second poisons the
		// first field name, so only the later events surface.
		wire: "\xEF\xBB\xBF\xEF\xBB\xBF" + "data:1" + "\n\n" +
			"data:2" + "\n\n",
		want: []wantEvent{{Data: "2"}},
	},
	{
		name: "format-utf-8",
		url:  "wpt:eventsource/format-utf-8.any.js",
		// The stream is always decoded as UTF-8 regardless of any charset
		// parameter; non-ASCII payloads pass through unchanged.
		wire: "data:döm" + "\n\n",
		want: []wantEvent{{Data: "döm"}},
	},
	{
		name: "format-data-before-final-empty-line",
		url:  "wpt:eventsource/format-data-before-final-empty-line.any.js",
		// newline=none: the final frame never gets its blank line, so it is
		// discarded at EOF — and its id:test must not leak into the last event
		// ID of any dispatched event.
		wire: "retry:1000" + "\n" +
			"data:test1" + "\n\n" +
			"id:test" + "\n" +
			"data:test2" + "\n",
		want: []wantEvent{{Data: "test1", Retry: 1000}},
	},
}

// specExamples holds the example streams printed in the spec itself (§ 9.2.6
// "Examples"), with the trailing blank line each ongoing stream implies.
//
// https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
var specExamples = []sseCase{
	{
		name: "spec-example-test-stream",
		url:  "spec:§9.2.6 example 1",
		wire: ": test stream" + "\n\n" +
			"data: first event" + "\n" +
			"id: 1" + "\n\n" +
			"data:second event" + "\n" +
			"id: 2" + "\n\n" +
			"event: custom event" + "\n" +
			"data: third event" + "\n" +
			"id: 3" + "\n\n" +
			": this is a test stream" + "\n",
		want: []wantEvent{
			{Data: "first event", ID: "1"},
			{Data: "second event", ID: "2"},
			{Type: "custom event", Data: "third event", ID: "3"},
		},
	},
	{
		name: "spec-example-two-identical-events",
		url:  "spec:§9.2.6 examples (data: test ≡ data:test)",
		// The spec pairs "data: test" with "data:test" to show that the single
		// space after the colon is optional: both fire the same event.
		wire: "data:test" + "\n\n" +
			"data: test" + "\n\n",
		want: []wantEvent{
			{Data: "test"},
			{Data: "test"},
		},
	},
	{
		name: "spec-example-stock-ticker",
		url:  "spec:§9.2.6 examples (stock ticker)",
		wire: "data: YHOO" + "\n" +
			"data: +2" + "\n" +
			"data: 10" + "\n\n",
		want: []wantEvent{{Data: "YHOO\n+2\n10"}},
	},
}

// chromiumParserCases mirrors the pure unit tests Chromium runs against its
// event-source wire codec (event_source_parser_test.cc). They pin parser
// behaviors the WPT corpus only observes indirectly.
//
// https://source.chromium.org/chromium/chromium/src/+/main:third_party/blink/renderer/core/event_source_parser_test.cc
var chromiumParserCases = []sseCase{
	{
		name: "LastEventIdShouldNotBeReset",
		url:  "chromium:event_source_parser_test.cc LastEventIdShouldNotBeReset",
		// The last event ID persists across frames until the next id field:
		// the third event reports "2", and its frame never restated the id.
		wire: "data:1" + "\n" + "id:1" + "\n\n" +
			"id:2" + "\n" + "data:2" + "\n\n" +
			"data:3" + "\n\n",
		want: []wantEvent{
			{Data: "1", ID: "1"},
			{Data: "2", ID: "2"},
			{Data: "3", ID: "2"},
		},
	},
	{
		name: "LastEventIdCanBeUpdatedEvenWhenDataIsEmpty",
		url:  "chromium:event_source_parser_test.cc LastEventIdCanBeUpdatedEvenWhenDataIsEmpty",
		// An id line updates the buffer in a frame that never dispatches, and
		// an empty id value resets the buffer to "".
		wire: "id:9" + "\n\n" +
			"data:x" + "\n\n" +
			"id:" + "\n\n" +
			"data:y" + "\n\n",
		want: []wantEvent{
			{Data: "x", ID: "9"},
			{Data: "y", ID: ""},
		},
	},
	{
		name: "RetryTakesEffectEvenWhenNotDispatching",
		url:  "chromium:event_source_parser_test.cc RetryTakesEffectEvenWhenNotDispatching",
		// A retry line in a dataless frame still updates the reconnection
		// time; the next dispatched event reports it.
		wire: "retry:2500" + "\n\n" +
			"data:x" + "\n\n",
		want: []wantEvent{{Data: "x", Retry: 2500}},
	},
	{
		name: "NonDigitRetryShouldBeIgnored",
		url:  "chromium:event_source_parser_test.cc NonDigitRetryShouldBeIgnored",
		// Values (after the spec's single-space strip) that are not pure
		// ASCII digits are ignored and do not reset a previously set value.
		// " 1234" survives the strip as "1234" and therefore counts — a
		// single space after the colon is stripped per spec § 9.2.6.
		wire: "retry:2000" + "\n" +
			"retry:a0" + "\n" +
			"retry:1234x" + "\n" +
			"retry:0x10" + "\n" +
			"retry:1.5" + "\n" +
			"retry:+1" + "\n" +
			"retry:1_000" + "\n" +
			"retry: 1234" + "\n\n" +
			"data:x" + "\n\n",
		// The final "retry: 1234" line: single space stripped → valid → 1234.
		want: []wantEvent{{Data: "x", Retry: 1234}},
	},
	{
		name: "UnrecognizedFieldShouldBeIgnored",
		url:  "chromium:event_source_parser_test.cc UnrecognizedFieldShouldBeIgnored",
		wire: "data:1" + "\n" + "unknown: x" + "\n" + "data:2" + "\n\n",
		want: []wantEvent{{Data: "1\n2"}},
	},
	{
		name: "EventTypeShouldBeReset",
		url:  "chromium:event_source_parser_test.cc EventTypeShouldBeReset",
		wire: "event: x" + "\n" + "data:1" + "\n\n" +
			"data:2" + "\n\n",
		want: []wantEvent{
			{Type: "x", Data: "1"},
			{Data: "2"},
		},
	},
	{
		name: "DataShouldBeReset",
		url:  "chromium:event_source_parser_test.cc DataShouldBeReset",
		wire: "data:1" + "\n\n" +
			"data:2" + "\n\n",
		want: []wantEvent{
			{Data: "1"},
			{Data: "2"},
		},
	},
	{
		name: "TrailingCREndsLineAtEOF",
		url:  "derived:§9.2.5 end-of-line = cr lf / cr / lf",
		// A lone CR terminates the line even when EOF follows immediately, so
		// this frame's data line is complete — but the blank line required to
		// dispatch it never arrives, and the frame is discarded at EOF.
		wire: "data: x" + "\r",
		want: nil,
	},
}

// TestWPTFormatCorpus runs every transcribed conformance vector through the
// reader and asserts the exact observable output.
func TestWPTFormatCorpus(t *testing.T) {
	t.Parallel()

	for _, sseCase := range allConformanceCases() {
		t.Run(sseCase.name, func(t *testing.T) {
			t.Parallel()

			events, err := datastartest.ReadEvents(strings.NewReader(sseCase.wire))
			if err != nil {
				t.Fatalf("%s: read events: %v", sseCase.url, err)
			}

			requireEventsMatch(t, sseCase, events)
		})
	}
}

// allConformanceCases is the full corpus: WPT transcriptions, spec § 9.2.6
// examples, and Chromium parser unit cases.
func allConformanceCases() []sseCase {
	total := len(wptCorpus) + len(specExamples) + len(chromiumParserCases)

	all := make([]sseCase, 0, total)
	all = append(all, wptCorpus...)
	all = append(all, specExamples...)
	all = append(all, chromiumParserCases...)

	return all
}

// dataOf joins an event's data lines the way a browser's MessageEvent.data
// joins them (the spec's data buffer with one trailing LF stripped).
func dataOf(evt datastartest.Event) string {
	return strings.Join(evt.DataLines, "\n")
}

// requireEventsMatch asserts the parsed events equal the vector's expected
// observable output, with a failure message that includes the wire bytes.
func requireEventsMatch(t *testing.T, vector sseCase, got []datastartest.Event) {
	t.Helper()

	if len(got) != len(vector.want) {
		t.Fatalf("%s: event count: got %d, want %d\nwire: %q\ngot:\n%s",
			vector.url, len(got), len(vector.want), vector.wire, datastartest.EventsString(got))
	}

	for i, want := range vector.want {
		evt := got[i]

		if evt.Type != want.Type {
			t.Errorf("%s: event[%d] type: got %q, want %q", vector.url, i, evt.Type, want.Type)
		}

		if data := dataOf(evt); data != want.Data {
			t.Errorf("%s: event[%d] data: got %q, want %q", vector.url, i, data, want.Data)
		}

		if evt.ID != want.ID {
			t.Errorf("%s: event[%d] last event id: got %q, want %q", vector.url, i, evt.ID, want.ID)
		}

		if evt.Retry != want.Retry {
			t.Errorf("%s: event[%d] retry: got %d, want %d", vector.url, i, evt.Retry, want.Retry)
		}
	}
}
