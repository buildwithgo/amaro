package cache_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/cache"
	"github.com/buildwithgo/amaro/routers"
)

func TestCachePage(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))
	store := cache.NewMemoryCache()

	// Cached Endpoint
	app.GET("/time", func(c *amaro.Context) error {
		return c.String(http.StatusOK, time.Now().Format(time.RFC3339))
	}, cache.CachePage(store, 1*time.Second))

	server := httptest.NewServer(app)
	defer server.Close()
	client := server.Client()

	// A. First Hit
	resp, err := client.Get(server.URL + "/time")
	if err != nil {
		t.Fatal(err)
	}
	firstTime := readBody(resp)

	// B. Second Hit (Should be cached)
	resp, err = client.Get(server.URL + "/time")
	if err != nil {
		t.Fatal(err)
	}
	secondTime := readBody(resp)

	if firstTime != secondTime {
		t.Errorf("Cache failed. Got different times: %s vs %s", firstTime, secondTime)
	}

	// Verify Cache Header
	if resp.Header.Get("X-Cache") != "HIT" {
		t.Error("Expected X-Cache: HIT header")
	}
}

func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	return string(buf)
}
