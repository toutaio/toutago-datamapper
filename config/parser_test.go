package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_LoadFile_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")

	content := `namespace: users
version: "1.0"
sources:
  main-db:
    adapter: mysql
    connection: "localhost:3306"
mappings:
  user-crud:
    object: User
    source: main-db
    operations:
      fetch:
        statement: "SELECT * FROM users WHERE id = ?"
`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	cfg, err := parser.GetConfig("users")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
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

func TestParser_LoadFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.json")

	content := `{
  "namespace": "orders",
  "version": "1.0",
  "sources": {
    "db": {
      "adapter": "postgres",
      "connection": "postgres://localhost/db"
    }
  },
  "mappings": {
    "order-crud": {
      "object": "Order",
      "source": "db"
    }
  }
}`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	cfg, err := parser.GetConfig("orders")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.Namespace != "orders" {
		t.Errorf("Namespace = %v, want orders", cfg.Namespace)
	}
}

func TestParser_LoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple config files
	configs := map[string]string{
		"users.yaml": `namespace: users
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
`,
		"orders.yaml": `namespace: orders
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "localhost"
mappings:
  order:
    object: Order
    source: db
`,
		"readme.txt": "This should be ignored",
	}

	for filename, content := range configs {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	parser := NewParser()
	if err := parser.LoadDirectory(tmpDir); err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	namespaces := parser.GetAllNamespaces()
	if len(namespaces) != 2 {
		t.Errorf("len(namespaces) = %v, want 2", len(namespaces))
	}
}

func TestParser_NamespaceCollision(t *testing.T) {
	tmpDir := t.TempDir()

	content := `namespace: test
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "localhost"
mappings:
  item:
    object: Item
    source: db
`

	file1 := filepath.Join(tmpDir, "config1.yaml")
	file2 := filepath.Join(tmpDir, "config2.yaml")

	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(file1); err != nil {
		t.Fatalf("LoadFile(file1) error = %v", err)
	}

	// Second file with same namespace should error
	err := parser.LoadFile(file2)
	if err == nil {
		t.Error("LoadFile(file2) should error due to namespace collision")
	}
}

