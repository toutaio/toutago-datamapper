// Package datamapper provides a production-ready, configuration-driven data mapping library for Go.
//
// Datamapper enables mapping domain objects to ANY data source (SQL, Files, APIs, NoSQL)
// through YAML/JSON configuration, with complete abstraction from data sources and zero
// dependencies on your domain objects.
//
// # Features
//
//   - Zero dependencies on domain objects
//   - Configuration-driven mapping (YAML/JSON)
//   - Complete abstraction from data sources
//   - Pluggable adapter architecture
//   - Built-in filesystem adapter for JSON data
//   - MySQL adapter available (github.com/toutaio/toutago-datamapper-mysql)
//   - PostgreSQL adapter available (github.com/toutaio/toutago-datamapper-postgres)
//   - Secure credential management with environment variables
//   - Custom actions and hooks
//   - Transaction support
//   - Bulk operations
//   - CQRS patterns (read/write separation, event sourcing)
//   - 80%+ test coverage
//   - Production-ready CI/CD pipeline
//
// # Quick Start
//
// Create a mapper from a configuration file:
//
//	mapper, err := engine.NewMapper("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mapper.Close()
//
//	// Register adapters
//	mapper.RegisterAdapter("filesystem", func(source config.Source) (adapter.Adapter, error) {
//	    return filesystem.NewFilesystemAdapter(source.Connection)
//	})
//
//	// Use the mapper
//	user := &User{ID: "1", Name: "John", Email: "john@example.com"}
//	err = mapper.Insert(context.Background(), "User", user)
//
// # Configuration
//
// Define mappings in YAML:
//
//	sources:
//	  - name: main
//	    type: mysql
//	    connection:
//	      host: ${DB_HOST:-localhost}
//	      port: ${DB_PORT:-3306}
//	      database: mydb
//	      user: ${DB_USER}
//	      password: ${DB_PASSWORD}
//
//	mappings:
//	  - entity: User
//	    source: main
//	    table: users
//	    fields:
//	      - entity: ID
//	        column: id
//	        type: string
//	      - entity: Name
//	        column: name
//	        type: string
//
// # Adapters
//
// Built-in adapters:
//
//   - MySQL: Full SQL support with transactions
//   - Filesystem: JSON-based file storage
//
// Custom adapters can be registered using RegisterAdapter.
//
// # Credential Management
//
// Environment variable substitution with defaults:
//
//	${DB_HOST:-localhost}  // Use DB_HOST or default to localhost
//	${DB_PASSWORD}          // Use DB_PASSWORD or empty string
//
// Load credentials from file or environment:
//
//	creds, err := config.LoadCredentials("credentials.yaml")
//	mapper.WithCredentials(creds)
//
// # Thread Safety
//
// The mapper is designed to be used concurrently. Adapters are responsible
// for their own thread-safety guarantees.
//
// # Version
//
// This is version 0.1.0 - production ready with 80%+ test coverage.
// Requires Go 1.22 or higher.
package datamapper
