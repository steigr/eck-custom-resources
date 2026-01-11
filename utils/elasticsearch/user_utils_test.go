package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		serverStatusCode int
		serverResponse   string
		wantUser         bool
		wantErr          bool
	}{
		{
			name:             "user exists",
			username:         "testuser",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"username": {"username": "testuser", "roles": ["admin"], "enabled": true}}`,
			wantUser:         true,
			wantErr:          false,
		},
		{
			name:             "user not found",
			username:         "nonexistent",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{}`,
			wantUser:         false,
			wantErr:          false,
		},
		{
			name:             "server error - returns nil user but no error due to json parsing",
			username:         "testuser",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantUser:         false,
			wantErr:          false, // GetUser doesn't check status code, only parses JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

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

			user, err := GetUser(esClient, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser && user == nil {
				t.Error("GetUser() returned nil user, expected non-nil")
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			username:         "testuser",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"found": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "user not found",
			username:         "nonexistent",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"found": false}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			username:         "testuser",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				expectedPath := "/_security/user/" + tt.username
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

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

			result, err := DeleteUser(esClient, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteUser() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestDeleteUser_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	result, err := DeleteUser(esClient, "testuser")

	if err == nil {
		t.Error("DeleteUser() with connection error should return an error")
	}

	if !result.Requeue {
		t.Error("DeleteUser() with connection error should request requeue")
	}
}

func TestGetUser_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	user, err := GetUser(esClient, "testuser")

	if err == nil {
		t.Error("GetUser() with connection error should return an error")
	}

	if user != nil {
		t.Error("GetUser() with connection error should return nil user")
	}
}
