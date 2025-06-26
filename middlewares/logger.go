package middlewares

import (
	"fmt"
	"time"

	"github.com/buildwithgo/amaro"
)

func Logger() amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			fmt.Printf("[%s] %s %s - %s %v\n",
				time.Now().Format("2006-01-02 15:04:05"),
				c.Request.Method,
				c.Request.URL.Path,
				duration,
				err,
			)
			return err
		}
	}
}
