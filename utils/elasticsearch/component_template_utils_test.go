package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eck-custom-resources/api/es.eck/v1alpha1"

	"github.com/elastic/go-elasticsearch/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteComponentTemplate(t *testing.T) {
	tests := []struct {
		name             string
		templateName     string
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name:             "successful deletion",
			templateName:     "test-template",
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name:             "template not found",
			templateName:     "nonexistent-template",
			serverStatusCode: http.StatusNotFound,
			serverResponse:   `{"error": {"type": "resource_not_found_exception"}}`,
			wantRequeue:      true,
			wantErr:          false,
		},
		{
			name:             "server error",
			templateName:     "test-template",
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
				expectedPath := "/_component_template/" + tt.templateName
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

			result, err := DeleteComponentTemplate(esClient, tt.templateName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComponentTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("DeleteComponentTemplate() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestUpsertComponentTemplate(t *testing.T) {
	tests := []struct {
		name             string
		template         v1alpha1.ComponentTemplate
		serverStatusCode int
		serverResponse   string
		wantRequeue      bool
		wantErr          bool
	}{
		{
			name: "successful creation",
			template: v1alpha1.ComponentTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-template",
				},
				Spec: v1alpha1.ComponentTemplateSpec{
					Body: `{"template": {"mappings": {"properties": {"field1": {"type": "text"}}}}}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "successful update",
			template: v1alpha1.ComponentTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-template",
				},
				Spec: v1alpha1.ComponentTemplateSpec{
					Body: `{"template": {"mappings": {"properties": {"field1": {"type": "keyword"}}}}}`,
				},
			},
			serverStatusCode: http.StatusOK,
			serverResponse:   `{"acknowledged": true}`,
			wantRequeue:      false,
			wantErr:          false,
		},
		{
			name: "invalid template body",
			template: v1alpha1.ComponentTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-template",
				},
				Spec: v1alpha1.ComponentTemplateSpec{
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
			template: v1alpha1.ComponentTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-template",
				},
				Spec: v1alpha1.ComponentTemplateSpec{
					Body: `{"template": {}}`,
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
				expectedPath := "/_component_template/" + tt.template.Name
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

			result, err := UpsertComponentTemplate(esClient, tt.template)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertComponentTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Requeue != tt.wantRequeue {
				t.Errorf("UpsertComponentTemplate() Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
			}
		})
	}
}

func TestComponentTemplateExists(t *testing.T) {
	tests := []struct {
		name             string
		templateName     string
		serverStatusCode int
		wantExists       bool
		wantErr          bool
	}{
		{
			name:             "template exists",
			templateName:     "existing-template",
			serverStatusCode: http.StatusOK,
			wantExists:       true,
			wantErr:          false,
		},
		{
			name:             "template does not exist",
			templateName:     "nonexistent-template",
			serverStatusCode: http.StatusNotFound,
			wantExists:       false,
			wantErr:          false,
		},
		{
			name:             "server error",
			templateName:     "test-template",
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
				expectedPath := "/_component_template/" + tt.templateName
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

			exists, err := ComponentTemplateExists(esClient, tt.templateName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComponentTemplateExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("ComponentTemplateExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}
