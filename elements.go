package datastar

import (
	"fmt"
	"strings"
	"time"

	"github.com/larsartmann/go-sse"
)

// ElementsPatch patches HTML elements into the DOM on the DataStar client.
// It is the most-used DataStar patch type.
//
// Construct one with [NewElementsPatch] and functional options:
//
//	patch := datastar.NewElementsPatch("<div>Hello</div>",
//	    datastar.WithSelector("#feed"),
//	    datastar.WithMode(datastar.ElementPatchModeInner),
//	)
//	stream.Send(patch.Event())
type ElementsPatch struct {
	// Selector is the CSS selector for the target element.
	// Empty means no selector data line is emitted.
	Selector string

	// Mode controls how the element is merged. The default (outer) is never
	// emitted on the wire — it is the DataStar client's default behavior.
	Mode ElementPatchMode

	// Namespace specifies the XML namespace. The default (html) is never
	// emitted on the wire.
	Namespace Namespace

	// UseViewTransitions enables the View Transition API for the merge.
	UseViewTransitions bool

	// ViewTransitionSelector scopes the view transition to a specific element.
	ViewTransitionSelector string

	// HTML is the element content to patch into the DOM.
	HTML string

	// EventID is an optional SSE event identifier for reconnection replay.
	EventID string

	// RetryDuration overrides the default SSE retry interval. Only emitted
	// on the wire when > 0 and != [DefaultRetryDuration].
	RetryDuration time.Duration
}

// ElementPatchOption configures an [ElementsPatch].
type ElementPatchOption func(*ElementsPatch)

// WithSelector sets the CSS selector for the element patch target.
func WithSelector(selector string) ElementPatchOption {
	return func(p *ElementsPatch) { p.Selector = selector }
}

// WithSelectorf is a printf-style variant of [WithSelector].
func WithSelectorf(format string, args ...any) ElementPatchOption {
	return WithSelector(fmt.Sprintf(format, args...))
}

// WithMode overrides the [DefaultElementPatchMode] for the element.
func WithMode(mode ElementPatchMode) ElementPatchOption {
	return func(p *ElementsPatch) { p.Mode = mode }
}

// WithNamespace specifies the XML namespace for the element.
func WithNamespace(ns Namespace) ElementPatchOption {
	return func(p *ElementsPatch) { p.Namespace = ns }
}

// WithViewTransitions enables the View Transition API for the merge.
func WithViewTransitions(enable bool) ElementPatchOption {
	return func(p *ElementsPatch) { p.UseViewTransitions = enable }
}

// WithViewTransitionSelector scopes the view transition to a CSS selector.
func WithViewTransitionSelector(selector string) ElementPatchOption {
	return func(p *ElementsPatch) { p.ViewTransitionSelector = selector }
}

// WithElementsEventID sets the SSE event ID for the element patch.
func WithElementsEventID(id string) ElementPatchOption {
	return func(p *ElementsPatch) { p.EventID = id }
}

// WithElementsRetryDuration overrides the SSE retry duration for the element patch.
func WithElementsRetryDuration(d time.Duration) ElementPatchOption {
	return func(p *ElementsPatch) { p.RetryDuration = d }
}

// NewElementsPatch creates an [ElementsPatch] with the given HTML and options.
// The default mode is [DefaultElementPatchMode] (outer), which is never emitted
// on the wire.
func NewElementsPatch(html string, opts ...ElementPatchOption) ElementsPatch {
	patch := ElementsPatch{
		HTML:          html,
		Mode:          DefaultElementPatchMode,
		RetryDuration: DefaultRetryDuration,
	}
	for _, opt := range opts {
		opt(&patch)
	}

	return patch
}

// Event returns the [sse.Event] for this element patch. The data lines are
// constructed in the exact order the DataStar JS client expects:
//
//  1. selector (if non-empty)
//  2. mode (if not outer)
//  3. namespace (if non-empty and not html)
//  4. useViewTransition true (if enabled)
//  5. viewTransitionSelector (if non-empty and view transitions enabled)
//  6. elements <line> (one per line of HTML)
func (p ElementsPatch) Event() sse.Event {
	var dataLines []string

	if p.Selector != "" {
		dataLines = append(dataLines, SelectorDatalineKey+p.Selector)
	}

	if p.Mode != ElementPatchModeOuter {
		dataLines = append(dataLines, ModeDatalineKey+string(p.Mode))
	}

	if p.Namespace != "" && p.Namespace != NamespaceHTML {
		dataLines = append(dataLines, NamespaceDatalineKey+string(p.Namespace))
	}

	if p.UseViewTransitions {
		dataLines = append(dataLines, UseViewTransitionDatalineKey+"true")

		if p.ViewTransitionSelector != "" {
			dataLines = append(
				dataLines,
				ViewTransitionSelectorDatalineKey+p.ViewTransitionSelector,
			)
		}
	}

	if p.HTML != "" {
		for line := range strings.SplitSeq(p.HTML, "\n") {
			dataLines = append(dataLines, ElementsDatalineKey+line)
		}
	}

	evt := sse.Event{
		Event: string(EventTypePatchElements),
		Data:  sse.JoinLines(dataLines...),
	}

	if p.EventID != "" {
		evt.ID = sse.NewEventID(p.EventID)
	}

	if p.RetryDuration > 0 && p.RetryDuration != DefaultRetryDuration {
		evt.Retry = uint(
			p.RetryDuration.Milliseconds(),
		)
	}

	return evt
}
