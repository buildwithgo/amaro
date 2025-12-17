package amaro_test

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/openapi"
	"github.com/buildwithgo/amaro/addons/streaming"
	"github.com/buildwithgo/amaro/middlewares"
	"github.com/buildwithgo/amaro/routers"
)

func TestStreamingIntegration(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Use Compress middleware to verify it doesn't break streaming
	app.Use(middlewares.Compress())

	gen := openapi.NewGenerator(openapi.Info{Title: "Stream API", Version: "1.0"})

	// Define streaming handler
	streamHandler := func(c *amaro.Context) error {
		return streaming.StreamText(c, func(st *streaming.StreamTextContext) {
			st.WriteLn("chunk1")
			c.Writer.(http.Flusher).Flush() // explicit flush
			st.WriteLn("chunk2")
			c.Writer.(http.Flusher).Flush()
		})
	}

	app.GET("/stream", openapi.WrapStream(gen, "GET", "/stream", streamHandler))

	app.GET("/swagger.json", func(c *amaro.Context) error {
		return c.JSON(http.StatusOK, gen.Spec)
	})

	t.Run("StreamingThroughCompress", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream", nil)
		req.Header.Set("Accept-Encoding", "gzip") // Request Compression
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		// Check Content-Encoding
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected gzip response")
		}

		// Check Transfer-Encoding or flushing behavior is hard in httptest.ResponseRecorder
		// because it buffers. But we can check that we got the data.
		// Detailed flush timing verification usually requires a real network conn or pipe.
		// For now, verification is that it didn't panic and returned correct gzipped data.

		gr, err := gzip.NewReader(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer gr.Close()

		scanner := bufio.NewScanner(gr)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if len(lines) != 2 || lines[0] != "chunk1" || lines[1] != "chunk2" {
			t.Errorf("Unexpected body lines: %v", lines)
		}
	})

	t.Run("OpenAPIRegistration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/swagger.json", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		var spec openapi.OpenAPI
		json.Unmarshal(w.Body.Bytes(), &spec)

		path := spec.Paths["/stream"]
		if path == nil || path.Get == nil {
			t.Fatal("Spec missing /stream path")
		}

		content := path.Get.Responses["200"].Content
		if _, ok := content["text/event-stream"]; !ok {
			t.Error("Spec missing text/event-stream content type")
		}
	})
}
