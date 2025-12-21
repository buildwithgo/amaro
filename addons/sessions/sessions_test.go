package sessions_test

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

// UserData acts as a "Fixed Struct" of types.
type UserData struct {
	Username string
	Role     string
	Views    int
}

// TestSessionWithFixedStruct verifies stronger typing using Generics.
// Manager[UserData] ensures the session data is ALWAYS UserData.
func TestSessionWithFixedStruct(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// 1. Generic Cache for UserData
	store := cache.NewMemoryCache()

	// 2. Generic Manager for UserData
	sessMgr := sessions.NewManager[UserData](store, "struct_sess", 10*time.Minute)

	// 3. Generic Middleware
	app.Use(sessions.Start(sessMgr))

	app.GET("/login", func(c *amaro.Context) error {
		// No need to "Get" first if we are initializing.
		// Use Provider to get/new
		sess := sessions.Get[UserData](c)

		// DIRECT ACCESS: Data is UserData struct! No "Set(key, val)" needed.
		sess.Data.Username = "bernardo"
		sess.Data.Role = "admin"
		sess.Data.Views = 0

		return c.String(http.StatusOK, "logged in")
	})

	app.GET("/profile", func(c *amaro.Context) error {
		sess := sessions.Get[UserData](c)

		// DIRECT ACCESS without type assertion
		sess.Data.Views++

		// No explicit "Set" - Data is modified in place (struct copy/ref considerations pending implementation details)
		// Since Session stores Data T by value in struct, but Session is pointer...
		// In previous generic impl: `Data T`. So `sess.Data.Views++` modifies `sess.Data`.

		return c.String(http.StatusOK, fmt.Sprintf("User: %s, Views: %d", sess.Data.Username, sess.Data.Views))
	})

	server := httptest.NewServer(app)
	defer server.Close()
	client := server.Client()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	// A. Login
	resp, _ := client.Get(server.URL + "/login")
	if resp.StatusCode != 200 {
		t.Errorf("Login failed")
	}

	// B. Profile 1
	resp, _ = client.Get(server.URL + "/profile")
	body := readBody(resp)
	if body != "User: bernardo, Views: 1" {
		t.Errorf("Unexpected profile: %s", body)
	}

	// C. Profile 2 (Count increment)
	resp, _ = client.Get(server.URL + "/profile")
	body = readBody(resp)
	if body != "User: bernardo, Views: 2" {
		t.Errorf("Unexpected count: %s", body)
	}
}

// TestSessionWithMap verifies the "Interfaces Map" behavior using Generics.
// Manager[map[string]interface{}] mimics the old untyped behavior.
func TestSessionWithMap(t *testing.T) {
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Map must be initialized. The default NewSession creates zero-value T.
	// For map, zero value is nil. We'll need to handle init in handler or custom NewSession logic.
	// Let's assume user inits map in handler if nil.
	type MapType map[string]interface{}

	store := cache.NewMemoryCache()
	sessMgr := sessions.NewManager[MapType](store, "map_sess", 10*time.Minute)

	app.Use(sessions.Start(sessMgr))

	app.GET("/set", func(c *amaro.Context) error {
		sess := sessions.Get[MapType](c)
		if sess.Data == nil {
			sess.Data = make(MapType)
		}
		sess.Data["foo"] = "bar"
		return c.String(200, "ok")
	})

	app.GET("/get", func(c *amaro.Context) error {
		sess := sessions.Get[MapType](c)
		val := sess.Data["foo"]
		return c.String(200, fmt.Sprintf("%v", val))
	})

	server := httptest.NewServer(app)
	defer server.Close()
	client := server.Client()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	client.Get(server.URL + "/set")
	resp, _ := client.Get(server.URL + "/get")
	body := readBody(resp)
	if body != "bar" {
		t.Errorf("Map storage failed, got: %s", body)
	}
}

func readBody(resp *http.Response) string {
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	return string(buf)
}
