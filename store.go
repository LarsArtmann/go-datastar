package datastar

import (
	"strconv"
	"sync"

	"github.com/larsartmann/go-sse"
)

// DefaultMemoryStoreCapacity is the default number of events retained for
// reconnection replay when no capacity is specified.
const DefaultMemoryStoreCapacity = 128

// MemoryStore is an in-memory ring buffer implementing [sse.EventStore].
// It keeps the last N events for SSE reconnection replay.
//
// Events are sequenced by their numeric [sse.EventID]: EventsAfter returns
// all stored events whose ID is strictly greater than the requested lastID.
// Non-numeric IDs are treated as zero, so an empty or non-numeric lastID
// replays the entire buffer.
//
// MemoryStore is safe for concurrent use. It is intended for single-process
// deployments and demos. For multi-instance setups, use a shared store
// (Redis, Postgres) that implements [sse.EventStore].
type MemoryStore struct {
	mu     sync.RWMutex
	events []sse.Event
	cap    int
}

// NewMemoryStore creates a [MemoryStore] that retains the last capacity events.
// If capacity is non-positive, [DefaultMemoryStoreCapacity] is used.
func NewMemoryStore(capacity int) *MemoryStore {
	if capacity <= 0 {
		capacity = DefaultMemoryStoreCapacity
	}

	return &MemoryStore{
		events: make([]sse.Event, 0, capacity),
		cap:    capacity,
	}
}

// Append stores an event for later replay. If the buffer is full, the oldest
// event is evicted.
func (s *MemoryStore) Append(evt sse.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, evt)

	if len(s.events) > s.cap {
		s.events = s.events[len(s.events)-s.cap:]
	}
}

// EventsAfter returns stored events with IDs strictly greater than lastID,
// ordered ascending by sequence number. Non-numeric lastID values are treated
// as zero, replaying the entire buffer.
func (s *MemoryStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	lastSeq := parseSeq(lastID.Get())

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []sse.Event

	for _, evt := range s.events {
		if parseSeq(evt.ID.Get()) > lastSeq {
			result = append(result, evt)
		}
	}

	return result, nil
}

// Len returns the number of events currently stored.
func (s *MemoryStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.events)
}

// parseSeq parses a numeric event ID string into an int. Non-numeric values
// return 0, so they sort before all numeric IDs.
func parseSeq(id string) int {
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0
	}

	return n
}
