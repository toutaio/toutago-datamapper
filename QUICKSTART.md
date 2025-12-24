# Quick Start: Phase 1 Implementation

This guide helps you get started with implementing Phase 1 of the toutago-datamapper project.

## Prerequisites

- Go 1.21 or later installed
- Git configured
- Text editor or IDE (VS Code, GoLand, etc.)

## Step 1: Initialize Project (30 minutes)

### 1.1 Create Repository

```bash
# Create project directory
mkdir -p ~/Proyects/toutago-datamapper
cd ~/Proyects/toutago-datamapper

# Initialize Git
git init
git branch -M main

# Create basic structure
mkdir -p {adapter,config,engine,filesystem,examples}
```

### 1.2 Initialize Go Module

```bash
# Initialize module
go mod init github.com/yourusername/toutago-datamapper

# Create go.mod with minimal dependencies
```

### 1.3 Create .gitignore

```bash
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# IDE files
.idea/
.vscode/
*.swp
*.swo
*~

# Credentials (NEVER commit these!)
credentials*.yaml
credentials*.yml
credentials*.json
.env
.env.*
!.env.example

# OS files
.DS_Store
Thumbs.db

# Build artifacts
*.log
EOF
```

### 1.4 Create Initial README

```bash
cat > README.md << 'EOF'
# toutago-datamapper

A Go library for configuration-driven data mapping with complete source abstraction.

## Status

🚧 **Under Development** - Phase 1 in progress

## Vision

Map domain objects to ANY data source (SQL, Files, APIs, NoSQL) through YAML/JSON configuration, with zero dependencies on domain objects.

## Features (Planned)

- ✅ YAML/JSON configuration
- ✅ CRUD operations (fetch, insert, update, delete)
- ✅ Bulk operations
- ✅ Custom actions (stored procedures, complex queries)
- ✅ Multi-source support
- ✅ Credential management (env vars, separate files)
- ✅ CQRS patterns (read/write separation, event sourcing)
- ✅ Pluggable adapters (filesystem, MySQL, PostgreSQL, custom)

## Installation

```bash
go get github.com/yourusername/toutago-datamapper
```

## Quick Example

```go
// Coming soon...
```

## Development

See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for detailed roadmap.

## License

MIT
EOF
```

## Step 2: Define Adapter Interface (2-3 hours)

### 2.1 Create adapter/adapter.go

```go
package adapter

import "context"

// Adapter defines the interface all data source adapters must implement
type Adapter interface {
	// Fetch retrieves one or more objects based on operation and parameters
	Fetch(ctx context.Context, op *Operation, params map[string]interface{}) ([]interface{}, error)

	// Insert creates new objects in the data source
	Insert(ctx context.Context, op *Operation, objects []interface{}) error

	// Update modifies existing objects
	Update(ctx context.Context, op *Operation, objects []interface{}) error

	// Delete removes objects from the data source
	Delete(ctx context.Context, op *Operation, identifiers []interface{}) error

	// Execute runs custom actions (stored procedures, complex queries)
	Execute(ctx context.Context, action *Action, params map[string]interface{}) (interface{}, error)

	// Connect establishes connection to the data source
	Connect(ctx context.Context, config map[string]interface{}) error

	// Close releases all resources
	Close() error

	// Name returns the adapter type name
	Name() string
}

// OperationType represents the type of database operation
type OperationType string

const (
	OpFetch  OperationType = "fetch"
	OpInsert OperationType = "insert"
	OpUpdate OperationType = "update"
	OpDelete OperationType = "delete"
	OpAction OperationType = "action"
)

// Operation represents a configured data operation
type Operation struct {
	Type       OperationType
	Statement  string              // SQL query, file path template, etc.
	Properties []PropertyMapping   // Object property to data field mappings
	Identifier []PropertyMapping   // Fields that identify the object (for update/delete)
	Generated  []PropertyMapping   // Fields that are auto-generated (e.g., auto-increment ID)
	Condition  []PropertyMapping   // Conditional fields (e.g., optimistic locking version)
	Bulk       bool                // Whether this is a bulk operation
	Multi      bool                // Whether to return multiple results (fetch)
	Source     string              // Source name (for CQRS)
	Fallback   *Operation          // Fallback operation (for CQRS)
	After      []AfterAction       // Actions to run after operation (cache invalidation, etc.)
}

// PropertyMapping maps an object property to a data field
type PropertyMapping struct {
	ObjectField string // Name of the field in the Go struct
	DataField   string // Name of the field in the data source
	Type        string // Type conversion hint (timestamp, json, etc.)
	Generated   bool   // Whether this field is auto-generated
}

// Action represents a custom action (stored procedure, complex query)
type Action struct {
	Name       string
	Statement  string
	Parameters []PropertyMapping
	Result     *ResultMapping
}

// ResultMapping defines how to map results back to objects
type ResultMapping struct {
	Type       string            // Go type name
	Multi      bool              // Whether to return multiple results
	Properties []PropertyMapping // Field mappings
}

// AfterAction represents an action to execute after an operation
type AfterAction struct {
	Type      string                 // invalidate, cache_set, etc.
	Source    string                 // Source to execute on
	Statement string                 // Statement to execute
	Config    map[string]interface{} // Additional configuration
}

// Error types
var (
	ErrNotFound      = &AdapterError{Code: "NOT_FOUND", Message: "object not found"}
	ErrValidation    = &AdapterError{Code: "VALIDATION", Message: "validation failed"}
	ErrConnection    = &AdapterError{Code: "CONNECTION", Message: "connection failed"}
	ErrAdapter       = &AdapterError{Code: "ADAPTER", Message: "adapter error"}
	ErrConfiguration = &AdapterError{Code: "CONFIGURATION", Message: "configuration error"}
)

// AdapterError represents an error from an adapter
type AdapterError struct {
	Code    string
	Message string
	Cause   error
}

func (e *AdapterError) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *AdapterError) Unwrap() error {
	return e.Cause
}
```

