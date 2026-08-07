package datastar

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-sse"
)

// ReadSignals extracts DataStar signals from an HTTP request and unmarshals
// them into the signals target (a pointer to a struct).
//
// For GET and DELETE requests, signals are read from the "datastar" query parameter.
// For all other methods, signals are read from the JSON request body.
//
// Returns nil (with no data written) if no signals are present (empty query
// param or empty body).
func ReadSignals(r *http.Request, signals any) error {
	var input []byte

	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		dsJSON := r.URL.Query().Get(DatastarKey)
		if dsJSON == "" {
			return nil
		}
		input = []byte(dsJSON)
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			if err == http.ErrBodyReadAfterClose {
				return ErrBodyReadAfterClose
			}
			return errorfamily.Wrapf(err, errorfamily.Transient,
				CodeBodyReadFailed, "read request body").
				WithContext("method", r.Method)
		}
		if len(body) == 0 {
			return nil
		}
		input = body
	}

	if err := json.Unmarshal(input, signals); err != nil {
		return errorfamily.Wrapf(err, errorfamily.Rejection,
			CodeSignalsUnmarshalFailed, "unmarshal signals JSON into %T", signals).
			WithContext("method", r.Method).
			WithContext("input_bytes", strconv.Itoa(len(input)))
	}
	return nil
}

// LastEventID extracts the last event ID from an HTTP request. It checks the
// standard "Last-Event-ID" header first, then falls back to the "lastEventId"
// query parameter (which the DataStar JS client sends on reconnection).
//
// Returns an empty [sse.EventID] if no event ID is present.
func LastEventID(r *http.Request) sse.EventID {
	// Header takes priority
	if headerID := r.Header.Get("Last-Event-ID"); headerID != "" {
		return sse.NewEventID(headerID)
	}

	// DataStar JS client also sends lastEventId as a query param
	if queryID := r.URL.Query().Get("lastEventId"); queryID != "" {
		return sse.NewEventID(queryID)
	}

	return sse.EventID{}
}
