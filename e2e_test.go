package datastar_test

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

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
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// --- Verify HTTP headers ---
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type: got %q, want %q", ct, "text/event-stream")
	}

	if cc := resp.Header.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control: got %q, want %q", cc, "no-cache")
	}

	if conn := resp.Header.Get("Connection"); conn != "keep-alive" {
		t.Errorf("Connection: got %q, want %q", conn, "keep-alive")
	}

	// --- Parse SSE events from wire ---
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	type sseEvent struct {
		eventType string
		dataLines []string
	}

	var events []sseEvent
	var current *sseEvent

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line terminates the current event
			if current != nil {
				events = append(events, *current)
				current = nil
			}

			continue
		}

		if current == nil {
			current = &sseEvent{}
		}

		switch {
		case strings.HasPrefix(line, "event: "):
			current.eventType = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			current.dataLines = append(current.dataLines, strings.TrimPrefix(line, "data: "))
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner: %v", err)
	}

	// Should have 4 events
	if len(events) != 4 {
		t.Fatalf("event count: got %d, want 4", len(events))
	}

	// --- Event 1: patch-elements with selector + mode + elements ---
	ev := events[0]
	if ev.eventType != "datastar-patch-elements" {
		t.Errorf("event[0] type: got %q, want %q", ev.eventType, "datastar-patch-elements")
	}

	if !sliceContains(ev.dataLines, "selector #feed") {
		t.Errorf("event[0] should contain selector #feed; got %v", ev.dataLines)
	}

	if !sliceContains(ev.dataLines, "mode append") {
		t.Errorf("event[0] should contain mode append; got %v", ev.dataLines)
	}

	if !sliceContains(ev.dataLines, "elements <div>hello</div>") {
		t.Errorf("event[0] should contain elements line; got %v", ev.dataLines)
	}

	// --- Event 2: patch-signals ---
	ev = events[1]
	if ev.eventType != "datastar-patch-signals" {
		t.Errorf("event[1] type: got %q, want %q", ev.eventType, "datastar-patch-signals")
	}

	if len(ev.dataLines) == 0 {
		t.Error("event[1] should have signal data lines")
	}

	// The signals dataline should contain the JSON
	foundSignals := false
	for _, dl := range ev.dataLines {
		if strings.HasPrefix(dl, "signals ") && strings.Contains(dl, "count") {
			foundSignals = true
			break
		}
	}

	if !foundSignals {
		t.Errorf("event[1] should contain signals data with count; got %v", ev.dataLines)
	}

	// --- Event 3: ExecuteScript → patch-elements (NOT execute-script) ---
	ev = events[2]
	if ev.eventType != "datastar-patch-elements" {
		t.Errorf("event[2] type: got %q, want %q (ExecuteScript uses patch-elements per SDK)",
			ev.eventType, "datastar-patch-elements")
	}

	if !sliceContains(ev.dataLines, "selector body") {
		t.Errorf("event[2] should contain selector body; got %v", ev.dataLines)
	}

	if !sliceContains(ev.dataLines, "mode append") {
		t.Errorf("event[2] should contain mode append; got %v", ev.dataLines)
	}

	// The script should be wrapped in <script> tags
	foundScript := false
	for _, dl := range ev.dataLines {
		if strings.HasPrefix(dl, "elements ") && strings.Contains(dl, "console.log('hi')") {
			foundScript = true
			break
		}
	}

	if !foundScript {
		t.Errorf("event[2] should contain script in elements line; got %v", ev.dataLines)
	}

	// --- Event 4: RemoveElement → patch-elements with mode remove ---
	ev = events[3]
	if ev.eventType != "datastar-patch-elements" {
		t.Errorf("event[3] type: got %q, want %q", ev.eventType, "datastar-patch-elements")
	}

	if !sliceContains(ev.dataLines, "selector #stale") {
		t.Errorf("event[3] should contain selector #stale; got %v", ev.dataLines)
	}

	if !sliceContains(ev.dataLines, "mode remove") {
		t.Errorf("event[3] should contain mode remove; got %v", ev.dataLines)
	}
}

func sliceContains(slice []string, want string) bool {
	for _, s := range slice {
		if s == want {
			return true
		}
	}

	return false
}
