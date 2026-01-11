package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteSnapshotRepository(t *testing.T) {
	tests := []struct {
		name             string
		repoName         string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			repoName:         "test-repo",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "repository not found",
			repoName:         "nonexistent-repo",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "repository_missing_exception"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			repoName:         "test-repo",
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
				expectedPath := "/_snapshot/" + tt.repoName
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

			result, err := DeleteSnapshotRepository(esClient, tt.repoName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSnapshotRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteSnapshotRepository() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestUpsertSnapshotRepository(t *testing.T) {
	tests := []struct {
		name             string
		repo             v1alpha1.SnapshotRepository
		getStatusCode    int
		getResponse      string
		createStatusCode int
		createResponse   string
		deleteStatusCode int
		deleteResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "create new repository",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "new-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "fs", "settings": {"location": "/backup"}}`,
				},
			},
			getStatusCode:    http.StatusNotFound,
			getResponse:      `{"error": {"type": "repository_missing_exception"}}`,
			createStatusCode: http.StatusOK,
			createResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "update existing repository",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "fs", "settings": {"location": "/new-backup"}}`,
				},
			},
			getStatusCode:    http.StatusOK,
			getResponse:      `{"existing-repo": {"type": "fs", "settings": {"location": "/backup"}}}`,
			deleteStatusCode: http.StatusOK,
			deleteResponse:   `{"acknowledged": true}`,
			createStatusCode: http.StatusOK,
			createResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "create fails",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fail-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "invalid"}`,
				},
			},
			getStatusCode:    http.StatusNotFound,
			getResponse:      `{"error": {"type": "repository_missing_exception"}}`,
			createStatusCode: http.StatusBadRequest,
			createResponse:   `{"error": {"type": "repository_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")

				if r.Method == http.MethodGet {
					w.WriteHeader(tt.getStatusCode)
					w.Write([]byte(tt.getResponse))
					return
				}

				if r.Method == http.MethodDelete {
					w.WriteHeader(tt.deleteStatusCode)
					w.Write([]byte(tt.deleteResponse))
					return
				}

				if r.Method == http.MethodPut {
					requestCount++
					w.WriteHeader(tt.createStatusCode)
					w.Write([]byte(tt.createResponse))
					return
				}
			}))
			defer server.Close()

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			result, err := UpsertSnapshotRepository(esClient, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertSnapshotRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("UpsertSnapshotRepository() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestCreateSnapshotRepository(t *testing.T) {
	tests := []struct {
		name             string
		repo             v1alpha1.SnapshotRepository
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "fs", "settings": {"location": "/backup"}}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "invalid repository type",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "unknown_type"}`,
				},
			},
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "repository_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "server error",
			repo: v1alpha1.SnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.SnapshotRepositorySpec{
					Body: `{"type": "fs", "settings": {"location": "/backup"}}`,
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
				expectedPath := "/_snapshot/" + tt.repo.Name
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

			result, err := createSnapshotRepository(esClient, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("createSnapshotRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("createSnapshotRepository() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}
