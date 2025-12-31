package amaro

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// StaticConfig defines configuration for serving static files.
type StaticConfig struct {
	// Root is the filesystem to serve from.
	Root fs.FS

	// Prefix is the URL path prefix.
	Prefix string

	// Index is the index file name (default: "index.html").
	Index string

	// Browse enables directory listing (default: false).
	Browse bool

	// SPA mode: if file not found, serve Index (default: false).
	SPA bool

	// ModifyResponse allows setting custom headers.
	ModifyResponse func(c *Context)
}

// StaticHandler creates a handler that serves static files.
func StaticHandler(config StaticConfig) Handler {
	if config.Index == "" {
		config.Index = "index.html"
	}

	// Normalize prefix
	if config.Prefix != "" {
		if config.Prefix[0] != '/' {
			config.Prefix = "/" + config.Prefix
		}
		config.Prefix = strings.TrimRight(config.Prefix, "/")
	}

	return func(c *Context) error {
		if config.ModifyResponse != nil {
			config.ModifyResponse(c)
		}

		urlPath := c.Request.URL.Path
		filepath := urlPath
		if config.Prefix != "" {
			if strings.HasPrefix(filepath, config.Prefix) {
				filepath = filepath[len(config.Prefix):]
			}
		}

		// Clean path
		filepath = path.Clean(filepath)
		if filepath == "." || filepath == "/" {
			filepath = ""
		}
		filepath = strings.TrimPrefix(filepath, "/")

		// Try to open file
		f, err := config.Root.Open(filepath)
		if err != nil {
			// File not found or other error
			if os.IsNotExist(err) {
				if config.SPA {
					return serveFile(c, config.Root, config.Index)
				}
				// Return 404 error
				return NewHTTPError(http.StatusNotFound, "File Not Found").SetInternal(err)
			}
			return err
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		if stat.IsDir() {
			// Check for index file
			indexFunc := func() error {
				indexPath := path.Join(filepath, config.Index)
				indexFile, err := config.Root.Open(indexPath)
				if err == nil {
					defer indexFile.Close()
					indexStat, err := indexFile.Stat()
					if err == nil {
						return serveContent(c, config.Index, indexStat.ModTime(), indexFile)
					}
				}
				return err
			}

			if err := indexFunc(); err == nil {
				return nil
			}

			if config.Browse {
				// TODO: Implement directory listing
				// For now fallback to 403
				return NewHTTPError(http.StatusForbidden, "Directory Listing Forbidden")
			}

			if config.SPA {
				return serveFile(c, config.Root, config.Index)
			}

			return NewHTTPError(http.StatusNotFound, "File Not Found")
		}

		return serveContent(c, stat.Name(), stat.ModTime(), f)
	}
}

func serveFile(c *Context, fsys fs.FS, name string) error {
	f, err := fsys.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	return serveContent(c, stat.Name(), stat.ModTime(), f)
}

func serveContent(c *Context, name string, modtime time.Time, content fs.File) error {
	rs, ok := content.(io.ReadSeeker)
	if !ok {
		return fmt.Errorf("file does not support seeking")
	}

	http.ServeContent(c.Writer, c.Request, name, modtime, rs)
	return nil
}
