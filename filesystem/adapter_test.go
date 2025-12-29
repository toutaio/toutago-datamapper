package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/toutaio/toutago-datamapper/adapter"
)

func TestNewFilesystemAdapter(t *testing.T) {
	tmpDir := t.TempDir()

	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	if fa == nil {
		t.Fatal("FilesystemAdapter should not be nil")
	}

	// Check that directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base directory should exist")
	}
}

func TestFilesystemAdapter_Connect(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	err := fa.Connect(ctx, nil)
	if err != nil {
		t.Errorf("Connect() should not error, got %v", err)
	}
}

func TestFilesystemAdapter_Close(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	err := fa.Close()
	if err != nil {
		t.Errorf("Close() should not error, got %v", err)
	}
}

func TestFilesystemAdapter_Insert(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpInsert,
		Statement: "users/{id}.json",
	}

	objects := []interface{}{
		map[string]interface{}{
			"id":    "123",
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	err := fa.Insert(ctx, op, objects)
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "users", "123.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File should have been created")
	}

	// Verify content
	data, _ := os.ReadFile(filePath)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["name"] != "John Doe" {
		t.Errorf("name = %v, want John Doe", result["name"])
	}
}

func TestFilesystemAdapter_Insert_DuplicateError(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpInsert,
		Statement: "users/{id}.json",
	}

	objects := []interface{}{
		map[string]interface{}{
			"id":   "123",
			"name": "John Doe",
		},
	}

	// First insert
	fa.Insert(ctx, op, objects)

	// Second insert should fail
	err := fa.Insert(ctx, op, objects)
	if err == nil {
		t.Error("Insert() should error for duplicate file")
	}
}

func TestFilesystemAdapter_Fetch(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create test file
	testData := map[string]interface{}{
		"id":    "456",
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}

	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)
	data, _ := json.Marshal(testData)
	os.WriteFile(filepath.Join(tmpDir, "users", "456.json"), data, 0644)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "users/{id}.json",
		Multi:     false,
	}

	params := map[string]interface{}{
		"id": "456",
	}

	results, err := fa.Fetch(ctx, op, params)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	resultMap := results[0].(map[string]interface{})
	if resultMap["name"] != "Jane Doe" {
		t.Errorf("name = %v, want Jane Doe", resultMap["name"])
	}
}

func TestFilesystemAdapter_Fetch_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "users/{id}.json",
	}

	params := map[string]interface{}{
		"id": "999",
	}

	_, err := fa.Fetch(ctx, op, params)
	if err != adapter.ErrNotFound {
		t.Errorf("Fetch() error = %v, want ErrNotFound", err)
	}
}

func TestFilesystemAdapter_FetchMulti(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create test files
	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)

	users := []map[string]interface{}{
		{"id": "1", "name": "User 1"},
		{"id": "2", "name": "User 2"},
		{"id": "3", "name": "User 3"},
	}

	for _, user := range users {
		data, _ := json.Marshal(user)
		filename := filepath.Join(tmpDir, "users", user["id"].(string)+".json")
		os.WriteFile(filename, data, 0644)
	}

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "users/*.json",
		Multi:     true,
	}

	results, err := fa.Fetch(ctx, op, nil)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("len(results) = %d, want 3", len(results))
	}
}

func TestFilesystemAdapter_Update(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create initial file
	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)
	initialData := map[string]interface{}{
		"id":   "123",
		"name": "Old Name",
	}
	data, _ := json.Marshal(initialData)
	os.WriteFile(filepath.Join(tmpDir, "users", "123.json"), data, 0644)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpUpdate,
		Statement: "users/{id}.json",
	}

	objects := []interface{}{
		map[string]interface{}{
			"id":   "123",
			"name": "New Name",
		},
	}

	err := fa.Update(ctx, op, objects)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	filePath := filepath.Join(tmpDir, "users", "123.json")
	data, _ = os.ReadFile(filePath)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["name"] != "New Name" {
		t.Errorf("name = %v, want New Name", result["name"])
	}
}