func TestParser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Namespace: "test",
				Version:   "1.0",
				Sources: map[string]Source{
					"db": {Adapter: "mysql", Connection: "localhost"},
				},
				Mappings: map[string]Mapping{
					"user": {Object: "User", Source: "db"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing namespace",
			config: Config{
				Version: "1.0",
				Mappings: map[string]Mapping{
					"user": {Object: "User"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			config: Config{
				Namespace: "test",
				Mappings: map[string]Mapping{
					"user": {Object: "User"},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported version",
			config: Config{
				Namespace: "test",
				Version:   "2.0",
				Mappings: map[string]Mapping{
					"user": {Object: "User"},
				},
			},
			wantErr: true,
		},
		{
			name: "no mappings",
			config: Config{
				Namespace: "test",
				Version:   "1.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			err := parser.validateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_ValidateSourceReferences(t *testing.T) {
	cfg := Config{
		Namespace: "test",
		Version:   "1.0",
		Sources: map[string]Source{
			"db": {Adapter: "mysql", Connection: "localhost"},
		},
		Mappings: map[string]Mapping{
			"user": {
				Object: "User",
				Source: "nonexistent", // Invalid source reference
			},
		},
	}

	parser := NewParser()
	err := parser.validateSourceReferences(&cfg)
	if err == nil {
		t.Error("validateSourceReferences() should error for invalid source reference")
	}
}

func TestParser_GetMapping(t *testing.T) {
	parser := NewParser()
	parser.configs["users"] = &Config{
		Namespace: "users",
		Version:   "1.0",
		Mappings: map[string]Mapping{
			"user-crud": {
				Object: "User",
			},
		},
	}

	// Valid mapping
	mapping, cfg, err := parser.GetMapping("users.user-crud")
	if err != nil {
		t.Fatalf("GetMapping() error = %v", err)
	}
	if mapping.Object != "User" {
		t.Errorf("mapping.Object = %v, want User", mapping.Object)
	}
	if cfg.Namespace != "users" {
		t.Errorf("cfg.Namespace = %v, want users", cfg.Namespace)
	}

	// Invalid format
	_, _, err = parser.GetMapping("invalid")
	if err == nil {
		t.Error("GetMapping('invalid') should error")
	}

	// Nonexistent namespace
	_, _, err = parser.GetMapping("nonexistent.mapping")
	if err == nil {
		t.Error("GetMapping('nonexistent.mapping') should error")
	}

	// Nonexistent mapping
	_, _, err = parser.GetMapping("users.nonexistent")
	if err == nil {
		t.Error("GetMapping('users.nonexistent') should error")
	}
}

func TestParser_WithCredentials(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config file with env var placeholders
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `namespace: app
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"
mappings:
  user:
    object: User
    source: db
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create env file
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `DB_USER=testuser
DB_PASS=testpass
DB_HOST=localhost
DB_PORT=3306
DB_NAME=testdb
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadEnvFile(envFile); err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	cfg, err := parser.GetConfig("app")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	expectedConn := "testuser:testpass@tcp(localhost:3306)/testdb"
	actualConn := cfg.Sources["db"].Connection
	if actualConn != expectedConn {
		t.Errorf("Connection = %v, want %v", actualConn, expectedConn)
	}
}

func TestParser_LoadCredentialsFile(t *testing.T) {
	tmpDir := t.TempDir()
	credsFile := filepath.Join(tmpDir, "creds.yaml")

	content := `credentials:
  db:
    connection: "user:pass@localhost:3306/dbname"
`
	if err := os.WriteFile(credsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create credentials file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadCredentialsFile(credsFile); err != nil {
		t.Fatalf("LoadCredentialsFile() error = %v", err)
	}

	// Verify credentials were loaded by creating a config that uses them
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `namespace: test
version: "1.0"
sources:
  main:
    adapter: mysql
    connection: "@credentials:db"
mappings:
  user:
    object: User
    source: main
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	cfg, err := parser.GetConfig("test")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// The connection should be resolved from credentials
	if cfg.Sources["main"].Connection != "user:pass@localhost:3306/dbname" {
		t.Errorf("Connection = %v, want 'user:pass@localhost:3306/dbname'", cfg.Sources["main"].Connection)
	}
}

func TestParser_Validate_NoConfigs(t *testing.T) {
	parser := NewParser()
	err := parser.Validate()
	if err == nil {
		t.Error("Validate() expected error for no configurations, got nil")
	}
	if err.Error() != "no configurations loaded" {
		t.Errorf("Validate() error = %v, want 'no configurations loaded'", err)
	}
}

func TestParser_GetAllNamespaces(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two config files with different namespaces
	config1 := filepath.Join(tmpDir, "config1.yaml")
	content1 := `namespace: users
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "localhost"
mappings:
  user:
    object: User
    source: db
`
	if err := os.WriteFile(config1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}

	config2 := filepath.Join(tmpDir, "config2.yaml")
	content2 := `namespace: products
version: "1.0"
sources:
  db:
    adapter: mysql
    connection: "localhost"
mappings:
  product:
    object: Product
    source: db
`
	if err := os.WriteFile(config2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(config1); err != nil {
		t.Fatalf("LoadFile(config1) error = %v", err)
	}
	if err := parser.LoadFile(config2); err != nil {
		t.Fatalf("LoadFile(config2) error = %v", err)
	}

	namespaces := parser.GetAllNamespaces()
	if len(namespaces) != 2 {
		t.Errorf("GetAllNamespaces() returned %d namespaces, want 2", len(namespaces))
	}

	hasUsers := false
	hasProducts := false
	for _, ns := range namespaces {
		if ns == "users" {
			hasUsers = true
		}
		if ns == "products" {
			hasProducts = true
		}
	}

	if !hasUsers {
		t.Error("GetAllNamespaces() missing 'users' namespace")
	}
	if !hasProducts {
		t.Error("GetAllNamespaces() missing 'products' namespace")
	}
}

func TestParser_ValidateSourceReferences_OperationSource(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Config with invalid operation source reference
	content := `namespace: test
version: "1.0"
sources:
  db1:
    adapter: mysql
    connection: "localhost"
mappings:
  user:
    object: User
    source: db1
    operations:
      fetch:
        source: nonexistent_source
        statement: "SELECT * FROM users"
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	err := parser.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid operation source, got nil")
	}
}

func TestParser_ValidateSourceReferences_AfterAction(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Config with invalid after action source reference
	content := `namespace: test
version: "1.0"
sources:
  db1:
    adapter: mysql
    connection: "localhost"
mappings:
  user:
    object: User
    source: db1
    operations:
      insert:
        statement: "INSERT INTO users"
        after:
          - source: invalid_source
            action: "notify"
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	err := parser.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid after action source, got nil")
	}
}

func TestParser_ValidateSourceReferences_FallbackChain(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Config with invalid source in fallback chain
	content := `namespace: test
version: "1.0"
sources:
  db1:
    adapter: mysql
    connection: "localhost"
mappings:
  user:
    object: User
    source: db1
    operations:
      fetch:
        statement: "SELECT * FROM users"
        sources:
          - name: invalid_db
            statement: "SELECT * FROM users"
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	parser := NewParser()
	if err := parser.LoadFile(configFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	err := parser.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid fallback chain source, got nil")
	}
}
