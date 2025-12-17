package openapi

import "github.com/buildwithgo/amaro"

// WrapStream registers a handler as a streaming endpoint (text/event-stream) in the OpenAPI spec.
// It returns the original handler unchanged.
func WrapStream(g *Generator, method, path string, handler amaro.Handler) amaro.Handler {
	op := Operation{
		Summary: path,
		Responses: map[string]*Response{
			"200": {
				Description: "Stream Response",
				Content: map[string]*MediaType{
					"text/event-stream": {
						Schema: &Schema{Type: "string"},
					},
				},
			},
		},
	}
	g.AddRoute(method, path, op)
	return handler
}
