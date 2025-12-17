package openapi

// OpenAPI represents the root document of the OpenAPI v3 specification.
type OpenAPI struct {
	OpenAPI    string      `json:"openapi"`
	Info       Info        `json:"info"`
	Servers    []Server    `json:"servers,omitempty"`
	Paths      Paths       `json:"paths"`
	Components *Components `json:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type Paths map[string]*PathItem

type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // query, header, path, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]*MediaType `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

type Response struct {
	Description string                `json:"description"`
	Content     map[string]*MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type   string `json:"type"`
	Name   string `json:"name,omitempty"`
	In     string `json:"in,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Bearer string `json:"bearerFormat,omitempty"`
}

type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Description string             `json:"description,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
}
