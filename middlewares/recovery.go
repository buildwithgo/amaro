package middlewares

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/buildwithgo/amaro"
)

// Recovery recovers from panics, logs the stack trace, and returns an Internal Server Error.
func Recovery() amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
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
