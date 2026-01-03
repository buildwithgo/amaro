// Package amaro implements a blazing fast, zero-dependency, zero-allocation HTTP router and framework for Go.
package amaro

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Handler is a function that handles an HTTP request.
// It returns an error which can be handled by middlewares or the framework.
type Handler func(*Context) error

// Middleware is a function that wraps a Handler to provide additional functionality.
type Middleware func(next Handler) Handler

// ErrorHandler is a function that handles errors occurred during request processing.
type ErrorHandler func(c *Context, err error, code int)

// App is the main entry point for the Amaro framework.
// It holds the router, global middlewares, and a context pool.
type App struct {
	router       Router
	middlewares  []Middleware
	pool         *sync.Pool
	handler      Handler
	once         sync.Once
	errorHandler ErrorHandler
}

// WithErrorHandler returns an AppOption that configures the App to use the specified ErrorHandler.
func WithErrorHandler(handler ErrorHandler) AppOption {
	return func(app *App) {
		app.errorHandler = handler
	}
}

// Use adds a global middleware to the application.
// Global middlewares are applied to all routes in the order they are added.
func (a *App) Use(middleware Middleware) {
	a.middlewares = append(a.middlewares, middleware)
}

// GET registers a new GET route with a handler and optional route-specific middlewares.
func (a *App) GET(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodGet, path, handler, middlewares...)
}

func (a *App) POST(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodPost, path, handler, middlewares...)
}

func (a *App) PUT(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodPut, path, handler, middlewares...)
}

func (a *App) DELETE(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodDelete, path, handler, middlewares...)
}

func (a *App) PATCH(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodPatch, path, handler, middlewares...)
}

func (a *App) OPTIONS(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodOptions, path, handler, middlewares...)
}

func (a *App) HEAD(path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(http.MethodHead, path, handler, middlewares...)
}

// Add registers a new route with the specified method, path, handler, and middlewares.
func (a *App) Add(method, path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(method, path, handler, middlewares...)
}

func (a *App) Group(prefix string) *Group {
	return a.router.Group(prefix)
}

func (a *App) StaticFS(pathPrefix string, fs fs.FS) {
	a.router.StaticFS(pathPrefix, fs)
}

// Static serves files from the local filesystem.
func (a *App) Static(pathPrefix, root string) {
	a.StaticFS(pathPrefix, os.DirFS(root))
}

func (a *App) Find(method, path string) (*Route, error) {
	return a.router.Find(method, path, nil)
}

// Test executes a request against the application and returns the response recorder.
// This is a helper for writing tests.
func (a *App) Test(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	a.ServeHTTP(w, req)
	return w
}

// AppOption defines a function to configure the App during initialization.
type AppOption func(*App)

// New creates a new instance of the Amaro App with optional configuration.
func New(options ...AppOption) *App {
	app := &App{
		middlewares: []Middleware{Recovery()}, // Add Recovery middleware by default
		pool: &sync.Pool{
			New: func() interface{} {
				// We can't fully init here because we need w/r, but we create the struct
				// The slice capacity is set in context.go
				return NewContext(nil, nil)
			},
		},
		errorHandler: func(c *Context, err error, code int) {
			if he, ok := err.(*HTTPError); ok {
				code = he.Code
				if msg, ok := he.Message.(string); ok {
					http.Error(c.Writer, msg, code)
				} else {
					http.Error(c.Writer, http.StatusText(code), code)
				}
				return
			}
			http.Error(c.Writer, err.Error(), code)
		},
	}

	for _, option := range options {
		option(app)
	}

	return app
}

// Run starts the HTTP server with graceful shutdown support.
func (a *App) Run(address string) error {
	return a.startServer(address, "", "")
}

// RunTLS starts the HTTPS server with graceful shutdown support.
func (a *App) RunTLS(address, certFile, keyFile string) error {
	return a.startServer(address, certFile, keyFile)
}

func (a *App) startServer(address, certFile, keyFile string) error {
	if !strings.HasPrefix(address, ":") {
		address = ":" + address
	}

	// We do NOT modify a.middlewares here anymore.
	// setup() will ignore any middlewares added after first compile if not careful,
	// but standard app lifecycle is: New -> Use... -> Run.
	// We just rely on Dispatch compiled in setup().

	srv := &http.Server{
		Addr:    address,
		Handler: a,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		a.setup() // Ensure middlewares are compiled before starting
		log.Printf("Server is starting on %s...", address)
		if certFile != "" && keyFile != "" {
			serverErrors <- srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			serverErrors <- srv.ListenAndServe()
		}
	}()

	// Buffered channel to receive OS signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received or an error occurs
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("shutdown started: signal %v", sig)

		// Create a context with a timeout for the shutdown process.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Ask the server to shut down gracefully.
		if err := srv.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure the handler chain is built (Lazy init for testing/direct usage)
	if a.handler == nil {
		a.setup()
	}

	ctx := a.pool.Get().(*Context)
	ctx.Reset(w, r)
	defer a.pool.Put(ctx)

	if err := a.handler(ctx); err != nil {
		a.errorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
}

func (a *App) setup() {
	a.once.Do(func() {
		// Compile the global middlewares with the router handler (dispatch)
		// This ensures that global middlewares run even if the route is not found
		a.handler = Compile(a.dispatch, a.middlewares...)
	})
}

func (a *App) dispatch(c *Context) error {
	// Pass ctx to Find so it can populate params without allocation
	route, err := a.router.Find(c.Request.Method, c.Request.URL.Path, c)
	if err != nil {
		a.errorHandler(c, err, http.StatusNotFound)
		return nil
	}
	// route.Middlewares are already compiled into route.Handler
	return route.Handler(c)
}

func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func Compile(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
