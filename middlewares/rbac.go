package middlewares

import (
	"net/http"

	"github.com/buildwithgo/amaro"
)

// RBACConfig holds configuration for RBAC middleware.
type RBACConfig struct {
	// RoleExtractor extracts the role from the context.
	// The role is usually populated by a previous Auth middleware (JWT, Basic, Session).
	RoleExtractor func(c *amaro.Context) (string, error)

	// Roles is a map of Path -> []AllowedRoles.
	// OR use a policy function.
	// For simplicity, let's allow passing a required role to the generator.
	// BUT middleware is usually global or per-route.
	// If per-route, we generate it: middlewares.RBAC("admin")

	// ErrorHandler handles forbidden access.
	ErrorHandler func(c *amaro.Context, err error) error
}

// RBAC returns a middleware that enforces a required role.
func RBAC(requiredRole string, roleExtractor func(c *amaro.Context) (string, error)) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			role, err := roleExtractor(c)
			if err != nil {
				return amaro.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			if role != requiredRole {
				// Simple check. Could be hierarchical or list check.
				return amaro.NewHTTPError(http.StatusForbidden, "Forbidden")
			}

			return next(c)
		}
	}
}

// ACL is a more flexible version allowing multiple roles.
func ACL(allowedRoles []string, roleExtractor func(c *amaro.Context) (string, error)) amaro.Middleware {
	return func(next amaro.Handler) amaro.Handler {
		return func(c *amaro.Context) error {
			role, err := roleExtractor(c)
			if err != nil {
				return amaro.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			return amaro.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
	}
}
