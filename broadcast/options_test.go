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

// recordingStore is a broadcast.Store that records appended events and
// counts replay queries, delegating storage to an embedded MemoryStore.
type recordingStore struct {
	*datastar.MemoryStore

	mu            sync.Mutex
	appended      []sse.Event
	eventsAfter   int
	lastEventByID string
}

func newRecordingStore() *recordingStore {
	return &recordingStore{MemoryStore: datastar.NewMemoryStore(16)}
}

func (s *recordingStore) Append(evt sse.Event) {
	s.mu.Lock()
	s.appended = append(s.appended, evt)
	s.mu.Unlock()

	s.MemoryStore.Append(evt)
}

func (s *recordingStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	s.mu.Lock()
	s.eventsAfter++
	s.lastEventByID = lastID.Get()
	s.mu.Unlock()

	return s.MemoryStore.EventsAfter(lastID)
}

func (s *recordingStore) appendCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.appended)
}

func (s *recordingStore) replayQueries() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.eventsAfter
}

// numberedEvents builds n uniquely-IDed raw events ("item-1".."item-n").
func numberedEvents(n int) []sse.Event {
	evts := make([]sse.Event, 0, n)
	for i := range n {
		evts = append(evts, sse.Event{
			ID:    sse.NewEventID(strconv.Itoa(i + 1)),
			Event: "feed",
			Data:  fmt.Sprintf("item-%d", i+1),
		})
	}

	return evts
}

// replayBody connects to b with the given Last-Event-ID, waits for
// registration, disconnects, and returns the response body.
func replayBody(t *testing.T, b *broadcast.Broadcaster, lastEventID string) string {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/events", nil)
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	done := make(chan struct{})

	go func() {
		b.ServeHTTP(recorder, req)
		close(done)
	}()

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	cancel()
	<-done

	return recorder.Body.String()
}

func TestWithStoreAppendBeforeFanOut(t *testing.T) {
	t.Parallel()

	store := newRecordingStore()
	b := broadcast.NewBroadcaster(broadcast.WithStore(store))
	defer b.Close()

	sub := b.Subscribe()

	evt := sse.Event{ID: sse.NewEventID("1"), Event: "feed", Data: "item-1"}
	b.BroadcastEvent(evt)

	received, ok := <-sub
	if !ok {
		t.Fatal("subscriber channel closed before delivery")
	}

	// The ordering guarantee under test: by the time a subscriber holds the
	// event, the store already contains it — a client reconnecting at this
	// instant replays the event instead of missing it.
	if got := store.appendCount(); got != 1 {
		t.Fatalf("store append count when subscriber received event: got %d, want 1", got)
	}

	if received.ID != evt.ID || received.Data != evt.Data {
		t.Errorf("received %+v, want %+v", received, evt)
	}
}

func TestWithStoreReplayOnReconnect(t *testing.T) {
	t.Parallel()

	store := newRecordingStore()
	b := broadcast.NewBroadcaster(broadcast.WithStore(store))
	defer b.Close()

	for _, evt := range numberedEvents(3) {
		b.BroadcastEvent(evt)
	}

	body := replayBody(t, b, "1")

	for _, want := range []string{"item-2", "item-3"} {
		if !strings.Contains(body, want) {
			t.Errorf("replayed body %q does not contain %q (replay must serve from the injected store)", body, want)
		}
	}

	if strings.Contains(body, "item-1") {
		t.Errorf("replayed body %q must not contain already-seen %q", body, "item-1")
	}

	if got := store.replayQueries(); got == 0 {
		t.Error("injected store's EventsAfter was never queried — replay did not use the injected store")
	}
}

func TestWithStoreNilDisablesReplay(t *testing.T) {
	t.Parallel()

	b := broadcast.NewBroadcaster(broadcast.WithStore(nil))
	defer b.Close()

	for _, evt := range numberedEvents(3) {
		b.BroadcastEvent(evt)
	}

	if body := replayBody(t, b, "1"); strings.Contains(body, "item-") {
		t.Errorf("body %q must contain no replayed items (nil store disables replay)", body)
	}
}

