package datastar

import "time"

// DefaultRetryDuration is the default SSE retry interval the DataStar client
// uses after a connection reset (1 second). Patches that leave RetryDuration at
// this value do not emit a retry field; only deviations from this default are
// sent on the wire.
const DefaultRetryDuration = 1000 * time.Millisecond

// EventType is the DataStar protocol event type sent as the SSE event: field.
type EventType string

const (
	// EventTypePatchElements is the event for patching HTML elements into the DOM.
	EventTypePatchElements EventType = "datastar-patch-elements"

	// EventTypePatchSignals is the event for patching reactive signals.
	EventTypePatchSignals EventType = "datastar-patch-signals"
)

// ElementPatchMode controls how an element is merged into the DOM.
type ElementPatchMode string

const (
	// DefaultElementPatchMode is the default mode (outer).
	DefaultElementPatchMode ElementPatchMode = ElementPatchModeOuter

	// ElementPatchModeOuter morphs the element into the existing element.
	ElementPatchModeOuter ElementPatchMode = "outer"

	// ElementPatchModeInner replaces the inner HTML of the existing element.
	ElementPatchModeInner ElementPatchMode = "inner"

	// ElementPatchModeRemove removes the existing element.
	ElementPatchModeRemove ElementPatchMode = "remove"

	// ElementPatchModeReplace replaces the existing element with the new element.
	ElementPatchModeReplace ElementPatchMode = "replace"

	// ElementPatchModePrepend prepends the element inside the existing element.
	ElementPatchModePrepend ElementPatchMode = "prepend"

	// ElementPatchModeAppend appends the element inside the existing element.
	ElementPatchModeAppend ElementPatchMode = "append"

	// ElementPatchModeBefore inserts the element before the existing element.
	ElementPatchModeBefore ElementPatchMode = "before"

	// ElementPatchModeAfter inserts the element after the existing element.
	ElementPatchModeAfter ElementPatchMode = "after"
)

// Namespace is the XML namespace to use when patching elements into the DOM.
type Namespace string

const (
	// NamespaceHTML is the default namespace for HTML elements.
	NamespaceHTML Namespace = "html"

	// NamespaceSVG is the namespace for SVG elements.
	NamespaceSVG Namespace = "svg"

	// NamespaceMathML is the namespace for MathML elements.
	NamespaceMathML Namespace = "mathml"
)

// Dataline key constants. These have a trailing space baked in, matching the
// DataStar wire format (e.g., "selector #feed" is emitted as a data line).
const (
	SelectorDatalineKey               = "selector "
	ModeDatalineKey                   = "mode "
	NamespaceDatalineKey              = "namespace "
	UseViewTransitionDatalineKey      = "useViewTransition "
	ViewTransitionSelectorDatalineKey = "viewTransitionSelector "
	ElementsDatalineKey               = "elements "
	SignalsDatalineKey                = "signals "
	OnlyIfMissingDatalineKey          = "onlyIfMissing "
)

// DatastarKey is the query parameter key for DataStar signals on GET/DELETE requests.
const DatastarKey = "datastar"
