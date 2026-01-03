package amaro_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestAppTestHelper(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	app.GET("/hello", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "world")
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := app.Test(req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "world" {
		t.Errorf("Expected 'world', got %s", w.Body.String())
	}
}
