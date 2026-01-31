package amaro

import (
	"io/fs"
	"net/http"
	"strings"
)

// Group represents a route group with a common path prefix and shared middlewares.
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

// Use adds a middleware to the group.
// These middlewares are applied to all routes registered in this group.
func (g *Group) Use(middleware Middleware) {
	g.middlewares = append(g.middlewares, middleware)
}

func (g *Group) Add(method, path string, handler Handler, middlewares ...Middleware) error {
	var fullPath strings.Builder
	fullPath.Grow(len(g.prefix) + len(path)) // Pre-allocate capacity
	fullPath.WriteString(g.prefix)
	fullPath.WriteString(path)
	return g.router.Add(method, fullPath.String(), handler, middlewares...)
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

// Any registers a route that matches all standard HTTP methods.
func (g *Group) Any(path string, handler Handler, middlewares ...Middleware) error {
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
		if err := g.Add(method, path, handler, middlewares...); err != nil {
			return err
		}
	}
	return nil
}

// Mount registers an http.Handler (e.g., grpc-gateway mux) at the specified path prefix within the group.
// It registers the handler for all standard HTTP methods for the exact path and all subpaths.
func (g *Group) Mount(path string, handler http.Handler) error {
	h := WrapHTTPHandler(handler)

	// Exact match
	if err := g.Any(path, h); err != nil {
		return err
	}

	// Wildcard match for subpaths
	wildcardPath := path
	if !strings.HasSuffix(wildcardPath, "/") {
		wildcardPath += "/"
	}
	wildcardPath += "*filepath"

	return g.Any(wildcardPath, h)
}

func (g *Group) Group(prefix string) *Group {
	return NewGroup(g.prefix+prefix, g.router)
}

func (g *Group) Find(method, path string) (*Route, error) {
	return g.router.Find(method, g.calculatePath(path), nil)
}

func (g *Group) StaticFS(pathPrefix string, fs fs.FS) {
	g.router.StaticFS(g.calculatePath(pathPrefix), fs)
}

func (g *Group) calculatePath(path string) string {
	var fullPath strings.Builder
	fullPath.Grow(len(g.prefix) + len(path))
	fullPath.WriteString(g.prefix)
	fullPath.WriteString(path)
	return fullPath.String()
}
