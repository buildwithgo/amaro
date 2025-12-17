package middlewares

import (
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	AllowOrigins []string
	// AllowMethods is a list of methods the client is allowed to use with cross-domain requests.
	AllowMethods []string
	// AllowHeaders is a list of non-simple headers the client is allowed to use with cross-domain requests.
	AllowHeaders []string
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}
}

// CORS returns a Cross-Origin Resource Sharing middleware.
func CORS(config ...CORSConfig) amaro.Middleware {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			origin := c.Request.Header.Get("Origin")
			allowOrigin := ""

			for _, o := range cfg.AllowOrigins {
				if o == "*" || o == origin {
					allowOrigin = o
					if o == "*" {
						allowOrigin = "*" // Or echo back origin if credentials needed
					}
					break
				}
			}

			if allowOrigin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ","))
				c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ","))
			}

			// Handle Preflight
			if c.Request.Method == http.MethodOptions {
				c.Writer.WriteHeader(http.StatusNoContent)
				return nil
			}

			return next(c)
		}
	}
}
