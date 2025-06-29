package middlewares

import (
	"log"
	"time"

	"github.com/buildwithgo/amaro"
)

type LoggerOption func(*loggerConfig)
type LoggerPrintFunc func(logger *log.Logger, duration time.Duration, c *amaro.Context)

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

func Logger(opts ...LoggerOption) amaro.Middleware {
	cfg := &loggerConfig{
		logger: log.Default(),
		printFunc: func(logger *log.Logger, duration time.Duration, c *amaro.Context) {
			logger.Printf("%s %s - %s  \n",
				c.Request.Method,
				c.Request.URL.Path,
				duration,
			)
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)
			cfg.printFunc(cfg.logger, duration, c)
			return err
		}
	}
}
