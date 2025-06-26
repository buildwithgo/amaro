package amaro

import (
	"net/http"
)

type Context struct {
	Request    *http.Request
	Writer     http.ResponseWriter
	PathParams map[string]string // Path parameters
}

type ContextOption func(*Context)

// NewContext creates a new context for the request
func NewContext(w http.ResponseWriter, r *http.Request, options ...ContextOption) *Context {
	ctx := &Context{
		Request: r,
		Writer:  w,
	}
	for _, option := range options {
		option(ctx)
	}
	return ctx
}

func (c *Context) String(statusCode int, s string) error {
	c.Writer.WriteHeader(statusCode)
	_, err := c.Writer.Write([]byte(s))
	return err
}
