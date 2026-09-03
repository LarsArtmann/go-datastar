package main

import (
	"compress/gzip"
	"fmt"
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
	return http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(writer, r)

			return
		}

		compressor := gzip.NewWriter(writer)
		defer compressor.Close()

		header := writer.Header()
		header.Del("Content-Length")
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")

		next.ServeHTTP(&gzipSSEWriter{ResponseWriter: writer, gz: compressor}, r)
	})
}

// gzipSSEWriter flushes the compressor on every Flush so events stream
// instead of buffering until the handler returns.
type gzipSSEWriter struct {
	http.ResponseWriter

	gz *gzip.Writer
}

func (writer *gzipSSEWriter) Write(p []byte) (int, error) {
	n, err := writer.gz.Write(p)
	if err != nil {
		return n, fmt.Errorf("gzip write: %w", err)
	}

	return n, nil
}

func (writer *gzipSSEWriter) Flush() {
	if err := writer.gz.Flush(); err != nil {
		return
	}

	if err := http.NewResponseController(writer.ResponseWriter).Flush(); err != nil {
		return
	}
}

var _ http.Flusher = (*gzipSSEWriter)(nil)
