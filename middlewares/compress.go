package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.Header().Del("Content-Length") // Content-length is no longer valid after compression
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Flush() {
	if f, ok := w.Writer.(*gzip.Writer); ok {
		f.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Compress returns a middleware that compresses HTTP responses using Gzip.
func Compress() amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
				return next(c)
			}

			// Set Header
			c.Writer.Header().Set("Content-Encoding", "gzip")
			c.Writer.Header().Set("Vary", "Accept-Encoding")

			gz := gzip.NewWriter(c.Writer)
			defer gz.Close()

			gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}

			// Temporarily replace writer
			originalWriter := c.Writer
			c.Writer = gzw

			err := next(c)

			// Restore
			c.Writer = originalWriter
			return err
		}
	}
}
