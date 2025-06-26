package amaro

import (
	"encoding/json"
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

func (c *Context) JSON(statusCode int, v interface{}) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(statusCode)
	// Assuming you have a JSON encoder function
	err := json.NewEncoder(c.Writer).Encode(v)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) HTML(statusCode int, html string) error {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(statusCode)
	_, err := c.Writer.Write([]byte(html))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) Redirect(statusCode int, url string) error {
	c.Writer.Header().Set("Location", url)
	c.Writer.WriteHeader(statusCode)
	return nil
}

func (c *Context) QueryParam(name string) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.URL.Query().Get(name)
}

func (c *Context) PathParam(name string) string {
	if c.PathParams == nil {
		return ""
	}
	return c.PathParams[name]
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) GetHeader(key string) string {
	if c.Request == nil {
		return ""
	}
	return c.Request.Header.Get(key)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	if c.Request == nil {
		return nil, http.ErrNoCookie
	}
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}
