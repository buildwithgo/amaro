package cache

import (
	"bytes"
	"net/http"
	"time"

	"github.com/buildwithgo/amaro"
)

// responseRecorder captures the response status and body for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// CachePage returns a middleware that caches the response body for a given duration.
// It uses the Cache interface.
func CachePage(store Cache, ttl time.Duration) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			// Only cache GET requests
			if c.Request.Method != http.MethodGet {
				return next(c)
			}

			key := "route_cache:" + c.Request.URL.String()

			// Check cache
			if val, ok := store.Get(key); ok {
				// Hit - We must assert to []byte
				if bodyBytes, ok := val.([]byte); ok {
					c.Writer.Header().Set("X-Cache", "HIT")
					c.Writer.Write(bodyBytes)
					return nil
				}
			}

			// Miss
			recorder := &responseRecorder{
				ResponseWriter: c.Writer,
				statusCode:     http.StatusOK, // Default
				body:           &bytes.Buffer{},
			}
			c.Writer = recorder

			// Process request
			err := next(c)

			// If successful, cache the result
			if err == nil && recorder.statusCode == http.StatusOK {
				store.Set(key, recorder.body.Bytes(), ttl)
			}

			return err
		}
	}
}
