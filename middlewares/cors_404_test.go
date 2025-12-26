package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/middlewares"
	"github.com/buildwithgo/amaro/routers"
)

func TestCORS_NotFound(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	app.Use(middlewares.CORS(middlewares.DefaultCORSConfig()))

	// Route that exists
	app.GET("/hello", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "Hello")
	})

	t.Run("GET 404 with CORS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/not-found", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected proper CORS header on 404")
		}
	})

	t.Run("OPTIONS 404 with CORS", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/not-found", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		// With my fix, this should be 204 (handled by middleware completely)
		// WITHOUT my fix, it would be 404 (because middleware wouldn't run, router would fail finding OPTIONS route)
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected proper CORS header on OPTIONS 404")
		}
	})
}
