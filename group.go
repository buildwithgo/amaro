package amaro

import "net/http"

type Group struct {
	prefix      string
	router      Router
	middlewares []Middleware
}

func NewGroup(prefix string, router Router) *Group {
	return &Group{
		prefix:      prefix,
		router:      router,
		middlewares: make([]Middleware, 0),
	}
}

func (g *Group) wrapWithMiddlewares(handler Handler) Handler {
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		handler = g.middlewares[i](handler)
	}
	return handler
}

func (g *Group) Use(middleware Middleware) {
	g.middlewares = append(g.middlewares, middleware)
}

func (g *Group) Add(method, path string, handler Handler, middlewares ...Middleware) error {
	path = g.prefix + path
	if len(middlewares) > 0 {
		handler = g.wrapWithMiddlewares(handler)
		middlewares = append(g.middlewares, middlewares...)
	} else {
		middlewares = g.middlewares
	}
	return g.router.GET(path, handler, middlewares...)
}

func (g *Group) GET(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodGet, path, handler, middlewares...)
}

func (g *Group) POST(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodPost, path, handler, middlewares...)
}

func (g *Group) PUT(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodPut, path, handler, middlewares...)
}

func (g *Group) DELETE(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodDelete, path, handler, middlewares...)
}

func (g *Group) PATCH(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodPatch, path, handler, middlewares...)
}

func (g *Group) OPTIONS(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodOptions, path, handler, middlewares...)
}

func (g *Group) HEAD(path string, handler Handler, middlewares ...Middleware) error {
	return g.Add(http.MethodHead, path, handler, middlewares...)
}

func (g *Group) Group(prefix string) *Group {
	return NewGroup(g.prefix+prefix, g.router)
}

func (g *Group) Find(method, path string) (*Route, error) {
	path = g.prefix + path
	return g.router.Find(method, path)
}
