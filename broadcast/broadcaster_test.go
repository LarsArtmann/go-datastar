package broadcast_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	datastar "github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/broadcast"
	"github.com/larsartmann/go-sse"
)

// waitFor polls cond until it returns true or the timeout elapses.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

func TestNewBroadcaster(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()
	if b == nil {
		t.Fatal("NewBroadcaster returned nil")
	}

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("initial SubscriberCount: got %d, want 0", got)
	}
}

func TestNewBroadcasterWithBufferSize(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcasterWithBufferSize(128)
	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("initial SubscriberCount: got %d, want 0", got)
	}
}

// connectSubscriber starts one SSE connection against b in a goroutine and
// returns a disconnect func. It blocks until the subscriber is registered.
func connectSubscriber(t *testing.T, b *broadcast.Broadcaster) func() {
	t.Helper()

	expected := b.SubscriberCount() + 1

	ctx, ctxCancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/events", nil)
		b.ServeHTTP(recorder, req)
		close(done)
	}()

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() >= expected })

	return func() {
		ctxCancel()
		<-done
	}
}

func TestBroadcasterSubscriberCount(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	disconnect1 := connectSubscriber(t, b)

	disconnect2 := connectSubscriber(t, b)
	if got := b.SubscriberCount(); got != 2 {
		t.Fatalf("SubscriberCount after 2 connects: got %d, want 2", got)
	}

	disconnect1()
	waitFor(t, "first disconnect", func() bool { return b.SubscriberCount() == 1 })

	disconnect2()
	waitFor(t, "second disconnect", func() bool { return b.SubscriberCount() == 0 })
}

// readResult carries the outcome of a background SSE body read.
type readResult struct {
	body string
	err  error
}

// startReader connects to server in a goroutine and accumulates the response
// body until the server closes the stream (or a 5s deadline elapses). The
// connect-then-collect split lets the test wait for the subscriber to register
// before broadcasting; draining until EOF keeps assertions deterministic — no
// chunk-splitting races.
func startReader(t *testing.T, server *httptest.Server) <-chan readResult {
	t.Helper()

	out := make(chan readResult, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		if err != nil {
			out <- readResult{err: fmt.Errorf("NewRequestWithContext: %w", err)}

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			out <- readResult{err: fmt.Errorf("Do: %w", err)}

			return
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		out <- readResult{body: string(body), err: err}
	}()

	return out
}

// readBody collects a startReader result, failing the test on read errors.
func readBody(t *testing.T, results <-chan readResult) string {
	t.Helper()

	res := <-results
	if res.err != nil {
		t.Fatalf("read body: %v", res.err)
	}

	return res.body
}

func TestBroadcasterBroadcastDeliversPatch(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	server := httptest.NewServer(b)
	defer server.Close()

	results := startReader(t, server)

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	patch, err := datastar.NewSignalsPatch(map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	b.Broadcast(patch)
	b.Close()

	body := readBody(t, results)
	if want := "datastar-patch-signals"; !strings.Contains(body, want) {
		t.Errorf("body %q does not contain %q", body, want)
	}

	if !strings.Contains(body, "hello") {
		t.Errorf("body %q does not contain %q", body, "hello")
	}
}

func TestBroadcasterBroadcastMany(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	server := httptest.NewServer(b)
	defer server.Close()

	results := startReader(t, server)

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	sigPatch, err := datastar.NewSignalsPatch(map[string]any{"step": 1})
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	b.BroadcastMany(
		sigPatch,
		datastar.NewElementsPatch("<div>update</div>"),
	)
	b.Close()

	body := readBody(t, results)
	for _, want := range []string{"datastar-patch-signals", "datastar-patch-elements", "step", "update"} {
		if !strings.Contains(body, want) {
			t.Errorf("body %q does not contain %q", body, want)
		}
	}
}

func TestBroadcasterBroadcastEvent(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	server := httptest.NewServer(b)
	defer server.Close()

	results := startReader(t, server)

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	b.BroadcastEvent(sse.Event{Event: "raw", Data: "payload"})
	b.Close()

	body := readBody(t, results)
	if want := "event: raw"; !strings.Contains(body, want) {
		t.Errorf("body %q does not contain %q", body, want)
	}

	if want := "data: payload"; !strings.Contains(body, want) {
		t.Errorf("body %q does not contain %q", body, want)
	}
}

func TestBroadcasterCloseDisconnectsAll(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	_ = connectSubscriber(t, b)
	_ = connectSubscriber(t, b)

	if got := b.SubscriberCount(); got != 2 {
		t.Fatalf("SubscriberCount after 2 connects: got %d, want 2", got)
	}

	b.Close()

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("SubscriberCount after Close: got %d, want 0", got)
	}
}

func TestBroadcasterServeHTTPLifecycle(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()
	disconnect := connectSubscriber(t, b)

	if got := b.SubscriberCount(); got != 1 {
		t.Fatalf("SubscriberCount: got %d, want 1", got)
	}

	disconnect()

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("SubscriberCount after disconnect: got %d, want 0", got)
	}
}

