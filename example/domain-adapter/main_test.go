package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	"github.com/larsartmann/go-sse"
)

func TestBridge_UserPosted(t *testing.T) {
	t.Parallel()

	patches, err := Bridge(UserPosted{User: "alice", Message: "hello", Seq: 7})
	if err != nil {
		t.Fatalf("Bridge: %v", err)
	}

	if len(patches) != 2 {
		t.Fatalf("Bridge produced %d patches; want 2 (element + signals)", len(patches))
	}

	elements := patches[0].Event().Data
	if !strings.Contains(elements, `<strong>alice</strong> hello`) {
		t.Errorf("element patch should render the post; got:\n%s", elements)
	}

	if !strings.Contains(elements, "selector #feed") {
		t.Errorf("element patch should target #feed; got:\n%s", elements)
	}

	signals := patches[1].Event().Data
	if !strings.Contains(signals, `"lastSeq":7`) {
		t.Errorf("signals patch should carry lastSeq; got:\n%s", signals)
	}
}

func TestBridge_WarningRaised(t *testing.T) {
	t.Parallel()

	patches, err := Bridge(WarningRaised{Text: "disk 90%"})
	if err != nil {
		t.Fatalf("Bridge: %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("Bridge produced %d patches; want 1", len(patches))
	}

	if got := patches[0].Event().Data; !strings.Contains(got, `class="toast"`) {
		t.Errorf("warning should render a toast; got:\n%s", got)
	}
}

func TestBridge_UnknownEvent(t *testing.T) {
	t.Parallel()

	if _, err := Bridge(unknownEvent{}); err == nil {
		t.Fatal("expected error for unknown domain event")
	}
}

type unknownEvent struct{}

func (unknownEvent) EventName() string { return "unknown" }

// TestPostHandler_E2E exercises the transport through a real HTTP server:
// POST a domain event, then confirm the SSE feed renders it.
func TestPostHandler_E2E(t *testing.T) {
	t.Parallel()

	patches, err := Bridge(UserPosted{User: "bob", Message: "from test", Seq: 1})
	if err != nil {
		t.Fatalf("Bridge: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		resp := datastar.NewResponse(stream)

		for _, p := range patches {
			if err := resp.Send(p.Event()); err != nil {
				t.Errorf("send: %v", err)
			}
		}
	})

	events := datastartest.Collect(t, handler)

	datastartest.RequireEventCount(t, events, 2)
	datastartest.RequireElements(t, events[0], "#feed", "append",
		`<div class="post" data-post-seq="1"><strong>bob</strong> from test</div>`)
}
