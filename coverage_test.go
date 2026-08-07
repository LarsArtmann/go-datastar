package datastar_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

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

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func (failingWriter) WriteHeader(int)           {}
func (failingWriter) Header() http.Header           { return make(http.Header) }
func (failingWriter) Flush()                    {}

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
