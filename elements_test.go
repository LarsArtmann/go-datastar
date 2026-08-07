package datastar_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"
)

// writeEvent is a test helper that serializes a patch to wire format.
func writeEvent(t *testing.T, p datastar.Patch) string {
	t.Helper()

	var buf bytes.Buffer
	if err := sse.WriteEvent(&buf, p.Event()); err != nil {
		t.Fatalf("WriteEvent: %v", err)
	}

	return buf.String()
}

func TestElementsPatch_DefaultOuter(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div>hello</div>")
	got := patch.Event()

	want := sse.Event{
		Event: "datastar-patch-elements",
		Data:  "elements <div>hello</div>",
	}

	if got.Event != want.Event {
		t.Errorf("Event: got %q, want %q", got.Event, want.Event)
	}

	if got.Data != want.Data {
		t.Errorf("Data: got %q, want %q", got.Data, want.Data)
	}

	if !got.ID.IsZero() {
		t.Errorf("ID: got %q, want zero", got.ID.Get())
	}

	if got.Retry != 0 {
		t.Errorf("Retry: got %d, want 0 (default should not emit)", got.Retry)
	}
}

func TestElementsPatch_WithSelector(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div>hi</div>",
		datastar.WithSelector("#feed"),
	)
	got := patch.Event()

	wantData := "selector #feed\nelements <div>hi</div>"
	if got.Data != wantData {
		t.Errorf("Data: got %q, want %q", got.Data, wantData)
	}
}

func TestElementsPatch_WithSelectorf(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div>hi</div>",
		datastar.WithSelectorf("#item-%d", 42),
	)
	got := patch.Event()

	wantData := "selector #item-42\nelements <div>hi</div>"
	if got.Data != wantData {
		t.Errorf("Data: got %q, want %q", got.Data, wantData)
	}
}

func TestElementsPatch_AllModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode datastar.ElementPatchMode
		want string // expected mode line; empty = not emitted
	}{
		{"outer", datastar.ElementPatchModeOuter, ""},
		{"inner", datastar.ElementPatchModeInner, "mode inner"},
		{"remove", datastar.ElementPatchModeRemove, "mode remove"},
		{"replace", datastar.ElementPatchModeReplace, "mode replace"},
		{"prepend", datastar.ElementPatchModePrepend, "mode prepend"},
		{"append", datastar.ElementPatchModeAppend, "mode append"},
		{"before", datastar.ElementPatchModeBefore, "mode before"},
		{"after", datastar.ElementPatchModeAfter, "mode after"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			patch := datastar.NewElementsPatch("<div>x</div>",
				datastar.WithMode(testCase.mode),
			)
			got := patch.Event()

			if testCase.want == "" {
				if strings.Contains(got.Data, "mode ") {
					t.Errorf("mode should not be emitted for outer; Data=%q", got.Data)
				}
			} else {
				if !strings.Contains(got.Data, testCase.want) {
					t.Errorf("Data should contain %q; got %q", testCase.want, got.Data)
				}
			}
		})
	}
}

func TestElementsPatch_Namespace(t *testing.T) {
	t.Parallel()

	t.Run("svg emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<circle/>",
			datastar.WithNamespace(datastar.NamespaceSVG),
		)
		got := patch.Event()

		if !strings.Contains(got.Data, "namespace svg") {
			t.Errorf("should contain 'namespace svg'; got %q", got.Data)
		}
	})

	t.Run("html not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithNamespace(datastar.NamespaceHTML),
		)
		got := patch.Event()

		if strings.Contains(got.Data, "namespace ") {
			t.Errorf("should NOT contain 'namespace'; got %q", got.Data)
		}
	})

	t.Run("mathml emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<mi>x</mi>",
			datastar.WithNamespace(datastar.NamespaceMathML),
		)
		got := patch.Event()

		if !strings.Contains(got.Data, "namespace mathml") {
			t.Errorf("should contain 'namespace mathml'; got %q", got.Data)
		}
	})
}

