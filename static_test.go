package amaro_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

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

	// 2. EmbedFS Simulation using fstest.MapFS
	// This mocks an embedded filesystem where .local is the root or a subdir.
	// We use MapFS to simulate "embedded" files.

	mockFS := fstest.MapFS{
		".local/hello.txt":     &fstest.MapFile{Data: []byte("Hello Local")},
		".local/test_file.txt": &fstest.MapFile{Data: []byte("Content of test file")},
	}

	// Sub to get into .local
	subFS, _ := fs.Sub(mockFS, ".local")
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

		if w2.Code != http.StatusOK {
			t.Errorf("Expected 200 for test_file.txt, got %d", w2.Code)
		}
	})
}
