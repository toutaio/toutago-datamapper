package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
)

// MockAdapter is a mock adapter for testing
type MockAdapter struct {
	name        string
	connected   bool
	closed      bool
	connectErr  error
	closeErr    error
}

func NewMockAdapter(name string) *MockAdapter {
	return &MockAdapter{name: name}
}

func (m *MockAdapter) Fetch(ctx context.Context, op *adapter.Operation, params map[string]interface{}) ([]interface{}, error) {
	return nil, nil
}

func (m *MockAdapter) Insert(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	return nil
}

func (m *MockAdapter) Update(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	return nil
}

func (m *MockAdapter) Delete(ctx context.Context, op *adapter.Operation, identifiers []interface{}) error {
	return nil
}

func (m *MockAdapter) Execute(ctx context.Context, action *adapter.Action, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}

func (m *MockAdapter) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	m.closed = true
	return nil
}

func (m *MockAdapter) Name() string {
	return m.name
}

func TestAdapterRegistry_Register(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter("test"), nil
	}

	registry.Register("test", factory)

	if !registry.HasFactory("test") {
		t.Error("Factory should be registered")
	}
	if registry.HasFactory("nonexistent") {
		t.Error("Nonexistent factory should not be registered")
	}
}

func TestAdapterRegistry_GetAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	adapter1, err := registry.GetAdapter(ctx, source, "source1")
	if err != nil {
		t.Fatalf("GetAdapter() error = %v", err)
	}
	if adapter1 == nil {
		t.Fatal("Adapter should not be nil")
	}

	// Should return same instance
	adapter2, err := registry.GetAdapter(ctx, source, "source1")
	if err != nil {
		t.Fatalf("GetAdapter() error = %v", err)
	}
	if adapter1 != adapter2 {
		t.Error("Should return same adapter instance")
	}
}

func TestAdapterRegistry_GetAdapter_UnregisteredType(t *testing.T) {
	registry := NewAdapterRegistry()

	source := config.Source{
		Adapter:    "unknown",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	_, err := registry.GetAdapter(ctx, source, "source1")
	if err == nil {
		t.Error("GetAdapter() should error for unregistered adapter type")
	}
}

func TestAdapterRegistry_GetAdapter_FactoryError(t *testing.T) {
	registry := NewAdapterRegistry()

	expectedErr := errors.New("factory error")
	factory := func(source config.Source) (adapter.Adapter, error) {
		return nil, expectedErr
	}

	registry.Register("failing", factory)

	source := config.Source{
		Adapter:    "failing",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	_, err := registry.GetAdapter(ctx, source, "source1")
	if err == nil {
		t.Error("GetAdapter() should error when factory fails")
	}
}

func TestAdapterRegistry_GetAdapter_ConnectError(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		mock := NewMockAdapter("test")
		mock.connectErr = errors.New("connection failed")
		return mock, nil
	}

	registry.Register("failing-connect", factory)

	source := config.Source{
		Adapter:    "failing-connect",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	_, err := registry.GetAdapter(ctx, source, "source1")
	if err == nil {
		t.Error("GetAdapter() should error when Connect() fails")
	}
}

func TestAdapterRegistry_Close(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	
	// Create multiple instances
	_, _ = registry.GetAdapter(ctx, source, "source1")
	_, _ = registry.GetAdapter(ctx, source, "source2")

	if len(registry.ListInstances()) != 2 {
		t.Errorf("Should have 2 instances, got %d", len(registry.ListInstances()))
	}

	// Close all
	if err := registry.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if len(registry.ListInstances()) != 0 {
		t.Error("All instances should be closed")
	}
}

func TestAdapterRegistry_CloseInstance(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	instance, _ := registry.GetAdapter(ctx, source, "source1")

	mock := instance.(*MockAdapter)
	if mock.closed {
		t.Error("Instance should not be closed yet")
	}

	// Close specific instance
	if err := registry.CloseInstance("source1"); err != nil {
		t.Errorf("CloseInstance() error = %v", err)
	}

	if !mock.closed {
		t.Error("Instance should be closed")
	}

	// Try to close again
	err := registry.CloseInstance("source1")
	if err == nil {
		t.Error("CloseInstance() should error for non-existent instance")
	}
}

func TestAdapterRegistry_GetInstance(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()
	
	// No instance yet
	_, exists := registry.GetInstance("source1")
	if exists {
		t.Error("Instance should not exist yet")
	}

	// Create instance
	original, _ := registry.GetAdapter(ctx, source, "source1")

	// Get instance
	retrieved, exists := registry.GetInstance("source1")
	if !exists {
		t.Error("Instance should exist")
	}
	if retrieved != original {
		t.Error("Should return same instance")
	}
}

func TestAdapterRegistry_ListInstances(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()

	// No instances initially
	if len(registry.ListInstances()) != 0 {
		t.Error("Should have no instances initially")
	}

	// Create instances
	registry.GetAdapter(ctx, source, "source1")
	registry.GetAdapter(ctx, source, "source2")
	registry.GetAdapter(ctx, source, "source3")

	instances := registry.ListInstances()
	if len(instances) != 3 {
		t.Errorf("Should have 3 instances, got %d", len(instances))
	}
}

func TestAdapterRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewAdapterRegistry()

	factory := func(source config.Source) (adapter.Adapter, error) {
		return NewMockAdapter(source.Adapter), nil
	}

	registry.Register("mock", factory)

	source := config.Source{
		Adapter:    "mock",
		Connection: "test://localhost",
	}

	ctx := context.Background()

	// Concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := registry.GetAdapter(ctx, source, "concurrent-source")
			if err != nil {
				t.Errorf("GetAdapter() error = %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should only have 1 instance despite concurrent access
	instances := registry.ListInstances()
	if len(instances) != 1 {
		t.Errorf("Should have 1 instance, got %d", len(instances))
	}
}
