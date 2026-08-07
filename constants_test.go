package datastar_test

import (
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
)

func TestConstants_RetryDuration(t *testing.T) {
	t.Parallel()

	if datastar.DefaultRetryDuration != 1000*time.Millisecond {
		t.Errorf("DefaultRetryDuration = %v, want 1000ms", datastar.DefaultRetryDuration)
	}
}

func TestConstants_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value datastar.EventType
		want  string
	}{
		{"patch elements", datastar.EventTypePatchElements, "datastar-patch-elements"},
		{"patch signals", datastar.EventTypePatchSignals, "datastar-patch-signals"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if string(tc.value) != tc.want {
				t.Errorf("got %q, want %q", tc.value, tc.want)
			}
		})
	}
}

func TestConstants_ElementPatchModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value datastar.ElementPatchMode
		want  string
	}{
		{"default", datastar.DefaultElementPatchMode, "outer"},
		{"outer", datastar.ElementPatchModeOuter, "outer"},
		{"inner", datastar.ElementPatchModeInner, "inner"},
		{"remove", datastar.ElementPatchModeRemove, "remove"},
		{"replace", datastar.ElementPatchModeReplace, "replace"},
		{"prepend", datastar.ElementPatchModePrepend, "prepend"},
		{"append", datastar.ElementPatchModeAppend, "append"},
		{"before", datastar.ElementPatchModeBefore, "before"},
		{"after", datastar.ElementPatchModeAfter, "after"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if string(tc.value) != tc.want {
				t.Errorf("got %q, want %q", tc.value, tc.want)
			}
		})
	}
}

func TestConstants_Namespaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value datastar.Namespace
		want  string
	}{
		{"html", datastar.NamespaceHTML, "html"},
		{"svg", datastar.NamespaceSVG, "svg"},
		{"mathml", datastar.NamespaceMathML, "mathml"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if string(tc.value) != tc.want {
				t.Errorf("got %q, want %q", tc.value, tc.want)
			}
		})
	}
}

func TestConstants_DatalineKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"selector", datastar.SelectorDatalineKey, "selector "},
		{"mode", datastar.ModeDatalineKey, "mode "},
		{"namespace", datastar.NamespaceDatalineKey, "namespace "},
		{"useViewTransition", datastar.UseViewTransitionDatalineKey, "useViewTransition "},
		{"viewTransitionSelector", datastar.ViewTransitionSelectorDatalineKey, "viewTransitionSelector "},
		{"elements", datastar.ElementsDatalineKey, "elements "},
		{"signals", datastar.SignalsDatalineKey, "signals "},
		{"onlyIfMissing", datastar.OnlyIfMissingDatalineKey, "onlyIfMissing "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.value != tc.want {
				t.Errorf("got %q, want %q", tc.value, tc.want)
			}
		})
	}
}

func TestConstants_DefaultModeIsOuter(t *testing.T) {
	t.Parallel()

	if datastar.DefaultElementPatchMode != datastar.ElementPatchModeOuter {
		t.Errorf("DefaultElementPatchMode = %q, want %q",
			datastar.DefaultElementPatchMode, datastar.ElementPatchModeOuter)
	}
}

func TestConstants_DatastarKey(t *testing.T) {
	t.Parallel()

	if datastar.DatastarKey != "datastar" {
		t.Errorf("DatastarKey = %q, want %q", datastar.DatastarKey, "datastar")
	}
}
