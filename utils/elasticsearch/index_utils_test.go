package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVerifyIndexExists(t *testing.T) {
	tests := []struct {
		name             string
		indexName        string
		serverStatusCode int
		wantExists       bool
		wantErr          bool
	}{
		{
			name:             "index exists",
			indexName:        "test-index",
			serverStatusCode: http.StatusOK,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "index does not exist",
			indexName:        "nonexistent-index",
			serverStatusCode: http.StatusNotFound,
			wantExists:       false,
			wantErr:          false,
		},
		{
			name:             "server error",
			indexName:        "test-index",
			serverStatusCode: http.StatusInternalServerError,
			wantExists:       false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("Expected HEAD request, got %s", r.Method)
				}
				expectedPath := "/" + tt.indexName
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
			}))
			defer server.Close()

			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			exists, err := VerifyIndexExists(esClient, tt.indexName)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyIndexExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("VerifyIndexExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestVerifyIndexEmpty(t *testing.T) {
	tests := []struct {
		name             string
		indexName        string
		serverStatusCode int
		serverResponse   string
		wantEmpty        bool
		wantErr          bool
	}{
		{
			name:             "index is empty",
			indexName:        "empty-index",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"count": 0, "_shards": {"total": 1, "successful": 1, "skipped": 0, "failed": 0}}`,
			wantEmpty:        true,
			wantErr:          false,
		},
		{
			name:             "index has documents",
			indexName:        "populated-index",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"count": 100, "_shards": {"total": 1, "successful": 1, "skipped": 0, "failed": 0}}`,
			wantEmpty:        false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/" + tt.indexName + "/_count"
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

			empty, err := VerifyIndexEmpty(esClient, tt.indexName)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyIndexEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && empty != tt.wantEmpty {
				t.Errorf("VerifyIndexEmpty() = %v, want %v", empty, tt.wantEmpty)
			}
		})
	}
}

func TestDeleteIndex(t *testing.T) {
	tests := []struct {
		name             string
		indexName        string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			indexName:        "test-index",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "index not found",
			indexName:        "nonexistent-index",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "index_not_found_exception"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			indexName:        "test-index",
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
				expectedPath := "/" + tt.indexName
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

			result, err := DeleteIndex(esClient, tt.indexName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteIndex() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		name             string
		index            v1alpha1.Index
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			index: v1alpha1.Index{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-index",
				},
				Spec: v1alpha1.IndexSpec{
					Body: `{"settings": {"number_of_shards": 1}, "mappings": {"properties": {"field1": {"type": "text"}}}}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true, "shards_acknowledged": true, "index": "test-index"}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "index already exists",
			index: v1alpha1.Index{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-index",
				},
				Spec: v1alpha1.IndexSpec{
					Body: `{"settings": {"number_of_shards": 1}}`,
				},
			},
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "resource_already_exists_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "invalid index body",
			index: v1alpha1.Index{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-index",
				},
				Spec: v1alpha1.IndexSpec{
					Body: `{"invalid": "body"}`,
				},
			},
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "mapper_parsing_exception"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "server error",
			index: v1alpha1.Index{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-index",
				},
				Spec: v1alpha1.IndexSpec{
					Body: `{"settings": {}}`,
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
				expectedPath := "/" + tt.index.Name
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

			result, err := CreateIndex(esClient, tt.index)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("CreateIndex() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestVerifyIndexExists_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	exists, err := VerifyIndexExists(esClient, "test-index")

	if err == nil {
		t.Error("VerifyIndexExists() with connection error should return an error")
	}

	if exists {
		t.Error("VerifyIndexExists() with connection error should return false")
	}
}

func TestDeleteIndex_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	result, err := DeleteIndex(esClient, "test-index")

	if err == nil {
		t.Error("DeleteIndex() with connection error should return an error")
	}

	if !result.Requeue {
		t.Error("DeleteIndex() with connection error should request requeue")
	}
}

func TestCreateIndex_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	index := v1alpha1.Index{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-index",
		},
		Spec: v1alpha1.IndexSpec{
			Body: `{"settings": {}}`,
		},
	}

	result, err := CreateIndex(esClient, index)

	if err == nil {
		t.Error("CreateIndex() with connection error should return an error")
	}

	if !result.Requeue {
		t.Error("CreateIndex() with connection error should request requeue")
	}
}
