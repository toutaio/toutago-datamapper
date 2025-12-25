# Example: toutago-datamapper with MySQL

This example demonstrates using toutago-datamapper with a MySQL database, showing full CRUD operations, bulk processing, optimistic locking, and custom actions.

## Important Note

This example requires the external MySQL adapter package:

```bash
go get github.com/toutaio/toutago-datamapper-mysql
```

The MySQL adapter is now maintained in a separate repository to keep the core library lightweight and database-agnostic.

## What This Example Shows

- ✅ **Basic CRUD** - Create, Read, Update, Delete operations
- ✅ **Bulk Operations** - High-performance batch inserts
- ✅ **Optimistic Locking** - Concurrent update protection with version fields
- ✅ **Custom Actions** - Complex queries and aggregations
- ✅ **Auto-increment IDs** - Handling database-generated primary keys
- ✅ **Prepared Statements** - SQL injection protection
- ✅ **Connection Pooling** - Efficient database connection management

## Prerequisites

### MySQL Database

You need a running MySQL instance. Choose one option:

**Option 1: Docker (Recommended)**
```bash
docker run --name mysql-test \
  -e MYSQL_ROOT_PASSWORD=testpass \
  -e MYSQL_DATABASE=testdb \
  -p 3306:3306 \
  -d mysql:8.0
```

**Option 2: Local MySQL**
- Install MySQL 5.7+ or 8.0+
- Create a database for testing

## Setup

### 1. Set Environment Variables

Create a `.env` file:
```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=testpass
DB_NAME=testdb
DB_SSL=false
```

Or export them directly:
```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=testpass
export DB_NAME=testdb
```

### 2. Initialize Database Schema

```bash
# If using Docker
docker exec -i mysql-test mysql -uroot -ptestpass testdb < schema.sql

# If using local MySQL
mysql -u root -p testdb < schema.sql
```

The schema creates:
- **users** table with auto-increment ID, timestamps, and version field
- **products** table for bulk operations demo

### 3. Run the Example

```bash
# Load environment variables (if using .env)
source .env

# Run the example
go run main.go
```

## Expected Output

```
=== MySQL Adapter Example ===

--- Basic CRUD Operations ---
Creating user: Alice Johnson (alice@example.com)
✓ User created with ID: 1

Fetching user with ID: 1
✓ User fetched: Alice Johnson (alice@example.com)

Updating user email to: alice.johnson@newdomain.com
✓ User updated
✓ Verified update: alice.johnson@newdomain.com

Deleting user with ID: 1
✓ User deleted
✓ Verified deletion (user not found)

--- Bulk Operations ---
Bulk inserting 4 products...
✓ 4 products inserted in 45ms (avg: 11ms per product)

Fetching all products...
✓ Fetched 4 products
  1. Laptop Pro - $1299.99 (Stock: 50)
  2. Wireless Mouse - $29.99 (Stock: 200)
  3. USB-C Hub - $49.99 (Stock: 150)
  4. Mechanical Keyboard - $149.99 (Stock: 75)

Updating product stock...
✓ All product stocks updated

Cleaning up test products...
✓ Test products cleaned up

--- Optimistic Locking ---
Creating user with version control...
✓ User created with ID: 2, Version: 1

Simulating concurrent updates...
Update 1: Changing email with Version 1
✓ Update 1 succeeded
Update 2: Trying to change email with stale Version 1
✓ Update 2 failed as expected (version mismatch)

Cleaning up...

--- Custom Actions ---
Creating 3 test users...
✓ Test users created

Executing custom action: count users
✓ Total users in database: 3

Executing custom action: search by email pattern
✓ Found 3 users matching pattern
  1. Alice (alice@example.com)
  2. Bob (bob@example.com)
  3. Charlie (charlie@example.com)

Cleaning up test users...
✓ Test users cleaned up

=== Example Complete ===
```

## Code Structure

### Domain Objects

```go
// Zero database dependencies
type User struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int
}
```

### Adapter Registration

```go
mapper.RegisterAdapter("mysql", func() adapter.Adapter {
    return mysql.NewMySQLAdapter()
})
```

### Operations

**Insert with Auto-generated ID:**
```go
newUser := map[string]interface{}{
    "Name":  "Alice Johnson",
    "Email": "alice@example.com",
}
mapper.Insert(ctx, "users.insert", newUser)
// newUser["ID"] now contains auto-generated value
```

**Fetch:**
```go
var user map[string]interface{}
mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
    "id": 123,
}, &user)
```

