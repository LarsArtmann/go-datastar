package broadcast_test

import (
	"context"
	"fmt"
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
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		req = req.WithContext(ctx)
		b.ServeHTTP(w, req)
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

// readFirstResponse connects to server, reads the first chunk of the body,
// and hands it to assert via the returned channel.
func readFirstResponse(t *testing.T, server *httptest.Server) <-chan string {
	t.Helper()

	out := make(chan string, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		if err != nil {
			out <- ""

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			out <- ""

			return
		}
		defer func() { _ = resp.Body.Close() }()

		buf := make([]byte, 8192)

		n, _ := resp.Body.Read(buf)
		out <- string(buf[:n])
	}()

	return out
}

func TestBroadcasterBroadcastDeliversPatch(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	server := httptest.NewServer(b)
	defer server.Close()

	bodyCh := readFirstResponse(t, server)

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	patch, err := datastar.NewSignalsPatch(map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	b.Broadcast(patch)
	b.Close()

	body := <-bodyCh
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
	disconnect := connectSubscriber(t, b)

	sigPatch, err := datastar.NewSignalsPatch(map[string]any{"step": 1})
	if err != nil {
		t.Fatalf("NewSignalsPatch: %v", err)
	}

	b.BroadcastMany(
		sigPatch,
		datastar.NewElementsPatch("<div>update</div>"),
	)

	disconnect()

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("SubscriberCount after disconnect: got %d, want 0", got)
	}
}

func TestBroadcasterBroadcastEvent(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster()

	server := httptest.NewServer(b)
	defer server.Close()

	bodyCh := readFirstResponse(t, server)

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	b.Broadcast(datastar.NewElementsPatch("<div>raw</div>"))
	b.Close()

	body := <-bodyCh
	if want := "datastar-patch-elements"; !strings.Contains(body, want) {
		t.Errorf("body %q does not contain %q", body, want)
	}

	if want := "elements <div>raw</div>"; !strings.Contains(body, want) {
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
		count int
		mu    sync.Mutex
	)

	b.OnSubscribe(func() {
		mu.Lock()
		count++
		mu.Unlock()
	})

	disconnect := connectSubscriber(t, b)

	mu.Lock()
	if count != 1 {
		mu.Unlock()
		t.Fatalf("OnSubscribe calls: got %d, want 1", count)
	}
	mu.Unlock()

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
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Last-Event-ID", "1")
	req = req.WithContext(ctx)

	done := make(chan struct{})

	go func() {
		b.ServeHTTP(w, req)
		close(done)
	}()

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	cancel()
	<-done

	body := w.Body.String()
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

	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	b.BroadcastEvent(sse.Event{Event: "cross", Data: "transport"})

	waitFor(t, "cross-transport event", func() bool {
		select {
		case evt := <-ch:
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

	ch := b.SubscribeFilter(func(evt sse.Event) bool { return evt.Event == "wanted" })
	defer b.Unsubscribe(ch)

	b.BroadcastEvent(sse.Event{Event: "skipped", Data: "no"})
	b.BroadcastEvent(sse.Event{Event: "wanted", Data: "yes"})

	evt := <-ch
	if evt.Event != "wanted" || evt.Data != "yes" {
		t.Fatalf("filtered event: got %+v, want {wanted yes}", evt)
	}

	select {
	case evt := <-ch:
		t.Fatalf("unexpected extra event: %+v", evt)
	default:
	}
}
