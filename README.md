# toutago-datamapper

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
    "github.com/toutago/toutago-datamapper/engine"
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
go get github.com/toutago/toutago-datamapper
```

## Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Detailed roadmap
- [Quick Start](QUICKSTART.md) - Getting started guide
- [Configuration Guide](openspec/changes/add-mapper-configuration/SUMMARY.md) - Configuration reference
- [CQRS Patterns](openspec/changes/add-mapper-configuration/CQRS.md) - CQRS implementation guide
- [Credential Management](openspec/changes/add-mapper-configuration/CREDENTIALS.md) - Security best practices
- [Examples](openspec/changes/add-mapper-configuration/EXAMPLES.md) - Configuration examples

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

### Phase 1: Foundation (Week 1-2) - 🟡 In Progress
- [x] Project setup
- [x] Go module initialized
- [ ] Adapter interface definition
- [ ] Basic tests
- [ ] CI/CD pipeline

### Phase 2: Configuration (Week 2-3) - ⬜ Not Started
- [ ] Configuration parser
- [ ] Credential management
- [ ] CQRS support

### Phase 3: Core Engine (Week 3-4) - ⬜ Not Started
- [ ] Orchestration engine
- [ ] Property mapper
- [ ] Adapter registry

### Phases 4-7 - ⬜ Not Started
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
