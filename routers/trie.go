package routers

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

type node struct {
	// Static children
	children map[string]*node

	// Dynamic children
	paramNode    *node
	paramName    string

	catchAllNode *node
	catchAllName string

	amaro.Route
}

// ParamParser defines a function that checks if a path segment is a parameter.
// It returns true and the parameter name if it is, false otherwise.
type ParamParser func(segment string) (bool, string)

// WildcardParser defines a function that checks if a path segment is a wildcard.
// It returns true and the wildcard name if it is, false otherwise.
type WildcardParser func(segment string) (bool, string)

// TrieRouterConfig defines configuration for TrieRouter.
type TrieRouterConfig struct {
	ParamParser    ParamParser
	WildcardParser WildcardParser
}

// DefaultParamParser implements the standard :param and {param} syntax.
func DefaultParamParser(segment string) (bool, string) {
	if len(segment) > 0 && segment[0] == ':' {
		return true, segment[1:]
	}
	if len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}' {
		return true, segment[1 : len(segment)-1]
	}
	return false, ""
}

// DefaultWildcardParser implements the standard *wildcard syntax.
func DefaultWildcardParser(segment string) (bool, string) {
	if len(segment) > 0 && segment[0] == '*' {
		return true, segment[1:]
	}
	return false, ""
}

// DefaultTrieRouterConfig returns the default configuration.
func DefaultTrieRouterConfig() TrieRouterConfig {
	return TrieRouterConfig{
		ParamParser:    DefaultParamParser,
		WildcardParser: DefaultWildcardParser,
	}
}

// TrieRouter is a trie-based router using a map for children.
// It supports :param and *wildcard parameters.
type TrieRouter struct {
	root              map[string]*node // method -> root node
	globalMiddlewares []amaro.Middleware
	config            TrieRouterConfig
}

// TrieRouterOption configures TrieRouter.
type TrieRouterOption func(*TrieRouter)

// WithConfig sets the router configuration.
func WithConfig(config TrieRouterConfig) TrieRouterOption {
	return func(r *TrieRouter) {
		r.config = config
	}
}

