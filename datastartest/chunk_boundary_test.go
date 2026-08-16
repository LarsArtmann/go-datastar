package datastartest_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

// TestParserChunkBoundaryIndependence runs the entire conformance corpus
// (WPT vectors, spec § 9.2.6 examples, Chromium parser cases) through readers
// that deliver the stream in small fixed-size chunks, asserting identical
// results to the single-read parse. This mirrors Chromium's
// event_source_parser_test.cc EnqueueOneByOne trick: the parser must be
// independent of TCP chunking.
func TestParserChunkBoundaryIndependence(t *testing.T) {
	t.Parallel()

	for _, chunkSize := range []int{1, 2, 3, 5, 7, 4096} {
		for _, tc := range allConformanceCases() {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				whole, err := datastartest.ReadEvents(strings.NewReader(tc.wire))
				if err != nil {
					t.Fatalf("%s: baseline parse: %v", tc.url, err)
				}

				chunked, err := datastartest.ReadEvents(&chunkedReader{data: []byte(tc.wire), size: chunkSize})
				if err != nil {
					t.Fatalf("%s: chunked parse (size %d): %v", tc.url, chunkSize, err)
				}

				if len(chunked) != len(whole) {
					t.Fatalf("%s: chunk size %d: event count: got %d, want %d\nwire: %q",
						tc.url, chunkSize, len(chunked), len(whole), tc.wire)
				}

				for i := range whole {
					a, b := whole[i], chunked[i]

					if a.Type != b.Type || a.ID != b.ID || a.Retry != b.Retry || dataOf(a) != dataOf(b) {
						t.Fatalf("%s: chunk size %d: event[%d] differs:\nwhole:  %+v\nchunked:%+v\nwire: %q",
							tc.url, chunkSize, i, a, b, tc.wire)
					}
				}
			})
		}
	}
}
