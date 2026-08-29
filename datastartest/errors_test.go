package datastartest_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/go-datastar/datastartest"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

// These tests verify the error-family classification contract for errors
// returned by the datastartest helpers. The failingReader and errTestReadFailure
// sentinel are defined in collect_test.go and reused here.

func TestError_ReadEvents_ScanFailed_Transient(t *testing.T) {
	t.Parallel()

	_, err := datastartest.ReadEvents(failingReader{})
	if err == nil {
		t.Fatal("expected error from failing reader")
	}

	errorfamilytest.AssertCode(t, err, datastartest.CodeSSEScanFailed)
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertRetryable(t, err, true)

	if !errors.Is(err, errTestReadFailure) {
		t.Errorf("errors.Is(err, errTestReadFailure) = false; want true (chain traversal)")
	}

	if _, ok := errors.AsType[*errorfamily.Error](err); !ok {
		t.Errorf("errors.AsType[*errorfamily.Error](err) = false; want true")
	}
}

func TestError_ReadNEvents_ScanFailed_Transient(t *testing.T) {
	t.Parallel()

	_, err := datastartest.ReadNEvents(failingReader{}, 1)
	if err == nil {
		t.Fatal("expected error from failing reader")
	}

	errorfamilytest.AssertCode(t, err, datastartest.CodeSSEScanFailed)
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
}

func TestError_UnmarshalSignals_Rejection(t *testing.T) {
	t.Parallel()

	// Build an event with malformed signals JSON.
	input := "event: datastar-patch-signals\n" +
		"data: signals {bad json\n" +
		"\n"

	events, err := datastartest.ReadEvents(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	var target map[string]any

	unmarshalErr := events[0].UnmarshalSignals(&target)
	if unmarshalErr == nil {
		t.Fatal("expected unmarshal error from malformed JSON")
	}

	errorfamilytest.AssertCode(t, unmarshalErr, datastartest.CodeSignalsUnmarshalFailed)
	errorfamilytest.AssertFamily(t, unmarshalErr, errorfamily.Rejection)
	errorfamilytest.AssertRetryable(t, unmarshalErr, false)

	// The underlying cause must be preserved in the error chain so callers
	// can inspect the original decode failure.
	if cause := errors.Unwrap(unmarshalErr); cause == nil {
		t.Errorf("errors.Unwrap(unmarshalErr) = nil; want non-nil (cause preserved)")
	}
}

// TestError_AllPaths_AreErrorFamilyType verifies that every error returned by
// the test helpers is an *errorfamily.Error extractable via AsType.
func TestError_AllPaths_AreErrorFamilyType(t *testing.T) {
	t.Parallel()

	_, scanErr := datastartest.ReadEvents(failingReader{})

	events, _ := datastartest.ReadEvents(strings.NewReader(
		"event: datastar-patch-signals\ndata: signals {bad\n\n"))

	var unmarshalErr error
	if len(events) > 0 {
		unmarshalErr = events[0].UnmarshalSignals(new(map[string]any))
	}

	_, nEventsErr := datastartest.ReadNEvents(failingReader{}, 1)

	cases := []struct {
		name string
		err  error
	}{
		{"ReadEvents_ScanFailed", scanErr},
		{"UnmarshalSignals_Rejection", unmarshalErr},
		{"ReadNEvents_ScanFailed", nEventsErr},
	}

	for _, testCase := range cases {
		if testCase.err == nil {
			t.Errorf("%s: error is nil", testCase.name)

			continue
		}

		if _, ok := errors.AsType[*errorfamily.Error](testCase.err); !ok {
			t.Errorf("%s: errors.AsType[*errorfamily.Error](err) = false; want true", testCase.name)
		}
	}
}
