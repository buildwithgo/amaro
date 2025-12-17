package middlewares

import (
	"github.com/buildwithgo/amaro"
)

type SecureConfig struct {
	XSSProtection         string
	ContentTypeOptions    string
	FrameOptions          string
	HSTSMaxAge            int
	HSTSExcludeSubdomains bool
}

func DefaultSecureConfig() SecureConfig {
	return SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeOptions: "nosniff",
		FrameOptions:       "SAMEORIGIN",
		HSTSMaxAge:         31536000,
	}
}

// Secure adds security headers to the response.
func Secure(config ...SecureConfig) amaro.Middleware {
	cfg := DefaultSecureConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			if cfg.XSSProtection != "" {
				c.Writer.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}
			if cfg.ContentTypeOptions != "" {
				c.Writer.Header().Set("X-Content-Type-Options", cfg.ContentTypeOptions)
			}
			if cfg.FrameOptions != "" {
				c.Writer.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			// HSTS
			if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
				val := "max-age=31536000"
				if cfg.HSTSExcludeSubdomains {
					val += "; includeSubDomains"
				}
				c.Writer.Header().Set("Strict-Transport-Security", val)
			}
			return next(c)
		}
	}
}
