package routers

import (
	"net/http"
	"testing"

	"github.com/buildwithgo/amaro"
)

func TestTrieRouter_CustomParamSyntax(t *testing.T) {
	// 1. Test Default support for :param and {param}
	r := NewTrieRouter()
	r.GET("/default/:id", func(c *amaro.Context) error { return nil })
	r.GET("/braces/{name}", func(c *amaro.Context) error { return nil })

	ctx := amaro.NewContext(nil, nil)

	// Check :id
	_, err := r.Find(http.MethodGet, "/default/123", ctx)
	if err != nil { t.Fatal(err) }
	if ctx.PathParam("id") != "123" { t.Errorf("Expected 123, got %s", ctx.PathParam("id")) }

	// Check {name}
	ctx.Reset(nil, nil)
	_, err = r.Find(http.MethodGet, "/braces/john", ctx)
	if err != nil { t.Fatal(err) }
	if ctx.PathParam("name") != "john" { t.Errorf("Expected john, got %s", ctx.PathParam("name")) }
}
