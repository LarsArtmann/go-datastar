package datastar_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func TestScriptPatch_DefaultAutoRemove(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("console.log('hi')")
	got := patch.Event()

	// Should use patch-elements event with selector body, mode append
	if got.Event != "datastar-patch-elements" {
		t.Errorf("Event: got %q, want datastar-patch-elements", got.Event)
	}

	if !strings.Contains(got.Data, "selector body") {
		t.Errorf("should contain 'selector body'; got %q", got.Data)
	}

	if !strings.Contains(got.Data, "mode append") {
		t.Errorf("should contain 'mode append'; got %q", got.Data)
	}
	// nil AutoRemove → data-effect="el.remove()"
	if !strings.Contains(got.Data, `data-effect="el.remove()"`) {
		t.Errorf("should contain auto-remove effect; got %q", got.Data)
	}
	// Should contain the script wrapped in <script>
	wantScript := `<script data-effect="el.remove()">console.log('hi')</script>`
	if !strings.Contains(got.Data, wantScript) {
		t.Errorf("should contain wrapped script; got %q", got.Data)
	}
}

func TestScriptPatch_AutoRemoveTrue(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x", datastar.WithScriptAutoRemove(true))
	got := patch.Event()

	if !strings.Contains(got.Data, `data-effect="el.remove()"`) {
		t.Errorf("should contain auto-remove; got %q", got.Data)
	}
}

func TestScriptPatch_AutoRemoveFalse(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x", datastar.WithScriptAutoRemove(false))
	got := patch.Event()

	if strings.Contains(got.Data, `data-effect="el.remove()"`) {
		t.Errorf("should NOT contain auto-remove; got %q", got.Data)
	}
}

func TestScriptPatch_WithAttributes(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x",
		datastar.WithScriptAutoRemove(false),
		datastar.WithScriptAttributes(`type="module"`),
	)
	got := patch.Event()

	if !strings.Contains(got.Data, `<script type="module">x</script>`) {
		t.Errorf("should contain script with type attribute; got %q", got.Data)
	}
}

func TestScriptPatch_WithAttributeKVs(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x",
		datastar.WithScriptAutoRemove(false),
		datastar.WithScriptAttributeKVs("type", "speculationrules"),
	)
	got := patch.Event()

	if !strings.Contains(got.Data, `<script type="speculationrules">x</script>`) {
		t.Errorf("should contain script with speculationrules type; got %q", got.Data)
	}
}

func TestScriptPatch_MultipleAttributes(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x",
		datastar.WithScriptAutoRemove(false),
		datastar.WithScriptAttributes(`type="module"`, `async`),
	)
	got := patch.Event()

	if !strings.Contains(got.Data, `<script type="module" async>x</script>`) {
		t.Errorf("should contain script with multiple attrs; got %q", got.Data)
	}
}

func TestScriptPatch_EventID(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("x",
		datastar.WithScriptEventID("script-1"),
	)
	got := patch.Event()

	if got.ID.Get() != "script-1" {
		t.Errorf("ID: got %q, want script-1", got.ID.Get())
	}
}

func TestScriptPatch_FullWireFormat(t *testing.T) {
	t.Parallel()

	patch := datastar.NewScriptPatch("console.log('hello')")

	wire := writeEvent(t, patch)

	expectedLines := []string{
		"event: datastar-patch-elements\n",
		"data: selector body\n",
		"data: mode append\n",
		`data: elements <script data-effect="el.remove()">console.log('hello')</script>` + "\n",
	}

	for _, line := range expectedLines {
		if !strings.Contains(wire, line) {
			t.Errorf("wire format missing %q; got:\n%s", line, wire)
		}
	}

	if wire[len(wire)-2:] != "\n\n" {
		t.Errorf("wire must end with \\n\\n; got tail %q", wire[len(wire)-4:])
	}
}

func TestScriptPatch_ImplementsPatch(t *testing.T) {
	t.Parallel()

	var (
		_ datastar.Patch = datastar.ScriptPatch{}
		_ datastar.Patch = datastar.NewScriptPatch("x")
	)
}
