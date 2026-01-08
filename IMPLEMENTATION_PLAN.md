# Implementation Plan: toutago-datamapper

> **📜 HISTORICAL DOCUMENT**: This was the original implementation plan. The project is now complete and in production. See [CHANGELOG.md](CHANGELOG.md) for actual release history and [Releases](https://github.com/toutaio/toutago-datamapper/releases) for current versions.

## Project Overview

**Purpose**: A Go library that provides complete data source abstraction through configuration-driven mapping, allowing applications to perform CRUD operations on domain objects without any knowledge of the underlying storage mechanism.

**Core Philosophy**: Zero dependencies on domain objects, pluggable adapters, configuration-driven mapping.

---

## Architecture Overview

```
toutago-datamapper/
├── Core Library (stdlib only)
│   ├── Configuration Parser (YAML/JSON)
│   ├── Credential Manager (env vars, files)
│   ├── Adapter Interface
│   ├── Orchestration Engine (CQRS support)
│   └── Reference Adapter (filesystem)
│
├── Separate Adapter Packages
│   ├── toutago-datamapper-mysql
│   ├── toutago-datamapper-postgres
│   └── toutago-datamapper-* (user-defined)
│
└── Examples & Documentation
    ├── Configuration examples
    ├── Credential management guides
    └── CQRS pattern guides
```

---

## Phase 1: Foundation & Core Interfaces (Week 1-2)

### Milestone 1.1: Project Setup
**Duration**: 1-2 days

**Tasks**:
- [ ] Initialize Go module (`go mod init github.com/yourusername/toutago-datamapper`)
- [ ] Set up project structure (packages: config, adapter, engine, filesystem)
- [ ] Configure CI/CD (GitHub Actions: lint, test, build)
- [ ] Add .gitignore (credentials files, IDE files)
- [ ] Create initial README with project vision

**Deliverables**:
- Working Go module
- CI/CD pipeline
- Project structure documented

---

### Milestone 1.2: Adapter Interface Definition
**Duration**: 2-3 days

**Tasks**:
- [ ] Define `Adapter` interface (Fetch, Insert, Update, Delete, Execute)
- [ ] Define `Source` interface/struct (connection details, adapter type)
- [ ] Define `Operation` types (Fetch, Insert, Update, Delete, Action)
- [ ] Define `PropertyMapping` struct
- [ ] Define error types (`ErrNotFound`, `ErrValidation`, `ErrAdapter`)
- [ ] Add comprehensive godoc comments

**Deliverables**:
```go
// adapter/adapter.go
package adapter

type Adapter interface {
    // Fetch retrieves one or more objects
    Fetch(ctx context.Context, op *Operation, params map[string]interface{}) ([]interface{}, error)
    
    // Insert creates new objects
    Insert(ctx context.Context, op *Operation, objects []interface{}) error
    
    // Update modifies existing objects
    Update(ctx context.Context, op *Operation, objects []interface{}) error
    
    // Delete removes objects
    Delete(ctx context.Context, op *Operation, identifiers []interface{}) error
    
    // Execute runs custom actions
    Execute(ctx context.Context, action *Action, params map[string]interface{}) (interface{}, error)
    
    // Connect establishes connection to data source
    Connect(ctx context.Context, config map[string]interface{}) error
    
    // Close releases resources
    Close() error
}

type Operation struct {
    Type       OperationType
    Statement  string
    Properties []PropertyMapping
    Identifier []PropertyMapping
    Generated  []PropertyMapping
    Bulk       bool
    Multi      bool
}

type PropertyMapping struct {
    ObjectField string
    DataField   string
    Type        string  // timestamp, json, etc.
}
```

---

## Phase 2: Configuration System (Week 2-3)

### Milestone 2.1: Configuration Schema Design
**Duration**: 2-3 days

**Tasks**:
- [ ] Define configuration structs (Config, Source, Mapping, Operation)
- [ ] Implement struct tags for YAML/JSON parsing
- [ ] Define validation rules
- [ ] Create JSON schema file for external validation

**Deliverables**:
```go
// config/schema.go
package config

type Config struct {
    Namespace string              `yaml:"namespace" json:"namespace"`
    Version   string              `yaml:"version" json:"version"`
    Sources   map[string]Source   `yaml:"sources" json:"sources"`
    Mappings  map[string]Mapping  `yaml:"mappings" json:"mappings"`
}

type Source struct {
    Adapter    string                 `yaml:"adapter" json:"adapter"`
    Connection string                 `yaml:"connection" json:"connection"`
    Options    map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

type Mapping struct {
    Object     string                    `yaml:"object" json:"object"`
    Source     string                    `yaml:"source,omitempty" json:"source,omitempty"`
    Operations map[string]OperationConfig `yaml:"operations,omitempty" json:"operations,omitempty"`
    Actions    map[string]ActionConfig    `yaml:"actions,omitempty" json:"actions,omitempty"`
}
```

---

### Milestone 2.2: Configuration Parser
**Duration**: 3-4 days

**Tasks**:
- [ ] Implement YAML parser (use `gopkg.in/yaml.v3`)
- [ ] Implement JSON parser (use stdlib `encoding/json`)
- [ ] Add configuration validation logic
- [ ] Implement namespace collision detection
- [ ] Add detailed error reporting with file/line numbers

**Deliverables**:
```go
// config/parser.go
package config

type Parser struct {
    configs map[string]*Config  // namespace -> config
}

func NewParser() *Parser

// LoadFile loads a single configuration file
func (p *Parser) LoadFile(path string) error

// LoadDirectory loads all config files from a directory
func (p *Parser) LoadDirectory(path string) error

// Validate checks all loaded configurations
func (p *Parser) Validate() error

// GetConfig returns config by namespace
func (p *Parser) GetConfig(namespace string) (*Config, error)

// GetMapping returns a specific mapping
func (p *Parser) GetMapping(fullyQualifiedID string) (*Mapping, error)
```

**Test Coverage**: 90%+
- Valid YAML/JSON parsing
- Invalid format detection
- Namespace collisions
- Missing required fields
- Multi-file loading

---

### Milestone 2.3: Credential Management
**Duration**: 3-4 days

**Tasks**:
- [ ] Implement environment variable resolver
- [ ] Support placeholder syntax (`${VAR_NAME}`)
- [ ] Support default values (`${VAR_NAME:-default}`)
- [ ] Implement credentials file loader
- [ ] Support credential references (`@credentials:name`)
- [ ] Add credential exposure warnings
- [ ] Implement log sanitization

**Deliverables**:
```go
// config/credentials.go
package config

type CredentialResolver struct {
    envVars     map[string]string
    credentials map[string]interface{}
}

func NewCredentialResolver() *CredentialResolver

// LoadEnvFile loads .env file
func (cr *CredentialResolver) LoadEnvFile(path string) error

// LoadCredentialsFile loads credentials.yaml
func (cr *CredentialResolver) LoadCredentialsFile(path string) error

// Resolve replaces placeholders in a string
func (cr *CredentialResolver) Resolve(value string) (string, error)

// Sanitize removes credentials from error messages
func (cr *CredentialResolver) Sanitize(message string) string
```

**Test Coverage**: 95%+
- Environment variable resolution
- Default values
- Credentials file loading
- Missing variable errors
- Sanitization

---

### Milestone 2.4: CQRS Support
**Duration**: 2-3 days

**Tasks**:
- [ ] Implement operation-specific source resolution
- [ ] Add source fallback chain logic
- [ ] Support cache invalidation hooks
- [ ] Add "after" action execution

**Deliverables**:
```go
// config/cqrs.go
package config

type SourceResolver struct {
    sources map[string]Source
}

// ResolveSource returns the appropriate source for an operation
func (sr *SourceResolver) ResolveSource(mapping *Mapping, opType string) (Source, error)

// ResolveFallbackChain returns ordered list of sources to try
func (sr *SourceResolver) ResolveFallbackChain(op *OperationConfig) ([]Source, error)
```

---

## Phase 3: Core Engine (Week 3-4)

### Milestone 3.1: Orchestration Engine
**Duration**: 4-5 days

**Tasks**:
- [ ] Implement Mapper orchestration engine
- [ ] Add adapter registry
- [ ] Implement operation execution logic
- [ ] Add reflection-based property mapping
- [ ] Implement bulk operation handling

**Deliverables**:
```go
// engine/mapper.go
package engine

type Mapper struct {
    config   *config.Config
    registry *AdapterRegistry
    resolver *config.CredentialResolver
}

func NewMapper(configPath string) (*Mapper, error)

// Fetch retrieves objects
func (m *Mapper) Fetch(ctx context.Context, mappingID string, params map[string]interface{}) (interface{}, error)

// FetchMulti retrieves multiple objects
func (m *Mapper) FetchMulti(ctx context.Context, mappingID string, params map[string]interface{}) ([]interface{}, error)

// Insert creates objects
func (m *Mapper) Insert(ctx context.Context, mappingID string, objects interface{}) error

// Update modifies objects
func (m *Mapper) Update(ctx context.Context, mappingID string, objects interface{}) error

// Delete removes objects
func (m *Mapper) Delete(ctx context.Context, mappingID string, identifiers interface{}) error

// Execute runs custom actions
func (m *Mapper) Execute(ctx context.Context, actionID string, params map[string]interface{}) (interface{}, error)
```

---

### Milestone 3.2: Property Mapper
**Duration**: 3-4 days

**Tasks**:
- [ ] Implement reflection-based property mapping
- [ ] Support nested struct navigation
- [ ] Add type conversion (timestamp, JSON, etc.)
- [ ] Handle pointer vs value semantics
- [ ] Add validation for required fields

**Deliverables**:
```go
// engine/property.go
package engine

type PropertyMapper struct{}

func NewPropertyMapper() *PropertyMapper

// MapToObject maps data fields to object properties
func (pm *PropertyMapper) MapToObject(data map[string]interface{}, target interface{}, mappings []config.PropertyMapping) error

// MapFromObject extracts data fields from object
func (pm *PropertyMapper) MapFromObject(obj interface{}, mappings []config.PropertyMapping) (map[string]interface{}, error)

// ConvertType converts between types based on hints
func (pm *PropertyMapper) ConvertType(value interface{}, typeHint string) (interface{}, error)
```

---

### Milestone 3.3: Adapter Registry
**Duration**: 2-3 days

**Tasks**:
- [ ] Implement adapter registration system
- [ ] Support runtime adapter loading
- [ ] Add adapter lifecycle management
- [ ] Implement connection pooling

**Deliverables**:
```go
// engine/registry.go
package engine

type AdapterRegistry struct {
    adapters map[string]AdapterFactory
    instances map[string]adapter.Adapter
}

type AdapterFactory func(config map[string]interface{}) (adapter.Adapter, error)

func NewAdapterRegistry() *AdapterRegistry

// Register adds an adapter factory
func (ar *AdapterRegistry) Register(adapterType string, factory AdapterFactory)

// GetAdapter returns or creates an adapter instance
func (ar *AdapterRegistry) GetAdapter(ctx context.Context, source config.Source) (adapter.Adapter, error)

// Close closes all adapter instances
func (ar *AdapterRegistry) Close() error
```

---

## Phase 4: Reference Implementation (Week 4-5)

### Milestone 4.1: Filesystem Adapter
**Duration**: 4-5 days

**Tasks**:
- [ ] Implement filesystem adapter (JSON files)
- [ ] Support path templates (`{id}.json`)
- [ ] Add glob pattern support for listing
- [ ] Implement atomic file writes
- [ ] Add error handling (file not found, permissions)

**Deliverables**:
```go
// filesystem/adapter.go
package filesystem

type FilesystemAdapter struct {
    basePath string
    format   string  // json, yaml, xml
}

func NewFilesystemAdapter(config map[string]interface{}) (*FilesystemAdapter, error)

// Implements adapter.Adapter interface
func (fa *FilesystemAdapter) Fetch(ctx context.Context, op *adapter.Operation, params map[string]interface{}) ([]interface{}, error)
func (fa *FilesystemAdapter) Insert(ctx context.Context, op *adapter.Operation, objects []interface{}) error
func (fa *FilesystemAdapter) Update(ctx context.Context, op *adapter.Operation, objects []interface{}) error
func (fa *FilesystemAdapter) Delete(ctx context.Context, op *adapter.Operation, identifiers []interface{}) error
func (fa *FilesystemAdapter) Execute(ctx context.Context, action *adapter.Action, params map[string]interface{}) (interface{}, error)
```

**Test Coverage**: 85%+
- File read/write operations
- Path template resolution
- Glob pattern matching
- Error handling

---

## Phase 5: Examples & Documentation (Week 5-6)

### Milestone 5.1: Working Examples
**Duration**: 3-4 days

**Tasks**:
- [ ] Create examples/simple-crud (basic User CRUD with filesystem)
- [ ] Create examples/multi-source (database + cache + files)
- [ ] Create examples/cqrs (read replica + master)
- [ ] Create examples/credentials (env vars + credentials file)
- [ ] Create examples/bulk-operations
- [ ] Create examples/custom-actions

---

### Milestone 5.2: Documentation
**Duration**: 3-4 days

**Tasks**:
- [ ] Write comprehensive README
- [ ] Create Getting Started guide
- [ ] Document configuration format
- [ ] Add credential management guide
- [ ] Add CQRS patterns guide
- [ ] Create API documentation (godoc)
- [ ] Add migration guide

---

## Phase 6: External Adapters (Week 6-8)

### Milestone 6.1: MySQL Adapter (Separate Module)
**Duration**: 4-5 days

**Tasks**:
- [ ] Create `toutago-datamapper-mysql` repository
- [ ] Implement MySQL adapter using `database/sql`
- [ ] Support prepared statements
- [ ] Handle generated IDs (auto-increment)
- [ ] Add transaction support
- [ ] Comprehensive tests with test containers

---

### Milestone 6.2: PostgreSQL Adapter (Separate Module)
**Duration**: 4-5 days

**Tasks**:
- [ ] Create `toutago-datamapper-postgres` repository
- [ ] Implement PostgreSQL adapter
- [ ] Support RETURNING clause for generated IDs
- [ ] Add COPY support for bulk operations
- [ ] Comprehensive tests

---

## Phase 7: Testing & Refinement (Week 7-8)

### Milestone 7.1: Integration Testing
**Duration**: 3-4 days

**Tasks**:
- [ ] Integration tests with real adapters
- [ ] End-to-end tests with multiple sources
- [ ] Performance benchmarks
- [ ] Memory profiling
- [ ] Concurrency testing

---

### Milestone 7.2: Polish & Release Prep
**Duration**: 2-3 days

**Tasks**:
- [ ] Code review and refactoring
- [ ] Performance optimization
- [ ] Documentation review
- [ ] CHANGELOG creation
- [ ] Prepare v0.1.0 release

---

## Testing Strategy

### Unit Tests (Per Package)
- **Target**: 85%+ coverage
- **Tools**: Standard Go testing, testify for assertions
- **Focus**: Edge cases, error handling, validation

### Integration Tests
- **Target**: 70%+ coverage
- **Tools**: Testcontainers for databases
- **Focus**: Multi-adapter scenarios, CQRS patterns

### Benchmarks
- **Metrics**: Operations/second, memory allocation
- **Scenarios**: Single operations, bulk operations, caching

---

## Success Criteria

### Phase 1-2 (Foundation)
- ✅ Core interfaces defined and documented
- ✅ Configuration parser handles YAML/JSON
- ✅ Credential management working
- ✅ 90%+ test coverage

### Phase 3-4 (Core + Reference)
- ✅ Orchestration engine working
- ✅ Filesystem adapter fully functional
- ✅ Examples running successfully
- ✅ 85%+ test coverage

### Phase 5-6 (Documentation + Adapters)
- ✅ Complete documentation
- ✅ MySQL/PostgreSQL adapters working
- ✅ All examples validated

### Phase 7 (Release)
- ✅ v0.1.0 released
- ✅ CI/CD pipeline green
- ✅ Documentation published

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1: Foundation | Week 1-2 | Adapter interfaces, project structure |
| Phase 2: Configuration | Week 2-3 | Parser, credentials, CQRS support |
| Phase 3: Core Engine | Week 3-4 | Orchestration, property mapping, registry |
| Phase 4: Reference Impl | Week 4-5 | Filesystem adapter |
| Phase 5: Examples/Docs | Week 5-6 | Working examples, documentation |
| Phase 6: External Adapters | Week 6-8 | MySQL, PostgreSQL adapters |
| Phase 7: Testing/Release | Week 7-8 | Integration tests, v0.1.0 release |

**Total Duration**: 8 weeks (2 months)

---

## Risks & Mitigation

### Risk: Reflection Performance
**Impact**: Medium
**Mitigation**: Benchmark early, cache reflection metadata, consider code generation for hot paths

### Risk: Configuration Complexity
**Impact**: High
**Mitigation**: Comprehensive examples, validation with clear errors, JSON schema support

### Risk: Adapter Interface Too Rigid
**Impact**: Medium
**Mitigation**: Design for extensibility, gather feedback early, version the interface

### Risk: CQRS Complexity
**Impact**: Medium
**Mitigation**: Start with simple patterns, document thoroughly, provide examples

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Approve the spec** (`add-mapper-configuration`)
3. **Create GitHub repository** and initialize Go module
4. **Start Phase 1** with adapter interface definition
5. **Set up CI/CD** pipeline early
6. **Create milestone tracking** in GitHub Projects

---

## Resources Needed

- **Go 1.21+** development environment
- **Docker** for test containers (MySQL, PostgreSQL, Redis)
- **GitHub** for repository hosting and CI/CD
- **Documentation hosting** (GitHub Pages or similar)

---

## Future Enhancements (Post v1.0)

- Configuration generation from database schemas
- GraphQL adapter
- MongoDB adapter
- Redis adapter
- Excel adapter
- Migration tooling
- Performance monitoring and metrics
- Configuration hot-reload
- Web-based configuration editor
