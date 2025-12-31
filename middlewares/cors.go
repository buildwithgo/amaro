package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/buildwithgo/amaro"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is []string{"*"}.
	AllowOrigins []string

	// AllowOriginFunc is a custom function to validate the origin. It takes the origin as an argument
	// and returns true if allowed or false otherwise. If this function is set, AllowOrigins is ignored.
	AllowOriginFunc func(origin string) bool

	// AllowMethods is a list of methods the client is allowed to use with cross-domain requests.
	// Default value is allowedMethodsDefault.
	AllowMethods []string

	// AllowHeaders is a list of non-simple headers the client is allowed to use with cross-domain requests.
	AllowHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool

	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS API specification.
	ExposeHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	MaxAge int
}

var allowedMethodsDefault = []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: allowedMethodsDefault,
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:       86400,
	}
}

// CORS returns a Cross-Origin Resource Sharing middleware.
func CORS(config ...CORSConfig) amaro.Middleware {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		c := config[0]
		// Merge with default if empty
		if len(c.AllowOrigins) > 0 {
			cfg.AllowOrigins = c.AllowOrigins
		}
		if c.AllowOriginFunc != nil {
			cfg.AllowOriginFunc = c.AllowOriginFunc
		}
		if len(c.AllowMethods) > 0 {
			cfg.AllowMethods = c.AllowMethods
		}
		if len(c.AllowHeaders) > 0 {
			cfg.AllowHeaders = c.AllowHeaders
		}
		if c.AllowCredentials {
			cfg.AllowCredentials = true
		}
		if len(c.ExposeHeaders) > 0 {
			cfg.ExposeHeaders = c.ExposeHeaders
		}
		if c.MaxAge > 0 {
			cfg.MaxAge = c.MaxAge
		}
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			req := c.Request
			res := c.Writer
			origin := req.Header.Get("Origin")
			allowOrigin := ""

			// Preflight request?
			preflight := req.Method == http.MethodOptions

			c.Writer.Header().Add("Vary", "Origin")

			if cfg.AllowOriginFunc != nil {
				if cfg.AllowOriginFunc(origin) {
					allowOrigin = origin
				}
			} else {
				for _, o := range cfg.AllowOrigins {
					if o == "*" && cfg.AllowCredentials {
						allowOrigin = origin
						break
					}
					if o == "*" || o == origin {
						allowOrigin = o
						if o == "*" {
							allowOrigin = "*"
						}
						break
					}
				}
			}

			if allowOrigin != "" {
				res.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				if cfg.AllowCredentials {
					res.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if len(cfg.ExposeHeaders) > 0 {
					res.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ","))
				}
			} else {
				// Origin not allowed
				if preflight {
					return c.String(http.StatusNoContent, "")
				}
				return next(c)
			}

			if preflight {
				res.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ","))
				res.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ","))
				res.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
				return c.String(http.StatusNoContent, "")
			}

			return next(c)
		}
	}
}
