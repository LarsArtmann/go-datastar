package datastartest_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

func TestCollect_WithPath(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(writer http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>from /events</div>", datastar.WithSelector("#feed"))
	})

	events := datastartest.Collect(t, mux, datastartest.WithPath("/events"))
	datastartest.RequireEventCount(t, events, 1)

	if got := events[0].Elements(); got != "<div>from /events</div>" {
		t.Errorf("elements: got %q, want %q", got, "<div>from /events</div>")
	}
}

func TestCollect_WithDatastarSignals(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		var signals struct {
			Filter string `json:"filter"`
		}

		// GET requests read signals from the "datastar" query parameter.
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.MarshalAndPatchSignals(map[string]any{"filter": signals.Filter})
	})

	events := datastartest.Collect(
		t,
		handler,
		datastartest.WithDatastarSignals(`{"filter":"alerts"}`),
	)

	datastartest.RequireEventCount(t, events, 1)
	datastartest.RequireSignals(t, events[0], `{"filter":"alerts"}`)
}

func TestCollect_WithHeader(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test-Trace"); got != "abc123" {
			http.Error(writer, fmt.Sprintf("missing trace header: %q", got), http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>ok</div>", datastar.WithSelector("#feed"))
	})

	events := datastartest.Collect(t, handler,
		datastartest.WithHeader("X-Test-Trace", "abc123"),
		datastartest.WithHeader("X-Test-Trace", "duplicate-appends"),
	)

	datastartest.RequireEventCount(t, events, 1)
}

func TestCollect_WithLastEventID_HeaderArrives(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		lastID := datastar.LastEventID(r)

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.MarshalAndPatchSignals(map[string]any{"lastId": lastID.Get()})
	})

	events := datastartest.Collect(t, handler, datastartest.WithLastEventID("42"))

	datastartest.RequireEventCount(t, events, 1)

	var got struct {
		LastID string `json:"lastId"`
	}

	if err := events[0].UnmarshalSignals(&got); err != nil {
		t.Fatalf("unmarshal signals: %v", err)
	}

	if got.LastID != "42" {
		t.Errorf("Last-Event-ID header: got %q, want %q", got.LastID, "42")
	}
}

func TestCollectN_WithOptions(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(writer http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Last-Event-ID"); got != "1" {
			http.Error(writer, "missing Last-Event-ID", http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		for i := range 5 {
			_ = resp.PatchElements(fmt.Sprintf("<div>%d</div>", i), datastar.WithSelector("#feed"))
		}

		<-r.Context().Done()
	})

	events := datastartest.CollectN(t, mux, 2,
		datastartest.WithPath("/stream"),
		datastartest.WithLastEventID("1"),
	)
	datastartest.RequireEventCount(t, events, 2)
}

func TestCollectPost_WithOptions(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", func(writer http.ResponseWriter, r *http.Request) {
		var signals struct {
			Name string `json:"name"`
		}

		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>"+signals.Name+"</div>", datastar.WithSelector("#greeting"))
	})

	events := datastartest.CollectPost(t, mux, `{"name":"alice"}`, datastartest.WithPath("/submit"))
	datastartest.RequireEventCount(t, events, 1)
	datastartest.RequireElements(t, events[0], "#greeting", "outer", "<div>alice</div>")
}

func TestCollectWithTimeout_WithOptions(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(writer http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>timed</div>", datastar.WithSelector("#feed"))
	})

	events := datastartest.CollectWithTimeout(t, mux, 5*time.Second, datastartest.WithPath("/events"))
	datastartest.RequireEventCount(t, events, 1)
}
