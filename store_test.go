package datastar_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

func TestMemoryStore_AppendAndLen(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(10)
	if got := store.Len(); got != 0 {
		t.Fatalf("initial Len: got %d, want 0", got)
	}

	store.Append(sse.Event{ID: sse.NewEventID("1"), Event: "test", Data: "a"})

	if got := store.Len(); got != 1 {
		t.Fatalf("after 1 append: got %d, want 1", got)
	}

	store.Append(sse.Event{ID: sse.NewEventID("2"), Event: "test", Data: "b"})

	if got := store.Len(); got != 2 {
		t.Fatalf("after 2 appends: got %d, want 2", got)
	}
}

func TestMemoryStore_EventsAfter(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(10)

	for i := range 5 {
		store.Append(sse.Event{
			ID:    sse.NewEventID(strconv.Itoa(i + 1)),
			Event: "feed",
			Data:  "item-" + strconv.Itoa(i+1),
		})
	}

	events, err := store.EventsAfter(sse.NewEventID("2"))
	if err != nil {
		t.Fatalf("EventsAfter: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("event count: got %d, want 3", len(events))
	}

	if events[0].Data != "item-3" {
		t.Errorf("first event data: got %q, want %q", events[0].Data, "item-3")
	}

	if events[2].Data != "item-5" {
		t.Errorf("last event data: got %q, want %q", events[2].Data, "item-5")
	}
}

func TestMemoryStore_EventsAfterEmptyIDReplaysAll(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(10)

	for i := range 3 {
		store.Append(sse.Event{ID: sse.NewEventID(strconv.Itoa(i + 1)), Data: "x"})
	}

	events, err := store.EventsAfter(sse.EventID{})
	if err != nil {
		t.Fatalf("EventsAfter: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("event count: got %d, want 3", len(events))
	}
}

func TestMemoryStore_NonNumericIDTreatedAsZero(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(10)
	store.Append(sse.Event{ID: sse.NewEventID("abc"), Data: "non-numeric"})
	store.Append(sse.Event{ID: sse.NewEventID("5"), Data: "numeric"})

	// Non-numeric lastID treated as seq 0; only events with seq > 0 are returned.
	// "abc" has seq 0, so it is NOT included; "5" has seq 5, so it IS.
	events, err := store.EventsAfter(sse.NewEventID("xyz"))
	if err != nil {
		t.Fatalf("EventsAfter: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("event count: got %d, want 1", len(events))
	}

	if events[0].Data != "numeric" {
		t.Errorf("event data: got %q, want %q", events[0].Data, "numeric")
	}
}

func TestMemoryStore_RingBufferEviction(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(3)

	for i := range 6 {
		store.Append(sse.Event{ID: sse.NewEventID(strconv.Itoa(i + 1)), Data: strconv.Itoa(i + 1)})
	}

	if got := store.Len(); got != 3 {
		t.Fatalf("Len after eviction: got %d, want 3", got)
	}

	events, err := store.EventsAfter(sse.EventID{})
	if err != nil {
		t.Fatalf("EventsAfter: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("event count: got %d, want 3", len(events))
	}

	if events[0].Data != "4" {
		t.Errorf("first event data: got %q, want %q", events[0].Data, "4")
	}

	if events[2].Data != "6" {
		t.Errorf("last event data: got %q, want %q", events[2].Data, "6")
	}
}

func TestMemoryStore_DefaultCapacity(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(0) // non-positive → default

	for i := range datastar.DefaultMemoryStoreCapacity + 5 {
		store.Append(sse.Event{ID: sse.NewEventID(strconv.Itoa(i + 1))})
	}

	if got := store.Len(); got != datastar.DefaultMemoryStoreCapacity {
		t.Fatalf("Len: got %d, want %d", got, datastar.DefaultMemoryStoreCapacity)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(100)

	var waitGroup sync.WaitGroup

	waitGroup.Go(func() {
		for i := range 50 {
			store.Append(sse.Event{ID: sse.NewEventID(strconv.Itoa(i + 1))})
		}
	})

	for range 10 {
		_, _ = store.EventsAfter(sse.EventID{})
	}

	waitGroup.Wait()

	if got := store.Len(); got != 50 {
		t.Fatalf("Len: got %d, want 50", got)
	}
}

func TestMemoryStore_EventsAfterUnknownID(t *testing.T) {
	t.Parallel()

	store := datastar.NewMemoryStore(10)
	store.Append(sse.Event{ID: sse.NewEventID("1")})
	store.Append(sse.Event{ID: sse.NewEventID("2")})

	events, err := store.EventsAfter(sse.NewEventID("99"))
	if err != nil {
		t.Fatalf("EventsAfter: %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("event count: got %d, want 0", len(events))
	}
}
