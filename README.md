# Go Auth

A lightweight, production-ready JWT authentication library for Go applications with built-in middleware support.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/oolio-group/loyalty-2/packages/go-auth)](https://goreportcard.com/report/github.com/oolio-group/loyalty-2/packages/go-auth)

## Features

- **JWT Token Generation**: Create access and refresh token pairs with customizable claims
- **Token Validation**: Validate and parse JWT tokens with type checking
- **HTTP Middleware**: Drop-in middleware for securing HTTP routes
- **Refresh Token Support**: Built-in refresh token flow
- **Context Integration**: Automatic user context injection into requests
- **Flexible Public Routes**: Configurable public endpoints that bypass authentication
- **Structured Logging**: Uses Go's `log/slog` for observability
- **Zero Dependencies**: Only requires `golang-jwt/jwt/v4`

## Installation

```bash
go get github.com/oolio-group/loyalty-2/packages/go-auth
```

## Quick Start

### 1. Set Environment Variable

```bash
export JWT_SECRET="your-secret-key-here"
```

### 2. Generate Token Pair

```go
package main

import (
    "fmt"
    goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

func main() {
    tokens, err := goauth.GenerateTokenPair("user123", "+1234567890", "org456")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Access Token: %s\n", tokens.AccessToken)
    fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)
}
```

### 3. Protect Your HTTP Routes

```go
package main

import (
    "net/http"
    goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/api/protected", func(w http.ResponseWriter, r *http.Request) {
        // Access user claims from context
        user := r.Context().Value(goauth.UserContextKey)
        w.Write([]byte("Protected endpoint"))
    })

    // Wrap with JWT middleware
    handler := goauth.JWTAuth(mux)

    http.ListenAndServe(":8080", handler)
}
```

## API Documentation

### Token Generation

#### `GenerateTokenPair(memberID, phone, orgID string) (*TokenPair, error)`

Generates an access token and refresh token pair.

**Parameters:**
- `memberID` - Unique user identifier
- `phone` - User phone number
- `orgID` - Organization identifier

**Returns:**
- `TokenPair` with access token, refresh token, and expiry times
- `error` if token generation fails

**Token Durations:**
- Access Token: 24 hours
- Refresh Token: 365 days

**Example:**
```go
tokens, err := goauth.GenerateTokenPair("user123", "+1234567890", "org456")
if err != nil {
    log.Fatal(err)
}
```

### Token Validation

#### `ValidateToken(tokenString string, expectedType string) (*TokenClaims, error)`

Validates a JWT token and returns the claims.

**Parameters:**
- `tokenString` - JWT token to validate
- `expectedType` - Expected token type ("access" or "refresh")

**Returns:**
- `TokenClaims` containing user information
- `error` if validation fails

**Example:**
```go
claims, err := goauth.ValidateToken(token, "access")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s, Org: %s\n", claims.MemberID, claims.OrgID)
```

### Token Refresh

#### `RefreshAccessToken(refreshToken string) (*TokenPair, error)`

Generates a new token pair using a valid refresh token.

**Parameters:**
- `refreshToken` - Valid refresh token

**Returns:**
- New `TokenPair` with fresh access and refresh tokens
- `error` if refresh fails

**Example:**
```go
newTokens, err := goauth.RefreshAccessToken(oldRefreshToken)
if err != nil {
    log.Fatal(err)
}
```

### Middleware

#### `JWTAuth(next http.Handler) http.Handler`

HTTP middleware that validates JWT tokens and injects user claims into request context.

**Behavior:**
- Extracts token from `Authorization` header (supports both `Bearer <token>` and raw token)
- Skips authentication for public endpoints
- Injects user claims into context under `UserContextKey`
- Passes through to next handler (authentication is advisory, not blocking)

**Example:**
```go
protectedHandler := goauth.JWTAuth(yourHandler)
```

### Data Structures

#### `TokenPair`
```go
type TokenPair struct {
    AccessToken      string `json:"accessToken"`
    RefreshToken     string `json:"refreshToken"`
    ExpiresIn        int64  `json:"expiresIn"`        // seconds
    RefreshExpiresIn int64  `json:"refreshExpiresIn"` // seconds
}
```

#### `TokenClaims`
```go
type TokenClaims struct {
    MemberID string `json:"memberId"`
    Phone    string `json:"phone"`
    OrgID    string `json:"orgId"`
    Type     string `json:"type"` // "access" or "refresh"
    jwt.RegisteredClaims
}
```

## Configuration

### JWT Secret Configuration

The library provides three flexible ways to configure the JWT secret:

#### Method 1: Default Environment Variable (Recommended)

```bash
export JWT_SECRET="your-secret-key"
```

#### Method 2: Custom Environment Variable Name

If you're integrating this library and want to use your own environment variable name:

```go
import goauth "github.com/oolio-group/loyalty-2/packages/go-auth"

func init() {
    // Override the default environment variable name
    goauth.JWTSecretEnvVar = "MY_APP_JWT_SECRET"
}
```

Then set your custom environment variable:
```bash
export MY_APP_JWT_SECRET="your-secret-key"
```

#### Method 3: Direct Secret Configuration (Highest Priority)

Set the secret directly in code (useful for testing or when loading from secure vaults):

```go
import goauth "github.com/oolio-group/loyalty-2/packages/go-auth"

func init() {
    // This takes precedence over any environment variable
    goauth.JWTSecret = loadSecretFromVault()
}
```

**Priority Order:**
1. `goauth.JWTSecret` (if set) - Highest priority
2. Environment variable specified by `goauth.JWTSecretEnvVar`
3. Default `JWT_SECRET` environment variable

### Environment Variables

| Variable             | Required | Description                        | Example              | Can Override            |
|----------------------|----------|------------------------------------|----------------------|-------------------------|
| `JWT_SECRET`         | Yes*     | Secret key for signing JWT tokens  | `your-256-bit-secret`| Yes, via `JWTSecretEnvVar` |

*Required unless you set `goauth.JWTSecret` directly or configure a custom variable name

### Public Endpoints

By default, these endpoints bypass authentication:
- `/otp/generate`
- `/otp/validate`
- `/otp/refresh`
- `/health`
- `/metrics`

To customize public endpoints, modify the `isPublicEndpoint` function in [middleware.go:51](middleware.go#L51).

## Complete Example

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"

    goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

func main() {
    // Configure JWT secret (choose one method):

    // Method 1: Default environment variable
    os.Setenv("JWT_SECRET", "your-secret-key")

    // Method 2: Custom environment variable name
    // goauth.JWTSecretEnvVar = "MY_JWT_SECRET"
    // os.Setenv("MY_JWT_SECRET", "your-secret-key")

    // Method 3: Direct configuration
    // goauth.JWTSecret = "your-secret-key"

    mux := http.NewServeMux()

    // Public endpoint - generate tokens
    mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
        tokens, err := goauth.GenerateTokenPair("user123", "+1234567890", "org456")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(tokens)
    })

    // Public endpoint - refresh tokens
    mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
        refreshToken := r.Header.Get("X-Refresh-Token")
        tokens, err := goauth.RefreshAccessToken(refreshToken)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }
        json.NewEncoder(w).Encode(tokens)
    })

    // Protected endpoint
    mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
        user := r.Context().Value(goauth.UserContextKey)
        if user == nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        userClaims := user.(map[string]interface{})
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Welcome to your profile",
            "user":    userClaims,
        })
    })

    // Apply JWT middleware
    handler := goauth.JWTAuth(mux)

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

