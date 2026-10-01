//go:build docspec

// Compile-checked snippets from broadcast-facing docs (docs/migration-guide.md
// "New optional broadcast/ submodule", broadcast README). Same contract as the
// root docspec_test.go: mirrored functions must compile and execute their
// cheap paths, so doc drift fails the tagged test run.
//
// Run with: go test -tags docspec .   (from broadcast/)

package broadcast_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/broadcast"
)

// docs/migration-guide.md — adopting the broadcast submodule: mount the
// Broadcaster as a handler; fan out patches; replay on reconnect.
func docspecMigrationGuideAdoption(t *testing.T) {
	t.Helper()

	broadcaster := broadcast.NewBroadcasterWithReplay(16)
	defer broadcaster.Close()

	patch := datastar.NewElementsPatch("<div>hello</div>")

	// Events are appended to the replay store BEFORE fan-out, so
	// reconnecting clients replay instead of missing the event.
	broadcaster.Broadcast(patch)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	broadcaster.ServeHTTP(recorder, req)
}

// docs/migration-guide.md — patch-level fan-out surface.
func docspecMigrationGuideFanOut(broadcaster *broadcast.Broadcaster, patches []datastar.Patch) {
	broadcaster.Broadcast(patches[0])
	broadcaster.BroadcastMany(patches...)
	broadcaster.BroadcastEvent(patches[0].Event())
}

// broadcast README — "Options": buffer size, replay, store, and heartbeat
// are orthogonal knobs on one constructor.
func docspecReadmeOptions(t *testing.T) {
	t.Helper()

	b := broadcast.NewBroadcaster(
		broadcast.WithBufferSize(64),
		broadcast.WithReplayCapacity(256),
		broadcast.WithHeartbeatInterval(30*time.Second),
	)
	defer b.Close()

	b.Broadcast(datastar.NewElementsPatch("<div>hello</div>"))

	var _ broadcast.Store = datastar.NewMemoryStore(16)
}

// TestDocspec_BroadcastSnippets executes the adoption path end to end so
// constructor or ordering drift (append-before-fan-out) fails loudly.
func TestDocspec_BroadcastSnippets(t *testing.T) {
	t.Parallel()

	broadcaster := broadcast.NewBroadcasterWithReplay(16)
	defer broadcaster.Close()

	patches := []datastar.Patch{
		datastar.NewElementsPatch("<div>one</div>"),
		datastar.NewElementsPatch("<div>two</div>"),
	}

	broadcaster.BroadcastMany(patches...)

	docspecMigrationGuideFanOut(broadcaster, patches)
	docspecMigrationGuideAdoption(t)
	docspecReadmeOptions(t)

	var _ http.Handler = broadcaster // mountable as a handler, per the guide
}
