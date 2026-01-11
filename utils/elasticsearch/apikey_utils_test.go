package elasticsearch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestValidateExpiration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
		wantValue string
	}{
		{
			name:      "valid -1 (never expires)",
			input:     "-1",
			wantValid: true,
			wantValue: "-1",
		},
		{
			name:      "valid 0 (immediate expiration)",
			input:     "0",
			wantValid: true,
			wantValue: "0",
		},
		{
			name:      "valid seconds",
			input:     "30s",
			wantValid: true,
			wantValue: "30s",
		},
		{
			name:      "valid minutes",
			input:     "30m",
			wantValid: true,
			wantValue: "30m",
		},
		{
			name:      "valid hours",
			input:     "12h",
			wantValid: true,
			wantValue: "12h",
		},
		{
			name:      "valid days",
			input:     "7d",
			wantValid: true,
			wantValue: "7d",
		},
		{
			name:      "valid milliseconds",
			input:     "1000ms",
			wantValid: true,
			wantValue: "1000ms",
		},
		{
			name:      "valid microseconds",
			input:     "1000micros",
			wantValid: true,
			wantValue: "1000micros",
		},
		{
			name:      "valid nanoseconds",
			input:     "1000nanos",
			wantValid: true,
			wantValue: "1000nanos",
		},
		{
			name:      "valid decimal hours",
			input:     "1.5h",
			wantValid: true,
			wantValue: "1.5h",
		},
		{
			name:      "valid uppercase",
			input:     "30M",
			wantValid: true,
			wantValue: "30m",
		},
		{
			name:      "valid with whitespace",
			input:     "  30m  ",
			wantValid: true,
			wantValue: "30m",
		},
		{
			name:      "invalid - empty string",
			input:     "",
			wantValid: false,
		},
		{
			name:      "invalid - just number",
			input:     "30",
			wantValid: false,
		},
		{
			name:      "invalid - invalid unit",
			input:     "30x",
			wantValid: false,
		},
		{
			name:      "invalid - negative number with unit",
			input:     "-30m",
			wantValid: false,
		},
		{
			name:      "invalid - string",
			input:     "never",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateExpiration(tt.input)

			if tt.wantValid {
				if err != nil {
					t.Errorf("validateExpiration(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.wantValue {
					t.Errorf("validateExpiration(%q) = %v, want %v", tt.input, result, tt.wantValue)
				}
			} else {
				if err == nil {
					t.Errorf("validateExpiration(%q) expected error, got nil", tt.input)
				}
			}
		})
	}
}

