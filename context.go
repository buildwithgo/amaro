package amaro

import (
	"encoding/json"
	"net/http"
)

// Param represents a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Context represents the context of the current HTTP request.
// It holds the request and response objects, URL parameters, and provides helper methods.
// It is designed to be reused via sync.Pool to minimize allocations.
type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Params  []Param // efficient slice instead of map
}

type ContextOption func(*Context)

// Reset resets the context to be reused in sync.Pool
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Writer = w
	// Resize params slice to capacity to avoid allocation if possible
	if cap(c.Params) < 10 {
		c.Params = make([]Param, 0, 10)
	} else {
		c.Params = c.Params[:0]
	}
}

// NewContext creates a new context for the request
func NewContext(w http.ResponseWriter, r *http.Request, options ...ContextOption) *Context {
	ctx := &Context{
		Request: r,
		Writer:  w,
		Params:  make([]Param, 0, 10),
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
	for _, p := range c.Params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

func (c *Context) AddParam(key, value string) {
	c.Params = append(c.Params, Param{Key: key, Value: value})
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

func (c *Context) Status(statusCode int) {
	c.Writer.WriteHeader(statusCode)
}