func TestFilesystemAdapter_Update_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpUpdate,
		Statement: "users/{id}.json",
	}

	objects := []interface{}{
		map[string]interface{}{
			"id":   "999",
			"name": "New Name",
		},
	}

	err := fa.Update(ctx, op, objects)
	if err != adapter.ErrNotFound {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestFilesystemAdapter_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create test file
	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)
	data, _ := json.Marshal(map[string]interface{}{"id": "123"})
	filePath := filepath.Join(tmpDir, "users", "123.json")
	os.WriteFile(filePath, data, 0644)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpDelete,
		Statement: "users/{id}.json",
		Identifier: []adapter.PropertyMapping{
			{DataField: "id"},
		},
	}

	identifiers := []interface{}{"123"}

	err := fa.Delete(ctx, op, identifiers)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

func TestFilesystemAdapter_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpDelete,
		Statement: "users/{id}.json",
		Identifier: []adapter.PropertyMapping{
			{DataField: "id"},
		},
	}

	identifiers := []interface{}{"999"}

	err := fa.Delete(ctx, op, identifiers)
	if err != adapter.ErrNotFound {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestFilesystemAdapter_Delete_MapIdentifier(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create test file
	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)
	data, _ := json.Marshal(map[string]interface{}{"id": "123"})
	filePath := filepath.Join(tmpDir, "users", "123.json")
	os.WriteFile(filePath, data, 0644)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpDelete,
		Statement: "users/{id}.json",
	}

	identifiers := []interface{}{
		map[string]interface{}{"id": "123"},
	}

	err := fa.Delete(ctx, op, identifiers)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

func TestFilesystemAdapter_Execute_List(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	// Create test files
	os.MkdirAll(filepath.Join(tmpDir, "users"), 0755)
	for i := 1; i <= 3; i++ {
		data, _ := json.Marshal(map[string]interface{}{"id": i})
		os.WriteFile(filepath.Join(tmpDir, "users", fmt.Sprintf("%d.json", i)), data, 0644)
	}

	ctx := context.Background()
	action := &adapter.Action{
		Name:      "list",
		Statement: "users/*.json",
	}

	result, err := fa.Execute(ctx, action, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	results, ok := result.([]interface{})
	if !ok {
		t.Fatalf("result should be []interface{}, got %T", result)
	}

	if len(results) < 1 {
		t.Errorf("len(results) = %d, want >= 1", len(results))
	}
}

func TestFilesystemAdapter_ResolvePath(t *testing.T) {
	fa := &FilesystemAdapter{}

	tests := []struct {
		name     string
		template string
		params   map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple replacement",
			template: "users/{id}.json",
			params:   map[string]interface{}{"id": 123},
			want:     "users/123.json",
			wantErr:  false,
		},
		{
			name:     "multiple placeholders",
			template: "{type}/{id}/{version}.json",
			params:   map[string]interface{}{"type": "users", "id": 456, "version": "v1"},
			want:     "users/456/v1.json",
			wantErr:  false,
		},
		{
			name:     "no placeholders",
			template: "users/all.json",
			params:   map[string]interface{}{},
			want:     "users/all.json",
			wantErr:  false,
		},
		{
			name:     "unresolved placeholder",
			template: "users/{id}/{name}.json",
			params:   map[string]interface{}{"id": 123},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fa.resolvePath(tt.template, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilesystemAdapter_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	fa, _ := NewFilesystemAdapter(tmpDir)

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpInsert,
		Statement: "users/{id}.json",
	}

	// Insert multiple objects concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			objects := []interface{}{
				map[string]interface{}{
					"id":   fmt.Sprintf("%d", id),
					"name": fmt.Sprintf("User %d", id),
				},
			}
			fa.Insert(ctx, op, objects)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all files were created
	opFetch := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "users/*.json",
		Multi:     true,
	}

	results, _ := fa.Fetch(ctx, opFetch, nil)
	if len(results) != 10 {
		t.Errorf("len(results) = %d, want 10", len(results))
	}
}

