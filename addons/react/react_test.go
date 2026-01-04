package react_test

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/react"
)

func TestRender(t *testing.T) {
	engine := react.New(react.Config{
		Version: "1.0",
	})

	t.Run("Initial Load (HTML)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/home", nil)
		c := amaro.NewContext(w, req)

		err := engine.Render(c, "Home", map[string]string{"msg": "hello"})
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
			t.Errorf("Expected HTML content type")
		}
		body := w.Body.String()
		if !strings.Contains(body, `data-page='{"component":"Home"`) {
			t.Error("Expected data-page attribute with JSON")
		}
		if !strings.Contains(body, `"props":{"msg":"hello"}`) {
			t.Error("Expected props in JSON")
		}
	})

	t.Run("Subsequent Navigation (JSON)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/about", nil)
		req.Header.Set("X-Inertia", "true")
		c := amaro.NewContext(w, req)

		err := engine.Render(c, "About", map[string]int{"id": 1})
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if w.Header().Get("X-Inertia") != "true" {
			t.Error("Expected X-Inertia header in response")
		}

		var page react.Page
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if page.Component != "About" {
			t.Errorf("Expected component About, got %s", page.Component)
		}
		if page.URL != "/about" {
			t.Errorf("Expected URL /about, got %s", page.URL)
		}
	})
}

func TestRedirect(t *testing.T) {
	engine := react.New(react.Config{})

	t.Run("Standard Redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", nil)
		c := amaro.NewContext(w, req)

		engine.Redirect(c, "/dashboard")
		if w.Code != http.StatusFound { // 302
			t.Errorf("Expected 302, got %d", w.Code)
		}
		if w.Header().Get("Location") != "/dashboard" {
			t.Errorf("Expected Location /dashboard, got %s", w.Header().Get("Location"))
		}
	})

	t.Run("Inertia Redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", nil)
		req.Header.Set("X-Inertia", "true")
		c := amaro.NewContext(w, req)

		engine.Redirect(c, "/dashboard")
		if w.Code != http.StatusSeeOther { // 303
			t.Errorf("Expected 303, got %d", w.Code)
		}
		if w.Header().Get("Location") != "/dashboard" {
			t.Errorf("Expected Location /dashboard, got %s", w.Header().Get("Location"))
		}
	})
}

func TestCustomTemplate(t *testing.T) {
	tmpl := template.Must(template.New("custom").Parse(`My App: {{ .Page }}`))
	engine := react.New(react.Config{Template: tmpl})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := amaro.NewContext(w, req)

	engine.Render(c, "Test", nil)
	if !strings.Contains(w.Body.String(), "My App:") {
		t.Error("Custom template was not used")
	}
}
