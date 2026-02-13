package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/buildwithgo/amaro"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type bufferedWriter struct {
	http.ResponseWriter
	header http.Header
	code   int
	buf    *bytes.Buffer
}

func (w *bufferedWriter) Header() http.Header {
	return w.header
}

func (w *bufferedWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.buf.Write(b)
}

func (w *bufferedWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

// Timeout middleware cancels the context if the request processing time exceeds the given duration.
func Timeout(timeout time.Duration) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()

			c.Request = c.Request.WithContext(ctx)

			// Get buffer from pool
			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()

			// Use buffered writer
			bw := &bufferedWriter{
				ResponseWriter: c.Writer,
				header:         make(http.Header),
				buf:            buf,
			}
			originalWriter := c.Writer

			// We clone the context for the goroutine because `c` will be reset if we return early.
			cClone := c.Clone()
			cClone.Writer = bw

			done := make(chan error, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						// Recover from panic to prevent crashing the server
						// We send an error to the channel so the main request loop can handle it
						// Note: The stack trace is lost here unless we log it or include it in error.
						// For now, we return a generic panic error.
						// Ideally, use a logger here.
						select {
						case done <- fmt.Errorf("panic in timeout handler: %v", r):
						default:
							// Channel might be full if next() returned and then we panicked?
							// Unlikely in this structure.
						}
					}
				}()
				// Execute the handler with the cloned context
				done <- next(cClone)
			}()

			select {
			case err := <-done:
				if err != nil {
					// If error occurred (or panic), discard buffer and return error
					// ensuring upstream error handler can write the response.
					bufferPool.Put(buf)
					c.Writer = originalWriter
					return err
				}

				// Success: copy buffer to original writer

				// Copy headers
				for k, vv := range bw.header {
					for _, v := range vv {
						originalWriter.Header().Add(k, v)
					}
				}

				code := bw.code
				if code == 0 {
					code = http.StatusOK
				}
				originalWriter.WriteHeader(code)
				originalWriter.Write(buf.Bytes())

				// Return buffer to pool
				bufferPool.Put(buf)

				return nil

			case <-ctx.Done():
				// Timeout
				// Do not return buffer to pool as the goroutine might still use it
				c.String(http.StatusServiceUnavailable, "Service Unavailable")
				return ctx.Err()
			}
		}
	}
}
