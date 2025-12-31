package middlewares

import (
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestBasicAuth(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Middleware
	mw := BasicAuth(func(username, password string, c *amaro.Context) (bool, error) {
		if username == "admin" && password == "secret" {
			return true, nil
		}
		return false, nil
	})

	app.GET("/protected", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "Allowed")
	}, mw)

	// Case 1: No Auth
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.code)
	}

	// Case 2: Invalid Auth
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("admin", "wrong")
	w = &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.code)
	}

	// Case 3: Valid Auth
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("admin", "secret")
	w = &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.code)
	}
	if w.body != "Allowed" {
		t.Errorf("Expected 'Allowed', got '%s'", w.body)
	}
}

func TestKeyAuth(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	mw := KeyAuth(func(key string, c *amaro.Context) (bool, error) {
		return key == "valid-api-key", nil
	})

	app.GET("/api", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "Success")
	}, mw)

	// Case 1: Missing Key
	req, _ := http.NewRequest("GET", "/api", nil)
	w := &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.code)
	}

	// Case 2: Invalid Key
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("X-API-Key", "bad-key")
	w = &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.code)
	}

	// Case 3: Valid Key
	req, _ = http.NewRequest("GET", "/api", nil)
	req.Header.Set("X-API-Key", "valid-api-key")
	w = &mockWriter{}
	app.ServeHTTP(w, req)
	if w.code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.code)
	}
}

// Mock Writer
type mockWriter struct {
	code   int
	body   string
	header http.Header
}
func (m *mockWriter) Header() http.Header {
	if m.header == nil { m.header = make(http.Header) }
	return m.header
}
func (m *mockWriter) Write(b []byte) (int, error) {
	m.body = string(b)
	return len(b), nil
}
func (m *mockWriter) WriteHeader(statusCode int) {
	m.code = statusCode
}
