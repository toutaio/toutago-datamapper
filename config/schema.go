// Package config provides configuration parsing and management for the toutago-datamapper.
// It supports YAML and JSON configuration formats with environment variable substitution
// and credential management.
package config

// Config represents the root configuration for a mapper.
// Each configuration file should have a unique namespace to avoid collisions.
type Config struct {
	// Namespace uniquely identifies this configuration file.
	// Used to prevent mapping ID collisions across multiple files.
	Namespace string `yaml:"namespace" json:"namespace"`

	// Version specifies the configuration format version.
	// Current version is "1.0".
	Version string `yaml:"version" json:"version"`

	// Sources defines named data sources (databases, files, APIs, etc.).
	Sources map[string]Source `yaml:"sources,omitempty" json:"sources,omitempty"`

	// Mappings defines object-to-data-source mappings.
	Mappings map[string]Mapping `yaml:"mappings" json:"mappings"`
}

// Source defines a data source connection configuration.
type Source struct {
	// Adapter specifies the adapter type (mysql, postgres, filesystem, redis, etc.).
	Adapter string `yaml:"adapter" json:"adapter"`

	// Connection is the connection string or reference.
	// Can contain environment variable placeholders: ${VAR_NAME}
	// Can reference credentials file: @credentials:source-name
	Connection string `yaml:"connection" json:"connection"`

	// Options contains adapter-specific configuration options.
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// Mapping defines how a domain object maps to data operations.
type Mapping struct {
	// Object is the Go type name (e.g., "User", "Order").
	Object string `yaml:"object" json:"object"`

	// Source is the default source name for all operations.
	// Individual operations can override this.
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// Operations defines CRUD operations for this object.
	Operations map[string]OperationConfig `yaml:"operations,omitempty" json:"operations,omitempty"`

	// Actions defines custom actions (stored procedures, complex queries).
	Actions map[string]ActionConfig `yaml:"actions,omitempty" json:"actions,omitempty"`
}

// OperationConfig defines configuration for a single operation (fetch, insert, update, delete).
type OperationConfig struct {
	// Source overrides the default source for this operation (CQRS pattern).
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// Sources defines a fallback chain of sources to try (CQRS pattern).
	Sources []SourceRef `yaml:"sources,omitempty" json:"sources,omitempty"`

	// Statement is the adapter-specific statement (SQL query, file path, etc.).
	Statement string `yaml:"statement" json:"statement"`

	// Parameters defines input parameter mappings.
	Parameters []PropertyMap `yaml:"parameters,omitempty" json:"parameters,omitempty"`

	// Properties defines object property to data field mappings.
	Properties []PropertyMap `yaml:"properties,omitempty" json:"properties,omitempty"`

	// Identifier defines fields that identify the object (for update/delete).
	Identifier []PropertyMap `yaml:"identifier,omitempty" json:"identifier,omitempty"`

	// Generated defines auto-generated fields (auto-increment IDs, timestamps).
	Generated []PropertyMap `yaml:"generated,omitempty" json:"generated,omitempty"`

	// Condition defines conditional fields (optimistic locking, version checks).
	Condition []PropertyMap `yaml:"condition,omitempty" json:"condition,omitempty"`

	// Result defines how to map results back to objects.
	Result *ResultConfig `yaml:"result,omitempty" json:"result,omitempty"`

	// Bulk indicates this is a bulk operation (multiple objects).
	Bulk bool `yaml:"bulk,omitempty" json:"bulk,omitempty"`

	// Fallback defines an alternative operation if this one fails.
	Fallback *OperationConfig `yaml:"fallback,omitempty" json:"fallback,omitempty"`

	// After defines actions to run after the operation (cache invalidation, etc.).
	After []AfterActionConfig `yaml:"after,omitempty" json:"after,omitempty"`
}

// SourceRef references a source with fallback behavior (for CQRS).
type SourceRef struct {
	// Name is the source name.
	Name string `yaml:"name" json:"name"`

	// OnMiss specifies what to do on cache miss ("next" to try next source).
	OnMiss string `yaml:"on_miss,omitempty" json:"on_miss,omitempty"`

	// OnError specifies what to do on error ("next" to try next source).
	OnError string `yaml:"on_error,omitempty" json:"on_error,omitempty"`
}

// PropertyMap maps an object property to a data field.
type PropertyMap struct {
	// Object is the object field name (in Go struct).
	Object string `yaml:"object" json:"object"`

	// Field is the data field name (in database, file, etc.).
	Field string `yaml:"field" json:"field"`

	// Type is an optional type conversion hint (timestamp, json, base64, etc.).
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Generated indicates this field is auto-generated.
	Generated bool `yaml:"generated,omitempty" json:"generated,omitempty"`
}

// ResultConfig defines how to map operation results to objects.
type ResultConfig struct {
	// Type is the Go type name to create.
	Type string `yaml:"type" json:"type"`

	// Multi indicates whether to return multiple results.
	Multi bool `yaml:"multi,omitempty" json:"multi,omitempty"`

	// Properties maps data fields to object properties.
	Properties []PropertyMap `yaml:"properties" json:"properties"`
}

// ActionConfig defines a custom action configuration.
type ActionConfig struct {
	// Source specifies which source to execute the action on.
	Source string `yaml:"source,omitempty" json:"source,omitempty"`

	// Statement is the adapter-specific statement to execute.
	Statement string `yaml:"statement" json:"statement"`

	// Parameters defines input parameter mappings.
	Parameters []PropertyMap `yaml:"parameters,omitempty" json:"parameters,omitempty"`

	// Result defines how to map action results.
	Result *ResultConfig `yaml:"result,omitempty" json:"result,omitempty"`
}

// AfterActionConfig defines an action to execute after an operation.
type AfterActionConfig struct {
	// Action is the action type (invalidate, cache_set, publish, etc.).
	Action string `yaml:"action" json:"action"`

	// Source is the source to execute the action on.
	Source string `yaml:"source" json:"source"`

	// Statement is the adapter-specific statement.
	Statement string `yaml:"statement,omitempty" json:"statement,omitempty"`

	// Config contains additional configuration.
	Config map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// CredentialsConfig represents a credentials file structure.
type CredentialsConfig struct {
	// Credentials maps source names to their connection details.
	Credentials map[string]CredentialSource `yaml:"credentials" json:"credentials"`
}

// CredentialSource contains connection details for a source.
type CredentialSource struct {
	// Connection is the actual connection string.
	Connection string `yaml:"connection" json:"connection"`

	// Options contains adapter-specific options.
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}
