package datastar

import (
	"encoding/json/v2"
	"fmt"
	"net/url"
	"strings"

	"github.com/larsartmann/go-sse"
)

// NewRedirectPatch creates a [ScriptPatch] that redirects the browser to the
// given URL using setTimeout.
func NewRedirectPatch(targetURL string, opts ...ScriptPatchOption) ScriptPatch {
	js := fmt.Sprintf("setTimeout(() => window.location.href = %q)", targetURL)

	return NewScriptPatch(js, opts...)
}

// NewRedirectfPatch is a printf-style variant of [NewRedirectPatch].
func NewRedirectfPatch(format string, args ...any) ScriptPatch {
	return NewRedirectPatch(fmt.Sprintf(format, args...))
}

// NewConsoleLogPatch creates a [ScriptPatch] that calls console.log with the
// given message. The message is JS-quoted via %q.
func NewConsoleLogPatch(msg string, opts ...ScriptPatchOption) ScriptPatch {
	return NewScriptPatch(fmt.Sprintf("console.log(%q)", msg), opts...)
}

// NewConsoleLogfPatch is a printf-style variant of [NewConsoleLogPatch].
func NewConsoleLogfPatch(format string, args ...any) ScriptPatch {
	return NewConsoleLogPatch(fmt.Sprintf(format, args...))
}

// NewConsoleErrorPatch creates a [ScriptPatch] that calls console.error with
// the given error's message. The message is JS-quoted via %q.
func NewConsoleErrorPatch(err error, opts ...ScriptPatchOption) ScriptPatch {
	return NewScriptPatch(fmt.Sprintf("console.error(%q)", err.Error()), opts...)
}

// DispatchCustomEventPatch dispatches a custom DOM event on the client via
// script execution. The detail value is marshaled to JSON and passed as the
// event's detail property.
type DispatchCustomEventPatch struct {
	EventName string
	Detail    any

	Selector   string
	Bubbles    bool
	Cancelable bool
	Composed   bool

	EventID       string
	RetryDuration int64 // milliseconds; 0 = default
}

// DispatchCustomEventOption configures a [DispatchCustomEventPatch].
type DispatchCustomEventOption func(*DispatchCustomEventPatch)

// WithCustomEventSelector replaces the default target (document) with a CSS selector.
func WithCustomEventSelector(s string) DispatchCustomEventOption {
	return func(p *DispatchCustomEventPatch) { p.Selector = s }
}

// WithCustomEventBubbles overrides the default bubbling (true).
func WithCustomEventBubbles(b bool) DispatchCustomEventOption {
	return func(p *DispatchCustomEventPatch) { p.Bubbles = b }
}

// WithCustomEventCancelable overrides the default cancelability (true).
func WithCustomEventCancelable(b bool) DispatchCustomEventOption {
	return func(p *DispatchCustomEventPatch) { p.Cancelable = b }
}

// WithCustomEventComposed overrides the default composed flag (true).
func WithCustomEventComposed(b bool) DispatchCustomEventOption {
	return func(p *DispatchCustomEventPatch) { p.Composed = b }
}

// WithCustomEventEventID sets the SSE event ID.
func WithCustomEventEventID(id string) DispatchCustomEventOption {
	return func(p *DispatchCustomEventPatch) { p.EventID = id }
}

const defaultCustomEventSelector = "document"

// NewDispatchCustomEventPatch creates a [DispatchCustomEventPatch] with the
// given event name and detail value. The detail is marshaled to JSON when
// [DispatchCustomEventPatch.Event] is called.
func NewDispatchCustomEventPatch(
	eventName string,
	detail any,
	opts ...DispatchCustomEventOption,
) (DispatchCustomEventPatch, error) {
	if eventName == "" {
		return DispatchCustomEventPatch{}, ErrEventNameRequired
	}

	patch := DispatchCustomEventPatch{
		EventName:  eventName,
		Detail:     detail,
		Selector:   defaultCustomEventSelector,
		Bubbles:    true,
		Cancelable: true,
		Composed:   true,
	}
	for _, opt := range opts {
		opt(&patch)
	}

	return patch, nil
}

// Event returns the [sse.Event] for this custom event dispatch.
func (p DispatchCustomEventPatch) Event() sse.Event {
	detailsJSON, err := json.Marshal(p.Detail)
	if err != nil {
		detailsJSON = []byte("null")
	}

	elementsJS := "[document]"
	if p.Selector != "" && p.Selector != defaultCustomEventSelector {
		elementsJS = fmt.Sprintf(`document.querySelectorAll(%q)`, p.Selector)
	}

	scriptBody := fmt.Sprintf(`{
	const elements = %s

	const event = new CustomEvent(%q, {
		bubbles: %t,
		cancelable: %t,
		composed: %t,
		detail: %s,
	});

	elements.forEach((element) => {
		element.dispatchEvent(event);
	});
}
	`,
		elementsJS,
		p.EventName,
		p.Bubbles,
		p.Cancelable,
		p.Composed,
		string(detailsJSON),
	)

	scriptOpts := []ScriptPatchOption{}
	if p.EventID != "" {
		scriptOpts = append(scriptOpts, WithScriptEventID(p.EventID))
	}

	return NewScriptPatch(scriptBody, scriptOpts...).Event()
}

// NewReplaceURLPatch creates a [ScriptPatch] that replaces the browser URL
// using history.replaceState.
func NewReplaceURLPatch(u url.URL, opts ...ScriptPatchOption) ScriptPatch {
	js := fmt.Sprintf(`window.history.replaceState({}, "", %q)`, u.String())

	return NewScriptPatch(js, opts...)
}

// NewPrefetchPatch creates a [ScriptPatch] that injects a speculation rules
// JSON block to prefetch the given URLs.
func NewPrefetchPatch(urls ...string) ScriptPatch {
	quoted := make([]string, 0, len(urls))
	for _, u := range urls {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, u))
	}

	script := fmt.Sprintf(`{
	"prefetch": [
		{
			"source": "list",
			"urls": [
				%s
			]
		}
	]
		}`, strings.Join(quoted, ",\n\t\t\t\t"))

	return NewScriptPatch(script,
		WithScriptAutoRemove(false),
		WithScriptAttributes(`type="speculationrules"`),
	)
}
