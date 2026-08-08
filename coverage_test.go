package datastar_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

var errWriteFailed = errors.New("write failed")

func TestOptionConstructors(t *testing.T) {
	t.Parallel()

	t.Run("WithScriptRetryDuration", func(t *testing.T) {
		t.Parallel()

		d := 7500 * time.Millisecond
		patch := datastar.NewScriptPatch("console.log(1)", datastar.WithScriptRetryDuration(d))

		if patch.RetryDuration != d {
			t.Errorf("RetryDuration: got %v, want %v", patch.RetryDuration, d)
		}
	})

	t.Run("WithSignalsEventID", func(t *testing.T) {
		t.Parallel()

		patch := datastar.SignalsPatch{}
		datastar.WithSignalsEventID("evt-99")(&patch)

		if patch.EventID != "evt-99" {
			t.Errorf("EventID: got %q, want %q", patch.EventID, "evt-99")
		}
	})

	t.Run("WithSignalsRetryDuration", func(t *testing.T) {
		t.Parallel()

		d := 3000 * time.Millisecond
		patch := datastar.SignalsPatch{}
		datastar.WithSignalsRetryDuration(d)(&patch)

		if patch.RetryDuration != d {
			t.Errorf("RetryDuration: got %v, want %v", patch.RetryDuration, d)
		}
	})

	t.Run("WithCustomEventBubbles", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("evt", nil)
		if err != nil {
			t.Fatal(err)
		}

		datastar.WithCustomEventBubbles(false)(&patch)

		if patch.Bubbles {
			t.Error("Bubbles: got true, want false")
		}
	})

	t.Run("WithCustomEventCancelable", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("evt", nil)
		if err != nil {
			t.Fatal(err)
		}

		datastar.WithCustomEventCancelable(false)(&patch)

		if patch.Cancelable {
			t.Error("Cancelable: got true, want false")
		}
	})

	t.Run("WithCustomEventComposed", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("evt", nil)
		if err != nil {
			t.Fatal(err)
		}

		datastar.WithCustomEventComposed(false)(&patch)

		if patch.Composed {
			t.Error("Composed: got true, want false")
		}
	})

	t.Run("WithCustomEventEventID", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("evt", nil)
		if err != nil {
			t.Fatal(err)
		}

		datastar.WithCustomEventEventID("evt-42")(&patch)

		if patch.EventID != "evt-42" {
			t.Errorf("EventID: got %q, want %q", patch.EventID, "evt-42")
		}
	})
}

// failingWriter is an http.ResponseWriter that always returns an error on Write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errWriteFailed }

func (failingWriter) WriteHeader(int) {}

func (failingWriter) Header() http.Header { return make(http.Header) }

func (failingWriter) Flush() {}

func TestWrapStreamError_ErrorPath(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequestWithContext(context.Background(), "GET", "/events", nil)
	stream := sse.NewStream(failingWriter{}, r)

	resp := datastar.NewResponse(stream)

	err := resp.PatchElements("<div>hi</div>")
	if err == nil {
		t.Fatal("expected error from failing writer")
	}
}

// TestResponse_OptionApplication covers the option-application loop in
// PatchSignals, which is skipped when called without options.
func TestResponse_OptionApplication(t *testing.T) {
	t.Parallel()

	stream, _ := newTestStream()
	resp := datastar.NewResponse(stream)

	if err := resp.PatchSignals([]byte(`{"a":1}`), datastar.WithSignalsEventID("e1")); err != nil {
		t.Fatalf("PatchSignals with option: %v", err)
	}
}

// TestResponse_ConstructionErrors covers the early-return error branches in
// Response methods that build a patch before sending it.
func TestResponse_ConstructionErrors(t *testing.T) {
	t.Parallel()

	stream, _ := newTestStream()
	resp := datastar.NewResponse(stream)

	tests := []struct {
		name string
		err  error
	}{
		{"MarshalAndPatchSignals unmarshalable value", resp.MarshalAndPatchSignals(make(chan int))},
		{
			"PatchElementsTempl render failure",
			resp.PatchElementsTempl(fakeTemplComponent{err: io.ErrUnexpectedEOF}),
		},
		{"DispatchCustomEvent empty name", resp.DispatchCustomEvent("", nil)},
	}

	for _, tt := range tests {
		if tt.err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}

// TestResponse_SendErrorPaths covers the stream-send failure branches that are
// not exercised by the success-path tests.
func TestResponse_SendErrorPaths(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	failing := sse.NewStream(failingWriter{}, r)
	resp := datastar.NewResponse(failing)

	if err := resp.ApplyPatches(datastar.NewElementsPatch("<div/>")); err == nil {
		t.Error("ApplyPatches: expected send error, got nil")
	}

	if err := resp.PatchSignals([]byte(`{"a":1}`)); err == nil {
		t.Error("PatchSignals: expected send error, got nil")
	}
}
