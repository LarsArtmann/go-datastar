package datastar

import (
	"net/http"
	"net/url"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// Response wraps an [sse.Stream] and provides fluent methods for sending
// DataStar patches on a single HTTP connection. Each method constructs a Patch,
// calls its [Patch.Event] method, and sends the resulting [sse.Event] via the
// underlying stream.
//
// Create one per HTTP handler:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    stream := sse.NewStream(w, r)
//	    defer func() { _ = stream.Close() }()
//
//	    resp := datastar.NewResponse(stream)
//
//	    if err := resp.PatchElements("<div>Hello</div>", datastar.WithSelector("#feed")); err != nil {
//	        log.Printf("patch elements: %v", err)
//	        return
//	    }
//
//	    if err := resp.MarshalAndPatchSignals(map[string]any{"count": 1}); err != nil {
//	        log.Printf("patch signals: %v", err)
//	    }
//	}
type Response struct {
	stream *sse.Stream
}

// NewResponse creates a [Response] wrapping the given [sse.Stream].
func NewResponse(stream *sse.Stream) *Response {
	return &Response{stream: stream}
}

func wrapStreamError(err error) error {
	if err == nil {
		return nil
	}

	return errorfamily.WrapTransient(err, CodeStreamSendFailed, "send SSE event")
}

func (r *Response) PatchElements(html string, opts ...ElementPatchOption) error {
	return wrapStreamError(r.stream.Send(NewElementsPatch(html, opts...).Event()))
}

// PatchElementsTempl renders a [TemplComponent] to HTML and sends it as an
// [ElementsPatch].
func (r *Response) PatchElementsTempl(c TemplComponent, opts ...ElementPatchOption) error {
	patch, err := ElementsFromTempl(c, opts...)
	if err != nil {
		return err
	}

	return wrapStreamError(r.stream.Send(patch.Event()))
}

// PatchSignals sends a [SignalsPatch] with the given pre-encoded JSON.
func (r *Response) PatchSignals(signalsJSON []byte, opts ...SignalsPatchOption) error {
	patch := SignalsPatch{
		Signals:       signalsJSON,
		RetryDuration: DefaultRetryDuration,
	}
	for _, opt := range opts {
		opt(&patch)
	}

	return wrapStreamError(r.stream.Send(patch.Event()))
}

// MarshalAndPatchSignals marshals a Go value to JSON and sends it as a
// [SignalsPatch]. Returns an error if marshaling fails.
func (r *Response) MarshalAndPatchSignals(v any, opts ...SignalsPatchOption) error {
	patch, err := NewSignalsPatch(v, opts...)
	if err != nil {
		return err
	}

	return wrapStreamError(r.stream.Send(patch.Event()))
}

// RemoveElement sends an [ElementsPatch] that removes the given selector.
func (r *Response) RemoveElement(selector string, opts ...ElementPatchOption) error {
	return wrapStreamError(r.stream.Send(NewRemovePatch(selector, opts...).Event()))
}

// RemoveElementByID sends an [ElementsPatch] that removes the element with the
// given ID.
func (r *Response) RemoveElementByID(id string, opts ...ElementPatchOption) error {
	return wrapStreamError(r.stream.Send(NewRemoveByIDPatch(id, opts...).Event()))
}

// ExecuteScript sends a [ScriptPatch] on the underlying stream.
func (r *Response) ExecuteScript(script string, opts ...ScriptPatchOption) error {
	return wrapStreamError(r.stream.Send(NewScriptPatch(script, opts...).Event()))
}

// Redirect sends a redirect [ScriptPatch].
func (r *Response) Redirect(targetURL string, opts ...ScriptPatchOption) error {
	return wrapStreamError(r.stream.Send(NewRedirectPatch(targetURL, opts...).Event()))
}

// ConsoleLog sends a console.log [ScriptPatch].
func (r *Response) ConsoleLog(msg string, opts ...ScriptPatchOption) error {
	return wrapStreamError(r.stream.Send(NewConsoleLogPatch(msg, opts...).Event()))
}

// ConsoleError sends a console.error [ScriptPatch].
func (r *Response) ConsoleError(err error, opts ...ScriptPatchOption) error {
	return wrapStreamError(r.stream.Send(NewConsoleErrorPatch(err, opts...).Event()))
}

// DispatchCustomEvent dispatches a custom DOM event on the client.
func (r *Response) DispatchCustomEvent(
	eventName string,
	detail any,
	opts ...DispatchCustomEventOption,
) error {
	patch, err := NewDispatchCustomEventPatch(eventName, detail, opts...)
	if err != nil {
		return err
	}

	return wrapStreamError(r.stream.Send(patch.Event()))
}

// ReplaceURL sends a replaceState [ScriptPatch].
func (r *Response) ReplaceURL(u url.URL, opts ...ScriptPatchOption) error {
	return wrapStreamError(r.stream.Send(NewReplaceURLPatch(u, opts...).Event()))
}

// ReplaceURLQuerystring replaces the browser URL's query string with the
// encoded values, preserving the request's path (upstream-parity
// convenience over [NewReplaceURLQuerystringPatch]).
func (r *Response) ReplaceURLQuerystring(
	req *http.Request,
	values url.Values,
	opts ...ScriptPatchOption,
) error {
	return wrapStreamError(
		r.stream.Send(NewReplaceURLQuerystringPatch(req, values, opts...).Event()),
	)
}

// Prefetch sends a speculation rules [ScriptPatch] to prefetch the given URLs.
func (r *Response) Prefetch(urls ...string) error {
	return wrapStreamError(r.stream.Send(NewPrefetchPatch(urls...).Event()))
}

// ApplyPatches sends multiple patches in sequence.
func (r *Response) ApplyPatches(patches ...Patch) error {
	for _, p := range patches {
		if err := wrapStreamError(r.stream.Send(p.Event())); err != nil {
			return err
		}
	}

	return nil
}

// Send sends a raw [sse.Event] on the underlying stream.
func (r *Response) Send(evt sse.Event) error {
	return wrapStreamError(r.stream.Send(evt))
}

// Stream returns the underlying [sse.Stream].
func (r *Response) Stream() *sse.Stream { return r.stream }

const signalKeyMessage = "message"

// sendSignalsMap builds a [SignalsPatch] from a key→value map and sends it on
// the stream. It is the shared core of [ErrorResponse] and [NotificationResponse].
func sendSignalsMap(stream *sse.Stream, signals map[string]any) error {
	patch, err := NewSignalsPatch(signals)
	if err != nil {
		return err
	}

	return wrapStreamError(stream.Send(patch.Event()))
}

// ErrorResponse sends a signals patch with error information that the
// DataStar client can display.
func ErrorResponse(stream *sse.Stream, message string, code string) error {
	return sendSignalsMap(stream, map[string]any{
		"error": map[string]any{
			signalKeyMessage: message,
			"code":           code,
		},
	})
}

// ErrorResponseFromError sends a signals patch with error metadata extracted
// from a Go error using [errorfamily] classification. The payload includes the
// error message, stable code, behavioral family, retryability, and the HTTP
// status code that the family maps to — giving the DataStar client enough
// context to render an appropriate error UI.
//
// For non-errorfamily errors, code will be empty and Classify defaults to
// Transient (fail-open for retry), so family will be "transient", retryable
// will be true, and HTTPStatus will be 503.
//
// A nil error is caller misuse and returns a classified Rejection
// ([CodeErrorResponseNilError]) without sending anything.
func ErrorResponseFromError(stream *sse.Stream, err error) error {
	if err == nil {
		return errorfamily.NewRejection(CodeErrorResponseNilError, "ErrorResponseFromError called with nil error")
	}

	return sendSignalsMap(stream, map[string]any{
		"error": map[string]any{
			signalKeyMessage: err.Error(),
			"code":           errorfamily.Code(err),
			"family":         errorfamily.Classify(err).String(),
			"retryable":      errorfamily.IsRetryable(err),
			"httpStatus":     errorfamily.HTTPStatus(err),
		},
	})
}

// NotificationResponse sends a signals patch with a notification message.
func NotificationResponse(stream *sse.Stream, message string, kind string) error {
	return sendSignalsMap(stream, map[string]any{
		"notification": map[string]any{
			signalKeyMessage: message,
			"kind":           kind,
			"time":           time.Now().Unix(),
		},
	})
}

// ErrorResponse is the method form of [ErrorResponse] on this response's
// stream — the fluent spelling when you already hold a [Response].
func (r *Response) ErrorResponse(message, code string) error {
	return ErrorResponse(r.stream, message, code)
}

// ErrorResponseFromError is the method form of [ErrorResponseFromError] on
// this response's stream.
func (r *Response) ErrorResponseFromError(err error) error {
	return ErrorResponseFromError(r.stream, err)
}

// NotificationResponse is the method form of [NotificationResponse] on this
// response's stream.
func (r *Response) NotificationResponse(message, kind string) error {
	return NotificationResponse(r.stream, message, kind)
}

// NewResponseFromHTTP is a convenience that creates an [sse.Stream] from the
// ResponseWriter and Request, then wraps it in a [Response].
func NewResponseFromHTTP(w http.ResponseWriter, r *http.Request) *Response {
	return NewResponse(sse.NewStream(w, r))
}
