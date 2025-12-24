package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfig_Structure(t *testing.T) {
	cfg := Config{
		Namespace: "users",
		Version:   "1.0",
		Sources: map[string]Source{
			"main-db": {
				Adapter:    "mysql",
				Connection: "user:pass@tcp(localhost:3306)/db",
				Options: map[string]interface{}{
					"max_connections": 10,
				},
			},
		},
		Mappings: map[string]Mapping{
			"user-crud": {
				Object: "User",
				Source: "main-db",
			},
		},
	}

	if cfg.Namespace != "users" {
		t.Errorf("Namespace = %v, want users", cfg.Namespace)
	}
	if cfg.Version != "1.0" {
		t.Errorf("Version = %v, want 1.0", cfg.Version)
	}
	if len(cfg.Sources) != 1 {
		t.Errorf("len(Sources) = %v, want 1", len(cfg.Sources))
	}
	if len(cfg.Mappings) != 1 {
		t.Errorf("len(Mappings) = %v, want 1", len(cfg.Mappings))
	}
}

func TestConfig_YAMLMarshal(t *testing.T) {
	cfg := Config{
		Namespace: "test",
		Version:   "1.0",
		Sources: map[string]Source{
			"db": {
				Adapter:    "mysql",
				Connection: "localhost",
			},
		},
		Mappings: map[string]Mapping{
			"user": {
				Object: "User",
				Source: "db",
			},
		},
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	var cfg2 Config
	err = yaml.Unmarshal(data, &cfg2)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}

	if cfg2.Namespace != cfg.Namespace {
		t.Errorf("After YAML round-trip: Namespace = %v, want %v", cfg2.Namespace, cfg.Namespace)
	}
}

func TestConfig_JSONMarshal(t *testing.T) {
	cfg := Config{
		Namespace: "test",
		Version:   "1.0",
		Sources: map[string]Source{
			"db": {
				Adapter:    "postgres",
				Connection: "postgres://localhost/db",
			},
		},
		Mappings: map[string]Mapping{
			"order": {
				Object: "Order",
				Source: "db",
			},
		},
	}

	data, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var cfg2 Config
	err = json.Unmarshal(data, &cfg2)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if cfg2.Namespace != cfg.Namespace {
		t.Errorf("After JSON round-trip: Namespace = %v, want %v", cfg2.Namespace, cfg.Namespace)
	}
}

func TestOperationConfig_Structure(t *testing.T) {
	op := OperationConfig{
		Source:    "main-db",
		Statement: "SELECT * FROM users WHERE id = ?",
		Parameters: []PropertyMap{
			{Object: "ID", Field: "id"},
		},
		Properties: []PropertyMap{
			{Object: "ID", Field: "id"},
			{Object: "Name", Field: "name"},
			{Object: "Email", Field: "email"},
		},
		Result: &ResultConfig{
			Type:  "User",
			Multi: false,
			Properties: []PropertyMap{
				{Object: "ID", Field: "id"},
			},
		},
	}

	if op.Source != "main-db" {
		t.Error("Failed to set Source")
	}
	if len(op.Parameters) != 1 {
		t.Error("Failed to set Parameters")
	}
	if len(op.Properties) != 3 {
		t.Error("Failed to set Properties")
	}
	if op.Result == nil {
		t.Error("Result should not be nil")
	}
	if op.Result.Type != "User" {
		t.Error("Result.Type should be User")
	}
}

func TestOperationConfig_FallbackChain(t *testing.T) {
	op := OperationConfig{
		Sources: []SourceRef{
			{Name: "cache", OnMiss: "next"},
			{Name: "read-replica", OnError: "next"},
			{Name: "master-db"},
		},
		Statement: "SELECT * FROM users WHERE id = ?",
	}

	if len(op.Sources) != 3 {
		t.Errorf("len(Sources) = %v, want 3", len(op.Sources))
	}
	if op.Sources[0].Name != "cache" {
		t.Error("First source should be cache")
	}
	if op.Sources[0].OnMiss != "next" {
		t.Error("Cache OnMiss should be next")
	}
	if op.Sources[1].OnError != "next" {
		t.Error("Replica OnError should be next")
	}
}