func TestElementsPatch_ViewTransitions(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithViewTransitions(true),
		)
		got := patch.Event()

		if !strings.Contains(got.Data, "useViewTransition true") {
			t.Errorf("should contain 'useViewTransition true'; got %q", got.Data)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithViewTransitions(false),
		)
		got := patch.Event()

		if strings.Contains(got.Data, "useViewTransition") {
			t.Errorf("should NOT contain 'useViewTransition'; got %q", got.Data)
		}
	})

	t.Run("with selector", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithViewTransitions(true),
			datastar.WithViewTransitionSelector("#main"),
		)
		got := patch.Event()

		if !strings.Contains(got.Data, "viewTransitionSelector #main") {
			t.Errorf("should contain 'viewTransitionSelector #main'; got %q", got.Data)
		}
	})
}

func TestElementsPatch_EventID(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div/>",
		datastar.WithElementsEventID("evt-42"),
	)
	got := patch.Event()

	if got.ID.Get() != "evt-42" {
		t.Errorf("ID: got %q, want %q", got.ID.Get(), "evt-42")
	}
}

func TestElementsPatch_RetryDuration(t *testing.T) {
	t.Parallel()

	t.Run("custom retry emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithElementsRetryDuration(2000*time.Millisecond),
		)
		got := patch.Event()

		if got.Retry != 2000 {
			t.Errorf("Retry: got %d, want 2000", got.Retry)
		}
	})

	t.Run("default retry not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>")
		got := patch.Event()

		if got.Retry != 0 {
			t.Errorf("Retry: got %d, want 0 (default should not emit)", got.Retry)
		}
	})

	t.Run("zero retry not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.NewElementsPatch("<div/>",
			datastar.WithElementsRetryDuration(0),
		)
		got := patch.Event()

		if got.Retry != 0 {
			t.Errorf("Retry: got %d, want 0", got.Retry)
		}
	})
}

func TestElementsPatch_MultiLineHTML(t *testing.T) {
	t.Parallel()

	html := "<div>\n  <span>hi</span>\n</div>"
	patch := datastar.NewElementsPatch(html)
	got := patch.Event()

	wantData := "elements <div>\nelements   <span>hi</span>\nelements </div>"
	if got.Data != wantData {
		t.Errorf("Data: got %q, want %q", got.Data, wantData)
	}
}

func TestElementsPatch_EmptyHTML(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("")
	got := patch.Event()

	// No elements data lines should be present
	if got.Data != "" {
		t.Errorf("Data: got %q, want empty", got.Data)
	}
}

func TestElementsPatch_FullWireFormat(t *testing.T) {
	t.Parallel()

	patch := datastar.NewElementsPatch("<div id=\"feed\">\n  <span>1</span>\n</div>",
		datastar.WithSelector("#feed"),
		datastar.WithMode(datastar.ElementPatchModeInner),
		datastar.WithElementsEventID("evt-99"),
	)

	wire := writeEvent(t, patch)

	// The wire format must contain these lines in this exact order
	expectedLines := []string{
		"event: datastar-patch-elements\n",
		"data: selector #feed\n",
		"data: mode inner\n",
		"data: elements <div id=\"feed\">\n",
		"data: elements   <span>1</span>\n",
		"data: elements </div>\n",
		"id: evt-99\n",
	}

	for _, line := range expectedLines {
		if !strings.Contains(wire, line) {
			t.Errorf("wire format missing %q; got:\n%s", line, wire)
		}
	}

	// Must end with double newline (event separator)
	if wire[len(wire)-2:] != "\n\n" {
		t.Errorf("wire must end with \\n\\n; got tail %q", wire[len(wire)-4:])
	}
}

func TestElementsPatch_ImplementsPatch(t *testing.T) {
	t.Parallel()

	var (
		_ datastar.Patch = datastar.ElementsPatch{}
		_ datastar.Patch = datastar.NewElementsPatch("<div/>")
	)
}
