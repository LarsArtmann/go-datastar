package datastar_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/static"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// errSomethingFailed is a static sentinel used to fake a ConsoleError source.
var errSomethingFailed = errors.New("something failed")

func newTestStream() (*sse.Stream, *mockFlushWriter) {
	var buf mockFlushWriter

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	stream := sse.NewStream(&buf, r)

	return stream, &buf
}

// assertContains returns an output assertion that verifies the SSE output
// contains every expected substring.
func assertContains(want ...string) func(testing.TB, string) {
	return func(tb testing.TB, output string) {
		tb.Helper()

		for _, w := range want {
			if !strings.Contains(output, w) {
				tb.Errorf("should contain %q; got:\n%s", w, output)
			}
		}
	}
}

// mockTemplComponent implements datastar.TemplComponent for testing.
type mockTemplComponent struct {
	html string
}

func (m *mockTemplComponent) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, m.html)

	return err
}

func TestScriptHandler_Basic(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()

	t.Run("GET returns JS", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/datastar.js",
			nil,
		)
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

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/datastar.js",
			nil,
		)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		etag := rec.Header().Get("ETag")
		if etag == "" {
			t.Error("ETag should be set")
		}
	})

	t.Run("has Cache-Control", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/datastar.js",
			nil,
		)
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
	req1 := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/datastar.js",
		nil,
	)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	etag := rec1.Header().Get("ETag")

	// Second request with If-None-Match should get 304
	req2 := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/datastar.js",
		nil,
	)
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
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/datastar.js",
		nil,
	)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestScriptHandler_HEADReturnsNoBody(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()

	// GET to learn the expected Content-Length and ETag.
	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/datastar.js",
		nil,
	)
	handler.ServeHTTP(getRec, getReq)

	// HEAD must return 200, the same Content-Length and ETag, but no body
	// (RFC 7231 §4.3.2).
	headRec := httptest.NewRecorder()
	headReq := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodHead,
		"/datastar.js",
		nil,
	)
	handler.ServeHTTP(headRec, headReq)

	if headRec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", headRec.Code, http.StatusOK)
	}

	if headRec.Body.Len() != 0 {
		t.Errorf("HEAD must not return a body; got %d bytes", headRec.Body.Len())
	}

	gotLen := headRec.Header().Get("Content-Length")
	wantLen := getRec.Header().Get("Content-Length")

	if gotLen == "" || gotLen != wantLen {
		t.Errorf("Content-Length: got %q, want %q (non-empty, matching GET)", gotLen, wantLen)
	}

	if headRec.Header().Get("ETag") == "" {
		t.Error("HEAD must include ETag")
	}
}

func TestScriptHandlerWith(t *testing.T) {
	t.Parallel()

	customJS := []byte("// custom")
	handler := datastar.ScriptHandlerWith(customJS, "0.0.0-test")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/datastar.js", nil)
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
	if v != "1.0.3" {
		t.Errorf("got %q, want %q", v, "1.0.3")
	}
}

// --- Response Tests ---

func TestResponse_Actions(t *testing.T) {
	t.Parallel()

	templComp := &mockTemplComponent{html: "<div>templ-rendered</div>"}

	tests := []struct {
		name   string
		run    func(*sse.Stream) error
		assert func(testing.TB, string)
	}{
		{
			name: "PatchElements",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).
					PatchElements("<div>hi</div>", datastar.WithSelector("#feed"))
			},
			assert: assertContains("event: datastar-patch-elements", "selector #feed"),
		},
		{
			name: "PatchSignals",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).PatchSignals([]byte(`{"x":1}`))
			},
			assert: assertContains("event: datastar-patch-signals"),
		},
		{
			name: "ExecuteScript",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).ExecuteScript("console.log('test')")
			},
			assert: assertContains("console.log('test')"),
		},
		{
			name: "RemoveElement",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).RemoveElement("#feed")
			},
			assert: assertContains("mode remove"),
		},
		{
			name: "ReplaceURLQuerystring",
			run: func(stream *sse.Stream) error {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/search?q=old",
					nil,
				)

				return datastar.NewResponse(stream).
					ReplaceURLQuerystring(req, url.Values{"q": {"new"}, "page": {"2"}})
			},
			assert: assertContains(
				`window.history.replaceState({}, "", "/search?page=2&q=new")`,
			),
		},
		{
			name: "ApplyPatches",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).ApplyPatches(
					datastar.NewElementsPatch("<div>1</div>"),
					datastar.NewElementsPatch("<div>2</div>"),
				)
			},
			assert: func(tb testing.TB, output string) {
				tb.Helper()

				if strings.Count(output, "data: elements <div>1</div>") != 1 {
					tb.Errorf("should contain first patch once; got:\n%s", output)
				}

				if strings.Count(output, "data: elements <div>2</div>") != 1 {
					tb.Errorf("should contain second patch once; got:\n%s", output)
				}
			},
		},
		{
			name: "PatchElementsTempl",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).
					PatchElementsTempl(templComp, datastar.WithSelectorID("main"))
			},
			assert: assertContains("elements <div>templ-rendered</div>"),
		},
		{
			name: "MarshalAndPatchSignals",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).
					MarshalAndPatchSignals(map[string]any{"count": 42, "name": "test"})
			},
			assert: assertContains("event: datastar-patch-signals", "count", "42"),
		},
		{
			name: "RemoveElementByID",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).RemoveElementByID("todo-1")
			},
			assert: assertContains("selector #todo-1", "mode remove"),
		},
		{
			name: "Redirect",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).Redirect("/dashboard")
			},
			assert: assertContains("window.location.href", "/dashboard"),
		},
		{
			name: "ConsoleLog",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).ConsoleLog("debug info")
			},
			assert: assertContains("console.log", "debug info"),
		},
		{
			name: "ConsoleError",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).ConsoleError(errSomethingFailed)
			},
			assert: assertContains("console.error", "something failed"),
		},
		{
			name: "DispatchCustomEvent",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).
					DispatchCustomEvent("item-added", map[string]any{"id": 1})
			},
			assert: assertContains("CustomEvent", "item-added"),
		},
		{
			name: "ReplaceURL",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).ReplaceURL(url.URL{Path: "/new-path"})
			},
			assert: assertContains("replaceState", "/new-path"),
		},
		{
			name: "Prefetch",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).Prefetch("/page1", "/page2")
			},
			assert: assertContains("prefetch", "/page1", "/page2"),
		},
		{
			name: "Send",
			run: func(s *sse.Stream) error {
				return datastar.NewResponse(s).Send(sse.Event{Event: "custom", Data: "raw-data"})
			},
			assert: assertContains("event: custom", "raw-data"),
		},
		{
			name: "ErrorResponse",
			run: func(s *sse.Stream) error {
				return datastar.ErrorResponse(s, "something broke", "ERR_001")
			},
			assert: assertContains("something broke", "ERR_001"),
		},
		{
			name: "NotificationResponse",
			run: func(s *sse.Stream) error {
				return datastar.NotificationResponse(s, "saved!", "success")
			},
			assert: assertContains("saved!", "success"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stream, buf := newTestStream()

			if err := tt.run(stream); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}

			tt.assert(t, buf.String())
		})
	}
}

