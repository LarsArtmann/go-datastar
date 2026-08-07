package datastar_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

func newTestStream() (*sse.Stream, *mockFlushWriter) {
	var buf mockFlushWriter
	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	stream := sse.NewStream(&buf, r)
	return stream, &buf
}

func TestScriptHandler_Basic(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()

	t.Run("GET returns JS", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
		}

		ct := rec.Header().Get("Content-Type")
		if ct != "text/javascript; charset=utf-8" {
			t.Errorf("Content-Type: got %q, want %q", ct, "text/javascript; charset=utf-8")
		}

		if rec.Body.Len() == 0 {
			t.Error("body should not be empty")
		}

		if !strings.HasPrefix(rec.Body.String(), "// Datastar") {
			t.Error("should start with DataStar header")
		}
	})

	t.Run("has ETag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		etag := rec.Header().Get("ETag")
		if etag == "" {
			t.Error("ETag should be set")
		}
	})

	t.Run("has Cache-Control", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		cc := rec.Header().Get("Cache-Control")
		if cc == "" {
			t.Error("Cache-Control should be set")
		}
	})
}

func TestScriptHandler_ConditionalRequest(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()

	// First request to get the ETag
	req1 := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	etag := rec1.Header().Get("ETag")

	// Second request with If-None-Match should get 304
	req2 := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Errorf("status: got %d, want %d", rec2.Code, http.StatusNotModified)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("body should be empty on 304; got %d bytes", rec2.Body.Len())
	}
}

func TestScriptHandler_RejectsPost(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()
	req := httptest.NewRequest(http.MethodPost, "/datastar.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestScriptHandlerWith(t *testing.T) {
	t.Parallel()

	customJS := []byte("// custom")
	handler := datastar.ScriptHandlerWith(customJS, "0.0.0-test")

	req := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "// custom" {
		t.Errorf("body: got %q, want %q", rec.Body.String(), "// custom")
	}
}

func TestScriptTag(t *testing.T) {
	t.Parallel()

	got := datastar.ScriptTag("/static/datastar.js")
	want := `<script type="module" src="/static/datastar.js"></script>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	v := datastar.Version()
	if v != "1.0.2" {
		t.Errorf("got %q, want %q", v, "1.0.2")
	}
}

// --- Response Tests ---

func TestResponse_PatchElements(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.PatchElements("<div>hi</div>", datastar.WithSelector("#feed")); err != nil {
		t.Fatalf("PatchElements: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "event: datastar-patch-elements") {
		t.Errorf("should contain patch-elements event; got:\n%s", output)
	}
	if !strings.Contains(output, "selector #feed") {
		t.Errorf("should contain selector; got:\n%s", output)
	}
}

func TestResponse_PatchSignals(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.PatchSignals([]byte(`{"x":1}`)); err != nil {
		t.Fatalf("PatchSignals: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "event: datastar-patch-signals") {
		t.Errorf("should contain patch-signals event; got:\n%s", output)
	}
}

func TestResponse_ExecuteScript(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.ExecuteScript("console.log('test')"); err != nil {
		t.Fatalf("ExecuteScript: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "console.log('test')") {
		t.Errorf("should contain script; got:\n%s", output)
	}
}

func TestResponse_RemoveElement(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.RemoveElement("#feed"); err != nil {
		t.Fatalf("RemoveElement: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "mode remove") {
		t.Errorf("should contain mode remove; got:\n%s", output)
	}
}

func TestResponse_ApplyPatches(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()
	resp := datastar.NewResponse(stream)

	patches := []datastar.Patch{
		datastar.NewElementsPatch("<div>1</div>"),
		datastar.NewElementsPatch("<div>2</div>"),
	}

	if err := resp.ApplyPatches(patches...); err != nil {
		t.Fatalf("ApplyPatches: %v", err)
	}

	output := buf.String()
	if strings.Count(output, "data: elements <div>1</div>") != 1 {
		t.Errorf("should contain first patch once; got:\n%s", output)
	}
	if strings.Count(output, "data: elements <div>2</div>") != 1 {
		t.Errorf("should contain second patch once; got:\n%s", output)
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()

	if err := datastar.ErrorResponse(stream, "something broke", "ERR_001"); err != nil {
		t.Fatalf("ErrorResponse: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "something broke") {
		t.Errorf("should contain error message; got:\n%s", output)
	}
	if !strings.Contains(output, "ERR_001") {
		t.Errorf("should contain error code; got:\n%s", output)
	}
}

func TestNotificationResponse(t *testing.T) {
	t.Parallel()

	stream, buf := newTestStream()

	if err := datastar.NotificationResponse(stream, "saved!", "success"); err != nil {
		t.Fatalf("NotificationResponse: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "saved!") {
		t.Errorf("should contain notification message; got:\n%s", output)
	}
	if !strings.Contains(output, "success") {
		t.Errorf("should contain kind; got:\n%s", output)
	}
}

func TestResponse_Stream(t *testing.T) {
	t.Parallel()

	stream, _ := newTestStream()
	resp := datastar.NewResponse(stream)

	if resp.Stream() != stream {
		t.Error("Stream() should return the underlying stream")
	}
}

// mockFlushWriter implements http.ResponseWriter and http.Flusher for testing.
type mockFlushWriter struct {
	bytes []byte
}

func (m *mockFlushWriter) Write(b []byte) (int, error) {
	m.bytes = append(m.bytes, b...)
	return len(b), nil
}

func (m *mockFlushWriter) WriteHeader(int) {}

func (m *mockFlushWriter) Header() http.Header { return http.Header{} }

func (m *mockFlushWriter) Flush() {}

func (m *mockFlushWriter) String() string { return string(m.bytes) }
