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

// TrieRouter is a trie-based router using a map for children.
// It supports :param and *wildcard parameters.
type TrieRouter struct {
	root              map[string]*node // method -> root node
	globalMiddlewares []amaro.Middleware
}

// NewTrieRouter creates a new instance of TrieRouter.
func NewTrieRouter() *TrieRouter {
	return &TrieRouter{
		root: make(map[string]*node),
	}
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
		// Create a new slice to avoid modifying the original middlewares slice if reusing
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

			// Check if it's a param or wildcard
			if part[0] == ':' || (len(part) > 1 && part[0] == '{' && part[len(part)-1] == '}') {
				// Param
				pName := part[1:]
				if part[0] == '{' {
					pName = part[1 : len(part)-1]
				}

				if n.paramNode == nil {
					n.paramNode = &node{children: make(map[string]*node)}
					n.paramName = pName
				}
				if n.paramName != pName {
					return fmt.Errorf("param name conflict: %s vs %s", n.paramName, pName)
				}
				n = n.paramNode
			} else if part[0] == '*' {
				// Wildcard
				wName := part[1:]
				if n.catchAllNode == nil {
					n.catchAllNode = &node{children: make(map[string]*node)}
					n.catchAllName = wName
				}
				if n.catchAllName != wName {
					return fmt.Errorf("wildcard name conflict: %s vs %s", n.catchAllName, wName)
				}
				n = n.catchAllNode
				// Wildcard must be the last element usually
				// We return/break?
				// If there are more parts after wildcard, it's weird but we'll allow adding to catchAllNode
				// effectively treating wildcard as a segment.
				// BUT standard wildcard matches EVERYTHING remaining.
				// So we should stop here?
				// If the user defines /files/*path/extra, it's ambiguous.
				// We assume *path captures everything. So we should NOT continue.
				// BUT checking the loop, we iterate parts.
				// If we continue, we add children to catchAllNode.
				// This implies *path matches one segment.
				// Trie usually: *path matches REST.
				// So this node should be the terminal handler node (mostly).
				// We will continue to allow defining children, but Find logic will decide.
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
		// If consumed whole path
		if len(searchPath) == 0 {
			// Exact match handler?
			if n.Handler != nil {
				return &n.Route, nil
			}
			// Check if we have a catchAll that matches empty?
			// Usually * matches rest. If rest is empty, it depends on implementation.
			// Hono: /a/* -> /a/ matches? Yes. /a matches? Maybe.
			if n.catchAllNode != nil {
				if ctx != nil {
					ctx.AddParam(n.catchAllName, "")
				}
				if n.catchAllNode.Handler != nil {
					return &n.catchAllNode.Route, nil
				}
			}
			return nil, fmt.Errorf("route not found")
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
				// Capture remaining path
				value := part
				if len(searchPath) > 0 {
					value += "/" + searchPath
				}
				ctx.AddParam(n.catchAllName, value)
			}
			// CatchAll consumes everything, so we are done traversing parts.
			// But we need to return the route from the child node.
			if n.catchAllNode.Handler != nil {
				return &n.catchAllNode.Route, nil
			}
			// If catchAllNode has no handler?
			// It should.
			return nil, fmt.Errorf("route not found")
		}

		return nil, fmt.Errorf("route not found")
	}

	return nil, fmt.Errorf("route not found")
}

func (r *TrieRouter) StaticFS(pathPrefix string, fsys fs.FS) {
	handler := amaro.StaticHandler(amaro.StaticConfig{
		Root:   fsys,
		Prefix: pathPrefix,
	})

	// Register pathPrefix (without trailing /) and pathPrefix/*filepath
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
