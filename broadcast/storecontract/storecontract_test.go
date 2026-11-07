package storecontract

import (
	"strconv"
	"testing"

	datastar "github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/broadcast"
	"github.com/larsartmann/go-sse"
)

// fakeTB captures helper failures without aborting the running test,
// so the negative tests can assert a contract violation is DETECTED.
type fakeTB struct {
	testing.TB

	failures []string
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatal(args ...any) {
	f.failures = append(f.failures, "fatal")
}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.failures = append(f.failures, format)
}

func (f *fakeTB) Error(args ...any) {
	f.failures = append(f.failures, "error")
}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.failures = append(f.failures, format)
}

func ascendingEvents(count int) []sse.Event {
	events := make([]sse.Event, 0, count)

	for i := 1; i <= count; i++ {
		id := strconv.Itoa(i)
		events = append(events, sse.Event{ID: sse.NewEventID(id), Event: "feed", Data: "payload-" + id})
	}

	return events
}

func TestAssertReplayOrderingMemoryStorePasses(t *testing.T) {
	t.Parallel()

	AssertReplayOrdering(t, datastar.NewMemoryStore(16), ascendingEvents(5))
}

func TestAssertIDAssignmentCatchesUnassignedIDs(t *testing.T) {
	t.Parallel()

	ftb := &fakeTB{TB: t}
	AssertIDAssignment(ftb, datastar.NewMemoryStore(16), 3)

	if len(ftb.failures) == 0 {
		t.Fatal("MemoryStore stores ID-less events verbatim, so the ID-assignment assertion must fail — it did not")
	}
}

func TestAssertIDAssignmentPassesAssigningStore(t *testing.T) {
	t.Parallel()

	AssertIDAssignment(t, &assigningStore{next: 1}, 4)
}

func TestAssertIDAssignmentCatchesDuplicateIDs(t *testing.T) {
	t.Parallel()

	ftb := &fakeTB{TB: t}
	AssertIDAssignment(ftb, &staticIDStore{}, 3)

	if len(ftb.failures) == 0 {
		t.Fatal("a store assigning the same ID to every event must fail the monotonic check — it did not")
	}
}

func TestAssertReplayOrderingCatchesDescendingStore(t *testing.T) {
	t.Parallel()

	ftb := &fakeTB{TB: t}
	AssertReplayOrdering(ftb, &descendingStore{events: ascendingEvents(4)}, ascendingEvents(4))

	if len(ftb.failures) == 0 {
		t.Fatal("a store replaying in descending order must fail the ordering assertion — it did not")
	}
}

func TestAssertReplayOrderingCatchesDroppingStore(t *testing.T) {
	t.Parallel()

	ftb := &fakeTB{TB: t}
	AssertReplayOrdering(ftb, &droppingStore{}, ascendingEvents(4))

	if len(ftb.failures) == 0 {
		t.Fatal("a store dropping appended events must fail the completeness assertion — it did not")
	}
}

// assigningStore is a minimal compliant store: it assigns strictly
// increasing numeric IDs to the ID-less events it receives.
type assigningStore struct {
	events []sse.Event
	next   int
}

func (s *assigningStore) Append(evt sse.Event) {
	evt.ID = sse.NewEventID(strconv.Itoa(s.next))
	s.next++
	s.events = append(s.events, evt)
}

func (s *assigningStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	last, _ := strconv.Atoi(lastID.Get())

	var result []sse.Event

	for _, evt := range s.events {
		if id, _ := strconv.Atoi(evt.ID.Get()); id > last {
			result = append(result, evt)
		}
	}

	return result, nil
}

// staticIDStore assigns the same ID to every event — the monotonic
// violation class.
type staticIDStore struct {
	events []sse.Event
}

func (s *staticIDStore) Append(evt sse.Event) {
	evt.ID = sse.NewEventID("7")
	s.events = append(s.events, evt)
}

func (s *staticIDStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	last, _ := strconv.Atoi(lastID.Get())

	var result []sse.Event

	for _, evt := range s.events {
		if id, _ := strconv.Atoi(evt.ID.Get()); id > last {
			result = append(result, evt)
		}
	}

	return result, nil
}

// descendingStore replays in reverse order — the ordering violation class.
type descendingStore struct {
	events []sse.Event
}

func (s *descendingStore) Append(evt sse.Event) {
	s.events = append(s.events, evt)
}

func (s *descendingStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	last, _ := strconv.Atoi(lastID.Get())

	var result []sse.Event

	for i := len(s.events) - 1; i >= 0; i-- {
		if id, _ := strconv.Atoi(s.events[i].ID.Get()); id > last {
			result = append(result, s.events[i])
		}
	}

	return result, nil
}

// droppingStore loses every second appended event — the completeness
// violation class.
type droppingStore struct {
	events []sse.Event
}

func (s *droppingStore) Append(evt sse.Event) {
	if len(s.events)%2 == 0 {
		s.events = append(s.events, evt)
	}
}

func (s *droppingStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	last, _ := strconv.Atoi(lastID.Get())

	var result []sse.Event

	for _, evt := range s.events {
		if id, _ := strconv.Atoi(evt.ID.Get()); id > last {
			result = append(result, evt)
		}
	}

	return result, nil
}
