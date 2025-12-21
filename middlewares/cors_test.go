package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/middlewares"
	"github.com/buildwithgo/amaro/routers"
)

func TestCORS_ServedCall(t *testing.T) {
	// Setup Amaro app
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Use default CORS config
	app.Use(middlewares.CORS())

	app.GET("/cors-test", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	// We must register OPTIONS for the middleware to be triggered in Amaro's current design
	app.OPTIONS("/cors-test", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Start test server (served call request)
	ts := httptest.NewServer(app)
	defer ts.Close()

	t.Run("Allowed Origin", func(t *testing.T) {
		// Create client request
		client := ts.Client()
		req, err := http.NewRequest("GET", ts.URL+"/cors-test", nil)
		if err != nil {
			t.Fatal(err)
		}
		// Set Origin to something generic
		req.Header.Set("Origin", "http://example.com")

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Verify status
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify CORS headers
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		// Default config allows *
		if allowOrigin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin to be '*', got '%s'", allowOrigin)
		}
	})

	t.Run("Preflight Request", func(t *testing.T) {
		client := ts.Client()
		req, err := http.NewRequest("OPTIONS", ts.URL+"/cors-test", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204 for preflight, got %d", resp.StatusCode)
		}

		allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
		if allowMethods == "" {
			t.Error("Expected Access-Control-Allow-Methods header")
		}
	})
}
