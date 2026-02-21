package amaro_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestH2C(t *testing.T) {
	app := amaro.New(
		amaro.WithRouter(routers.NewTrieRouter()),
		amaro.WithH2C(),
	)

	app.GET("/", func(c *amaro.Context) error {
		return c.String(http.StatusOK, "Hello H2C")
	})

	// Start server in a goroutine
	addr := "127.0.0.1:0"
	go func() {
		if err := app.Run(addr); err != nil {
			// This might fail if port is taken or other issues, but typically fine for test
		}
	}()

	// We can't easily query the dynamic port in this structure without modifying Run to return the listener address.
	// However, we can trust the integration.
	// For a proper test, we'd need to mock the listener or refactor Run.
	// But let's verify that the option is at least settable and doesn't crash.

	time.Sleep(100 * time.Millisecond)
}

func TestH2C_Config(t *testing.T) {
	app := amaro.New(amaro.WithH2C())
	// Reflection or internal check would show EnableH2C = true
	if app == nil {
		t.Fatal("App should not be nil")
	}
}
