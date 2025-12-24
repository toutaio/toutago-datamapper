package engine

import (
	"context"
	"fmt"

	"github.com/toutago/toutago-datamapper/adapter"
	"github.com/toutago/toutago-datamapper/config"
)

// Mapper is the main orchestration engine that coordinates configuration,
// adapters, and property mapping to execute data operations.
type Mapper struct {
	parser   *config.Parser
	registry *AdapterRegistry
	propMap  *PropertyMapper
}

// NewMapper creates a new mapper instance by loading configuration from a file.
func NewMapper(configPath string) (*Mapper, error) {
	parser := config.NewParser()
	if err := parser.LoadFile(configPath); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := parser.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Mapper{
		parser:   parser,
		registry: NewAdapterRegistry(),
		propMap:  NewPropertyMapper(),
	}, nil
}

// NewMapperWithParser creates a mapper with an existing parser.
// Useful when you want to load multiple config files or use custom credential resolution.
func NewMapperWithParser(parser *config.Parser) (*Mapper, error) {
	if err := parser.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Mapper{
		parser:   parser,
		registry: NewAdapterRegistry(),
		propMap:  NewPropertyMapper(),
	}, nil
}

// RegisterAdapter registers an adapter factory for a specific adapter type.
func (m *Mapper) RegisterAdapter(adapterType string, factory AdapterFactory) {
	m.registry.Register(adapterType, factory)
}

// Fetch retrieves a single object using the specified mapping.
// params should contain the parameter values for the query.
// result must be a pointer to a struct where the data will be mapped.
func (m *Mapper) Fetch(ctx context.Context, mappingID string, params map[string]interface{}, result interface{}) error {
	mapping, cfg, err := m.parser.GetMapping(mappingID)
	if err != nil {
		return err
	}

	opConfig, exists := mapping.Operations["fetch"]
	if !exists {
		return fmt.Errorf("mapping '%s' does not have a 'fetch' operation", mappingID)
	}

	// Resolve source
	source, sourceID, err := m.resolveSource(cfg, mapping, &opConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve source for fetch: %w", err)
	}

	// Get adapter
	adp, err := m.registry.GetAdapter(ctx, source, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}

	// Build operation
	op := m.buildOperation(adapter.OpFetch, &opConfig)
	op.Multi = false

	// Execute fetch
	results, err := adp.Fetch(ctx, op, params)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	if len(results) == 0 {
		return adapter.ErrNotFound
	}

	// Map result to object
	if opConfig.Result != nil {
		dataMap, ok := results[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map[string]interface{}, got %T", results[0])
		}

		if err := m.propMap.MapToObject(dataMap, result, opConfig.Result.Properties); err != nil {
			return fmt.Errorf("failed to map result: %w", err)
		}
	}

	return nil
}

// FetchMulti retrieves multiple objects using the specified mapping.
// Returns a slice of results.
func (m *Mapper) FetchMulti(ctx context.Context, mappingID string, params map[string]interface{}, results interface{}) error {
	mapping, cfg, err := m.parser.GetMapping(mappingID)
	if err != nil {
		return err
	}

	opConfig, exists := mapping.Operations["fetch"]
	if !exists {
		return fmt.Errorf("mapping '%s' does not have a 'fetch' operation", mappingID)
	}

	// Resolve source
	source, sourceID, err := m.resolveSource(cfg, mapping, &opConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve source for fetch: %w", err)
	}

	// Get adapter
	adp, err := m.registry.GetAdapter(ctx, source, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}

	// Build operation
	op := m.buildOperation(adapter.OpFetch, &opConfig)
	op.Multi = true

	// Execute fetch
	data, err := adp.Fetch(ctx, op, params)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	// Map results to objects
	if opConfig.Result != nil {
		// TODO: Implement slice mapping
		_ = data
		_ = results
	}

	return nil
}

// Insert creates new objects in the data source.
// objects can be a single object or a slice of objects.
func (m *Mapper) Insert(ctx context.Context, mappingID string, objects interface{}) error {
	mapping, cfg, err := m.parser.GetMapping(mappingID)
	if err != nil {
		return err
	}

	opConfig, exists := mapping.Operations["insert"]
	if !exists {
		return fmt.Errorf("mapping '%s' does not have an 'insert' operation", mappingID)
	}

	// Resolve source
	source, sourceID, err := m.resolveSource(cfg, mapping, &opConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve source for insert: %w", err)
	}

	// Get adapter
	adp, err := m.registry.GetAdapter(ctx, source, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}

	// Build operation
	op := m.buildOperation(adapter.OpInsert, &opConfig)

	// Map objects to data
	var dataObjects []interface{}
	// TODO: Handle single vs multiple objects
	_ = objects

	// Execute insert
	if err := adp.Insert(ctx, op, dataObjects); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	// Execute after actions
	if err := m.executeAfterActions(ctx, cfg, opConfig.After, nil); err != nil {
		return fmt.Errorf("after actions failed: %w", err)
	}

	return nil
}

