package datastar_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/go-sse"
)

// These tests verify the strongly typed error contract: every failure path
// returns an errorfamily-classified error with the correct Family, a stable
// machine-readable Code, and (where relevant) structured context. They also
// confirm sentinel errors are matchable via errors.Is and stay context-pristine.

// errConnReset is a static sentinel used to fake a transient body-read failure.
var errConnReset = errors.New("connection reset")

// --- Render failures (Orchestration: internal output-production failure) ---

func TestError_ElementsFromTempl_Classified(t *testing.T) {
	t.Parallel()

	_, err := datastar.ElementsFromTempl(fakeTemplComponent{err: io.ErrUnexpectedEOF})

	errorfamilytest.AssertFamily(t, err, errorfamily.Orchestration)
	errorfamilytest.AssertCode(t, err, datastar.CodeTemplRenderFailed)
	errorfamilytest.AssertRetryable(t, err, false)

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("cause not preserved: errors.Is(err, io.ErrUnexpectedEOF) = false")
	}
}

func TestError_ElementsFromGostar_Classified(t *testing.T) {
	t.Parallel()

	_, err := datastar.ElementsFromGostar(fakeGoStarRenderer{err: io.ErrUnexpectedEOF})

	errorfamilytest.AssertFamily(t, err, errorfamily.Orchestration)
	errorfamilytest.AssertCode(t, err, datastar.CodeGostarRenderFailed)
	errorfamilytest.AssertRetryable(t, err, false)
}

// --- ReadSignals: body already closed (Rejection: API misuse) ---

type afterCloseBody struct{}

func (afterCloseBody) Read([]byte) (int, error) { return 0, http.ErrBodyReadAfterClose }
func (afterCloseBody) Close() error             { return nil }

func TestError_ReadSignals_BodyReadAfterClose(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	r.Body = afterCloseBody{}

	err := datastar.ReadSignals(r, new(map[string]any))

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeBodyReadAfterClose)
	errorfamilytest.AssertRetryable(t, err, false)

	if !errors.Is(err, datastar.ErrBodyReadAfterClose) {
		t.Errorf("errors.Is(err, ErrBodyReadAfterClose) = false; want true (sentinel match)")
	}

	if !errors.Is(err, http.ErrBodyReadAfterClose) {
		t.Errorf("underlying http.ErrBodyReadAfterClose cause not preserved in chain")
	}
}

// --- ReadSignals: body read failure (Transient: retryable I/O) ---

type failingBody struct{ err error }

func (f failingBody) Read([]byte) (int, error) { return 0, f.err }
func (failingBody) Close() error               { return nil }

func TestError_ReadSignals_BodyReadFailed_Transient(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	r.Body = failingBody{err: errConnReset}

	err := datastar.ReadSignals(r, new(map[string]any))

	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertCode(t, err, datastar.CodeBodyReadFailed)
	errorfamilytest.AssertRetryable(t, err, true)
	errorfamilytest.AssertContext(t, err, "method", http.MethodPost)
}

// --- ReadSignals: unmarshal failure (Rejection: malformed input) ---

func TestError_ReadSignals_UnmarshalFailed_Rejection(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/",
		strings.NewReader("{bad json"),
	)

	var s map[string]any

	err := datastar.ReadSignals(req, &s)

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeSignalsUnmarshalFailed)
	errorfamilytest.AssertRetryable(t, err, false)
	errorfamilytest.AssertContext(t, err, "method", http.MethodPost)
	errorfamilytest.AssertContext(t, err, "input_bytes", "9") // len("{bad json")
	errorfamilytest.AssertContext(t, err, "input_preview", "{bad json")
}

// --- MarshalSignals failure (Rejection: unmarshallable value) ---

func TestError_MarshalSignals_Classified(t *testing.T) {
	t.Parallel()

	_, err := datastar.MarshalSignals(make(chan int))

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeSignalsMarshalFailed)
	errorfamilytest.AssertRetryable(t, err, false)
}

func TestError_NewSignalsPatch_PropagatesMarshalError(t *testing.T) {
	t.Parallel()

	_, err := datastar.NewSignalsPatch(make(chan int))

	errorfamilytest.AssertCode(t, err, datastar.CodeSignalsMarshalFailed)
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
}

// --- DispatchCustomEvent: empty name (Rejection: missing required input) ---

func TestError_NewDispatchCustomEventPatch_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := datastar.NewDispatchCustomEventPatch("", nil)

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeEventNameRequired)
	errorfamilytest.AssertRetryable(t, err, false)

	if !errors.Is(err, datastar.ErrEventNameRequired) {
		t.Errorf("errors.Is(err, ErrEventNameRequired) = false; want true (sentinel match)")
	}
}

func TestError_NewDispatchCustomEventPatch_UnmarshallableDetail(t *testing.T) {
	t.Parallel()

	_, err := datastar.NewDispatchCustomEventPatch("test", make(chan int))

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeCustomEventDetailMarshalFailed)
	errorfamilytest.AssertRetryable(t, err, false)
}

