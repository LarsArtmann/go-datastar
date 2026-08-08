package datastar_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/larsartmann/go-datastar"
	errorfamily "github.com/larsartmann/go-error-family"
)

// Example_errorHandlingByCode shows how to match errors by their stable code.
// Codes are string constants that never change between versions, making them
// safe for programmatic branching.
func Example_errorHandlingByCode() {
	// Simulate a request with malformed JSON signals
	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/", strings.NewReader("{bad json"))

	var target map[string]any

	err := datastar.ReadSignals(req, &target)

	switch errorfamily.Code(err) {
	case datastar.CodeSignalsUnmarshalFailed:
		fmt.Println("bad JSON — ask the user to fix their input")
	case datastar.CodeBodyReadFailed:
		fmt.Println("transient I/O — safe to retry")
	default:
		fmt.Println("unexpected error")
	}
	// Output: bad JSON — ask the user to fix their input
}

// Example_errorHandlingBySentinel shows how to match errors by sentinel value.
// errors.Is works because errorfamily compares by (code, family), so even a
// context-enriched clone of a sentinel still matches.
func Example_errorHandlingBySentinel() {
	_, err := datastar.NewDispatchCustomEventPatch("", nil)

	if errors.Is(err, datastar.ErrEventNameRequired) {
		fmt.Println("eventName is required")
	}
	// Output: eventName is required
}

// Example_errorHandlingByFamily shows how to use the behavioral family to
// decide whether to retry. Transient errors (temporary I/O) are safe to retry
// with backoff; Rejection errors are not (the caller's input is wrong).
func Example_errorHandlingByFamily() {
	// MarshalSignals with an unmarshallable value (channel)
	_, err := datastar.MarshalSignals(make(chan int))

	if errorfamily.IsRetryable(err) {
		fmt.Println("retrying...")
	} else {
		fmt.Println("not retryable — fix the input")
	}
	// Output: not retryable — fix the input
}
