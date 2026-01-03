package htmx

import (
	"encoding/json"

	"github.com/buildwithgo/amaro"
)

// Is returns true if the request is an HTMX request.
func Is(c *amaro.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

// Trigger sets the HX-Trigger header to trigger a client-side event.
func Trigger(c *amaro.Context, event string) {
	c.SetHeader("HX-Trigger", event)
}

// TriggerJSON sets the HX-Trigger header with a JSON object for passing data to events.
func TriggerJSON(c *amaro.Context, events map[string]any) error {
	b, err := json.Marshal(events)
	if err != nil {
		return err
	}
	c.SetHeader("HX-Trigger", string(b))
	return nil
}

// PushURL sets the HX-Push-Url header to push a new URL into the history stack.
func PushURL(c *amaro.Context, url string) {
	c.SetHeader("HX-Push-Url", url)
}

// Redirect sets the HX-Redirect header to force a client-side redirect.
func Redirect(c *amaro.Context, url string) {
	c.SetHeader("HX-Redirect", url)
}

// Refresh sets the HX-Refresh header to force a full page refresh.
func Refresh(c *amaro.Context) {
	c.SetHeader("HX-Refresh", "true")
}

// Retarget sets the HX-Retarget header to update a different element than the one triggering the request.
func Retarget(c *amaro.Context, target string) {
	c.SetHeader("HX-Retarget", target)
}

// Reswap sets the HX-Reswap header to specify how the response should be swapped in.
func Reswap(c *amaro.Context, swap string) {
	c.SetHeader("HX-Reswap", swap)
}
