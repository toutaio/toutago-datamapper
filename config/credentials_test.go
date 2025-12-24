package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialResolver_EnvVars(t *testing.T) {
	cr := NewCredentialResolver()
	
	// Set test env var
	cr.SetEnvVar("TEST_VAR", "test_value")
	
	value, exists := cr.GetEnvVar("TEST_VAR")
	if !exists {
		t.Error("TEST_VAR should exist")
	}
	if value != "test_value" {
		t.Errorf("TEST_VAR = %v, want test_value", value)
	}
}

func TestCredentialResolver_ResolveSimple(t *testing.T) {
	cr := NewCredentialResolver()
	cr.SetEnvVar("DB_HOST", "localhost")
	cr.SetEnvVar("DB_PORT", "3306")
	
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single variable",
			input: "${DB_HOST}",
			want:  "localhost",
		},
		{
			name:  "multiple variables",
			input: "${DB_HOST}:${DB_PORT}",
			want:  "localhost:3306",
		},
		{
			name:  "variable in middle",
			input: "tcp://${DB_HOST}:3306",
			want:  "tcp://localhost:3306",
		},
		{
			name:  "no variables",
			input: "static-string",
			want:  "static-string",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cr.Resolve(tt.input)
			if err != nil {
				t.Errorf("Resolve() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialResolver_ResolveWithDefaults(t *testing.T) {
	cr := NewCredentialResolver()
	cr.SetEnvVar("DB_HOST", "localhost")
	// DB_PORT is not set
	
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "use default when var not set",
			input:   "${DB_PORT:-3306}",
			want:    "3306",
			wantErr: false,
		},
		{
			name:    "use actual value when var is set",
			input:   "${DB_HOST:-default_host}",
			want:    "localhost",
			wantErr: false,
		},
		{
			name:    "error when var not set and no default",
			input:   "${DB_PORT}",
			want:    "",
			wantErr: true,
		},
		{
			name:    "complex with defaults",
			input:   "${DB_HOST}:${DB_PORT:-5432}",
			want:    "localhost:5432",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cr.Resolve(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialResolver_ResolveCredentialsRef(t *testing.T) {
	cr := NewCredentialResolver()
	cr.credentials["main-db"] = CredentialSource{
		Connection: "user:pass@tcp(host:3306)/db",
	}
	
	got, err := cr.Resolve("@credentials:main-db")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}
	
	want := "user:pass@tcp(host:3306)/db"
	if got != want {
		t.Errorf("Resolve() = %v, want %v", got, want)
	}
}

func TestCredentialResolver_LoadEnvFile(t *testing.T) {
	// Create temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	
	content := `# Test env file
DB_HOST=localhost
DB_PORT=3306
DB_USER=testuser
# Comment line
DB_PASSWORD="secret123"
`
	
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}
	
	cr := NewCredentialResolver()
	if err := cr.LoadEnvFile(envFile); err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}
	
	// Check loaded variables
	tests := []struct {
		name  string
		value string
	}{
		{"DB_HOST", "localhost"},
		{"DB_PORT", "3306"},
		{"DB_USER", "testuser"},
		{"DB_PASSWORD", "secret123"},
	}
	
	for _, tt := range tests {
		got, exists := cr.GetEnvVar(tt.name)
		if !exists {
			t.Errorf("%s should exist", tt.name)
		}
		if got != tt.value {
			t.Errorf("%s = %v, want %v", tt.name, got, tt.value)
		}
	}
}

func TestCredentialResolver_LoadCredentialsFile(t *testing.T) {
	// Create temporary credentials file
	tmpDir := t.TempDir()
	credsFile := filepath.Join(tmpDir, "credentials.yaml")
	
	content := `credentials:
  main-db:
    connection: "user:pass@tcp(host:3306)/db"
    options:
      max_connections: 20
  cache:
    connection: "redis://localhost:6379"
`
	
	if err := os.WriteFile(credsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test credentials file: %v", err)
	}
	
	cr := NewCredentialResolver()
	if err := cr.LoadCredentialsFile(credsFile); err != nil {
		t.Fatalf("LoadCredentialsFile() error = %v", err)
	}
	
	// Check loaded credentials
	if len(cr.credentials) != 2 {
		t.Errorf("Should have 2 credentials, got %d", len(cr.credentials))
	}
	
	mainDB, exists := cr.credentials["main-db"]
	if !exists {
		t.Error("main-db credential should exist")
	}
	if mainDB.Connection != "user:pass@tcp(host:3306)/db" {
		t.Errorf("main-db connection incorrect: %v", mainDB.Connection)
	}
}

func TestCredentialResolver_Sanitize(t *testing.T) {
	cr := NewCredentialResolver()
	
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "sanitize password in connection string",
			input: "user:password123@tcp(host:3306)/db",
			want:  "user:***@tcp(host:3306)/db",
		},
		{
			name:  "sanitize API key",
			input: "GET /api?key=secret123",
			want:  "GET /api?key=***",
		},
		{
			name:  "sanitize Bearer token",
			input: "Authorization: Bearer abc123xyz",
			want:  "Authorization: Bearer ***",
		},
		{
			name:  "no sensitive data",
			input: "SELECT * FROM users",
			want:  "SELECT * FROM users",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cr.Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}
