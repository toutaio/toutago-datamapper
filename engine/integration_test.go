package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
	"github.com/toutaio/toutago-datamapper/filesystem"
)

// setupMapperWithFilesystem is a helper to create a mapper with filesystem adapter registered
func setupMapperWithFilesystem(t *testing.T, configContent, tempDir string) *Mapper {
	t.Helper()

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		t.Fatalf("Failed to create mapper: %v", err)
	}

	// Register filesystem adapter factory
	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
		basePath := source.Connection
		if basePath == "" {
			basePath = tempDir
		}
		return filesystem.NewFilesystemAdapter(basePath)
	})

	return mapper
}

// Integration tests for end-to-end scenarios with multiple sources

// TestIntegration_MultiSourceCRUD tests CRUD operations across multiple adapters
func TestIntegration_MultiSourceCRUD(t *testing.T) {
	tempDir := t.TempDir()

	// Create multi-source configuration
	configContent := fmt.Sprintf(`
namespace: test
version: "1.0"

sources:
  primary:
    adapter: filesystem
    connection: %s/primary
  secondary:
    adapter: filesystem
    connection: %s/secondary

mappings:
  user:
    object: User
    source: primary
    operations:
      fetch:
        statement: "user_{id}.json"
        result:
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
      insert:
        statement: "user_{id}.json"
        properties:
          - object: ID
            field: id
          - object: Name
            field: name
      update:
        statement: "user_{id}.json"
        identifier:
          - object: ID
            field: id
        properties:
          - object: Name
            field: name
      delete:
        statement: "user_{id}.json"
        identifier:
          - object: ID
            field: id
  cache:
    object: User
    source: secondary
    operations:
      fetch:
        statement: "cache_{id}.json"
        result:
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
      insert:
        statement: "cache_{id}.json"
        properties:
          - object: ID
            field: id
          - object: Name
            field: name
`, tempDir, tempDir)

	mapper := setupMapperWithFilesystem(t, configContent, tempDir)
	defer mapper.Close()

	ctx := context.Background()

	// Test 1: Insert into primary
	user := map[string]interface{}{
		"id":   "123",
		"name": "Alice",
	}
	if err := mapper.Insert(ctx, "test.user", user); err != nil {
		t.Errorf("Insert to primary failed: %v", err)
	}

	// Test 2: Fetch from primary
	var fetched map[string]interface{}
	params := map[string]interface{}{"id": "123"}
	if err := mapper.Fetch(ctx, "test.user", params, &fetched); err != nil {
		t.Errorf("Fetch from primary failed: %v", err)
	}
	if fetched["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", fetched["name"])
	}

	// Test 3: Insert into cache
	if err := mapper.Insert(ctx, "test.cache", user); err != nil {
		t.Errorf("Insert to cache failed: %v", err)
	}

	// Test 4: Fetch from cache
	var cachedUser map[string]interface{}
	if err := mapper.Fetch(ctx, "test.cache", params, &cachedUser); err != nil {
		t.Errorf("Fetch from cache failed: %v", err)
	}

	// Test 5: Update primary
	user["name"] = "Alice Updated"
	if err := mapper.Update(ctx, "test.user", user); err != nil {
		t.Errorf("Update primary failed: %v", err)
	}

	// Test 6: Verify update
	if err := mapper.Fetch(ctx, "test.user", params, &fetched); err != nil {
		t.Errorf("Fetch after update failed: %v", err)
	}
	if fetched["name"] != "Alice Updated" {
		t.Errorf("Expected name=Alice Updated, got %v", fetched["name"])
	}

	// Test 7: Delete from both sources
	if err := mapper.Delete(ctx, "test.user", "123"); err != nil {
		t.Errorf("Delete from primary failed: %v", err)
	}
	if err := mapper.Delete(ctx, "test.cache", "123"); err != nil {
		t.Errorf("Delete from cache failed: %v", err)
	}

	// Test 8: Verify deletion
	if err := mapper.Fetch(ctx, "test.user", params, &fetched); err != adapter.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

// TestIntegration_ConcurrentOperations tests thread safety
func TestIntegration_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()

	configContent := fmt.Sprintf(`
namespace: concurrent
version: "1.0"

sources:
  store:
    adapter: filesystem
    connection: %s

mappings:
  item:
    object: Item
    source: store
    operations:
      insert:
        statement: "item_{id}.json"
        properties:
          - object: ID
            field: id
          - object: Value
            field: value
      fetch:
        statement: "item_{id}.json"
        result:
          properties:
            - object: ID
              field: id
            - object: Value
              field: value
`, tempDir)

	mapper := setupMapperWithFilesystem(t, configContent, tempDir)
	defer mapper.Close()

	ctx := context.Background()
	numGoroutines := 10
	numOpsPerGoroutine := 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOpsPerGoroutine)

	// Concurrent inserts
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < numOpsPerGoroutine; i++ {
				item := map[string]interface{}{
					"id":    fmt.Sprintf("g%d-i%d", goroutineID, i),
					"value": fmt.Sprintf("value-%d-%d", goroutineID, i),
				}
				if err := mapper.Insert(ctx, "concurrent.item", item); err != nil {
					errors <- fmt.Errorf("goroutine %d insert %d: %w", goroutineID, i, err)
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	// Verify all items were inserted
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < numOpsPerGoroutine; i++ {
			var item map[string]interface{}
			params := map[string]interface{}{
				"id": fmt.Sprintf("g%d-i%d", g, i),
			}
			if err := mapper.Fetch(ctx, "concurrent.item", params, &item); err != nil {
				t.Errorf("Failed to fetch item g%d-i%d: %v", g, i, err)
			}
		}
	}
}

