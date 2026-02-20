package routers_test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestRouterConcurrency(t *testing.T) {
	router := routers.NewTrieRouter()

	// Pre-populate some routes
	router.Add(http.MethodGet, "/", func(c *amaro.Context) error { return nil })

	var wg sync.WaitGroup
	start := make(chan struct{})

	// Writers: Add routes concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			path := fmt.Sprintf("/route-%d", i)
			router.Add(http.MethodGet, path, func(c *amaro.Context) error { return nil })
		}(i)
	}

	// Readers: Find routes concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			// Randomly check root or non-existent
			_, _ = router.Find(http.MethodGet, "/", nil)
			_, _ = router.Find(http.MethodGet, "/non-existent", nil)
		}()
	}

	close(start)
	wg.Wait()

	// Verify routes were added
	routes := router.Routes()
	if len(routes) != 11 { // 1 initial + 10 added
		t.Errorf("Expected 11 routes, got %d", len(routes))
	}
}

func TestRouterMiddlewareConcurrency(t *testing.T) {
	router := routers.NewTrieRouter()

	var wg sync.WaitGroup

	// Add middleware concurrently while adding routes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			router.Use(func(next amaro.Handler) amaro.Handler { return next })
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			path := fmt.Sprintf("/dynamic-%d", i)
			router.Add(http.MethodGet, path, func(c *amaro.Context) error { return nil })
		}
	}()

	wg.Wait()
}
