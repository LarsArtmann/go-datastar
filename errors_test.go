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

// --- errors.As coverage: every error path must be an *errorfamily.Error ---

func TestError_AllPaths_AreErrorFamilyType(t *testing.T) {
	t.Parallel()

	var ef *errorfamily.Error

	cases := []struct {
		name string
		err  error
	}{
		{"ElementsFromTempl", func() error {
			_, e := datastar.ElementsFromTempl(fakeTemplComponent{err: io.ErrUnexpectedEOF})
			return e
		}()},
		{"ElementsFromGostar", func() error {
			_, e := datastar.ElementsFromGostar(fakeGoStarRenderer{err: io.ErrUnexpectedEOF})
			return e
		}()},
		{"ReadSignals_BodyReadAfterClose", func() error {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
			r.Body = afterCloseBody{}
			return datastar.ReadSignals(r, new(map[string]any))
		}()},
		{"ReadSignals_BodyReadFailed", func() error {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
			r.Body = failingBody{err: errConnReset}
			return datastar.ReadSignals(r, new(map[string]any))
		}()},
		{"ReadSignals_UnmarshalFailed", func() error {
			req := httptest.NewRequestWithContext(
				context.Background(), http.MethodPost, "/", strings.NewReader("{bad"))
			return datastar.ReadSignals(req, new(map[string]any))
		}()},
		{"MarshalSignals", func() error {
			_, e := datastar.MarshalSignals(make(chan int))
			return e
		}()},
		{"NewDispatchCustomEventPatch_EmptyName", func() error {
			_, e := datastar.NewDispatchCustomEventPatch("", nil)
			return e
		}()},
		{"NewDispatchCustomEventPatch_UnmarshallableDetail", func() error {
			_, e := datastar.NewDispatchCustomEventPatch("test", make(chan int))
			return e
		}()},
		{"ElementPatchModeFromString", func() error {
			_, e := datastar.ElementPatchModeFromString("bogus")
			return e
		}()},
		{"NamespaceFromString", func() error {
			_, e := datastar.NamespaceFromString("bogus")
			return e
		}()},
	}

	for _, tc := range cases {
		if !errors.As(tc.err, &ef) {
			t.Errorf("%s: errors.As(err, &*errorfamily.Error) = false; want true", tc.name)
		}
	}
}
