# Simple CRUD Example

This example demonstrates basic Create, Read, Update, Delete (CRUD) operations using the toutago-datamapper library with the filesystem adapter.

## What This Example Shows

- ✅ Creating a mapper from a configuration file
- ✅ Registering a custom adapter (filesystem)
- ✅ Inserting new objects (CREATE)
- ✅ Fetching a single object by ID (READ)
- ✅ Listing multiple objects (READ)
- ✅ Updating existing objects (UPDATE)
- ✅ Deleting objects (DELETE)
- ✅ Working with filesystem storage (JSON files)

## Domain Object

```go
type User struct {
    ID    string
    Name  string
    Email string
    Age   int
}
```

## Configuration

The `config.yaml` file defines:

1. **Source**: A filesystem adapter pointing to `./data` directory
2. **Mappings**:
   - `user-crud`: CRUD operations for individual users
   - `user-list`: Listing all users

## Running the Example

```bash
cd examples/simple-crud
go run main.go
```

## Output

The example will:
1. Create 3 users (Alice, Bob, Carol)
2. Fetch Alice's details
3. List all users
4. Update Bob's information
5. Delete Carol
6. Show final count (2 users)

All data is stored in `./data/users/` as JSON files.

## What Happens

### 1. Create Users

```go
user := User{ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Age: 30}
mapper.Insert(ctx, "users.user-crud", user)
```

Creates: `data/users/1.json`

### 2. Read Single User

```go
var user User
mapper.Fetch(ctx, "users.user-crud", map[string]interface{}{"id": "1"}, &user)
```

Reads: `data/users/1.json` and maps to `User` struct

### 3. List All Users

```go
var users []map[string]interface{}
mapper.FetchMulti(ctx, "users.user-list", nil, &users)
```

Reads all: `data/users/*.json` files

### 4. Update User

```go
updatedUser := User{ID: "2", Name: "Bob Smith Jr.", Email: "bob.smith@example.com", Age: 26}
mapper.Update(ctx, "users.user-crud", updatedUser)
```

Updates: `data/users/2.json`

### 5. Delete User

```go
mapper.Delete(ctx, "users.user-crud", "3")
```

Removes: `data/users/3.json`

## Key Features Demonstrated

### Configuration-Driven

All operations are defined in `config.yaml` - the application code never mentions file paths or storage details.

### Adapter Pattern

The filesystem adapter can be swapped for MySQL, PostgreSQL, or any other adapter without changing application code.

### Type-Safe

Domain objects (`User` struct) are strongly typed. The mapper handles all conversion between storage format (JSON) and Go structs.

### Zero Dependencies

Your domain objects (`User`) have no dependencies on the mapper library.

## Next Steps

- Try modifying the configuration to add new fields
- Implement custom validations
- Add a different adapter (MySQL, PostgreSQL)
- Explore CQRS patterns with multiple sources
