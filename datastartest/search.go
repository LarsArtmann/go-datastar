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
