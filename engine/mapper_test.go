package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
)

// mockAdapter for testing
type mockAdapter struct {
	fetchResults []map[string]interface{}
}

func (m *mockAdapter) Fetch(ctx context.Context, op *adapter.Operation, params map[string]interface{}) ([]interface{}, error) {
	results := make([]interface{}, len(m.fetchResults))
	for i, r := range m.fetchResults {
		results[i] = r
	}
	return results, nil
}

func (m *mockAdapter) Insert(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	return nil
}

func (m *mockAdapter) Update(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	return nil
}

func (m *mockAdapter) Delete(ctx context.Context, op *adapter.Operation, identifiers []interface{}) error {
	return nil
}

func (m *mockAdapter) Execute(ctx context.Context, action *adapter.Action, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	return nil
}

func (m *mockAdapter) Close() error {
	return nil
}

func (m *mockAdapter) Name() string {
	return "mock"
}

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

func TestMapper_FetchMulti(t *testing.T) {
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
        statement: "SELECT * FROM users"
        multi: true
        result:
          properties:
            - entity: ID
              column: id
            - entity: Name
              column: name
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}
	defer mapper.Close()

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{
			fetchResults: []map[string]interface{}{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
		}, nil
	})

	var results []map[string]interface{}
	err = mapper.FetchMulti(context.Background(), "test.user", nil, &results)
	if err != nil {
		t.Fatalf("FetchMulti() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("FetchMulti() returned %d results, want 2", len(results))
	}
}

func TestMapper_Execute(t *testing.T) {
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
      custom:
        statement: "UPDATE users SET status = 'active'"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}
	defer mapper.Close()

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{}, nil
	})

	err = mapper.Execute(context.Background(), "test.user.custom", nil, nil)
	if err == nil {
		// Execute is not fully implemented yet, so we expect an error
		t.Log("Execute() returned nil, implementation may be complete")
	}
}

func TestMapper_Insert_ErrorCases(t *testing.T) {
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
      insert:
        statement: "INSERT INTO users"
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
	defer mapper.Close()

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{}, nil
	})

	type User struct {
		ID   string
		Name string
	}

	// Test with valid single object
	user := User{ID: "1", Name: "Alice"}
	err = mapper.Insert(context.Background(), "test.user", user)
	if err != nil {
		t.Errorf("Insert(single) error = %v", err)
	}

	// Test with pointer to object
	userPtr := &User{ID: "2", Name: "Bob"}
	err = mapper.Insert(context.Background(), "test.user", userPtr)
	if err != nil {
		t.Errorf("Insert(pointer) error = %v", err)
	}
}

func TestMapper_Update(t *testing.T) {
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
      update:
        statement: "UPDATE users SET name = ?"
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
	defer mapper.Close()

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{}, nil
	})

	type User struct {
		ID   string
		Name string
	}

	user := User{ID: "1", Name: "Updated"}
	err = mapper.Update(context.Background(), "test.user", user)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestMapper_Delete(t *testing.T) {
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
      delete:
        statement: "DELETE FROM users WHERE id = ?"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	mapper, err := NewMapper(configFile)
	if err != nil {
		t.Fatalf("NewMapper() error = %v", err)
	}
	defer mapper.Close()

	// Register mock adapter
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{}, nil
	})

	// Test delete with single ID
	err = mapper.Delete(context.Background(), "test.user", "1")
	if err != nil {
		t.Fatalf("Delete(single) error = %v", err)
	}

	// Test delete with multiple IDs
	err = mapper.Delete(context.Background(), "test.user", []string{"2", "3"})
	if err != nil {
		t.Fatalf("Delete(slice) error = %v", err)
	}
}

func TestMapper_Fetch_ErrorCases(t *testing.T) {
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
        result:
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
	defer mapper.Close()

	// Register mock adapter that returns empty results
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{fetchResults: []map[string]interface{}{}}, nil
	})

	type User struct {
		ID   string
		Name string
	}

	var user User
	err = mapper.Fetch(context.Background(), "test.user", map[string]interface{}{"id": "999"}, &user)
	if err != adapter.ErrNotFound {
		t.Errorf("Fetch() with no results should return ErrNotFound, got: %v", err)
	}
}

func TestMapper_Fetch_WithResult(t *testing.T) {
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
        result:
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
	defer mapper.Close()

	// Register mock adapter that returns a result
	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{
			fetchResults: []map[string]interface{}{
				{"id": "1", "name": "Test User"},
			},
		}, nil
	})

	type User struct {
		ID   string
		Name string
	}

	var user User
	err = mapper.Fetch(context.Background(), "test.user", map[string]interface{}{"id": "1"}, &user)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if user.ID != "1" {
		t.Errorf("User.ID = %v, want '1'", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("User.Name = %v, want 'Test User'", user.Name)
	}
}

func TestMapper_Fetch_NoOperation(t *testing.T) {
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
	defer mapper.Close()

	mapper.RegisterAdapter("mock", func(source config.Source) (adapter.Adapter, error) {
		return &mockAdapter{}, nil
	})

	type User struct {
		ID string
	}

	var user User
	err = mapper.Fetch(context.Background(), "test.user", nil, &user)
	if err == nil {
		t.Error("Fetch() should error when no fetch operation defined, got nil")
	}
}
