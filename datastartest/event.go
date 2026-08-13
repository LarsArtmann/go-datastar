package datastartest

import (
	"encoding/json/v2"
	"fmt"
	"strings"

	"github.com/larsartmann/go-datastar"
	errorfamily "github.com/larsartmann/go-error-family"
)

// Event is a DataStar SSE event decoded from the wire format. It preserves
// the raw data lines and provides typed accessors that decode the DataStar
// dataline key-value pairs.
//
// Fields:
//   - Type is the SSE event type (e.g., "datastar-patch-elements").
//   - DataLines are the individual data: lines with their key prefixes intact.
//   - ID is the optional SSE event ID.
//   - Retry is the optional reconnection interval in milliseconds.
type Event struct {
	Type      string
	DataLines []string
	ID        string
	Retry     uint
}

// IsElements reports whether the event type is datastar-patch-elements.
func (e Event) IsElements() bool { return e.Type == string(datastar.EventTypePatchElements) }

// IsSignals reports whether the event type is datastar-patch-signals.
func (e Event) IsSignals() bool { return e.Type == string(datastar.EventTypePatchSignals) }

// IsScript reports whether the event is a patch-elements event whose HTML
// content starts with a <script> tag. Script patches (ExecuteScript, Redirect,
// ConsoleLog, ConsoleError, DispatchCustomEvent, ReplaceURL, Prefetch) all
// produce elements patches wrapping JavaScript in <script> tags.
func (e Event) IsScript() bool {
	return e.IsElements() && strings.HasPrefix(e.Elements(), "<script")
}

// Selector returns the CSS selector from a patch-elements event.
// Returns empty if not present (the client defaults to the merging element).
func (e Event) Selector() string {
	return e.firstValue(datastar.SelectorDatalineKey)
}

// Mode returns the element patch mode (e.g., "append", "inner", "remove").
// Returns "outer" (the DataStar default) if no mode dataline is present.
func (e Event) Mode() string {
	if m := e.firstValue(datastar.ModeDatalineKey); m != "" {
		return m
	}

	return string(datastar.DefaultElementPatchMode)
}

// Namespace returns the XML namespace for a patch-elements event.
// Returns "html" (the DataStar default) if no namespace dataline is present.
func (e Event) Namespace() string {
	if ns := e.firstValue(datastar.NamespaceDatalineKey); ns != "" {
		return ns
	}

	return string(datastar.NamespaceHTML)
}

// Elements returns the HTML content from a patch-elements event. Multi-line
// HTML that was split across multiple "elements" datalines is rejoined with
// "\n", reconstructing the original content.
func (e Event) Elements() string {
	return strings.Join(e.allValues(datastar.ElementsDatalineKey), "\n")
}

// SignalsJSON returns the raw JSON bytes from a patch-signals event.
// Multi-line JSON that was split across multiple "signals" datalines is
// rejoined with "\n", reconstructing the original payload.
func (e Event) SignalsJSON() []byte {
	return []byte(strings.Join(e.allValues(datastar.SignalsDatalineKey), "\n"))
}

// UnmarshalSignals decodes the signals JSON payload from a patch-signals event
// into the target. The target must be a pointer.
func (e Event) UnmarshalSignals(target any) error {
	raw := e.SignalsJSON()

	if err := json.Unmarshal(raw, target); err != nil {
		preview := string(raw)

		const maxPreviewLen = 200

		if len(preview) > maxPreviewLen {
			preview = preview[:maxPreviewLen] + "..."
		}

		return errorfamily.WrapRejectionf(
			err,
			CodeSignalsUnmarshalFailed,
			"unmarshal signals JSON %q",
			preview,
		)
	}

	return nil
}

// UseViewTransitions reports whether the event enables the View Transition API.
func (e Event) UseViewTransitions() bool {
	return e.firstValue(datastar.UseViewTransitionDatalineKey) == "true"
}

// ViewTransitionSelector returns the scoped view transition selector, if any.
func (e Event) ViewTransitionSelector() string {
	return e.firstValue(datastar.ViewTransitionSelectorDatalineKey)
}

// OnlyIfMissing reports whether the signals patch has the onlyIfMissing flag.
func (e Event) OnlyIfMissing() bool {
	return e.firstValue(datastar.OnlyIfMissingDatalineKey) == "true"
}

// ScriptContent extracts the JavaScript source from a script-bearing patch.
// Script patches (ExecuteScript, Redirect, ConsoleLog, ConsoleError,
// DispatchCustomEvent, ReplaceURL, Prefetch) wrap JS inside <script> tags
// within a patch-elements event. This method strips the <script ...> wrapper
// and returns the inner source code.
//
// Returns empty string if the event is not a script-bearing elements patch.
func (e Event) ScriptContent() string {
	el := e.Elements()

	afterTag, ok := strings.CutPrefix(el, "<script")
	if !ok {
		return ""
	}

	idx := indexTagEnd(afterTag)
	if idx < 0 {
		return ""
	}

	content := afterTag[idx+1:]

	if inner, ok := strings.CutSuffix(content, "</script>"); ok {
		return inner
	}

	return content
}

// DataValue returns the value after the first dataline matching the given key
// prefix (e.g., "selector ", "mode "). This is a generic escape hatch when no
// typed accessor covers a specific dataline key. Returns empty if not found.
func (e Event) DataValue(key string) string {
	return e.firstValue(key)
}

// String returns a human-readable debug representation of the event, showing
// the type, event ID (if any), retry (if non-zero), and dataline count.
// Useful for debugging test failures and logging.
func (e Event) String() string {
	if e.ID != "" || e.Retry > 0 {
		return fmt.Sprintf(
			"Event{type=%s id=%s retry=%d datalines=%d}",
			e.Type, e.ID, e.Retry, len(e.DataLines),
		)
	}

	return fmt.Sprintf("Event{type=%s datalines=%d}", e.Type, len(e.DataLines))
}

// EventsString returns a multi-line debug representation of an event slice,
// with one Event per line. Useful for logging test failures involving
// multiple events.
func EventsString(events []Event) string {
	if len(events) == 0 {
		return "(no events)"
	}

	var b strings.Builder

	for i, evt := range events {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(evt.String())
	}

	return b.String()
}

// indexTagEnd finds the index of the closing '>' of an HTML opening tag,
// skipping '>' characters that appear inside quoted attribute values.
// Returns -1 if no closing '>' is found.
func indexTagEnd(s string) int {
	var quote byte

	for i := range len(s) {
		char := s[i]

		if quote != 0 {
			if char == quote {
				quote = 0
			}

			continue
		}

		if char == '"' || char == '\'' {
			quote = char

			continue
		}

		if char == '>' {
			return i
		}
	}

	return -1
}

// firstValue returns the value after the first dataline matching the given key
// prefix. The key includes the trailing space (e.g., "selector ").
func (e Event) firstValue(key string) string {
	for _, line := range e.DataLines {
		if val, ok := strings.CutPrefix(line, key); ok {
			return val
		}
	}

	return ""
}

// allValues returns every value from datalines matching the given key prefix,
// preserving wire order. Used for multi-line fields (elements, signals).
func (e Event) allValues(key string) []string {
	var vals []string

	for _, line := range e.DataLines {
		if val, ok := strings.CutPrefix(line, key); ok {
			vals = append(vals, val)
		}
	}

	return vals
}
