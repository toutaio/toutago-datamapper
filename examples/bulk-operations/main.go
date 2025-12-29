package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
	"github.com/toutaio/toutago-datamapper/engine"
	"github.com/toutaio/toutago-datamapper/filesystem"
)

// Order represents an e-commerce order
type Order struct {
	ID         string
	CustomerID string
	Total      float64
	Status     string
	Items      []OrderItem
	CreatedAt  time.Time
}

// OrderItem represents a line item
type OrderItem struct {
	ProductID string
	Quantity  int
	Price     float64
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Bulk Operations Example ===")
	fmt.Println()

	// Create mapper
	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	// Register adapter
	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
		return filesystem.NewFilesystemAdapter(source.Connection)
	})

	// 1. Bulk Insert - Create many orders at once
	fmt.Println("1. Bulk Insert: Creating 100 orders...")
	orders := generateOrders(100)

	start := time.Now()
	if err := mapper.Insert(ctx, "orders.order-bulk", orders); err != nil {
		log.Fatalf("Bulk insert failed: %v", err)
	}
	elapsed := time.Since(start)

	fmt.Printf("   ✓ Created 100 orders in %v\n", elapsed)
	fmt.Printf("   ✓ Average: %v per order\n", elapsed/100)
	fmt.Println()

	// 2. Bulk Read - Fetch all orders
	fmt.Println("2. Bulk Read: Fetching all orders...")
	var allOrders []Order

	start = time.Now()
	err = mapper.FetchMulti(ctx, "orders.order-list", nil, &allOrders)
	elapsed = time.Since(start)

	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Fetched %d orders in %v\n", len(allOrders), elapsed)

		// Calculate statistics
		var totalRevenue float64
		statusCount := make(map[string]int)
		for _, order := range allOrders {
			totalRevenue += order.Total
			statusCount[order.Status]++
		}

		fmt.Printf("\n   Statistics:\n")
		fmt.Printf("   • Total Revenue: $%.2f\n", totalRevenue)
		fmt.Printf("   • Order Status:\n")
		for status, count := range statusCount {
			fmt.Printf("     - %s: %d orders\n", status, count)
		}
	}
	fmt.Println()

	// 3. Bulk Update - Update multiple orders
	fmt.Println("3. Bulk Update: Processing pending orders...")

	// Update first 20 orders to "processing"
	var ordersToUpdate []Order
	for i := 0; i < 20 && i < len(allOrders); i++ {
		if allOrders[i].Status == "pending" {
			allOrders[i].Status = "processing"
			ordersToUpdate = append(ordersToUpdate, allOrders[i])
		}
	}

	start = time.Now()
	if err := mapper.Update(ctx, "orders.order-bulk", ordersToUpdate); err != nil {
		log.Printf("Bulk update failed: %v", err)
	} else {
		elapsed = time.Since(start)
		fmt.Printf("   ✓ Updated %d orders to 'processing' in %v\n", len(ordersToUpdate), elapsed)
	}
	fmt.Println()

	// 4. Filtered Read - Fetch by status
	fmt.Println("4. Filtered Read: Fetching completed orders...")
	var completedOrders []Order

	// Note: This is a simple example. Real implementation would filter in adapter
	err = mapper.FetchMulti(ctx, "orders.order-list", nil, &completedOrders)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		// Filter in application (adapter should do this)
		var filtered []Order
		for _, order := range completedOrders {
			if order.Status == "completed" {
				filtered = append(filtered, order)
			}
		}
		fmt.Printf("   ✓ Found %d completed orders\n", len(filtered))
	}
	fmt.Println()

	// 5. Bulk Delete - Cancel old pending orders
	fmt.Println("5. Bulk Delete: Canceling old pending orders...")

	// Find orders to cancel (first 10 pending)
	var idsToDelete []string
	count := 0
	for _, order := range allOrders {
		if order.Status == "pending" && count < 10 {
			idsToDelete = append(idsToDelete, order.ID)
			count++
		}
	}

	start = time.Now()
	if err := mapper.Delete(ctx, "orders.order-bulk", idsToDelete); err != nil {
		log.Printf("Bulk delete failed: %v", err)
	} else {
		elapsed = time.Since(start)
		fmt.Printf("   ✓ Deleted %d orders in %v\n", len(idsToDelete), elapsed)
	}
	fmt.Println()

	// 6. Final count
	fmt.Println("6. Final order count...")
	var finalOrders []Order
	err = mapper.FetchMulti(ctx, "orders.order-list", nil, &finalOrders)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Total orders remaining: %d\n", len(finalOrders))
	}
	fmt.Println()

	fmt.Println("=== Bulk Operations Complete ===")
	fmt.Println()
	fmt.Println("Performance Benefits:")
	fmt.Println("  • Single transaction for multiple records")
	fmt.Println("  • Reduced network overhead")
	fmt.Println("  • Better resource utilization")
	fmt.Println("  • Atomic operations")
	fmt.Println()
	fmt.Println("Check ./data/orders directory for created files!")
}

// generateOrders creates test order data
func generateOrders(count int) []Order {
	orders := make([]Order, count)
	statuses := []string{"pending", "processing", "completed", "cancelled"}

	for i := 0; i < count; i++ {
		orders[i] = Order{
			ID:         fmt.Sprintf("ord-%03d", i+1),
			CustomerID: fmt.Sprintf("cust-%d", (i%20)+1),
			Total:      float64(50+i*10) + 0.99,
			Status:     statuses[i%len(statuses)],
			Items: []OrderItem{
				{
					ProductID: fmt.Sprintf("prod-%d", (i%10)+1),
					Quantity:  (i % 5) + 1,
					Price:     float64(20+i*2) + 0.99,
				},
			},
			CreatedAt: time.Now().Add(time.Duration(-i) * time.Hour),
		}
	}

	return orders
}
