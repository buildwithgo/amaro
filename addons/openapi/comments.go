package openapi

import (
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// CommentParser parses Go source files to extract struct documentation
type CommentParser struct {
	TypeDocs map[string]string // Struct Name -> Doc Comment
}

// NewCommentParser creates a new parser
func NewCommentParser() *CommentParser {
	return &CommentParser{
		TypeDocs: make(map[string]string),
	}
}

// ParseDocs parses all Go files in the given directory and subdirectories
// This is a simplified version scanning only the provided root path
func (cp *CommentParser) ParseDocs(root string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, root, func(fi os.FileInfo) bool {
		return true
	}, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		d := doc.New(pkg, root, doc.AllDecls)
		for _, t := range d.Types {
			// t.Name is the struct name
			// t.Doc is the comment block
			cp.TypeDocs[t.Name] = strings.TrimSpace(t.Doc)
		}
	}
	return nil
}

// RegisterDocs updates the generator's internal schemas with comments if available
// This requires the Generator to have access to the parser or TypeDocs.
// For simplicity, we can just export a function to apply docs to a Generator.
func ApplyComments(gen *Generator, root string) error {
	cp := NewCommentParser()
	if err := cp.ParseDocs(root); err != nil {
		return err
	}

	for name, schema := range gen.Spec.Components.Schemas {
		if doc, ok := cp.TypeDocs[name]; ok {
			schema.Description = doc
		}
	}
	return nil
}
