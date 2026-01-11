package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteIngestPipeline(t *testing.T) {
	tests := []struct {
		name             string
		pipelineId       string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			pipelineId:       "test-pipeline",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "pipeline not found",
			pipelineId:       "nonexistent-pipeline",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "resource_not_found_exception"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			pipelineId:       "test-pipeline",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request method and path
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				expectedPath := "/_ingest/pipeline/" + tt.pipelineId
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create an Elasticsearch client pointing to the test server
			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			result, err := DeleteIngestPipeline(esClient, tt.pipelineId)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteIngestPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteIngestPipeline() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestUpsertIngestPipeline(t *testing.T) {
	tests := []struct {
		name             string
		pipeline         v1alpha1.IngestPipeline
		body             string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			pipeline: v1alpha1.IngestPipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pipeline",
					Namespace: "default",
				},
			},
			body:             `{"description": "Test pipeline", "processors": []}`,
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "successful update",
			pipeline: v1alpha1.IngestPipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing-pipeline",
					Namespace: "default",
				},
			},
			body:             `{"description": "Updated pipeline", "processors": [{"set": {"field": "foo", "value": "bar"}}]}`,
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "invalid pipeline body",
			pipeline: v1alpha1.IngestPipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pipeline",
					Namespace: "default",
				},
			},
			body:             `{"invalid": "body"}`,
			serverStatusCode: http.StatusBadRequest,
			serverResponse:   `{"error": {"type": "parse_exception", "reason": "No processor type exists"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "server error",
			pipeline: v1alpha1.IngestPipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pipeline",
					Namespace: "default",
				},
			},
			body:             `{"description": "Test", "processors": []}`,
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantRequeue:      true,
			wantErr:          true,
		},
		{
			name: "pipeline with processors",
			pipeline: v1alpha1.IngestPipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pipeline-with-processors",
					Namespace: "default",
				},
			},
			body:             `{"description": "Pipeline with processors", "processors": [{"grok": {"field": "message", "patterns": ["%{GREEDYDATA:data}"]}}]}`,
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody string

			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request method and path
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				expectedPath := "/_ingest/pipeline/" + tt.pipeline.Name
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Read the request body
				buf := make([]byte, r.ContentLength)
				r.Body.Read(buf)
				receivedBody = string(buf)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.WriteHeader(tt.serverStatusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create an Elasticsearch client pointing to the test server
			esClient, err := elasticsearch.NewClient(elasticsearch.Config{
				Addresses: []string{server.URL},
			})
			if err != nil {
				t.Fatalf("Failed to create ES client: %v", err)
			}

			result, err := UpsertIngestPipeline(esClient, tt.pipeline, tt.body)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertIngestPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("UpsertIngestPipeline() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}

			// Verify the body was sent correctly
			if receivedBody != tt.body {
				t.Errorf("UpsertIngestPipeline() sent body = %v, want %v", receivedBody, tt.body)
			}
		})
	}
}

func TestUpsertIngestPipeline_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "parse_exception", "reason": "request body is required"}}`))
	}))
	defer server.Close()

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{server.URL},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	pipeline := v1alpha1.IngestPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "empty-body-pipeline",
			Namespace: "default",
		},
	}

	result, err := UpsertIngestPipeline(esClient, pipeline, "")

	if err == nil {
		t.Error("UpsertIngestPipeline() with empty body should return an error")
	}

	if !result.Requeue {
		t.Error("UpsertIngestPipeline() with empty body should request requeue")
	}
}

func TestDeleteIngestPipeline_ConnectionError(t *testing.T) {
	// Create a client with an invalid address to simulate connection error
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	result, err := DeleteIngestPipeline(esClient, "test-pipeline")

	if err == nil {
		t.Error("DeleteIngestPipeline() with connection error should return an error")
	}

	if !result.Requeue {
		t.Error("DeleteIngestPipeline() with connection error should request requeue")
	}
}

func TestUpsertIngestPipeline_ConnectionError(t *testing.T) {
	// Create a client with an invalid address to simulate connection error
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	pipeline := v1alpha1.IngestPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pipeline",
			Namespace: "default",
		},
	}

	result, err := UpsertIngestPipeline(esClient, pipeline, `{"processors": []}`)

	if err == nil {
		t.Error("UpsertIngestPipeline() with connection error should return an error")
	}

	if !result.Requeue {
		t.Error("UpsertIngestPipeline() with connection error should request requeue")
	}
}

