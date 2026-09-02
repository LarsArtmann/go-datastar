# Replay: serving events to reconnecting clients

Replay is the answer to one question: _a browser reconnects after a network
blip — what does it missed?_ Server-Sent Events give the client a way to say
"resume from here" (the `Last-Event-ID` header, or `?datastar=` reconnects);
your server needs an event history to answer with.

go-datastar makes patches first-class values (`type Patch interface {
Event() sse.Event }`), so history is just a list of events you already
produced — nothing about replay is special-cased.

## The pieces

| Piece                     | Where                           | Role                                                   |
| ------------------------- | ------------------------------- | ------------------------------------------------------ |
| `Patch.Event() sse.Event` | go-datastar                     | every patch renders to a storable event                |
| `sse.EventStore`          | go-sse                          | the storage contract: `Append` + `EventsAfter(lastID)` |
| `MemoryStore`             | go-datastar (`store.go`)        | in-memory ring buffer implementing `EventStore`        |
| `Last-Event-ID`           | go-sse (`Stream.LastEventID()`) | the client's resume cursor                             |

## Minimal replay setup

```go
store := datastar.NewMemoryStore(1024) // ring buffer of 1024 events

// On every state change: render the patch, store it, broadcast it.
func publish(store *datastar.MemoryStore, broadcaster *sse.Broadcaster[sse.Event], p datastar.Patch) {
	evt := p.Event()
	store.Append(evt)
	broadcaster.BroadcastMany(evt)
}
```

`MemoryStore.EventsAfter(lastID)` returns everything after the cursor in
order. The ring buffer drops the oldest events when full — a client that was
gone longer than `capacity` events cannot be fully caught up; reconnection
then replays the retained tail. Size the buffer for your longest tolerated
disconnect, not forever.

## The reconnection path

When a DataStar client reconnects, the browser sends
`Last-Event-ID: <id>`. Serve the backlog, then merge into the live stream:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	stream, err := sse.NewStream(w, r)
	if err != nil {
		return
	}
	defer func() { _ = stream.Close() }()
	resp := datastar.NewResponse(stream)

	// 1. Backlog: everything the client missed. The browser sends the
	//    Last-Event-ID header on reconnect; go-sse surfaces it typed.
	if backlog, err := store.EventsAfter(stream.LastEventID()); err == nil {
		for _, evt := range backlog {
			_ = resp.Send(evt)
		}
	}

	// 2. Live: subscribe AFTER the backlog is sent so no event is
	//    duplicated (events broadcast while we sent the backlog are
	//    covered by the next subscribe + dedupe, or use a broadcaster
	//    with replay support from go-sse).
	events := broadcaster.Subscribe()
	defer broadcaster.Unsubscribe(events)

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			_ = resp.Send(evt)
		}
	}
}
```

The exact backlog/live handoff (and its duplicate window) is a policy
choice — see the `example/` app for a working broadcaster + MemoryStore
combination.

## Testing replay with datastartest

`datastartest.CollectWithRequest` + `WithLastEventID` simulates the
reconnecting browser end to end:

```go
events := datastartest.CollectWithRequest(t, handler,
	http.MethodGet, "", "",
	datastartest.WithLastEventID("42"),
)

// The first events back must be the missed ones.
datastartest.RequireEventID(t, events[0], "43")
```

## What go-datastar deliberately does NOT do

- **No persistence.** `MemoryStore` is in-memory; a restart loses history.
  Implement `sse.EventStore` over your database if you need durable replay —
  the interface is two methods.
- **No global sequence.** IDs come from whatever you store in the event's
  `id:` field; `MemoryStore` expects monotonic numeric IDs.
- **No per-client cursor management.** The cursor lives in the browser
  (SSE standard); the server is stateless per connection.
