package datastar

import "fmt"

// GetSSE generates a DataStar @get SSE action attribute string.
func GetSSE(urlFormat string, args ...any) string {
	return fmt.Sprintf(`@get('%s')`, fmt.Sprintf(urlFormat, args...))
}

// PostSSE generates a DataStar @post SSE action attribute string.
func PostSSE(urlFormat string, args ...any) string {
	return fmt.Sprintf(`@post('%s')`, fmt.Sprintf(urlFormat, args...))
}

// PutSSE generates a DataStar @put SSE action attribute string.
func PutSSE(urlFormat string, args ...any) string {
	return fmt.Sprintf(`@put('%s')`, fmt.Sprintf(urlFormat, args...))
}

// PatchSSE generates a DataStar @patch SSE action attribute string.
func PatchSSE(urlFormat string, args ...any) string {
	return fmt.Sprintf(`@patch('%s')`, fmt.Sprintf(urlFormat, args...))
}

// DeleteSSE generates a DataStar @delete SSE action attribute string.
func DeleteSSE(urlFormat string, args ...any) string {
	return fmt.Sprintf(`@delete('%s')`, fmt.Sprintf(urlFormat, args...))
}
