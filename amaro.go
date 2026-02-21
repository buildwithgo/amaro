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

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Handler is a function that handles an HTTP request.
// It returns an error which can be handled by middlewares or the framework.
type Handler func(*Context) error

// Middleware is a function that wraps a Handler to provide additional functionality.
type Middleware func(next Handler) Handler

// ErrorHandler is a function that handles errors occurred during request processing.
type ErrorHandler func(c *Context, err error, code int)

// ServerConfig holds the configuration for the HTTP server.
type ServerConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	EnableH2C         bool
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
}

// App is the main entry point for the Amaro framework.
// It holds the router, global middlewares, and a context pool.
type App struct {
	router       Router
	middlewares  []Middleware
	pool         *sync.Pool
	handler      Handler
	once         sync.Once
	errorHandler ErrorHandler
	serverConfig ServerConfig
}

// WithErrorHandler returns an AppOption that configures the App to use the specified ErrorHandler.
func WithErrorHandler(handler ErrorHandler) AppOption {
	return func(app *App) {
		app.errorHandler = handler
	}
}

// WithServerConfig returns an AppOption that configures the App to use the specified ServerConfig.
func WithServerConfig(config ServerConfig) AppOption {
	return func(app *App) {
		app.serverConfig = config
	}
}

// WithReadHeaderTimeout sets the ReadHeaderTimeout for the HTTP server.
func WithReadHeaderTimeout(timeout time.Duration) AppOption {
	return func(app *App) {
		app.serverConfig.ReadHeaderTimeout = timeout
	}
}

// WithReadTimeout sets the ReadTimeout for the HTTP server.
func WithReadTimeout(timeout time.Duration) AppOption {
	return func(app *App) {
		app.serverConfig.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the WriteTimeout for the HTTP server.
func WithWriteTimeout(timeout time.Duration) AppOption {
	return func(app *App) {
		app.serverConfig.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the IdleTimeout for the HTTP server.
func WithIdleTimeout(timeout time.Duration) AppOption {
	return func(app *App) {
		app.serverConfig.IdleTimeout = timeout
	}
}

// WithMaxHeaderBytes sets the MaxHeaderBytes for the HTTP server.
func WithMaxHeaderBytes(maxBytes int) AppOption {
	return func(app *App) {
		app.serverConfig.MaxHeaderBytes = maxBytes
	}
}

// WithH2C enables HTTP/2 Cleartext (H2C) support.
// This is useful for backend services behind a proxy that terminates TLS.
func WithH2C() AppOption {
	return func(app *App) {
		app.serverConfig.EnableH2C = true
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

// Any registers a route that matches all standard HTTP methods.
func (a *App) Any(path string, handler Handler, middlewares ...Middleware) error {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
	}
	for _, method := range methods {
		if err := a.Add(method, path, handler, middlewares...); err != nil {
			return err
		}
	}
	return nil
}

// Mount registers an http.Handler (e.g., grpc-gateway mux) at the specified path prefix.
// It registers the handler for all standard HTTP methods for the exact path and all subpaths.
func (a *App) Mount(path string, handler http.Handler) error {
	h := WrapHTTPHandler(handler)

	// Exact match
	if err := a.Any(path, h); err != nil {
		return err
	}

	// Wildcard match for subpaths
	wildcardPath := path
	if !strings.HasSuffix(wildcardPath, "/") {
		wildcardPath += "/"
	}
	wildcardPath += "*filepath"

	return a.Any(wildcardPath, h)
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

	// Determine the handler (wrap in H2C if enabled)
	var handler http.Handler = a
	if a.serverConfig.EnableH2C {
		h2s := &http2.Server{}
		handler = h2c.NewHandler(a, h2s)
	}

	srv := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: a.serverConfig.ReadHeaderTimeout,
		ReadTimeout:       a.serverConfig.ReadTimeout,
		WriteTimeout:      a.serverConfig.WriteTimeout,
		IdleTimeout:       a.serverConfig.IdleTimeout,
		MaxHeaderBytes:    a.serverConfig.MaxHeaderBytes,
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

// WrapHTTPHandler converts a standard http.Handler to an amaro.Handler.
func WrapHTTPHandler(h http.Handler) Handler {
	return func(c *Context) error {
		h.ServeHTTP(c.Writer, c.Request)
		return nil
	}
}
