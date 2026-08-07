package datastar_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
)

// --- Render Adapter Tests ---

type fakeTemplComponent struct {
	html string
	err  error
}

func (f fakeTemplComponent) Render(_ context.Context, w io.Writer) error {
	if f.err != nil {
		return f.err
	}
	_, err := w.Write([]byte(f.html))
	return err
}

type fakeGoStarRenderer struct {
	html string
	err  error
}

func (f fakeGoStarRenderer) Render(w io.Writer) error {
	if f.err != nil {
		return f.err
	}
	_, err := w.Write([]byte(f.html))
	return err
}

func TestElementsFromTempl(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.ElementsFromTempl(
			fakeTemplComponent{html: "<div>from templ</div>"},
			datastar.WithSelector("#target"),
		)
		if err != nil {
			t.Fatalf("ElementsFromTempl: %v", err)
		}

		got := patch.Event()
		if !bytes.Contains([]byte(got.Data), []byte("elements <div>from templ</div>")) {
			t.Errorf("should contain rendered HTML; got %q", got.Data)
		}
	})

	t.Run("render error", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.ElementsFromTempl(
			fakeTemplComponent{err: io.ErrUnexpectedEOF},
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestElementsFromGostar(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		patch, err := datastar.ElementsFromGostar(
			fakeGoStarRenderer{html: "<span>gostar</span>"},
		)
		if err != nil {
			t.Fatalf("ElementsFromGostar: %v", err)
		}

		got := patch.Event()
		if !bytes.Contains([]byte(got.Data), []byte("elements <span>gostar</span>")) {
			t.Errorf("should contain rendered HTML; got %q", got.Data)
		}
	})

	t.Run("render error", func(t *testing.T) {
		t.Parallel()

		_, err := datastar.ElementsFromGostar(
			fakeGoStarRenderer{err: io.ErrUnexpectedEOF},
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- HTTP Verb Helper Tests ---

func TestHTTPVerbHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"get", datastar.GetSSE("/api/items"), "@get('/api/items')"},
		{"get with format", datastar.GetSSE("/api/items/%d", 42), "@get('/api/items/42')"},
		{"post", datastar.PostSSE("/api/items"), "@post('/api/items')"},
		{"put", datastar.PutSSE("/api/items/1"), "@put('/api/items/1')"},
		{"patch", datastar.PatchSSE("/api/items/1"), "@patch('/api/items/1')"},
		{"delete", datastar.DeleteSSE("/api/items/1"), "@delete('/api/items/1')"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

// --- ReadSignals Tests ---

func TestReadSignals_FromBody(t *testing.T) {
	t.Parallel()

	body := strings.NewReader(`{"name":"test","count":5}`)
	r := httptest.NewRequest(http.MethodPost, "/api", body)

	type signals struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var s signals

	if err := datastar.ReadSignals(r, &s); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if s.Name != "test" {
		t.Errorf("Name: got %q, want %q", s.Name, "test")
	}
	if s.Count != 5 {
		t.Errorf("Count: got %d, want 5", s.Count)
	}
}

func TestReadSignals_FromQuery(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/api?datastar=%7B%22x%22%3A1%7D", nil)

	var s map[string]int
	if err := datastar.ReadSignals(r, &s); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if s["x"] != 1 {
		t.Errorf("x: got %d, want 1", s["x"])
	}
}

func TestReadSignals_EmptyQuery(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/api", nil)

	var s map[string]any
	if err := datastar.ReadSignals(r, &s); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if len(s) != 0 {
		t.Errorf("should be empty; got %v", s)
	}
}

func TestReadSignals_EmptyBody(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(""))

	var s map[string]any
	if err := datastar.ReadSignals(r, &s); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if len(s) != 0 {
		t.Errorf("should be empty; got %v", s)
	}
}

func TestReadSignals_MalformedJSON(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader("{bad json"))

	var s map[string]any
	err := datastar.ReadSignals(r, &s)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestReadSignals_NestedStruct(t *testing.T) {
	t.Parallel()

	body := strings.NewReader(`{"user":{"name":"bob"},"items":[1,2,3]}`)
	r := httptest.NewRequest(http.MethodPut, "/api", body)

	type nested struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
		Items []int `json:"items"`
	}
	var s nested

	if err := datastar.ReadSignals(r, &s); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if s.User.Name != "bob" {
		t.Errorf("User.Name: got %q, want bob", s.User.Name)
	}
	if len(s.Items) != 3 {
		t.Errorf("Items: got %d, want 3", len(s.Items))
	}
}

// --- LastEventID Tests ---

func TestLastEventID_FromHeader(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	r.Header.Set("Last-Event-ID", "42")

	id := datastar.LastEventID(r)
	if id.Get() != "42" {
		t.Errorf("got %q, want 42", id.Get())
	}
}

func TestLastEventID_FromQuery(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/events?lastEventId=99", nil)

	id := datastar.LastEventID(r)
	if id.Get() != "99" {
		t.Errorf("got %q, want 99", id.Get())
	}
}

func TestLastEventID_HeaderTakesPriority(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/events?lastEventId=99", nil)
	r.Header.Set("Last-Event-ID", "42")

	id := datastar.LastEventID(r)
	if id.Get() != "42" {
		t.Errorf("got %q, want 42 (header should win)", id.Get())
	}
}

func TestLastEventID_None(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/events", nil)

	id := datastar.LastEventID(r)
	if !id.IsZero() {
		t.Errorf("got %q, want zero", id.Get())
	}
}

func TestLastEventID_ReturnsSSEEventID(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/events", nil)
	r.Header.Set("Last-Event-ID", "abc")

	id := datastar.LastEventID(r)
	_ = id // verify return type is sse.EventID
}
