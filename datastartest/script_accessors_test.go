package datastartest_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-datastar/datastartest"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// helperHandler runs a handler function that has access to a DataStar
// Response; the test server collects the SSE output.
func accessorHelperHandler(send func(*datastar.Response)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		send(datastar.NewResponse(stream))
	})
}

func TestEvent_RedirectURL(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, accessorHelperHandler(func(resp *datastar.Response) {
		_ = resp.Redirect("https://example.com/next?page=2")
	}))

	got := events[0].RedirectURL()
	if got != "https://example.com/next?page=2" {
		t.Errorf("RedirectURL() = %q; want %q", got, "https://example.com/next?page=2")
	}
}

func TestEvent_CustomEventAccessors(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, accessorHelperHandler(func(resp *datastar.Response) {
		_ = resp.DispatchCustomEvent("cart-updated", map[string]any{"count": 3})
	}))

	evt := events[0]

	if got := evt.CustomEventName(); got != "cart-updated" {
		t.Errorf("CustomEventName() = %q; want %q", got, "cart-updated")
	}

	var detail struct {
		Count int `json:"count"`
	}

	if err := evt.UnmarshalCustomEventDetail(&detail); err != nil {
		t.Fatalf("UnmarshalCustomEventDetail: %v", err)
	}

	if detail.Count != 3 {
		t.Errorf("detail.Count = %d; want 3", detail.Count)
	}

	if !strings.Contains(evt.CustomEventDetail(), `"count":3`) {
		t.Errorf("CustomEventDetail() = %q; want JSON containing count:3", evt.CustomEventDetail())
	}
}

func TestEvent_UnmarshalCustomEventDetail_ClassifiedError(t *testing.T) {
	t.Parallel()

	plain := datastartest.Event{Type: "datastar-patch-elements"}

	err := plain.UnmarshalCustomEventDetail(new(map[string]any))
	if err == nil {
		t.Fatal("expected error for event without detail")
	}

	if got := errorfamily.Code(err); got != datastartest.CodeCustomEventDetailUnmarshalFailed {
		t.Errorf("errorfamily.Code = %q; want %q", got, datastartest.CodeCustomEventDetailUnmarshalFailed)
	}
}

func TestEvent_ScriptAttributes(t *testing.T) {
	t.Parallel()

	events := datastartest.Collect(t, accessorHelperHandler(func(resp *datastar.Response) {
		_ = resp.Prefetch("/page1")
	}))

	attrs := events[0].ScriptAttributes()

	joined := strings.Join(attrs, " ")
	if !strings.Contains(joined, `type="speculationrules"`) {
		t.Errorf("ScriptAttributes() should contain speculationrules type; got %v", attrs)
	}
}

func TestEvent_ScriptAccessors_ZeroValuesOnNonScript(t *testing.T) {
	t.Parallel()

	plain := datastartest.Event{Type: "datastar-patch-elements"}

	if got := plain.RedirectURL(); got != "" {
		t.Errorf("RedirectURL() on non-script = %q; want empty", got)
	}

	if got := plain.CustomEventName(); got != "" {
		t.Errorf("CustomEventName() on non-script = %q; want empty", got)
	}

	if got := plain.CustomEventDetail(); got != "" {
		t.Errorf("CustomEventDetail() on non-script = %q; want empty", got)
	}

	if got := plain.ScriptAttributes(); got != nil {
		t.Errorf("ScriptAttributes() on non-script = %v; want nil", got)
	}
}
