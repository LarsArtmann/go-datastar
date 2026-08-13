package datastartest

// Error codes for datastartest. Each is a stable string accessible via
// [errorfamily.Code], enabling programmatic classification of errors returned
// by the test helpers without string matching on human-readable messages.
const (
	// CodeSSEScanFailed: [ReadEvents] or [ReadNEvents] encountered an I/O
	// error while scanning the SSE response stream.
	CodeSSEScanFailed = "datastartest.sse_scan_failed"

	// CodeSignalsUnmarshalFailed: [Event.UnmarshalSignals] could not decode
	// the signals JSON payload from a patch-signals event.
	CodeSignalsUnmarshalFailed = "datastartest.signals_unmarshal_failed"
)
