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

// User represents our domain object
type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

func main() {
	// Clean up any existing data from previous runs
	os.RemoveAll("./data")

	fmt.Println("=== toutago-datamapper Simple CRUD Example ===")
	fmt.Println()

	// Create mapper from configuration
	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	// Register filesystem adapter
	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
		return filesystem.NewFilesystemAdapter(source.Connection)
	})

	ctx := context.Background()

	// CREATE - Insert some users
	fmt.Println("1. Creating users...")
	users := []interface{}{
		User{ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Age: 30},
		User{ID: "2", Name: "Bob Smith", Email: "bob@example.com", Age: 25},
		User{ID: "3", Name: "Carol Williams", Email: "carol@example.com", Age: 35},
	}

	for _, user := range users {
		if err := mapper.Insert(ctx, "users.user-crud", user); err != nil {
			log.Printf("Failed to insert user: %v", err)
		} else {
			fmt.Printf("   ✓ Created user: %s\n", user.(User).Name)
		}
	}

	fmt.Println()

	// READ - Fetch a single user
	fmt.Println("2. Fetching user by ID...")
	var fetchedUser User
	if err := mapper.Fetch(ctx, "users.user-crud", map[string]interface{}{"id": "1"}, &fetchedUser); err != nil {
		log.Printf("Failed to fetch user: %v", err)
	} else {
		fmt.Printf("   ✓ Found user: %s (%s) - Age: %d\n",
			fetchedUser.Name, fetchedUser.Email, fetchedUser.Age)
	}

	fmt.Println()

	// READ - List all users
	fmt.Println("3. Listing all users...")
	var allUsers []map[string]interface{}
	if err := mapper.FetchMulti(ctx, "users.user-list", nil, &allUsers); err != nil {
		log.Printf("Failed to list users: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d users:\n", len(allUsers))
		for _, u := range allUsers {
			fmt.Printf("     - %s (%s)\n", u["name"], u["email"])
		}
	}

	fmt.Println()

	// UPDATE - Modify a user
	fmt.Println("4. Updating user...")
	updatedUser := User{
		ID:    "2",
		Name:  "Bob Smith Jr.",
		Email: "bob.smith@example.com",
		Age:   26,
	}

	if err := mapper.Update(ctx, "users.user-crud", updatedUser); err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		fmt.Printf("   ✓ Updated user: %s\n", updatedUser.Name)
	}

	// Verify update
	var verifyUser User
	if err := mapper.Fetch(ctx, "users.user-crud", map[string]interface{}{"id": "2"}, &verifyUser); err == nil {
		fmt.Printf("   ✓ Verified: %s (%s) - Age: %d\n",
			verifyUser.Name, verifyUser.Email, verifyUser.Age)
	}

	fmt.Println()

	// DELETE - Remove a user
	fmt.Println("5. Deleting user...")
	if err := mapper.Delete(ctx, "users.user-crud", "3"); err != nil {
		log.Printf("Failed to delete user: %v", err)
	} else {
		fmt.Println("   ✓ Deleted user ID: 3")
	}

	// Verify deletion
	var deletedUser User
	if err := mapper.Fetch(ctx, "users.user-crud", map[string]interface{}{"id": "3"}, &deletedUser); err != nil {
		fmt.Printf("   ✓ Verified deletion: %v\n", err)
	}

	fmt.Println()

	// Final count
	fmt.Println("6. Final user count...")
	var finalUsers []map[string]interface{}
	if err := mapper.FetchMulti(ctx, "users.user-list", nil, &finalUsers); err == nil {
		fmt.Printf("   ✓ Total users remaining: %d\n", len(finalUsers))
	}

	fmt.Println()
	fmt.Println("=== Example Complete ===")
	fmt.Println()
	fmt.Println("Check the ./data/users directory to see the JSON files created!")
}
