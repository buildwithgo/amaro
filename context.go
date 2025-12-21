package amaro

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.Request.FormFile(name)
	return fh, err
}

// SaveFile saves the uploaded file to the specified destination.
func (c *Context) SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

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
	Keys    map[string]interface{}
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
	// Reset Keys (nil them out or create new map if needed)
	c.Keys = nil
}

// NewContext creates a new context for the request
func NewContext(w http.ResponseWriter, r *http.Request, options ...ContextOption) *Context {
	ctx := &Context{
		Request: r,
		Writer:  w,
		Params:  make([]Param, 0, 10),
		Keys:    nil,
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

// Set stores a new key-value pair in the context for this request.
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get retrieves a value from the context.
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.Keys != nil {
		value, exists = c.Keys[key]
	}
	return
}
