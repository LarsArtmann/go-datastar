package datastar_test

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func TestNewRedirectPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRedirectPatch("https://example.com")
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte("elements <script data-effect=\"el.remove()\">setTimeout(() => window.location.href = \"https://example.com\")</script>")) {
		t.Errorf("should contain redirect script; got %q", got.Data)
	}
}

func TestNewRedirectfPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRedirectfPatch("/users/%d", 42)
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte(`window.location.href = "/users/42"`)) {
		t.Errorf("should contain formatted URL; got %q", got.Data)
	}
}

func TestNewConsoleLogPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleLogPatch("hello world")
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte(`console.log("hello world")`)) {
		t.Errorf("should contain console.log; got %q", got.Data)
	}
}

func TestNewConsoleLogfPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleLogfPatch("count: %d", 5)
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte(`console.log("count: 5")`)) {
		t.Errorf("should contain formatted console.log; got %q", got.Data)
	}
}

func TestNewConsoleErrorPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleErrorPatch(errTest("something broke"))
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte(`console.error("something broke")`)) {
		t.Errorf("should contain console.error; got %q", got.Data)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestNewDispatchCustomEventPatch(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("myEvent", map[string]any{"key": "val"})
		if err != nil {
			t.Fatalf("NewDispatchCustomEventPatch: %v", err)
		}
		got := patch.Event()
		if !bytes.Contains([]byte(got.Data), []byte(`new CustomEvent("myEvent"`)) {
			t.Errorf("should contain CustomEvent construction; got %q", got.Data)
		}
		if !bytes.Contains([]byte(got.Data), []byte(`"key":"val"`)) {
			t.Errorf("should contain detail JSON; got %q", got.Data)
		}
		if !bytes.Contains([]byte(got.Data), []byte(`elements = [document]`)) {
			t.Errorf("should use [document] for default selector; got %q", got.Data)
		}
	})

	t.Run("custom selector", func(t *testing.T) {
		t.Parallel()

		patch, _ := datastar.NewDispatchCustomEventPatch("evt", nil,
			datastar.WithCustomEventSelector("#my-element"),
		)
		got := patch.Event()

		if !bytes.Contains([]byte(got.Data), []byte(`document.querySelectorAll("#my-element")`)) {
			t.Errorf("should use querySelectorAll; got %q", got.Data)
		}
	})

	t.Run("empty event name returns error", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.NewDispatchCustomEventPatch("", nil)
		if err == nil {
			t.Fatal("expected error for empty event name")
		}
	})

	t.Run("default flags", func(t *testing.T) {
		t.Parallel()

		patch, _ := datastar.NewDispatchCustomEventPatch("evt", nil)
		got := patch.Event()

		for _, want := range []string{"bubbles: true", "cancelable: true", "composed: true"} {
			if !bytes.Contains([]byte(got.Data), []byte(want)) {
				t.Errorf("should contain %q; got %q", want, got.Data)
			}
		}
	})
}

func TestNewReplaceURLPatch(t *testing.T) {
	t.Parallel()

	u := url.URL{Scheme: "https", Host: "example.com", Path: "/new"}
	patch := datastar.NewReplaceURLPatch(u)
	got := patch.Event()

	if !bytes.Contains([]byte(got.Data), []byte(`window.history.replaceState({}, "", "https://example.com/new")`)) {
		t.Errorf("should contain replaceState; got %q", got.Data)
	}
}

func TestNewPrefetchPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewPrefetchPatch("/page1", "/page2")
	got := patch.Event()

	// Should have type="speculationrules" attribute
	if !bytes.Contains([]byte(got.Data), []byte(`type="speculationrules"`)) {
		t.Errorf("should contain speculationrules type; got %q", got.Data)
	}
	// Should NOT have auto-remove (false)
	if bytes.Contains([]byte(got.Data), []byte(`data-effect="el.remove()"`)) {
		t.Errorf("should NOT contain auto-remove; got %q", got.Data)
	}
	// Should contain both URLs
	if !bytes.Contains([]byte(got.Data), []byte(`"/page1"`)) {
		t.Errorf("should contain /page1; got %q", got.Data)
	}
	if !bytes.Contains([]byte(got.Data), []byte(`"/page2"`)) {
		t.Errorf("should contain /page2; got %q", got.Data)
	}
	// Should contain prefetch source
	if !bytes.Contains([]byte(got.Data), []byte(`"source": "list"`)) {
		t.Errorf("should contain source list; got %q", got.Data)
	}
}

func TestDispatchCustomEventPatch_ImplementsPatch(t *testing.T) {
	t.Parallel()

	var _ datastar.Patch = datastar.DispatchCustomEventPatch{}
}