func TestApiKeyNameExist(t *testing.T) {
	tests := []struct {
		name             string
		apiKeyName       string
		serverStatusCode int
		serverResponse   string
		wantExists       bool
	}{
		{
			name:             "api key exists",
			apiKeyName:       "test-key",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": [{"id": "key123", "name": "test-key"}]}`,
			wantExists:       true,
		},
		{
			name:             "api key does not exist",
			apiKeyName:       "nonexistent-key",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": []}`,
			wantExists:       false,
		},
		{
			name:             "server error",
			apiKeyName:       "test-key",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantExists:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			exists := ApiKeyNameExist(nil, nil, esClient, tt.apiKeyName)

			if exists != tt.wantExists {
				t.Errorf("ApiKeyNameExist() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestGetApiKeyWithName(t *testing.T) {
	tests := []struct {
		name             string
		apiKeyName       string
		serverStatusCode int
		serverResponse   string
		wantKeys         int
	}{
		{
			name:             "single key found",
			apiKeyName:       "test-key",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": [{"id": "key123", "name": "test-key"}]}`,
			wantKeys:         1,
		},
		{
			name:             "multiple keys found",
			apiKeyName:       "test-key",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": [{"id": "key123", "name": "test-key"}, {"id": "key456", "name": "test-key"}]}`,
			wantKeys:         2,
		},
		{
			name:             "no keys found",
			apiKeyName:       "nonexistent-key",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": []}`,
			wantKeys:         0,
		},
		{
			name:             "server error",
			apiKeyName:       "test-key",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantKeys:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			keys := GetApiKeyWithName(nil, nil, esClient, tt.apiKeyName)

			if len(keys) != tt.wantKeys {
				t.Errorf("GetApiKeyWithName() returned %d keys, want %d", len(keys), tt.wantKeys)
			}
		})
	}
}

func TestGetApiKeyWithID(t *testing.T) {
	tests := []struct {
		name             string
		apiKeyID         string
		serverStatusCode int
		serverResponse   string
		wantKey          bool
	}{
		{
			name:             "key found",
			apiKeyID:         "key123",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"api_keys": [{"id": "key123", "name": "test-key"}]}`,
			wantKey:          true,
		},
		{
			name:             "key not found - returns empty key",
			apiKeyID:         "nonexistent",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"api_keys": []}`,
			wantKey:          false,
		},
		{
			name:             "server error - returns empty key",
			apiKeyID:         "key123",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantKey:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			key, _ := GetApiKeyWithID(nil, nil, esClient, tt.apiKeyID)

			if tt.wantKey && key.ID == "" {
				t.Error("GetApiKeyWithID() returned empty key, expected non-empty")
			}

			if !tt.wantKey && key.ID != "" {
				t.Errorf("GetApiKeyWithID() returned key %v, expected empty", key)
			}
		})
	}
}

func TestContainsID(t *testing.T) {
	tests := []struct {
		name     string
		apiKeys  []APIKey
		id       string
		wantBool bool
	}{
		{
			name:     "id exists in list",
			apiKeys:  []APIKey{{ID: "123", Name: "key1"}, {ID: "456", Name: "key2"}},
			id:       "123",
			wantBool: true,
		},
		{
			name:     "id does not exist in list",
			apiKeys:  []APIKey{{ID: "123", Name: "key1"}, {ID: "456", Name: "key2"}},
			id:       "789",
			wantBool: false,
		},
		{
			name:     "empty list",
			apiKeys:  []APIKey{},
			id:       "123",
			wantBool: false,
		},
		{
			name:     "nil list",
			apiKeys:  nil,
			id:       "123",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsID(tt.apiKeys, tt.id)
			if result != tt.wantBool {
				t.Errorf("containsID() = %v, want %v", result, tt.wantBool)
			}
		})
	}
}

func TestRemoveField(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		field   string
		wantErr bool
	}{
		{
			name:    "remove existing field",
			json:    `{"name": "test", "value": "something"}`,
			field:   "name",
			wantErr: false,
		},
		{
			name:    "remove non-existing field",
			json:    `{"name": "test", "value": "something"}`,
			field:   "nonexistent",
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid json}`,
			field:   "name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := removeField(tt.json, tt.field)

			if (err != nil) != tt.wantErr {
				t.Errorf("removeField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == "" {
				t.Error("removeField() returned empty string for valid input")
			}
		})
	}
}

func TestRemoveField_FieldActuallyRemoved(t *testing.T) {
	input := `{"name": "test", "value": "something", "id": "123"}`
	result, err := removeField(input, "id")
	if err != nil {
		t.Fatalf("removeField() unexpected error: %v", err)
	}

	// Check that "id" is not in the result
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if _, exists := data["id"]; exists {
		t.Error("removeField() did not remove the field")
	}

	// Check that other fields are still present
	if _, exists := data["name"]; !exists {
		t.Error("removeField() removed field that should be kept")
	}
	if _, exists := data["value"]; !exists {
		t.Error("removeField() removed field that should be kept")
	}
}

func TestApiKeyNameExist_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	exists := ApiKeyNameExist(nil, nil, esClient, "test-key")

	if exists {
		t.Error("ApiKeyNameExist() with connection error should return false")
	}
}

func TestGetApiKeyWithName_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	keys := GetApiKeyWithName(nil, nil, esClient, "test-key")

	if keys != nil {
		t.Error("GetApiKeyWithName() with connection error should return nil")
	}
}

func TestGetApiKeyWithID_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	_, err = GetApiKeyWithID(nil, nil, esClient, "key123")

	if err == nil {
		t.Error("GetApiKeyWithID() with connection error should return an error")
	}
}
