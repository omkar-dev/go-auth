package main

import (
	"fmt"
	"log"
	"os"

	goauth "github.com/oolio-group/loyalty-2/packages/go-auth"
)

// This example demonstrates three ways to configure the JWT secret

func main() {
	fmt.Println("=== JWT Secret Configuration Examples ===\n")

	// Method 1: Default environment variable (JWT_SECRET)
	example1()

	// Method 2: Custom environment variable name
	example2()

	// Method 3: Direct secret configuration
	example3()
}

func example1() {
	fmt.Println("1. Using default environment variable (JWT_SECRET):")

	os.Setenv("JWT_SECRET", "my-default-secret")
	defer os.Unsetenv("JWT_SECRET")

	tokens, err := goauth.GenerateTokenPair("user1", "+1111111111", "org1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   ✓ Generated tokens using JWT_SECRET\n")
	fmt.Printf("   Access Token: %s...\n\n", tokens.AccessToken[:50])
}

func example2() {
	fmt.Println("2. Using custom environment variable name:")

	// Override the environment variable name
	goauth.JWTSecretEnvVar = "MY_CUSTOM_JWT_SECRET"
	defer func() {
		goauth.JWTSecretEnvVar = "JWT_SECRET" // Reset
	}()

	os.Setenv("MY_CUSTOM_JWT_SECRET", "my-custom-secret")
	defer os.Unsetenv("MY_CUSTOM_JWT_SECRET")

	tokens, err := goauth.GenerateTokenPair("user2", "+2222222222", "org2")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   ✓ Generated tokens using MY_CUSTOM_JWT_SECRET\n")
	fmt.Printf("   Access Token: %s...\n\n", tokens.AccessToken[:50])
}

func example3() {
	fmt.Println("3. Using direct secret configuration (highest priority):")

	// Set the secret directly - this takes precedence over environment variables
	goauth.JWTSecret = "my-direct-secret"
	defer func() {
		goauth.JWTSecret = "" // Reset
	}()

	// Even if env var is set, JWTSecret takes precedence
	os.Setenv("JWT_SECRET", "this-will-be-ignored")
	defer os.Unsetenv("JWT_SECRET")

	tokens, err := goauth.GenerateTokenPair("user3", "+3333333333", "org3")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   ✓ Generated tokens using direct JWTSecret (env var ignored)\n")
	fmt.Printf("   Access Token: %s...\n\n", tokens.AccessToken[:50])

	// Verify the tokens work
	claims, err := goauth.ValidateToken(tokens.AccessToken, "access")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   ✓ Validated token - User: %s, Org: %s\n", claims.MemberID, claims.OrgID)
}
