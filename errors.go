package datastar

import (
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
)

// go-datastar returns classified errors built on [go-error-family]. Every error
// carries a stable machine-readable code (below), a behavioral [errorfamily.Family]
// (Rejection, Transient, Orchestration), and structured context for diagnosis.
//
// Callers have three strongly typed ways to handle errors:
//
//  1. By code:     if errorfamily.Code(err) == datastar.CodeSignalsMarshalFailed
//  2. By sentinel: if errors.Is(err, datastar.ErrEventNameRequired)
//  3. By family:   if errorfamily.Classify(err) == errorfamily.Transient
//
// Families are chosen by whose fault the failure is and whether retrying helps:
//
//	Rejection     — bad or missing input from the caller. Not retryable.
//	Transient     — temporary I/O failure (e.g. reading the request body). Retryable.
//	Orchestration — an internal operation failed to produce output (rendering). Not retryable.
//
// Codes follow the convention "datastar.<operation>_<failure>".

// Error codes for go-datastar. Each is a stable string accessible via
// [errorfamily.Code], enabling programmatic handling, metrics, and structured
// logging without string matching on human-readable messages.
const (
	// CodeTemplRenderFailed: a [TemplComponent] failed to render to HTML.
	CodeTemplRenderFailed = "datastar.templ_render_failed"

	// CodeGostarRenderFailed: a [GoStarElementRenderer] failed to render to HTML.
	CodeGostarRenderFailed = "datastar.gostar_render_failed"

	// CodeBodyReadAfterClose: [ReadSignals] read the request body after it was
	// already closed, typically because an SSE stream consumed it first.
	CodeBodyReadAfterClose = "datastar.body_read_after_close"

	// CodeBodyReadFailed: [ReadSignals] could not read the request body.
	CodeBodyReadFailed = "datastar.body_read_failed"

	// CodeSignalsUnmarshalFailed: the inbound signals JSON could not be
	// unmarshaled into the caller's target.
	CodeSignalsUnmarshalFailed = "datastar.signals_unmarshal_failed"

	// CodeSignalsMarshalFailed: a Go value could not be marshaled to JSON for a
	// signals patch (e.g. a channel, function, or cyclic reference).
	CodeSignalsMarshalFailed = "datastar.signals_marshal_failed"

	// CodeEventNameRequired: [NewDispatchCustomEventPatch] was called with an
	// empty event name.
	CodeEventNameRequired = "datastar.event_name_required"

	// CodeElementPatchModeInvalid: [ElementPatchModeFromString] received an
	// unrecognized mode string.
	CodeElementPatchModeInvalid = "datastar.element_patch_mode_invalid"

	// CodeNamespaceInvalid: [NamespaceFromString] received an unrecognized
	// namespace string.
	CodeNamespaceInvalid = "datastar.namespace_invalid"

	// CodeStreamSendFailed: [Response] could not deliver an SSE event to the
	// underlying stream. Wraps any error returned by the transport.
	CodeStreamSendFailed = "datastar.stream_send_failed"
)

// Sentinel errors for fixed-message failure modes. Match with [errors.Is]:
//
//	if errors.Is(err, datastar.ErrEventNameRequired) { ... }
//
// [errorfamily] errors compare by (code, family), so a context-enriched clone
// returned by [errorfamily.Error.WithContext] still satisfies errors.Is against
// the sentinel. Dynamic-message errors (render failures, parse failures with the
// offending value) are constructed inline at their call site using the code
// constants above.
var (
	// ErrBodyReadAfterClose is returned by [ReadSignals] when the request body
	// was already closed before reading. This usually means an SSE stream was
	// created before ReadSignals ran; re-order so ReadSignals reads the body
	// first. The underlying [http.ErrBodyReadAfterClose] is preserved as the cause.
	ErrBodyReadAfterClose = errorfamily.WrapRejection(
		http.ErrBodyReadAfterClose,
		CodeBodyReadAfterClose,
		"request body already closed (create the SSE stream after calling ReadSignals)",
	)

	// ErrEventNameRequired is returned by [NewDispatchCustomEventPatch] when the
	// event name argument is empty.
	ErrEventNameRequired = errorfamily.NewRejection(
		CodeEventNameRequired,
		"eventName is required",
	)
)