Test the server:

```bash
# Login and get tokens
curl -X POST http://localhost:8080/auth/login

# Access protected endpoint
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/profile

# Refresh tokens
curl -H "X-Refresh-Token: <refresh_token>" http://localhost:8080/auth/refresh
```

## Security Considerations

### Best Practices

1. **Secret Management**
   - Never commit `JWT_SECRET` to version control
   - Use a cryptographically secure random string (minimum 256 bits)
   - Rotate secrets regularly in production

2. **Token Storage**
   - Store access tokens in memory or httpOnly cookies
   - Never store tokens in localStorage for web applications
   - Implement secure token transmission (HTTPS only)

3. **Token Expiry**
   - Default access token: 24 hours
   - Default refresh token: 365 days
   - Adjust durations based on your security requirements in [token.go:28-29](token.go#L28-L29)

4. **HTTPS Required**
   - Always use HTTPS in production
   - Tokens transmitted over HTTP are vulnerable to interception

5. **Middleware Behavior**
   - Current middleware is advisory (non-blocking)
   - Invalid tokens result in `nil` user context, not HTTP errors
   - Implement authorization checks in your handlers

### Hardening Recommendations

For production use, consider:
- Implementing token revocation/blacklisting
- Adding rate limiting for token generation
- Enabling CORS with strict origin policies
- Implementing token rotation on refresh
- Adding audit logging for token operations
- Using RS256 instead of HS256 for better key management

## Testing

```go
package main

import (
    "os"
    "testing"

    goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

func TestTokenGeneration(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")

    tokens, err := goauth.GenerateTokenPair("user123", "+1234567890", "org456")
    if err != nil {
        t.Fatalf("Failed to generate tokens: %v", err)
    }

    if tokens.AccessToken == "" {
        t.Error("Access token is empty")
    }

    if tokens.RefreshToken == "" {
        t.Error("Refresh token is empty")
    }
}

func TestTokenValidation(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")

    tokens, _ := goauth.GenerateTokenPair("user123", "+1234567890", "org456")

    claims, err := goauth.ValidateToken(tokens.AccessToken, "access")
    if err != nil {
        t.Fatalf("Token validation failed: %v", err)
    }

    if claims.MemberID != "user123" {
        t.Errorf("Expected MemberID 'user123', got '%s'", claims.MemberID)
    }
}
```

Run tests:
```bash
go test -v
```

## Migration Guide

### From Other JWT Libraries

If you're migrating from another JWT library:

1. **Replace token generation:**
   ```go
   // Before (example with different library)
   token := jwt.New(jwt.SigningMethodHS256)
   tokenString, _ := token.SignedString([]byte(secret))

   // After
   tokens, err := goauth.GenerateTokenPair(memberID, phone, orgID)
   ```

2. **Update middleware:**
   ```go
   // Before
   handler := jwtmiddleware.New(...)

   // After
   handler := goauth.JWTAuth(yourHandler)
   ```

3. **Access claims:**
   ```go
   // Before
   user := r.Context().Value("user")

   // After
   user := r.Context().Value(goauth.UserContextKey)
   ```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
git clone https://github.com/oolio-group/loyalty-2.git
cd loyalty-2/packages/go-auth
go mod download
go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/oolio-group/loyalty-2/issues)
- **Discussions**: [GitHub Discussions](https://github.com/oolio-group/loyalty-2/discussions)

## Roadmap

- [ ] Add support for RS256/ES256 signing algorithms
- [ ] Token revocation support with Redis backend
- [ ] Multi-organization claim support
- [ ] Custom claim injection
- [ ] Token rotation strategies
- [ ] Prometheus metrics integration

## Acknowledgments

Built with [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - the most popular JWT implementation for Go.

---

Made with ❤️ for the Go community
