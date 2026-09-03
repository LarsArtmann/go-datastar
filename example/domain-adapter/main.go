// Command domain-adapter demonstrates the "patches as values" architecture:
// a domain layer speaks its own events, an EventBridge translates them into
// go-datastar Patch values, and the transport layer stores/replays/broadcasts
// those values without knowing anything about the domain.
//
// This is the architectural style cqrs-htmx/datastar builds on — shown here
// in miniature, with no dependencies beyond this module and go-sse.
package main

import (
	"encoding/json/v2"
	"errors"
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

const (
	// replayBufferSize bounds the reconnection replay ring (events).
	replayBufferSize = 256

	// readHeaderTimeout protects against slowloris-style header floods.
	readHeaderTimeout = 5 * time.Second
)

var errUnknownDomainEvent = errors.New("unknown domain event")

// Bridge translates one domain event into the patches that render it. This
// is the ONLY place domain vocabulary meets the DataStar wire protocol.
// Patches are returned as values: the caller decides to send, store,
// replay, or broadcast them.
func Bridge(evt DomainEvent) ([]datastar.Patch, error) {
	switch domainEvt := evt.(type) {
	case UserPosted:
		html := fmt.Sprintf(
			`<div class="post" data-post-seq="%d"><strong>%s</strong> %s</div>`,
			domainEvt.Seq, domainEvt.User, domainEvt.Message,
		)

		signals := map[string]any{"feed": map[string]any{"lastSeq": domainEvt.Seq}}

		signalsJSON, err := datastar.MarshalSignals(signals)
		if err != nil {
			return nil, fmt.Errorf("bridge %s: %w", domainEvt.EventName(), err)
		}

		return []datastar.Patch{
			datastar.NewElementsPatch(
				html,
				datastar.WithSelector("#feed"),
				datastar.WithModeAppend(),
			),
			datastar.SignalsPatch{Signals: signalsJSON},
		}, nil

	case WarningRaised:
		html := fmt.Sprintf(`<div class="toast">%s</div>`, domainEvt.Text)

		return []datastar.Patch{
			datastar.NewElementsPatch(
				html,
				datastar.WithSelector("#toasts"),
				datastar.WithModeAppend(),
			),
		}, nil

	default:
		return nil, fmt.Errorf("bridge: %T: %w", evt, errUnknownDomainEvent)
	}
}

// ---- Transport layer: broadcast + replay, domain-agnostic ---------------

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	broadcaster := sse.NewBroadcaster[sse.Event]()
	defer broadcaster.Close()

	store := datastar.NewMemoryStore(replayBufferSize)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		handleEvents(w, r, broadcaster, store)
	})
	mux.HandleFunc("POST /post", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r, broadcaster, store)
	})

	server := &http.Server{
		Addr:              ":8766",
		ReadHeaderTimeout: readHeaderTimeout,
		Handler:           mux,
	}

	log.Println("domain-adapter listening on :8766")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

// handleEvents subscribes the client, replays the backlog first, then
// forwards the live broadcast.
func handleEvents(
	writer http.ResponseWriter,
	r *http.Request,
	broadcaster *sse.Broadcaster[sse.Event],
	store *datastar.MemoryStore,
) {
	stream := sse.NewStream(writer, r)
	defer stream.Close()

	resp := datastar.NewResponse(stream)

	if backlog, err := store.EventsAfter(stream.LastEventID()); err == nil {
		for _, evt := range backlog {
			if err := resp.Send(evt); err != nil {
				return
			}
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

			if err := resp.Send(evt); err != nil {
				return
			}
		}
	}
}

// handlePost accepts a domain event (JSON), bridges it, publishes it.
func handlePost(
	writer http.ResponseWriter,
	r *http.Request,
	broadcaster *sse.Broadcaster[sse.Event],
	store *datastar.MemoryStore,
) {
	var payload struct {
		Type    string `json:"type"`
		User    string `json:"user"`
		Message string `json:"message"`
	}

	if err := json.UnmarshalRead(r.Body, &payload); err != nil {
		http.Error(writer, "bad json", http.StatusBadRequest)

		return
	}

	var evt DomainEvent

	switch payload.Type {
	case "user_posted":
		evt = UserPosted{User: payload.User, Message: payload.Message, Seq: time.Now().Unix()}
	case "warning_raised":
		evt = WarningRaised{Text: payload.Message}
	default:
		http.Error(writer, "unknown event type", http.StatusBadRequest)

		return
	}

	patches, err := Bridge(evt)
	if err != nil {
		log.Printf("bridge: %v", err)
		http.Error(writer, "bridge failed", http.StatusInternalServerError)

		return
	}

	for _, p := range patches {
		evt := p.Event()
		store.Append(evt)
		broadcaster.BroadcastMany(evt)
	}

	writer.WriteHeader(http.StatusAccepted)
}
