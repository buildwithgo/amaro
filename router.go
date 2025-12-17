package amaro

type Route struct {
	Method      string
	Path        string
	Handler     Handler
	Middlewares []Middleware
}

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
}

func WithRouter(router Router) AppOption {
	return func(app *App) {
		app.router = router
	}
}
