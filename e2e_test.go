package datastar_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// sseEvent is a single SSE event parsed from the wire format.
type sseEvent struct {
	eventType string
	dataLines []string
}

// TestE2E_HTTPRoundTrip verifies the complete SSE wire format produced by
// go-datastar through a real HTTP server and client — not a mock writer.
// This catches header, Content-Type, and framing bugs that mock-based tests
// cannot detect.
func TestE2E_HTTPRoundTrip(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		emitAllPatches(t, resp)
	}))
	defer srv.Close()

	resp, err := fetchServerResponse(t, srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	verifySSEHeaders(t, resp)
	events := parseSSEEvents(t, resp)

	if len(events) != 4 {
		t.Fatalf("event count: got %d, want 4", len(events))
	}

	verifyEventPatchElements(t, events[0], "datastar-patch-elements",
		"#feed", "append", "<div>hello</div>")
	verifyEventPatchSignals(t, events[1])
	verifyEventExecuteScript(t, events[2])
	verifyEventRemoveElement(t, events[3])
}

// emitAllPatches sends the four patches that the e2e test asserts against.
// They live in a single handler for clarity but are not part of any production
// path; this is a wire-format smoke test.
func emitAllPatches(_ *testing.T, resp *datastar.Response) {
	// 1. Patch elements with selector + mode
	_ = resp.PatchElements("<div>hello</div>",
		datastar.WithSelector("#feed"),
		datastar.WithMode(datastar.ElementPatchModeAppend),
	)

	// 2. Patch signals
	_ = resp.MarshalAndPatchSignals(map[string]any{"count": 1})

	// 3. Execute script (should be patch-elements, NOT execute-script)
	_ = resp.ExecuteScript("console.log('hi')")

	// 4. Remove element
	_ = resp.RemoveElement("#stale")
}

// fetchServerResponse issues a GET against the running test server using a
// background-derived context, satisfying noctx without leaking a CancelFunc.
func fetchServerResponse(t *testing.T, url string) (*http.Response, error) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

// verifySSEHeaders checks the transport headers that the DataStar client
// depends on for connection management and reconnection.
func verifySSEHeaders(t *testing.T, resp *http.Response) {
	t.Helper()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type: got %q, want %q", ct, "text/event-stream")
	}

	if cc := resp.Header.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control: got %q, want %q", cc, "no-cache")
	}

	if conn := resp.Header.Get("Connection"); conn != "keep-alive" {
		t.Errorf("Connection: got %q, want %q", conn, "keep-alive")
	}
}

// parseSSEEvents reads the response body and parses the SSE framing into a
// sequence of events. The scanner uses a 1 MiB max line buffer to handle any
// reasonable DataStar payload.
func parseSSEEvents(t *testing.T, resp *http.Response) []sseEvent {
	t.Helper()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		events  []sseEvent
		current *sseEvent
	)

	for scanner.Scan() {
		line := scanner.Text()
		current = advanceSSEEvent(current, line)

		if line == "" && current != nil {
			events = append(events, *current)
			current = nil
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner: %v", err)
	}

	return events
}

// advanceSSEEvent folds a single SSE line into the event being built, opening
// a new event when one is not in progress.
func advanceSSEEvent(current *sseEvent, line string) *sseEvent {
	if line == "" {
		return current
	}

	if current == nil {
		current = &sseEvent{}
	}

	if strings.HasPrefix(line, "event: ") {
		current.eventType = strings.TrimPrefix(line, "event: ")
	} else if strings.HasPrefix(line, "data: ") {
		current.dataLines = append(current.dataLines, strings.TrimPrefix(line, "data: "))
	}

	return current
}

// verifyEventPatchElements asserts the elements patch carries the expected
// selector, mode, and at least one elements dataline.
func verifyEventPatchElements(t *testing.T, event sseEvent, wantType, wantSelector, wantMode, wantElements string) {
	t.Helper()

	if event.eventType != wantType {
		t.Errorf("event type: got %q, want %q", event.eventType, wantType)
	}

	wantLines := []string{
		"selector " + wantSelector,
		"mode " + wantMode,
		"elements " + wantElements,
	}
	for _, line := range wantLines {
		if !slices.Contains(event.dataLines, line) {
			t.Errorf("event should contain %q; got %v", line, event.dataLines)
		}
	}
}

// verifyEventPatchSignals asserts the signals event carries a signals dataline
// that mentions the "count" key produced by the handler.
func verifyEventPatchSignals(t *testing.T, event sseEvent) {
	t.Helper()

	if event.eventType != "datastar-patch-signals" {
		t.Errorf("event type: got %q, want %q", event.eventType, "datastar-patch-signals")
	}

	if len(event.dataLines) == 0 {
		t.Error("event should have signal data lines")
	}

	foundSignals := false

	for _, line := range event.dataLines {
		if strings.HasPrefix(line, "signals ") && strings.Contains(line, "count") {
			foundSignals = true

			break
		}
	}

	if !foundSignals {
		t.Errorf("event should contain signals data with count; got %v", event.dataLines)
	}
}

// verifyEventExecuteScript asserts ExecuteScript uses the patch-elements event
// (selector=body, mode=append) and wraps the script in <script> elements.
func verifyEventExecuteScript(t *testing.T, event sseEvent) {
	t.Helper()

	if event.eventType != "datastar-patch-elements" {
		t.Errorf("event type: got %q, want %q (ExecuteScript uses patch-elements per SDK)",
			event.eventType, "datastar-patch-elements")
	}

	wantLines := []string{"selector body", "mode append"}
	for _, line := range wantLines {
		if !slices.Contains(event.dataLines, line) {
			t.Errorf("event should contain %q; got %v", line, event.dataLines)
		}
	}

	foundScript := false

	for _, line := range event.dataLines {
		if strings.HasPrefix(line, "elements ") && strings.Contains(line, "console.log('hi')") {
			foundScript = true

			break
		}
	}

	if !foundScript {
		t.Errorf("event should contain script in elements line; got %v", event.dataLines)
	}
}

// verifyEventRemoveElement asserts the remove call becomes patch-elements
// with selector=#stale and mode=remove.
func verifyEventRemoveElement(t *testing.T, event sseEvent) {
	t.Helper()

	if event.eventType != "datastar-patch-elements" {
		t.Errorf("event type: got %q, want %q", event.eventType, "datastar-patch-elements")
	}

	wantLines := []string{"selector #stale", "mode remove"}
	for _, line := range wantLines {
		if !slices.Contains(event.dataLines, line) {
			t.Errorf("event should contain %q; got %v", line, event.dataLines)
		}
	}
}

func sliceContains(slice []string, want string) bool {
	return slices.Contains(slice, want)
}
