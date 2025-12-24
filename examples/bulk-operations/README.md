# Bulk Operations Example

This example demonstrates **bulk operations** for high-performance data processing with toutago-datamapper.

## What This Example Shows

- **Bulk Insert**: Create multiple records in one operation
- **Bulk Update**: Modify multiple records efficiently
- **Bulk Delete**: Remove multiple records at once
- **Bulk Read**: Fetch large datasets
- **Performance Metrics**: Measure operation timing

## Why Bulk Operations?

### Performance Benefits
- **Reduced Overhead**: Single transaction vs multiple
- **Better Throughput**: Network/IO optimization
- **Atomic Operations**: All-or-nothing semantics
- **Resource Efficiency**: Lower CPU/memory per record

### Use Cases
- Data imports/exports
- Batch processing
- ETL pipelines
- Report generation
- Cache warming

## Configuration

```yaml
mappings:
  order-bulk:
    operations:
      insert:
        bulk: true  # Enable bulk mode
        statement: "orders/{id}.json"
      
      update:
        bulk: true
        statement: "orders/{id}.json"
      
      delete:
        bulk: true
        statement: "orders/{id}.json"
```

## Running the Example

```bash
cd examples/bulk-operations
go run main.go
```

## Expected Output

```
=== Bulk Operations Example ===

1. Bulk Insert: Creating 100 orders...
   ✓ Created 100 orders in 45ms
   ✓ Average: 450µs per order

2. Bulk Read: Fetching all orders...
   ✓ Fetched 100 orders in 23ms

   Statistics:
   • Total Revenue: $54599.00
   • Order Status:
     - pending: 25 orders
     - processing: 25 orders
     - completed: 25 orders
     - cancelled: 25 orders

3. Bulk Update: Processing pending orders...
   ✓ Updated 20 orders to 'processing' in 8ms

4. Filtered Read: Fetching completed orders...
   ✓ Found 25 completed orders

5. Bulk Delete: Canceling old pending orders...
   ✓ Deleted 10 orders in 4ms

6. Final order count...
   ✓ Total orders remaining: 90

=== Bulk Operations Complete ===
```

## Usage Patterns

### 1. Bulk Insert
```go
orders := []Order{
    {ID: "1", Total: 100.00, Status: "pending"},
    {ID: "2", Total: 200.00, Status: "pending"},
    {ID: "3", Total: 300.00, Status: "pending"},
}

// Insert all at once
mapper.Insert(ctx, "orders.order-bulk", orders)
```

### 2. Bulk Update
```go
// Modify multiple records
for i := range orders {
    orders[i].Status = "processing"
}

// Update all at once
mapper.Update(ctx, "orders.order-bulk", orders)
```

### 3. Bulk Delete
```go
// Delete by IDs
ids := []string{"1", "2", "3"}
mapper.Delete(ctx, "orders.order-bulk", ids)
```

### 4. Bulk Read
```go
var orders []Order
mapper.FetchMulti(ctx, "orders.order-list", nil, &orders)
```

## Performance Comparison

### Individual Operations
```
100 inserts × 2ms = 200ms total
```

### Bulk Operation
```
1 bulk insert (100 records) = 45ms total
Speedup: 4.4x faster!
```

## Database-Specific Optimizations

### MySQL
```sql
-- Bulk Insert uses multi-row INSERT
INSERT INTO orders (id, total, status) VALUES
  ('1', 100.00, 'pending'),
  ('2', 200.00, 'pending'),
  ('3', 300.00, 'pending');
```

### PostgreSQL
```sql
-- Can use COPY for even faster bulk loads
COPY orders FROM STDIN;
```

### MongoDB
```javascript
// Uses insertMany
db.orders.insertMany([
  {id: '1', total: 100.00},
  {id: '2', total: 200.00},
  {id: '3', total: 300.00}
])
```

## Best Practices

### 1. Batch Size
```go
// Process in chunks to avoid memory issues
const batchSize = 1000

for i := 0; i < len(allOrders); i += batchSize {
    end := i + batchSize
    if end > len(allOrders) {
        end = len(allOrders)
    }
    
    batch := allOrders[i:end]
    mapper.Insert(ctx, "orders.order-bulk", batch)
}
```

### 2. Error Handling
```go
// Handle partial failures
for i := 0; i < len(batches); i++ {
    if err := mapper.Insert(ctx, "orders.order-bulk", batches[i]); err != nil {
        log.Printf("Batch %d failed: %v", i, err)
        // Retry or skip
    }
}
```

### 3. Progress Tracking
```go
total := len(orders)
for i := 0; i < total; i += batchSize {
    // Process batch
    mapper.Insert(ctx, "orders.order-bulk", batch)
    
    // Show progress
    progress := float64(i+batchSize) / float64(total) * 100
    fmt.Printf("Progress: %.1f%%\n", progress)
}
```

### 4. Memory Management
```go
// Stream large datasets instead of loading all at once
for {
    batch := fetchNextBatch() // Get 1000 records
    if len(batch) == 0 {
        break
    }
    
    mapper.Insert(ctx, "orders.order-bulk", batch)
}
```

## Real-World Scenarios

### Data Migration
```go
// Migrate from old system
oldData := loadFromLegacySystem()

const batchSize = 500
for i := 0; i < len(oldData); i += batchSize {
    batch := convertToNewFormat(oldData[i:i+batchSize])
    mapper.Insert(ctx, "orders.order-bulk", batch)
}
```

### Nightly Batch Processing
```go
// Process daily orders
orders := fetchPendingOrders()

// Bulk update statuses
mapper.Update(ctx, "orders.order-bulk", orders)

// Bulk insert to reporting database
mapper.Insert(ctx, "reports.daily-orders", orders)
```

### Data Export
```go
// Export all orders
var orders []Order
mapper.FetchMulti(ctx, "orders.order-list", nil, &orders)

// Write to CSV/Excel
exportToCSV(orders, "orders-export.csv")
```

## Performance Tips

1. **Use Transactions**: Wrap bulk operations in transactions
2. **Disable Indexes**: Temporarily for large imports
3. **Batch Commits**: Commit in chunks, not per-record
4. **Parallel Processing**: Use goroutines for independent batches
5. **Monitor Memory**: Watch heap growth with large batches

## Monitoring

```go
import "time"

start := time.Now()
count := 0

for batch := range batches {
    mapper.Insert(ctx, "orders.order-bulk", batch)
    count += len(batch)
}

elapsed := time.Since(start)
rate := float64(count) / elapsed.Seconds()

fmt.Printf("Processed %d records in %v (%.0f records/sec)\n", 
    count, elapsed, rate)
```

## Next Steps

- Try with real databases (MySQL, PostgreSQL)
- Implement parallel batch processing
- Add retry logic for failed batches
- Explore streaming APIs
- Set up performance benchmarks
