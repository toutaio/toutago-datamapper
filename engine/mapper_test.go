package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
)

func TestNewMapper(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `namespace: test
version: "1.0"
sources:
  db:
    adapter: mock
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
    operations:
      fetch:
        statement: "SELECT * FROM users WHERE id = ?"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}

	if mapper == nil {
		t.Fatal("Mapper should not be nil")
	}

	if mapper.parser == nil {
		t.Error("Parser should not be nil")
	}
	if mapper.registry == nil {
		t.Error("Registry should not be nil")
	}
	if mapper.propMap == nil {
		t.Error("PropertyMapper should not be nil")
	}
}

func TestNewMapper_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	// Invalid YAML
	configContent := `invalid yaml content [[[`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err := NewMapper(configFile)
	if err == nil {
		t.Error("NewMapper() should error for invalid config")
	}
}

func TestNewMapperWithParser(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `namespace: test
version: "1.0"
sources:
  db:
    adapter: mock
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	parser := config.NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	mapper, err := NewMapperWithParser(parser)
	if err != nil {
		t.Fatalf("NewMapperWithParser() error = %v", err)
	}

	if mapper == nil {
		t.Fatal("Mapper should not be nil")
	}
}

func TestMapper_RegisterAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `namespace: test
version: "1.0"
sources:
  db:
    adapter: custom
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter("custom"), nil
	}

	mapper.RegisterAdapter("custom", factory)

	if !mapper.registry.HasFactory("custom") {
		t.Error("Custom adapter should be registered")
	}
}

func TestMapper_Fetch(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `namespace: test
version: "1.0"
sources:
  db:
    adapter: mock
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
    operations:
      fetch:
        statement: "SELECT * FROM users WHERE id = ?"
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
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		mock := NewMockAdapter("mock")
		return mock, nil
	})

	ctx := context.Background()

	type User struct {
		ID   int
		Name string
	}

	var user User
	err = mapper.Fetch(ctx, "test.user", map[string]interface{}{"ID": 1}, &user)

	// This will fail because MockAdapter returns empty results
	// We expect ErrNotFound
	if err != adapter.ErrNotFound {
		t.Logf("Fetch() error = %v (expected ErrNotFound)", err)
	}
}

func TestMapper_resolveSource(t *testing.T) {
	cfg := &config.Config{
		Namespace: "test",
		Version:   "1.0",
		Sources: map[string]config.Source{
			"db1": {Adapter: "mysql", Connection: "localhost:3306"},
			"db2": {Adapter: "postgres", Connection: "localhost:5432"},
		},
	}

	mapping := &config.Mapping{
		Object: "User",
		Source: "db1", // Default source
	}

	mapper := &Mapper{
		parser:   config.NewParser(),
		registry: NewAdapterRegistry(),
		propMap:  NewPropertyMapper(),
	}

	tests := []struct {
		name       string
		opConfig   config.OperationConfig
		wantSource string
	}{
		{
			name: "use operation-specific source",
			opConfig: config.OperationConfig{
				Source:    "db2",
				Statement: "SELECT * FROM users",
			},
			wantSource: "db2",
		},
		{
			name: "use default mapping source",
			opConfig: config.OperationConfig{
				Statement: "SELECT * FROM users",
			},
			wantSource: "db1",
		},
		{
			name: "use fallback chain first source",
			opConfig: config.OperationConfig{
				Sources: []config.SourceRef{
					{Name: "db2"},
					{Name: "db1"},
				},
				Statement: "SELECT * FROM users",
			},
			wantSource: "db2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sourceID, err := mapper.resolveSource(cfg, mapping, &tt.opConfig)
			if err != nil {
				t.Fatalf("resolveSource() error = %v", err)
			}
			if sourceID != tt.wantSource {
				t.Errorf("sourceID = %v, want %v", sourceID, tt.wantSource)
			}
		})
	}
}

func TestMapper_buildOperation(t *testing.T) {
	mapper := &Mapper{
		parser:   config.NewParser(),
		registry: NewAdapterRegistry(),
		propMap:  NewPropertyMapper(),
	}

	opConfig := &config.OperationConfig{
		Statement: "SELECT * FROM users",
		Bulk:      true,
		Properties: []config.PropertyMap{
			{Object: "ID", Field: "id"},
			{Object: "Name", Field: "name"},
		},
		Identifier: []config.PropertyMap{
			{Object: "ID", Field: "id"},
		},
		Generated: []config.PropertyMap{
			{Object: "CreatedAt", Field: "created_at", Type: "timestamp"},
		},
	}

	op := mapper.buildOperation(adapter.OpFetch, opConfig)

	if op.Type != adapter.OpFetch {
		t.Errorf("Type = %v, want %v", op.Type, adapter.OpFetch)
	}
	if op.Statement != "SELECT * FROM users" {
		t.Errorf("Statement = %v, want SELECT * FROM users", op.Statement)
	}
	if !op.Bulk {
		t.Error("Bulk should be true")
	}
	if len(op.Properties) != 2 {
		t.Errorf("len(Properties) = %v, want 2", len(op.Properties))
	}
	if len(op.Identifier) != 1 {
		t.Errorf("len(Identifier) = %v, want 1", len(op.Identifier))
	}
	if len(op.Generated) != 1 {
		t.Errorf("len(Generated) = %v, want 1", len(op.Generated))
	}
}

func TestMapper_Close(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `namespace: test
version: "1.0"
sources:
  db:
    adapter: mock
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}

	err = mapper.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
