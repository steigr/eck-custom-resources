package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteSnapshotLifecyclePolicy(t *testing.T) {
	tests := []struct {
		name             string
		policyName       string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			policyName:       "test-policy",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "policy not found",
			policyName:       "nonexistent-policy",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "resource_not_found_exception"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			policyName:       "test-policy",
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
				expectedPath := "/_slm/policy/" + tt.policyName
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

			result, err := DeleteSnapshotLifecyclePolicy(esClient, tt.policyName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSnapshotLifecyclePolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteSnapshotLifecyclePolicy() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestUpsertSnapshotLifecyclePolicy(t *testing.T) {
	tests := []struct {
		name             string
		policy           v1alpha1.SnapshotLifecyclePolicy
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			policy: v1alpha1.SnapshotLifecyclePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: v1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{"schedule": "0 30 1 * * ?", "name": "<daily-snap-{now/d}>", "repository": "my_repository"}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "successful update",
			policy: v1alpha1.SnapshotLifecyclePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-policy",
				},
				Spec: v1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{"schedule": "0 0 2 * * ?", "name": "<weekly-snap-{now/d}>", "repository": "my_repository"}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "invalid policy body",
			policy: v1alpha1.SnapshotLifecyclePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-policy",
				},
				Spec: v1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{"invalid": "policy"}`,
				},
			},
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "illegal_argument_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "server error",
			policy: v1alpha1.SnapshotLifecyclePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: v1alpha1.SnapshotLifecyclePolicySpec{
					Body: `{"schedule": "0 30 1 * * ?", "name": "<snap-{now/d}>", "repository": "repo"}`,
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
				expectedPath := "/_slm/policy/" + tt.policy.Name
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

			result, err := UpsertSnapshotLifecyclePolicy(esClient, tt.policy)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertSnapshotLifecyclePolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("UpsertSnapshotLifecyclePolicy() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}
