# Go Auth Examples

This directory contains example implementations demonstrating how to use the go-auth library.

## Basic Server Example

A complete HTTP server with authentication endpoints.

### Running the Example

```bash
cd examples
go run basic_server.go
```

The server will start on `http://localhost:8080`

### Testing the Example

#### 1. Login to get tokens

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "memberId": "user123",
    "phone": "+1234567890",
    "orgId": "org456"
  }'
```

Response:
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "expiresIn": 86400,
  "refreshExpiresIn": 31536000
}
```

#### 2. Access protected endpoint

```bash
# Save the access token from the previous response
ACCESS_TOKEN="your-access-token-here"

curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Response:
```json
{
  "message": "Welcome to your profile",
  "user": {
    "sub": "user123",
    "phone": "+1234567890",
    "org_id": "org456",
    "org_ids": ["org456"]
  }
}
```

#### 3. Refresh the access token

```bash
# Save the refresh token from the login response
REFRESH_TOKEN="your-refresh-token-here"

curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"
```

Response:
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "expiresIn": 86400,
  "refreshExpiresIn": 31536000
}
```

#### 4. Test unauthorized access

```bash
curl http://localhost:8080/api/profile
```

Response:
```
Unauthorized - Please provide a valid access token
```

### Complete Test Script

Save this as `test.sh`:

```bash
#!/bin/bash

echo "1. Logging in..."
RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"memberId":"user123","phone":"+1234567890","orgId":"org456"}')

echo "$RESPONSE" | jq .

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r .accessToken)
REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r .refreshToken)

echo -e "\n2. Accessing protected endpoint..."
curl -s http://localhost:8080/api/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

echo -e "\n3. Refreshing token..."
curl -s -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}" | jq .

echo -e "\n4. Testing unauthorized access..."
curl -s http://localhost:8080/api/profile

echo -e "\n\nDone!"
```

Make it executable and run:
```bash
chmod +x test.sh
./test.sh
```

## Features Demonstrated

- Token generation (login)
- Token validation (protected endpoints)
- Token refresh
- HTTP middleware integration
- Context-based user claims
- Public vs protected routes
- CORS handling
- Error handling

## Production Considerations

This example uses simplified authentication for demonstration. In production:

1. **Validate credentials** - Add actual user validation in login handler
2. **Use HTTPS** - Never send tokens over HTTP
3. **Secure secrets** - Store JWT secret in secure environment variables
4. **Rate limiting** - Add rate limiting to prevent brute force
5. **Database integration** - Store and validate user credentials
6. **Token revocation** - Implement token blacklisting if needed
7. **Logging** - Add comprehensive logging for security events
8. **Input validation** - Validate all user inputs
9. **CORS** - Restrict CORS to specific origins in production
10. **Error messages** - Don't leak sensitive information in errors
