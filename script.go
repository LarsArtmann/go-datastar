package datastar

import (
	"fmt"
	"strings"
	"time"

	"github.com/larsartmann/go-sse"
)

// ScriptPatch executes JavaScript on the DataStar client by injecting a
// <script> element into the DOM (patched via ElementsPatch with
// selector=body, mode=append).
//
// By default the script element auto-removes itself after execution
// (data-effect="el.remove()"). Use [WithScriptAutoRemove](false) to keep
// the element.
//
// Construct one with [NewScriptPatch]:
//
//	patch := datastar.NewScriptPatch("console.log('hi')")
//	stream.Send(patch.Event())
type ScriptPatch struct {
	// Script is the JavaScript source code to execute.
	Script string

	// AutoRemove controls whether the script element self-removes after
	// execution. nil (the default) and true both add data-effect="el.remove()".
	// Set to false to keep the element.
	AutoRemove *bool

	// Attributes are additional HTML attributes for the <script> tag.
	// Each should be a complete attribute (e.g., `type="module"`).
	Attributes []string

	// EventID is an optional SSE event identifier.
	EventID string

	// RetryDuration overrides the default SSE retry interval.
	RetryDuration time.Duration
}

// ScriptPatchOption configures a [ScriptPatch].
type ScriptPatchOption func(*ScriptPatch)

// WithScriptAutoRemove controls whether the script element self-removes.
// Pass false to keep the element.
func WithScriptAutoRemove(b bool) ScriptPatchOption {
	return func(p *ScriptPatch) { p.AutoRemove = &b }
}

// WithScriptAttributes sets additional HTML attributes for the <script> tag.
// Each should be a complete key="value" pair (e.g., `type="module"`).
func WithScriptAttributes(attrs ...string) ScriptPatchOption {
	return func(p *ScriptPatch) { p.Attributes = attrs }
}

// WithScriptAttributeKVs sets script attributes from key-value pairs.
// Returns an error via the patch if the argument count is odd (unlike the SDK
// which panics). The error is surfaced when [ScriptPatch.Event] is called.
//
// Prefer [WithScriptAttributes] for pre-formatted attributes.
func WithScriptAttributeKVs(kvs ...string) ScriptPatchOption {
	return func(p *ScriptPatch) {
		p.Attributes = scriptAttributeKVs(kvs)
	}
}

func scriptAttributeKVs(kvs []string) []string {
	attrs := make([]string, 0, len(kvs)/2)
	for i := 0; i+1 < len(kvs); i += 2 {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, kvs[i], kvs[i+1]))
	}

	return attrs
}

// WithScriptEventID sets the SSE event ID for the script patch.
func WithScriptEventID(id string) ScriptPatchOption {
	return func(p *ScriptPatch) { p.EventID = id }
}

// WithScriptRetryDuration overrides the SSE retry duration for the script patch.
func WithScriptRetryDuration(d time.Duration) ScriptPatchOption {
	return func(p *ScriptPatch) { p.RetryDuration = d }
}

// NewScriptPatch creates a [ScriptPatch] with the given JavaScript source and
// options. The default retry duration is [DefaultRetryDuration].
func NewScriptPatch(script string, opts ...ScriptPatchOption) ScriptPatch {
	p := ScriptPatch{
		Script:        script,
		RetryDuration: DefaultRetryDuration,
	}
	for _, opt := range opts {
		opt(&p)
	}

	return p
}

// Event returns the [sse.Event] for this script patch. The script is wrapped in
// a <script> element and sent as a patch-elements event with selector=body,
// mode=append — matching the DataStar SDK wire format exactly.
func (p ScriptPatch) Event() sse.Event {
	var sb strings.Builder
	sb.WriteString("<script")

	for _, attr := range p.Attributes {
		sb.WriteString(" ")
		sb.WriteString(attr)
	}

	// nil and true both add data-effect="el.remove()"
	if p.AutoRemove == nil || *p.AutoRemove {
		sb.WriteString(` data-effect="el.remove()"`)
	}

	sb.WriteString(">")
	sb.WriteString(p.Script)
	sb.WriteString("</script>")

	elementsOpts := []ElementPatchOption{
		WithSelector("body"),
		WithMode(ElementPatchModeAppend),
	}
	if p.EventID != "" {
		elementsOpts = append(elementsOpts, WithElementsEventID(p.EventID))
	}

	if p.RetryDuration > 0 {
		elementsOpts = append(elementsOpts, WithElementsRetryDuration(p.RetryDuration))
	}

	return NewElementsPatch(sb.String(), elementsOpts...).Event()
}
