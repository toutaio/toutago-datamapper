# DataMapper - Configuration-Driven Data Mapping for Go

[![CI](https://github.com/toutaio/toutago-datamapper/actions/workflows/ci.yml/badge.svg)](https://github.com/toutaio/toutago-datamapper/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/toutaio/toutago-datamapper.svg)](https://pkg.go.dev/github.com/toutaio/toutago-datamapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/toutaio/toutago-datamapper)](https://goreportcard.com/report/github.com/toutaio/toutago-datamapper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> A production-ready Go library for configuration-driven data mapping with complete source abstraction. Part of the **Toutā Framework** ecosystem.

## Core Philosophy

- **Zero Dependencies**: Domain objects have no library dependencies
- **Configuration-Driven**: All mappings defined in YAML/JSON
- **Complete Abstraction**: Swap data sources without changing code
- **Pluggable Adapters**: Support any data source through adapters
- **Production Ready**: 80%+ test coverage, comprehensive CI/CD pipeline

## Quick Start

See the [examples directory](examples/) for complete working demonstrations.

```go
package main

import (
    "context"
    "github.com/toutaio/toutago-datamapper/adapter"
    "github.com/toutaio/toutago-datamapper/config"
    "github.com/toutaio/toutago-datamapper/engine"
    "github.com/toutaio/toutago-datamapper/filesystem"
)

type User struct {
    ID    string
    Name  string
    Email string
}

func main() {
    // Create mapper from config file
    mapper, _ := engine.NewMapper("config.yaml")
    defer mapper.Close()

    // Register filesystem adapter
    mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
        return filesystem.NewFilesystemAdapter(source.Connection)
    })

    ctx := context.Background()

    // Create a user
    user := User{ID: "1", Name: "Alice", Email: "alice@example.com"}
    mapper.Insert(ctx, "users.user-crud", user)

    // Fetch the user
    var fetched User
    mapper.Fetch(ctx, "users.user-crud", map[string]interface{}{"id": "1"}, &fetched)
    
    // Update the user
    fetched.Email = "alice.johnson@example.com"
    mapper.Update(ctx, "users.user-crud", fetched)
    
    // Delete the user
    mapper.Delete(ctx, "users.user-crud", "1")
}
```

## Features

### Core Capabilities
- ✅ **YAML/JSON Configuration** - Define all mappings declaratively
- ✅ **CRUD Operations** - Full support for fetch, insert, update, delete
- ✅ **Bulk Operations** - High-performance batch processing
- ✅ **Custom Actions** - Execute stored procedures and complex queries
- ✅ **Multi-Source Support** - Work with multiple data sources simultaneously
- ✅ **Credential Management** - Secure handling via environment variables and files
- ✅ **CQRS Patterns** - Read/write separation, event sourcing, fallback chains
- ✅ **Pluggable Adapters** - Filesystem, MySQL, PostgreSQL, and custom adapters

### Quality & Reliability
- ✅ **80%+ Test Coverage** - Comprehensive unit and integration tests
- ✅ **Production Ready** - Battle-tested CI/CD pipeline
- ✅ **Type Safe** - Leverages Go's type system for compile-time safety
- ✅ **Thread Safe** - Concurrent operation support
- ✅ **Zero Dependencies** - Core library uses only Go standard library

### Available Data Sources
- ✅ **Filesystem** (built-in) - JSON-based storage
- ✅ **MySQL** - Full SQL support via [toutago-datamapper-mysql](https://github.com/toutaio/toutago-datamapper-mysql)
- ✅ **PostgreSQL** - Full SQL support via [toutago-datamapper-postgres](https://github.com/toutaio/toutago-datamapper-postgres)
- 🔄 **Redis, MongoDB, SQLite** - Coming soon

## Quick Example

### With Filesystem Adapter (Built-in)

```yaml
# config/users.yaml
namespace: users
version: "1.0"

sources:
  file-storage:
    adapter: filesystem
    connection: "./data"

mappings:
  user-crud:
    object: User
    source: file-storage
    operations:
      fetch:
        statement: "users/{id}.json"
        result:
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
            - object: Email
              field: email
```

```go
package main

import (
    "context"
    "github.com/toutaio/toutago-datamapper/engine"
    "github.com/toutaio/toutago-datamapper/filesystem"
    "github.com/toutaio/toutago-datamapper/config"
)

type User struct {
    ID    string
    Name  string
    Email string
}

func main() {
    mapper, _ := engine.NewMapper("config/users.yaml")
    defer mapper.Close()

    // Register filesystem adapter
    mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
        return filesystem.NewFilesystemAdapter(source.Connection)
    })

    // Fetch a user
    var user User
    mapper.Fetch(context.Background(), "users.user-crud", map[string]interface{}{
        "id": "123",
    }, &user)
}
```

### With MySQL Adapter

```yaml
# config/users.yaml
namespace: users
version: "1.0"

sources:
  main-db:
    adapter: mysql
    connection: "${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"

mappings:
  user-crud:
    object: User
    source: main-db
    operations:
      fetch:
        statement: "SELECT id, name, email FROM users WHERE id = ?"
        parameters:
          - object: ID
            field: id
        result:
          type: User
          properties:
            - object: ID
              field: id
            - object: Name
              field: name
            - object: Email
              field: email
```

```go
package main

import (
    "context"
    "github.com/toutaio/toutago-datamapper/engine"
    mysql "github.com/toutaio/toutago-datamapper-mysql"
)

type User struct {
    ID    int
    Name  string
    Email string
}

func main() {
    mapper, _ := engine.NewMapper("config/users.yaml")
    defer mapper.Close()

    // Register MySQL adapter
    mapper.RegisterAdapter("mysql", mysql.NewAdapter)

    // Fetch a user - no SQL in your code!
    var user User
    mapper.Fetch(context.Background(), "users.user-crud", map[string]interface{}{
        "ID": 123,
    }, &user)
}
```

## Installation

```bash
go get github.com/toutaio/toutago-datamapper
```

**Requirements:**
- Go 1.22 or higher
- No external dependencies for core library

### Available Adapters

The core library is database-agnostic. Choose adapters based on your needs:

#### Built-in Adapters
- **Filesystem**: JSON-based file storage (included in core library)

#### Official Adapters
- **MySQL**: Production-ready MySQL adapter  
  `go get github.com/toutaio/toutago-datamapper-mysql`  
  [Documentation & Examples](https://github.com/toutaio/toutago-datamapper-mysql)

- **PostgreSQL**: Production-ready PostgreSQL adapter  
  `go get github.com/toutaio/toutago-datamapper-postgres`  
  [Documentation & Examples](https://github.com/toutaio/toutago-datamapper-postgres)

#### Planned Adapters
- Redis
- MongoDB
- SQLite
- REST APIs

See each adapter's repository for specific documentation, examples, and advanced features.

## Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Detailed roadmap
- [Quick Start](QUICKSTART.md) - Getting started guide
- [Configuration Guide](openspec/changes/add-mapper-configuration/SUMMARY.md) - Configuration reference
- [CQRS Patterns](openspec/changes/add-mapper-configuration/CQRS.md) - CQRS implementation guide
- [Credential Management](openspec/changes/add-mapper-configuration/CREDENTIALS.md) - Security best practices
- [Examples](openspec/changes/add-mapper-configuration/EXAMPLES.md) - Configuration examples

## Examples

- [Simple CRUD](examples/simple-crud/) - Basic create, read, update, delete operations
- [Multi-Source CQRS](examples/multi-source/) - Multiple data sources with read/write separation
- [Bulk Operations](examples/bulk-operations/) - High-performance batch processing
- [Credentials Management](examples/credentials/) - Secure credential handling
- [Custom Actions](examples/custom-actions/) - Stored procedures and complex queries

## Architecture

```
┌─────────────────┐
│  Application    │
│  (Your Code)    │
└────────┬────────┘
         │
    ┌────▼─────┐
    │  Mapper  │  ◄─── Configuration (YAML/JSON)
    │  Engine  │
    └────┬─────┘
         │
    ┌────▼────────────────────────┐
    │   Adapter Interface         │
    └─────────────────────────────┘
         │
    ┌────┼────────────┬───────────┐
    │    │            │           │
┌───▼──┐ │  ┌────────▼──┐  ┌────▼────┐
│ File │ │  │  MySQL    │  │ Custom  │
│System│ │  │ Postgres  │  │ Adapter │
└──────┘ │  └───────────┘  └─────────┘
         │
    More adapters...
```

## Project Roadmap

### ✅ Completed (v0.1.0)
- **Foundation**: Project setup, adapter interface, CI/CD pipeline
- **Configuration**: YAML/JSON parser, credential management, CQRS support
- **Core Engine**: Orchestration engine, property mapper, adapter registry
- **Reference Implementation**: Filesystem adapter with complete CRUD
- **Examples & Documentation**: 5 working examples, comprehensive documentation
- **Quality Assurance**: 80%+ test coverage, production-ready CI/CD
- **Database Adapters**: MySQL and PostgreSQL adapters available

### 🔄 In Progress
- **Performance Optimization**: Caching, connection pooling
- **Advanced Features**: Transactions, migrations, schema validation
- **Developer Tooling**: CLI tools, code generation utilities

### 📋 Planned
- Additional database adapters (SQLite)
- NoSQL adapters (Redis, MongoDB)
- API adapters (REST, GraphQL)
- Advanced CQRS features (event sourcing, saga patterns)

See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for the complete roadmap.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup
```bash
git clone https://github.com/toutaio/toutago-datamapper.git
cd toutago-datamapper
go mod download
go test ./...
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests (requires build tag)
go test -tags=integration ./...
```

Please ensure:
- All tests pass
- Code coverage remains above 80%
- Code is formatted with `gofmt`
- No linting errors from `staticcheck`

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright (c) 2024 Toutaio

## Project Information

- **Version**: 0.1.0 (Production Ready)
- **Started**: December 2024
- **Language**: Go 1.22+
- **Test Coverage**: 80.6%
- **Dependencies**: Zero (core library uses stdlib only)
- **Status**: Production Ready

## Support

- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/toutaio/toutago-datamapper)
- **Issues**: [GitHub Issues](https://github.com/toutaio/toutago-datamapper/issues)
- **Examples**: See [examples/](examples/) directory

## Acknowledgments

Built with ❤️ using:
- Go standard library
- YAML parser: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- Modern CI/CD practices
- Comprehensive testing and quality standards

---

## About Toutā Framework

**Toutā** (Proto-Celtic for "people" or "tribe") is a modular Go web framework emphasizing:
- Interface-first design for pluggability
- Configuration-driven architecture
- Dependency injection for testability
- Zero framework lock-in

DataMapper is a core component providing complete data source abstraction.
