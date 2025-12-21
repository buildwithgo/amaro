package sessions

import (
	"net/http"
	"time"

	"github.com/buildwithgo/amaro"
)

const ContextKey = "session"

// Start returns a generic middleware that handles session lifecycle for type T.
func Start[T any](p Provider[T]) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			cookieName, ttl := p.CookieConfig()

			// 1. Extract Session ID from Cookie
			cookie, err := c.GetCookie(cookieName)
			var sessionID string
			if err == nil {
				sessionID = cookie.Value
			}

			// 2. Retrieve/Create Session (Typed)
			session, err := p.Get(sessionID)
			if err != nil {
				session = p.NewSession()
			}

			// 3. Inject into Context
			c.Set(ContextKey, session)

			// 4. Set Cookie (Header)
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     cookieName,
				Value:    session.ID,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(ttl),
			})

			// 5. Call Next Handler
			err = next(c)

			// 6. Save Session
			p.Save(session)

			return err
		}
	}
}

// Get retrieves the typed session from the context.
func Get[T any](c *amaro.Context) *Session[T] {
	if val, ok := c.Get(ContextKey); ok {
		// Type assertion to *Session[T]
		if s, ok := val.(*Session[T]); ok {
			return s
		}
	}
	return nil
}
