package amaro_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestBasicRouting(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	app.GET("/hello", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "world")
	})

	app.GET("/users/{id}", func(c *amaro.Context) error {
		id := c.PathParam("id")
		return c.String(http.StatusOK, "user "+id)
	})

	t.Run("Static", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if w.Body.String() != "world" {
			t.Errorf("Expected 'world', got '%s'", w.Body.String())
		}
	})

	t.Run("Param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if w.Body.String() != "user 123" {
			t.Errorf("Expected 'user 123', got '%s'", w.Body.String())
		}
	})
}

func BenchmarkStaticRoute(b *testing.B) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))
	app.GET("/hello", func(c *amaro.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(w, req)
	}
}

func BenchmarkParamRoute(b *testing.B) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))
	app.GET("/users/{id}", func(c *amaro.Context) error {
		_ = c.PathParam("id")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(w, req)
	}
}

// Comparison with raw net/http to see overhead
func BenchmarkNetHttp(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}
