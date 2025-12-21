package websocket

import (
	"golang.org/x/net/websocket"

	"github.com/buildwithgo/amaro"
)

// Handler is a type alias for the websocket handler function.
type Handler func(*websocket.Conn)

// New creates a new Amaro handler that upgrades the connection to a WebSocket.
// It wraps golang.org/x/net/websocket.
func New(handler Handler) amaro.Handler {
	return func(c *amaro.Context) error {
		// Create the websocket.Handler
		wsHandler := websocket.Handler(handler)

		// ServeHTTP will handle the upgrade and hijacking.
		// NOTE: x/net/websocket's ServeHTTP expects a ResponseWriter and Request.
		// It will take over the connection.
		wsHandler.ServeHTTP(c.Writer, c.Request)

		// After the websocket handler returns (connection closed), we return nil.
		return nil
	}
}
