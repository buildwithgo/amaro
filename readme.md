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
)

func main() {
    app := amaro.New()

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

## 🛠️ Core Concepts

### Routing

Amaro uses a high-performance Trie router supporting standard HTTP methods and dynamic parameters.

### Grouping

Organize your API with groups.

```go
api := app.Group("/api")
{
    v1 := api.Group("/v1")
    v1.GET("/users", handler) // GET /api/v1/users
}
```

### Middleware

Add global or route-specific middleware.

```go
// Global
app.Use(func(next amaro.Handler) amaro.Handler {
    return func(c *amaro.Context) error {
        println("Request received")
        return next(c)
    }
})

// Per-route
app.GET("/admin", adminHandler, authMiddleware)
```

## 📖 Cookbook

### Accessing Path Parameters

```go
app.GET("/user/:id", func(c *amaro.Context) error {
    id := c.PathParam("id") // Use PathParam to get dynamic route parameters
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

### JSON Request & Response

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

app.POST("/users", func(c *amaro.Context) error {
    var u User
    // Standard Go JSON decoding
    if err := json.NewDecoder(c.Request.Body).Decode(&u); err != nil {
        return c.String(400, "Invalid JSON")
    }

    return c.JSON(201, map[string]interface{}{
        "message": "User created",
        "user":    u,
    })
})
```

### Custom Middleware

```go
// Logger middleware
func Logger() amaro.Middleware {
    return func(next amaro.Handler) amaro.Handler {
        return func(c *amaro.Context) error {
            start := time.Now()
            err := next(c)
            duration := time.Since(start)
            log.Printf("[%s] %s took %v", c.Request.Method, c.Request.URL.Path, duration)
            return err
        }
    }
}

app.Use(Logger())
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

// 1. Manual Route Registration
gen.AddRoute("GET", "/users", openapi.Operation{
    Summary: "List users",
    Responses: map[string]*openapi.Response{
        "200": {Description: "Successful response"},
    },
})

// 2. Type-Safe Handlers with Automatic Schema Generation
type CreateUserReq struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserRes struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// WrapHandler automatically generates OpenAPI schemas for Request/Response
handler := openapi.WrapHandler(gen, "POST", "/users", func(c *amaro.Context, req *CreateUserReq) (*UserRes, error) {
    // req is already bound and validated
    return &UserRes{
        ID:    "123",
        Name:  req.Name,
        Email: req.Email,
    }, nil
})

app.POST("/users", handler)
```

### Streaming

Built-in support for Server-Sent Events (SSE) and data streaming.

## 🏗️ Real World Example

Here's how to build a production-ready API with Authentication, Groups, and JSON validation.

```go
package main

import (
    "log"
    "net/http"
    "strings"
    
    "github.com/buildwithgo/amaro"
)

// AuthMiddleware - A simple token-based authentication middleware
func AuthMiddleware() amaro.Middleware {
    return func(next amaro.Handler) amaro.Handler {
        return func(c *amaro.Context) error {
            authHeader := c.GetHeader("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer secret-token") {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "Unauthorized access",
                })
            }
            return next(c)
        }
    }
}

// Product struct
type Product struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

func main() {
    app := amaro.New()

    // 1. Public Routes
    app.GET("/health", func(c *amaro.Context) error {
        return c.String(200, "OK")
    })

    // 2. Private API Group with Middleware
    api := app.Group("/api/v1")
    api.Use(AuthMiddleware()) // Apply Auth to all routes in this group

    // POST /api/v1/products
    api.POST("/products", func(c *amaro.Context) error {
        var p Product
        // Standard Go JSON decoding
        if err := json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
             return c.String(400, "Bad Request")
        }
        // In real app, save to DB...
        p.ID = "prod_123" 
        return c.JSON(201, p)
    })

    // GET /api/v1/products/:id
    api.GET("/products/:id", func(c *amaro.Context) error {
        id := c.PathParam("id")
        if id == "" {
             return c.JSON(400, map[string]string{"error": "ID required"})
        }
        
        return c.JSON(200, Product{
            ID:    id,
            Name:  "Super Widget",
            Price: 99.99,
        })
    })

    log.Println("Server running on :8080")
    app.Run("8080")
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [license](license.md) file for details.
