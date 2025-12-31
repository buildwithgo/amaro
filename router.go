package amaro

import "io/fs"

// Route represents a registered route.
type Route struct {
	Method      string
	Path        string
	Handler     Handler
	Middlewares []Middleware
}

// ParamParser defines a function that checks if a path segment is a parameter.
// It returns true and the parameter name if it is, false otherwise.
type ParamParser func(segment string) (bool, string)

// WildcardParser defines a function that checks if a path segment is a wildcard.
// It returns true and the wildcard name if it is, false otherwise.
type WildcardParser func(segment string) (bool, string)

// RouterConfig defines configuration for Router.
type RouterConfig struct {
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

// DefaultRouterConfig returns the default configuration.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		ParamParser:    DefaultParamParser,
		WildcardParser: DefaultWildcardParser,
	}
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
