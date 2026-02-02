# Contributing to Go Auth

Thank you for considering contributing to Go Auth! This document outlines the process and guidelines for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs actual behavior
- **Go version** and OS
- **Code samples** if applicable

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear use case** and motivation
- **Detailed description** of the proposed functionality
- **Examples** of how it would be used
- **Potential impact** on existing functionality

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** with clear, focused commits
3. **Add tests** for any new functionality
4. **Update documentation** including README if needed
5. **Ensure tests pass** with `go test -v`
6. **Follow code style** (use `gofmt` and `golint`)
7. **Submit pull request** with a clear description

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/loyalty-2.git
cd loyalty-2/packages/go-auth

# Install dependencies
go mod download

# Run tests
go test -v

# Run tests with coverage
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Code Style Guidelines

### Go Standards

- Follow standard Go conventions and idioms
- Use `gofmt` to format code
- Use `golint` and `go vet` for linting
- Write clear, self-documenting code

### Naming Conventions

- Use descriptive variable names
- Follow Go naming conventions (camelCase for private, PascalCase for public)
- Package name should be lowercase, single word

### Comments

- Add godoc comments for all exported functions, types, and constants
- Use complete sentences in comments
- Explain the "why" not just the "what"

Example:
```go
// GenerateTokenPair creates both an access token and refresh token for the given user.
// The access token is valid for 24 hours while the refresh token is valid for 365 days.
// Returns an error if the JWT secret is not configured.
func GenerateTokenPair(memberID, phone, orgID string) (*TokenPair, error) {
    // Implementation
}
```

### Error Handling

- Always handle errors explicitly
- Use structured logging with `slog`
- Return meaningful error messages
- Don't panic in library code

### Testing

- Write table-driven tests where appropriate
- Aim for >80% code coverage
- Test both success and failure cases
- Use meaningful test names

Example:
```go
func TestGenerateTokenPair(t *testing.T) {
    tests := []struct {
        name        string
        memberID    string
        phone       string
        orgID       string
        secretSet   bool
        expectError bool
    }{
        {
            name:        "valid input with secret set",
            memberID:    "user123",
            phone:       "+1234567890",
            orgID:       "org456",
            secretSet:   true,
            expectError: false,
        },
        {
            name:        "missing JWT secret",
            memberID:    "user123",
            phone:       "+1234567890",
            orgID:       "org456",
            secretSet:   false,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.secretSet {
                os.Setenv("JWT_SECRET", "test-secret")
            } else {
                os.Unsetenv("JWT_SECRET")
            }

            tokens, err := GenerateTokenPair(tt.memberID, tt.phone, tt.orgID)

            if tt.expectError && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.expectError && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            if !tt.expectError && tokens == nil {
                t.Error("Expected tokens but got nil")
            }
        })
    }
}
```

## Commit Message Guidelines

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(token): add support for custom token duration

Allow users to specify custom access and refresh token durations
through configuration instead of using hardcoded values.

Closes #123
```

```
fix(middleware): handle missing authorization header correctly

Previously, the middleware would panic if the Authorization header
was missing. Now it properly passes through to the next handler.

Fixes #456
```

## Security Issues

**DO NOT** open public issues for security vulnerabilities. Instead:

1. Email security concerns to: [security@oolio.io](mailto:security@oolio.io)
2. Include detailed description and steps to reproduce
3. Allow time for the issue to be fixed before public disclosure

## Questions?

Feel free to open a discussion on GitHub Discussions for any questions about contributing!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
