package amaro_test

import (
	"embed"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestStaticFS(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Create a temp file to serve
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(filePath, []byte("Hello Static"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. DirFS (using .local as requested)
	// Make sure .local/hello.txt exists (created by setup or we create here)
	os.Mkdir(".local", 0755)
	os.WriteFile(filepath.Join(".local", "hello.txt"), []byte("Hello Local"), 0644)

	app.StaticFS("/files", os.DirFS(".local"))

	t.Run("ServeFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/files/hello.txt", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if w.Body.String() != "Hello Local" {
			t.Errorf("Expected 'Hello Local', got '%s'", w.Body.String())
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/files/missing.txt", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	// 2. EmbedFS
	// With //go:embed .local/*, the fs root contains ".local" directory.
	// We usually want to strip that or serve from root.
	// Let's serve strict.

	var localFS embed.FS

	// We need to use fs.Sub to get into .local if we want /files/hello.txt to map to .local/hello.txt directly
	// without the client adding .local in path.
	subFS, _ := fs.Sub(localFS, ".local")
	app.StaticFS("/embed", subFS)

	t.Run("ServeEmbed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/embed/hello.txt", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if w.Body.String() != "Hello Local" {
			t.Errorf("Expected 'Hello Local', got '%s'", w.Body.String())
		}

		// Verify new test file
		req2 := httptest.NewRequest("GET", "/embed/test_file.txt", nil)
		w2 := httptest.NewRecorder()
		app.ServeHTTP(w2, req2)
		// Trim space because echo might add newline
		if w2.Code != http.StatusOK {
			t.Errorf("Expected 200 for test_file.txt, got %d", w2.Code)
		}
	})
}
