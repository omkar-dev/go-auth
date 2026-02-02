package goauth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMode defines the authentication enforcement mode
type AuthMode int

const (
	// AuthModeStrict returns 401 for missing or invalid tokens (default, recommended)
	AuthModeStrict AuthMode = iota
	// AuthModeOptional allows requests without valid tokens to proceed
	AuthModeOptional
)

// AuthConfig configures the JWT authentication middleware
type AuthConfig struct {
	// Mode determines authentication enforcement behavior
	// Default: AuthModeStrict (returns 401 for missing/invalid tokens)
	Mode AuthMode

	// PublicPaths is a list of path prefixes that bypass authentication
	// Default: empty (all paths require authentication)
	PublicPaths []string

	// PublicPathMatcher is a custom function to determine if a path is public
	// If set, this takes precedence over PublicPaths
	PublicPathMatcher func(path string) bool
}

// JWTAuth creates a JWT authentication middleware with strict mode (returns 401 for invalid auth).
// For more control, use JWTAuthWithConfig.
//
// This is the RECOMMENDED middleware for production use.
func JWTAuth(next http.Handler) http.Handler {
	return JWTAuthWithConfig(AuthConfig{Mode: AuthModeStrict})(next)
}

// JWTAuthOptional creates a JWT authentication middleware that allows requests without valid tokens.
// Use this only if you need to support both authenticated and unauthenticated access on the same endpoint.
//
// WARNING: Handlers MUST check if UserContextKey is present to enforce authorization.
func JWTAuthOptional(next http.Handler) http.Handler {
	return JWTAuthWithConfig(AuthConfig{Mode: AuthModeOptional})(next)
}

// JWTAuthWithConfig creates a JWT authentication middleware with custom configuration.
//
// Example:
//
//	middleware := goauth.JWTAuthWithConfig(goauth.AuthConfig{
//	    Mode: goauth.AuthModeStrict,
//	    PublicPaths: []string{"/health", "/metrics"},
//	})
func JWTAuthWithConfig(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a public endpoint
			if isPublicPath(r.URL.Path, config) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if config.Mode == AuthModeStrict {
					http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
					return
				}
				// Optional mode: allow through without auth
				next.ServeHTTP(w, r)
				return
			}

			// Parse Bearer token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				// No "Bearer " prefix, use the header value as-is
				token = authHeader
			}

			// Validate token
			claims, err := ValidateToken(token, "access")
			if err != nil {
				slog.Warn("Invalid JWT token", slog.Any("error", err))
				if config.Mode == AuthModeStrict {
					http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
					return
				}
				// Optional mode: allow through without auth
				next.ServeHTTP(w, r)
				return
			}

			// Token is valid - inject user claims into context
			userClaims := map[string]interface{}{
				"sub":     claims.MemberID,
				"phone":   claims.Phone,
				"org_id":  claims.OrgID,
				"org_ids": []string{claims.OrgID},
			}

			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicPath determines if a path should bypass authentication
func isPublicPath(path string, config AuthConfig) bool {
	// Use custom matcher if provided
	if config.PublicPathMatcher != nil {
		return config.PublicPathMatcher(path)
	}

	// Check against PublicPaths list
	for _, publicPath := range config.PublicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

// Helper function to get user claims from context
// Returns nil if no authenticated user
func GetUser(ctx context.Context) map[string]interface{} {
	user, _ := ctx.Value(UserContextKey).(map[string]interface{})
	return user
}

// Helper function to check if request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return ctx.Value(UserContextKey) != nil
}
