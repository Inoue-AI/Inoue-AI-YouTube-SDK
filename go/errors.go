package youtube

import (
	"errors"
	"fmt"
)

// Error represents an upstream YouTube Data API error. Google returns a
// structured "error" object with an HTTP-style code and a list of nested
// errors; this struct flattens the most useful fields.
type Error struct {
	StatusCode int      // HTTP status code returned by the API.
	Code       int      // YouTube error.code (mirrors HTTP status).
	Status     string   // YouTube error.status (e.g. "PERMISSION_DENIED").
	Message    string   // Top-level error message.
	Reasons    []string // Per-error reason codes.
	Body       []byte   // Raw response body, for diagnostics.
}

func (e *Error) Error() string {
	return fmt.Sprintf("youtube: [%d/%s] %s (reasons=%v)", e.StatusCode, e.Status, e.Message, e.Reasons)
}

// IsAuthError reports whether the error indicates an authentication or
// authorisation failure.
func (e *Error) IsAuthError() bool {
	return e.StatusCode == 401 || e.StatusCode == 403
}

// IsRateLimited reports whether the error indicates a quota or rate-limit hit.
func (e *Error) IsRateLimited() bool {
	if e.StatusCode == 429 {
		return true
	}
	for _, r := range e.Reasons {
		switch r {
		case "quotaExceeded", "userRateLimitExceeded", "rateLimitExceeded":
			return true
		}
	}
	return false
}

// IsNotFound reports whether the resource was not found.
func (e *Error) IsNotFound() bool { return e.StatusCode == 404 }

// IsServerError reports whether the upstream returned a 5xx-style failure.
func (e *Error) IsServerError() bool { return e.StatusCode >= 500 }

// AsError unwraps an error into a *Error if applicable.
func AsError(err error) (*Error, bool) {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
