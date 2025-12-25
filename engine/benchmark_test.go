package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-datamapper/config"
)

// Benchmark tests for performance measurement

// BenchmarkMapper_Insert measures insert operation performance
func BenchmarkMapper_Insert(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
          - object: Value
            field: value
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()
	item := map[string]interface{}{
		"id":    "test",
		"name":  "Test Item",
		"value": 12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item["id"] = fmt.Sprintf("item-%d", i)
		if err := mapper.Insert(ctx, "bench.item", item); err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

// BenchmarkMapper_Fetch measures fetch operation performance
func BenchmarkMapper_Fetch(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
      fetch:
        statement: "item_{id}.json"
        result:
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()

	// Prepare data
	item := map[string]interface{}{
		"id":   "bench-item",
		"name": "Benchmark Item",
	}
	if err := mapper.Insert(ctx, "bench.item", item); err != nil {
		b.Fatalf("Setup insert failed: %v", err)
	}

	params := map[string]interface{}{"id": "bench-item"}
	var result map[string]interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := mapper.Fetch(ctx, "bench.item", params, &result); err != nil {
			b.Fatalf("Fetch failed: %v", err)
		}
	}
}

// BenchmarkMapper_Update measures update operation performance
func BenchmarkMapper_Update(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
      update:
        statement: "item_{id}.json"
        identifier:
          - object: ID
            field: id
        properties:
          - object: Name
            field: name
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()

	// Prepare data
	item := map[string]interface{}{
		"id":   "bench-item",
		"name": "Original Name",
	}
	if err := mapper.Insert(ctx, "bench.item", item); err != nil {
		b.Fatalf("Setup insert failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item["name"] = fmt.Sprintf("Updated Name %d", i)
		if err := mapper.Update(ctx, "bench.item", item); err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkMapper_Delete measures delete operation performance
func BenchmarkMapper_Delete(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
      delete:
        statement: "item_{id}.json"
        identifier:
          - object: ID
            field: id
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()

	// Pre-populate items
	items := make([]map[string]interface{}, b.N)
	for i := 0; i < b.N; i++ {
		items[i] = map[string]interface{}{
			"id":   fmt.Sprintf("item-%d", i),
			"name": fmt.Sprintf("Item %d", i),
		}
		if err := mapper.Insert(ctx, "bench.item", items[i]); err != nil {
			b.Fatalf("Setup insert %d failed: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := mapper.Delete(ctx, "bench.item", fmt.Sprintf("item-%d", i)); err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// BenchmarkMapper_BulkInsert measures bulk insert performance
func BenchmarkMapper_BulkInsert(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Insert batches of 100
		for j := 0; j < 100; j++ {
			item := map[string]interface{}{
				"id":   fmt.Sprintf("batch-%d-%d", i, j),
				"name": fmt.Sprintf("Batch Item %d-%d", i, j),
			}
			if err := mapper.Insert(ctx, "bench.item", item); err != nil {
				b.Fatalf("Bulk insert failed: %v", err)
			}
		}
	}
}

// BenchmarkPropertyMapper_MapToObject measures property mapping to object
func BenchmarkPropertyMapper_MapToObject(b *testing.B) {
	pm := NewPropertyMapper()

	type TestStruct struct {
		ID    string
		Name  string
		Email string
		Age   int
	}

	data := map[string]interface{}{
		"id":    "123",
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
		{Object: "Age", Field: "age"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target TestStruct
		if err := pm.MapToObject(data, &target, mappings); err != nil {
			b.Fatalf("MapToObject failed: %v", err)
		}
	}
}

// BenchmarkPropertyMapper_MapFromObject measures property mapping from object
func BenchmarkPropertyMapper_MapFromObject(b *testing.B) {
	pm := NewPropertyMapper()

	type TestStruct struct {
		ID    string
		Name  string
		Email string
		Age   int
	}

	obj := TestStruct{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
		{Object: "Age", Field: "age"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := pm.MapFromObject(obj, mappings); err != nil {
			b.Fatalf("MapFromObject failed: %v", err)
		}
	}
}

// BenchmarkMapper_ConcurrentInserts measures concurrent insert performance
func BenchmarkMapper_ConcurrentInserts(b *testing.B) {
	tempDir := b.TempDir()

	configContent := fmt.Sprintf(`
namespace: bench
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
          - object: Name
            field: name
`, tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	mapper, err := NewMapper(configPath)
	if err != nil {
		b.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			item := map[string]interface{}{
				"id":   fmt.Sprintf("concurrent-%d", i),
				"name": fmt.Sprintf("Concurrent Item %d", i),
			}
			if err := mapper.Insert(ctx, "bench.item", item); err != nil {
				b.Fatalf("Concurrent insert failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkConfigParser_LoadFile measures configuration loading performance
func BenchmarkConfigParser_LoadFile(b *testing.B) {
	tempDir := b.TempDir()

	configContent := `
namespace: bench
version: "1.0"

sources:
  db:
    adapter: filesystem
    connection: "/tmp/data"

mappings:
  user:
    object: User
    source: db
    operations:
      fetch:
        statement: "user_{id}.json"
        result:
          properties:
            - object: ID
              field: id
`

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := config.NewParser()
		if err := parser.LoadFile(configPath); err != nil {
			b.Fatalf("LoadFile failed: %v", err)
		}
	}
}
