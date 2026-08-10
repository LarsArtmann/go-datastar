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
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, initialLineCap), maxLineBytes)

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
