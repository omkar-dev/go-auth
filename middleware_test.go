package goauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestJWTAuthStrict tests strict mode (default) returns 401 for missing/invalid auth
func TestJWTAuthStrict(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := JWTAuth(handler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: missing authorization header\n",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: invalid token\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestJWTAuthStrictWithValidToken tests that valid tokens work in strict mode
func TestJWTAuthStrictWithValidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid token
	tokens, err := GenerateTokenPair("user123", "+1234567890", "org456")
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	var capturedUser map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedUser == nil {
		t.Error("Expected user context to be set, got nil")
	}

	if capturedUser["sub"] != "user123" {
		t.Errorf("Expected user sub 'user123', got %v", capturedUser["sub"])
	}
}

// TestJWTAuthOptional tests optional mode allows requests without auth
func TestJWTAuthOptional(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAuthenticated(r.Context()) {
			w.Write([]byte("authenticated"))
		} else {
			w.Write([]byte("anonymous"))
		}
	})

	middleware := JWTAuthOptional(handler)

	tests := []struct {
		name         string
		authHeader   string
		expectedBody string
	}{
		{
			name:         "no auth header",
			authHeader:   "",
			expectedBody: "anonymous",
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalid-token",
			expectedBody: "anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/optional", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

// TestJWTAuthWithConfigPublicPaths tests public path configuration
func TestJWTAuthWithConfigPublicPaths(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := JWTAuthWithConfig(AuthConfig{
		Mode:        AuthModeStrict,
		PublicPaths: []string{"/health", "/metrics"},
	})(handler)

	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "public path without auth",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "public path with prefix without auth",
			path:           "/metrics/detailed",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected path without auth",
			path:           "/api/protected",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestJWTAuthWithConfigCustomMatcher tests custom path matcher
func TestJWTAuthWithConfigCustomMatcher(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Custom matcher that makes paths ending in .json public
	middleware := JWTAuthWithConfig(AuthConfig{
		Mode: AuthModeStrict,
		PublicPathMatcher: func(path string) bool {
			return len(path) >= 5 && path[len(path)-5:] == ".json"
		},
	})(handler)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "path ending in .json",
			path:           "/api/data.json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "path not ending in .json",
			path:           "/api/data",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestGetUserAndIsAuthenticated tests helper functions
func TestGetUserAndIsAuthenticated(t *testing.T) {
	// Test with no user in context
	ctx := context.Background()
	if IsAuthenticated(ctx) {
		t.Error("Expected IsAuthenticated to return false for empty context")
	}
	if GetUser(ctx) != nil {
		t.Error("Expected GetUser to return nil for empty context")
	}

	// Test with user in context
	userClaims := map[string]interface{}{
		"sub":    "user123",
		"org_id": "org456",
	}
	ctx = context.WithValue(ctx, UserContextKey, userClaims)

	if !IsAuthenticated(ctx) {
		t.Error("Expected IsAuthenticated to return true for context with user")
	}

	user := GetUser(ctx)
	if user == nil {
		t.Fatal("Expected GetUser to return user claims")
	}

	if user["sub"] != "user123" {
		t.Errorf("Expected sub 'user123', got %v", user["sub"])
	}
}
