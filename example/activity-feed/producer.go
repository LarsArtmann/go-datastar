package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	datastar "github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

const (
	categoryAlert   = "alert"
	categorySuccess = "success"
	categoryInfo    = "info"
)

const (
	badgeAlert   = "ALERT"
	badgeSuccess = "OK"
	badgeInfo    = "INFO"
)

// Named magic-number constants for the random message generators.
// These make the template strings below self-documenting.
const (
	nodeCount       = 20
	serviceCount    = 50
	endpointCount   = 10
	deployCount     = 999
	sessionCount    = 9999
	taskCount       = 500
	versionMinor    = 9
	versionPatch    = 9
	maxBuildSeconds = 30
)

// feedItemCategoryClass is the CSS-class marker embedding an item's category
// in its elements HTML. The alerts-only filter predicate matches on this
// marker: the Patch API emits canonical DataStar datalines only, so there is
// no custom metadata line to grep for.
const feedItemCategoryClass = "feed-item--"

// activityItem is a single entry in the activity feed.
type activityItem struct {
	category string // "info", "alert", "success"
	badge    string // "INFO", "ALERT", "OK"
	message  string
	time     string
}

// randomN wraps math/rand/v2.IntN for demo data generation. Not crypto-safe;
// this is a demo server generating fake activity messages.
//
//nolint:gosec // G404: non-crypto random for demo data
func randomN(n int) int {
	return rand.IntN(n)
}

// msgTemplate pairs a display category with a random message generator.
type msgTemplate struct {
	category string
	badge    string
	gen      func() string
}

//nolint:gochecknoglobals // read-only template pool, intentional for example
var msgTemplates = []msgTemplate{
	{
		categoryAlert,
		badgeAlert,
		func() string { return fmt.Sprintf("CPU usage above 90%% on node-%d", randomN(nodeCount)+1) },
	},
	{
		categoryAlert,
		badgeAlert,
		func() string { return fmt.Sprintf("Disk space below 10%% on node-%d", randomN(nodeCount)+1) },
	},
	{
		categoryAlert,
		badgeAlert,
		func() string { return fmt.Sprintf("Memory leak detected in service-%d", randomN(serviceCount)+1) },
	},
	{
		categoryAlert,
		badgeAlert,
		func() string {
			return fmt.Sprintf(
				"Response time exceeding SLA on endpoint-%d",
				randomN(endpointCount)+1,
			)
		},
	},
	{
		categorySuccess,
		badgeSuccess,
		func() string { return fmt.Sprintf("Deploy #%d passed all checks", randomN(deployCount)+1) },
	},
	{
		categorySuccess,
		badgeSuccess,
		func() string {
			return fmt.Sprintf(
				"Migration v1.%d.%d applied successfully",
				randomN(versionMinor),
				randomN(versionPatch),
			)
		},
	},
	{
		categorySuccess,
		badgeSuccess,
		func() string {
			return fmt.Sprintf(
				"Build #%d completed in %ds",
				randomN(deployCount)+1,
				randomN(maxBuildSeconds)+1,
			)
		},
	},
	{
		categorySuccess,
		badgeSuccess,
		func() string { return fmt.Sprintf("Health check passed for service-%d", randomN(serviceCount)+1) },
	},
	{
		categoryInfo, badgeInfo,
		func() string { return fmt.Sprintf("User session-%d started", randomN(sessionCount)+1) },
	},
	{
		categoryInfo,
		badgeInfo,
		func() string { return fmt.Sprintf("Cache invalidated for region-%d", randomN(endpointCount)+1) },
	},
	{
		categoryInfo, badgeInfo,
		func() string { return fmt.Sprintf("Scheduled task-%d completed", randomN(taskCount)+1) },
	},
	{
		categoryInfo,
		badgeInfo,
		func() string { return fmt.Sprintf("Configuration reloaded for service-%d", randomN(serviceCount)+1) },
	},
}

func generateItem() activityItem {
	t := msgTemplates[randomN(len(msgTemplates))]

	return activityItem{
		category: t.category,
		badge:    t.badge,
		message:  t.gen(),
		time:     time.Now().Format("15:04:05"),
	}
}

// feedItemEvent builds a DataStar patch-elements event that prepends a single
// feed item to the #feed div. The patch's event ID is the replay identity —
// the browser echoes it back as Last-Event-ID, and the memStore replays from
// it on reconnection.
func feedItemEvent(id sse.EventID, item activityItem) sse.Event {
	patch := datastar.NewElementsPatch(
		feedItemHTML(item),
		datastar.WithSelectorID("feed"),
		datastar.WithMode(datastar.ElementPatchModePrepend),
		datastar.WithElementsEventID(id.Get()),
	)

	return patch.Event()
}

// countEvent builds a DataStar patch-signals event updating the
// subscriberCount signal. No event ID — an ephemeral meta event that must
// not be replayed.
func countEvent(count int) (sse.Event, error) {
	patch, err := datastar.NewSignalsPatch(map[string]any{"subscriberCount": count})
	if err != nil {
		return sse.Event{}, err // already classified by go-datastar (Rejection)
	}

	return patch.Event(), nil
}

// replayEvent builds a DataStar patch-signals event setting the replayed
// signal, which drives the "Replayed N missed events" banner.
func replayEvent(n int) (sse.Event, error) {
	patch, err := datastar.NewSignalsPatch(map[string]any{"replayed": n})
	if err != nil {
		return sse.Event{}, err // already classified by go-datastar (Rejection)
	}

	return patch.Event(), nil
}

// totalEventSignal builds a DataStar patch-signals event updating the
// totalEvents counter (UI stat and empty-state check).
func totalEventSignal(n int64) (sse.Event, error) {
	patch, err := datastar.NewSignalsPatch(map[string]any{"totalEvents": n})
	if err != nil {
		return sse.Event{}, err // already classified by go-datastar (Rejection)
	}

	return patch.Event(), nil
}

func feedItemHTML(item activityItem) string {
	return fmt.Sprintf(
		`<div class="feed-item feed-item--%s"><span class="feed-item__badge">%s</span><span class="feed-item__message">%s</span><span class="feed-item__time">%s</span></div>`,
		item.category,
		item.badge,
		item.message,
		item.time,
	)
}

// startProducer runs a background goroutine that emits a random activity
// event every emitInterval. total counts emitted events; its decimal form is
// each feed event's sse.EventID so reconnecting clients can replay missed
// events.
func (s *activityServer) startProducer(ctx context.Context) {
	var total int64

	emit := func() {
		total++

		evt := feedItemEvent(sse.NewEventID(strconv.FormatInt(total, 10)), generateItem())

		sig, sigErr := totalEventSignal(total)
		if sigErr != nil {
			log.Printf("producer: %v", sigErr)

			s.store.Append(evt)
			s.broadcaster.Broadcast(evt)

			return
		}

		s.store.Append(evt)
		s.broadcaster.BroadcastMany(evt, sig)
	}

	emit() // emit one immediately so the user sees something fast

	ticker := time.NewTicker(emitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			emit()
		}
	}
}
