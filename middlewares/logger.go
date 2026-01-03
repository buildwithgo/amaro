package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/buildwithgo/amaro"
)

// ANSI color codes
const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type LoggerOption func(*loggerConfig)
type LoggerPrintFunc func(logger *log.Logger, duration time.Duration, c *amaro.Context, statusCode int)

type loggerConfig struct {
	logger    *log.Logger
	printFunc LoggerPrintFunc
}

func WithLogger(logger *log.Logger) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.logger = logger
	}
}

func WithLoggerLogFunc(logFunc LoggerPrintFunc) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.printFunc = logFunc
	}
}

// statusColor returns the ANSI color code for a given HTTP status code.
func statusColor(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// methodColor returns the ANSI color code for a given HTTP method.
func methodColor(method string) string {
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func Logger(opts ...LoggerOption) amaro.Middleware {
	cfg := &loggerConfig{
		logger: log.Default(),
		printFunc: func(logger *log.Logger, duration time.Duration, c *amaro.Context, statusCode int) {
			statusColor := statusColor(statusCode)
			methodColor := methodColor(c.Request.Method)
			resetColor := reset

			// Format: [STATUS] METHOD PATH - LATENCY
			// Example: [200] GET /users - 12ms
			logMsg := fmt.Sprintf("%s %3d %s %s %s %s %s %s",
				statusColor, statusCode, resetColor,
				methodColor, c.Request.Method, resetColor,
				c.Request.URL.Path,
				duration,
			)
			logger.Println(logMsg)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			start := time.Now()
			// Default status code to 200, if not written it stays 200 usually.
			// But we need to capture the status code.
			// amaro.Context doesn't expose the status code written unless we wrap the writer.
			// Since amaro is zero-allocation focused, wrapping might be tricky without allocating.
			// However, typically a logger wraps the writer.
			// Let's check if amaro.Context has a way to get the status.
			// It doesn't seem so from my read of context.go.
			// So I need to wrap the response writer.

			// Simple ResponseWriter wrapper to capture status code
			lrw := &loggingResponseWriter{ResponseWriter: c.Writer, statusCode: http.StatusOK}
			c.Writer = lrw

			err := next(c)
			duration := time.Since(start)

			// If error returned and not handled, status might be different?
			// Typically middleware is outer layer.
			// If next(c) returns error, the error handler eventually handles it.
			// But the logger middleware is executed *around* next(c).
			// If an error bubbles up, we might miss the final status code set by the error handler
			// UNLESS the error handler is inside the chain or we handle it here.
			// But usually Logger is the first middleware.

			cfg.printFunc(cfg.logger, duration, c, lrw.statusCode)
			return err
		}
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