// Update modifies existing objects in the data source.
func (m *Mapper) Update(ctx context.Context, mappingID string, objects interface{}) error {
	mapping, cfg, err := m.parser.GetMapping(mappingID)
	if err != nil {
		return err
	}

	opConfig, exists := mapping.Operations["update"]
	if !exists {
		return fmt.Errorf("mapping '%s' does not have an 'update' operation", mappingID)
	}

	// Resolve source
	source, sourceID, err := m.resolveSource(cfg, mapping, &opConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve source for update: %w", err)
	}

	// Get adapter
	adp, err := m.registry.GetAdapter(ctx, source, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}

	// Build operation
	op := m.buildOperation(adapter.OpUpdate, &opConfig)

	// Map objects to data
	var dataObjects []interface{}
	// TODO: Handle single vs multiple objects
	_ = objects

	// Execute update
	if err := adp.Update(ctx, op, dataObjects); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Execute after actions
	if err := m.executeAfterActions(ctx, cfg, opConfig.After, nil); err != nil {
		return fmt.Errorf("after actions failed: %w", err)
	}

	return nil
}

// Delete removes objects from the data source.
func (m *Mapper) Delete(ctx context.Context, mappingID string, identifiers interface{}) error {
	mapping, cfg, err := m.parser.GetMapping(mappingID)
	if err != nil {
		return err
	}

	opConfig, exists := mapping.Operations["delete"]
	if !exists {
		return fmt.Errorf("mapping '%s' does not have a 'delete' operation", mappingID)
	}

	// Resolve source
	source, sourceID, err := m.resolveSource(cfg, mapping, &opConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve source for delete: %w", err)
	}

	// Get adapter
	adp, err := m.registry.GetAdapter(ctx, source, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}

	// Build operation
	op := m.buildOperation(adapter.OpDelete, &opConfig)

	// Map identifiers
	var ids []interface{}
	// TODO: Handle identifier mapping
	_ = identifiers

	// Execute delete
	if err := adp.Delete(ctx, op, ids); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	// Execute after actions
	if err := m.executeAfterActions(ctx, cfg, opConfig.After, nil); err != nil {
		return fmt.Errorf("after actions failed: %w", err)
	}

	return nil
}

// Execute runs a custom action.
func (m *Mapper) Execute(ctx context.Context, actionID string, params map[string]interface{}, result interface{}) error {
	// TODO: Implement Execute properly
	// For now, return not implemented error
	return fmt.Errorf("Execute not yet implemented")
}

// Close closes all adapter instances and releases resources.
func (m *Mapper) Close() error {
	return m.registry.Close()
}

// resolveSource determines which source to use for an operation (CQRS support).
func (m *Mapper) resolveSource(cfg *config.Config, mapping *config.Mapping, opConfig *config.OperationConfig) (config.Source, string, error) {
	// Operation-specific source takes precedence
	if opConfig.Source != "" {
		source, exists := cfg.Sources[opConfig.Source]
		if !exists {
			return config.Source{}, "", fmt.Errorf("source '%s' not found", opConfig.Source)
		}
		return source, opConfig.Source, nil
	}

	// Fallback chain (CQRS)
	if len(opConfig.Sources) > 0 {
		// For now, use the first source
		// TODO: Implement fallback logic with on_miss and on_error
		sourceRef := opConfig.Sources[0]
		source, exists := cfg.Sources[sourceRef.Name]
		if !exists {
			return config.Source{}, "", fmt.Errorf("source '%s' not found", sourceRef.Name)
		}
		return source, sourceRef.Name, nil
	}

	// Default mapping source
	if mapping.Source != "" {
		source, exists := cfg.Sources[mapping.Source]
		if !exists {
			return config.Source{}, "", fmt.Errorf("source '%s' not found", mapping.Source)
		}
		return source, mapping.Source, nil
	}

	return config.Source{}, "", fmt.Errorf("no source configured for operation")
}

// buildOperation constructs an adapter.Operation from config.OperationConfig.
func (m *Mapper) buildOperation(opType adapter.OperationType, opConfig *config.OperationConfig) *adapter.Operation {
	op := &adapter.Operation{
		Type:      opType,
		Statement: opConfig.Statement,
		Bulk:      opConfig.Bulk,
	}

	// Convert property mappings
	op.Properties = make([]adapter.PropertyMapping, len(opConfig.Properties))
	for i, pm := range opConfig.Properties {
		op.Properties[i] = adapter.PropertyMapping{
			ObjectField: pm.Object,
			DataField:   pm.Field,
			Type:        pm.Type,
			Generated:   pm.Generated,
		}
	}

	// Convert identifier mappings
	op.Identifier = make([]adapter.PropertyMapping, len(opConfig.Identifier))
	for i, pm := range opConfig.Identifier {
		op.Identifier[i] = adapter.PropertyMapping{
			ObjectField: pm.Object,
			DataField:   pm.Field,
			Type:        pm.Type,
		}
	}

	// Convert generated mappings
	op.Generated = make([]adapter.PropertyMapping, len(opConfig.Generated))
	for i, pm := range opConfig.Generated {
		op.Generated[i] = adapter.PropertyMapping{
			ObjectField: pm.Object,
			DataField:   pm.Field,
			Type:        pm.Type,
			Generated:   true,
		}
	}

	return op
}

// executeAfterActions executes after-action hooks (cache invalidation, etc.).
func (m *Mapper) executeAfterActions(ctx context.Context, cfg *config.Config, actions []config.AfterActionConfig, data map[string]interface{}) error {
	// TODO: Implement after action execution
	_ = ctx
	_ = cfg
	_ = actions
	_ = data
	return nil
}
