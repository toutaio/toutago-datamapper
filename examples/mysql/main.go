package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mysql "github.com/toutaio/toutago-datamapper-mysql"
	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
	"github.com/toutaio/toutago-datamapper/engine"
)

// User represents a domain object with zero database dependencies
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

// Product represents another domain object
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
}

func main() {
	fmt.Println("=== MySQL Adapter Example ===")
	fmt.Println()

	// Create mapper with configuration
	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	// Register MySQL adapter
	mapper.RegisterAdapter("mysql", func(source config.Source) (adapter.Adapter, error) {
		return mysql.NewMySQLAdapter(), nil
	})

	ctx := context.Background()

	// Run examples
	if err := basicCRUD(ctx, mapper); err != nil {
		log.Printf("Basic CRUD error: %v", err)
	}

	if err := bulkOperations(ctx, mapper); err != nil {
		log.Printf("Bulk operations error: %v", err)
	}

	if err := optimisticLocking(ctx, mapper); err != nil {
		log.Printf("Optimistic locking error: %v", err)
	}

	if err := customActions(ctx, mapper); err != nil {
		log.Printf("Custom actions error: %v", err)
	}

	fmt.Println()
	fmt.Println("=== Example Complete ===")
}

func basicCRUD(ctx context.Context, mapper *engine.Mapper) error {
	fmt.Println("--- Basic CRUD Operations ---")

	// Create a new user
	newUser := map[string]interface{}{
		"Name":  "Alice Johnson",
		"Email": "alice@example.com",
	}

	fmt.Printf("Creating user: %s (%s)\n", newUser["Name"], newUser["Email"])
	if err := mapper.Insert(ctx, "users.insert", newUser); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✓ User created with ID: %v\n", newUser["ID"])

	// Fetch the user
	var fetchedUser map[string]interface{}
	fmt.Printf("\nFetching user with ID: %v\n", newUser["ID"])
	if err := mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
		"id": newUser["ID"],
	}, &fetchedUser); err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}
	fmt.Printf("✓ User fetched: %s (%s)\n", fetchedUser["Name"], fetchedUser["Email"])

	// Update the user
	fetchedUser["Email"] = "alice.johnson@newdomain.com"
	fmt.Printf("\nUpdating user email to: %s\n", fetchedUser["Email"])
	if err := mapper.Update(ctx, "users.update", fetchedUser); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	fmt.Println("✓ User updated")

	// Verify update
	var updatedUser map[string]interface{}
	if err := mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
		"id": newUser["ID"],
	}, &updatedUser); err != nil {
		return fmt.Errorf("fetch after update failed: %w", err)
	}
	fmt.Printf("✓ Verified update: %s\n", updatedUser["Email"])

	// Delete the user
	fmt.Printf("\nDeleting user with ID: %v\n", newUser["ID"])
	if err := mapper.Delete(ctx, "users.delete", newUser["ID"]); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("✓ User deleted")

	// Verify deletion
	var deletedUser map[string]interface{}
	err := mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
		"id": newUser["ID"],
	}, &deletedUser)
	if err == adapter.ErrNotFound {
		fmt.Println("✓ Verified deletion (user not found)")
	} else if err != nil {
		return fmt.Errorf("unexpected error after delete: %w", err)
	} else {
		fmt.Println("⚠ Warning: User still exists after delete")
	}

	fmt.Println()
	return nil
}

