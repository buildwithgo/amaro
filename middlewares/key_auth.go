package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

// KeyAuthConfig holds the configuration for Key Auth middleware.
type KeyAuthConfig struct {
	// KeyLookup is a string in the form of "header:Key-Name", "query:Key-Name", or "cookie:Key-Name".
	// Default is "header:X-API-Key".
	KeyLookup string

	// AuthScheme is the authentication scheme (e.g., "Bearer").
	// Only used if KeyLookup is "header". Default is "".
	AuthScheme string

	// Validator is the function to validate the key.
	Validator func(key string, c *amaro.Context) (bool, error)

	// ErrorHandler is called when an error occurs during key validation.
	ErrorHandler func(c *amaro.Context, err error) error

	// Skipper defines a function to skip middleware.
	Skipper func(c *amaro.Context) bool
}

// DefaultKeyAuthConfig returns a default configuration.
func DefaultKeyAuthConfig() KeyAuthConfig {
	return KeyAuthConfig{
		KeyLookup: "header:X-API-Key",
		Skipper:   func(c *amaro.Context) bool { return false },
		ErrorHandler: func(c *amaro.Context, err error) error {
			return amaro.NewHTTPError(http.StatusUnauthorized, err.Error())
		},
	}
}

// KeyAuth returns a Key Auth middleware.
func KeyAuth(validator func(key string, c *amaro.Context) (bool, error)) amaro.Middleware {
	config := DefaultKeyAuthConfig()
	config.Validator = validator
	return KeyAuthWithConfig(config)
}

// KeyAuthWithConfig returns a Key Auth middleware with custom configuration.
func KeyAuthWithConfig(config KeyAuthConfig) amaro.Middleware {
	if config.Validator == nil {
		panic("KeyAuth: validator function is required")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultKeyAuthConfig().Skipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultKeyAuthConfig().ErrorHandler
	}

	parts := strings.Split(config.KeyLookup, ":")
	extractor := func(c *amaro.Context) (string, error) {
		return "", errors.New("invalid key lookup configuration")
	}

	if len(parts) == 2 {
		switch parts[0] {
		case "header":
			extractor = func(c *amaro.Context) (string, error) {
				key := c.GetHeader(parts[1])
				if key == "" {
					return "", errors.New("missing key in header")
				}
				if config.AuthScheme != "" {
					if !strings.HasPrefix(key, config.AuthScheme+" ") {
						return "", errors.New("invalid key scheme")
					}
					return key[len(config.AuthScheme)+1:], nil
				}
				return key, nil
			}
		case "query":
			extractor = func(c *amaro.Context) (string, error) {
				key := c.QueryParam(parts[1])
				if key == "" {
					return "", errors.New("missing key in query")
				}
				return key, nil
			}
		case "cookie":
			extractor = func(c *amaro.Context) (string, error) {
				cookie, err := c.GetCookie(parts[1])
				if err != nil {
					return "", errors.New("missing key in cookie")
				}
				return cookie.Value, nil
			}
		}
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			key, err := extractor(c)
			if err != nil {
				return config.ErrorHandler(c, err)
			}

			valid, err := config.Validator(key, c)
			if err != nil {
				return config.ErrorHandler(c, err)
			}
			if !valid {
				return config.ErrorHandler(c, errors.New("invalid key"))
			}

			return next(c)
		}
	}
}
