package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles loading and parsing configuration files.
type Parser struct {
	// configs stores loaded configurations by namespace
	configs map[string]*Config

	// credentials resolver for environment variables and credentials files
	credResolver *CredentialResolver
}

// NewParser creates a new configuration parser.
func NewParser() *Parser {
	return &Parser{
		configs:      make(map[string]*Config),
		credResolver: NewCredentialResolver(),
	}
}

// LoadFile loads a single configuration file (YAML or JSON).
// The file extension determines the format (.yaml, .yml, .json).
func (p *Parser) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var cfg Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse JSON file %s: %w", path, err)
		}
	default:
		return fmt.Errorf("unsupported file extension %s (use .yaml, .yml, or .json)", ext)
	}

	// Validate basic structure
	if err := p.validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid configuration in %s: %w", path, err)
	}

	// Check for namespace collision
	if existing, exists := p.configs[cfg.Namespace]; exists {
		return fmt.Errorf("namespace collision: namespace '%s' already loaded from another file (existing version: %s)",
			cfg.Namespace, existing.Version)
	}

	// Resolve credentials in connection strings
	if err := p.resolveCredentials(&cfg); err != nil {
		return fmt.Errorf("failed to resolve credentials in %s: %w", path, err)
	}

	p.configs[cfg.Namespace] = &cfg
	return nil
}

// LoadDirectory loads all configuration files from a directory.
// Supports .yaml, .yml, and .json files.
func (p *Parser) LoadDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	loadedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Skip non-config files
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		// Skip credentials files
		if strings.Contains(strings.ToLower(filename), "credential") {
			continue
		}

		fullPath := filepath.Join(path, filename)
		if err := p.LoadFile(fullPath); err != nil {
			return fmt.Errorf("failed to load %s: %w", fullPath, err)
		}
		loadedCount++
	}

	if loadedCount == 0 {
		return fmt.Errorf("no configuration files found in %s", path)
	}

	return nil
}

// LoadCredentialsFile loads a credentials file.
func (p *Parser) LoadCredentialsFile(path string) error {
	return p.credResolver.LoadCredentialsFile(path)
}

// LoadEnvFile loads environment variables from a .env file.
func (p *Parser) LoadEnvFile(path string) error {
	return p.credResolver.LoadEnvFile(path)
}

// Validate checks all loaded configurations for errors.
func (p *Parser) Validate() error {
	if len(p.configs) == 0 {
		return fmt.Errorf("no configurations loaded")
	}

	for namespace, cfg := range p.configs {
		if err := p.validateConfig(cfg); err != nil {
			return fmt.Errorf("validation error in namespace '%s': %w", namespace, err)
		}

		// Validate source references in mappings
		if err := p.validateSourceReferences(cfg); err != nil {
			return fmt.Errorf("validation error in namespace '%s': %w", namespace, err)
		}
	}

	return nil
}

// GetConfig returns a configuration by namespace.
func (p *Parser) GetConfig(namespace string) (*Config, error) {
	cfg, exists := p.configs[namespace]
	if !exists {
		return nil, fmt.Errorf("configuration namespace '%s' not found", namespace)
	}
	return cfg, nil
}

// GetMapping returns a specific mapping by fully-qualified ID (namespace.mappingID).
func (p *Parser) GetMapping(fullyQualifiedID string) (*Mapping, *Config, error) {
	parts := strings.Split(fullyQualifiedID, ".")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid mapping ID '%s': must be in format 'namespace.mappingID'", fullyQualifiedID)
	}

	namespace := parts[0]
	mappingID := parts[1]

	cfg, err := p.GetConfig(namespace)
	if err != nil {
		return nil, nil, err
	}

	mapping, exists := cfg.Mappings[mappingID]
	if !exists {
		return nil, nil, fmt.Errorf("mapping '%s' not found in namespace '%s'", mappingID, namespace)
	}

	return &mapping, cfg, nil
}

// GetAllNamespaces returns all loaded namespace names.
func (p *Parser) GetAllNamespaces() []string {
	namespaces := make([]string, 0, len(p.configs))
	for ns := range p.configs {
		namespaces = append(namespaces, ns)
	}
	return namespaces
}

// validateConfig performs basic validation on a configuration.
func (p *Parser) validateConfig(cfg *Config) error {
	if cfg.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if cfg.Version != "1.0" {
		return fmt.Errorf("unsupported version '%s' (supported: 1.0)", cfg.Version)
	}

	if len(cfg.Mappings) == 0 {
		return fmt.Errorf("at least one mapping is required")
	}

	// Validate each mapping
	for mappingID, mapping := range cfg.Mappings {
		if mapping.Object == "" {
			return fmt.Errorf("mapping '%s': object type is required", mappingID)
		}

		// Must have either default source or operation-specific sources
		hasDefaultSource := mapping.Source != ""
		hasOperations := len(mapping.Operations) > 0 || len(mapping.Actions) > 0

		if !hasDefaultSource && !hasOperations {
			return fmt.Errorf("mapping '%s': must have either a default source or operations/actions", mappingID)
		}
	}

	return nil
}

// validateSourceReferences ensures all referenced sources exist.
func (p *Parser) validateSourceReferences(cfg *Config) error {
	for mappingID, mapping := range cfg.Mappings {
		// Check default source
		if mapping.Source != "" {
			if _, exists := cfg.Sources[mapping.Source]; !exists {
				return fmt.Errorf("mapping '%s': source '%s' not defined", mappingID, mapping.Source)
			}
		}

		// Check operation sources
		for opName, op := range mapping.Operations {
			if op.Source != "" {
				if _, exists := cfg.Sources[op.Source]; !exists {
					return fmt.Errorf("mapping '%s', operation '%s': source '%s' not defined",
						mappingID, opName, op.Source)
				}
			}

			// Check fallback chain sources
			for i, sourceRef := range op.Sources {
				if _, exists := cfg.Sources[sourceRef.Name]; !exists {
					return fmt.Errorf("mapping '%s', operation '%s', source[%d]: source '%s' not defined",
						mappingID, opName, i, sourceRef.Name)
				}
			}

			// Check after action sources
			for i, after := range op.After {
				if after.Source != "" {
					if _, exists := cfg.Sources[after.Source]; !exists {
						return fmt.Errorf("mapping '%s', operation '%s', after[%d]: source '%s' not defined",
							mappingID, opName, i, after.Source)
					}
				}
			}
		}

		// Check action sources
		for actionName, action := range mapping.Actions {
			if action.Source != "" {
				if _, exists := cfg.Sources[action.Source]; !exists {
					return fmt.Errorf("mapping '%s', action '%s': source '%s' not defined",
						mappingID, actionName, action.Source)
				}
			}
		}
	}

	return nil
}

// resolveCredentials resolves environment variables and credential references in sources.
func (p *Parser) resolveCredentials(cfg *Config) error {
	for sourceName, source := range cfg.Sources {
		resolved, err := p.credResolver.Resolve(source.Connection)
		if err != nil {
			return fmt.Errorf("source '%s': %w", sourceName, err)
		}
		source.Connection = resolved
		cfg.Sources[sourceName] = source
	}
	return nil
}
