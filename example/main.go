// Example: live feed using go-datastar patches as values with go-sse Broadcaster.
//
// Run: go run ./example/
// Open: http://localhost:8765
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

const addr = ":8765"

func main() {
	broadcaster := sse.NewBroadcaster[sse.Event]()
	defer broadcaster.Close()

	// Background producer: emits a new feed item every 2 seconds
	go startProducer(broadcaster)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /datastar.js", datastar.ScriptHandler().ServeHTTP)
	mux.HandleFunc("GET /", indexHandler)
	mux.HandleFunc("GET /events", eventsHandler(broadcaster))

	srv := &http.Server{Addr: addr, Handler: mux}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("go-datastar example on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	_ = broadcaster.Shutdown(shutdownCtx)
}

func startProducer(b *sse.Broadcaster[sse.Event]) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for i := 1; ; i++ {
		<-ticker.C

		html := fmt.Sprintf(`<div class="item" data-signals-count="%d">Item #%d — %s</div>`,
			i, i, time.Now().Format("15:04:05"))

		patch := datastar.NewElementsPatch(html,
			datastar.WithSelectorID("feed"),
			datastar.WithMode(datastar.ElementPatchModePrepend),
		)
		b.Broadcast(patch.Event())

		countPatch, _ := datastar.NewSignalsPatch(map[string]any{"total": i})
		b.Broadcast(countPatch.Event())
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>go-datastar Example</title>
%s
<style>
body { font-family: system-ui; max-width: 600px; margin: 2rem auto; padding: 0 1rem; }
#feed { display: flex; flex-direction: column; gap: 0.5rem; }
.item { padding: 0.75rem; background: #f4f4f0; border-radius: 0.5rem; }
</style>
</head>
<body>
<h1>go-datastar Live Feed</h1>
<p>Total items: <span data-signals-total="0">0</span></p>
<div id="feed" data-on-evt-feed-item="__event"></div>
<script type="module">
import { SSE } from "/events";
const es = new EventSource("/events");
es.addEventListener("datastar-patch-elements", (e) => {
	const parser = new DOMParser();
	const doc = parser.parseFromString(e.data.replace(/^data: /gm, ""), "text/html");
	console.log("elements event", e.data);
});
es.addEventListener("datastar-patch-signals", (e) => {
	console.log("signals event", e.data);
});
</script>
</body>
</html>`, datastar.ScriptTag("/datastar.js"))
}

func eventsHandler(b *sse.Broadcaster[sse.Event]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		ch := b.Subscribe()
		defer b.Unsubscribe(ch)

		for {
			select {
			case <-r.Context().Done():
				return
			case evt, ok := <-ch:
				if !ok {
					return
				}
				if err := stream.Send(evt); err != nil {
					return
				}
			}
		}
	}
}
