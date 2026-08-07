package datastar_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/larsartmann/go-datastar"
)

func TestSignalsPatch_Basic(t *testing.T) {
	t.Parallel()

	patch := datastar.SignalsPatch{
		Signals: []byte(`{"count":42}`),
	}
	got := patch.Event()

	if got.Event != "datastar-patch-signals" {
		t.Errorf("Event: got %q, want %q", got.Event, "datastar-patch-signals")
	}

	wantData := "signals " + `{"count":42}`
	if got.Data != wantData {
		t.Errorf("Data: got %q, want %q", got.Data, wantData)
	}
}

func TestSignalsPatch_MultiLineJSON(t *testing.T) {
	t.Parallel()

	jsonPayload := []byte("{\n  \"count\": 42,\n  \"name\": \"test\"\n}")
	patch := datastar.SignalsPatch{Signals: jsonPayload}
	got := patch.Event()

	wantData := "signals {\nsignals   \"count\": 42,\nsignals   \"name\": \"test\"\nsignals }"
	if got.Data != wantData {
		t.Errorf("Data: got %q, want %q", got.Data, wantData)
	}
}

func TestSignalsPatch_OnlyIfMissing(t *testing.T) {
	t.Parallel()

	patch := datastar.SignalsPatch{
		Signals:       []byte(`{"x":1}`),
		OnlyIfMissing: true,
	}
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte("onlyIfMissing true\n")) {
		t.Errorf("should contain 'onlyIfMissing true'; got %q", got.Data)
	}
	if !bytes.Contains([]byte(got.Data), []byte("signals {\"x\":1}")) {
		t.Errorf("should contain 'signals {\"x\":1}'; got %q", got.Data)
	}
}

func TestSignalsPatch_EventID(t *testing.T) {
	t.Parallel()

	patch := datastar.SignalsPatch{
		Signals: []byte(`{}`),
		EventID: "sig-1",
	}
	got := patch.Event()

	if got.ID.Get() != "sig-1" {
		t.Errorf("ID: got %q, want %q", got.ID.Get(), "sig-1")
	}
}

func TestSignalsPatch_RetryDuration(t *testing.T) {
	t.Parallel()

	t.Run("custom", func(t *testing.T) {
		t.Parallel()

		patch := datastar.SignalsPatch{
			Signals:       []byte(`{}`),
			RetryDuration: 500 * time.Millisecond,
		}
		got := patch.Event()

		if got.Retry != 500 {
			t.Errorf("Retry: got %d, want 500", got.Retry)
		}
	})

	t.Run("default not emitted", func(t *testing.T) {
		t.Parallel()

		patch := datastar.SignalsPatch{
			Signals:       []byte(`{}`),
			RetryDuration: datastar.DefaultRetryDuration,
		}
		got := patch.Event()

		if got.Retry != 0 {
			t.Errorf("Retry: got %d, want 0", got.Retry)
		}
	})
}

func TestNewSignalsPatch(t *testing.T) {
	t.Parallel()

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewSignalsPatch(map[string]any{"count": 42})
		if err != nil {
			t.Fatalf("NewSignalsPatch: %v", err)
		}

		got := patch.Event()
		if !bytes.Contains([]byte(got.Data), []byte(`signals {"count":42}`)) {
			t.Errorf("should contain signals data; got %q", got.Data)
		}
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		type signals struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		patch, err := datastar.NewSignalsPatch(signals{Name: "test", Count: 5})
		if err != nil {
			t.Fatalf("NewSignalsPatch: %v", err)
		}

		got := patch.Event()
		if !bytes.Contains([]byte(got.Data), []byte(`"name":"test"`)) {
			t.Errorf("should contain name field; got %q", got.Data)
		}
		if !bytes.Contains([]byte(got.Data), []byte(`"count":5`)) {
			t.Errorf("should contain count field; got %q", got.Data)
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.NewSignalsPatch(make(chan int))
		if err == nil {
			t.Fatal("expected error for unmarshallable type")
		}
	})
}

func TestNewSignalsIfMissingPatch(t *testing.T) {
	t.Parallel()

	patch, err := datastar.NewSignalsIfMissingPatch(map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("NewSignalsIfMissingPatch: %v", err)
	}

	if !patch.OnlyIfMissing {
		t.Error("OnlyIfMissing should be true")
	}
}

func TestMarshalSignals(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		b, err := datastar.MarshalSignals(map[string]int{"a": 1})
		if err != nil {
			t.Fatalf("MarshalSignals: %v", err)
		}
		if string(b) != `{"a":1}` {
			t.Errorf("got %q, want %q", b, `{"a":1}`)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.MarshalSignals(make(chan int))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSignalsPatch_ImplementsPatch(t *testing.T) {
	t.Parallel()

	var _ datastar.Patch = datastar.SignalsPatch{}
}

func TestSignalsPatch_FullWireFormat(t *testing.T) {
	t.Parallel()

	patch := datastar.SignalsPatch{
		Signals:       []byte(`{"count":42}`),
		EventID:       "sig-5",
		RetryDuration: 3 * time.Second,
	}

	wire := writeEvent(t, patch)

	expectedLines := []string{
		"event: datastar-patch-signals\n",
		"data: signals {\"count\":42}\n",
		"id: sig-5\n",
		"retry: 3000\n",
	}

	for _, line := range expectedLines {
		if !bytes.Contains([]byte(wire), []byte(line)) {
			t.Errorf("wire format missing %q; got:\n%s", line, wire)
		}
	}

	if wire[len(wire)-2:] != "\n\n" {
		t.Errorf("wire must end with \\n\\n; got tail %q", wire[len(wire)-4:])
	}
}

// Verify SignalsPatch satisfies sse.Event compatibility
func TestSignalsPatch_EventIsSSEEvent(t *testing.T) {
	t.Parallel()

	patch := datastar.SignalsPatch{Signals: []byte(`{}`)}
	evt := patch.Event()
	_ = evt // verify Event() returns sse.Event
}
