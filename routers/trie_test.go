package routers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
)

func TestTrieRouter_Basic(t *testing.T) {
	r := NewTrieRouter()

	handler := func(c *amaro.Context) error { return nil }

	r.GET("/hello", handler)
	r.POST("/world", handler)

	route, err := r.Find(http.MethodGet, "/hello", nil)
	if err != nil {
		t.Fatalf("Expected match, got error: %v", err)
	}
	if route == nil || route.Path != "/hello" {
		t.Errorf("Expected route /hello, got %v", route)
	}

	_, err = r.Find(http.MethodGet, "/world", nil)
	if err == nil {
		t.Error("Expected error for GET /world, got match")
	}

	route, err = r.Find(http.MethodPost, "/world", nil)
	if err != nil {
		t.Fatalf("Expected match, got error: %v", err)
	}
}

func TestTrieRouter_Params(t *testing.T) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }

	r.GET("/users/:id", handler)
	r.GET("/users/:id/posts/:post_id", handler)

	ctx := amaro.NewContext(nil, nil)

	// Test /users/123
	_, err := r.Find(http.MethodGet, "/users/123", ctx)
	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}
	if val := ctx.PathParam("id"); val != "123" {
		t.Errorf("Expected id=123, got %s", val)
	}

	// Test /users/123/posts/456
	ctx.Reset(nil, nil)
	_, err = r.Find(http.MethodGet, "/users/123/posts/456", ctx)
	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}
	if val := ctx.PathParam("id"); val != "123" {
		t.Errorf("Expected id=123, got %s", val)
	}
	if val := ctx.PathParam("post_id"); val != "456" {
		t.Errorf("Expected post_id=456, got %s", val)
	}
}

func TestTrieRouter_Wildcard(t *testing.T) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }

	r.GET("/static/*filepath", handler)

	ctx := amaro.NewContext(nil, nil)

	cases := []struct {
		path string
		want string
	}{
		{"/static/css/style.css", "css/style.css"},
		{"/static/js/app.js", "js/app.js"},
		{"/static/", ""}, // Empty match?
	}

	for _, tc := range cases {
		ctx.Reset(nil, nil)
		_, err := r.Find(http.MethodGet, tc.path, ctx)
		if err != nil {
			t.Errorf("Failed to find wildcard route for %s: %v", tc.path, err)
			continue
		}
		if got := ctx.PathParam("filepath"); got != tc.want {
			t.Errorf("For path %s, expected filepath=%q, got %q", tc.path, tc.want, got)
		}
	}
}

func TestTrieRouter_Wildcard_Root(t *testing.T) {
	r := NewTrieRouter()
	handler := func(c *amaro.Context) error { return nil }

	// Catch all at root
	r.GET("/*all", handler)

	ctx := amaro.NewContext(nil, nil)
	_, err := r.Find(http.MethodGet, "/anything/goes/here", ctx)
	if err != nil {
		t.Fatalf("Failed to match root wildcard: %v", err)
	}
	if got := ctx.PathParam("all"); got != "anything/goes/here" {
		t.Errorf("Expected all='anything/goes/here', got %q", got)
	}
}

func TestTrieRouter_DynamicConflict(t *testing.T) {
	// Our router allows adding multiple dynamics.
	// We prioritize static > param > wildcard.

	r := NewTrieRouter()
	// Use distinguishable handlers
	handlerStatic := func(c *amaro.Context) error { return fmt.Errorf("static") }
	handlerParam := func(c *amaro.Context) error { return fmt.Errorf("param") }

	r.GET("/users/search", handlerStatic)
	r.GET("/users/:id", handlerParam)

	ctx := amaro.NewContext(nil, nil)

	// 1. /users/search should match static
	route, err := r.Find(http.MethodGet, "/users/search", ctx)
	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}
	// Execute handler to verify identity
	if err := route.Handler(ctx); err == nil || err.Error() != "static" {
		t.Errorf("Expected static handler, got error: %v", err)
	}

	// 2. /users/123 should match param
	ctx.Reset(nil, nil)
	route, err = r.Find(http.MethodGet, "/users/123", ctx)
	if err != nil {
		t.Fatalf("Failed to find route: %v", err)
	}
	if err := route.Handler(ctx); err == nil || err.Error() != "param" {
		t.Errorf("Expected param handler, got error: %v", err)
	}
	if val := ctx.PathParam("id"); val != "123" {
		t.Errorf("Expected id=123, got %s", val)
	}
}
