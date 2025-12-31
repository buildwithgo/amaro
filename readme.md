# Amaro

<div align="center">
  <h3>🔥 The blazing fast, zero-dependency Go HTTP framework</h3>
  <p>Inspired by Hono.js, built for performance and simplicity.</p>
</div>

---

**Amaro** is a lightweight, high-performance web framework for Go. It prioritizes zero allocations in the hot path, zero external dependencies, and a developer-friendly API that feels like modern JavaScript frameworks but with the power of Go.

## 🚀 Features

- **Zero Dependency**: Runs on pure Go standard library.
- **Blazing Fast**: Optimized Trie-based router with zero-allocation context pooling.
- **Decoupled Architecture**: Router implementation is fully decoupled from the core framework.
- **Configurable Syntax**: Support for customizable parameter delimiters (e.g. `:id` or `{id}`).
- **Robust Static Serving**: Built-in support for serving static files, SPAs, and directory browsing (configurable).
- **Production-Grade Middlewares**: Includes Auth (Basic, Key, Session, RBAC), CORS, Cache, and more.
- **Group Routing**: Organize routes with prefixes and shared middlewares.
- **Context Pooling**: Reuses request contexts to minimize GC pressure.
- **Addon System**: Extensible with powerful addons like OpenAPI generation and Streaming.

## 📦 Installation

```bash
go get github.com/buildwithgo/amaro
```

## ⚡ Quick Start

```go
package main

import (
    "net/http"
    "github.com/buildwithgo/amaro"
    "github.com/buildwithgo/amaro/routers"
)

func main() {
    // Initialize with the optimized TrieRouter
    app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

    app.GET("/", func(c *amaro.Context) error {
        return c.String(http.StatusOK, "Hello, Amaro! 🔥")
    })

    app.GET("/json", func(c *amaro.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Blazing fast JSON",
        })
    })

    app.Run("8080")
}
```

## 🛠️ Advanced Configuration

### Customizing Router Syntax

Amaro's TrieRouter supports configurable parameter syntax. You can use standard colon syntax (`:id`) or brackets (`{id}`), or define your own.

```go
config := routers.DefaultTrieRouterConfig()
// Enable custom bracket syntax if desired (default is {} and :)
config.ParamPrefix = "<"
config.ParamSuffix = ">"

r := routers.NewTrieRouter(routers.WithConfig(config))
app := amaro.New(amaro.WithRouter(r))

app.GET("/users/<id>", handler) // Matches /users/123
```

### Static File Serving

Serve static files with robust support for SPAs (Single Page Applications).

```go
app.StaticFS("/assets", os.DirFS("./public"))

// Or using the robust Static handler manually for more control
app.GET("/app/*filepath", amaro.StaticHandler(amaro.StaticConfig{
    Root: os.DirFS("./dist"),
    SPA:  true, // Serve index.html on 404
    Index: "index.html",
}))
```

## 🛡️ Middlewares

Amaro comes with a suite of production-grade middlewares.

### Authentication & Authorization

```go
import "github.com/buildwithgo/amaro/middlewares"

// Basic Auth
app.Use(middlewares.BasicAuth(func(user, pass string, c *amaro.Context) (bool, error) {
    return user == "admin" && pass == "secret", nil
}))

// API Key Auth
app.Use(middlewares.KeyAuth(func(key string, c *amaro.Context) (bool, error) {
    return key == "valid-api-key", nil
}))

// Session Auth (requires addons/sessions)
app.Use(middlewares.SessionAuth[User](validatorFunc))

// RBAC (Role-Based Access Control)
app.GET("/admin", middlewares.RBAC("admin", roleExtractor), adminHandler)
```

### CORS & Caching

```go
// CORS with options
app.Use(middlewares.CORS(middlewares.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowCredentials: true,
}))

// Cache responses
store := cache.NewMemoryCache()
app.GET("/cached-data", middlewares.CachePage(store, 5*time.Minute), handler)
```

## 📖 Cookbook

### Accessing Path Parameters

```go
app.GET("/user/:id", func(c *amaro.Context) error {
    id := c.PathParam("id") // Works with :id or {id} or configured syntax
    return c.String(200, "User ID: "+id)
})
```

### Accessing Query Parameters

```go
// GET /search?q=golang
app.GET("/search", func(c *amaro.Context) error {
    query := c.QueryParam("q")
    return c.String(200, "Searching for: "+query)
})
```

## 🔌 Addons

### OpenAPI Generator

Amaro includes a built-in OpenAPI v3 generator to automatically document your API.

```go
import "github.com/buildwithgo/amaro/addons/openapi"

// Create generator
gen := openapi.NewGenerator(openapi.Info{
    Title:   "My API",
    Version: "1.0.0",
})
// ... (see full docs for usage)
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [license](license.md) file for details.
