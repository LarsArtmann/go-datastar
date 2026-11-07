package datastar

// applyOptions folds functional options onto base in order, so later options
// win. Every patch constructor in this package folds its options through here,
// giving the apply semantics a single home.
func applyOptions[T any, O ~func(*T)](base T, opts []O) T {
	for _, opt := range opts {
		opt(&base)
	}

	return base
}
