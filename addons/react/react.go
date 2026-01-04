package react

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/buildwithgo/amaro"
)

// Config holds the configuration for the React engine.
type Config struct {
	// ViteDevURL is the URL of the Vite dev server (e.g., "http://localhost:5173").
	// If set, scripts will be loaded from here.
	ViteDevURL string
	// Assets is the filesystem containing built assets (dist/).
	// Used when ViteDevURL is empty.
	Assets fs.FS
	// Template is the HTML template for the root view.
	// It must contain a specific placeholder for the React mount point.
	// Defaults to a simple internal template if nil.
	Template *template.Template
	// Version is the asset version hash. Used for cache busting and force-reloading.
	Version string
}

// Engine manages the React integration.
type Engine struct {
	config Config
}

// New creates a new React engine.
func New(config Config) *Engine {
	if config.Template == nil {
		config.Template = defaultTemplate
	}
	return &Engine{config: config}
}

// Page represents the data sent to the client.
type Page struct {
	Component string `json:"component"`
	Props     any    `json:"props"`
	URL       string `json:"url"`
	Version   string `json:"version"`
}

// Render renders a React component.
// If the request is an X-Inertia request, it returns JSON.
// Otherwise, it returns the full HTML page with the component mounted.
func (e *Engine) Render(c *amaro.Context, component string, props any) error {
	page := Page{
		Component: component,
		Props:     props,
		URL:       c.Request.RequestURI,
		Version:   e.config.Version,
	}

	// Check if strictly Inertia request
	if c.GetHeader("X-Inertia") == "true" {
		c.SetHeader("X-Inertia", "true")
		c.SetHeader("Vary", "Accept")
		return c.JSON(http.StatusOK, page)
	}

	// Initial page load
	data, err := json.Marshal(page)
	if err != nil {
		return err
	}

	// Create view data
	viewData := map[string]any{
		"Page": template.HTML(data), // Safe because we just marshaled it
	}

	if e.config.ViteDevURL != "" {
		viewData["Vite"] = e.config.ViteDevURL
		viewData["IsDev"] = true
	} else {
		viewData["IsDev"] = false
		// In a real implementation, we would parse manifest.json here to find entry points.
		// For simplicity, we assume a standard entry point or let the user handle it in template.
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	return e.config.Template.Execute(c.Writer, viewData)
}

// Redirect performs a redirect compatible with the React adapter.
// It uses 303 See Other for PUT/PATCH/DELETE -> GET redirects which is standard for this pattern.
func (e *Engine) Redirect(c *amaro.Context, url string) error {
	if c.GetHeader("X-Inertia") == "true" {
		c.Writer.WriteHeader(http.StatusSeeOther) // 303
		c.Writer.Header().Set("Location", url)
		return nil
	}
	return c.Redirect(http.StatusFound, url)
}

var defaultTemplate = template.Must(template.New("react").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0" />
    {{ if .IsDev }}
    <script type="module" src="{{ .Vite }}/@vite/client"></script>
    <script type="module" src="{{ .Vite }}/src/main.jsx"></script>
    {{ else }}
    <script type="module" src="/assets/index.js"></script>
    <link rel="stylesheet" href="/assets/index.css" />
    {{ end }}
</head>
<body>
    <div id="app" data-page='{{ .Page }}'></div>
</body>
</html>
`))
