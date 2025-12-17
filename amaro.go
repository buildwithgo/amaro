// Package amaro implements a blazing fast, zero-dependency, zero-allocation HTTP router and framework for Go.
package amaro

import (
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

// Handler is a function that handles an HTTP request.
// It returns an error which can be handled by middlewares or the framework.
type Handler func(*Context) error

// Middleware is a function that wraps a Handler to provide additional functionality.
type Middleware func(next Handler) Handler

// App is the main entry point for the Amaro framework.
// It holds the router, global middlewares, and a context pool.
type App struct {
	router      Router
	middlewares []Middleware
	pool        *sync.Pool
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

func (a *App) Find(method, path string) (*Route, error) {
	return a.router.Find(method, path, nil)
}

// AppOption defines a function to configure the App during initialization.
type AppOption func(*App)

// New creates a new instance of the Amaro App with optional configuration.
func New(options ...AppOption) *App {
	app := &App{
		middlewares: make([]Middleware, 0),
		pool: &sync.Pool{
			New: func() interface{} {
				// We can't fully init here because we need w/r, but we create the struct
				// The slice capacity is set in context.go
				return NewContext(nil, nil)
			},
		},
	}

	for _, option := range options {
		option(app)
	}

	return app
}

func (a *App) Run(port string) error {
	compiledMiddlewares := Chain(a.middlewares...)
	a.middlewares = []Middleware{compiledMiddlewares}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return http.ListenAndServe(port, a)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := a.pool.Get().(*Context)
	ctx.Reset(w, r)
	defer a.pool.Put(ctx)

	// Pass ctx to Find so it can populate params without allocation
	route, err := a.router.Find(r.Method, r.URL.Path, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// route.Middlewares are already compiled into route.Handler
	// We only need to apply global middlewares
	if err := Compile(route.Handler, a.middlewares...)(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func Compile(hendler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		hendler = middlewares[i](hendler)
	}
	return hendler
}
