package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteRole(t *testing.T) {
	tests := []struct {
		name             string
		roleName         string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			roleName:         "test-role",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"found": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "role not found",
			roleName:         "nonexistent-role",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"found": false}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			roleName:         "test-role",
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
				expectedPath := "/_security/role/" + tt.roleName
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

			result, err := DeleteRole(esClient, tt.roleName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteRole() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestUpsertRole(t *testing.T) {
	tests := []struct {
		name             string
		role             v1alpha1.ElasticsearchRole
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			role: v1alpha1.ElasticsearchRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-role",
				},
				Spec: v1alpha1.ElasticsearchRoleSpec{
					Body: `{"cluster": ["monitor"], "indices": [{"names": ["*"], "privileges": ["read"]}]}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"role": {"created": true}}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "successful update",
			role: v1alpha1.ElasticsearchRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-role",
				},
				Spec: v1alpha1.ElasticsearchRoleSpec{
					Body: `{"cluster": ["all"], "indices": [{"names": ["logs-*"], "privileges": ["all"]}]}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"role": {"created": false}}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "invalid role body",
			role: v1alpha1.ElasticsearchRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-role",
				},
				Spec: v1alpha1.ElasticsearchRoleSpec{
					Body: `{"invalid": "privileges"}`,
				},
			},
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "illegal_argument_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "server error",
			role: v1alpha1.ElasticsearchRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-role",
				},
				Spec: v1alpha1.ElasticsearchRoleSpec{
					Body: `{"cluster": ["monitor"]}`,
				},
			},
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				expectedPath := "/_security/role/" + tt.role.Name
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

			result, err := UpsertRole(esClient, tt.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("UpsertRole() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}
