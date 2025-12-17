package middlewares

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/buildwithgo/amaro"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds the configuration for JWT middleware
type JWTConfig struct {
	// Secret key for HMAC signing
	Secret []byte

	// RSA public key for RSA signing verification
	PublicKey *rsa.PublicKey

	// JKS keystore configuration
	JKSConfig *JKSConfig

	// Token lookup configuration
	TokenLookup string // "header:Authorization", "query:token", "cookie:jwt"

	// Auth scheme for header lookup
	AuthScheme string // "Bearer"

	// Claims key to store user data in context
	ContextKey string

	// Error handler
	ErrorHandler func(*amaro.Context, error) error

	// Success handler called after successful validation
	SuccessHandler func(*amaro.Context, jwt.Token) error

	// Skipper function to skip middleware for certain requests
	Skipper func(*amaro.Context) bool

	// Signing method
	SigningMethod jwt.SigningMethod
}

// JKSConfig holds Java KeyStore configuration
type JKSConfig struct {
	KeystoreData []byte
	Password     string
	Alias        string
}

// JWTOption is a function type for configuring JWT middleware
type JWTOption func(*JWTConfig)

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		TokenLookup:   "header:Authorization",
		AuthScheme:    "Bearer",
		ContextKey:    "user",
		SigningMethod: jwt.SigningMethodHS256,
		ErrorHandler: func(c *amaro.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "unauthorized",
				"message": err.Error(),
			})
		},
		Skipper: func(c *amaro.Context) bool {
			return false
		},
	}
}

// WithSecret sets the HMAC secret
func WithSecret(secret string) JWTOption {
	return func(config *JWTConfig) {
		config.Secret = []byte(secret)
		config.SigningMethod = jwt.SigningMethodHS256
	}
}

// WithSecretBytes sets the HMAC secret from bytes
func WithSecretBytes(secret []byte) JWTOption {
	return func(config *JWTConfig) {
		config.Secret = secret
		config.SigningMethod = jwt.SigningMethodHS256
	}
}

// WithRSAPublicKey sets the RSA public key for verification
func WithRSAPublicKey(publicKey *rsa.PublicKey) JWTOption {
	return func(config *JWTConfig) {
		config.PublicKey = publicKey
		config.SigningMethod = jwt.SigningMethodRS256
	}
}

// WithRSAPublicKeyFromPEM sets the RSA public key from PEM string
func WithRSAPublicKeyFromPEM(pemStr string) JWTOption {
	return func(config *JWTConfig) {
		block, _ := pem.Decode([]byte(pemStr))
		if block == nil {
			return
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return
		}

		if rsaPubKey, ok := pubKey.(*rsa.PublicKey); ok {
			config.PublicKey = rsaPubKey
			config.SigningMethod = jwt.SigningMethodRS256
		}
	}
}

// WithJKS sets the JKS configuration
func WithJKS(keystoreData []byte, password, alias string) JWTOption {
	return func(config *JWTConfig) {
		config.JKSConfig = &JKSConfig{
			KeystoreData: keystoreData,
			Password:     password,
			Alias:        alias,
		}
		config.SigningMethod = jwt.SigningMethodRS256
	}
}

// WithTokenLookup sets where to look for the token
func WithTokenLookup(lookup string) JWTOption {
	return func(config *JWTConfig) {
		config.TokenLookup = lookup
	}
}

// WithAuthScheme sets the authorization scheme
func WithAuthScheme(scheme string) JWTOption {
	return func(config *JWTConfig) {
		config.AuthScheme = scheme
	}
}

// WithContextKey sets the context key for storing claims
func WithContextKey(key string) JWTOption {
	return func(config *JWTConfig) {
		config.ContextKey = key
	}
}

// WithErrorHandler sets custom error handler
func WithErrorHandler(handler func(*amaro.Context, error) error) JWTOption {
	return func(config *JWTConfig) {
		config.ErrorHandler = handler
	}
}

// WithSuccessHandler sets custom success handler
func WithSuccessHandler(handler func(*amaro.Context, jwt.Token) error) JWTOption {
	return func(config *JWTConfig) {
		config.SuccessHandler = handler
	}
}

// WithSkipper sets the skipper function
func WithSkipper(skipper func(*amaro.Context) bool) JWTOption {
	return func(config *JWTConfig) {
		config.Skipper = skipper
	}
}

// WithSigningMethod sets the signing method
func WithSigningMethod(method jwt.SigningMethod) JWTOption {
	return func(config *JWTConfig) {
		config.SigningMethod = method
	}
}

