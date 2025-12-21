package websocket_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/websocket"
	"github.com/buildwithgo/amaro/routers"
	xws "golang.org/x/net/websocket"
)

func TestWebSocket(t *testing.T) {
	// 1. Setup Amaro Server
	app := amaro.New(amaro.WithRouter(routers.NewTrieRouter()))

	// Echo handler
	app.GET("/ws", websocket.New(func(ws *xws.Conn) {
		defer ws.Close()
		var msg string
		for {
			if err := xws.Message.Receive(ws, &msg); err != nil {
				return
			}
			if err := xws.Message.Send(ws, "Echo: "+msg); err != nil {
				return
			}
		}
	}))

	// 2. Start Test Server
	ts := httptest.NewServer(app)
	defer ts.Close()

	// 3. Convert HTTP URL to WS URL
	wsURL := strings.Replace(ts.URL, "http", "ws", 1) + "/ws"

	// 4. Connect Client
	ws, err := xws.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// 5. Send Message
	message := "Hello Amaro"
	if err := xws.Message.Send(ws, message); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 6. Receive Response
	var response string
	if err := xws.Message.Receive(ws, &response); err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	// 7. Verify
	expected := "Echo: " + message
	if response != expected {
		t.Errorf("Expected '%s', got '%s'", expected, response)
	}
}
