package routers

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

type node struct {
	children map[string]*node
	amaro.Route
}

// TrieRouter is a trie-based router using a map for children.
// It supports :param and *wildcard parameters.
type TrieRouter struct {
	root map[string]*node // method -> root node
}

// NewTrieRouter creates a new instance of TrieRouter.
func NewTrieRouter() *TrieRouter {
	return &TrieRouter{
		root: make(map[string]*node),
	}
}

// Use adds a global middleware to the router.
// Note: In this framework, global middlewares are typically handled by App.
// This method is provided to satisfy the Router interface.
func (r *TrieRouter) Use(middleware amaro.Middleware) {
	// No-op for now
}

func (r *TrieRouter) Add(method, path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
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
			if n.children == nil {
				n.children = make(map[string]*node)
			}
			if _, ok := n.children[part]; !ok {
				n.children[part] = &node{children: make(map[string]*node)}
			}
			n = n.children[part]
		}
	}

	// Compile middlewares into handler
	finalHandler := handler
	if len(middlewares) > 0 {
		finalHandler = amaro.Compile(handler, middlewares...)
	}

	n.Handler = finalHandler
	n.Middlewares = middlewares // Store for introspection if needed
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
	// Remove leading slash for processing
	if len(searchPath) > 0 && searchPath[0] == '/' {
		searchPath = searchPath[1:]
	}
	// Remove trailing slash if needed?
	if len(searchPath) > 0 && searchPath[len(searchPath)-1] == '/' {
		searchPath = searchPath[:len(searchPath)-1]
	}

	// Zero-allocation iteration over parts
	for len(searchPath) > 0 || n != nil {
		// If we consumed the whole path
		if len(searchPath) == 0 {
			if n.Handler == nil {
				// Check for * child
				if n.children != nil {
					for key, child := range n.children {
						if key[0] == '*' {
							if ctx != nil {
								// Match rest (empty)
								ctx.AddParam(key[1:], "")
							}
							return &child.Route, nil
						}
					}
				}

				// Return handler if exists (exact match)
				if n.Handler != nil {
					return &n.Route, nil
				}
				return nil, fmt.Errorf("route not found")
			}
			return &n.Route, nil
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

		// Look for child
		if child, found := n.children[part]; found {
			n = child
			continue
		}

		// Check for dynamic children (param or wildcard)
		matched := false
		for key, child := range n.children {
			// Check for :param
			if key[0] == ':' || (len(key) > 1 && key[0] == '{' && key[len(key)-1] == '}') {
				if ctx != nil {
					paramName := key[1:]
					if key[0] == '{' {
						paramName = key[1 : len(key)-1]
					}
					ctx.AddParam(paramName, part)
				}
				n = child
				matched = true
				break
			}

			// Check for *wildcard
			if key[0] == '*' {
				if ctx != nil {
					value := part
					if len(searchPath) > 0 {
						value += "/" + searchPath
					}
					ctx.AddParam(key[1:], value)
				}
				return &child.Route, nil
			}
		}

		if !matched {
			return nil, fmt.Errorf("route not found")
		}
	}
	return nil, fmt.Errorf("route not found")
}

func (r *TrieRouter) StaticFS(pathPrefix string, fsys fs.FS) {
	fileServer := http.FileServer(http.FS(fsys))
	handler := func(c *amaro.Context) error {
		http.StripPrefix(pathPrefix, fileServer).ServeHTTP(c.Writer, c.Request)
		return nil
	}

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
