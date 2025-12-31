package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/buildwithgo/amaro"
	"golang.org/x/oauth2"
)

// Config holds OAuth2 configuration.
type Config struct {
	oauth2.Config

	// SuccessHandler is called after successful token exchange.
	// It should handle session creation or token response.
	SuccessHandler func(c *amaro.Context, token *oauth2.Token) error

	// ErrorHandler handles errors during the flow.
	ErrorHandler func(c *amaro.Context, err error) error

	// StateGenerator generates the state string.
	StateGenerator func(c *amaro.Context) string

	// StateValidator validates the state string.
	StateValidator func(c *amaro.Context, state string) bool
}

// LoginHandler returns a handler that redirects to the OAuth2 provider.
func LoginHandler(config *Config) amaro.Handler {
	return func(c *amaro.Context) error {
		state := ""
		if config.StateGenerator != nil {
			state = config.StateGenerator(c)
		}
		url := config.AuthCodeURL(state)
		return c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// CallbackHandler returns a handler that processes the OAuth2 callback.
func CallbackHandler(config *Config) amaro.Handler {
	return func(c *amaro.Context) error {
		code := c.QueryParam("code")
		state := c.QueryParam("state")

		if config.StateValidator != nil {
			if !config.StateValidator(c, state) {
				return config.ErrorHandler(c, fmt.Errorf("invalid state"))
			}
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			if config.ErrorHandler != nil {
				return config.ErrorHandler(c, err)
			}
			return err
		}

		if config.SuccessHandler != nil {
			return config.SuccessHandler(c, token)
		}

		return c.JSON(http.StatusOK, token)
	}
}
