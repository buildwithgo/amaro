package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/buildwithgo/amaro"
)

// Timeout middleware cancels the context if the request processing time exceeds the given duration.
func Timeout(timeout time.Duration) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()

			// Update request with new context
			c.Request = c.Request.WithContext(ctx)

			done := make(chan error, 1)

			go func() {
				// We need to be careful with concurrency here.
				// The next handler typically writes to c.Writer.
				// If we timeout, we shouldn't write anymore from this goroutine ideally,
				// but stdlib http.ResponseWriter is not thread safe.
				// For a simple middleware without buffering, we rely on the handler checking ctx.Done().
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				// Timeout
				c.Writer.WriteHeader(http.StatusServiceUnavailable)
				return ctx.Err()
			}
		}
	}
}
