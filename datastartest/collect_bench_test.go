package datastartest

import (
	"bytes"
	"net/http"
	"testing"
)

// benchHandler emits count element patches, then returns (ending the
// stream) — the same shape the Collect helpers are built for.
func benchHandler(count int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		for i := 0; i < count; i++ {
			_, _ = w.Write([]byte(
				"event: datastar-patch-elements\ndata: selector #bench\ndata: elements <span>x</span>\ndata: mode replace\n\n"))
			flusher.Flush()
		}
	})
}

// BenchmarkCollect measures the full Collect round trip: httptest server,
// GET, SSE wire parsing, and DataStar dataline decoding for a fixed stream.
// It is the consumer-facing end-to-end cost of the helper.
func BenchmarkCollect(b *testing.B) {
	const events = 16
	handler := benchHandler(events)

	b.ReportAllocs()

	for b.Loop() {
		got := Collect(b, handler)
		if len(got) != events {
			b.Fatalf("Collect returned %d events; want %d", len(got), events)
		}
	}
}

// BenchmarkReadEvents measures raw SSE wire parsing plus dataline decoding
// without the HTTP layer — the parser floor under BenchmarkCollect.
func BenchmarkReadEvents(b *testing.B) {
	const frames = 16
	stream := ""
	for i := 0; i < frames; i++ {
		stream += "event: datastar-patch-elements\n" +
			"data: selector #bench\n" +
			"data: elements <span>x</span>\n" +
			"data: mode replace\n\n"
	}
	payload := []byte(stream)

	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()

	for b.Loop() {
		if _, err := ReadEvents(bytes.NewReader(payload)); err != nil {
			b.Fatalf("ReadEvents: %v", err)
		}
	}
}
