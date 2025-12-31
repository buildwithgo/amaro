package middlewares

import (
	"fmt"
	"net/http"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/sessions"
)

// SessionAuthConfig holds configuration for session-based auth.
type SessionAuthConfig[T any] struct {
	// Validator checks if the session data indicates an authenticated user.
	Validator func(data T, c *amaro.Context) (bool, error)

	// ErrorHandler handles errors (e.g., session not found).
	ErrorHandler func(c *amaro.Context, err error) error

	// Skipper skips middleware.
	Skipper func(c *amaro.Context) bool
}

// DefaultSessionAuthConfig returns defaults.
func DefaultSessionAuthConfig[T any]() SessionAuthConfig[T] {
	return SessionAuthConfig[T]{
		Skipper: func(c *amaro.Context) bool { return false },
		ErrorHandler: func(c *amaro.Context, err error) error {
			return amaro.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}
}

// SessionAuth returns a middleware that checks for a valid session.
// It assumes sessions.Start middleware is already applied.
func SessionAuth[T any](validator func(data T, c *amaro.Context) (bool, error)) amaro.Middleware {
	config := DefaultSessionAuthConfig[T]()
	config.Validator = validator
	return SessionAuthWithConfig(config)
}

// SessionAuthWithConfig returns middleware with custom config.
func SessionAuthWithConfig[T any](config SessionAuthConfig[T]) amaro.Middleware {
	if config.Validator == nil {
		panic("SessionAuth: validator function is required")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSessionAuthConfig[T]().Skipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultSessionAuthConfig[T]().ErrorHandler
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Retrieve session
			// Note: This relies on sessions package generic Get function
			// We assume T matches the T used in sessions.Start
			sess := sessions.Get[T](c)
			if sess == nil {
				return config.ErrorHandler(c, fmt.Errorf("session not found"))
			}

			// Validate
			valid, err := config.Validator(sess.Data, c)
			if err != nil {
				return config.ErrorHandler(c, err)
			}
			if !valid {
				return config.ErrorHandler(c, fmt.Errorf("invalid session"))
			}

			return next(c)
		}
	}
}
