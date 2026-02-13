package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/middlewares"
)

func TestRateLimiter(t *testing.T) {
	// 5 req/sec, burst 1
	limiter := middlewares.RateLimiter(5, 1)

	handler := limiter(func(c *amaro.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	performRequest := func() int {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:1234" // Mock IP
		w := httptest.NewRecorder()
		ctx := amaro.NewContext(w, req)
		_ = handler(ctx)
		return w.Code
	}

	// First request: OK
	if code := performRequest(); code != http.StatusOK {
		t.Errorf("First request should be OK, got %d", code)
	}

	// Second request immediately: Should fail (burst 1 used)
	if code := performRequest(); code != http.StatusTooManyRequests {
		t.Errorf("Second request should be 429, got %d", code)
	}

	// Wait 200ms (1/5 sec): Token should refill
	time.Sleep(250 * time.Millisecond)
	if code := performRequest(); code != http.StatusOK {
		t.Errorf("Request after wait should be OK, got %d", code)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	// High rate to avoid limiting, test concurrency safety
	limiter := middlewares.RateLimiter(1000, 1000)

	handler := limiter(func(c *amaro.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			w := httptest.NewRecorder()
			ctx := amaro.NewContext(w, req)
			_ = handler(ctx)
		}()
	}
	wg.Wait()
}
