package routers_test

import (
	"sort"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestRoutesIntrospection(t *testing.T) {
	r := routers.NewTrieRouter()

	dummyHandler := func(c *amaro.Context) error { return nil }

	r.GET("/users", dummyHandler)
	r.POST("/users", dummyHandler)
	r.GET("/users/:id", dummyHandler)
	r.GET("/assets/*filepath", dummyHandler)

	routes := r.Routes()

	// Expected routes
	// GET /users
	// POST /users
	// GET /users/:id
	// GET /assets/*filepath

	expected := []struct {
		Method string
		Path   string
	}{
		{"GET", "/assets/*filepath"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"POST", "/users"},
	}

	// Helper to sort actual routes for comparison (though implementation already sorts)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Method != routes[j].Method {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	if len(routes) != len(expected) {
		t.Fatalf("Expected %d routes, got %d", len(expected), len(routes))
	}

	for i, exp := range expected {
		if routes[i].Method != exp.Method || routes[i].Path != exp.Path {
			t.Errorf("Index %d: Expected %s %s, got %s %s", i, exp.Method, exp.Path, routes[i].Method, routes[i].Path)
		}
	}
}