// NewTrieRouter creates a new instance of TrieRouter.
func NewTrieRouter(opts ...TrieRouterOption) *TrieRouter {
	r := &TrieRouter{
		root:   make(map[string]*node),
		config: DefaultTrieRouterConfig(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Use adds a global middleware to the router.
// Note: These middlewares are applied to all routes registered AFTER calling Use.
// They are wrapped around the handler in Add.
func (r *TrieRouter) Use(middleware amaro.Middleware) {
	r.globalMiddlewares = append(r.globalMiddlewares, middleware)
}

func (r *TrieRouter) Add(method, path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	// Prepend router-level middlewares to the route-specific middlewares
	if len(r.globalMiddlewares) > 0 {
		combined := make([]amaro.Middleware, 0, len(r.globalMiddlewares)+len(middlewares))
		combined = append(combined, r.globalMiddlewares...)
		combined = append(combined, middlewares...)
		middlewares = combined
	}
	if _, ok := r.root[method]; !ok {
		r.root[method] = &node{children: make(map[string]*node)}
	}
	n := r.root[method]

	// Normalize path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	searchPath := strings.Trim(path, "/")

	if searchPath != "" {
		parts := strings.Split(searchPath, "/")
		for _, part := range parts {
			if part == "" {
				continue
			}

			// Use configured parsers
			isParam, paramName := false, ""
			if r.config.ParamParser != nil {
				isParam, paramName = r.config.ParamParser(part)
			}

			isWildcard, wildcardName := false, ""
			if !isParam && r.config.WildcardParser != nil {
				isWildcard, wildcardName = r.config.WildcardParser(part)
			}

			if isParam {
				if n.paramNode == nil {
					n.paramNode = &node{children: make(map[string]*node)}
					n.paramName = paramName
				}
				if n.paramName != paramName {
					return fmt.Errorf("param name conflict: %s vs %s", n.paramName, paramName)
				}
				n = n.paramNode
			} else if isWildcard {
				if n.catchAllNode == nil {
					n.catchAllNode = &node{children: make(map[string]*node)}
					n.catchAllName = wildcardName
				}
				if n.catchAllName != wildcardName {
					return fmt.Errorf("wildcard name conflict: %s vs %s", n.catchAllName, wildcardName)
				}
				n = n.catchAllNode
			} else {
				// Static
				if n.children == nil {
					n.children = make(map[string]*node)
				}
				if _, ok := n.children[part]; !ok {
					n.children[part] = &node{children: make(map[string]*node)}
				}
				n = n.children[part]
			}
		}
	}

	// Compile middlewares into handler
	finalHandler := handler
	if len(middlewares) > 0 {
		finalHandler = amaro.Compile(handler, middlewares...)
	}

	n.Handler = finalHandler
	n.Middlewares = middlewares
	n.Path = path
	n.Method = method

	return nil
}

func (r *TrieRouter) Find(method, path string, ctx *amaro.Context) (*amaro.Route, error) {
	n, ok := r.root[method]
	if !ok {
		return nil, fmt.Errorf("method not found")
	}

	searchPath := path
	if len(searchPath) > 0 && searchPath[0] == '/' {
		searchPath = searchPath[1:]
	}
	if len(searchPath) > 0 && searchPath[len(searchPath)-1] == '/' {
		searchPath = searchPath[:len(searchPath)-1]
	}

	// Zero-allocation iteration
	for len(searchPath) > 0 || n != nil {
		if len(searchPath) == 0 {
			if n.Handler != nil {
				return &n.Route, nil
			}
			if n.catchAllNode != nil {
				if ctx != nil {
					ctx.AddParam(n.catchAllName, "")
				}
				if n.catchAllNode.Handler != nil {
					return &n.catchAllNode.Route, nil
				}
			}
			return nil, amaro.NewHTTPError(http.StatusNotFound, "route not found")
		}

		var part string
		i := strings.IndexByte(searchPath, '/')
		if i < 0 {
			part = searchPath
			searchPath = ""
		} else {
			part = searchPath[:i]
			searchPath = searchPath[i+1:]
		}

		if part == "" {
			continue
		}

		// Priority: Static > Param > Wildcard

		// 1. Static
		if child, found := n.children[part]; found {
			n = child
			continue
		}

		// 2. Param
		if n.paramNode != nil {
			if ctx != nil {
				ctx.AddParam(n.paramName, part)
			}
			n = n.paramNode
			continue
		}

		// 3. CatchAll
		if n.catchAllNode != nil {
			if ctx != nil {
				value := part
				if len(searchPath) > 0 {
					value += "/" + searchPath
				}
				ctx.AddParam(n.catchAllName, value)
			}
			if n.catchAllNode.Handler != nil {
				return &n.catchAllNode.Route, nil
			}
			return nil, amaro.NewHTTPError(http.StatusNotFound, "route not found")
		}

		return nil, amaro.NewHTTPError(http.StatusNotFound, "route not found")
	}

	return nil, amaro.NewHTTPError(http.StatusNotFound, "route not found")
}

func (r *TrieRouter) StaticFS(pathPrefix string, fsys fs.FS) {
	handler := amaro.StaticHandler(amaro.StaticConfig{
		Root:   fsys,
		Prefix: pathPrefix,
	})

	path := strings.TrimRight(pathPrefix, "/")
	r.Add(http.MethodGet, path, handler)
	r.Add(http.MethodHead, path, handler)

	wildcardPath := path + "/*filepath"
	r.Add(http.MethodGet, wildcardPath, handler)
	r.Add(http.MethodHead, wildcardPath, handler)
}

func (r *TrieRouter) GET(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodGet, path, handler, middlewares...)
}
func (r *TrieRouter) POST(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodPost, path, handler, middlewares...)
}
func (r *TrieRouter) PUT(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodPut, path, handler, middlewares...)
}
func (r *TrieRouter) DELETE(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodDelete, path, handler, middlewares...)
}
func (r *TrieRouter) PATCH(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodPatch, path, handler, middlewares...)
}
func (r *TrieRouter) OPTIONS(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodOptions, path, handler, middlewares...)
}
func (r *TrieRouter) HEAD(path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	return r.Add(http.MethodHead, path, handler, middlewares...)
}
func (r *TrieRouter) Group(prefix string) *amaro.Group {
	return amaro.NewGroup(prefix, r)
}
