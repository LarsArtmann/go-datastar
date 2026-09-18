package datastartest

// FindElement returns the first patch-elements event whose CSS selector matches
// the given value, along with true. Returns false if no match is found.
//
// Useful when a handler sends multiple elements patches and you need to assert
// on a specific one without indexing by position.
func FindElement(events []Event, selector string) (Event, bool) {
	for _, evt := range events {
		if evt.IsElements() && evt.Selector() == selector {
			return evt, true
		}
	}

	return Event{}, false
}

// FindSignals returns the first patch-signals event, along with true. Returns
// false if the slice contains no signals events.
func FindSignals(events []Event) (Event, bool) {
	for _, evt := range events {
		if evt.IsSignals() {
			return evt, true
		}
	}

	return Event{}, false
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

// FindAllElements returns every patch-elements event in events whose selector
// equals selector, in stream order — the plural counterpart of [FindElement].
// Events of other types (signals, comments, keep-alives) are ignored. An empty
// selector matches patches without an explicit selector dataline (the client
// merges those into the target element). Returns nil when nothing matches.
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
