package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/buildwithgo/amaro"
)

// CachedResponse stores the response data.
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// responseRecorder captures the response status, headers, and body for caching.
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

// KeyGenerator allows customizing the cache key.
type KeyGenerator func(c *amaro.Context) string

func DefaultKeyGenerator(c *amaro.Context) string {
	return "route_cache:" + c.Request.URL.String()
}

// CachePage returns a middleware that caches the response for a given duration.
// It uses the Cache interface.
func CachePage(store Cache, ttl time.Duration, keyGen ...KeyGenerator) amaro.Middleware {
	getKey := DefaultKeyGenerator
	if len(keyGen) > 0 {
		getKey = keyGen[0]
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			// Only cache GET requests
			if c.Request.Method != http.MethodGet {
				return next(c)
			}

			key := getKey(c)

			// Check cache
			if val, ok := store.Get(key); ok {
				if cachedBytes, ok := val.([]byte); ok {
					var cached CachedResponse
					// Use Gob for simple serialization of struct with headers
					buf := bytes.NewBuffer(cachedBytes)
					if err := gob.NewDecoder(buf).Decode(&cached); err == nil {
						// Replay headers
						for k, v := range cached.Headers {
							for _, h := range v {
								c.Writer.Header().Add(k, h)
							}
						}
						c.Writer.Header().Set("X-Cache", "HIT")
						c.Writer.WriteHeader(cached.StatusCode)
						c.Writer.Write(cached.Body)
						return nil
					}
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
			if err == nil && recorder.statusCode < 400 {
				// Create cached response
				resp := CachedResponse{
					StatusCode: recorder.statusCode,
					Headers:    recorder.Header().Clone(), // Copy headers
					Body:       recorder.body.Bytes(),
				}

				var buf bytes.Buffer
				if err := gob.NewEncoder(&buf).Encode(resp); err == nil {
					store.Set(key, buf.Bytes(), ttl)
				} else {
					fmt.Println("Cache encode error:", err)
				}
			}

			return err
		}
	}
}
