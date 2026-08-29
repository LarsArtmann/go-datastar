package datastar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

func TestNewRedirectPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRedirectPatch("https://example.com")
	got := patch.Event()

	wantScript := `elements <script data-effect="el.remove()">setTimeout(() => ` +
		`window.location.href = "https://example.com")</script>`
	if !strings.Contains(got.Data, wantScript) {
		t.Errorf("should contain redirect script; got %q", got.Data)
	}
}

func TestNewRedirectfPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewRedirectfPatch("/users/%d", 42)
	got := patch.Event()

	if !strings.Contains(got.Data, `window.location.href = "/users/42"`) {
		t.Errorf("should contain formatted URL; got %q", got.Data)
	}
}

func TestNewConsoleLogPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleLogPatch("hello world")
	got := patch.Event()

	if !strings.Contains(got.Data, `console.log("hello world")`) {
		t.Errorf("should contain console.log; got %q", got.Data)
	}
}

func TestNewConsoleLogfPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleLogfPatch("count: %d", 5)
	got := patch.Event()

	if !strings.Contains(got.Data, `console.log("count: 5")`) {
		t.Errorf("should contain formatted console.log; got %q", got.Data)
	}
}

func TestNewConsoleErrorPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewConsoleErrorPatch(fakeError("something broke"))
	got := patch.Event()

	if !strings.Contains(got.Data, `console.error("something broke")`) {
		t.Errorf("should contain console.error; got %q", got.Data)
	}
}

type fakeError string

func (e fakeError) Error() string { return string(e) }

func TestNewDispatchCustomEventPatch(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.NewDispatchCustomEventPatch("myEvent", map[string]any{"key": "val"})
		if err != nil {
			t.Fatalf("NewDispatchCustomEventPatch: %v", err)
		}

		got := patch.Event()
		if !strings.Contains(got.Data, `new CustomEvent("myEvent"`) {
			t.Errorf("should contain CustomEvent construction; got %q", got.Data)
		}

		if !strings.Contains(got.Data, `"key":"val"`) {
			t.Errorf("should contain detail JSON; got %q", got.Data)
		}

		if !strings.Contains(got.Data, `elements = [document]`) {
			t.Errorf("should use [document] for default selector; got %q", got.Data)
		}
	})

	t.Run("custom selector", func(t *testing.T) {
		t.Parallel()

		patch, _ := datastar.NewDispatchCustomEventPatch("evt", nil,
			datastar.WithCustomEventSelector("#my-element"),
		)
		got := patch.Event()

		if !strings.Contains(got.Data, `document.querySelectorAll("#my-element")`) {
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
			if !strings.Contains(got.Data, want) {
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

	if !strings.Contains(
		got.Data,
		`window.history.replaceState({}, "", "https://example.com/new")`,
	) {
		t.Errorf("should contain replaceState; got %q", got.Data)
	}
}

func TestNewReplaceURLQuerystringPatch(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/search?q=old#frag",
		nil,
	)
	values := url.Values{"q": {"new"}, "page": {"2"}}
	patch := datastar.NewReplaceURLQuerystringPatch(req, values)
	got := patch.Event()

	// Path is preserved, query replaced (url.Values.Encode sorts keys),
	// fragment dropped — mirroring upstream ReplaceURLQuerystring semantics.
	if !strings.Contains(
		got.Data,
		`window.history.replaceState({}, "", "/search?page=2&q=new")`,
	) {
		t.Errorf("should contain replaceState with new query; got %q", got.Data)
	}
}

func TestNewReplaceURLQuerystringPatch_ParityWireFormat(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/search?q=old",
		nil,
	)
	patch := datastar.NewReplaceURLQuerystringPatch(req, url.Values{"q": {"new"}})
	got := patch.Event()

	// Parity item 5: script patches always append to selector body.
	for _, want := range []string{"selector body", "mode append"} {
		if !strings.Contains(got.Data, want) {
			t.Errorf("wire format should contain %q; got %q", want, got.Data)
		}
	}

	// Parity item 4: default AutoRemove (nil) adds the data-effect attribute.
	if !strings.Contains(got.Data, `data-effect="el.remove()"`) {
		t.Errorf("should contain auto-remove effect; got %q", got.Data)
	}
}

func TestNewPrefetchPatch(t *testing.T) {
	t.Parallel()

	patch := datastar.NewPrefetchPatch("/page1", "/page2")
	got := patch.Event()

	// Should have type="speculationrules" attribute
	if !strings.Contains(got.Data, `type="speculationrules"`) {
		t.Errorf("should contain speculationrules type; got %q", got.Data)
	}
	// Should NOT have auto-remove (false)
	if strings.Contains(got.Data, `data-effect="el.remove()"`) {
		t.Errorf("should NOT contain auto-remove; got %q", got.Data)
	}
	// Should contain both URLs
	if !strings.Contains(got.Data, `"/page1"`) {
		t.Errorf("should contain /page1; got %q", got.Data)
	}

	if !strings.Contains(got.Data, `"/page2"`) {
		t.Errorf("should contain /page2; got %q", got.Data)
	}
	// Should contain prefetch source
	if !strings.Contains(got.Data, `"source": "list"`) {
		t.Errorf("should contain source list; got %q", got.Data)
	}
}

func TestDispatchCustomEventPatch_ImplementsPatch(t *testing.T) {
	t.Parallel()

	var _ datastar.Patch = datastar.DispatchCustomEventPatch{}
}
