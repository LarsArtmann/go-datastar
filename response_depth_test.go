package datastar_test

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/larsartmann/go-datastar"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// signalsPayload decodes the patch-signals payload from raw SSE output. The
// signals datalines are split on "\n" and every line is emitted unconditionally,
// so reassembly is a plain strip-and-join of the "signals " prefixed lines.
func signalsPayload(tb testing.TB, output string) map[string]any {
	tb.Helper()

	var raw strings.Builder

	for line := range strings.SplitSeq(output, "\n") {
		if payload, ok := strings.CutPrefix(line, "data: signals "); ok {
			raw.WriteString(payload)
			raw.WriteString("\n")
		}
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw.String()), &decoded); err != nil {
		tb.Fatalf("decode signals payload %q: %v", raw.String(), err)
	}

	return decoded
}

// TestErrorResponseFromError_TypedPayload asserts the decoded JSON fields, not
// substrings — so key renames or shape changes fail loudly and specifically.
func TestErrorResponseFromError_TypedPayload(t *testing.T) {
	t.Parallel()

	err := errorfamily.NewRejection("test.bad_input", "invalid input")
	stream, buf := newTestStream()

	if sendErr := datastar.ErrorResponseFromError(stream, err); sendErr != nil {
		t.Fatalf("ErrorResponseFromError: %v", sendErr)
	}

	payload := signalsPayload(t, buf.String())

	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("payload should carry an error object; got %v", payload)
	}

	// message is err.Error() — errorfamily's Error() includes the
	// "[family:code] " prefix, so it repeats the family/code fields.
	if errObj["message"] != "[rejection:test.bad_input] invalid input" {
		t.Errorf("message: got %v", errObj["message"])
	}

	if errObj["code"] != "test.bad_input" {
		t.Errorf("code: got %v", errObj["code"])
	}

	if errObj["family"] != "rejection" {
		t.Errorf("family: got %v", errObj["family"])
	}

	retryable, ok := errObj["retryable"].(bool)
	if !ok || retryable {
		t.Errorf("retryable: got %v (want false)", errObj["retryable"])
	}

	if status, ok := errObj["httpStatus"].(float64); !ok || int(status) != http.StatusBadRequest {
		t.Errorf("httpStatus: got %v (want 400)", errObj["httpStatus"])
	}
}

// TestErrorResponseFromError_NilError pins the guard: a nil error is caller
// misuse and must come back as a classified Rejection, never a panic.
func TestErrorResponseFromError_NilError(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()

	sendErr := datastar.ErrorResponseFromError(stream, nil)
	if sendErr == nil {
		t.Fatal("expected a classified error for a nil error input, got nil")
	}

	if got := errorfamily.Classify(sendErr); got != errorfamily.Rejection {
		t.Errorf("nil error should classify as Rejection, got %v", got)
	}

	if len(buf.bytes) != 0 {
		t.Errorf("nothing should be sent on a nil error input; sent %d bytes", len(buf.bytes))
	}
}

// TestErrorResponseFromError_NilDetailDispatchCustomEvent pins that a nil
// detail marshals to JSON null and the event still dispatches.
func TestErrorResponseFromError_NilDetailDispatchCustomEvent(t *testing.T) {
	t.Parallel()

	patch, err := datastar.NewDispatchCustomEventPatch("saved", nil)
	if err != nil {
		t.Fatalf("NewDispatchCustomEventPatch(nil detail): %v", err)
	}

	evt := patch.Event()

	if !strings.Contains(evt.Data, "detail: null") {
		t.Errorf("nil detail should marshal to null; data: %q", evt.Data)
	}
}

// TestErrorResponseAndNotification_EdgeCases covers empty, unicode, very long,
// and empty-code payloads with decoded-JSON assertions.
func TestErrorResponseAndNotification_EdgeCases(t *testing.T) {
	t.Parallel()

	longMessage := strings.Repeat("x", 10_000)

	t.Run("ErrorResponse empty message and code", func(t *testing.T) {
		t.Parallel()

		stream, buf := newTestStream()
		if err := datastar.ErrorResponse(stream, "", ""); err != nil {
			t.Fatalf("ErrorResponse: %v", err)
		}

		errObj := signalsPayload(t, buf.String())["error"].(map[string]any)
		if errObj["message"] != "" || errObj["code"] != "" {
			t.Errorf("empty fields should survive the round trip; got %v", errObj)
		}
	})

	t.Run("ErrorResponse unicode message", func(t *testing.T) {
		t.Parallel()

		stream, buf := newTestStream()

		msg := "Fehler: Zahlungsdaten ungültig — Verbindungsabbruch 🚨"
		if err := datastar.ErrorResponse(stream, msg, "pay.unicode"); err != nil {
			t.Fatalf("ErrorResponse: %v", err)
		}

		errObj := signalsPayload(t, buf.String())["error"].(map[string]any)
		if errObj["message"] != msg {
			t.Errorf("unicode message should survive; got %q", errObj["message"])
		}
	})

	t.Run("ErrorResponse very long message", func(t *testing.T) {
		t.Parallel()

		stream, buf := newTestStream()
		if err := datastar.ErrorResponse(stream, longMessage, "pay.long"); err != nil {
			t.Fatalf("ErrorResponse: %v", err)
		}

		errObj := signalsPayload(t, buf.String())["error"].(map[string]any)
		if errObj["message"] != longMessage {
			t.Errorf("long message truncated: got %d bytes, want %d",
				len(errObj["message"].(string)), len(longMessage))
		}
	})

	t.Run("NotificationResponse kinds", func(t *testing.T) {
		t.Parallel()

		stream, buf := newTestStream()
		if err := datastar.NotificationResponse(stream, "Saved", "success"); err != nil {
			t.Fatalf("NotificationResponse: %v", err)
		}

		noteObj := signalsPayload(t, buf.String())["notification"].(map[string]any)
		if noteObj["message"] != "Saved" || noteObj["kind"] != "success" {
			t.Errorf("notification fields: got %v", noteObj)
		}

		if _, ok := noteObj["time"].(float64); !ok {
			t.Errorf("notification should carry a numeric time; got %v", noteObj["time"])
		}
	})
}