**Bulk Insert:**
```go
products := []interface{}{
    map[string]interface{}{"Name": "Product 1", ...},
    map[string]interface{}{"Name": "Product 2", ...},
}
mapper.Insert(ctx, "products.bulk-insert", products)
```

**Custom Action:**
```go
result, err := mapper.Execute(ctx, "users.count", nil)
```

## Configuration Highlights

### Source Configuration

```yaml
sources:
  mysql-db:
    adapter: mysql
    options:
      host: "${DB_HOST:-localhost}"
      port: ${DB_PORT:-3306}
      user: "${DB_USER:-root}"
      password: "${DB_PASSWORD}"
      database: "${DB_NAME:-testdb}"
      max_connections: 10
      max_idle: 5
```

### Mapping with Auto-increment

```yaml
operations:
  insert:
    statement: "users"
    properties:
      - object: Name
        data: name
      - object: Email
        data: email
    generated:
      - object: ID
        data: id  # Auto-populated after insert
```

### Optimistic Locking

```yaml
operations:
  update-versioned:
    statement: "users"
    properties:
      - object: Name
        data: name
      - object: Version
        data: version
    identifier:
      - object: ID
        data: id
    condition:
      - object: Version
        data: version  # UPDATE fails if version doesn't match
```

## Features Demonstrated

### 1. Basic CRUD
- Insert with auto-generated ID
- Fetch by ID
- Update existing record
- Delete record
- Verify operations

### 2. Bulk Operations
- Bulk insert (single SQL statement)
- Batch processing
- Performance measurement
- Stock updates (batch update)

### 3. Optimistic Locking
- Version-based concurrency control
- Detecting concurrent updates
- Preventing lost updates
- Handling version conflicts

### 4. Custom Actions
- COUNT queries
- Pattern matching (LIKE)
- Custom result sets
- Aggregations

## Real-World Applications

### E-Commerce Platform
```go
// High-performance product catalog updates
products := loadProductsFromCSV()
mapper.Insert(ctx, "products.bulk-insert", products)
```

### User Management System
```go
// Safe concurrent user updates with optimistic locking
user.Email = newEmail
user.Version++
err := mapper.Update(ctx, "users.update-versioned", user)
if err == adapter.ErrNotFound {
    // Handle conflict - user was modified by another request
}
```

### Reporting System
```go
// Custom aggregation queries
stats, _ := mapper.Execute(ctx, "users.statistics", params)
```

## Troubleshooting

### Connection Refused
```
Error: mysql: failed to ping database: dial tcp [::1]:3306: connect: connection refused
```

**Solution:** Ensure MySQL is running and accessible:
```bash
# Check Docker container
docker ps | grep mysql

# Check local MySQL
systemctl status mysql

# Test connection
mysql -h localhost -u root -p
```

### Access Denied
```
Error: Access denied for user 'root'@'localhost'
```

**Solution:** Verify credentials in environment variables:
```bash
echo $DB_USER
echo $DB_PASSWORD
```

### Table Doesn't Exist
```
Error: Table 'testdb.users' doesn't exist
```

**Solution:** Run the schema file:
```bash
docker exec -i mysql-test mysql -uroot -ptestpass testdb < schema.sql
```

### Version Conflicts
```
Error: record not found
```

This is expected behavior when optimistic locking detects a conflict. Reload the record and retry the update.

## Performance Tips

1. **Use Bulk Operations** - For inserting multiple records, use `bulk: true`
2. **Connection Pooling** - Adjust `max_connections` based on load
3. **Indexes** - Add indexes on frequently queried columns
4. **Prepared Statements** - Automatically used by the adapter
5. **Batch Processing** - Process large datasets in chunks

## Clean Up

### Stop and Remove Docker Container
```bash
docker stop mysql-test
docker rm mysql-test
```

### Clean Database (if keeping MySQL running)
```bash
docker exec -i mysql-test mysql -uroot -ptestpass testdb <<EOF
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
EOF
```

## Next Steps

- **Try PostgreSQL adapter** - Similar API, different database
- **Implement CQRS** - Separate read replicas from write master
- **Add caching** - Combine with filesystem or Redis adapter
- **Custom queries** - Add your own actions for complex operations
- **Transactions** - Wrap multiple operations in transactions

## Resources

- [MySQL Adapter Documentation](../../adapter/mysql/README.md)
- [toutago-datamapper Documentation](../../README.md)
- [MySQL Docker Image](https://hub.docker.com/_/mysql)
- [MySQL Documentation](https://dev.mysql.com/doc/)
