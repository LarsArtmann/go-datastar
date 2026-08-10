package datastartest_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

func TestCollectPost_JSONSignals(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
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

	events := datastartest.CollectPost(t, handler, `{"name":"alice"}`)
	datastartest.RequireEventCount(t, events, 1)

	if got := events[0].Elements(); got != "<div>alice</div>" {
		t.Errorf("elements: got %q, want %q", got, "<div>alice</div>")
	}
}

func TestCollectWithRequest_PutWithBody(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>put received</div>", datastar.WithSelector("#result"))
	})

	events := datastartest.CollectWithRequest(t, handler, http.MethodPut,
		nil, "application/json")
	datastartest.RequireEventCount(t, events, 1)

	if got := events[0].Elements(); got != "<div>put received</div>" {
		t.Errorf("elements: got %q, want %q", got, "<div>put received</div>")
	}
}

func TestCollectN_StreamingHandler(t *testing.T) {
	t.Parallel()

	totalEvents := 10

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		for i := range totalEvents {
			_ = resp.PatchElements(
				fmt.Sprintf("<div>%d</div>", i),
				datastar.WithSelector("#feed"),
			)
		}

		<-r.Context().Done()
	})

	want := 3
	events := datastartest.CollectN(t, handler, want)
	datastartest.RequireEventCount(t, events, want)

	for i, evt := range events {
		wantHTML := fmt.Sprintf("<div>%d</div>", i)
		if got := evt.Elements(); got != wantHTML {
			t.Errorf("event[%d] elements: got %q, want %q", i, got, wantHTML)
		}
	}
}

func TestCollectN_AllEvents(t *testing.T) {
	t.Parallel()

	handler := helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.PatchElements("<div>2</div>", datastar.WithSelector("#b"))
	})

	events := datastartest.CollectN(t, handler, 2)
	datastartest.RequireEventCount(t, events, 2)
}

var errTestReadFailure = errors.New("test: simulated read failure")

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errTestReadFailure
}

func TestReadEvents_FailingReader(t *testing.T) {
	t.Parallel()

	events, err := datastartest.ReadEvents(failingReader{})
	if err == nil {
		t.Fatal("expected error from ReadEvents with failing reader")
	}

	if len(events) != 0 {
		t.Errorf("failing reader should return 0 events; got %d", len(events))
	}
}

func TestCollectN_ZeroCount(t *testing.T) {
	t.Parallel()

	handler := helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
	})

	events := datastartest.CollectN(t, handler, 0)

	if len(events) != 0 {
		t.Errorf("CollectN(0): got %d events, want 0", len(events))
	}
}

func TestCollectN_FewerThanRequested(t *testing.T) {
	t.Parallel()

	handler := helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))
		_ = resp.PatchElements("<div>2</div>", datastar.WithSelector("#b"))
	})

	events := datastartest.CollectN(t, handler, 10)

	if len(events) != 2 {
		t.Errorf("CollectN(10) with 2-event handler: got %d events, want 2", len(events))
	}
}

func TestCollectWithTimeout(t *testing.T) {
	t.Parallel()

	handler := helperHandler(func(resp *datastar.Response) {
		_ = resp.PatchElements("<div>hi</div>", datastar.WithSelector("#feed"))
	})

	events := datastartest.CollectWithTimeout(t, handler, 5*time.Second)
	datastartest.RequireEventCount(t, events, 1)

	if got := events[0].Elements(); got != "<div>hi</div>" {
		t.Errorf("elements: got %q, want %q", got, "<div>hi</div>")
	}
}

func TestCollectWithTimeout_StreamingReturnsPartial(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(writer, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)
		_ = resp.PatchElements("<div>1</div>", datastar.WithSelector("#a"))

		<-r.Context().Done()
	})

	events := datastartest.CollectWithTimeout(t, handler, 200*time.Millisecond)

	if len(events) != 1 {
		t.Errorf("expected 1 event before timeout; got %d", len(events))
	}
}
