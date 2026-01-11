package elasticsearch

import (
	"testing"
	"time"
)

func TestGetResourceCreatedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name        string
		meta        map[string]any
		expectTime  *time.Time
		expectError bool
	}{
		{
			name:        "nil meta",
			meta:        nil,
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "empty meta",
			meta:        map[string]any{},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "missing created_at key",
			meta:        map[string]any{"updated_at": now.Format(time.RFC3339)},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "valid created_at timestamp",
			meta:        map[string]any{"created_at": now.Format(time.RFC3339)},
			expectTime:  &now,
			expectError: false,
		},
		{
			name:        "created_at is not a string",
			meta:        map[string]any{"created_at": 12345},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "invalid timestamp format",
			meta:        map[string]any{"created_at": "not-a-valid-timestamp"},
			expectTime:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetResourceCreatedAt(tt.meta)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expectTime == nil {
				if result != nil {
					t.Errorf("Expected nil time, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil time, got nil")
				} else if !result.Equal(*tt.expectTime) {
					t.Errorf("GetResourceCreatedAt() = %v, want %v", result, tt.expectTime)
				}
			}
		})
	}
}

func TestGetResourceUpdatedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name        string
		meta        map[string]any
		expectTime  *time.Time
		expectError bool
	}{
		{
			name:        "nil meta",
			meta:        nil,
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "empty meta",
			meta:        map[string]any{},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "missing updated_at key",
			meta:        map[string]any{"created_at": now.Format(time.RFC3339)},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "valid updated_at timestamp",
			meta:        map[string]any{"updated_at": now.Format(time.RFC3339)},
			expectTime:  &now,
			expectError: false,
		},
		{
			name:        "updated_at is not a string",
			meta:        map[string]any{"updated_at": 12345},
			expectTime:  nil,
			expectError: false,
		},
		{
			name:        "invalid timestamp format",
			meta:        map[string]any{"updated_at": "invalid-date-format"},
			expectTime:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetResourceUpdatedAt(tt.meta)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expectTime == nil {
				if result != nil {
					t.Errorf("Expected nil time, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil time, got nil")
				} else if !result.Equal(*tt.expectTime) {
					t.Errorf("GetResourceUpdatedAt() = %v, want %v", result, tt.expectTime)
				}
			}
		})
	}
}

func TestGetTimestampFromMeta_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		meta        map[string]any
		key         string
		expectTime  bool
		expectError bool
	}{
		{
			name:        "empty string timestamp",
			meta:        map[string]any{"timestamp": ""},
			key:         "timestamp",
			expectTime:  false,
			expectError: true,
		},
		{
			name:        "whitespace timestamp",
			meta:        map[string]any{"timestamp": "   "},
			key:         "timestamp",
			expectTime:  false,
			expectError: true,
		},
		{
			name:        "partial date format",
			meta:        map[string]any{"timestamp": "2026-01-10"},
			key:         "timestamp",
			expectTime:  false,
			expectError: true,
		},
		{
			name:        "unix timestamp as string",
			meta:        map[string]any{"timestamp": "1736496000"},
			key:         "timestamp",
			expectTime:  false,
			expectError: true,
		},
		{
			name:        "null value",
			meta:        map[string]any{"timestamp": nil},
			key:         "timestamp",
			expectTime:  false,
			expectError: false,
		},
		{
			name:        "boolean value",
			meta:        map[string]any{"timestamp": true},
			key:         "timestamp",
			expectTime:  false,
			expectError: false,
		},
		{
			name:        "float value",
			meta:        map[string]any{"timestamp": 1736496000.0},
			key:         "timestamp",
			expectTime:  false,
			expectError: false,
		},
		{
			name:        "valid RFC3339 with timezone",
			meta:        map[string]any{"timestamp": "2026-01-10T12:00:00+02:00"},
			key:         "timestamp",
			expectTime:  true,
			expectError: false,
		},
		{
			name:        "valid RFC3339 with Z timezone",
			meta:        map[string]any{"timestamp": "2026-01-10T12:00:00Z"},
			key:         "timestamp",
			expectTime:  true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getTimestampFromMeta(tt.meta, tt.key)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expectTime {
				if result == nil {
					t.Error("Expected non-nil time, got nil")
				}
			} else if !tt.expectError {
				if result != nil {
					t.Errorf("Expected nil time, got %v", result)
				}
			}
		})
	}
}
