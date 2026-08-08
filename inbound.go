package datastar

import (
	"encoding/json/v2"
	"errors"
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
func ReadSignals(req *http.Request, signals any) error {
	input, err := readSignalsInput(req)
	if err != nil {
		return err
	}

	if len(input) == 0 {
		return nil
	}

	if err := json.Unmarshal(input, signals); err != nil {
		preview := string(input)
		if len(preview) > 200 {
			preview = preview[:200]
		}

		return errorfamily.WrapOncef(err, errorfamily.Rejection,
			CodeSignalsUnmarshalFailed, "unmarshal signals JSON into %T", signals).
			WithContext("method", req.Method).
			WithContext("input_bytes", strconv.Itoa(len(input))).
			WithContext("input_preview", preview)
	}

	return nil
}

// readSignalsInput extracts the raw bytes that should be unmarshaled into
// signals, choosing the query parameter for GET/DELETE and the body otherwise.
// Returns a nil slice (no error) if no signals are present in either location.
func readSignalsInput(req *http.Request) ([]byte, error) {
	if isQueryMethod(req.Method) {
		return readSignalsFromQuery(req)
	}

	return readSignalsFromBody(req)
}

// isQueryMethod reports whether the HTTP method carries signals in the URL
// query string rather than the request body. DataStar's protocol assigns
// GET/DELETE to query encoding and everything else to the body.
func isQueryMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodDelete
}

// readSignalsFromQuery returns the raw "datastar" query parameter bytes, or
// nil if the parameter is absent.
func readSignalsFromQuery(req *http.Request) ([]byte, error) {
	dsJSON := req.URL.Query().Get(DatastarKey)
	if dsJSON == "" {
		return nil, nil
	}

	return []byte(dsJSON), nil
}

// readSignalsFromBody reads and returns the full request body, or nil for an
// empty body. Returns an errorfamily-classified error for I/O failures.
func readSignalsFromBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		if errors.Is(err, http.ErrBodyReadAfterClose) {
			return nil, ErrBodyReadAfterClose
		}

		return nil, errorfamily.WrapOncef(err, errorfamily.Transient,
			CodeBodyReadFailed, "read request body").
			WithContext("method", req.Method)
	}

	if len(body) == 0 {
		return nil, nil
	}

	return body, nil
}

// LastEventID extracts the last event ID from an HTTP request. It checks the
// standard "Last-Event-ID" header first, then falls back to the "lastEventId"
// query parameter (which the DataStar JS client sends on reconnection).
//
// Returns an empty [sse.EventID] if no event ID is present.
func LastEventID(req *http.Request) sse.EventID {
	// Header takes priority
	if headerID := req.Header.Get("Last-Event-ID"); headerID != "" {
		return sse.NewEventID(headerID)
	}

	// DataStar JS client also sends lastEventId as a query param
	if queryID := req.URL.Query().Get("lastEventId"); queryID != "" {
		return sse.NewEventID(queryID)
	}

	return sse.EventID{}
}