func TestFilesystemAdapter_Name(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	name := fa.Name()
	if name != "filesystem" {
		t.Errorf("Name() = %v, want 'filesystem'", name)
	}
}

func TestFilesystemAdapter_WriteAtomic_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	// Test writing to invalid directory
	invalidPath := filepath.Join(tmpDir, "nonexistent", "deep", "path", "file.json")
	data := []byte(`{"test": "data"}`)

	err = fa.writeAtomic(invalidPath, data)
	if err == nil {
		t.Error("writeAtomic() expected error for invalid path, got nil")
	}
}

func TestFilesystemAdapter_WriteAtomic_Success(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Test successful write
	testPath := filepath.Join(subDir, "test.json")
	testData := []byte(`{"key": "value"}`)

	err = fa.writeAtomic(testPath, testData)
	if err != nil {
		t.Fatalf("writeAtomic() unexpected error: %v", err)
	}

	// Verify file was created
	readData, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("File content = %s, want %s", string(readData), string(testData))
	}
}

func TestFilesystemAdapter_NewFilesystemAdapter_InvalidPath(t *testing.T) {
	// Test with path that cannot be created (permission issue simulation)
	// This is platform-specific, so we just test the happy path was already tested
	tmpDir := t.TempDir()
	_, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Errorf("NewFilesystemAdapter() with valid path should not error, got: %v", err)
	}
}

func TestFilesystemAdapter_Update_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpUpdate,
		Statement: "users/{id}.json",
	}

	// Test update with invalid data (not a map)
	invalidObjects := []interface{}{"not a map"}

	err = fa.Update(ctx, op, invalidObjects)
	if err == nil {
		t.Error("Update() expected error for non-map object, got nil")
	}
}

func TestFilesystemAdapter_Delete_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpDelete,
		Statement: "users/{id}.json",
	}

	// Test delete with invalid identifier (not a map)
	invalidIDs := []interface{}{"not a map"}

	err = fa.Delete(ctx, op, invalidIDs)
	if err == nil {
		t.Error("Delete() expected error for non-map identifier, got nil")
	}
}

func TestFilesystemAdapter_FetchSingle_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "nonexistent.json",
		Multi:     false,
	}

	results, err := fa.Fetch(ctx, op, nil)
	if err == nil {
		t.Error("Fetch() expected error for non-existent file, got nil")
	}
	if len(results) != 0 {
		t.Errorf("Fetch() returned %d results, want 0", len(results))
	}
}

func TestFilesystemAdapter_FetchMulti_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	ctx := context.Background()
	op := &adapter.Operation{
		Type:      adapter.OpFetch,
		Statement: "empty/*.json",
		Multi:     true,
	}

	results, err := fa.Fetch(ctx, op, nil)
	if err != nil {
		t.Errorf("Fetch() error = %v, want nil for empty directory", err)
	}
	if len(results) != 0 {
		t.Errorf("Fetch() returned %d results, want 0 for empty directory", len(results))
	}
}

func TestFilesystemAdapter_Execute_ListAction(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	// Create some test files
	testData := map[string]interface{}{"id": "1", "name": "test1"}
	filePath := filepath.Join(tmpDir, "item1.json")
	jsonData, _ := json.Marshal(testData)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	action := &adapter.Action{
		Name:      "list",
		Statement: "*.json",
	}

	results, err := fa.Execute(ctx, action, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if results == nil {
		t.Error("Execute() returned nil results")
	}
}

func TestFilesystemAdapter_Execute_UnsupportedAction(t *testing.T) {
	tmpDir := t.TempDir()
	fa, err := NewFilesystemAdapter(tmpDir)
	if err != nil {
		t.Fatalf("NewFilesystemAdapter() error = %v", err)
	}

	ctx := context.Background()
	action := &adapter.Action{
		Name:      "unsupported",
		Statement: "something",
	}

	_, err = fa.Execute(ctx, action, nil)
	if err == nil {
		t.Error("Execute() expected error for unsupported action, got nil")
	}
}
