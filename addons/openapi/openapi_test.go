package openapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/openapi"
	"github.com/buildwithgo/amaro/routers"
)

type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestOpenAPIIntegration(t *testing.T) {
	// 1. Setup Amaro App
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// 2. Setup OpenAPI Generator
	gen := openapi.NewGenerator(openapi.Info{
		Title:   "Test API",
		Version: "1.0.0",
	})

	// 3. Define Typed Handler
	createHandler := func(c *amaro.Context, req *CreateUserRequest) (*UserResponse, error) {
		if req.Name == "bad" {
			return nil, c.String(http.StatusBadRequest, "bad name")
		}
		return &UserResponse{
			ID:   "123",
			Name: req.Name,
		}, nil
	}

	// 4. Register Routes
	app.POST("/users", openapi.WrapHandler(gen, "POST", "/users", createHandler))

	app.GET("/swagger.json", func(c *amaro.Context) error {
		return c.JSON(http.StatusOK, gen.Spec)
	})

	// 5. Test API Logic (Typing/Binding)
	t.Run("TypedHandler", func(t *testing.T) {
		reqBody := `{"name": "john", "age": 30}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var res UserResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if res.Name != "john" || res.ID != "123" {
			t.Errorf("Unexpected response: %+v", res)
		}
	})

	// 6. Test OpenAPI Spec Generation
	t.Run("SpecGeneration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/swagger.json", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 for spec, got %d", w.Code)
		}

		var spec openapi.OpenAPI
		if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
			t.Fatal(err)
		}

		if spec.Info.Title != "Test API" {
			t.Errorf("Expected Title 'Test API', got '%s'", spec.Info.Title)
		}

		pathItem, ok := spec.Paths["/users"]
		if !ok {
			t.Fatal("Expected /users path")
		}
		if pathItem.Post == nil {
			t.Fatal("Expected POST operation")
		}

		// Check Request Schema Ref
		reqContent := pathItem.Post.RequestBody.Content["application/json"]
		if reqContent == nil || reqContent.Schema.Ref != "#/components/schemas/CreateUserRequest" {
			t.Errorf("Expected request ref to CreateUserRequest, got %+v", reqContent)
		}

		// Check Components
		schema, ok := spec.Components.Schemas["CreateUserRequest"]
		if !ok {
			t.Fatal("Expected CreateUserRequest schema in components")
		}
		if schema.Properties["name"].Type != "string" {
			t.Errorf("Expected name property type string")
		}
	})
}
