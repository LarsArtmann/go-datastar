package datastartest

import (
	"slices"

	"github.com/larsartmann/go-datastar"
)

// FilterElements returns only the patch-elements events from the slice.
func FilterElements(events []Event) []Event {
	return slices.DeleteFunc(slices.Clone(events), func(e Event) bool {
		return e.Type != string(datastar.EventTypePatchElements)
	})
}

// FilterSignals returns only the patch-signals events from the slice.
func FilterSignals(events []Event) []Event {
	return slices.DeleteFunc(slices.Clone(events), func(e Event) bool {
		return e.Type != string(datastar.EventTypePatchSignals)
	})
}
