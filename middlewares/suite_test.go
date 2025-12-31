package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/middlewares"
	"github.com/buildwithgo/amaro/routers"
)

func TestMiddlewares(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// app.Use(amaro.Recovery()) // Already added by default in amaro.New()
	app.Use(middlewares.RequestID())
	app.Use(middlewares.Secure())
	app.Use(middlewares.CORS())
	app.Use(middlewares.Compress())

	// Route that panics
	app.GET("/panic", func(c *amaro.Context) error {
		panic("oops")
	})

	// Route that sleeps
	app.GET("/sleep", func(c *amaro.Context) error {
		time.Sleep(100 * time.Millisecond)
		return c.String(http.StatusOK, "woke up")
	}, middlewares.Timeout(50*time.Millisecond))

	// Normal route
	app.GET("/hello", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	t.Run("Recovery", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()

		// Capture stdout to avoid noise? Or just let it print.
		app.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %d", w.Code)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sleep", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected 503, got %d", w.Code)
		}
	})

	t.Run("Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/hello", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Header().Get("X-Request-ID") == "" {
			t.Error("Expected X-Request-ID header")
		}
		if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
			t.Error("Expected X-Frame-Options header")
		}
	})
}
