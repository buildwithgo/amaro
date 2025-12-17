package amaro

import "io/fs"

// Route represents a registered route.
type Route struct {
	Method      string
	Path        string
	Handler     Handler
	Middlewares []Middleware
}

// Router is the interface that all router implementations must satisfy.
// It allows for swappable routing strategies.
type Router interface {
	GET(path string, handler Handler, middlewares ...Middleware) error
	POST(path string, handler Handler, middlewares ...Middleware) error
	PUT(path string, handler Handler, middlewares ...Middleware) error
	DELETE(path string, handler Handler, middlewares ...Middleware) error
	PATCH(path string, handler Handler, middlewares ...Middleware) error
	OPTIONS(path string, handler Handler, middlewares ...Middleware) error
	HEAD(path string, handler Handler, middlewares ...Middleware) error
	Add(method, path string, handler Handler, middlewares ...Middleware) error
	Use(middleware Middleware)
	Group(prefix string) *Group
	Find(method, path string, ctx *Context) (*Route, error)
	StaticFS(pathPrefix string, fs fs.FS)
}

// WithRouter returns an AppOption that configures the App to use the specified router.
func WithRouter(router Router) AppOption {
	return func(app *App) {
		app.router = router
	}
}
