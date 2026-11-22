package main

import (
	"fmt"

	datastar "github.com/larsartmann/go-datastar"
)

// indexHTML renders the single-page UI. The page is fully client-hydrated:
// DataStar connects to /events (or /events?filter=alerts) via data-init and
// every dynamic value flows in as patch-signals / patch-elements updates.
//
// Converted from the original templ page; the markup, ids, and classes are
// unchanged (styles.css and app.js key on them), only the branding and the
// JS-serving pattern differ: the pinned DataStar runtime is served by
// ScriptHandler and referenced through ScriptTagVersioned (cache-busting
// version query) instead of a vendored copy under /static/.
func indexHTML(alertsOnly bool) string {
	filterBar := `<span class="filter-bar__active">All events</span>
						<a class="filter-bar__link" href="/?filter=alerts">Alerts only</a>`
	initDiv := `<div data-init="@get('/events')"></div>`

	if alertsOnly {
		filterBar = `<a class="filter-bar__link" href="/">All events</a>
						<span class="filter-bar__active">Alerts only</span>`
		initDiv = `<div data-init="@get('/events?filter=alerts')"></div>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1"/>
		<title>go-datastar — Live Activity Feed</title>
		<meta property="og:title" content="go-datastar — Live Activity Feed"/>
		<meta property="og:description" content="DataStar patches over SSE with live fan-out, filtering, and reconnection replay."/>
		<meta property="og:type" content="website"/>
		<link rel="icon" href="data:image/svg+xml,%%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%%3E%%3Ctext y='.9em' font-size='90'%%3E%%E2%%9A%%A1%%3C/text%%3E%%3C/svg%%3E"/>
		<link rel="stylesheet" href="/static/styles.css"/>
		%s
		<script src="/static/app.js" defer></script>
	</head>
	<body>
		<main class="card">
			<div data-signals="{subscriberCount: 0, replayed: 0, totalEvents: 0, paused: false}">
				<header class="card__header">
					<div class="header-row">
						<h1>go-datastar</h1>
						<button class="theme-toggle">
							<span class="theme-toggle__dark">Dark</span>
							<span class="theme-toggle__light">Light</span>
						</button>
					</div>
					<p class="subtitle">
						Live activity feed powered by DataStar patches over SSE. Open multiple tabs
						to see real-time fan-out — every tab receives the same events simultaneously.
					</p>
				</header>
				<nav class="filter-bar">
					%s
					<span class="filter-bar__hint">press 'a' for alerts</span>
				</nav>
				<div class="stats-bar">
					<div class="stat">
						<span class="stat__value" data-text="$subscriberCount"></span>
						<span class="stat__label">clients connected</span>
					</div>
					<div class="stat">
						<span class="stat__value" data-text="$totalEvents"></span>
						<span class="stat__label">events sent</span>
					</div>
					<div class="stat stat--live">
						<span class="live-dot"></span>
						<span class="stat__label">live</span>
					</div>
				</div>
				<div class="replay-banner" data-show="$replayed > 0">
					Replayed <span data-text="$replayed"></span> missed events on reconnect
				</div>
				<div class="feed-controls">
					<button class="feed-controls__btn" data-on-click="@set('paused', !$paused)">
						<span data-show="!$paused">Pause</span>
						<span data-show="$paused">Resume</span>
					</button>
				</div>
				<div class="feed-empty" data-show="$totalEvents === 0 && !$paused">
					Waiting for events…
				</div>
				<div class="feed-paused" data-show="$paused">Paused — updates will appear on resume</div>
				<div class="feed" id="feed" data-show="$totalEvents > 0 && !$paused"></div>
				<footer class="card__footer">
					<p>
						Powered by <code>github.com/larsartmann/go-datastar</code>. Events are
						<code>ElementsPatch</code>/<code>SignalsPatch</code> values fanned out via
						go-sse's <code>Broadcaster</code>, replayed via <code>EventStore</code>,
						filtered via <code>SubscribeFilter</code>, kept alive via <code>Heartbeat</code>.
					</p>
				</footer>
				%s
			</div>
		</main>
	</body>
</html>`,
		datastar.ScriptTagVersioned("/datastar.js"),
		filterBar,
		initDiv,
	)
}
