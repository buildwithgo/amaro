package amaro

import (
	"fmt"
	"net/http"
)

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	Code    int
	Message interface{}
	Internal error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// SetInternal sets the internal error.
func (e *HTTPError) SetInternal(err error) *HTTPError {
	e.Internal = err
	return e
}

// Unwrap returns the internal error.
func (e *HTTPError) Unwrap() error {
	return e.Internal
}
