package amaro

import (
	"fmt"
	"net/http"
	"runtime"
)

// Recovery recovers from panics, logs the stack trace, and returns an Internal Server Error.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 4096)
					n := runtime.Stack(stack, false)
					stackTrace := string(stack[:n])

					fmt.Printf("panic: %v\nStack trace:\n%s\n", err, stackTrace)

					c.String(http.StatusInternalServerError, "Internal Server Error")
				}
			}()
			return next(c)
		}
	}
}
