// Command domain-adapter demonstrates the "patches as values" architecture:
// a domain layer speaks its own events, an EventBridge translates them into
// go-datastar Patch values, and the transport layer stores/replays/broadcasts
// those values without knowing anything about the domain.
//
// This is the architectural style cqrs-htmx/datastar builds on — shown here
// in miniature, with no dependencies beyond this module and go-sse.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// ---- Domain layer: events in the domain's own language -----------------

// DomainEvent is something that happened. The domain never mentions HTML,
// selectors, or SSE.
type DomainEvent interface {
	EventName() string
}

// UserPosted is a feed item.
type UserPosted struct {
	User    string
	Message string
	Seq     int64
}

func (e UserPosted) EventName() string { return "user_posted" }

// WarningRaised is a system toast.
type WarningRaised struct {
	Text string
}

func (e WarningRaised) EventName() string { return "warning_raised" }

// ---- The EventBridge: domain events -> Patch values ---------------------

// Bridge translates one domain event into the patches that render it. This
// is the ONLY place domain vocabulary meets the DataStar wire protocol.
// Patches are returned as values: the caller decides to send, store,
// replay, or broadcast them.
func Bridge(evt DomainEvent) ([]datastar.Patch, error) {
	switch e := evt.(type) {
	case UserPosted:
		html := fmt.Sprintf(
			`<div class="post" data-post-seq="%d"><strong>%s</strong> %s</div>`,
			e.Seq, e.User, e.Message,
		)

		signals := map[string]any{"feed": map[string]any{"lastSeq": e.Seq}}

		signalsJSON, err := datastar.MarshalSignals(signals)
		if err != nil {
			return nil, fmt.Errorf("bridge %s: %w", e.EventName(), err)
		}

		return []datastar.Patch{
			datastar.NewElementsPatch(html, datastar.WithSelector("#feed"), datastar.WithModeAppend()),
			datastar.SignalsPatch{Signals: signalsJSON},
		}, nil

	case WarningRaised:
		html := fmt.Sprintf(`<div class="toast">%s</div>`, e.Text)

		return []datastar.Patch{
			datastar.NewElementsPatch(html, datastar.WithSelector("#toasts"), datastar.WithModeAppend()),
		}, nil

	default:
		return nil, fmt.Errorf("bridge: unknown domain event %T", evt)
	}
}

// ---- Transport layer: broadcast + replay, domain-agnostic ---------------

func main() {
	broadcaster := sse.NewBroadcaster[sse.Event]()
	defer broadcaster.Close()

	mux := http.NewServeMux()

	// GET /events — subscribe, with MemoryStore replay for reconnects.
	store := datastar.NewMemoryStore(256)
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		if backlog, err := store.EventsAfter(stream.LastEventID()); err == nil {
			for _, evt := range backlog {
				_ = resp.Send(evt)
			}
		}

		events := broadcaster.Subscribe()
		defer broadcaster.Unsubscribe(events)

		for {
			select {
			case <-r.Context().Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				_ = resp.Send(evt)
			}
		}
	})

	// POST /post — accept a domain event (JSON), bridge it, publish.
	mux.HandleFunc("POST /post", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Type    string `json:"type"`
			User    string `json:"user"`
			Message string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)

			return
		}

		var evt DomainEvent

		switch payload.Type {
		case "user_posted":
			evt = UserPosted{User: payload.User, Message: payload.Message, Seq: time.Now().Unix()}
		case "warning_raised":
			evt = WarningRaised{Text: payload.Message}
		default:
			http.Error(w, "unknown event type", http.StatusBadRequest)

			return
		}

		patches, err := Bridge(evt)
		if err != nil {
			log.Printf("bridge: %v", err)
			http.Error(w, "bridge failed", http.StatusInternalServerError)

			return
		}

		for _, p := range patches {
			evt := p.Event()
			store.Append(evt)
			broadcaster.BroadcastMany(evt)
		}

		w.WriteHeader(http.StatusAccepted)
	})

	log.Println("domain-adapter listening on :8766")
	log.Fatal(http.ListenAndServe(":8766", mux))
}
