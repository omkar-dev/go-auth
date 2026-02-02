package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

func main() {
	// Set JWT secret (in production, use environment variables)
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "demo-secret-key-change-in-production")
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/auth/refresh", refreshHandler)

	// Protected endpoints
	mux.HandleFunc("/api/profile", profileHandler)
	mux.HandleFunc("/api/protected", protectedHandler)

	// Apply JWT middleware with public paths configuration
	// Using strict mode (default) - returns 401 for missing/invalid tokens
	handler := goauth.JWTAuthWithConfig(goauth.AuthConfig{
		Mode: goauth.AuthModeStrict,
		PublicPaths: []string{
			"/health",
			"/auth/login",
			"/auth/refresh",
		},
	})(mux)

	// Add CORS middleware (optional)
	handler = corsMiddleware(handler)

	log.Println("Server starting on http://localhost:8080")
	log.Println("\nAvailable endpoints:")
	log.Println("  POST /auth/login       - Get access and refresh tokens")
	log.Println("  POST /auth/refresh     - Refresh access token")
	log.Println("  GET  /api/profile      - Get user profile (requires auth)")
	log.Println("  GET  /api/protected    - Protected endpoint (requires auth)")
	log.Println("  GET  /health           - Health check")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// loginHandler generates tokens for a user
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In production, validate credentials here
	var req struct {
		MemberID string `json:"memberId"`
		Phone    string `json:"phone"`
		OrgID    string `json:"orgId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults for demo
	if req.MemberID == "" {
		req.MemberID = "demo-user-123"
	}
	if req.Phone == "" {
		req.Phone = "+1234567890"
	}
	if req.OrgID == "" {
		req.OrgID = "demo-org-456"
	}

	// Generate token pair
	tokens, err := goauth.GenerateTokenPair(req.MemberID, req.Phone, req.OrgID)
	if err != nil {
		log.Printf("Failed to generate tokens: %v", err)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// refreshHandler refreshes access token using refresh token
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Refresh the token
	tokens, err := goauth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// profileHandler returns user profile (protected)
func profileHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(goauth.UserContextKey)
	if user == nil {
		http.Error(w, "Unauthorized - Please provide a valid access token", http.StatusUnauthorized)
		return
	}

	userClaims := user.(map[string]interface{})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Welcome to your profile",
		"user":    userClaims,
	})
}

// protectedHandler demonstrates a protected endpoint
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(goauth.UserContextKey)
	if user == nil {
		http.Error(w, "Unauthorized - Please provide a valid access token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "This is a protected endpoint",
		"data":    "Secret data only for authenticated users",
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
