package datastar_test

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func TestNewRemovePatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRemovePatch("#feed")
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte("selector #feed")) {
		t.Errorf("should contain selector; got %q", got.Data)
	}

	if !bytes.Contains([]byte(got.Data), []byte("mode remove")) {
		t.Errorf("should contain mode remove; got %q", got.Data)
	}
	// Should NOT contain any elements data lines
	if bytes.Contains([]byte(got.Data), []byte("elements ")) {
		t.Errorf("should NOT contain elements data lines; got %q", got.Data)
	}
}

func TestNewRemoveByIDPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRemoveByIDPatch("my-element")
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte("selector #my-element")) {
		t.Errorf("should contain #my-element; got %q", got.Data)
	}
}

func TestSugar_ModeHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opt  datastar.ElementPatchOption
		want string
	}{
		{"outer", datastar.WithModeOuter(), ""},
		{"inner", datastar.WithModeInner(), "mode inner"},
		{"remove", datastar.WithModeRemove(), "mode remove"},
		{"replace", datastar.WithModeReplace(), "mode replace"},
		{"prepend", datastar.WithModePrepend(), "mode prepend"},
		{"append", datastar.WithModeAppend(), "mode append"},
		{"before", datastar.WithModeBefore(), "mode before"},
		{"after", datastar.WithModeAfter(), "mode after"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			patch := datastar.NewElementsPatch("<div/>", tc.opt)
			got := patch.Event()

			if tc.want == "" {
				if bytes.Contains([]byte(got.Data), []byte("mode ")) {
					t.Errorf("should not emit mode for outer; got %q", got.Data)
				}
			} else {
				if !bytes.Contains([]byte(got.Data), []byte(tc.want)) {
					t.Errorf("should contain %q; got %q", tc.want, got.Data)
				}
			}
		})
	}
}

func TestSugar_WithSelectorID(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div/>", datastar.WithSelectorID("main"))
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte("selector #main")) {
		t.Errorf("should contain selector #main; got %q", got.Data)
	}
}

func TestSugar_NamespaceHelpers(t *testing.T) {
	t.Parallel()

	t.Run("svg", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<circle/>", datastar.WithNamespaceSVG())
		got := patch.Event()

		if !bytes.Contains([]byte(got.Data), []byte("namespace svg")) {
			t.Errorf("should contain namespace svg; got %q", got.Data)
		}
	})

	t.Run("mathml", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<mi/>", datastar.WithNamespaceMathML())
		got := patch.Event()

		if !bytes.Contains([]byte(got.Data), []byte("namespace mathml")) {
			t.Errorf("should contain namespace mathml; got %q", got.Data)
		}
	})

	t.Run("html not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>", datastar.WithNamespaceHTML())
		got := patch.Event()

		if bytes.Contains([]byte(got.Data), []byte("namespace ")) {
			t.Errorf("should NOT contain namespace; got %q", got.Data)
		}
	})
}

func TestSugar_ViewTransitionsHelpers(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>", datastar.WithViewTransitionsEnabled())
		got := patch.Event()

		if !bytes.Contains([]byte(got.Data), []byte("useViewTransition true")) {
			t.Errorf("should contain viewTransition; got %q", got.Data)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>", datastar.WithoutViewTransitions())
		got := patch.Event()

		if bytes.Contains([]byte(got.Data), []byte("useViewTransition")) {
			t.Errorf("should NOT contain viewTransition; got %q", got.Data)
		}
	})
}

func TestValidation_ElementPatchModeFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		valid := []string{
			"outer",
			"inner",
			"remove",
			"replace",
			"prepend",
			"append",
			"before",
			"after",
		}
		for _, s := range valid {
			m, err := datastar.ElementPatchModeFromString(s)
			if err != nil {
				t.Errorf("ElementPatchModeFromString(%q): %v", s, err)
			}

			if string(m) != s {
				t.Errorf("got %q, want %q", m, s)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.ElementPatchModeFromString("bogus")
		if err == nil {
			t.Fatal("expected error for invalid mode")
		}
	})
}

func TestValidation_NamespaceFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		valid := []string{"html", "svg", "mathml"}
		for _, s := range valid {
			ns, err := datastar.NamespaceFromString(s)
			if err != nil {
				t.Errorf("NamespaceFromString(%q): %v", s, err)
			}

			if string(ns) != s {
				t.Errorf("got %q, want %q", ns, s)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.NamespaceFromString("bogus")
		if err == nil {
			t.Fatal("expected error for invalid namespace")
		}
	})
}

func TestValidation_ValidLists(t *testing.T) {
	t.Parallel()

	if len(datastar.ValidElementPatchModes) != 8 {
		t.Errorf("ValidElementPatchModes: got %d, want 8", len(datastar.ValidElementPatchModes))
	}

	if len(datastar.ValidNamespaces) != 3 {
		t.Errorf("ValidNamespaces: got %d, want 3", len(datastar.ValidNamespaces))
	}
}
