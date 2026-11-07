// Package storecontract provides reusable assertions for the replay
// contract [broadcast.Store] documents, so implementations of custom
// stores (Redis, Postgres, ...) can prove they honor it inside their own
// tests. Every helper accepts [testing.TB], matching the datastartest
// convention: they work with *testing.T, *testing.B, and Ginkgo's
// GinkgoT().
//
// The contract being asserted (see [broadcast.Store]): replay is keyed on
// numeric event IDs, and a store receiving ID-less events must assign its
// own monotonic IDs — the broadcaster appends verbatim and never assigns
// IDs, so a store that skips the assignment makes Last-Event-ID replay
// degenerate.
package storecontract

import (
	"strconv"
	"testing"

	"github.com/larsartmann/go-datastar/broadcast"
	"github.com/larsartmann/go-sse"
)

// AssertReplayOrdering appends events to store and asserts the replay
// slice of the [broadcast.Store] contract: EventsAfter(zero) returns
// every appended event verbatim, ordered by ascending numeric ID;// EventsAfter(id) returns exactly the events after that ID; no call
// errors. This is the behavior Last-Event-ID reconnection replay depends
// on. The events must carry strictly increasing numeric string IDs (the
// same convention [datastar.MemoryStore] and the broadcaster tests use).
func AssertReplayOrdering(tb testing.TB, store broadcast.Store, events []sse.Event) {
	tb.Helper()

	previous := 0
	for i, evt := range events {
		seq := requireSeq(tb, i, evt.ID)
		if seq <= previous {
			tb.Fatalf("event %d carries ID %d, not greater than event %d's %d: the assertion expects strictly increasing numeric IDs as input", i, seq, i-1, previous)
		}

		previous = seq
		store.Append(evt)
	}

	all, err := store.EventsAfter(sse.EventID{})
	if err != nil {
		tb.Fatalf("EventsAfter(zero) errored: %v", err)
	}

	if len(all) != len(events) {
		tb.Fatalf("EventsAfter(zero) returned %d event(s); appended %d — a compliant store replays every appended event", len(all), len(events))
	}

	for i, evt := range all {
		if evt != events[i] {
			tb.Fatalf("EventsAfter(zero)[%d] = %+v; want the appended %+v verbatim", i, evt, events[i])
		}
	}

	for i := len(events) - 1; i >= 0; i-- {
		after, err := store.EventsAfter(events[i].ID)
		if err != nil {
			tb.Fatalf("EventsAfter(%q) errored: %v", events[i].ID.Get(), err)
		}

		if want := len(events) - i - 1; len(after) != want {
			tb.Fatalf("EventsAfter(%q) returned %d event(s); want %d — exactly the suffix after that Last-Event-ID", events[i].ID.Get(), len(after), want)

			continue
		}

		if len(after) > 0 && after[0] != events[i+1] {
			tb.Fatalf("EventsAfter(%q)[0] = %+v; want %+v", events[i].ID.Get(), after[0], events[i+1])
		}
	}
}

// AssertIDAssignment appends count ID-less events to store and asserts the
// assignment slice of the [broadcast.Store] contract: every stored event
// comes back from EventsAfter carrying a non-empty numeric ID, strictly
// increasing across the sequence. A store that skips the assignment makes
// Last-Event-ID replay degenerate (ID-less events are unreachable for
// every real lastID). [datastar.MemoryStore] deliberately does NOT pass
// this assertion — its contract is caller-assigned numeric IDs, which is
// what [broadcast.WithReplayCapacity] consumers do today.
func AssertIDAssignment(tb testing.TB, store broadcast.Store, count int) {
	tb.Helper()

	for i := 0; i < count; i++ {
		store.Append(sse.Event{Event: "tick", Data: strconv.Itoa(i + 1)})
	}

	all, err := store.EventsAfter(sse.EventID{})
	if err != nil {
		tb.Fatalf("EventsAfter(zero) errored: %v", err)
	}

	if len(all) != count {
		tb.Fatalf("EventsAfter(zero) returned %d of %d appended events — ID-less events stored without an assigned ID are unreachable for every real Last-Event-ID", len(all), count)
	}

	previous := 0
	for i, evt := range all {
		seq := requireSeq(tb, i, evt.ID)
		if seq <= previous {
			tb.Fatalf("event %d carries ID %d, not greater than the previous %d: replay requires strictly increasing IDs", i, seq, previous)
		}

		previous = seq
	}
}

// requireSeq parses evtID as a non-empty numeric ID, failing tb otherwise.
func requireSeq(tb testing.TB, index int, evtID sse.EventID) int {
	tb.Helper()

	text := evtID.Get()
	seq, err := strconv.Atoi(text)
	if err != nil || text == "" {
		tb.Fatalf("event %d carries ID %q: replay is keyed on numeric IDs, so the contract assertions require non-empty numeric string IDs", index, text)

		return 0
	}

	return seq
}
