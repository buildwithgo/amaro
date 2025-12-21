package addons_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/cache"
	"github.com/buildwithgo/amaro/addons/sessions"
	"github.com/buildwithgo/amaro/routers"
)

// UserData acts as a "Fixed Struct" of types, demonstrating
// that the context keys can hold strongly typed structs.
type UserData struct {
	Username string
	Role     string
	Views    int
}

func TestAddonsIntegration(t *testing.T) {
	// 1. Setup App & Decoupled Backend
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))
	store := cache.NewMemoryCache()
	// FIX: Explicitly instantiate Generic Manager with Map Type to support Set/Get usage below
	sessMgr := sessions.NewManager[map[string]interface{}](store, "test_sess", 10*time.Minute)

	app.Use(sessions.Start(sessMgr))

	// 2. Define Routes
	app.GET("/login", func(c *amaro.Context) error {
		// Generic Get for Map Type
		sess := sessions.Get[map[string]interface{}](c)

		// Store a simple value
		sess.Set("username", "bernardo")

		// Store a Struct (Simulating "fixed struct of types" as requested)
		data := UserData{Username: "bernardo", Role: "admin", Views: 0}
		sess.Set("data", data)

		return c.String(http.StatusOK, "logged in")
	})

	app.GET("/profile", func(c *amaro.Context) error {
		sess := sessions.Get[map[string]interface{}](c)
		val := sess.Get("data")
		if val == nil {
			return c.String(401, "no data")
		}

		// Type Assertion retrieves the "Fixed Struct"
		data := val.(UserData)
		data.Views++
		sess.Set("data", data) // Save back

		return c.String(http.StatusOK, fmt.Sprintf("User: %s, Role: %s, Views: %d", data.Username, data.Role, data.Views))
	})

	// Cached Endpoint
	app.GET("/time", func(c *amaro.Context) error {
		return c.String(http.StatusOK, time.Now().Format(time.RFC3339))
	}, cache.CachePage(store, 1*time.Second))

	// 3. Run Tests
	server := httptest.NewServer(app)
	defer server.Close()
	client := server.Client() // Handles cookies automatically
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	t.Run("Session Flow with Structs", func(t *testing.T) {
		// A. Login
		resp, err := client.Get(server.URL + "/login")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Login failed: %d", resp.StatusCode)
		}

		// B. Profile (Should have session data)
		resp, err = client.Get(server.URL + "/profile")
		if err != nil {
			t.Fatal(err)
		}
		body := readBody(resp)
		if body != "User: bernardo, Role: admin, Views: 1" {
			t.Errorf("Unexpected profile response: %s", body)
		}

		// C. Profile again (Increment Views struct field)
		resp, err = client.Get(server.URL + "/profile")
		if err != nil {
			t.Fatal(err)
		}
		body = readBody(resp)
		if body != "User: bernardo, Role: admin, Views: 2" {
			t.Errorf("Unexpected profile count: %s", body)
		}
	})

	t.Run("Cache Page", func(t *testing.T) {
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
	})
}

func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	return string(buf)
}