func TestGetIngestPipeline(t *testing.T) {
	tests := []struct {
		name             string
		pipelineId       string
		serverStatusCode int
		serverResponse   string
		wantPipeline     bool
		wantErr          bool
	}{
		{
			name:             "pipeline exists",
			pipelineId:       "test-pipeline",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"test-pipeline": {"description": "Test pipeline", "processors": [], "_meta": {"created_at": "2026-01-10T12:00:00Z"}}}`,
			wantPipeline:     true,
			wantErr:          false,
		},
		{
			name:             "pipeline not found",
			pipelineId:       "nonexistent-pipeline",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "resource_not_found_exception"}}`,
			wantPipeline:     false,
			wantErr:          true,
		},
		{
			name:             "server error",
			pipelineId:       "test-pipeline",
			serverStatusCode: http.StatusInternalServerError,
			serverResponse:   `{"error": {"type": "internal_server_error"}}`,
			wantPipeline:     false,
			wantErr:          true,
		},
		{
			name:             "pipeline with meta",
			pipelineId:       "meta-pipeline",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"meta-pipeline": {"description": "Pipeline with meta", "processors": [{"set": {"field": "test", "value": "value"}}], "_meta": {"created_at": "2026-01-10T10:00:00Z", "updated_at": "2026-01-10T12:00:00Z"}}}`,
			wantPipeline:     true,
			wantErr:          false,
		},
		{
			name:             "empty response",
			pipelineId:       "empty-pipeline",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{}`,
			wantPipeline:     false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				expectedPath := "/_ingest/pipeline/" + tt.pipelineId
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

			pipeline, err := GetIngestPipeline(esClient, tt.pipelineId)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetIngestPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantPipeline && pipeline == nil {
				t.Error("GetIngestPipeline() returned nil pipeline, expected non-nil")
			}

			if !tt.wantPipeline && pipeline != nil {
				t.Errorf("GetIngestPipeline() returned pipeline %v, expected nil", pipeline)
			}
		})
	}
}

func TestGetIngestPipeline_ParsesMetaCorrectly(t *testing.T) {
	pipelineId := "test-pipeline"
	serverResponse := `{
		"test-pipeline": {
			"description": "Test pipeline",
			"processors": [{"set": {"field": "test", "value": "value"}}],
			"on_failure": [{"set": {"field": "error", "value": "{{_ingest.on_failure_message}}"}}],
			"version": 1,
			"_meta": {
				"created_at": "2026-01-10T10:00:00Z",
				"updated_at": "2026-01-10T12:00:00Z"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(serverResponse))
	}))
	defer server.Close()

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{server.URL},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	pipeline, err := GetIngestPipeline(esClient, pipelineId)
	if err != nil {
		t.Fatalf("GetIngestPipeline() unexpected error: %v", err)
	}

	if pipeline == nil {
		t.Fatal("GetIngestPipeline() returned nil pipeline")
	}

	if pipeline.Description != "Test pipeline" {
		t.Errorf("GetIngestPipeline() Description = %v, want %v", pipeline.Description, "Test pipeline")
	}

	if len(pipeline.Processors) != 1 {
		t.Errorf("GetIngestPipeline() Processors length = %v, want %v", len(pipeline.Processors), 1)
	}

	if len(pipeline.OnFailure) != 1 {
		t.Errorf("GetIngestPipeline() OnFailure length = %v, want %v", len(pipeline.OnFailure), 1)
	}

	if pipeline.Version != 1 {
		t.Errorf("GetIngestPipeline() Version = %v, want %v", pipeline.Version, 1)
	}

	if pipeline.Meta == nil {
		t.Fatal("GetIngestPipeline() Meta is nil")
	}

	createdAt, ok := pipeline.Meta["created_at"].(string)
	if !ok || createdAt != "2026-01-10T10:00:00Z" {
		t.Errorf("GetIngestPipeline() Meta.created_at = %v, want %v", createdAt, "2026-01-10T10:00:00Z")
	}

	updatedAt, ok := pipeline.Meta["updated_at"].(string)
	if !ok || updatedAt != "2026-01-10T12:00:00Z" {
		t.Errorf("GetIngestPipeline() Meta.updated_at = %v, want %v", updatedAt, "2026-01-10T12:00:00Z")
	}
}

func TestGetIngestPipeline_ConnectionError(t *testing.T) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:99999"},
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	pipeline, err := GetIngestPipeline(esClient, "test-pipeline")

	if err == nil {
		t.Error("GetIngestPipeline() with connection error should return an error")
	}

	if pipeline != nil {
		t.Error("GetIngestPipeline() with connection error should return nil pipeline")
	}
}
