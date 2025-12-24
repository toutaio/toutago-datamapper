package adapter

import (
	"errors"
	"testing"
)

func TestOperationType(t *testing.T) {
	tests := []struct {
		name string
		op   OperationType
		want string
	}{
		{"fetch", OpFetch, "fetch"},
		{"insert", OpInsert, "insert"},
		{"update", OpUpdate, "update"},
		{"delete", OpDelete, "delete"},
		{"action", OpAction, "action"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("OperationType: got %v, want %v", tt.op, tt.want)
			}
		})
	}
}

func TestAdapterError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AdapterError
		want string
	}{
		{
			name: "error without cause",
			err: &AdapterError{
				Code:    "TEST_ERROR",
				Message: "test message",
			},
			want: "TEST_ERROR: test message",
		},
		{
			name: "error with cause",
			err: &AdapterError{
				Code:    "TEST_ERROR",
				Message: "test message",
				Cause:   errors.New("underlying error"),
			},
			want: "TEST_ERROR: test message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("AdapterError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AdapterError{
		Code:    "TEST",
		Message: "test",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test errors.Is
	if !errors.Is(err, cause) {
		t.Error("errors.Is() should return true for cause")
	}
}

func TestNewAdapterError(t *testing.T) {
	cause := errors.New("underlying")
	err := NewAdapterError("CODE", "message", cause)

	if err.Code != "CODE" {
		t.Errorf("Code = %v, want CODE", err.Code)
	}
	if err.Message != "message" {
		t.Errorf("Message = %v, want message", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AdapterError
		code string
	}{
		{"ErrNotFound", ErrNotFound, "NOT_FOUND"},
		{"ErrValidation", ErrValidation, "VALIDATION"},
		{"ErrConnection", ErrConnection, "CONNECTION"},
		{"ErrAdapter", ErrAdapter, "ADAPTER"},
		{"ErrConfiguration", ErrConfiguration, "CONFIGURATION"},
		{"ErrConflict", ErrConflict, "CONFLICT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
			if tt.err.Message == "" {
				t.Errorf("%s.Message is empty", tt.name)
			}
		})
	}
}

func TestOperation_Structure(t *testing.T) {
	// Test that Operation struct can be created with all fields
	op := &Operation{
		Type:      OpFetch,
		Statement: "SELECT * FROM users",
		Properties: []PropertyMapping{
			{ObjectField: "ID", DataField: "id"},
		},
		Identifier: []PropertyMapping{
			{ObjectField: "ID", DataField: "id"},
		},
		Generated: []PropertyMapping{
			{ObjectField: "CreatedAt", DataField: "created_at", Type: "timestamp"},
		},
		Condition: []PropertyMapping{
			{ObjectField: "Version", DataField: "version"},
		},
		Bulk:   false,
		Multi:  true,
		Source: "main-db",
		Fallback: &Operation{
			Type:      OpFetch,
			Statement: "SELECT * FROM users_cache",
		},
		After: []AfterAction{
			{Type: "invalidate", Source: "cache", Statement: "user:{id}"},
		},
	}

	if op.Type != OpFetch {
		t.Error("Failed to set Type")
	}
	if op.Statement != "SELECT * FROM users" {
		t.Error("Failed to set Statement")
	}
	if len(op.Properties) != 1 {
		t.Error("Failed to set Properties")
	}
	if op.Multi != true {
		t.Error("Failed to set Multi")
	}
	if op.Fallback == nil {
		t.Error("Failed to set Fallback")
	}
	if len(op.After) != 1 {
		t.Error("Failed to set After actions")
	}
}

func TestPropertyMapping_Structure(t *testing.T) {
	pm := PropertyMapping{
		ObjectField: "CreatedAt",
		DataField:   "created_at",
		Type:        "timestamp",
		Generated:   true,
	}

	if pm.ObjectField != "CreatedAt" {
		t.Error("Failed to set ObjectField")
	}
	if pm.DataField != "created_at" {
		t.Error("Failed to set DataField")
	}
	if pm.Type != "timestamp" {
		t.Error("Failed to set Type")
	}
	if pm.Generated != true {
		t.Error("Failed to set Generated")
	}
}

func TestAction_Structure(t *testing.T) {
	action := Action{
		Name:      "get-user-stats",
		Statement: "CALL GetUserStats(?)",
		Parameters: []PropertyMapping{
			{ObjectField: "ID", DataField: "user_id"},
		},
		Result: &ResultMapping{
			Type:  "UserStats",
			Multi: false,
			Properties: []PropertyMapping{
				{ObjectField: "LoginCount", DataField: "login_count"},
			},
		},
	}

	if action.Name != "get-user-stats" {
		t.Error("Failed to set Name")
	}
	if action.Result == nil {
		t.Error("Failed to set Result")
	}
	if action.Result.Type != "UserStats" {
		t.Error("Failed to set Result.Type")
	}
}
