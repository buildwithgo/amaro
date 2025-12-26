package amaro_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestErrorHandler(t *testing.T) {
	// 1. Default Handler (Plain Text)
	t.Run("DefaultErrorHandler", func(t *testing.T) {
		app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

		// Force a 404
		req := httptest.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
		if w.Body.String() != "404 page not found\n" { // http.Error adds newline
			t.Errorf("Expected default http.Error output, got '%s'", w.Body.String())
		}
	})

	// 2. Custom JSON Handler
	t.Run("CustomJSONHandler", func(t *testing.T) {
		type ErrorResponse struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
			Code    int    `json:"code"`
		}

		customHandler := func(c *amaro.Context, err error, code int) {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(code)
			json.NewEncoder(c.Writer).Encode(ErrorResponse{
				Success: false,
				Error:   err.Error(),
				Code:    code,
			})
		}

		app := amaro.New(
			amaro.WithRouter(routers.NewTrieRouter()),
			amaro.WithErrorHandler(customHandler),
		)

		// Case A: 404 Not Found
		t.Run("NotFound", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/missing", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected 404, got %d", w.Code)
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Expected application/json content type")
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Success != false || resp.Code != 404 {
				t.Errorf("Unexpected JSON response: %+v", resp)
			}
		})

		// Case B: 500 Internal Error (via handler error)
		app.GET("/panic", func(c *amaro.Context) error {
			return errors.New("something went wrong")
		})

		t.Run("InternalError", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/panic", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected 500, got %d", w.Code)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error != "something went wrong" {
				t.Errorf("Expected error message 'something went wrong', got '%s'", resp.Error)
			}
		})
	})
}
