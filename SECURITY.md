# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do Not Open a Public Issue

Security vulnerabilities should **not** be reported through public GitHub issues.

### 2. Contact Us Privately

Email your findings to: **security@oolio.io**

Include the following information:
- Type of vulnerability
- Full paths of source file(s) related to the issue
- Location of the affected source code (tag/branch/commit/URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: 24-48 hours
  - High: 7 days
  - Medium: 30 days
  - Low: 90 days

### 4. Disclosure Policy

- Security vulnerabilities will be disclosed after a fix is released
- Credit will be given to the reporter (if desired)
- We follow responsible disclosure practices

## Security Best Practices

### For Users of This Library

1. **Secret Management**
   ```bash
   # Use strong, random secrets (minimum 256 bits)
   openssl rand -base64 32

   # Never commit secrets to version control
   echo "JWT_SECRET=your-secret" >> .env
   echo ".env" >> .gitignore
   ```

2. **Environment Variables**
   - Store `JWT_SECRET` in secure environment variables
   - Use secret management services (AWS Secrets Manager, HashiCorp Vault, etc.)
   - Rotate secrets regularly

3. **HTTPS Only**
   ```go
   // Enforce HTTPS in production
   if os.Getenv("ENVIRONMENT") == "production" {
       handler = enforceHTTPS(handler)
   }
   ```

4. **Token Storage**
   - **DO NOT** store tokens in localStorage (XSS vulnerable)
   - **DO** use httpOnly cookies or memory storage
   - **DO** set Secure and SameSite flags on cookies

5. **CORS Configuration**
   ```go
   // Restrict origins in production
   cors.AllowedOrigins([]string{"https://yourdomain.com"})
   ```

6. **Rate Limiting**
   Implement rate limiting for authentication endpoints to prevent brute force attacks.

### Known Security Considerations

#### 1. HS256 Signing Algorithm

This library uses HS256 (HMAC with SHA-256) for JWT signing. While secure when properly implemented:

**Advantages:**
- Fast and efficient
- Single secret for signing and verification
- Suitable for symmetric key scenarios

**Limitations:**
- Shared secret between all parties
- Not ideal for distributed systems with multiple verifiers

**Recommendation:** For systems with multiple token verifiers, consider implementing RS256 (RSA) support.

#### 2. Token Revocation

Current implementation does not include token revocation. Consider implementing:
- Token blacklisting with Redis
- Short-lived tokens with frequent rotation
- Centralized token validation service

#### 3. Middleware Behavior

The current middleware is **advisory** (non-blocking):
- Invalid tokens result in `nil` user context
- Requests continue to handlers even without valid tokens
- **Action Required**: Implement authorization checks in your handlers

Example:
```go
func protectedHandler(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value(goauth.UserContextKey)
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Handle authenticated request
}
```

#### 4. Token Expiry

Default token durations:
- Access Token: 24 hours
- Refresh Token: 365 days

**Recommendation:** Adjust based on your security requirements. Shorter durations are more secure but less convenient.

### Secure Configuration Examples

#### Production Environment

```go
package main

import (
    "crypto/rand"
    "encoding/base64"
    "log"
    "os"
)

func init() {
    // Verify JWT secret is set and strong
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET must be set")
    }

    // Verify secret strength (minimum 32 bytes)
    decoded, err := base64.StdEncoding.DecodeString(secret)
    if err == nil && len(decoded) < 32 {
        log.Fatal("JWT_SECRET must be at least 32 bytes")
    }

    // Verify HTTPS in production
    if os.Getenv("ENVIRONMENT") == "production" {
        if os.Getenv("FORCE_HTTPS") != "true" {
            log.Fatal("FORCE_HTTPS must be enabled in production")
        }
    }
}

// GenerateSecret creates a cryptographically secure secret
func GenerateSecret() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(err)
    }
    return base64.StdEncoding.EncodeToString(b)
}
```

## Security Checklist

Before deploying to production:

- [ ] JWT secret is at least 256 bits (32 bytes)
- [ ] JWT secret is stored securely (not in code)
- [ ] HTTPS is enforced
- [ ] Token expiry times are appropriate for your use case
- [ ] Authorization checks are implemented in all protected handlers
- [ ] Rate limiting is configured for auth endpoints
- [ ] CORS is properly configured
- [ ] Tokens are stored securely on client side
- [ ] Logging does not include sensitive data
- [ ] Dependencies are up to date
- [ ] Security headers are set (CSP, X-Frame-Options, etc.)

## Security Updates

Subscribe to security updates:
- Watch this repository for security advisories
- Subscribe to the [security mailing list](mailto:security-subscribe@oolio.io)
- Check [GitHub Security Advisories](https://github.com/oolio-group/loyalty-2/security/advisories)

## Additional Resources

- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

Thank you for helping keep Go Auth and our users safe!
