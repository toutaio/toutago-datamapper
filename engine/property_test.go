package engine

import (
	"testing"
	"time"

	"github.com/toutago/toutago-datamapper/config"
)

// Test structs
type TestUser struct {
	ID        int
	Name      string
	Email     string
	Age       int
	Active    bool
	CreatedAt time.Time
	UpdatedAt *time.Time
	Metadata  map[string]interface{}
}

type TestProfile struct {
	UserID   int
	Bio      string
	Location string
}

func TestPropertyMapper_MapToObject(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{
		"id":    123,
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
		{Object: "Age", Field: "age"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	if user.ID != 123 {
		t.Errorf("ID = %v, want 123", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Name = %v, want John Doe", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Email = %v, want john@example.com", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("Age = %v, want 30", user.Age)
	}
}

func TestPropertyMapper_MapToObject_NilTarget(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{"id": 1}
	mappings := []config.PropertyMap{{Object: "ID", Field: "id"}}

	err := pm.MapToObject(data, nil, mappings)
	if err == nil {
		t.Error("MapToObject() should error for nil target")
	}
}

func TestPropertyMapper_MapToObject_NonPointerTarget(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{"id": 1}
	mappings := []config.PropertyMap{{Object: "ID", Field: "id"}}

	var user TestUser
	err := pm.MapToObject(data, user, mappings) // Not a pointer
	if err == nil {
		t.Error("MapToObject() should error for non-pointer target")
	}
}

func TestPropertyMapper_MapToObject_InvalidField(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{"unknown": 1}
	mappings := []config.PropertyMap{{Object: "NonExistent", Field: "unknown"}}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err == nil {
		t.Error("MapToObject() should error for non-existent field")
	}
}

func TestPropertyMapper_MapFromObject(t *testing.T) {
	pm := NewPropertyMapper()

	user := TestUser{
		ID:    456,
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
		{Object: "Age", Field: "age"},
	}

	data, err := pm.MapFromObject(user, mappings)
	if err != nil {
		t.Fatalf("MapFromObject() error = %v", err)
	}

	if data["id"] != 456 {
		t.Errorf("id = %v, want 456", data["id"])
	}
	if data["name"] != "Jane Doe" {
		t.Errorf("name = %v, want Jane Doe", data["name"])
	}
	if data["email"] != "jane@example.com" {
		t.Errorf("email = %v, want jane@example.com", data["email"])
	}
	if data["age"] != 25 {
		t.Errorf("age = %v, want 25", data["age"])
	}
}

func TestPropertyMapper_MapFromObject_Pointer(t *testing.T) {
	pm := NewPropertyMapper()

	user := &TestUser{
		ID:   789,
		Name: "Bob Smith",
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
	}

	data, err := pm.MapFromObject(user, mappings)
	if err != nil {
		t.Fatalf("MapFromObject() error = %v", err)
	}

	if data["id"] != 789 {
		t.Errorf("id = %v, want 789", data["id"])
	}
	if data["name"] != "Bob Smith" {
		t.Errorf("name = %v, want Bob Smith", data["name"])
	}
}

func TestPropertyMapper_MapFromObject_SkipGenerated(t *testing.T) {
	pm := NewPropertyMapper()

	user := TestUser{
		ID:   999,
		Name: "Test",
	}

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id", Generated: true}, // Should be skipped
		{Object: "Name", Field: "name"},
	}

	data, err := pm.MapFromObject(user, mappings)
	if err != nil {
		t.Fatalf("MapFromObject() error = %v", err)
	}

	if _, exists := data["id"]; exists {
		t.Error("Generated field 'id' should be skipped")
	}
	if data["name"] != "Test" {
		t.Errorf("name = %v, want Test", data["name"])
	}
}

func TestPropertyMapper_Timestamp(t *testing.T) {
	pm := NewPropertyMapper()

	now := time.Now()
	data := map[string]interface{}{
		"created_at": now,
	}

	mappings := []config.PropertyMap{
		{Object: "CreatedAt", Field: "created_at", Type: "timestamp"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	if !user.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", user.CreatedAt, now)
	}
}

func TestPropertyMapper_TimestampFromString(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{
		"created_at": "2024-01-15T10:30:00Z",
	}

	mappings := []config.PropertyMap{
		{Object: "CreatedAt", Field: "created_at", Type: "timestamp"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	expected, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if !user.CreatedAt.Equal(expected) {
		t.Errorf("CreatedAt = %v, want %v", user.CreatedAt, expected)
	}
}

func TestPropertyMapper_TimestampFromUnix(t *testing.T) {
	pm := NewPropertyMapper()

	unixTime := int64(1705315800)
	data := map[string]interface{}{
		"created_at": unixTime,
	}

	mappings := []config.PropertyMap{
		{Object: "CreatedAt", Field: "created_at", Type: "timestamp"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	expected := time.Unix(unixTime, 0)
	if !user.CreatedAt.Equal(expected) {
		t.Errorf("CreatedAt = %v, want %v", user.CreatedAt, expected)
	}
}

func TestPropertyMapper_JSON(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{
		"metadata": `{"role":"admin","level":5}`,
	}

	mappings := []config.PropertyMap{
		{Object: "Metadata", Field: "metadata", Type: "json"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	if user.Metadata["role"] != "admin" {
		t.Errorf("Metadata[role] = %v, want admin", user.Metadata["role"])
	}
	if user.Metadata["level"] != float64(5) {
		t.Errorf("Metadata[level] = %v, want 5", user.Metadata["level"])
	}
}

func TestPropertyMapper_PointerField(t *testing.T) {
	pm := NewPropertyMapper()

	now := time.Now()
	data := map[string]interface{}{
		"updated_at": now,
	}

	mappings := []config.PropertyMap{
		{Object: "UpdatedAt", Field: "updated_at", Type: "timestamp"},
	}

	var user TestUser
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	if user.UpdatedAt == nil {
		t.Fatal("UpdatedAt should not be nil")
	}
	if !user.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", *user.UpdatedAt, now)
	}
}

func TestPropertyMapper_NilValue(t *testing.T) {
	pm := NewPropertyMapper()

	data := map[string]interface{}{
		"name": nil,
	}

	mappings := []config.PropertyMap{
		{Object: "Name", Field: "name"},
	}

	user := TestUser{Name: "Original"}
	err := pm.MapToObject(data, &user, mappings)
	if err != nil {
		t.Fatalf("MapToObject() error = %v", err)
	}

	if user.Name != "" {
		t.Errorf("Name should be empty string (zero value), got %v", user.Name)
	}
}

func TestPropertyMapper_GetFieldNames(t *testing.T) {
	pm := NewPropertyMapper()

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
	}

	names := pm.GetFieldNames(mappings)
	expected := []string{"id", "name", "email"}

	if len(names) != len(expected) {
		t.Fatalf("len(names) = %d, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %v, want %v", i, name, expected[i])
		}
	}
}

func TestPropertyMapper_GetObjectFieldNames(t *testing.T) {
	pm := NewPropertyMapper()

	mappings := []config.PropertyMap{
		{Object: "ID", Field: "id"},
		{Object: "Name", Field: "name"},
		{Object: "Email", Field: "email"},
	}

	names := pm.GetObjectFieldNames(mappings)
	expected := []string{"ID", "Name", "Email"}

	if len(names) != len(expected) {
		t.Fatalf("len(names) = %d, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %v, want %v", i, name, expected[i])
		}
	}
}

func TestPropertyMapper_ValidateMapping(t *testing.T) {
	pm := NewPropertyMapper()

	tests := []struct {
		name     string
		target   interface{}
		mappings []config.PropertyMap
		wantErr  bool
	}{
		{
			name:   "valid mapping",
			target: &TestUser{},
			mappings: []config.PropertyMap{
				{Object: "ID", Field: "id"},
				{Object: "Name", Field: "name"},
			},
			wantErr: false,
		},
		{
			name:   "invalid field",
			target: &TestUser{},
			mappings: []config.PropertyMap{
				{Object: "NonExistent", Field: "something"},
			},
			wantErr: true,
		},
		{
			name:   "struct without pointer",
			target: TestUser{},
			mappings: []config.PropertyMap{
				{Object: "ID", Field: "id"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateMapping(tt.target, tt.mappings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
