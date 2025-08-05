package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTMiddleware(t *testing.T) {
	// Create a simple test handler
	testHandler := func(c *amaro.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Test 1: Valid token
	t.Run("ValidToken", func(t *testing.T) {
		// Create middleware
		middleware := JWT(WithSecret("test-secret"))
		handler := middleware(testHandler)

		// Create test token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatal(err)
		}

		// Create request
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		// Create context and call handler
		ctx := amaro.NewContext(w, req)
		err = handler(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	// Test 2: Missing token
	t.Run("MissingToken", func(t *testing.T) {
		// Create a middleware with custom error handler that returns error
		middleware := JWT(
			WithSecret("test-secret"),
			WithErrorHandler(func(c *amaro.Context, err error) error {
				c.Writer.WriteHeader(http.StatusUnauthorized)
				return err // Return the error so we can test it
			}),
		)
		handler := middleware(testHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		ctx := amaro.NewContext(w, req)
		err := handler(ctx)

		// Should return an error (which gets handled by ErrorHandler)
		if err == nil {
			t.Error("Expected error for missing token")
		}

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 3: Invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		middleware := JWT(
			WithSecret("test-secret"),
			WithErrorHandler(func(c *amaro.Context, err error) error {
				c.Writer.WriteHeader(http.StatusUnauthorized)
				return err
			}),
		)
		handler := middleware(testHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		ctx := amaro.NewContext(w, req)
		err := handler(ctx)

		if err == nil {
			t.Error("Expected error for invalid token")
		}

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 4: Expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		middleware := JWT(
			WithSecret("test-secret"),
			WithErrorHandler(func(c *amaro.Context, err error) error {
				c.Writer.WriteHeader(http.StatusUnauthorized)
				return err
			}),
		)
		handler := middleware(testHandler)

		// Create expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
			"iat": time.Now().Add(-2 * time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		ctx := amaro.NewContext(w, req)
		err = handler(ctx)

		if err == nil {
			t.Error("Expected error for expired token")
		}

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	// Test 5: Query parameter token
	t.Run("QueryParamToken", func(t *testing.T) {
		middleware := JWT(
			WithSecret("test-secret"),
			WithTokenLookup("query:token"),
			WithAuthScheme(""),
		)
		handler := middleware(testHandler)

		// Create test token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/protected?token="+tokenString, nil)
		w := httptest.NewRecorder()

		ctx := amaro.NewContext(w, req)
		err = handler(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test 6: Skipper function
	t.Run("SkipperFunction", func(t *testing.T) {
		middleware := JWT(
			WithSecret("test-secret"),
			WithSkipper(func(c *amaro.Context) bool {
				return c.Request.URL.Path == "/public"
			}),
		)
		handler := middleware(testHandler)

		req := httptest.NewRequest("GET", "/public", nil)
		w := httptest.NewRecorder()

		ctx := amaro.NewContext(w, req)
		err := handler(ctx)

		// Should not require authentication due to skipper
		if err != nil {
			t.Errorf("Expected no error for skipped route, got %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestTokenExtraction(t *testing.T) {
	config := DefaultJWTConfig()
	config.Secret = []byte("test-secret")

	// Test header extraction
	t.Run("HeaderExtraction", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		ctx := amaro.NewContext(httptest.NewRecorder(), req)

		config.TokenLookup = "header:Authorization"
		config.AuthScheme = "Bearer"

		token, err := extractToken(ctx, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token != "test-token" {
			t.Errorf("Expected 'test-token', got '%s'", token)
		}
	})

	// Test query extraction
	t.Run("QueryExtraction", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?access_token=test-token", nil)
		ctx := amaro.NewContext(httptest.NewRecorder(), req)

		config.TokenLookup = "query:access_token"
		config.AuthScheme = ""

		token, err := extractToken(ctx, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token != "test-token" {
			t.Errorf("Expected 'test-token', got '%s'", token)
		}
	})

	// Test cookie extraction
	t.Run("CookieExtraction", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "jwt",
			Value: "test-token",
		})
		ctx := amaro.NewContext(httptest.NewRecorder(), req)

		config.TokenLookup = "cookie:jwt"
		config.AuthScheme = ""

		token, err := extractToken(ctx, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token != "test-token" {
			t.Errorf("Expected 'test-token', got '%s'", token)
		}
	})
}

func TestCreateToken(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":  "user123",
		"name": "John Doe",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	config := DefaultJWTConfig()
	config.Secret = []byte("test-secret")

	token, err := CreateToken(claims, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify the token can be parsed back
	parsedToken, err := parseToken(token, config)
	if err != nil {
		t.Errorf("Expected no error parsing created token, got %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Created token should be valid")
	}

	if parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if parsedClaims["sub"] != "user123" {
			t.Errorf("Expected sub claim 'user123', got '%v'", parsedClaims["sub"])
		}
	} else {
		t.Error("Expected MapClaims")
	}
}
