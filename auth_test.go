package goauth

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TestGenerateTokenPair tests token pair generation
func TestGenerateTokenPair(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, err := GenerateTokenPair("user123", "+1234567890", "org456")
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if tokens.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if tokens.ExpiresIn != int64(AccessTokenDuration.Seconds()) {
		t.Errorf("Expected ExpiresIn %d, got %d", int64(AccessTokenDuration.Seconds()), tokens.ExpiresIn)
	}

	if tokens.RefreshExpiresIn != int64(RefreshTokenDuration.Seconds()) {
		t.Errorf("Expected RefreshExpiresIn %d, got %d", int64(RefreshTokenDuration.Seconds()), tokens.RefreshExpiresIn)
	}
}

// TestGenerateTokenPairWithoutSecret tests error handling when secret is missing
func TestGenerateTokenPairWithoutSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	_, err := GenerateTokenPair("user123", "+1234567890", "org456")
	if err == nil {
		t.Error("Expected error when JWT secret is not set, got nil")
	}
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, err := GenerateTokenPair("user123", "+1234567890", "org456")
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Validate access token
	claims, err := ValidateToken(tokens.AccessToken, "access")
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.MemberID != "user123" {
		t.Errorf("Expected MemberID 'user123', got '%s'", claims.MemberID)
	}

	if claims.Phone != "+1234567890" {
		t.Errorf("Expected Phone '+1234567890', got '%s'", claims.Phone)
	}

	if claims.OrgID != "org456" {
		t.Errorf("Expected OrgID 'org456', got '%s'", claims.OrgID)
	}

	if claims.Type != "access" {
		t.Errorf("Expected Type 'access', got '%s'", claims.Type)
	}

	// Validate refresh token
	refreshClaims, err := ValidateToken(tokens.RefreshToken, "refresh")
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.Type != "refresh" {
		t.Errorf("Expected Type 'refresh', got '%s'", refreshClaims.Type)
	}
}

// TestValidateTokenWithWrongType tests type validation
func TestValidateTokenWithWrongType(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, _ := GenerateTokenPair("user123", "+1234567890", "org456")

	// Try to validate access token as refresh token
	_, err := ValidateToken(tokens.AccessToken, "refresh")
	if err == nil {
		t.Error("Expected error when validating access token as refresh token, got nil")
	}

	// Try to validate refresh token as access token
	_, err = ValidateToken(tokens.RefreshToken, "access")
	if err == nil {
		t.Error("Expected error when validating refresh token as access token, got nil")
	}
}

// TestValidateInvalidToken tests validation of invalid tokens
func TestValidateInvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "invalid.token.format"},
		{"malformed token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, "access")
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

// TestValidateTokenWithWrongSecret tests validation with wrong secret
func TestValidateTokenWithWrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	tokens, _ := GenerateTokenPair("user123", "+1234567890", "org456")

	// Change the secret
	os.Setenv("JWT_SECRET", "different-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateToken(tokens.AccessToken, "access")
	if err == nil {
		t.Error("Expected error when validating token with wrong secret, got nil")
	}
}

// TestRefreshAccessToken tests token refresh functionality
func TestRefreshAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// Generate initial tokens
	tokens, err := GenerateTokenPair("user123", "+1234567890", "org456")
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Wait a second to ensure new tokens have different timestamps
	time.Sleep(1 * time.Second)

	// Refresh tokens
	newTokens, err := RefreshAccessToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh tokens: %v", err)
	}

	if newTokens.AccessToken == "" {
		t.Error("New access token is empty")
	}

	if newTokens.RefreshToken == "" {
		t.Error("New refresh token is empty")
	}

	// Verify new tokens are different
	if newTokens.AccessToken == tokens.AccessToken {
		t.Error("New access token should be different from old token")
	}

	// Verify new tokens contain same user info
	claims, err := ValidateToken(newTokens.AccessToken, "access")
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if claims.MemberID != "user123" {
		t.Errorf("Expected MemberID 'user123', got '%s'", claims.MemberID)
	}
}

// TestRefreshWithInvalidToken tests refresh with invalid refresh token
func TestRefreshWithInvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	_, err := RefreshAccessToken("invalid-token")
	if err == nil {
		t.Error("Expected error when refreshing with invalid token, got nil")
	}
}

// TestRefreshWithAccessToken tests that access token cannot be used to refresh
func TestRefreshWithAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, _ := GenerateTokenPair("user123", "+1234567890", "org456")

	// Try to refresh using access token instead of refresh token
	_, err := RefreshAccessToken(tokens.AccessToken)
	if err == nil {
		t.Error("Expected error when refreshing with access token, got nil")
	}
}

// TestTokenExpiry tests that expired tokens are rejected
func TestTokenExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping expiry test in short mode")
	}

	secret := "test-secret-key"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Create an expired token manually
	claims := TokenClaims{
		MemberID: "user123",
		Phone:    "+1234567890",
		OrgID:    "org456",
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "loyalty-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Use the same secret that ValidateToken will use
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = ValidateToken(tokenString, "access")
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}

	// Verify it's actually an expiry error, not a configuration error
	if !strings.Contains(err.Error(), "expired") && !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("Expected expiry error, got: %v", err)
	}
}

// TestTokenClaims tests all claim fields
func TestTokenClaims(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, _ := GenerateTokenPair("user123", "+1234567890", "org456")
	claims, _ := ValidateToken(tokens.AccessToken, "access")

	// Test all custom claims
	if claims.MemberID != "user123" {
		t.Errorf("MemberID: expected 'user123', got '%s'", claims.MemberID)
	}

	if claims.Phone != "+1234567890" {
		t.Errorf("Phone: expected '+1234567890', got '%s'", claims.Phone)
	}

	if claims.OrgID != "org456" {
		t.Errorf("OrgID: expected 'org456', got '%s'", claims.OrgID)
	}

	if claims.Type != "access" {
		t.Errorf("Type: expected 'access', got '%s'", claims.Type)
	}

	// Test standard claims
	if claims.Issuer != "loyalty-service" {
		t.Errorf("Issuer: expected 'loyalty-service', got '%s'", claims.Issuer)
	}

	if claims.Subject != "user123" {
		t.Errorf("Subject: expected 'user123', got '%s'", claims.Subject)
	}

	// Test expiry is in the future
	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}

	// Test issued at is in the past
	if claims.IssuedAt.After(time.Now()) {
		t.Error("IssuedAt should be in the past")
	}
}

// BenchmarkGenerateTokenPair benchmarks token generation
func BenchmarkGenerateTokenPair(b *testing.B) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateTokenPair("user123", "+1234567890", "org456")
	}
}

// BenchmarkValidateToken benchmarks token validation
func BenchmarkValidateToken(b *testing.B) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	tokens, _ := GenerateTokenPair("user123", "+1234567890", "org456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(tokens.AccessToken, "access")
	}
}
