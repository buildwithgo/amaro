package routers

import (
	"fmt"
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

	node.Handler = handler
	node.Middlewares = middlewares
	return nil
}

func (r *TrieRouter) findNode(method, path string) (*trieNode, map[string]string, error) {
	node, ok := r.root[method]
	if !ok {
		return nil, nil, fmt.Errorf("method not found")
	}
	params := make(map[string]string)
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if n, ok := node.children[part]; ok {
			node = n
		} else {
			matched := false
			for key, dyn := range node.children {
				if len(key) > 1 && key[0] == '{' && key[len(key)-1] == '}' {
					paramName := key[1 : len(key)-1]
					params[paramName] = part
					node = dyn
					matched = true
					break
				}
			}
			if !matched {
				return nil, nil, fmt.Errorf("route not found")
			}
		}
	}
	if node.Handler == nil {
		return nil, nil, fmt.Errorf("route not found")
	}
	return node, params, nil
}

func (r *TrieRouter) Find(method, path string) (*amaro.Route, error) {
	node, params, err := r.findNode(method, path)
	if err != nil {
		return nil, err
	}
	wrappedHandler := func(ctx *amaro.Context) error {
		ctx.PathParams = params
		return node.Handler(ctx)
	}
	return &amaro.Route{
		Method:      method,
		Path:        path,
		Handler:     wrappedHandler,
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
