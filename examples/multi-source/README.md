# Multi-Source CQRS Example

This example demonstrates using **multiple data sources** with **CQRS (Command Query Responsibility Segregation)** pattern.

## What This Example Shows

- **Multiple Sources**: Primary storage + cache storage
- **CQRS Pattern**: Separate read and write paths
- **Cache Management**: Population and invalidation
- **Source Selection**: Different sources for different operations

## Architecture

```
┌─────────────────────────────────────┐
│         Application                  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│      toutago-datamapper              │
│  ┌──────────────┐  ┌──────────────┐ │
│  │  Write Ops   │  │   Read Ops   │ │
│  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────┘
         │                    │
         ▼                    ▼
┌──────────────┐    ┌──────────────┐
│   Primary    │    │    Cache     │
│   Storage    │    │   Storage    │
└──────────────┘    └──────────────┘
```

## CQRS Pattern

### Commands (Write Operations)
- Always write to **primary source**
- Ensures data consistency
- Invalidates cache on updates

### Queries (Read Operations)
- Try **cache source** first (fast)
- Fallback to **primary source** (consistent)
- Applications can choose based on needs

## Configuration

```yaml
sources:
  primary:
    adapter: filesystem
    connection: "./data/primary"
  
  cache:
    adapter: filesystem
    connection: "./data/cache"

mappings:
  product-write:
    source: primary  # Writes go here
  
  product-read-cache:
    source: cache    # Fast reads
  
  product-read-primary:
    source: primary  # Consistent reads
```

## Running the Example

```bash
cd examples/multi-source
go run main.go
```

## Expected Output

```
=== Multi-Source CQRS Example ===

1. Creating products in primary storage...
   ✓ Created product: Laptop
   ✓ Created product: Mouse
   ✓ Created product: Keyboard

2. Reading from primary storage (cache miss scenario)...
   ✓ Read from primary: Laptop - $1299.99

3. Populating cache...
   ✓ Product cached: Laptop

4. Reading from cache (cache hit scenario)...
   ✓ Read from cache: Laptop - $1299.99 (fast!)

5. Updating product in primary storage...
   ✓ Updated product: Laptop - New price: $1199.99

6. Invalidating cache...
   ✓ Cache invalidated for: p1

7. Reading after cache invalidation...
   ✓ Read from primary: Laptop - $1199.99 (updated price!)

8. Listing all products from primary storage...
   ✓ Found 3 products:
      - Laptop: $1199.99 (stock: 45)
      - Mouse: $29.99 (stock: 200)
      - Keyboard: $149.99 (stock: 75)

=== CQRS Pattern Demonstrated ===
```

## Use Cases

### 1. **Performance Optimization**
```go
// Fast read from cache
mapper.Fetch(ctx, "products.product-read-cache", params, &product)

// Consistent read from primary
mapper.Fetch(ctx, "products.product-read-primary", params, &product)
```

### 2. **Write-Through Cache**
```go
// Update primary
mapper.Update(ctx, "products.product-write", product)

// Update cache
mapper.Update(ctx, "products.product-cache", product)
```

### 3. **Cache Invalidation**
```go
// Update primary
mapper.Update(ctx, "products.product-write", product)

// Invalidate cache
mapper.Delete(ctx, "products.product-cache", product.ID)
```

## Real-World Applications

### Database + Redis Cache
```yaml
sources:
  postgres:
    adapter: postgres
    connection: "${DB_CONNECTION}"
  
  redis:
    adapter: redis
    connection: "${REDIS_URL}"
```

### Read Replica Pattern
```yaml
sources:
  master:
    adapter: mysql
    connection: "${MASTER_DB}"
  
  replica:
    adapter: mysql
    connection: "${REPLICA_DB}"
```

### Multi-Region Setup
```yaml
sources:
  us-east:
    adapter: postgres
    connection: "${US_EAST_DB}"
  
  eu-west:
    adapter: postgres
    connection: "${EU_WEST_DB}"
```

## Key Benefits

1. **Performance**: Cache reduces database load
2. **Scalability**: Read replicas distribute load
3. **Flexibility**: Easy to change sources
4. **Consistency**: Write to single source of truth
5. **No Code Changes**: Configuration-driven

## Next Steps

- Try with real databases (MySQL + Redis)
- Implement cache warming strategies
- Add cache TTL and eviction policies
- Explore event-driven cache invalidation
