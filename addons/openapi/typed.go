package openapi

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/buildwithgo/amaro"
)

// Bind decodes the request body into a new instance of T
func Bind[T any](c *amaro.Context) (*T, error) {
	var req T
	if c.Request.Body == nil {
		return &req, nil
	}
	defer c.Request.Body.Close()
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

type TypedHandler[Req any, Res any] func(*amaro.Context, *Req) (*Res, error)

// Handle registers a typed handler with the generator and returns a standard amaro.Handler
// It automatically generates request and response schemas.
func (g *Generator) Handle(method, path string, handler TypedHandler[any, any]) amaro.Handler {
	// This generic function signature is tricky because Go doesn't allow method generics easily
	// on non-generic types in the way we might want for inference if we want to extract Req/Res types
	// from the handler function itself without passing them as type params to Handle.
	//
	// However, to make it truly infer from the function, we need the function to be passed.
	// But `Handle[Req, Res]` requires instantiation at call site if not inferred.
	//
	// Let's try a wrapper function instead of a method on Generator if we want generics.
	panic("Use WrapHandler instead")
}

// WrapHandler wraps a typed handler and registers it with the generator.
func WrapHandler[Req any, Res any](g *Generator, method, path string, handler TypedHandler[Req, Res]) amaro.Handler {
	// 1. Generate Schema for Req
	var reqModel Req
	reqSchema := g.GenerateSchema(reqModel)

	// 2. Generate Schema for Res
	var resModel Res
	resSchema := g.GenerateSchema(resModel)

	// 3. Register Operation
	op := Operation{
		Summary: path,
		Responses: map[string]*Response{
			"200": {
				Description: "OK",
				Content: map[string]*MediaType{
					"application/json": {Schema: resSchema},
				},
			},
		},
	}

	// Add Request Body if Req is not struct{} (or check fields)
	// For simplicity, always add if not nil/empty struct?
	// Let's rely on type reflection.
	reqType := reflect.TypeOf(reqModel)
	if reqType.Kind() == reflect.Struct && reqType.NumField() > 0 {
		op.RequestBody = &RequestBody{
			Description: "Request Body",
			Required:    true,
			Content: map[string]*MediaType{
				"application/json": {Schema: reqSchema},
			},
		}
	}

	g.AddRoute(method, path, op)

	// 4. Return standard handler
	return func(c *amaro.Context) error {
		req, err := Bind[Req](c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid Request")
		}
		res, err := handler(c, req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}