// --- Parse failures (Rejection: unrecognized input string) ---

func TestError_ElementPatchModeFromString_Invalid(t *testing.T) {
	t.Parallel()

	_, err := datastar.ElementPatchModeFromString("bogus")

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeElementPatchModeInvalid)
	errorfamilytest.AssertRetryable(t, err, false)
}

func TestError_NamespaceFromString_Invalid(t *testing.T) {
	t.Parallel()

	_, err := datastar.NamespaceFromString("bogus")

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, datastar.CodeNamespaceInvalid)
	errorfamilytest.AssertRetryable(t, err, false)
}

// --- Sentinel pristineness: shared sentinels must never leak caller context ---

func TestError_Sentinels_AreContextPristine(t *testing.T) {
	t.Parallel()

	// Using a sentinel via WithContext returns a clone; the original stays clean.
	// First exercise a path that enriches a derived error, then confirm the
	// package-level sentinel itself carries no context.
	_, _ = datastar.NewDispatchCustomEventPatch("", nil)

	errorfamilytest.AssertContextMissing(t, datastar.ErrEventNameRequired, "anything")
	errorfamilytest.AssertContextMissing(t, datastar.ErrBodyReadAfterClose, "anything")
}

// --- Cross-layer composition: go-datastar wraps go-sse's classified errors ---

// TestWrapStreamError_DoubleWrapComposition verifies that when go-sse v0.5.0
// returns an errorfamily-classified Send error, go-datastar's wrapStreamError
// adds a second classification layer that wins (outermost family/code), while
// errors.Is still traverses the full chain to the original cause.
func TestWrapStreamError_DoubleWrapComposition(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events", nil)
	stream := sse.NewStream(failingWriter{}, r)
	resp := datastar.NewResponse(stream)

	err := resp.PatchElements("<div>hi</div>")
	if err == nil {
		t.Fatal("expected error from failing writer")
	}

	// Outermost layer is go-datastar's classification.
	errorfamilytest.AssertCode(t, err, datastar.CodeStreamSendFailed)
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertRetryable(t, err, true)

	// errors.Is traverses through both errorfamily wraps to the original cause.
	if !errors.Is(err, errWriteFailed) {
		t.Errorf("errors.Is(err, errWriteFailed) = false; want true (chain traversal)")
	}

	// The error is an *errorfamily.Error (extractable via AsType).
	if _, ok := errors.AsType[*errorfamily.Error](err); !ok {
		t.Errorf("errors.AsType[*errorfamily.Error](err) = false; want true")
	}
}

// --- AsType coverage: every error path must be an *errorfamily.Error ---

func TestError_AllPaths_AreErrorFamilyType(t *testing.T) {
	t.Parallel()

	// Compute one error from each path that returns an error.
	_, templErr := datastar.ElementsFromTempl(fakeTemplComponent{err: io.ErrUnexpectedEOF})
	_, gostarErr := datastar.ElementsFromGostar(fakeGoStarRenderer{err: io.ErrUnexpectedEOF})

	closeReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	closeReq.Body = afterCloseBody{}
	bodyCloseErr := datastar.ReadSignals(closeReq, new(map[string]any))

	failReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	failReq.Body = failingBody{err: errConnReset}
	bodyReadErr := datastar.ReadSignals(failReq, new(map[string]any))

	badReq := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/", strings.NewReader("{bad"))
	unmarshalErr := datastar.ReadSignals(badReq, new(map[string]any))

	_, marshalErr := datastar.MarshalSignals(make(chan int))
	_, emptyNameErr := datastar.NewDispatchCustomEventPatch("", nil)
	_, badDetailErr := datastar.NewDispatchCustomEventPatch("test", make(chan int))
	_, modeErr := datastar.ElementPatchModeFromString("bogus")
	_, nsErr := datastar.NamespaceFromString("bogus")

	cases := []struct {
		name string
		err  error
	}{
		{"ElementsFromTempl", templErr},
		{"ElementsFromGostar", gostarErr},
		{"ReadSignals_BodyReadAfterClose", bodyCloseErr},
		{"ReadSignals_BodyReadFailed", bodyReadErr},
		{"ReadSignals_UnmarshalFailed", unmarshalErr},
		{"MarshalSignals", marshalErr},
		{"NewDispatchCustomEventPatch_EmptyName", emptyNameErr},
		{"NewDispatchCustomEventPatch_UnmarshallableDetail", badDetailErr},
		{"ElementPatchModeFromString", modeErr},
		{"NamespaceFromString", nsErr},
	}

	for _, tc := range cases {
		if _, ok := errors.AsType[*errorfamily.Error](tc.err); !ok {
			t.Errorf("%s: errors.AsType[*errorfamily.Error](err) = false; want true", tc.name)
		}
	}
}
