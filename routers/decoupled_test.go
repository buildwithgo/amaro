package routers

import (
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
)

func TestDecoupledParamSyntax(t *testing.T) {
	// Custom parser: matches <param>
	customParser := func(segment string) (bool, string) {
		if len(segment) > 2 && segment[0] == '<' && segment[len(segment)-1] == '>' {
			return true, segment[1 : len(segment)-1]
		}
		return false, ""
	}

	config := amaro.DefaultRouterConfig()
	config.ParamParser = customParser

	r := NewTrieRouter(WithConfig(config))
	r.GET("/users/<id>", func(c *amaro.Context) error { return nil })

	ctx := amaro.NewContext(nil, nil)

	// Should match /users/123
	_, err := r.Find(http.MethodGet, "/users/123", ctx)
	if err != nil {
		t.Fatalf("Failed to match custom syntax: %v", err)
	}
	if ctx.PathParam("id") != "123" {
		t.Errorf("Expected id=123, got %s", ctx.PathParam("id"))
	}

	// Should NOT match /users/:id syntax anymore (since we replaced parser)
	// Add another route with : syntax? No, add checks against parser logic.
	// Try to add /posts/:id and see if it's treated as static
	r.GET("/posts/:id", func(c *amaro.Context) error { return nil })

	ctx.Reset(nil, nil)
	// /posts/abc should NOT match because ":id" is treated as static literal ":id"
	_, err = r.Find(http.MethodGet, "/posts/abc", ctx)
	if err == nil {
		t.Error("Expected error matching /posts/abc against static /posts/:id")
	}

	// /posts/:id should match exact static
	_, err = r.Find(http.MethodGet, "/posts/:id", ctx)
	if err != nil {
		t.Errorf("Expected match for static /posts/:id: %v", err)
	}
}
