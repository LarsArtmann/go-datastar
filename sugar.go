package datastar

import (
	"slices"

	errorfamily "github.com/larsartmann/go-error-family"
)

// NewRemovePatch creates an [ElementsPatch] that removes the element matching
// the given CSS selector from the DOM. It is equivalent to:
//
//	NewElementsPatch("", WithModeRemove(), WithSelector(selector))
func NewRemovePatch(selector string, opts ...ElementPatchOption) ElementsPatch {
	allOpts := append([]ElementPatchOption{WithModeRemove(), WithSelector(selector)}, opts...)

	return NewElementsPatch("", allOpts...)
}

// NewRemoveByIDPatch creates an [ElementsPatch] that removes the element with
// the given ID. Equivalent to NewRemovePatch("#" + id).
func NewRemoveByIDPatch(id string, opts ...ElementPatchOption) ElementsPatch {
	return NewRemovePatch("#"+id, opts...)
}

// --- Sugar option helpers ---

// WithModeOuter creates an option that uses the outer merge mode (default).
func WithModeOuter() ElementPatchOption { return WithMode(ElementPatchModeOuter) }

// WithModeInner creates an option that replaces the inner HTML of the target.
func WithModeInner() ElementPatchOption { return WithMode(ElementPatchModeInner) }

// WithModeRemove creates an option that removes the target element.
func WithModeRemove() ElementPatchOption { return WithMode(ElementPatchModeRemove) }

// WithModeReplace creates an option that replaces the target element without morphing.
func WithModeReplace() ElementPatchOption { return WithMode(ElementPatchModeReplace) }

// WithModePrepend creates an option that prepends inside the target element.
func WithModePrepend() ElementPatchOption { return WithMode(ElementPatchModePrepend) }

// WithModeAppend creates an option that appends inside the target element.
func WithModeAppend() ElementPatchOption { return WithMode(ElementPatchModeAppend) }

// WithModeBefore creates an option that inserts before the target element.
func WithModeBefore() ElementPatchOption { return WithMode(ElementPatchModeBefore) }

// WithModeAfter creates an option that inserts after the target element.
func WithModeAfter() ElementPatchOption { return WithMode(ElementPatchModeAfter) }

// WithSelectorID is a convenience for WithSelector("#" + id).
func WithSelectorID(id string) ElementPatchOption { return WithSelector("#" + id) }

// WithNamespaceHTML sets the namespace to HTML (the default — usually a no-op).
func WithNamespaceHTML() ElementPatchOption { return WithNamespace(NamespaceHTML) }

// WithNamespaceSVG sets the namespace to SVG.
func WithNamespaceSVG() ElementPatchOption { return WithNamespace(NamespaceSVG) }

// WithNamespaceMathML sets the namespace to MathML.
func WithNamespaceMathML() ElementPatchOption { return WithNamespace(NamespaceMathML) }

// WithViewTransitionsEnabled is shorthand for WithViewTransitions(true).
func WithViewTransitionsEnabled() ElementPatchOption { return WithViewTransitions(true) }

// WithoutViewTransitions is shorthand for WithViewTransitions(false).
func WithoutViewTransitions() ElementPatchOption { return WithViewTransitions(false) }

// --- Validation helpers ---

// ValidElementPatchModes lists all valid element patch modes. It is a public
// package-level slice so callers can iterate the set without depending on the
// private definition. Treated as immutable.
//
//nolint:gochecknoglobals // Public API; iterated by callers.
var ValidElementPatchModes = []ElementPatchMode{
	ElementPatchModeOuter,
	ElementPatchModeInner,
	ElementPatchModeRemove,
	ElementPatchModePrepend,
	ElementPatchModeAppend,
	ElementPatchModeBefore,
	ElementPatchModeAfter,
	ElementPatchModeReplace,
}

// ValidNamespaces lists all valid namespaces. It is a public package-level
// slice so callers can iterate the set without depending on the private
// definition. Treated as immutable.
//
//nolint:gochecknoglobals // Public API; iterated by callers.
var ValidNamespaces = []Namespace{
	NamespaceHTML,
	NamespaceSVG,
	NamespaceMathML,
}

// ElementPatchModeFromString converts a string to an [ElementPatchMode].
// Returns an error for invalid mode strings.
func ElementPatchModeFromString(modeStr string) (ElementPatchMode, error) {
	i := slices.IndexFunc(
		ValidElementPatchModes,
		func(m ElementPatchMode) bool { return string(m) == modeStr },
	)
	if i < 0 {
		return "", errorfamily.Newf(errorfamily.Rejection,
			CodeElementPatchModeInvalid, "unrecognized element patch mode %q", modeStr)
	}

	return ValidElementPatchModes[i], nil
}

// NamespaceFromString converts a string to a [Namespace].
// Returns an error for invalid namespace strings.
func NamespaceFromString(nsStr string) (Namespace, error) {
	i := slices.IndexFunc(ValidNamespaces, func(ns Namespace) bool { return string(ns) == nsStr })
	if i < 0 {
		return "", errorfamily.Newf(errorfamily.Rejection,
			CodeNamespaceInvalid, "unrecognized namespace %q", nsStr)
	}

	return ValidNamespaces[i], nil
}