### 2.2 Create adapter/adapter_test.go

```go
package adapter

import (
	"testing"
)

func TestOperationType(t *testing.T) {
	tests := []struct {
		name string
		op   OperationType
		want string
	}{
		{"fetch", OpFetch, "fetch"},
		{"insert", OpInsert, "insert"},
		{"update", OpUpdate, "update"},
		{"delete", OpDelete, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("got %v, want %v", tt.op, tt.want)
			}
		})
	}
}

func TestAdapterError(t *testing.T) {
	err := &AdapterError{
		Code:    "TEST_ERROR",
		Message: "test message",
	}

	if err.Error() != "TEST_ERROR: test message" {
		t.Errorf("unexpected error string: %v", err.Error())
	}
}
```

## Step 3: Create Basic Tests (1 hour)

### 3.1 Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Step 4: Set Up CI/CD (1 hour)

### 4.1 Create .github/workflows/ci.yml

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      run: go build -v ./...
    
    - name: Test
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Coverage
      run: go tool cover -func=coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

## Step 5: Commit and Push

```bash
# Add all files
git add .

# Commit
git commit -m "feat: initial project setup and adapter interface"

# Create GitHub repo (via GitHub CLI or web)
gh repo create toutago-datamapper --public --source=. --push

# Or push to existing repo
git remote add origin https://github.com/yourusername/toutago-datamapper.git
git push -u origin main
```

## Next Steps

After completing Phase 1.1 and 1.2:

1. **Review** the adapter interface with team
2. **Start Milestone 2.1**: Configuration schema design
3. **Update** IMPLEMENTATION_PLAN.md with progress
4. **Track** progress in GitHub Projects or Issues

## Validation Checklist

Phase 1 Complete:
- [ ] Go module initialized
- [ ] Project structure created
- [ ] .gitignore configured
- [ ] README created
- [ ] Adapter interface defined with godoc
- [ ] Basic tests passing
- [ ] CI/CD pipeline green
- [ ] Code pushed to GitHub

## Getting Help

- Check IMPLEMENTATION_PLAN.md for detailed guidance
- Review design.md for architecture decisions
- See EXAMPLES.md for configuration patterns
- Consult CQRS.md for CQRS-specific guidance

## Common Issues

### Import Paths Not Working
Make sure go.mod has correct module path:
```
module github.com/yourusername/toutago-datamapper
```

### Tests Failing
Ensure you're in the right directory:
```bash
cd ~/Proyects/toutago-datamapper
go test ./...
```

### CI/CD Not Running
Check .github/workflows/ci.yml is committed and pushed

---

Good luck with Phase 1! 🚀
