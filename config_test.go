package amaro_test

import (
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/routers"
)

func TestServerConfig(t *testing.T) {
	// Custom config
	config := amaro.ServerConfig{
		ReadTimeout: 1 * time.Second,
	}

	app := amaro.New(
		amaro.WithRouter(routers.NewTrieRouter()),
		amaro.WithServerConfig(config),
	)

	if app == nil {
		t.Fatal("App should not be nil")
	}

	go func() {
		// Try to run on a random port to ensure it doesn't crash
		_ = app.Run(":0")
	}()
	time.Sleep(100 * time.Millisecond)
}

func TestGranularServerConfig(t *testing.T) {
	app := amaro.New(
		amaro.WithRouter(routers.NewTrieRouter()),
		amaro.WithReadTimeout(2*time.Second),
		amaro.WithWriteTimeout(4*time.Second),
		amaro.WithIdleTimeout(10*time.Second),
		amaro.WithReadHeaderTimeout(1*time.Second),
		amaro.WithMaxHeaderBytes(1024),
	)

	if app == nil {
		t.Fatal("App should not be nil")
	}

	// Ensure it starts
	go func() {
		_ = app.Run(":0")
	}()
	time.Sleep(100 * time.Millisecond)
}

func TestDefaultServerConfig(t *testing.T) {
	config := amaro.DefaultServerConfig()
	if config.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("Expected ReadHeaderTimeout 5s, got %v", config.ReadHeaderTimeout)
	}
}