// TestResponse_ConcurrentMethods documents that a single Response may be used
// from multiple goroutines: all events must arrive exactly once. Run under
// -race in the standard suite.
func TestResponse_ConcurrentMethods(t *testing.T) {
	t.Parallel()

	const goroutines = 16

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	var waitGroup sync.WaitGroup
	for i := range goroutines {
		waitGroup.Go(func() {
			if err := resp.PatchElements("<div>event</div>", datastar.WithSelectorf("#n%d", i)); err != nil {
				t.Errorf("PatchElements: %v", err)
			}
		})
	}

	waitGroup.Wait()

	if got := strings.Count(buf.String(), "event: datastar-patch-elements"); got != goroutines {
		t.Errorf("sent %d events, want %d", got, goroutines)
	}
}

// TestResponse_LargeMultiLineElements pins the splitting rule: elements split
// on "\n" and every source line becomes its own "data: elements" dataline.
func TestResponse_LargeMultiLineElements(t *testing.T) {
	t.Parallel()

	const lineCount = 200

	var html strings.Builder
	for range lineCount {
		html.WriteString("<div>line ")
		html.WriteString(strings.Repeat("x", 100))
		html.WriteString("</div>\n")
	}

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.PatchElements(html.String()); err != nil {
		t.Fatalf("PatchElements: %v", err)
	}

	dataLines := 0

	for line := range strings.SplitSeq(buf.String(), "\n") {
		if strings.HasPrefix(line, "data: elements ") {
			dataLines++
		}
	}

	if dataLines != lineCount+1 { // html ends with a trailing newline → one extra empty line
		t.Errorf("got %d elements datalines, want %d", dataLines, lineCount+1)
	}
}

// TestMarshalAndPatchSignals_Nested pins nested outbound signal structures.
func TestMarshalAndPatchSignals_Nested(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"user": map[string]any{
			"name": "Ada",
			"tags": []any{"admin", "beta"},
			"prefs": map[string]any{
				"theme": "dark",
				"count": float64(42),
			},
		},
	}

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.MarshalAndPatchSignals(want); err != nil {
		t.Fatalf("MarshalAndPatchSignals: %v", err)
	}

	got := signalsPayload(t, buf.String())

	gotUser, ok := got["user"].(map[string]any)
	if !ok {
		t.Fatalf("user signal missing or not an object; got %v", got)
	}

	if gotUser["name"] != "Ada" {
		t.Errorf("user.name: got %v", gotUser["name"])
	}

	prefs, ok := gotUser["prefs"].(map[string]any)
	if !ok || prefs["theme"] != "dark" || prefs["count"] != float64(42) {
		t.Errorf("user.prefs: got %v", gotUser["prefs"])
	}
}

// TestReadSignals_SourcePrecedence pins that the HTTP method alone decides the
// source: GET/DELETE read the query parameter and ignore the body; every other
// method reads the body and ignores the query parameter.
func TestReadSignals_SourcePrecedence(t *testing.T) {
	t.Parallel()

	type payload struct {
		Source string `json:"source"`
	}

	t.Run("GET with both query and body uses the query", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet,
			"/?datastar="+`{"source":"query"}`,
			strings.NewReader(`{"source":"body"}`),
		)

		var got payload
		if err := datastar.ReadSignals(r, &got); err != nil {
			t.Fatalf("ReadSignals: %v", err)
		}

		if got.Source != "query" {
			t.Errorf("GET should read the query parameter; got %q", got.Source)
		}
	})

	t.Run("POST with both query and body uses the body", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost,
			"/?datastar="+`{"source":"query"}`,
			strings.NewReader(`{"source":"body"}`),
		)

		var got payload
		if err := datastar.ReadSignals(r, &got); err != nil {
			t.Fatalf("ReadSignals: %v", err)
		}

		if got.Source != "body" {
			t.Errorf("POST should read the body; got %q", got.Source)
		}
	})

	t.Run("DELETE with both query and body uses the query", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(
			context.Background(), http.MethodDelete,
			"/?datastar="+`{"source":"query"}`,
			strings.NewReader(`{"source":"body"}`),
		)

		var got payload
		if err := datastar.ReadSignals(r, &got); err != nil {
			t.Fatalf("ReadSignals: %v", err)
		}

		if got.Source != "query" {
			t.Errorf("DELETE should read the query parameter; got %q", got.Source)
		}
	})
}

// TestErrorResponseFromError_E2E wire shape: through a real handler + real SSE
// stream, the payload decodes with the datastartest-shaped fields.
func TestErrorResponseFromError_WireShape(t *testing.T) {
	t.Parallel()

	var buf mockFlushWriter

	handler := func(w http.ResponseWriter, _ *http.Request) {
		stream := sse.NewStream(w, httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/events", nil,
		))
		defer func() { _ = stream.Close() }()

		err := errorfamily.NewRejection("test.wire", "wire failure")
		if sendErr := datastar.ErrorResponseFromError(stream, err); sendErr != nil {
			t.Errorf("ErrorResponseFromError: %v", sendErr)
		}
	}

	handler(&buf, nil)

	output := buf.String()
	for _, want := range []string{
		"event: datastar-patch-signals",
		`"retryable":false`,
		"test.wire",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("wire output should contain %q; got:\n%s", want, output)
		}
	}
}
