package datastar_test

import (
	"encoding/json/v2"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func BenchmarkElementsPatch_Event(b *testing.B) {
	html := `<div class="item"><span>Hello</span></div>`

	b.ResetTimer()

	for range b.N {
		patch := datastar.NewElementsPatch(html,
			datastar.WithSelector("#feed"),
			datastar.WithMode(datastar.ElementPatchModeInner),
		)

		_ = patch.Event()
	}
}

func BenchmarkSignalsPatch_Event(b *testing.B) {
	signals := map[string]any{
		"count":    42,
		"name":     "test",
		"active":   true,
		"tags":     []string{"a", "b", "c"},
		"metadata": map[string]any{"version": "1.0"},
	}

	b.ResetTimer()

	for range b.N {
		patch, _ := datastar.NewSignalsPatch(signals)

		_ = patch.Event()
	}
}

func BenchmarkScriptPatch_Event(b *testing.B) {
	script := `console.log("hello world")`

	b.ResetTimer()

	for range b.N {
		patch := datastar.NewScriptPatch(script,
			datastar.WithScriptAutoRemove(false),
			datastar.WithScriptAttributes(`type="module"`),
		)

		_ = patch.Event()
	}
}

func BenchmarkMarshalSignals(b *testing.B) {
	signals := map[string]any{
		"count":    42,
		"name":     "test",
		"active":   true,
		"tags":     []string{"a", "b", "c"},
		"metadata": map[string]any{"version": "1.0"},
	}

	b.ResetTimer()

	for range b.N {
		_, _ = datastar.MarshalSignals(signals)
	}
}

// FuzzMarshalSignalsRoundtrip hardens the outbound marshaling boundary. It takes
// arbitrary bytes, tries to unmarshal them as JSON into a generic value, then
// re-marshals via MarshalSignals. The invariant: MarshalSignals must never panic
// — it returns JSON bytes or a classified error.
func FuzzMarshalSignalsRoundtrip(f *testing.F) {
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(`{"nested":{"deep":{"value":42}}}`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(`123`))
	f.Add([]byte(`true`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		t.Parallel()

		var val any

		if err := json.Unmarshal(data, &val); err != nil {
			t.Skip() // not valid JSON — not interesting for this fuzz
		}

		_, _ = datastar.MarshalSignals(val)
	})
}
