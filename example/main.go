// Example: live feed using go-datastar patches as values with go-sse Broadcaster.
//
// Run: go run ./example/
// Open: http://localhost:8765
package main

import (
	"context"
	"errors"
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

	go startProducer(broadcaster)

	mux := http.NewServeMux()
	mux.Handle("GET /datastar.js", datastar.ScriptHandler())
	mux.HandleFunc("GET /", indexHandler)
	mux.HandleFunc("GET /events", eventsHandler(broadcaster))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("go-datastar example on http://localhost%s", addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}

	if err := broadcaster.Shutdown(shutdownCtx); err != nil {
		log.Printf("broadcaster shutdown: %v", err)
	}
}

func startProducer(b *sse.Broadcaster[sse.Event]) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for i := 1; ; i++ {
		<-ticker.C

		html := fmt.Sprintf(`<div class="item">Item #%d — %s</div>`,
			i, time.Now().Format("15:04:05"))

		elementsPatch := datastar.NewElementsPatch(html,
			datastar.WithSelectorID("feed"),
			datastar.WithMode(datastar.ElementPatchModePrepend),
		)
		b.Broadcast(elementsPatch.Event())

		countPatch, err := datastar.NewSignalsPatch(map[string]any{"total": i})
		if err != nil {
			log.Printf("producer: marshal count signals: %v", err)

			continue
		}
		b.Broadcast(countPatch.Event())
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>go-datastar Example — Live Feed</title>
%s
<style>
body { font-family: system-ui, max-width: 600px; margin: 2rem auto; padding: 0 1rem; color: #1a1a1a; }
h1 { margin-bottom: 0.25rem; }
.stats { display: flex; gap: 1rem; align-items: baseline; margin-bottom: 1rem; color: #666; }
.stats strong { font-size: 1.5rem; color: #1a1a1a; }
#feed { display: flex; flex-direction: column-reverse; gap: 0.5rem; }
.item { padding: 0.75rem 1rem; background: #f4f4f0; border-radius: 0.5rem; animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: none; } }
</style>
</head>
<body>
<div data-signals="{total: 0}">
	<h1>go-datastar Live Feed</h1>
	<div class="stats">
		<span>Total items: <strong data-text="$total">0</strong></span>
	</div>
	<div id="feed"></div>
</div>
<div data-init="@get('/events')"></div>
</body>
</html>`, datastar.ScriptTag("/datastar.js"))
}

func eventsHandler(b *sse.Broadcaster[sse.Event]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() {
			if err := stream.Close(); err != nil {
				log.Printf("events: close stream: %v", err)
			}
		}()

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
