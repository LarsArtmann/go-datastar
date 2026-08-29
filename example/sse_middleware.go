package main

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// gzipSSEMiddleware compresses SSE responses for clients that send
// Accept-Encoding: gzip. It exists to document the pattern: go-datastar
// deliberately leaves compression to middleware (unlike the upstream SDK,
// which embeds it in the generator), because compression is a transport
// concern owned by the HTTP layer — a reverse proxy (nginx, Caddy) usually
// does this better with idle-buffering tuned for SSE.
//
// SSE-specific details this middleware gets right:
//   - flush the gzip writer after EVERY write, not at handler end — SSE
//     consumers (the DataStar client) need events delivered immediately;
//   - strip Content-Length (streaming); set Vary so caches do not serve
//     compressed responses to uncompressed clients;
//   - leave text/event-stream negotiation to the handler.
func gzipSSEMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)

			return
		}

		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()

		header := w.Header()
		header.Del("Content-Length")
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")

		next.ServeHTTP(&gzipSSEWriter{ResponseWriter: w, gz: gz}, r)
	})
}

// gzipSSEWriter flushes the compressor on every Flush so events stream
// instead of buffering until the handler returns.
type gzipSSEWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipSSEWriter) Write(p []byte) (int, error) {
	return w.gz.Write(p)
}

func (w *gzipSSEWriter) Flush() {
	if err := w.gz.Flush(); err != nil {
		return
	}

	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

var _ http.Flusher = (*gzipSSEWriter)(nil)
