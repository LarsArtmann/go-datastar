package datastartest

import (
	"encoding/json/v2"
	"strconv"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Typed accessors over script-bearing events. They parse the JavaScript the
// library itself emits (the sugar patches in go-datastar's
// script_convenience.go), so a test can assert intent ("redirected to /next")
// instead of exact script text. All of them return zero values when the event
// is not a script or does not match the known emission shape — assert on the
// raw [Event.ScriptContent] for anything more exotic.

// RedirectURL extracts the target URL from a redirect script patch
// (emitted as `setTimeout(() => window.location.href = %q)`). Returns empty
// if the event carries no redirect of that shape.
func (e Event) RedirectURL() string {
	script := e.ScriptContent()

	const marker = "window.location.href = "

	return quotedSuffix(script, marker)
}

// CustomEventName extracts the event name from a custom-event dispatch
// script patch (`new CustomEvent(%q, ...)`). Returns empty if not found.
func (e Event) CustomEventName() string {
	script := e.ScriptContent()

	_, name, found := strings.Cut(script, "new CustomEvent(")
	if !found {
		return ""
	}

	return unquotePrefix(name)
}

// CustomEventDetail returns the raw JSON detail expression from a
// custom-event dispatch script patch (the value after `detail: ` up to the
// end of its emission line). Returns empty if not found.
func (e Event) CustomEventDetail() string {
	script := e.ScriptContent()

	const marker = "detail: "

	_, rest, found := strings.Cut(script, marker)
	if !found {
		return ""
	}

	lineEnd := strings.IndexByte(rest, '\n')
	if lineEnd >= 0 {
		rest = rest[:lineEnd]
	}

	return strings.TrimRight(rest, " \t,")
}

// UnmarshalCustomEventDetail decodes the raw custom-event detail JSON into
// target. Malformed or missing detail yields a classified
// `datastartest.custom_event_detail_unmarshal_failed` error.
func (e Event) UnmarshalCustomEventDetail(target any) error {
	raw := e.CustomEventDetail()
	if raw == "" {
		return errorfamily.New(errorfamily.Rejection,
			CodeCustomEventDetailUnmarshalFailed,
			"custom event detail not found in script content").
			WithContext("eventType", e.Type)
	}

	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return errorfamily.Wrapf(err, errorfamily.Rejection,
			CodeCustomEventDetailUnmarshalFailed,
			"unmarshal custom event detail").
			WithContext("detailLength", strconv.Itoa(len(raw)))
	}

	return nil
}

// ScriptAttributes returns the attributes on the injected `<script>` tag
// (e.g. `type="speculationrules"`, `data-effect="el.remove()"`). Values keep
// their quotes. Returns nil for non-script events.
func (e Event) ScriptAttributes() []string {
	el := e.Elements()

	afterTag, ok := strings.CutPrefix(el, "<script")
	if !ok {
		return nil
	}

	idx := indexTagEnd(afterTag)
	if idx < 0 {
		return nil
	}

	return strings.Fields(afterTag[:idx])
}

// quotedSuffix finds marker in s and unquotes the double-quoted string that
// immediately follows it (the library emits Go %q-quoted JS string literals).
// Returns empty if the marker or a well-formed quoted string is missing.
func quotedSuffix(s, marker string) string {
	_, rest, found := strings.Cut(s, marker)
	if !found {
		return ""
	}

	return unquotePrefix(rest)
}

// unquotePrefix unquotes the double-quoted string at the start of s (after
// leading spaces). Returns empty if s does not start with a well-formed
// quoted string.
func unquotePrefix(s string) string {
	rest := strings.TrimLeft(s, " \t")
	if !strings.HasPrefix(rest, `"`) {
		return ""
	}

	quoted, err := strconv.QuotedPrefix(rest)
	if err != nil {
		return ""
	}

	unquoted, err := strconv.Unquote(quoted)
	if err != nil {
		return ""
	}

	return unquoted
}
