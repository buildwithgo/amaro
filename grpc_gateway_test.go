package amaro_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestMount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "foo")
	})
	mux.HandleFunc("/bar/baz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "barbaz")
	})

	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Mount the mux at /gateway
	// Use StripPrefix so mux sees paths starting from /foo, /bar/baz
	app.Mount("/gateway", http.StripPrefix("/gateway", mux))

	t.Run("Exact Match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/gateway/foo", nil)
		w := app.Test(req)
		if w.Body.String() != "foo" {
			t.Errorf("Expected 'foo', got '%s'", w.Body.String())
		}
	})

	t.Run("Deep Match", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/gateway/bar/baz", nil)
		w := app.Test(req)
		if w.Body.String() != "barbaz" {
			t.Errorf("Expected 'barbaz', got '%s'", w.Body.String())
		}
	})

	t.Run("Method Support", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/gateway/foo", nil)
		w := app.Test(req)
		if w.Body.String() != "foo" {
			t.Errorf("Expected 'foo', got '%s'", w.Body.String())
		}
	})
}

func TestGroupMount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/echo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "echo")
	})

	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))
	api := app.Group("/api")

	// Mount at /api -> so it handles /api/v1/echo
	// Strip /api so mux sees /v1/echo
	api.Mount("/", http.StripPrefix("/api", mux))

	req := httptest.NewRequest("GET", "/api/v1/echo", nil)
	w := app.Test(req)
	if w.Body.String() != "echo" {
		t.Errorf("Expected 'echo', got '%s'", w.Body.String())
	}
}

func TestAny(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	app.Any("/all", func(c *amaro.Context) error {
		return c.String(200, c.Request.Method)
	})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, m := range methods {
		req := httptest.NewRequest(m, "/all", nil)
		w := app.Test(req)
		if w.Body.String() != m {
			t.Errorf("Expected %s, got %s", m, w.Body.String())
		}
	}
}
