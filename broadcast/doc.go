// Package broadcast provides the connection-lifecycle layer for serving
// DataStar patches over SSE: fan-out, optional reconnection replay, and
// cross-transport hub sharing on top of [go-datastar] patches and the
// [go-sse] transport.
//
// The root go-datastar module is a pure protocol vocabulary — patches are
// values, but the library deliberately manages no connections. This module
// is the optional glue that turns those values into a live SSE endpoint:
//
//	broadcaster := broadcast.NewBroadcasterWithReplay(128)
//	mux.Handle("GET /events", broadcaster)
//
//	broadcaster.Broadcast(datastar.NewElementsPatch("<div>hi</div>",
//		datastar.WithSelectorID("feed")))
//
// It is domain-agnostic: mapping domain events to patches (an EventBridge)
// stays a consumer concern, per the root module's non-goals.
//
// [go-datastar]: https://github.com/LarsArtmann/go-datastar
// [go-sse]: https://github.com/LarsArtmann/go-sse
package broadcast