func TestErrorResponseFromError(t *testing.T) {
	t.Parallel()

	t.Run("Rejection", func(t *testing.T) {
		t.Parallel()

		err := errorfamily.NewRejection("test.bad_input", "invalid input")
		stream, buf := newTestStream()

		if sendErr := datastar.ErrorResponseFromError(stream, err); sendErr != nil {
			t.Fatalf("ErrorResponseFromError: %v", sendErr)
		}

		check := assertContains(
			"event: datastar-patch-signals",
			"invalid input",
			"test.bad_input",
			`"rejection"`,
			`"retryable":false`,
		)
		check(t, buf.String())
	})

	t.Run("Transient", func(t *testing.T) {
		t.Parallel()

		err := errorfamily.NewTransient("test.io_failed", "connection reset")
		stream, buf := newTestStream()

		if sendErr := datastar.ErrorResponseFromError(stream, err); sendErr != nil {
			t.Fatalf("ErrorResponseFromError: %v", sendErr)
		}

		check := assertContains(
			"event: datastar-patch-signals",
			"connection reset",
			"test.io_failed",
			`"transient"`,
			`"retryable":true`,
		)
		check(t, buf.String())
	})

	t.Run("NonErrorFamilyError", func(t *testing.T) {
		t.Parallel()

		// Classify defaults to Transient (fail-open) for non-errorfamily errors.
		stream, buf := newTestStream()

		if sendErr := datastar.ErrorResponseFromError(stream, errSomethingFailed); sendErr != nil {
			t.Fatalf("ErrorResponseFromError: %v", sendErr)
		}

		check := assertContains(
			"event: datastar-patch-signals",
			"something failed",
			`"transient"`,
			`"retryable":true`,
		)
		check(t, buf.String())
	})
}

func TestResponse_Stream(t *testing.T) {
	t.Parallel()

	stream, _ := newTestStream()
	resp := datastar.NewResponse(stream)

	if resp.Stream() != stream {
		t.Error("Stream() should return the underlying stream")
	}
}

func TestNewResponseFromHTTP(t *testing.T) {
	t.Parallel()

	var buf mockFlushWriter

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)

	resp := datastar.NewResponseFromHTTP(&buf, req)
	if resp == nil {
		t.Fatal("NewResponseFromHTTP returned nil")
	}

	if err := resp.PatchElements("<div>test</div>"); err != nil {
		t.Fatalf("PatchElements via HTTP response: %v", err)
	}

	if !strings.Contains(buf.String(), "datastar-patch-elements") {
		t.Errorf("should contain patch-elements event; got:\n%s", buf.String())
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

func TestStaticVersionConsistency(t *testing.T) {
	t.Parallel()

	if datastar.DatastarJSVersion != static.Version {
		t.Errorf("DatastarJSVersion (%q) != static.Version (%q)",
			datastar.DatastarJSVersion, static.Version)
	}
}

func TestScriptHandler_ServesStaticBytes(t *testing.T) {
	t.Parallel()

	handler := datastar.ScriptHandler()
	srv := httptest.NewServer(handler)

	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET ScriptHandler: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	staticBytes := static.Bytes()
	if len(body) != len(staticBytes) {
		t.Fatalf("body length: got %d, want %d (static.Bytes)", len(body), len(staticBytes))
	}

	for i, b := range body {
		if b != staticBytes[i] {
			t.Fatalf("byte %d: got %x, want %x", i, b, staticBytes[i])
		}
	}
}
