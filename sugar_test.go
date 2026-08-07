package datastar_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func TestNewRemovePatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRemovePatch("#feed")
	got := patch.Event()

	if !strings.Contains(got.Data, "selector #feed") {
		t.Errorf("should contain selector; got %q", got.Data)
	}

	if !strings.Contains(got.Data, "mode remove") {
		t.Errorf("should contain mode remove; got %q", got.Data)
	}
	// Should NOT contain any elements data lines
	if strings.Contains(got.Data, "elements ") {
		t.Errorf("should NOT contain elements data lines; got %q", got.Data)
	}
}

func TestNewRemoveByIDPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRemoveByIDPatch("my-element")
	got := patch.Event()

	if !strings.Contains(got.Data, "selector #my-element") {
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			patch := datastar.NewElementsPatch("<div/>", testCase.opt)
			got := patch.Event()

			if testCase.want == "" {
				if strings.Contains(got.Data, "mode ") {
					t.Errorf("should not emit mode for outer; got %q", got.Data)
				}
			} else {
				if !strings.Contains(got.Data, testCase.want) {
					t.Errorf("should contain %q; got %q", testCase.want, got.Data)
				}
			}
		})
	}
}

func TestSugar_WithSelectorID(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div/>", datastar.WithSelectorID("main"))
	got := patch.Event()

	if !strings.Contains(got.Data, "selector #main") {
		t.Errorf("should contain selector #main; got %q", got.Data)
	}
}

func TestSugar_NamespaceHelpers(t *testing.T) {
	t.Parallel()

	t.Run("svg", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<circle/>", datastar.WithNamespaceSVG())
		got := patch.Event()

		if !strings.Contains(got.Data, "namespace svg") {
			t.Errorf("should contain namespace svg; got %q", got.Data)
		}
	})

	t.Run("mathml", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<mi/>", datastar.WithNamespaceMathML())
		got := patch.Event()

		if !strings.Contains(got.Data, "namespace mathml") {
			t.Errorf("should contain namespace mathml; got %q", got.Data)
		}
	})

	t.Run("html not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>", datastar.WithNamespaceHTML())
		got := patch.Event()

		if strings.Contains(got.Data, "namespace ") {
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

		if !strings.Contains(got.Data, "useViewTransition true") {
			t.Errorf("should contain viewTransition; got %q", got.Data)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>", datastar.WithoutViewTransitions())
		got := patch.Event()

		if strings.Contains(got.Data, "useViewTransition") {
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
		for _, modeStr := range valid {
			m, err := datastar.ElementPatchModeFromString(modeStr)
			if err != nil {
				t.Errorf("ElementPatchModeFromString(%q): %v", modeStr, err)
			}

			if string(m) != modeStr {
				t.Errorf("got %q, want %q", m, modeStr)
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
		for _, nsStr := range valid {
			namespace, err := datastar.NamespaceFromString(nsStr)
			if err != nil {
				t.Errorf("NamespaceFromString(%q): %v", nsStr, err)
			}

			if string(namespace) != nsStr {
				t.Errorf("got %q, want %q", namespace, nsStr)
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
