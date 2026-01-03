package amaro_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		// Let's test the middleware function directly to be sure about config.
		router := routers.NewTrieRouter()
		mw := amaro.Recovery()

		router.Use(mw)
		router.GET("/panic", func(c *amaro.Context) error {
			panic("oops")
		})

		// We need to simulate the app/router execution manually since we aren't using amaro.New
		// But router.ServeHTTP isn't a thing, we need App or to wrap it.
		// amaro.App wraps the router.

		// Let's use a dummy handler wrapped by the middleware
		handler := mw(func(c *amaro.Context) error {
			panic("oops")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		c := amaro.NewContext(w, req)

		err := handler(c)
		if err != nil {
			t.Errorf("Expected nil error (recovered), got %v", err)
		}

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %d", w.Code)
		}
		if w.Body.String() != "Internal Server Error" {
			t.Errorf("Expected 'Internal Server Error', got %s", w.Body.String())
		}
	})

	t.Run("HTMLDebug", func(t *testing.T) {
		mw := amaro.Recovery(amaro.WithHTMLDebug(true))

		handler := mw(func(c *amaro.Context) error {
			panic("debug me")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		c := amaro.NewContext(w, req)

		_ = handler(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("Expected HTML response")
		}
		if !strings.Contains(body, "debug me") {
			t.Error("Expected panic message in body")
		}
	})
}
