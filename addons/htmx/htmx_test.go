package htmx_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/buildwithgo/amaro"
	"github.com/buildwithgo/amaro/addons/htmx"
)

func TestHTMX(t *testing.T) {
	t.Run("Is", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("HX-Request", "true")
		w := httptest.NewRecorder()
		c := amaro.NewContext(w, req)

		if !htmx.Is(c) {
			t.Error("Expected IsHTMX to return true")
		}
	})

	t.Run("Trigger", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		c := amaro.NewContext(w, req)

		htmx.Trigger(c, "myEvent")
		if w.Header().Get("HX-Trigger") != "myEvent" {
			t.Errorf("Expected HX-Trigger header 'myEvent', got %s", w.Header().Get("HX-Trigger"))
		}
	})

	t.Run("TriggerJSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		c := amaro.NewContext(w, req)

		events := map[string]any{
			"event1": "data1",
			"event2": 123,
		}
		if err := htmx.TriggerJSON(c, events); err != nil {
			t.Fatalf("TriggerJSON failed: %v", err)
		}

		header := w.Header().Get("HX-Trigger")
		var decoded map[string]any
		if err := json.Unmarshal([]byte(header), &decoded); err != nil {
			t.Fatalf("Failed to unmarshal HX-Trigger header: %v", err)
		}

		if decoded["event1"] != "data1" {
			t.Errorf("Expected event1 data1, got %v", decoded["event1"])
		}
	})

	t.Run("Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		c := amaro.NewContext(w, req)

		htmx.PushURL(c, "/new-url")
		htmx.Redirect(c, "/redirect")
		htmx.Refresh(c)
		htmx.Retarget(c, "#target")
		htmx.Reswap(c, "outerHTML")

		if w.Header().Get("HX-Push-Url") != "/new-url" {
			t.Error("PushURL failed")
		}
		if w.Header().Get("HX-Redirect") != "/redirect" {
			t.Error("Redirect failed")
		}
		if w.Header().Get("HX-Refresh") != "true" {
			t.Error("Refresh failed")
		}
		if w.Header().Get("HX-Retarget") != "#target" {
			t.Error("Retarget failed")
		}
		if w.Header().Get("HX-Reswap") != "outerHTML" {
			t.Error("Reswap failed")
		}
	})
}
