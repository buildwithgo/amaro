package routers

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/buildwithgo/amaro"
)

type trieNode struct {
	children map[string]*trieNode
	amaro.Route
}

type TrieRouter struct {
	root              map[string]*trieNode // method -> root node
	globalMiddlewares []amaro.Middleware
}

func NewTrieRouter() *TrieRouter {
	return &TrieRouter{
		root: make(map[string]*trieNode),
	}
}

func (r *TrieRouter) Use(mw amaro.Middleware) {
	r.globalMiddlewares = append(r.globalMiddlewares, mw)
}

func (r *TrieRouter) Add(method, path string, handler amaro.Handler, middlewares ...amaro.Middleware) error {
	if _, ok := r.root[method]; !ok {
		r.root[method] = &trieNode{children: make(map[string]*trieNode)}
	}
	node := r.root[method]
	path = strings.Trim(path, "/")
	if path != "" {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			if part == "" {
				continue
			}
			if _, ok := node.children[part]; !ok {
				node.children[part] = &trieNode{children: make(map[string]*trieNode)}
			}
			node = node.children[part]
		}
	}

	// Pre-compile middlewares into the handler to avoid per-request chain construction
	// Iterate backwards to wrap the handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	node.Handler = handler
	node.Middlewares = nil // Middlewares are now baked into the handler
	node.Middlewares = nil // Middlewares are now baked into the handler
	return nil
}

func (r *TrieRouter) StaticFS(pathPrefix string, fsys fs.FS) {
	// Create a handler that serves from fs
	fileServer := http.FileServer(http.FS(fsys))
	handler := func(c *amaro.Context) error {
		http.StripPrefix(pathPrefix, fileServer).ServeHTTP(c.Writer, c.Request)
		return nil
	}

	// Register GET/HEAD for pathPrefix/*
	// We use a wildcard route. We need to support it in Add/Find.
	// Convention: /assets/*filepath
	path := strings.TrimRight(pathPrefix, "/") + "/*filepath"
	r.Add(http.MethodGet, path, handler)
	r.Add(http.MethodHead, path, handler)
}

func (r *TrieRouter) findNode(method, path string, ctx *amaro.Context) (*trieNode, error) {
	node, ok := r.root[method]
	if !ok {
		return nil, fmt.Errorf("method not found")
	}

	searchPath := strings.Trim(path, "/")

	// Zero-alloc path iteration
	for len(searchPath) > 0 {
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

		if n, ok := node.children[part]; ok {
			node = n
		} else {
			matched := false
			for key, dyn := range node.children {
				if len(key) > 1 && key[0] == '{' && key[len(key)-1] == '}' {
					if ctx != nil {
						paramName := key[1 : len(key)-1]
						ctx.AddParam(paramName, part)
					}
					node = dyn
					matched = true
					break
				}
				// Wildcard check
				if key[0] == '*' {
					if ctx != nil {
						paramName := key[1:] // e.g. "filepath"
						// For wildcard, we want to match the Rest of the path?
						// But this loop iterates by parts.
						// Standard Trie wildcard matches until end.
						// We need to consume all remaining parts or change loop logic?
						// For zero-allocation, we can just say this node handles everything.
						// But findNode iterates parts.
						// If key is "*filepath", we should match ALL remaining path.
						// But standard trie logic usually puts wildcard at the end.

						// If we found a wildcard child, we break the loop and assume match.
						// But we need to verify if this logic holds for nested parts.
						// Typically `*` is a terminal node.

						// If part matches wildcard, we should append this part and all subsequent to param?
						// Or just return the node?
						// If we return the node here, findNode loop continues?
						// No, findNode splits by `/`.
						// If we are at `/assets` and part is `css`, we go into `*filepath`.
						// Next part `main.css` should also be handled by `*filepath`?
						// If `*filepath` node has no children, it should handle it?

						// Implementation detail:
						// If we encounter a wildcard node, we stop traversing parts and return it?
						// Yes, because * should capture the rest.
						// But we need to handle the case where we are inside the loop.

						// Let's check how we handle params. We update node = dyn and break.
						// Loop continues to next part.
						// But wildcard matches multiple parts.
						// So we should break the OUTER loop?
						// Or simple hack:

						// If wildcard, we consume the rest of searchPath + part?
						// Actually `searchPath` is modified in the loop.
						// `part` is current part.
						// If we match wildcard, we want to set param = part + "/" + searchPath  and return node.

						ctx.AddParam(paramName, part+"/"+searchPath)
					}
					node = dyn
					matched = true
					// We must assume wildcard is terminal and catches everything
					// Break outer loop?
					// Use goto?
					// Or just return here.
					return node, nil
				}
			}
			if !matched {
				return nil, fmt.Errorf("route not found")
			}
		}
	}

	if node.Handler == nil {
		// Handle root path or check if we are at a node that has a handler
		// But wait, the loop runs for parts. If path is "/", Trim returns "". Loop doesn't run.
		// node is root. If root has handler, return it.
		// Logic check: if path was just "/", we are at root[method].
		if node.Handler == nil {
			return nil, fmt.Errorf("route not found")
		}
	}
	return node, nil
}

func (r *TrieRouter) Find(method, path string, ctx *amaro.Context) (*amaro.Route, error) {
	node, err := r.findNode(method, path, ctx)
	if err != nil {
		return nil, err
	}

	// Return the raw handler without wrapping
	// The params are already inside ctx (if ctx was provided)
	return &amaro.Route{
		Method:      method,
		Path:        path,
		Handler:     node.Handler,
		Middlewares: node.Middlewares,
	}, nil
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
