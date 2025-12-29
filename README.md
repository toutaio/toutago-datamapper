# toutago-datamapper

[![CI](https://github.com/toutaio/toutago-datamapper/actions/workflows/ci.yml/badge.svg)](https://github.com/toutaio/toutago-datamapper/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/toutaio/toutago-datamapper.svg)](https://pkg.go.dev/github.com/toutaio/toutago-datamapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/toutaio/toutago-datamapper)](https://goreportcard.com/report/github.com/toutaio/toutago-datamapper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for configuration-driven data mapping with complete source abstraction.

## Status

🚧 **Under Development** - Phase 1 in progress

## Vision

Map domain objects to ANY data source (SQL, Files, APIs, NoSQL) through YAML/JSON configuration, with zero dependencies on domain objects.

## Core Philosophy

- **Zero Dependencies**: Domain objects have no library dependencies
- **Configuration-Driven**: All mappings defined in YAML/JSON
- **Complete Abstraction**: Swap data sources without changing code
- **Pluggable Adapters**: Support any data source through adapters

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

- ✅ YAML/JSON configuration
- ✅ CRUD operations (fetch, insert, update, delete)
- ✅ Bulk operations
- ✅ Custom actions (stored procedures, complex queries)
- ✅ Multi-source support
- ✅ Credential management (env vars, separate files)
- ✅ CQRS patterns (read/write separation, event sourcing)
- ✅ Pluggable adapters (filesystem, MySQL, PostgreSQL, custom)

## Quick Example

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
)

type User struct {
    ID    int
    Name  string
    Email string
}

func main() {
    // Load mapper from configuration
    mapper, _ := engine.NewMapper("config/users.yaml")
    defer mapper.Close()

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

### Available Adapters

The core library is database-agnostic. Install adapters separately based on your needs:

#### Official Adapters

- **MySQL**: `go get github.com/toutaio/toutago-datamapper-mysql`
- **Filesystem**: Built-in (included in core library)

#### Coming Soon

- PostgreSQL
- Redis
- MongoDB
- REST APIs

See each adapter's repository for specific documentation and examples.

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

## Development Status

### Phase 1: Foundation (Week 1-2) - ✅ COMPLETE
- [x] Project setup
- [x] Go module initialized
- [x] Adapter interface definition
- [x] Basic tests (100% coverage)
- [x] CI/CD pipeline

### Phase 2: Configuration (Week 2-3) - ✅ COMPLETE
- [x] Configuration parser (YAML/JSON)
- [x] Credential management (env vars, credentials files)
- [x] CQRS support (fallback chains, after actions)
- [x] Multi-file loading
- [x] Comprehensive tests (75%+ coverage)

### Phase 3: Core Engine (Week 3-4) - ✅ COMPLETE
- [x] Orchestration engine
- [x] Property mapper (reflection-based)
- [x] Adapter registry
- [x] Comprehensive tests (55%+ coverage)

### Phase 4: Reference Impl (Week 4-5) - ✅ COMPLETE
- [x] Filesystem adapter (76% coverage)
- [x] Complete CRUD operations
- [x] Working examples
- [x] Example documentation

### Phase 5: Examples & Documentation (Week 5-6) - ✅ COMPLETE
- [x] Simple CRUD example
- [x] Multi-source CQRS example
- [x] Bulk operations example
- [x] Credentials management example
- [x] Custom actions example
- [x] Comprehensive README files
- [x] Usage patterns documentation

### Phases 6-7 - 🔄 READY TO START
See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for complete roadmap.

## Contributing

This project is currently in initial development. Contributions will be welcome after v0.1.0 release.

## License

MIT License

## Project Information

- **Started**: December 2024
- **Target v0.1.0**: 8 weeks from start
- **Language**: Go 1.21+
- **Dependencies**: Zero (core library uses stdlib only)

## Related Documentation

- [OpenSpec](openspec/) - Complete specification
- [Project Status](PROJECT_STATUS.md) - Current metrics and progress