func bulkOperations(ctx context.Context, mapper *engine.Mapper) error {
	fmt.Println("--- Bulk Operations ---")

	// Create multiple products
	products := []interface{}{
		map[string]interface{}{
			"Name":        "Laptop Pro",
			"Description": "High-performance laptop",
			"Price":       1299.99,
			"Stock":       50,
		},
		map[string]interface{}{
			"Name":        "Wireless Mouse",
			"Description": "Ergonomic wireless mouse",
			"Price":       29.99,
			"Stock":       200,
		},
		map[string]interface{}{
			"Name":        "USB-C Hub",
			"Description": "7-in-1 USB-C hub",
			"Price":       49.99,
			"Stock":       150,
		},
		map[string]interface{}{
			"Name":        "Mechanical Keyboard",
			"Description": "RGB mechanical keyboard",
			"Price":       149.99,
			"Stock":       75,
		},
	}

	fmt.Printf("Bulk inserting %d products...\n", len(products))
	start := time.Now()
	if err := mapper.Insert(ctx, "products.bulk-insert", products); err != nil {
		return fmt.Errorf("bulk insert failed: %w", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("✓ %d products inserted in %v (avg: %v per product)\n",
		len(products), elapsed, elapsed/time.Duration(len(products)))

	// Fetch all products
	var allProducts []interface{}
	fmt.Println("\nFetching all products...")
	if err := mapper.FetchMulti(ctx, "products.fetch-all", nil, &allProducts); err != nil {
		return fmt.Errorf("fetch all failed: %w", err)
	}
	fmt.Printf("✓ Fetched %d products\n", len(allProducts))

	for i, p := range allProducts {
		prod := p.(map[string]interface{})
		fmt.Printf("  %d. %s - $%.2f (Stock: %v)\n",
			i+1, prod["Name"], prod["Price"], prod["Stock"])
	}

	// Update stock for all products (simulate sales)
	fmt.Println("\nUpdating product stock...")
	for _, p := range allProducts {
		prod := p.(map[string]interface{})
		currentStock := int(prod["Stock"].(int64))
		prod["Stock"] = currentStock - 10 // Simulate selling 10 units
	}

	if err := mapper.Update(ctx, "products.update", allProducts); err != nil {
		return fmt.Errorf("bulk update failed: %w", err)
	}
	fmt.Println("✓ All product stocks updated")

	// Clean up - delete all test products
	fmt.Println("\nCleaning up test products...")
	for _, p := range allProducts {
		prod := p.(map[string]interface{})
		if err := mapper.Delete(ctx, "products.delete", prod["ID"]); err != nil {
			log.Printf("Warning: failed to delete product %v: %v", prod["ID"], err)
		}
	}
	fmt.Println("✓ Test products cleaned up")

	fmt.Println()
	return nil
}

func optimisticLocking(ctx context.Context, mapper *engine.Mapper) error {
	fmt.Println("--- Optimistic Locking ---")

	// Create a user with version field
	user := map[string]interface{}{
		"Name":    "Bob Smith",
		"Email":   "bob@example.com",
		"Version": 1,
	}

	fmt.Println("Creating user with version control...")
	if err := mapper.Insert(ctx, "users.insert", user); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✓ User created with ID: %v, Version: %v\n", user["ID"], user["Version"])

	// Fetch the user
	var fetchedUser map[string]interface{}
	if err := mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
		"id": user["ID"],
	}, &fetchedUser); err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	// Simulate concurrent update scenario
	fmt.Println("\nSimulating concurrent updates...")

	// First update (should succeed)
	user1 := make(map[string]interface{})
	for k, v := range fetchedUser {
		user1[k] = v
	}
	user1["Email"] = "bob.smith.v2@example.com"
	fmt.Printf("Update 1: Changing email with Version %v\n", user1["Version"])
	if err := mapper.Update(ctx, "users.update-versioned", user1); err != nil {
		return fmt.Errorf("first update failed: %w", err)
	}
	fmt.Println("✓ Update 1 succeeded")

	// Second update with old version (should fail)
	user2 := make(map[string]interface{})
	for k, v := range fetchedUser {
		user2[k] = v
	}
	user2["Email"] = "bob.smith.conflict@example.com"
	fmt.Printf("Update 2: Trying to change email with stale Version %v\n", user2["Version"])
	err := mapper.Update(ctx, "users.update-versioned", user2)
	if err == adapter.ErrNotFound {
		fmt.Println("✓ Update 2 failed as expected (version mismatch)")
	} else if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	} else {
		fmt.Println("⚠ Warning: Update 2 succeeded but should have failed")
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	if err := mapper.Delete(ctx, "users.delete", user["ID"]); err != nil {
		log.Printf("Warning: cleanup failed: %v", err)
	}

	fmt.Println()
	return nil
}

func customActions(ctx context.Context, mapper *engine.Mapper) error {
	fmt.Println("--- Custom Actions ---")

	// Create some test users
	testUsers := []interface{}{
		map[string]interface{}{"Name": "Alice", "Email": "alice@example.com"},
		map[string]interface{}{"Name": "Bob", "Email": "bob@example.com"},
		map[string]interface{}{"Name": "Charlie", "Email": "charlie@example.com"},
	}

	fmt.Printf("Creating %d test users...\n", len(testUsers))
	if err := mapper.Insert(ctx, "users.insert", testUsers); err != nil {
		return fmt.Errorf("test user creation failed: %w", err)
	}
	fmt.Println("✓ Test users created")

	// Execute custom action: count users
	fmt.Println("\nExecuting custom action: count users")
	var countResult []interface{}
	err := mapper.Execute(ctx, "users.count", nil, &countResult)
	if err != nil {
		return fmt.Errorf("count action failed: %w", err)
	}

	if len(countResult) > 0 {
		if countMap, ok := countResult[0].(map[string]interface{}); ok {
			fmt.Printf("✓ Total users in database: %v\n", countMap["count"])
		}
	}

	// Execute custom action: search by email pattern
	fmt.Println("\nExecuting custom action: search by email pattern")
	var searchResults []interface{}
	err = mapper.Execute(ctx, "users.search-by-email", map[string]interface{}{
		"pattern": "%example.com%",
	}, &searchResults)
	if err != nil {
		return fmt.Errorf("search action failed: %w", err)
	}

	fmt.Printf("✓ Found %d users matching pattern\n", len(searchResults))
	for i, r := range searchResults {
		if userMap, ok := r.(map[string]interface{}); ok {
			fmt.Printf("  %d. %s (%s)\n", i+1, userMap["Name"], userMap["Email"])
		}
	}

	// Clean up test users
	fmt.Println("\nCleaning up test users...")
	for _, u := range testUsers {
		user := u.(map[string]interface{})
		if id, ok := user["ID"]; ok {
			if err := mapper.Delete(ctx, "users.delete", id); err != nil {
				log.Printf("Warning: cleanup failed for user %v: %v", id, err)
			}
		}
	}
	fmt.Println("✓ Test users cleaned up")

	fmt.Println()
	return nil
}
