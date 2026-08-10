package datastartest

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

const (
	maxLineBytes   = 1024 * 1024 // 1 MiB max single line
	initialLineCap = 64 * 1024   // 64 KiB initial buffer
)

// ReadEvents parses the SSE wire format from r and returns all decoded events.
// It reads until EOF, so the source must close or end the stream (e.g., an HTTP
// response body from a handler that sends all patches and returns).
//
// The parser handles the standard SSE fields: event, data, id, retry, and
// comment lines (starting with ":"). Each blank line dispatches the current
// event. An event without a trailing blank line at EOF is still returned.
//
// DataStar datalines are preserved individually in [Event.DataLines] with their
// key prefixes intact (e.g., "selector #feed"), so typed accessors like
// [Event.Selector] and [Event.Elements] can decode them.
func ReadEvents(r io.Reader) ([]Event, error) {
	scanner := newSSEScanner(r)

	var (
		events  []Event
		current Event
		started bool
	)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if started {
				events = append(events, current)
				current = Event{}
				started = false
			}

			continue
		}

		started = true

		applySSELine(&current, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan SSE stream: %w", err)
	}

	if started {
		events = append(events, current)
	}

	return events, nil
}

// MustReadEvents is like [ReadEvents] but calls t.Fatal on error.
func MustReadEvents(t *testing.T, r io.Reader) []Event {
	t.Helper()

	events, err := ReadEvents(r)
	if err != nil {
		t.Fatalf("read SSE events: %v", err)
	}

	return events
}

// MustReadNEvents is like [ReadNEvents] but calls t.Fatal on error.
// Use this with streaming SSE connections that do not close on their own.
func MustReadNEvents(t *testing.T, r io.Reader, count int) []Event {
	t.Helper()

	events, err := ReadNEvents(r, count)
	if err != nil {
		t.Fatalf("read %d SSE events: %v", count, err)
	}

	return events
}

// applySSELine parses a single SSE wire line and folds it into the event.
func applySSELine(evt *Event, line string) {
	if strings.HasPrefix(line, ":") {
		return // SSE comment, ignore
	}

	field, value := parseSSEField(line)

	switch field {
	case "event":
		evt.Type = value
	case "data":
		evt.DataLines = append(evt.DataLines, value)
	case "id":
		evt.ID = value
	case "retry":
		if ms, err := strconv.ParseUint(value, 10, 32); err == nil {
			evt.Retry = uint(ms)
		}
	}
}

// newSSEScanner creates a bufio.Scanner configured for SSE wire-format parsing
// with the package's standard buffer sizes. Shared by ReadEvents and readNEvents
// to keep scanner setup in one place.
func newSSEScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, initialLineCap), maxLineBytes)

	return scanner
}

// parseSSEField splits an SSE line into field name and value. Per the SSE spec,
// the value is everything after the first colon, with a single leading space
// stripped if present. Lines without a colon produce the full line as the field
// with an empty value.
func parseSSEField(line string) (string, string) {
	field, value, found := strings.Cut(line, ":")
	if !found {
		return line, ""
	}

	return field, strings.TrimPrefix(value, " ")
}
