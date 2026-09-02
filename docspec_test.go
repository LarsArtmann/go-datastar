//go:build docspec

// Compile-checked documentation snippets. Every function here mirrors a
// snippet from a guide under docs/ (or example/README.md) — if the API the
// docs describe drifts from the real API, this file stops compiling and the
// drift is caught before consumers read stale docs.
//
// Run with: go test -tags docspec ./...   (or: nix run .#docspec)
//
// Contract: when you change a public API that a guide snippet uses, update
// the snippet AND the mirrored function here in the same commit. The
// mirrored functions execute the cheap paths so semantic drift (not just
// type drift) is caught.

package datastar_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/go-datastar"
)

// docs/replay.md — "Minimal replay setup": render, store, broadcast.
func docspecReplayPublish(
	store *datastar.MemoryStore,
	broadcaster *sse.Broadcaster[sse.Event],
	p datastar.Patch,
) {
	evt := p.Event()
	store.Append(evt)
	broadcaster.BroadcastMany(evt)
}

// docs/replay.md — "The reconnection path": backlog then live merge.
func docspecReconnectHandler(
	store *datastar.MemoryStore,
	broadcaster *sse.Broadcaster[sse.Event],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// docs/testing.md lives in the datastartest module; its snippets are
// compile-checked by datastartest/docspec_test.go (same build tag).

// docs/error-system.md — the three matching dimensions.
func docspecErrorMatching(err error, stream *sse.Stream) {
	// 1. By code (stable string, no message matching):
	if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed {
		return
	}

	// 2. By sentinel (errors.Is matches by code+family):
	if errors.Is(err, datastar.ErrEventNameRequired) {
		return
	}

	// 3. By family (behavioral — retryable? whose fault?):
	if errorfamily.IsRetryable(err) {
		_ = stream
	}
}

// example/README.md — heartbeat keep-alive.
func docspecHeartbeat(w http.ResponseWriter, r *http.Request) {
	stream := sse.NewStream(w, r)

	go stream.Heartbeat(r.Context(), 15*time.Second) //nolint:mnd // documented typical interval
}

// docs/error-system.md — the three matching dimensions.
// TestDocspec_GuideSnippets executes the cheap snippet paths so semantic
// drift (a renamed dataline key, a changed error payload shape) fails loudly.
func TestDocspec_GuideSnippets(t *testing.T) {
	t.Parallel()

	t.Run("replay publish + backlog", func(t *testing.T) {
		t.Parallel()

		store := datastar.NewMemoryStore(16)
		broadcaster := sse.NewBroadcaster[sse.Event]()
		defer broadcaster.Close()

		docspecReplayPublish(store, broadcaster,
			datastar.NewElementsPatch("<div>one</div>", datastar.WithElementsEventID("43")))

		if store.Len() != 1 {
			t.Fatalf("store should hold the published event; got %d", store.Len())
		}

		replay, err := store.EventsAfter(sse.NewEventID("42"))
		if err != nil {
			t.Fatalf("EventsAfter: %v", err)
		}

		if len(replay) != 1 || replay[0].ID.Get() != "43" {
			t.Fatalf("replay after 42 should return event 43; got %d events", len(replay))
		}
	})

	t.Run("error payload shape", func(t *testing.T) {
		t.Parallel()

		stream, buf := newTestStream()
		err := errorfamily.NewRejection("test.docspec", "docspec failure")
		if sendErr := datastar.ErrorResponseFromError(stream, err); sendErr != nil {
			t.Fatalf("ErrorResponseFromError: %v", sendErr)
		}

		if !strings.Contains(buf.String(), `"family":"rejection"`) {
			t.Errorf("error payload shape drifted; got:\n%s", buf.String())
		}
	})

	t.Run("handler context", func(t *testing.T) {
		t.Parallel()

		var probe http.ResponseWriter = httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

		docspecHeartbeat(probe, req)
	})
}
