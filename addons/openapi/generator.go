package openapi

import (
	"reflect"
	"strings"
	"time"
)

type Generator struct {
	Spec *OpenAPI
}

func NewGenerator(info Info) *Generator {
	return &Generator{
		Spec: &OpenAPI{
			OpenAPI: "3.0.0",
			Info:    info,
			Paths:   make(Paths),
			Components: &Components{
				Schemas: make(map[string]*Schema),
			},
		},
	}
}

func (g *Generator) AddRoute(method, path string, op Operation) {
	if g.Spec.Paths[path] == nil {
		g.Spec.Paths[path] = &PathItem{}
	}
	item := g.Spec.Paths[path]
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		item.Get = &op
	case "POST":
		item.Post = &op
	case "PUT":
		item.Put = &op
	case "DELETE":
		item.Delete = &op
	case "PATCH":
		item.Patch = &op
	case "OPTIONS":
		item.Options = &op
	case "HEAD":
		item.Head = &op
	}
}

// GenerateSchema creates a schema for v and registers it in Components if it's a struct
func (g *Generator) GenerateSchema(v interface{}) *Schema {
	t := reflect.TypeOf(v)
	return g.generateSchemaType(t)
}

func (g *Generator) generateSchemaType(t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference pointer
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		// Special case for byte slice -> string/binary
		if t.Elem().Kind() == reflect.Uint8 {
			return &Schema{Type: "string", Format: "binary"}
		}
		return &Schema{
			Type:  "array",
			Items: g.generateSchemaType(t.Elem()),
		}
	case reflect.Map:
		return &Schema{
			Type: "object",
			// AdditionalProperties? For now simple object
		}
	case reflect.Struct:
		// Check for time.Time
		if t == reflect.TypeOf(time.Time{}) {
			return &Schema{Type: "string", Format: "date-time"}
		}

		name := t.Name()
		// If unnamed struct, generate inline
		if name == "" {
			return g.generateStructSchema(t)
		}

		// Register in Components if not exists
		if _, ok := g.Spec.Components.Schemas[name]; !ok {
			// Placeholder to prevent infinite recursion
			g.Spec.Components.Schemas[name] = &Schema{}
			schema := g.generateStructSchema(t)
			g.Spec.Components.Schemas[name] = schema
		}
		return &Schema{Ref: "#/components/schemas/" + name}

	default:
		return &Schema{Type: "string"} // Fallback
	}
}

func (g *Generator) generateStructSchema(t reflect.Type) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Ignore unexported fields
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		name := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			name = parts[0]
		}

		schema.Properties[name] = g.generateSchemaType(field.Type)
	}
	return schema
}