// TestIntegration_BulkOperations tests bulk insert/update/delete
func TestIntegration_BulkOperations(t *testing.T) {
	tempDir := t.TempDir()

	configContent := fmt.Sprintf(`
namespace: bulk
version: "1.0"

sources:
  store:
    adapter: filesystem
    connection: %s

mappings:
  product:
    object: Product
    source: store
    operations:
      insert:
        statement: "product_{id}.json"
        properties:
          - object: ID
            field: id
          - object: Name
            field: name
          - object: Price
            field: price
      fetch-multi:
        statement: "product_*.json"
        result:
          multi: true
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
            - object: Price
              field: price
      delete:
        statement: "product_{id}.json"
        identifier:
          - object: ID
            field: id
`, tempDir)

	mapper := setupMapperWithFilesystem(t, configContent, tempDir)
	defer mapper.Close()

	ctx := context.Background()

	// Insert 100 products
	for i := 0; i < 100; i++ {
		product := map[string]interface{}{
			"id":    fmt.Sprintf("p%03d", i),
			"name":  fmt.Sprintf("Product %d", i),
			"price": float64(i * 10),
		}
		if err := mapper.Insert(ctx, "bulk.product", product); err != nil {
			t.Fatalf("Bulk insert failed at %d: %v", i, err)
		}
	}

	// Fetch all
	var results []map[string]interface{}
	if err := mapper.FetchMulti(ctx, "bulk.product", nil, &results); err != nil {
		t.Fatalf("FetchMulti failed: %v", err)
	}
	if len(results) != 100 {
		t.Errorf("Expected 100 products, got %d", len(results))
	}

	// Delete all
	for i := 0; i < 100; i++ {
		if err := mapper.Delete(ctx, "bulk.product", fmt.Sprintf("p%03d", i)); err != nil {
			t.Errorf("Bulk delete failed at %d: %v", i, err)
		}
	}

	// Verify deletion
	results = nil
	if err := mapper.FetchMulti(ctx, "bulk.product", nil, &results); err != nil {
		t.Fatalf("FetchMulti after delete failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 products after delete, got %d", len(results))
	}
}

// TestIntegration_ErrorHandling tests error scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	configContent := fmt.Sprintf(`
namespace: errors
version: "1.0"

sources:
  store:
    adapter: filesystem
    connection: %s

mappings:
  user:
    object: User
    source: store
    operations:
      fetch:
        statement: "user_{id}.json"
        result:
          properties:
            - object: ID
              field: id
`, tempDir)

	mapper := setupMapperWithFilesystem(t, configContent, tempDir)
	defer mapper.Close()

	ctx := context.Background()

	// Test 1: Fetch non-existent item
	var user map[string]interface{}
	if err := mapper.Fetch(ctx, "errors.user", map[string]interface{}{"id": "999"}, &user); err != adapter.ErrNotFound {
		t.Errorf("Expected ErrNotFound for missing item, got %v", err)
	}

	// Test 2: Invalid mapping ID
	if err := mapper.Fetch(ctx, "invalid.mapping", nil, &user); err == nil {
		t.Error("Expected error for invalid mapping ID")
	}

	// Test 3: Nil context (should handle gracefully)
	if err := mapper.Fetch(context.Background(), "errors.user", map[string]interface{}{"id": "999"}, &user); err != adapter.ErrNotFound {
		t.Errorf("Expected ErrNotFound with background context, got %v", err)
	}
}

// TestIntegration_AdapterLifecycle tests adapter connection and cleanup
func TestIntegration_AdapterLifecycle(t *testing.T) {
	tempDir := t.TempDir()

	configContent := fmt.Sprintf(`
namespace: lifecycle
version: "1.0"

sources:
  store1:
    adapter: filesystem
    connection: %s/store1
  store2:
    adapter: filesystem
    connection: %s/store2

mappings:
  data1:
    object: Data
    source: store1
    operations:
      insert:
        statement: "data.json"
        properties:
          - object: Value
            field: value
  data2:
    object: Data
    source: store2
    operations:
      insert:
        statement: "data.json"
        properties:
          - object: Value
            field: value
`, tempDir, tempDir)

	mapper := setupMapperWithFilesystem(t, configContent, tempDir)

	ctx := context.Background()

	// Use both adapters
	data1 := map[string]interface{}{"value": "test1"}
	data2 := map[string]interface{}{"value": "test2"}

	if err := mapper.Insert(ctx, "lifecycle.data1", data1); err != nil {
		t.Errorf("Insert to store1 failed: %v", err)
	}
	if err := mapper.Insert(ctx, "lifecycle.data2", data2); err != nil {
		t.Errorf("Insert to store2 failed: %v", err)
	}

	// Close mapper (should close all adapters)
	if err := mapper.Close(); err != nil {
		t.Errorf("Mapper close failed: %v", err)
	}

	// Verify adapters are closed (operations should fail)
	if err := mapper.Insert(ctx, "lifecycle.data1", data1); err == nil {
		t.Error("Expected error when using mapper after close")
	}
}
