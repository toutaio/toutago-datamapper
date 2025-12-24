package main

import (
	"context"
	"fmt"
	"log"

	"github.com/toutago/toutago-datamapper/adapter"
	"github.com/toutago/toutago-datamapper/config"
	"github.com/toutago/toutago-datamapper/engine"
	"github.com/toutago/toutago-datamapper/filesystem"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          string
	AccountID   string
	Amount      float64
	Type        string // debit, credit
	Description string
	Balance     float64
}

// AccountSummary is a computed result
type AccountSummary struct {
	AccountID      string
	TotalDebits    float64
	TotalCredits   float64
	TransactionCount int
	CurrentBalance float64
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Custom Actions Example ===")
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

	// 1. Create sample transactions
	fmt.Println("1. Creating sample transactions...")
	transactions := []Transaction{
		{ID: "txn-001", AccountID: "acc-100", Amount: 1000.00, Type: "credit", Description: "Initial deposit", Balance: 1000.00},
		{ID: "txn-002", AccountID: "acc-100", Amount: 50.00, Type: "debit", Description: "ATM withdrawal", Balance: 950.00},
		{ID: "txn-003", AccountID: "acc-100", Amount: 200.00, Type: "credit", Description: "Salary payment", Balance: 1150.00},
		{ID: "txn-004", AccountID: "acc-100", Amount: 75.50, Type: "debit", Description: "Grocery shopping", Balance: 1074.50},
		{ID: "txn-005", AccountID: "acc-100", Amount: 30.00, Type: "debit", Description: "Restaurant", Balance: 1044.50},
		
		{ID: "txn-006", AccountID: "acc-101", Amount: 500.00, Type: "credit", Description: "Initial deposit", Balance: 500.00},
		{ID: "txn-007", AccountID: "acc-101", Amount: 100.00, Type: "debit", Description: "Online purchase", Balance: 400.00},
		{ID: "txn-008", AccountID: "acc-101", Amount: 250.00, Type: "credit", Description: "Freelance payment", Balance: 650.00},
	}

	for _, txn := range transactions {
		if err := mapper.Insert(ctx, "transactions.transaction-crud", txn); err != nil {
			log.Printf("Error creating transaction %s: %v", txn.ID, err)
		}
	}
	fmt.Printf("   ✓ Created %d transactions\n", len(transactions))
	fmt.Println()

	// 2. List all transactions (standard action)
	fmt.Println("2. Listing all transactions...")
	var allTxns []interface{}
	err = mapper.Execute(ctx, "transactions.list-all", nil, &allTxns)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d transactions\n", len(allTxns))
	}
	fmt.Println()

	// 3. Get transactions by account (custom action with params)
	fmt.Println("3. Getting transactions for account acc-100...")
	var accountTxns []interface{}
	err = mapper.Execute(ctx, "transactions.by-account", map[string]interface{}{
		"account_id": "acc-100",
	}, &accountTxns)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d transactions for acc-100\n", len(accountTxns))
		for i, txn := range accountTxns {
			t := txn.(map[string]interface{})
			fmt.Printf("      %d. %s - %s $%.2f (Balance: $%.2f)\n", 
				i+1, t["type"], t["description"], t["amount"], t["balance"])
		}
	}
	fmt.Println()

	// 4. Calculate account summary (custom aggregation)
	fmt.Println("4. Calculating account summary for acc-100...")
	var summary map[string]interface{}
	err = mapper.Execute(ctx, "transactions.account-summary", map[string]interface{}{
		"account_id": "acc-100",
	}, &summary)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   Account Summary:\n")
		fmt.Printf("   • Total Credits: $%.2f\n", summary["total_credits"])
		fmt.Printf("   • Total Debits: $%.2f\n", summary["total_debits"])
		fmt.Printf("   • Transaction Count: %.0f\n", summary["transaction_count"])
		fmt.Printf("   • Current Balance: $%.2f\n", summary["current_balance"])
	}
	fmt.Println()

	// 5. Get recent transactions (custom action with limit)
	fmt.Println("5. Getting 3 most recent transactions...")
	var recentTxns []interface{}
	err = mapper.Execute(ctx, "transactions.recent", map[string]interface{}{
		"limit": 3,
	}, &recentTxns)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Recent transactions:\n")
		for i, txn := range recentTxns {
			t := txn.(map[string]interface{})
			fmt.Printf("      %d. [%s] %s - $%.2f\n", 
				i+1, t["account_id"], t["description"], t["amount"])
		}
	}
	fmt.Println()

	// 6. Get transactions by type (custom filter)
	fmt.Println("6. Getting all credit transactions...")
	var creditTxns []interface{}
	err = mapper.Execute(ctx, "transactions.by-type", map[string]interface{}{
		"type": "credit",
	}, &creditTxns)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d credit transactions\n", len(creditTxns))
		var totalCredits float64
		for _, txn := range creditTxns {
			t := txn.(map[string]interface{})
			totalCredits += t["amount"].(float64)
		}
		fmt.Printf("   ✓ Total credit amount: $%.2f\n", totalCredits)
	}
	fmt.Println()

	// 7. Execute stored procedure simulation
	fmt.Println("7. Executing balance reconciliation (stored procedure simulation)...")
	var reconciled map[string]interface{}
	err = mapper.Execute(ctx, "transactions.reconcile-balance", map[string]interface{}{
		"account_id": "acc-100",
	}, &reconciled)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   ✓ Balance reconciled:\n")
		fmt.Printf("      Expected: $%.2f\n", reconciled["expected_balance"])
		fmt.Printf("      Actual: $%.2f\n", reconciled["actual_balance"])
		fmt.Printf("      Status: %s\n", reconciled["status"])
	}
	fmt.Println()

	fmt.Println("=== Custom Actions Demonstrated ===")
	fmt.Println()
	fmt.Println("Action Types:")
	fmt.Println("  • List/Query actions - Retrieve filtered data")
	fmt.Println("  • Aggregation actions - Calculate summaries")
	fmt.Println("  • Stored procedures - Complex business logic")
	fmt.Println("  • Custom operations - Domain-specific tasks")
	fmt.Println()
	fmt.Println("Benefits:")
	fmt.Println("  • Encapsulate complex queries")
	fmt.Println("  • Reusable business logic")
	fmt.Println("  • Database-agnostic interface")
	fmt.Println("  • Easy to test and maintain")
}
