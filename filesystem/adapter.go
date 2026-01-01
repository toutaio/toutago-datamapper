package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/toutaio/toutago-datamapper/adapter"
)

// FilesystemAdapter implements the adapter.Adapter interface for filesystem storage.
// It stores data as JSON files in a directory structure.
type FilesystemAdapter struct {
	basePath string
	mu       sync.RWMutex
}

// NewFilesystemAdapter creates a new filesystem adapter.
func NewFilesystemAdapter(basePath string) (*FilesystemAdapter, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}

	// Ensure base directory exists
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FilesystemAdapter{
		basePath: absPath,
	}, nil
}

// Connect is a no-op for filesystem adapter as it doesn't need a connection.
func (fa *FilesystemAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Close is a no-op for filesystem adapter as it doesn't hold connections.
func (fa *FilesystemAdapter) Close() error {
	return nil
}

// Name returns the adapter name.
func (fa *FilesystemAdapter) Name() string {
	return "filesystem"
}

// Fetch retrieves objects from the filesystem.
func (fa *FilesystemAdapter) Fetch(ctx context.Context, op *adapter.Operation, params map[string]interface{}) ([]interface{}, error) {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	// Resolve file path from statement (treat statement as path template)
	path, err := fa.resolvePath(op.Statement, params)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if we need to list multiple files (glob pattern)
	if op.Multi || strings.Contains(path, "*") {
		return fa.fetchMulti(path)
	}

	// Fetch single file
	data, err := fa.fetchSingle(path)
	if err != nil {
		return nil, err
	}

	return []interface{}{data}, nil
}

// fetchSingle retrieves a single file.
func (fa *FilesystemAdapter) fetchSingle(path string) (map[string]interface{}, error) {
	fullPath := filepath.Join(fa.basePath, path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, adapter.ErrNotFound
	}

	// Read file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// fetchMulti retrieves multiple files matching a pattern.
func (fa *FilesystemAdapter) fetchMulti(pattern string) ([]interface{}, error) {
	fullPattern := filepath.Join(fa.basePath, pattern)

	// Find matching files
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return []interface{}{}, nil
	}

	// Read all matching files
	results := make([]interface{}, 0, len(matches))
	for _, match := range matches {
		// Skip directories
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		// Read and parse file
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// Insert creates new objects in the filesystem.
func (fa *FilesystemAdapter) Insert(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	for _, obj := range objects {
		dataMap, ok := obj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("object must be map[string]interface{}, got %T", obj)
		}

		// Resolve file path
		path, err := fa.resolvePath(op.Statement, dataMap)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		fullPath := filepath.Join(fa.basePath, path)

		// Check if file already exists
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("file already exists: %s", path)
		}

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Marshal data to JSON
		data, err := json.MarshalIndent(dataMap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Write atomically using temp file
		if err := fa.writeAtomic(fullPath, data); err != nil {
			return err
		}
	}

	return nil
}

// Update modifies existing objects in the filesystem.
func (fa *FilesystemAdapter) Update(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	for _, obj := range objects {
		dataMap, ok := obj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("object must be map[string]interface{}, got %T", obj)
		}

		// Resolve file path
		path, err := fa.resolvePath(op.Statement, dataMap)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		fullPath := filepath.Join(fa.basePath, path)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return adapter.ErrNotFound
		}

		// Marshal data to JSON
		data, err := json.MarshalIndent(dataMap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Write atomically
		if err := fa.writeAtomic(fullPath, data); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes objects from the filesystem.
func (fa *FilesystemAdapter) Delete(ctx context.Context, op *adapter.Operation, identifiers []interface{}) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	for _, id := range identifiers {
		// Convert identifier to params map
		var params map[string]interface{}
		switch v := id.(type) {
		case map[string]interface{}:
			params = v
		case string, int, int64:
			// Single value identifier, use first identifier field name
			if len(op.Identifier) > 0 {
				params = map[string]interface{}{
					op.Identifier[0].DataField: v,
				}
			} else {
				return fmt.Errorf("no identifier mapping defined")
			}
		default:
			return fmt.Errorf("unsupported identifier type: %T", id)
		}

		// Resolve file path
		path, err := fa.resolvePath(op.Statement, params)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		fullPath := filepath.Join(fa.basePath, path)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return adapter.ErrNotFound
		}

		// Delete file
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	return nil
}

// Execute runs custom actions (e.g., list all files, search).
func (fa *FilesystemAdapter) Execute(ctx context.Context, action *adapter.Action, params map[string]interface{}) (interface{}, error) {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	// For now, support basic list action
	if action.Name == "list" {
		pattern := action.Statement
		if pattern == "" {
			pattern = "*.json"
		}

		// Resolve pattern
		resolvedPattern, err := fa.resolvePath(pattern, params)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve pattern: %w", err)
		}

		return fa.fetchMulti(resolvedPattern)
	}

	return nil, fmt.Errorf("unsupported action: %s", action.Name)
}

// resolvePath resolves a path template with parameters.
// Example: "users/{id}.json" with params {"id": 123} -> "users/123.json"
func (fa *FilesystemAdapter) resolvePath(template string, params map[string]interface{}) (string, error) {
	result := template

	// Replace {param} placeholders
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
		}
	}

	// Check if there are unresolved placeholders
	if strings.Contains(result, "{") && strings.Contains(result, "}") {
		return "", fmt.Errorf("unresolved placeholders in path: %s", result)
	}

	return result, nil
}

// writeAtomic writes data to a file atomically using a temp file.
func (fa *FilesystemAdapter) writeAtomic(path string, data []byte) error {
	// Create temp file in same directory
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	// Write data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil

	// Rename atomically
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
