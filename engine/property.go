package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/toutago/toutago-datamapper/config"
)

// PropertyMapper handles mapping between data fields and object properties using reflection.
type PropertyMapper struct{}

// NewPropertyMapper creates a new property mapper.
func NewPropertyMapper() *PropertyMapper {
	return &PropertyMapper{}
}

// MapToObject maps data fields to object properties.
// target must be a pointer to a struct.
func (pm *PropertyMapper) MapToObject(data map[string]interface{}, target interface{}, mappings []config.PropertyMap) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer, got %s", targetValue.Kind())
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got pointer to %s", targetValue.Kind())
	}

	for _, mapping := range mappings {
		// Get data value
		dataValue, exists := data[mapping.Field]
		if !exists {
			// Skip if field doesn't exist in data
			continue
		}

		// Get target field
		field := targetValue.FieldByName(mapping.Object)
		if !field.IsValid() {
			return fmt.Errorf("field '%s' not found in target struct", mapping.Object)
		}
		if !field.CanSet() {
			return fmt.Errorf("field '%s' cannot be set (unexported?)", mapping.Object)
		}

		// Convert and set value
		if err := pm.setValue(field, dataValue, mapping.Type); err != nil {
			return fmt.Errorf("failed to set field '%s': %w", mapping.Object, err)
		}
	}

	return nil
}

// MapFromObject extracts data fields from object properties.
// obj can be a struct or a pointer to a struct.
func (pm *PropertyMapper) MapFromObject(obj interface{}, mappings []config.PropertyMap) (map[string]interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("object cannot be nil")
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be a struct or pointer to struct, got %s", objValue.Kind())
	}

	data := make(map[string]interface{})

	for _, mapping := range mappings {
		// Skip generated fields when extracting
		if mapping.Generated {
			continue
		}

		// Get object field
		field := objValue.FieldByName(mapping.Object)
		if !field.IsValid() {
			return nil, fmt.Errorf("field '%s' not found in object", mapping.Object)
		}

		// Extract value
		value, err := pm.getValue(field, mapping.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get field '%s': %w", mapping.Object, err)
		}

		data[mapping.Field] = value
	}

	return data, nil
}

// setValue sets a field value with type conversion.
func (pm *PropertyMapper) setValue(field reflect.Value, value interface{}, typeHint string) error {
	if value == nil {
		// Set zero value for nil
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// Handle type conversions based on hint
	switch typeHint {
	case "timestamp":
		return pm.setTimestamp(field, value)
	case "json":
		return pm.setJSON(field, value)
	default:
		return pm.setDirect(field, value)
	}
}

// setDirect sets a field value directly with basic type conversion.
func (pm *PropertyMapper) setDirect(field reflect.Value, value interface{}) error {
	valueReflect := reflect.ValueOf(value)

	// Handle pointer fields
	if field.Kind() == reflect.Ptr {
		if valueReflect.Kind() != reflect.Ptr {
			// Create pointer and set the value
			ptr := reflect.New(field.Type().Elem())
			if err := pm.setDirect(ptr.Elem(), value); err != nil {
				return err
			}
			field.Set(ptr)
			return nil
		}
	}

	// Direct assignment if types match
	if valueReflect.Type().AssignableTo(field.Type()) {
		field.Set(valueReflect)
		return nil
	}

	// Type conversion if possible
	if valueReflect.Type().ConvertibleTo(field.Type()) {
		field.Set(valueReflect.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign %s to %s", valueReflect.Type(), field.Type())
}

// setTimestamp sets a timestamp field from various input types.
func (pm *PropertyMapper) setTimestamp(field reflect.Value, value interface{}) error {
	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v != nil {
			t = *v
		}
	case string:
		// Try common timestamp formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			t, err = time.Parse(format, v)
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}
	case int64:
		t = time.Unix(v, 0)
	default:
		return fmt.Errorf("unsupported timestamp type: %T", value)
	}

	// Set the field
	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(field.Type().Elem())
		ptr.Elem().Set(reflect.ValueOf(t))
		field.Set(ptr)
	} else {
		field.Set(reflect.ValueOf(t))
	}

	return nil
}

// setJSON sets a field by unmarshaling JSON.
func (pm *PropertyMapper) setJSON(field reflect.Value, value interface{}) error {
	var jsonData []byte

	switch v := value.(type) {
	case string:
		jsonData = []byte(v)
	case []byte:
		jsonData = v
	default:
		// Marshal to JSON first
		var err error
		jsonData, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
	}

	// Create new instance if field is pointer and nil
	if field.Kind() == reflect.Ptr && field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	// Unmarshal into field
	if err := json.Unmarshal(jsonData, field.Addr().Interface()); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// getValue gets a field value with type conversion.
func (pm *PropertyMapper) getValue(field reflect.Value, typeHint string) (interface{}, error) {
	// Handle pointer fields
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil, nil
		}
		field = field.Elem()
	}

	// Handle type conversions based on hint
	switch typeHint {
	case "timestamp":
		return pm.getTimestamp(field)
	case "json":
		return pm.getJSON(field)
	default:
		return field.Interface(), nil
	}
}

// getTimestamp gets a timestamp value in standard format.
func (pm *PropertyMapper) getTimestamp(field reflect.Value) (interface{}, error) {
	if field.Type() != reflect.TypeOf(time.Time{}) {
		return nil, fmt.Errorf("field is not a time.Time")
	}

	t := field.Interface().(time.Time)
	return t.Format(time.RFC3339), nil
}

// getJSON gets a field value as JSON.
func (pm *PropertyMapper) getJSON(field reflect.Value) (interface{}, error) {
	data, err := json.Marshal(field.Interface())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

// GetFieldNames returns a list of field names from mappings.
func (pm *PropertyMapper) GetFieldNames(mappings []config.PropertyMap) []string {
	names := make([]string, len(mappings))
	for i, m := range mappings {
		names[i] = m.Field
	}
	return names
}

// GetObjectFieldNames returns a list of object field names from mappings.
func (pm *PropertyMapper) GetObjectFieldNames(mappings []config.PropertyMap) []string {
	names := make([]string, len(mappings))
	for i, m := range mappings {
		names[i] = m.Object
	}
	return names
}

// ValidateMapping validates that all mapped fields exist in the target struct.
func (pm *PropertyMapper) ValidateMapping(target interface{}, mappings []config.PropertyMap) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct or pointer to struct")
	}

	var missingFields []string
	for _, mapping := range mappings {
		if _, found := targetType.FieldByName(mapping.Object); !found {
			missingFields = append(missingFields, mapping.Object)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("fields not found in struct: %s", strings.Join(missingFields, ", "))
	}

	return nil
}
