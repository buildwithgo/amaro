package amaro

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"strings"
)

// RecoveryOption configures the Recovery middleware.
type RecoveryOption func(*recoveryConfig)

type recoveryConfig struct {
	htmlDebug bool
}

// WithHTMLDebug enables rendering a pretty HTML debug page for panics.
// WARNING: Do not use this in production as it exposes stack traces.
func WithHTMLDebug(enabled bool) RecoveryOption {
	return func(c *recoveryConfig) {
		c.htmlDebug = enabled
	}
}

// Recovery recovers from panics, logs the stack trace, and returns an Internal Server Error.
func Recovery(opts ...RecoveryOption) Middleware {
	cfg := &recoveryConfig{htmlDebug: false}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next Handler) Handler {
		return func(c *Context) error {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 4096)
					n := runtime.Stack(stack, false)
					stackTrace := string(stack[:n])

					fmt.Printf("panic: %v\nStack trace:\n%s\n", err, stackTrace)

					if cfg.htmlDebug {
						c.HTML(http.StatusInternalServerError, renderDebugPage(err, stackTrace))
					} else {
						c.String(http.StatusInternalServerError, "Internal Server Error")
					}
				}
			}()
			return next(c)
		}
	}
}

func renderDebugPage(err interface{}, stack string) string {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Internal Server Error - Amaro</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #f8f9fa; color: #212529; margin: 0; padding: 2rem; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        h1 { color: #dc3545; border-bottom: 2px solid #eee; padding-bottom: 0.5rem; }
        .error-message { font-size: 1.25rem; font-weight: bold; margin-bottom: 1rem; color: #343a40; }
        pre { background: #212529; color: #f8f9fa; padding: 1rem; border-radius: 4px; overflow-x: auto; font-size: 0.9rem; line-height: 1.5; }
        .footer { margin-top: 2rem; font-size: 0.875rem; color: #6c757d; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Internal Server Error</h1>
        <div class="error-message">Panic: {{.Error}}</div>
        <h3>Stack Trace:</h3>
        <pre>{{.Stack}}</pre>
        <div class="footer">Amaro Framework Debugger</div>
    </div>
</body>
</html>
`
	data := struct {
		Error interface{}
		Stack string
	}{
		Error: err,
		Stack: stack,
	}

	t, parseErr := template.New("debug").Parse(tmpl)
	if parseErr != nil {
		return "Internal Server Error (Failed to render debug page)"
	}

	var buf strings.Builder
	if execErr := t.Execute(&buf, data); execErr != nil {
		return "Internal Server Error (Failed to execute debug template)"
	}

	return buf.String()
}
