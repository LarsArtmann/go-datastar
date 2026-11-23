package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/larsartmann/go-sse"
)

// activityServer holds the broadcaster and event store.
// The broadcaster fans out events to all connected SSE clients; the store
// keeps recent events for reconnection replay.
type activityServer struct {
	broadcaster *sse.Broadcaster[sse.Event]
	store       *memStore
}

func newActivityServer() *activityServer {
	broadcaster := sse.NewBroadcaster[sse.Event]()

	server := &activityServer{
		broadcaster: broadcaster,
		store:       newMemStore(maxStoredEvents),
	}
	// Broadcast the new subscriber count whenever a client connects or
	// disconnects. This gives every tab a live "N clients connected" indicator.
	// We reset $replayed to 0 alongside the count update so the replay banner
	// clears when the subscriber count changes (e.g. on new connection).
	//
	// NOTE: There is a benign TOCTOU race here — SubscriberCount() is read
	// before Broadcast runs, so a concurrent subscribe/unsubscribe could
	// produce a slightly stale count. This is acceptable for a demo; a
	// production system would use an atomic counter or a single-state
	// broadcast event.
	broadcaster.OnSubscribe(func() {
		server.broadcastSignal("replayed", func() (sse.Event, error) {
			return replayEvent(0)
		})
		server.broadcastSignal("subscriberCount", func() (sse.Event, error) {
			return countEvent(broadcaster.SubscriberCount())
		})
	})
	broadcaster.OnUnsubscribe(func() {
		server.broadcastSignal("subscriberCount", func() (sse.Event, error) {
			return countEvent(broadcaster.SubscriberCount())
		})
	})

	return server
}

func (s *activityServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	alertsOnly := r.URL.Query().Get("filter") == "alerts"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := fmt.Fprint(w, indexHTML(alertsOnly)); err != nil {
		log.Printf("render index page: %v", err)
	}
}

func (s *activityServer) eventsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// Allow cross-origin SSE consumption (enables embedding the feed from
	// other domains).
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")

	// NewStream sets the SSE headers (text/event-stream, no-cache, etc.)
	// and writes 200 OK immediately. Do not write to w before this call.
	stream := sse.NewStream(responseWriter, request)
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("close stream: %v", err)
		}
	}()

	ctx := stream.Context()

	if !s.replayMissedEvents(stream) {
		return
	}

	events := s.subscribeForRequest(request)
	defer s.broadcaster.Unsubscribe(events)

	// Heartbeat to keep the connection alive through reverse proxies.
	//
	//nolint:contextcheck // ctx is stream.Context() which is r.Context()
	go stream.Heartbeat(ctx, heartbeatEvery)

	// Event loop: forward broadcast events to the SSE stream.
	// stream.Send serializes on a mutex (http.ResponseWriter is not
	// goroutine-safe), so it's safe to call here while the heartbeat
	// goroutine may also be writing.
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-events:
			if !ok || stream.Send(evt) != nil {
				return
			}
		}
	}
}

// replayMissedEvents replays events from the store based on the Last-Event-ID
// header and signals the count of replayed events to the client. Returns false
// if a write failed and the handler should abort.
func (s *activityServer) replayMissedEvents(stream *sse.Stream) bool {
	lastID := stream.LastEventID()
	if lastID.IsZero() {
		return true
	}

	// sse.Replay reads EventsAfter from the store and sends each event
	// in order via stream.Send.
	replayed, err := sse.Replay(stream, s.store, lastID)
	if err != nil {
		log.Printf("replay failed: %v", err)

		return true
	}

	if replayed == 0 {
		return true
	}

	evt, evtErr := replayEvent(replayed)
	if evtErr != nil {
		log.Printf("replay signal: %v", evtErr)

		return true
	}

	if err := stream.Send(evt); err != nil {
		return false
	}

	return true
}

// subscribeForRequest registers a subscriber for this request, optionally
// filtered to alerts only. Meta events (subscriber count, replay indicators)
// have no event ID and always pass through the filter so every tab sees the
// subscriber count.
func (s *activityServer) subscribeForRequest(r *http.Request) <-chan sse.Event {
	if r.URL.Query().Get("filter") != "alerts" {
		return s.broadcaster.Subscribe()
	}

	return s.broadcaster.SubscribeFilter(func(evt sse.Event) bool {
		if evt.ID.IsZero() {
			return true
		}

		return strings.Contains(evt.Data, feedItemCategoryClass+categoryAlert)
	})
}

// broadcastSignal builds a signals patch via build and fans it out; a failed
// build is logged with its name and skipped rather than broadcast as an
// empty event. The error is already classified by go-datastar (Rejection).
func (s *activityServer) broadcastSignal(name string, build func() (sse.Event, error)) {
	evt, err := build()
	if err != nil {
		log.Printf("signal %s: %v", name, err)

		return
	}

	s.broadcaster.Broadcast(evt)
}