func TestOperationConfig_WithAfterActions(t *testing.T) {
	op := OperationConfig{
		Source:    "main-db",
		Statement: "UPDATE users SET name = ? WHERE id = ?",
		After: []AfterActionConfig{
			{
				Action:    "invalidate",
				Source:    "cache",
				Statement: "user:{id}",
			},
		},
	}

	if len(op.After) != 1 {
		t.Error("Should have 1 after action")
	}
	if op.After[0].Action != "invalidate" {
		t.Error("After action should be invalidate")
	}
	if op.After[0].Source != "cache" {
		t.Error("After action source should be cache")
	}
}

func TestPropertyMap_Structure(t *testing.T) {
	pm := PropertyMap{
		Object:    "CreatedAt",
		Field:     "created_at",
		Type:      "timestamp",
		Generated: true,
	}

	if pm.Object != "CreatedAt" {
		t.Error("Failed to set Object")
	}
	if pm.Field != "created_at" {
		t.Error("Failed to set Field")
	}
	if pm.Type != "timestamp" {
		t.Error("Failed to set Type")
	}
	if !pm.Generated {
		t.Error("Generated should be true")
	}
}

func TestCredentialsConfig_Structure(t *testing.T) {
	creds := CredentialsConfig{
		Credentials: map[string]CredentialSource{
			"main-db": {
				Connection: "user:pass@tcp(host:3306)/db",
				Options: map[string]interface{}{
					"max_connections": 20,
				},
			},
		},
	}

	if len(creds.Credentials) != 1 {
		t.Error("Should have 1 credential")
	}
	if creds.Credentials["main-db"].Connection == "" {
		t.Error("Connection should not be empty")
	}
}

func TestCredentialsConfig_YAMLMarshal(t *testing.T) {
	creds := CredentialsConfig{
		Credentials: map[string]CredentialSource{
			"db": {
				Connection: "secret-connection-string",
				Options: map[string]interface{}{
					"timeout": "5s",
				},
			},
		},
	}

	data, err := yaml.Marshal(&creds)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	var creds2 CredentialsConfig
	err = yaml.Unmarshal(data, &creds2)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}

	if creds2.Credentials["db"].Connection != creds.Credentials["db"].Connection {
		t.Error("Connection mismatch after YAML round-trip")
	}
}

func TestMapping_CompleteStructure(t *testing.T) {
	mapping := Mapping{
		Object: "User",
		Source: "main-db",
		Operations: map[string]OperationConfig{
			"fetch": {
				Statement: "SELECT * FROM users WHERE id = ?",
				Parameters: []PropertyMap{
					{Object: "ID", Field: "id"},
				},
			},
			"insert": {
				Statement: "INSERT INTO users (name, email) VALUES (?, ?)",
				Properties: []PropertyMap{
					{Object: "Name", Field: "name"},
					{Object: "Email", Field: "email"},
				},
				Generated: []PropertyMap{
					{Object: "ID", Field: "id", Generated: true},
				},
			},
		},
		Actions: map[string]ActionConfig{
			"search": {
				Statement: "SELECT * FROM users WHERE name LIKE ?",
				Parameters: []PropertyMap{
					{Object: "SearchTerm", Field: "term"},
				},
				Result: &ResultConfig{
					Type:  "User",
					Multi: true,
				},
			},
		},
	}

	if mapping.Object != "User" {
		t.Error("Object should be User")
	}
	if len(mapping.Operations) != 2 {
		t.Error("Should have 2 operations")
	}
	if len(mapping.Actions) != 1 {
		t.Error("Should have 1 action")
	}
	if mapping.Operations["insert"].Generated == nil {
		t.Error("Insert operation should have Generated fields")
	}
	if len(mapping.Operations["insert"].Generated) != 1 {
		t.Error("Insert should have 1 generated field")
	}
}
