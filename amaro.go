package amaro

import (
	"net/http"
	"strings"
)

type Handler func(*Context) error

type Middleware func(*Context, func() error) error

type App struct {
	router      Router
	middlewares []Middleware
}

func (a *App) Use(middleware Middleware) {
	a.middlewares = append(a.middlewares, middleware)
}

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

func (a *App) Add(method, path string, handler Handler, middlewares ...Middleware) error {
	return a.router.Add(method, path, handler, middlewares...)
}

func (a *App) Group(prefix string) *Group {
	return a.router.Group(prefix)
}

func (a *App) Find(method, path string) (*Route, error) {
	return a.router.Find(method, path)
}

type AppOption func(*App)

// New creates a new Amaro app instance
func New(options ...AppOption) *App {
	app := &App{
		middlewares: make([]Middleware, 0),
	}

	for _, option := range options {
		option(app)
	}

	return app
}

func (a *App) Run(port string) error {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return http.ListenAndServe(port, a)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := Context{
		Request: r,
		Writer:  w,
	}
	route, err := a.router.Find(r.Method, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- a.executeWithMiddlewares(&ctx, route.Handler)
	}()

	if err := <-done; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) executeWithMiddlewares(ctx *Context, handler Handler) error {
	index := 0
	var next func() error
	next = func() error {
		if index >= len(a.middlewares) {
			return handler(ctx)
		}
		middleware := a.middlewares[index]
		index++
		return middleware(ctx, next)
	}
	return next()
}
