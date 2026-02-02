package goauth

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// JWTSecretEnvVar is the environment variable name for the JWT secret.
	// Can be overridden by libraries using this package.
	// Defaults to "JWT_SECRET".
	//
	// IMPORTANT: This is a global variable. Set it once at application startup
	// before any concurrent operations. Modifying it during runtime in a
	// concurrent application can lead to race conditions and unpredictable behavior.
	// Avoid modifying this in tests running with t.Parallel().
	JWTSecretEnvVar = "JWT_SECRET"

	// JWTSecret is the actual secret key used for signing tokens.
	// If set directly, this takes precedence over the environment variable.
	//
	// IMPORTANT: This is a global variable. Set it once at application startup
	// before any concurrent operations. Modifying it during runtime in a
	// concurrent application can lead to race conditions and unpredictable behavior.
	// Avoid modifying this in tests running with t.Parallel().
	JWTSecret string
)

type TokenPair struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
}

type TokenClaims struct {
	MemberID string `json:"memberId"`
	Phone    string `json:"phone"`
	OrgID    string `json:"orgId"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

const (
	AccessTokenDuration  = 24 * time.Hour
	RefreshTokenDuration = 365 * 24 * time.Hour
)

func GenerateTokenPair(memberID, phone, orgID string) (*TokenPair, error) {
	secret := getJWTSecret()
	if secret == "" {
		slog.Error("JWT secret not configured", slog.String("env_var", JWTSecretEnvVar))
		return nil, errors.New("JWT secret not configured")
	}

	now := time.Now()
	accessExpiry := now.Add(AccessTokenDuration)
	refreshExpiry := now.Add(RefreshTokenDuration)

	accessClaims := TokenClaims{
		MemberID: memberID,
		Phone:    phone,
		OrgID:    orgID,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "loyalty-service",
			Subject:   memberID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		slog.Error("Failed to sign access token", slog.Any("error", err))
		return nil, err
	}

	refreshClaims := TokenClaims{
		MemberID: memberID,
		Phone:    phone,
		OrgID:    orgID,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "loyalty-service",
			Subject:   memberID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		slog.Error("Failed to sign refresh token", slog.Any("error", err))
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		ExpiresIn:        int64(AccessTokenDuration.Seconds()),
		RefreshExpiresIn: int64(RefreshTokenDuration.Seconds()),
	}, nil
}

func ValidateToken(tokenString string, expectedType string) (*TokenClaims, error) {
	secret := getJWTSecret()
	if secret == "" {
		return nil, errors.New("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		slog.Error("Failed to parse token", slog.Any("error", err))
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if expectedType != "" && claims.Type != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

func RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := ValidateToken(refreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	return GenerateTokenPair(claims.MemberID, claims.Phone, claims.OrgID)
}

// getJWTSecret returns the JWT secret from JWTSecret variable or environment.
// If JWTSecret is set directly, it takes precedence over the environment variable.
func getJWTSecret() string {
	if JWTSecret != "" {
		return JWTSecret
	}
	return os.Getenv(JWTSecretEnvVar)
}
