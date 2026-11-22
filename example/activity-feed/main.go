// Example: the live activity feed — DataStar patches as values over go-sse
// transport, showcasing the full stack:
//   - Patch API: ElementsPatch/SignalsPatch values with event IDs for replay
//   - Broadcaster: multi-client fan-out (open multiple tabs to see it)
//   - SubscribeFilter: "alerts only" view via ?filter=alerts query param
//   - EventStore + Replay: missed events replayed on reconnect (Last-Event-ID)
//   - Heartbeat: keeps connections alive through proxies
//   - ScriptHandler + ScriptTagVersioned: pinned DataStar JS, no vendoring
//
// Run: go run ./example/activity-feed/
// Then: open http://localhost:8767 in your browser
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	datastar "github.com/larsartmann/go-datastar"
)

const (
	defaultFeedPort = "8767"
	emitInterval    = 2 * time.Second
	heartbeatEvery  = 15 * time.Second
	maxStoredEvents = 50
	shutdownTimeout = 5 * time.Second
	readHeaderLimit = 5 * time.Second
)

// idleTimeout is a generous idle timeout for long-lived SSE connections.
// ReadTimeout/WriteTimeout would kill SSE connections, so we use
// IdleTimeout instead which only applies to idle keep-alive connections.
const idleTimeout = 5 * time.Minute

//go:embed all:static
var staticFiles embed.FS

// listenAddr returns ":$PORT" when set, else the default port — lets smoke
// tests pick a free port without editing the example.
func listenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return ":" + defaultFeedPort
}

func main() {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static sub FS: %v", err)
	}

	server := newActivityServer()
	feedAddr := listenAddr()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go server.startProducer(ctx)

	mux := http.NewServeMux()
	mux.Handle("GET /datastar.js", datastar.ScriptHandler())
	mux.HandleFunc("GET /", server.indexHandler)
	mux.HandleFunc("GET /events", server.eventsHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	//nolint:exhaustruct // example server, most fields are intentionally default
	feedServer := &http.Server{
		Addr:              feedAddr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderLimit,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		log.Printf("go-datastar activity feed on http://localhost%s", feedAddr)
		log.Print("Open the URL in multiple browser tabs to see real-time fan-out.")

		if err := feedServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Print("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Graceful shutdown sequence:
	// 1. HTTP server drains active requests (including SSE connections)
	// 2. Broadcaster drains subscriber buffers, then closes channels
	if err := feedServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown: %v", err)
	}

	if err := server.broadcaster.Shutdown(shutdownCtx); err != nil {
		log.Printf("broadcaster shutdown: %v", err)
	}
}
