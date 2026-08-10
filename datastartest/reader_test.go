package datastartest_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
)

func TestReadEvents_EmptyInput(t *testing.T) {
	t.Parallel()

	events, err := datastartest.ReadEvents(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadEvents empty: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("empty input: got %d events, want 0", len(events))
	}
}

func TestReadEvents_SingleEvent(t *testing.T) {
	t.Parallel()

	input := "event: datastar-patch-elements\n" +
		"data: selector #feed\n" +
		"data: mode append\n" +
		"data: elements <div>hello</div>\n" +
		"\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	evt := events[0]
	if evt.Type != "datastar-patch-elements" {
		t.Errorf("type: got %q", evt.Type)
	}

	if len(evt.DataLines) != 3 {
		t.Fatalf("data lines: got %d, want 3", len(evt.DataLines))
	}

	if evt.DataLines[0] != "selector #feed" {
		t.Errorf("data[0]: got %q", evt.DataLines[0])
	}
}

func TestReadEvents_NoTrailingBlankLine(t *testing.T) {
	t.Parallel()

	input := "event: datastar-patch-signals\n" +
		`data: signals {"x":1}` + "\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (flushed at EOF)", len(events))
	}
}

func TestReadEvents_EventIDAndRetry(t *testing.T) {
	t.Parallel()

	input := "event: datastar-patch-elements\n" +
		"data: elements <div>hi</div>\n" +
		"id: 42\n" +
		"retry: 5000\n" +
		"\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	if events[0].ID != "42" {
		t.Errorf("ID: got %q, want %q", events[0].ID, "42")
	}

	if events[0].Retry != 5000 {
		t.Errorf("Retry: got %d, want 5000", events[0].Retry)
	}
}

func TestReadEvents_CommentLines(t *testing.T) {
	t.Parallel()

	input := ": heartbeat\n" +
		"event: datastar-patch-signals\n" +
		": another comment\n" +
		"data: signals {\"ok\":true}\n" +
		"\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	if events[0].Type != "datastar-patch-signals" {
		t.Errorf("type: got %q", events[0].Type)
	}
}

func TestReadEvents_MultipleEvents(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(
		"event: datastar-patch-elements\ndata: elements <div>1</div>\n\n" +
			"event: datastar-patch-signals\ndata: signals {\"n\":1}\n\n" +
			"event: datastar-patch-elements\ndata: elements <div>2</div>\n\n",
	)

	events, err := datastartest.ReadEvents(input)
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}

	if events[0].Elements() != "<div>1</div>" {
		t.Errorf("event[0] elements: got %q", events[0].Elements())
	}

	if events[2].Elements() != "<div>2</div>" {
		t.Errorf("event[2] elements: got %q", events[2].Elements())
	}
}

func TestReadEvents_MultiLineElements(t *testing.T) {
	t.Parallel()

	input := "event: datastar-patch-elements\n" +
		"data: elements <div>a</div>\n" +
		"data: elements <div>b</div>\n" +
		"data: elements <div>c</div>\n" +
		"\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	want := "<div>a</div>\n<div>b</div>\n<div>c</div>"
	if got := events[0].Elements(); got != want {
		t.Errorf("Elements: got %q, want %q", got, want)
	}
}

func TestParseSSEField_NoSpaceAfterColon(t *testing.T) {
	t.Parallel()

	// SSE spec: if no space after colon, value is as-is
	// "data:value" → value = "value"
	input := "event:datastar-patch-signals\ndata:{\"ok\":true}\n\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if events[0].Type != "datastar-patch-signals" {
		t.Errorf("type: got %q", events[0].Type)
	}

	if len(events[0].DataLines) != 1 || events[0].DataLines[0] != `{"ok":true}` {
		t.Errorf("data: got %v", events[0].DataLines)
	}
}

func TestParseSSEField_MultiColon(t *testing.T) {
	t.Parallel()

	// A data line with multiple colons: the value is everything after the
	// first colon, with a single leading space stripped.
	input := `event:datastar-patch-signals
data:{"json":"with:colons"}

`

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	want := `{"json":"with:colons"}`
	if got := events[0].DataLines[0]; got != want {
		t.Errorf("multi-colon data: got %q, want %q", got, want)
	}
}

func TestParseSSEField_EmptyData(t *testing.T) {
	t.Parallel()

	// "data:" with nothing after it (no space, no value) → empty data line
	input := "event:datastar-patch-signals\ndata:\n\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	if len(events[0].DataLines) != 1 {
		t.Fatalf("data lines: got %d, want 1", len(events[0].DataLines))
	}

	if events[0].DataLines[0] != "" {
		t.Errorf("empty data: got %q, want empty", events[0].DataLines[0])
	}
}
