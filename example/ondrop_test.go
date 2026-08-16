package main

import (
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// TestWithOnDropFiresWhenSubscriberBufferFull proves the drop-observability
// wiring the example demonstrates: a subscriber that stops reading has its
// per-subscriber buffer fill, and every event broadcast beyond that capacity
// surfaces through sse.WithOnDrop instead of blocking the producer.
//
// This is the pattern to copy when a real handler needs to know that a slow
// client is losing events (see main's broadcaster construction).
func TestWithOnDropFiresWhenSubscriberBufferFull(t *testing.T) {
	const bufferSize = 2

	dropped := make(chan sse.Event, 16)

	broadcaster := sse.NewBroadcaster[sse.Event](
		sse.WithBufferSize[sse.Event](bufferSize),
		sse.WithOnDrop(func(evt sse.Event) { dropped <- evt }),
	)
	defer broadcaster.Close()

	// The slow subscriber: subscribed, but never reads.
	subscriber := broadcaster.Subscribe()
	defer broadcaster.Unsubscribe(subscriber)

	sent := make([]sse.Event, 0, bufferSize+3)
	for i := 0; i < bufferSize+3; i++ {
		patch := datastar.NewElementsPatch(
			"<div>Item #"+string(rune('0'+i))+"</div>",
			datastar.WithSelectorID("feed"),
			datastar.WithMode(datastar.ElementPatchModePrepend),
		)
		evt := patch.Event()
		sent = append(sent, evt)
		broadcaster.Broadcast(evt)
	}

	// The first `bufferSize` events fit the subscriber's buffer; the rest
	// must arrive through the drop callback, in broadcast order.
	for i := bufferSize; i < len(sent); i++ {
		select {
		case got := <-dropped:
			if got.Data != sent[i].Data || got.Event != sent[i].Event {
				t.Errorf("dropped event %d: got (%q, %q), want (%q, %q)",
					i, got.Event, got.Data, sent[i].Event, sent[i].Data)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("expected event %d to be dropped", i)
		}
	}

	// Nothing beyond the expected drops was reported.
	select {
	case got := <-dropped:
		t.Errorf("unexpected extra drop: %q", got.Data)
	default:
	}

	// The subscriber's buffer still holds exactly the first events, in order.
	if len(subscriber) != bufferSize {
		t.Fatalf("subscriber buffer: got %d events, want %d", len(subscriber), bufferSize)
	}

	for i := 0; i < bufferSize; i++ {
		got := <-subscriber
		if got.Data != sent[i].Data {
			t.Errorf("buffered event %d: got %q, want %q", i, got.Data, sent[i].Data)
		}
	}
}
