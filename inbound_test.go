package datastar_test

import (
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
		if !strings.Contains(got.Data, "elements <div>from templ</div>") {
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
		if !strings.Contains(got.Data, "elements <span>gostar</span>") {
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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api", body)

	type signals struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var decoded signals

	if err := datastar.ReadSignals(req, &decoded); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if decoded.Name != "test" {
		t.Errorf("Name: got %q, want %q", decoded.Name, "test")
	}

	if decoded.Count != 5 {
		t.Errorf("Count: got %d, want 5", decoded.Count)
	}
}

func TestReadSignals_FromQuery(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api?datastar=%7B%22x%22%3A1%7D",
		nil,
	)

	var decoded map[string]int
	if err := datastar.ReadSignals(req, &decoded); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if decoded["x"] != 1 {
		t.Errorf("x: got %d, want 1", decoded["x"])
	}
}

func TestReadSignals_EmptyQuery(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api", nil)

	var decoded map[string]any
	if err := datastar.ReadSignals(req, &decoded); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("should be empty; got %v", decoded)
	}
}

func TestReadSignals_EmptyBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api",
		strings.NewReader(""),
	)

	var decoded map[string]any
	if err := datastar.ReadSignals(req, &decoded); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("should be empty; got %v", decoded)
	}
}

func TestReadSignals_MalformedJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api",
		strings.NewReader("{bad json"),
	)

	var decoded map[string]any

	err := datastar.ReadSignals(req, &decoded)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestReadSignals_NestedStruct(t *testing.T) {
	t.Parallel()

	body := strings.NewReader(`{"user":{"name":"bob"},"items":[1,2,3]}`)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api", body)

	type nested struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
		Items []int `json:"items"`
	}

	var decoded nested

	if err := datastar.ReadSignals(req, &decoded); err != nil {
		t.Fatalf("ReadSignals: %v", err)
	}

	if decoded.User.Name != "bob" {
		t.Errorf("User.Name: got %q, want bob", decoded.User.Name)
	}

	if len(decoded.Items) != 3 {
		t.Errorf("Items: got %d, want 3", len(decoded.Items))
	}
}

// --- LastEventID Tests ---

func TestLastEventID_FromHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	req.Header.Set("Last-Event-ID", "42")

	id := datastar.LastEventID(req)
	if id.Get() != "42" {
		t.Errorf("got %q, want 42", id.Get())
	}
}

func TestLastEventID_FromQuery(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events?lastEventId=99",
		nil,
	)

	id := datastar.LastEventID(req)
	if id.Get() != "99" {
		t.Errorf("got %q, want 99", id.Get())
	}
}

func TestLastEventID_HeaderTakesPriority(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events?lastEventId=99",
		nil,
	)
	req.Header.Set("Last-Event-ID", "42")

	id := datastar.LastEventID(req)
	if id.Get() != "42" {
		t.Errorf("got %q, want 42 (header should win)", id.Get())
	}
}

func TestLastEventID_None(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)

	id := datastar.LastEventID(req)
	if !id.IsZero() {
		t.Errorf("got %q, want zero", id.Get())
	}
}

func TestLastEventID_ReturnsSSEEventID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	req.Header.Set("Last-Event-ID", "abc")

	id := datastar.LastEventID(req)
	_ = id // verify return type is sse.EventID
}
