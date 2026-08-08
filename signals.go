package datastar

import (
	"bytes"
	"encoding/json/v2"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// SignalsPatch patches reactive signals on the DataStar client.
// The signals payload must be JSON-encoded bytes.
//
// Construct one with [NewSignalsPatch] (marshals a Go value) or directly
// (if you already have JSON bytes):
//
//	// From a Go struct:
//	patch, err := datastar.NewSignalsPatch(map[string]any{"count": 42})
//
//	// From pre-encoded JSON:
//	patch := datastar.SignalsPatch{Signals: []byte(`{"count":42}`)}
type SignalsPatch struct {
	// Signals is the JSON-encoded signal payload.
	Signals []byte

	// OnlyIfMissing instructs the client to only set signals that don't
	// already exist.
	OnlyIfMissing bool

	// EventID is an optional SSE event identifier.
	EventID string

	// RetryDuration overrides the default SSE retry interval. Only emitted
	// when > 0 and != [DefaultRetryDuration].
	RetryDuration time.Duration
}

// SignalsPatchOption configures a [SignalsPatch].
type SignalsPatchOption func(*SignalsPatch)

// WithSignalsEventID sets the SSE event ID for the signals patch.
func WithSignalsEventID(id string) SignalsPatchOption {
	return func(p *SignalsPatch) { p.EventID = id }
}

// WithSignalsRetryDuration overrides the SSE retry duration for the signals patch.
func WithSignalsRetryDuration(d time.Duration) SignalsPatchOption {
	return func(p *SignalsPatch) { p.RetryDuration = d }
}

// WithOnlyIfMissing instructs the client to only patch signals that are missing.
func WithOnlyIfMissing(onlyIfMissing bool) SignalsPatchOption {
	return func(p *SignalsPatch) { p.OnlyIfMissing = onlyIfMissing }
}

// NewSignalsPatch creates a [SignalsPatch] from a Go value, marshaling it to
// JSON. Returns an error if marshaling fails (unlike the SDK's panicking
// MarshalAndPatchSignals).
func NewSignalsPatch(v any, opts ...SignalsPatchOption) (SignalsPatch, error) {
	b, err := MarshalSignals(v)
	if err != nil {
		return SignalsPatch{}, err
	}

	patch := SignalsPatch{
		Signals:       b,
		RetryDuration: DefaultRetryDuration,
	}
	for _, opt := range opts {
		opt(&patch)
	}

	return patch, nil
}

// NewSignalsIfMissingPatch creates a [SignalsPatch] with OnlyIfMissing=true.
func NewSignalsIfMissingPatch(v any, opts ...SignalsPatchOption) (SignalsPatch, error) {
	return NewSignalsPatch(v, append(opts, WithOnlyIfMissing(true))...)
}

// MarshalSignals marshals a Go value to JSON for use as a DataStar signals
// payload. Returns an error instead of panicking.
func MarshalSignals(v any) ([]byte, error) { //nolint:erraudit // returns error interface by design — idiomatic Go, consistent with go-sse
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errorfamily.Wrapf(err, errorfamily.Rejection,
			CodeSignalsMarshalFailed, "marshal signals value of type %T", v)
	}

	return b, nil
}

// Event returns the [sse.Event] for this signals patch. The data lines are:
//
//  1. onlyIfMissing true (if OnlyIfMissing is set)
//  2. signals <line> (one per line of the JSON payload)
func (p SignalsPatch) Event() sse.Event {
	var dataLines []string

	if p.OnlyIfMissing {
		dataLines = append(dataLines, OnlyIfMissingDatalineKey+"true")
	}

	// SDK uses bytes.SplitSeq — split on \n, emit every line unconditionally
	for line := range bytes.SplitSeq(p.Signals, []byte("\n")) {
		dataLines = append(dataLines, SignalsDatalineKey+string(line))
	}

	evt := sse.Event{
		Event: string(EventTypePatchSignals),
		Data:  strings.Join(dataLines, "\n"),
	}

	if p.EventID != "" {
		evt.ID = sse.NewEventID(p.EventID)
	}

	if p.RetryDuration > 0 && p.RetryDuration != DefaultRetryDuration {
		evt.Retry = uint(
			p.RetryDuration.Milliseconds(),
		)
	}

	return evt
}
