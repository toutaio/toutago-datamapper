package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/toutago/toutago-datamapper/adapter"
	"github.com/toutago/toutago-datamapper/config"
)

// AdapterFactory is a function that creates an adapter instance.
type AdapterFactory func(source config.Source) (adapter.Adapter, error)

// AdapterRegistry manages adapter factories and instances.
// It provides lifecycle management and connection pooling for adapters.
type AdapterRegistry struct {
	// factories maps adapter type names to their factory functions
	factories map[string]AdapterFactory

	// instances maps source identifiers to active adapter instances
	instances map[string]adapter.Adapter

	// mu protects concurrent access to instances
	mu sync.RWMutex
}

// NewAdapterRegistry creates a new adapter registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		factories: make(map[string]AdapterFactory),
		instances: make(map[string]adapter.Adapter),
	}
}

// Register registers an adapter factory for a specific adapter type.
// The adapterType should match the "adapter" field in source configuration.
func (ar *AdapterRegistry) Register(adapterType string, factory AdapterFactory) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.factories[adapterType] = factory
}

// GetAdapter returns an adapter instance for the given source.
// If an instance already exists, it is reused. Otherwise, a new one is created.
func (ar *AdapterRegistry) GetAdapter(ctx context.Context, source config.Source, sourceID string) (adapter.Adapter, error) {
	// Check if instance already exists
	ar.mu.RLock()
	if instance, exists := ar.instances[sourceID]; exists {
		ar.mu.RUnlock()
		return instance, nil
	}
	ar.mu.RUnlock()

	// Create new instance
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Double-check after acquiring write lock
	if instance, exists := ar.instances[sourceID]; exists {
		return instance, nil
	}

	// Get factory
	factory, exists := ar.factories[source.Adapter]
	if !exists {
		return nil, fmt.Errorf("no adapter factory registered for type '%s'", source.Adapter)
	}

	// Create adapter instance
	instance, err := factory(source)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter instance for '%s': %w", source.Adapter, err)
	}

	// Connect to data source
	if err := instance.Connect(ctx, source.Options); err != nil {
		return nil, fmt.Errorf("failed to connect adapter '%s': %w", source.Adapter, err)
	}

	// Store instance
	ar.instances[sourceID] = instance
	return instance, nil
}

// Close closes all adapter instances and releases resources.
func (ar *AdapterRegistry) Close() error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	var errs []error
	for sourceID, instance := range ar.instances {
		if err := instance.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing adapter '%s': %w", sourceID, err))
		}
	}

	// Clear instances
	ar.instances = make(map[string]adapter.Adapter)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing adapters: %v", errs)
	}
	return nil
}

// HasFactory returns true if a factory is registered for the given adapter type.
func (ar *AdapterRegistry) HasFactory(adapterType string) bool {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	_, exists := ar.factories[adapterType]
	return exists
}

// GetInstance returns an existing adapter instance if one exists.
func (ar *AdapterRegistry) GetInstance(sourceID string) (adapter.Adapter, bool) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	instance, exists := ar.instances[sourceID]
	return instance, exists
}

// CloseInstance closes a specific adapter instance.
func (ar *AdapterRegistry) CloseInstance(sourceID string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	instance, exists := ar.instances[sourceID]
	if !exists {
		return fmt.Errorf("no adapter instance found for source '%s'", sourceID)
	}

	if err := instance.Close(); err != nil {
		return fmt.Errorf("failed to close adapter for source '%s': %w", sourceID, err)
	}

	delete(ar.instances, sourceID)
	return nil
}

// ListInstances returns the IDs of all active adapter instances.
func (ar *AdapterRegistry) ListInstances() []string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	ids := make([]string, 0, len(ar.instances))
	for id := range ar.instances {
		ids = append(ids, id)
	}
	return ids
}
