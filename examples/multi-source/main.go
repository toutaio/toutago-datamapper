package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toutago/toutago-datamapper/adapter"
	"github.com/toutago/toutago-datamapper/config"
	"github.com/toutago/toutago-datamapper/engine"
	"github.com/toutago/toutago-datamapper/filesystem"
)

// Product represents our domain object
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Stock       int
	UpdatedAt   time.Time
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Multi-Source CQRS Example ===")
	fmt.Println()

	// Create mapper from config
	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	// Register filesystem adapter (simulating both cache and primary storage)
	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
		return filesystem.NewFilesystemAdapter(source.Connection)
	})

	// Demonstrate CQRS pattern: Read from cache, write to primary

	// 1. Create products (writes to primary source)
	fmt.Println("1. Creating products in primary storage...")
	products := []Product{
		{
			ID:          "p1",
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       1299.99,
			Stock:       50,
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "p2",
			Name:        "Mouse",
			Description: "Wireless mouse",
			Price:       29.99,
			Stock:       200,
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "p3",
			Name:        "Keyboard",
			Description: "Mechanical keyboard",
			Price:       149.99,
			Stock:       75,
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range products {
		if err := mapper.Insert(ctx, "products.product-write", p); err != nil {
			log.Printf("Error creating product %s: %v", p.ID, err)
		} else {
			fmt.Printf("   ✓ Created product: %s\n", p.Name)
		}
	}
	fmt.Println()

	// 2. Read from primary (simulating cache miss)
	fmt.Println("2. Reading from primary storage (cache miss scenario)...")
	var product Product
	err = mapper.Fetch(ctx, "products.product-read-primary", map[string]interface{}{"id": "p1"}, &product)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Read from primary: %s - $%.2f\n", product.Name, product.Price)
	}
	fmt.Println()

	// 3. Simulate cache population (copy to cache)
	fmt.Println("3. Populating cache...")
	if err := mapper.Insert(ctx, "products.product-cache", product); err != nil {
		log.Printf("Error populating cache: %v", err)
	} else {
		fmt.Printf("   ✓ Product cached: %s\n", product.Name)
	}
	fmt.Println()

	// 4. Read from cache (fast read)
	fmt.Println("4. Reading from cache (cache hit scenario)...")
	var cachedProduct Product
	err = mapper.Fetch(ctx, "products.product-read-cache", map[string]interface{}{"id": "p1"}, &cachedProduct)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Read from cache: %s - $%.2f (fast!)\n", cachedProduct.Name, cachedProduct.Price)
	}
	fmt.Println()

	// 5. Update product (write to primary, invalidate cache)
	fmt.Println("5. Updating product in primary storage...")
	product.Price = 1199.99
	product.Stock = 45
	product.UpdatedAt = time.Now()

	if err := mapper.Update(ctx, "products.product-write", product); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Updated product: %s - New price: $%.2f\n", product.Name, product.Price)
	}
	fmt.Println()

	// 6. Invalidate cache (delete from cache)
	fmt.Println("6. Invalidating cache...")
	if err := mapper.Delete(ctx, "products.product-cache", product.ID); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Cache invalidated for: %s\n", product.ID)
	}
	fmt.Println()

	// 7. Read again - will need to fetch from primary
	fmt.Println("7. Reading after cache invalidation...")
	var updatedProduct Product
	err = mapper.Fetch(ctx, "products.product-read-primary", map[string]interface{}{"id": "p1"}, &updatedProduct)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Read from primary: %s - $%.2f (updated price!)\n", updatedProduct.Name, updatedProduct.Price)
	}
	fmt.Println()

	// 8. List all products from primary
	fmt.Println("8. Listing all products from primary storage...")
	var allProducts []Product
	err = mapper.FetchMulti(ctx, "products.product-list", nil, &allProducts)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d products:\n", len(allProducts))
		for _, p := range allProducts {
			fmt.Printf("      - %s: $%.2f (stock: %d)\n", p.Name, p.Price, p.Stock)
		}
	}
	fmt.Println()

	fmt.Println("=== CQRS Pattern Demonstrated ===")
	fmt.Println()
	fmt.Println("Key Concepts:")
	fmt.Println("  • Writes go to primary storage")
	fmt.Println("  • Reads can use cache for performance")
	fmt.Println("  • Cache invalidation on updates")
	fmt.Println("  • Multiple data sources configured separately")
	fmt.Println()
	fmt.Println("Check ./data/primary and ./data/cache directories!")
}
