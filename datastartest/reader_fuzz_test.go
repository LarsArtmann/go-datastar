package datastartest_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

// FuzzReadEvents hardens the SSE parser against malformed input. The invariant:
// ReadEvents must never panic for any input — it returns events, an error, or both.
//
// Seeds cover valid SSE, truncated streams, and edge cases. Run with
// `go test -fuzz=FuzzReadEvents` to explore the input space.
func FuzzReadEvents(f *testing.F) {
	// Valid SSE.
	f.Add([]byte("event: datastar-patch-elements\ndata: elements <div>hi</div>\n\n"))
	f.Add([]byte("event: datastar-patch-signals\ndata: signals {\"x\":1}\n\n"))

	// Edge cases.
	f.Add([]byte(""))                        // empty
	f.Add([]byte("\n\n\n"))                  // blank lines only
	f.Add([]byte("data: hello\n"))           // no event type, no blank line
	f.Add([]byte(":::"))                     // colons everywhere
	f.Add([]byte("data: \x00\x01\x02\n\n"))  // control characters
	f.Add([]byte("retry: not-a-number\n\n")) // invalid retry
	f.Add([]byte("event:\ndata:\n\n"))       // empty field values

	// Dataless frames must never dispatch (SSE spec): heartbeats, id-only,
	// retry-only, and fully-named-but-empty events.
	f.Add([]byte(": heartbeat\n\n"))
	f.Add([]byte("id: 1\n\nid: 2\n\n"))
	f.Add([]byte("retry: 3000\n\n"))
	f.Add([]byte("event: named\nid: 9\nretry: 100\n\n"))

	// Conformance parity with go-sse's ssetest (same parser under the hood;
	// 2026-08-29 port). WPT format-* vectors, sticky-id, BOM, and terminator
	// edge cases — see ssetest/reader_fuzz_test.go for the full citations.
	f.Add([]byte("data:test\r\ndata\ndata:test\r\r\n"))                     // format-newlines
	f.Add([]byte("data:\ttest\rdata: \ndata:test\n\n"))                     // format-leading-space
	f.Add([]byte("\xEF\xBB\xBFdata:1\n\n\xEF\xBB\xBFdata:2\n\ndata:3\n\n")) // format-bom
	f.Add([]byte("id:\x00\ndata:x\n\n"))                                    // format-field-id-null
	f.Add(
		[]byte("retry:2000\n\nretry\n\ndata:x\n\n"),
	) // format-field-retry-empty
	f.Add([]byte("data:1\nid:1\n\nid:2\ndata:2\n\ndata:3\n\n")) // sticky last-event-id
	f.Add([]byte("data: x\r"))                                  // trailing CR
	f.Add(
		[]byte("data: x\n"),
	) // trailing LF regression
	f.Add(
		[]byte("0data: hello\n\n"),
	) // crasher: substring but different field name — dispatches nothing

	f.Fuzz(func(t *testing.T, input []byte) {
		// The only invariant under fuzz: never panic.
		_, _ = datastartest.ReadEvents(strings.NewReader(string(input)))
	})
}

// BenchmarkReadEvents measures the throughput of the SSE parser on a realistic
// DataStar patch stream (10 elements + signals events).
func BenchmarkReadEvents(b *testing.B) {
	var builder strings.Builder

	for i := range 10 {
		idx := strconv.Itoa(i)

		builder.WriteString("event: datastar-patch-elements\n")
		builder.WriteString("data: selector #item-" + idx + "\n")
		builder.WriteString("data: mode append\n")
		builder.WriteString("data: elements <div class=\"item\">" + idx + "</div>\n\n")

		builder.WriteString("event: datastar-patch-signals\n")
		builder.WriteString(
			"data: signals {\"count\":" + idx + ",\"name\":\"item-" + idx + "\"}\n\n",
		)
	}

	input := builder.String()

	b.SetBytes(int64(len(input)))
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = datastartest.ReadEvents(strings.NewReader(input))
	}
}
