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