func TestBroadcasterShutdown(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	_ = connectSubscriber(t, b)
	if got := b.SubscriberCount(); got != 1 {
		t.Fatalf("SubscriberCount: got %d, want 1", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := b.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestBroadcasterHealth(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()
	health := b.Health()

	if health.SubscriberCount != 0 {
		t.Errorf("Health().SubscriberCount: got %d, want 0", health.SubscriberCount)
	}

	if health.Closed {
		t.Error("Health().Closed: got true, want false")
	}

	if health.Draining {
		t.Error("Health().Draining: got true, want false")
	}
}

func TestBroadcasterOnSubscribeCallback(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	var (
		count   int
		countMu sync.Mutex
	)

	b.OnSubscribe(func() {
		countMu.Lock()
		count++
		countMu.Unlock()
	})

	disconnect := connectSubscriber(t, b)

	countMu.Lock()
	if count != 1 {
		countMu.Unlock()
		t.Fatalf("OnSubscribe calls: got %d, want 1", count)
	}
	countMu.Unlock()

	disconnect()
}

func TestBroadcasterReplayOnReconnect(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcasterWithReplay(10)

	for i := range 3 {
		b.BroadcastEvent(sse.Event{
			ID:    sse.NewEventID(strconv.Itoa(i + 1)),
			Event: "feed",
			Data:  fmt.Sprintf("item-%d", i+1),
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/events", nil)
	req.Header.Set("Last-Event-ID", "1")

	done := make(chan struct{})

	go func() {
		b.ServeHTTP(recorder, req)
		close(done)
	}()

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	cancel()
	<-done

	body := recorder.Body.String()
	for _, want := range []string{"item-2", "item-3"} {
		if !strings.Contains(body, want) {
			t.Errorf("replayed body %q does not contain %q", body, want)
		}
	}

	if strings.Contains(body, "item-1") {
		t.Errorf("replayed body %q must not contain already-seen %q", body, "item-1")
	}
}

func TestBroadcasterHub(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()
	if b.Hub() == nil {
		t.Fatal("Hub returned nil")
	}

	if b.Broadcaster != b.Hub() {
		t.Error("Hub must return the embedded broadcaster")
	}
}

func TestNewBroadcasterFromHub(t *testing.T) {
	t.Parallel()

	hub := sse.NewBroadcaster[sse.Event]()

	b := broadcast.NewBroadcasterFromHub(hub)
	if b.Hub() != hub {
		t.Error("Hub must return the wrapped hub")
	}
}

func TestBroadcasterHubSharesFanOut(t *testing.T) {
	t.Parallel()

	hub := sse.NewBroadcaster[sse.Event]()
	b := broadcast.NewBroadcasterFromHub(hub)

	events := hub.Subscribe()
	defer hub.Unsubscribe(events)

	b.BroadcastEvent(sse.Event{Event: "cross", Data: "transport"})

	waitFor(t, "cross-transport event", func() bool {
		select {
		case evt := <-events:
			return evt.Event == "cross"
		default:
			return false
		}
	})
}

// The embedded hub's SubscribeFilter promotes to the broadcast Broadcaster —
// consumers can filter without unwrapping the hub.
func TestBroadcasterPromotedSubscribeFilter(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	events := b.SubscribeFilter(func(evt sse.Event) bool { return evt.Event == "wanted" })
	defer b.Unsubscribe(events)

	b.BroadcastEvent(sse.Event{Event: "skipped", Data: "no"})
	b.BroadcastEvent(sse.Event{Event: "wanted", Data: "yes"})

	evt := <-events
	if evt.Event != "wanted" || evt.Data != "yes" {
		t.Fatalf("filtered event: got %+v, want {wanted yes}", evt)
	}

	select {
	case evt := <-events:
		t.Fatalf("unexpected extra event: %+v", evt)
	default:
	}
}
