package kibana

import (
	"testing"
)

func TestInjectId(t *testing.T) {
	tests := []struct {
		name       string
		objectJson string
		id         string
		wantErr    bool
		wantId     string
	}{
		{
			name:       "inject id into simple object",
			objectJson: `{"name": "test"}`,
			id:         "my-id",
			wantErr:    false,
			wantId:     "my-id",
		},
		{
			name:       "inject id into empty object",
			objectJson: `{}`,
			id:         "empty-obj-id",
			wantErr:    false,
			wantId:     "empty-obj-id",
		},
		{
			name:       "inject id overwrites existing id",
			objectJson: `{"id": "old-id", "name": "test"}`,
			id:         "new-id",
			wantErr:    false,
			wantId:     "new-id",
		},
		{
			name:       "inject id into complex object",
			objectJson: `{"name": "test", "nested": {"key": "value"}, "array": [1, 2, 3]}`,
			id:         "complex-id",
			wantErr:    false,
			wantId:     "complex-id",
		},
		{
			name:       "invalid json returns error",
			objectJson: `{invalid json}`,
			id:         "test-id",
			wantErr:    true,
		},
		{
			name:       "empty json string returns error",
			objectJson: ``,
			id:         "test-id",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InjectId(tt.objectJson, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("InjectId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if result == nil {
				t.Error("InjectId() returned nil result for valid input")
				return
			}

			// Verify the result contains the expected id
			if !contains(*result, `"id":"`) && !contains(*result, `"id": "`) {
				t.Errorf("InjectId() result does not contain id field: %s", *result)
			}

			if !contains(*result, tt.wantId) {
				t.Errorf("InjectId() result does not contain expected id value %q: %s", tt.wantId, *result)
			}
		})
	}
}

func TestInjectId_PreservesOtherFields(t *testing.T) {
	objectJson := `{"name": "test-name", "description": "test description"}`
	id := "injected-id"

	result, err := InjectId(objectJson, id)
	if err != nil {
		t.Fatalf("InjectId() unexpected error: %v", err)
	}

	// Verify original fields are preserved
	if !contains(*result, "test-name") {
		t.Error("InjectId() did not preserve 'name' field")
	}
	if !contains(*result, "test description") {
		t.Error("InjectId() did not preserve 'description' field")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
