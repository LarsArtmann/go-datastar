// Testable examples for the go-sse routing primitives as used with
// go-datastar patches: Broadcaster fan-out, typed event-type filtering via
// SubscribeFilter, and reconnection replay through MemoryStore.

package datastar_test

import (
	"fmt"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// ExampleNewBroadcaster demonstrates the core patches-as-values payoff:
// construct a patch without a connection, then fan it out to every
// subscriber.
func ExampleNewBroadcaster() {
	broadcaster := sse.NewBroadcaster[sse.Event]()
	defer broadcaster.Close()

	ch := broadcaster.Subscribe()
	defer broadcaster.Unsubscribe(ch)

	patch := datastar.NewElementsPatch("<div>hello everyone</div>",
		datastar.WithSelectorID("feed"),
		datastar.WithModePrepend(),
	)

	broadcaster.Broadcast(patch.Event())

	evt := <-ch
	fmt.Println("fanned out:", evt.Event)
	fmt.Println("has selector line:", containsLine(evt.Data, "selector #feed"))

	// Output:
	// fanned out: datastar-patch-elements
	// has selector line: true
}

// ExampleBroadcaster_SubscribeFilter demonstrates per-subscriber filtering:
// only signals patches reach this subscriber, elements patches are dropped
// before entering its buffer.
func ExampleBroadcaster_SubscribeFilter() {
	broadcaster := sse.NewBroadcaster[sse.Event]()
	defer broadcaster.Close()

	signalsOnly := broadcaster.SubscribeFilter(func(evt sse.Event) bool {
		return evt.Event == "datastar-patch-signals"
	})
	defer broadcaster.Unsubscribe(signalsOnly)

	broadcaster.Broadcast(datastar.NewElementsPatch("<div>ignored</div>").Event())
	broadcaster.Broadcast(mustSignalsEvent(tExample, `{"count":1}`))
	broadcaster.Broadcast(datastar.NewElementsPatch("<div>ignored too</div>").Event())

	evt := <-signalsOnly
	fmt.Println("received:", evt.Event)

	// Output:
	// received: datastar-patch-signals
}

// ExampleMemoryStore demonstrates replay for reconnecting clients: the store
// keeps the last N events, and EventsAfter returns everything after a given
// Last-Event-ID.
func ExampleMemoryStore() {
	store := datastar.NewMemoryStore(16)

	first := datastar.NewElementsPatch("<div>one</div>").Event()
	second := datastar.NewSignalsPatchIfMissing(map[string]any{"count": 2}).Event()

	first.ID = sse.NewEventID("1").String()
	second.ID = sse.NewEventID("2").String()

	store.Append(first)
	store.Append(second)

	fmt.Println("stored:", store.Len())

	for _, evt := range store.EventsAfter(sse.NewEventID("1")) {
		fmt.Println("replay after 1:", evt.Event)
	}

	// Output:
	// stored: 2
	// replay after 1: datastar-patch-signals
}
