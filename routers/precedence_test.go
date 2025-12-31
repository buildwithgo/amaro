package routers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
)

func TestTrieRouter_Precedence(t *testing.T) {
	r := NewTrieRouter()

	// Register conflicting routes
	// Static > Param > Wildcard

	r.GET("/users/search", func(c *amaro.Context) error {
		return fmt.Errorf("static")
	})

	r.GET("/users/:id", func(c *amaro.Context) error {
		return fmt.Errorf("param")
	})

	// Can't easily test Wildcard vs Param at SAME level with current Add logic (it might merge/conflict?)
	// Let's test wildcards deeper.
	r.GET("/files/*path", func(c *amaro.Context) error {
		return fmt.Errorf("wildcard")
	})

	ctx := amaro.NewContext(nil, nil)

	// 1. Static Priority
	route, err := r.Find(http.MethodGet, "/users/search", ctx)
	if err != nil { t.Fatal(err) }
	if err := route.Handler(ctx); err == nil || err.Error() != "static" {
		t.Errorf("Expected static handler, got %v", err)
	}

	// 2. Param Priority
	ctx.Reset(nil, nil)
	route, err = r.Find(http.MethodGet, "/users/123", ctx)
	if err != nil { t.Fatal(err) }
	if err := route.Handler(ctx); err == nil || err.Error() != "param" {
		t.Errorf("Expected param handler, got %v", err)
	}

	// 3. Wildcard
	ctx.Reset(nil, nil)
	route, err = r.Find(http.MethodGet, "/files/css/main.css", ctx)
	if err != nil { t.Fatal(err) }
	if err := route.Handler(ctx); err == nil || err.Error() != "wildcard" {
		t.Errorf("Expected wildcard handler, got %v", err)
	}
	if v := ctx.PathParam("path"); v != "css/main.css" {
		t.Errorf("Expected wildcard value 'css/main.css', got '%s'", v)
	}
}

func TestTrieRouter_ConflictDetection(t *testing.T) {
	r := NewTrieRouter()

	r.GET("/users/:id", func(c *amaro.Context) error { return nil })

	// Should fail if we try to register different param name at same level
	err := r.GET("/users/:user_id", func(c *amaro.Context) error { return nil })
	if err == nil {
		t.Error("Expected error for conflicting param name, got nil")
	}

	r.GET("/files/*path", func(c *amaro.Context) error { return nil })
	err = r.GET("/files/*filepath", func(c *amaro.Context) error { return nil })
	if err == nil {
		t.Error("Expected error for conflicting wildcard name, got nil")
	}
}
