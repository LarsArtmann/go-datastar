package datastartest

// Search helpers locate events of interest inside a collected stream. They
// pair with the Require* assertions: find first (optionally failing if
// absent), then assert on the result.

// FindAllElements returns every patch-elements event in events whose selector
// equals selector, in stream order. Events of other types (signals, comments,
// keep-alives) are ignored. An empty selector matches patches without an
// explicit selector dataline (the client merges those into the target
// element). Returns nil when nothing matches.
//
// Use [RequireElementsOrdered] when count and order matter; use
// [EventToSelectorMap] for O(1) lookup when selectors are unique.
func FindAllElements(events []Event, selector string) []Event {
	var matches []Event

	for _, evt := range events {
		if evt.IsElements() && evt.Selector() == selector {
			matches = append(matches, evt)
		}
	}

	return matches
}

// FindScript returns the first script-bearing patch-elements event in events
// (e.g., from Redirect, ExecuteScript, or ConsoleLog). The bool reports
// whether one was found; when false, the returned Event is the zero value.
// Use [RequireScript] to also assert the exact JavaScript content.
func FindScript(events []Event) (Event, bool) {
	for _, evt := range events {
		if evt.IsScript() {
			return evt, true
		}
	}

	return Event{}, false
}

// EventToSelectorMap indexes the patch-elements events in events by their
// selector for O(1) lookup. Events of other types are ignored. If several
// events share a selector, the last one wins: later patches overwrite earlier
// DOM state, so the last event is what the client's DOM ends up with.
// Script-bearing patches are elements patches too and participate in the map.
// Patches without an explicit selector are keyed under "".
//
// Use [FindAllElements] when a selector can legitimately be patched more than
// once and you need every occurrence.
func EventToSelectorMap(events []Event) map[string]Event {
	bySelector := make(map[string]Event)

	for _, evt := range events {
		if evt.IsElements() {
			bySelector[evt.Selector()] = evt
		}
	}

	return bySelector
}
