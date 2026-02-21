package middlewares_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/middlewares"
)

func TestTimeoutRace(t *testing.T) {
	// This test attempts to trigger the data race in Timeout middleware.
	// Run with 'go test -race'

	// Create a handler that takes longer than the timeout
	longRunningHandler := func(c *amaro.Context) error {
		time.Sleep(50 * time.Millisecond)
		return c.String(http.StatusOK, "Too late")
	}

	// Apply Timeout middleware
	m := middlewares.Timeout(10 * time.Millisecond)
	h := m(longRunningHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Create context
	ctx := amaro.NewContext(w, req)

	// Execute handler
	_ = h(ctx)
}

func TestTimeoutPanic(t *testing.T) {
	// Ensure panic in handler is recovered and returned as error
	m := middlewares.Timeout(100 * time.Millisecond)

	handler := m(func(c *amaro.Context) error {
		panic("oops")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := amaro.NewContext(w, req)

	err := handler(ctx)
	if err == nil {
		t.Fatal("Expected error from panic, got nil")
	}
	expected := "panic in timeout handler: oops"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestTimeoutErrorDiscard(t *testing.T) {
	// Ensure buffer is discarded on error
	m := middlewares.Timeout(100 * time.Millisecond)

	handler := m(func(c *amaro.Context) error {
		c.String(http.StatusOK, "should be discarded")
		return fmt.Errorf("handler error")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := amaro.NewContext(w, req)

	err := handler(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Body should be empty (discarded)
	if w.Body.Len() > 0 {
		t.Errorf("Expected empty body, got '%s'", w.Body.String())
	}
}
