package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
	"github.com/toutaio/toutago-datamapper/engine"
	"github.com/toutaio/toutago-datamapper/filesystem"
)

// Account represents a user account
type Account struct {
	ID       string
	Username string
	Email    string
	Role     string
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Credentials Management Example ===")
	fmt.Println()

	// Show environment setup
	fmt.Println("Environment Setup:")
	fmt.Println("  • Set environment variables before running")
	fmt.Println("  • Create .env file for local development")
	fmt.Println("  • Use credentials.yaml for secrets (DO NOT COMMIT!)")
	fmt.Println()

	// Check for required environment variables
	checkEnvironment()

	// Create mapper - credentials are resolved from environment
	fmt.Println("1. Creating mapper with credential resolution...")
	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()
	fmt.Println("   ✓ Mapper created successfully")
	fmt.Println("   ✓ Credentials resolved from environment")
	fmt.Println()

	// Register adapter
	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
		return filesystem.NewFilesystemAdapter(source.Connection)
	})

	// 2. Use the mapper - credentials are hidden
	fmt.Println("2. Performing operations with secure credentials...")

	account := Account{
		ID:       "acc-001",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "administrator",
	}

	if err := mapper.Insert(ctx, "accounts.account-crud", account); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Created account: %s\n", account.Username)
	}
	fmt.Println()

	// 3. Fetch the account
	fmt.Println("3. Fetching account...")
	var fetched Account
	err = mapper.Fetch(ctx, "accounts.account-crud", map[string]interface{}{"id": "acc-001"}, &fetched)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Retrieved: %s (%s)\n", fetched.Username, fetched.Email)
	}
	fmt.Println()

	fmt.Println("=== Security Best Practices Demonstrated ===")
	fmt.Println()
	fmt.Println("✅ Configuration Separation:")
	fmt.Println("   • config.yaml - Committed to git")
	fmt.Println("   • .env - Local development only")
	fmt.Println("   • credentials.yaml - NEVER commit!")
	fmt.Println()
	fmt.Println("✅ Environment Variables:")
	fmt.Println("   • Production uses real env vars")
	fmt.Println("   • Local uses .env file")
	fmt.Println("   • CI/CD uses secrets manager")
	fmt.Println()
	fmt.Println("✅ Multiple Environments:")
	fmt.Println("   • Development: .env.development")
	fmt.Println("   • Staging: .env.staging")
	fmt.Println("   • Production: System environment")
	fmt.Println()
}

func checkEnvironment() {
	required := []string{
		"DB_PATH",
		"DB_FORMAT",
	}

	optional := []string{
		"DB_USER",
		"DB_PASSWORD",
		"DB_HOST",
		"DB_PORT",
	}

	fmt.Println("Checking environment variables...")

	allSet := true
	for _, envVar := range required {
		value := os.Getenv(envVar)
		if value == "" {
			fmt.Printf("   ✗ %s: NOT SET (required)\n", envVar)
			allSet = false
		} else {
			fmt.Printf("   ✓ %s: SET\n", envVar)
		}
	}

	for _, envVar := range optional {
		value := os.Getenv(envVar)
		if value == "" {
			fmt.Printf("   - %s: not set (optional)\n", envVar)
		} else {
			fmt.Printf("   ✓ %s: SET\n", envVar)
		}
	}

	fmt.Println()

	if !allSet {
		fmt.Println("⚠️  Some required variables are missing.")
		fmt.Println()
		fmt.Println("To run this example:")
		fmt.Println("  export DB_PATH=\"./data\"")
		fmt.Println("  export DB_FORMAT=\"json\"")
		fmt.Println()
		fmt.Println("Or use the provided .env file:")
		fmt.Println("  source .env")
		fmt.Println()

		// Set defaults for demo
		if os.Getenv("DB_PATH") == "" {
			os.Setenv("DB_PATH", "./data")
		}
		if os.Getenv("DB_FORMAT") == "" {
			os.Setenv("DB_FORMAT", "json")
		}

		fmt.Println("Using default values for demonstration...")
		fmt.Println()
	}
}
