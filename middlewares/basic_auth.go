package middlewares

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

// BasicAuthConfig holds the configuration for Basic Auth middleware.
type BasicAuthConfig struct {
	// Validator is the function to validate username and password.
	Validator func(username, password string, c *amaro.Context) (bool, error)

	// Realm is the authentication realm. Default is "Restricted".
	Realm string

	// Skipper defines a function to skip middleware.
	Skipper func(c *amaro.Context) bool
}

// BasicAuthValidator defines the function signature for validating credentials.
type BasicAuthValidator func(username, password string, c *amaro.Context) (bool, error)

// DefaultBasicAuthConfig returns a default configuration.
func DefaultBasicAuthConfig() BasicAuthConfig {
	return BasicAuthConfig{
		Realm:   "Restricted",
		Skipper: func(c *amaro.Context) bool { return false },
	}
}

// BasicAuth returns a Basic Auth middleware.
func BasicAuth(validator BasicAuthValidator) amaro.Middleware {
	config := DefaultBasicAuthConfig()
	config.Validator = validator
	return BasicAuthWithConfig(config)
}

// BasicAuthWithConfig returns a Basic Auth middleware with custom configuration.
func BasicAuthWithConfig(config BasicAuthConfig) amaro.Middleware {
	if config.Validator == nil {
		panic("BasicAuth: validator function is required")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultBasicAuthConfig().Skipper
	}
	if config.Realm == "" {
		config.Realm = "Restricted"
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			auth := c.GetHeader("Authorization")
			if auth == "" {
				c.SetHeader("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
				return amaro.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			const prefix = "Basic "
			if !strings.HasPrefix(auth, prefix) {
				return amaro.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header")
			}

			decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
			if err != nil {
				return amaro.NewHTTPError(http.StatusUnauthorized, "Invalid base64")
			}

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 {
				return amaro.NewHTTPError(http.StatusUnauthorized, "Invalid credentials format")
			}

			valid, err := config.Validator(creds[0], creds[1], c)
			if err != nil {
				return err
			}
			if !valid {
				c.SetHeader("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
				return amaro.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			return next(c)
		}
	}
}