// JWT creates a new JWT middleware with the given options
func JWT(opts ...JWTOption) amaro.Middleware {
	config := DefaultJWTConfig()

	for _, opt := range opts {
		opt(config)
	}

	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			// Skip if skipper returns true
			if config.Skipper(c) {
				return next(c)
			}

			// Extract token from request
			token, err := extractToken(c, config)
			if err != nil {
				return config.ErrorHandler(c, err)
			}

			// Parse and validate token
			parsedToken, err := parseToken(token, config)
			if err != nil {
				return config.ErrorHandler(c, err)
			}

			// Store claims in context (you might need to extend Context to support this)
			if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
				// For now, we'll store it in a header or you can extend the Context struct
				c.SetHeader("X-JWT-Claims", fmt.Sprintf("%v", claims))
			}

			// Call success handler if provided
			if config.SuccessHandler != nil {
				if err := config.SuccessHandler(c, *parsedToken); err != nil {
					return config.ErrorHandler(c, err)
				}
			}

			return next(c)
		}
	}
}

// extractToken extracts the JWT token from the request
func extractToken(c *amaro.Context, config *JWTConfig) (string, error) {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid token lookup format")
	}

	method := parts[0]
	key := parts[1]

	switch method {
	case "header":
		auth := c.GetHeader(key)
		if auth == "" {
			return "", errors.New("missing authorization header")
		}

		if config.AuthScheme != "" {
			prefix := config.AuthScheme + " "
			if !strings.HasPrefix(auth, prefix) {
				return "", fmt.Errorf("invalid authorization scheme, expected %s", config.AuthScheme)
			}
			return strings.TrimPrefix(auth, prefix), nil
		}
		return auth, nil

	case "query":
		token := c.QueryParam(key)
		if token == "" {
			return "", errors.New("missing token in query parameters")
		}
		return token, nil

	case "cookie":
		cookie, err := c.GetCookie(key)
		if err != nil {
			return "", errors.New("missing token in cookie")
		}
		return cookie.Value, nil

	default:
		return "", errors.New("unsupported token lookup method")
	}
}

// parseToken parses and validates the JWT token
func parseToken(tokenString string, config *JWTConfig) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method != config.SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		switch config.SigningMethod {
		case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
			if config.Secret == nil {
				return nil, errors.New("HMAC secret not configured")
			}
			return config.Secret, nil

		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
			if config.PublicKey != nil {
				return config.PublicKey, nil
			}

			if config.JKSConfig != nil {
				// Extract public key from JKS
				pubKey, err := extractPublicKeyFromJKS(config.JKSConfig)
				if err != nil {
					return nil, fmt.Errorf("failed to extract public key from JKS: %v", err)
				}
				return pubKey, nil
			}

			return nil, errors.New("RSA public key not configured")

		default:
			return nil, fmt.Errorf("unsupported signing method: %v", config.SigningMethod)
		}
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check standard claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, errors.New("token has expired")
			}
		}

		// Check not before
		if nbf, ok := claims["nbf"].(float64); ok {
			if time.Now().Unix() < int64(nbf) {
				return nil, errors.New("token not valid yet")
			}
		}

		// Check issued at (optional validation)
		if iat, ok := claims["iat"].(float64); ok {
			if time.Now().Unix() < int64(iat) {
				return nil, errors.New("token issued in the future")
			}
		}
	}

	return token, nil
}

// extractPublicKeyFromJKS extracts a public key from JKS keystore
// Note: This is a simplified implementation. For production use,
// consider using a proper JKS library like github.com/pavel-v-chernykh/keystore-go
func extractPublicKeyFromJKS(_ *JKSConfig) (*rsa.PublicKey, error) {
	// This is a placeholder implementation
	// In a real scenario, you would need to:
	// 1. Parse the JKS file format
	// 2. Extract the certificate for the given alias
	// 3. Get the public key from the certificate

	return nil, errors.New("JKS support requires additional implementation - consider using github.com/pavel-v-chernykh/keystore-go")
}

// Helper functions for creating JWT tokens (useful for testing)

// CreateToken creates a JWT token with the given claims and config
func CreateToken(claims jwt.MapClaims, config *JWTConfig) (string, error) {
	token := jwt.NewWithClaims(config.SigningMethod, claims)

	switch config.SigningMethod {
	case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
		if config.Secret == nil {
			return "", errors.New("HMAC secret not configured")
		}
		return token.SignedString(config.Secret)

	case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
		// For token creation, you would typically use a private key
		// This is just for demonstration - in practice, token creation
		// would be done by an authentication service with access to private keys
		return "", errors.New("RSA token creation requires private key")

	default:
		return "", fmt.Errorf("unsupported signing method: %v", config.SigningMethod)
	}
}