// TestNewBroadcasterOptionsMatrix walks the constructor × broadcast-method
// × replay matrix, including the previously-impossible buffer-size × replay
// combination and last-option-wins store semantics.
func TestNewBroadcasterOptionsMatrix(t *testing.T) {
	t.Parallel()

	newPatches := func() (datastar.Patch, datastar.Patch) {
		sig, err := datastar.NewSignalsPatch(map[string]any{"step": 1})
		if err != nil {
			t.Fatalf("NewSignalsPatch: %v", err)
		}

		return sig, datastar.NewElementsPatch("<div>update</div>")
	}

	tests := []struct {
		name      string
		opts      []broadcast.Option
		wantItems bool
	}{
		{
			name: "no options",
		},
		{
			name: "buffer size only",
			opts: []broadcast.Option{broadcast.WithBufferSize(8)},
		},
		{
			name:      "replay capacity only",
			opts:      []broadcast.Option{broadcast.WithReplayCapacity(4)},
			wantItems: true,
		},
		{
			name: "buffer size and replay",
			opts: []broadcast.Option{
				broadcast.WithBufferSize(8),
				broadcast.WithReplayCapacity(4),
			},
			wantItems: true,
		},
		{
			name: "store overrides earlier replay capacity",
			opts: []broadcast.Option{
				broadcast.WithReplayCapacity(4),
				broadcast.WithStore(nil),
			},
		},
		{
			name: "replay capacity overrides earlier nil store",
			opts: []broadcast.Option{
				broadcast.WithStore(nil),
				broadcast.WithReplayCapacity(4),
			},
			wantItems: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := broadcast.NewBroadcaster(tt.opts...)
			defer b.Close()

			server := httptest.NewServer(b)
			defer server.Close()

			results := startReader(t, server)

			waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

			sigPatch, elPatch := newPatches()

			// Every broadcast method must deliver through the configured hub.
			b.Broadcast(sigPatch)
			b.BroadcastMany(sigPatch, elPatch)
			b.BroadcastEvent(sse.Event{ID: sse.NewEventID("1"), Event: "raw", Data: "payload"})

			replayed := replayBody(t, b, "0")
			if tt.wantItems && !strings.Contains(replayed, "raw") {
				t.Errorf("replayed body %q must replay stored items", replayed)
			}

			if !tt.wantItems && strings.Contains(replayed, "raw") {
				t.Errorf("replayed body %q must be empty (replay disabled)", replayed)
			}

			b.Close()

			body := readBody(t, results)
			for _, want := range []string{
				"datastar-patch-signals",
				"datastar-patch-elements",
				"event: raw",
				"payload",
			} {
				if !strings.Contains(body, want) {
					t.Errorf("body %q does not contain %q", body, want)
				}
			}
		})
	}
}

// TestWithHeartbeatIntervalFast pins the heartbeat option: a 20ms interval
// must produce comment pings well inside the default 15s, proving the
// option reaches the per-connection Heartbeat goroutine.
func TestWithHeartbeatIntervalFast(t *testing.T) {
	b := broadcast.NewBroadcaster(broadcast.WithHeartbeatInterval(20 * time.Millisecond))
	defer b.Close()

	server := httptest.NewServer(b)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", server.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	waitFor(t, "subscriber to connect", func() bool { return b.SubscriberCount() == 1 })

	buf := make([]byte, 512)

	type readResult struct {
		n   int
		err error
	}

	reads := make(chan readResult, 1)

	go func() {
		n, err := resp.Body.Read(buf)
		reads <- readResult{n: n, err: err}
	}()

	// Any bytes within 500ms can only come from a 20ms heartbeat ping (no
	// events are broadcast; the default 15s interval could not produce them).
	var body string

	select {
	case r := <-reads:
		if r.n == 0 {
			t.Fatalf("empty read (err %v) — option not applied", r.err)
		}

		body = string(buf[:r.n])
	case <-time.After(500 * time.Millisecond):
		t.Fatal("no heartbeat ping within 500ms — option not applied")
	}

	if !strings.Contains(body, ":") {
		t.Errorf("early bytes %q do not look like a heartbeat comment ping", body)
	}
}
