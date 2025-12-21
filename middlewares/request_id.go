package middlewares

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/buildwithgo/amaro"
)

const RequestIDKey = "request_id"

// RequestID adds an X-Request-ID header to the response and context.
func RequestID() amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			rid := c.Request.Header.Get("X-Request-ID")
			if rid == "" {
				id := make([]byte, 16)
				rand.Read(id)
				rid = hex.EncodeToString(id)
			}
			c.Writer.Header().Set("X-Request-ID", rid)
			c.Set(RequestIDKey, rid)
			return next(c)
		}
	}
}
