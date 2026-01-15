package template

import (
	"context"
	"encoding/json"
	"testing"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// jsonVal is a helper function to create apiextensionsv1.JSON from any value
func jsonVal(v interface{}) apiextensionsv1.JSON {
	data, _ := json.Marshal(v)
	return apiextensionsv1.JSON{Raw: data}
}

func TestFetchAndRenderTemplate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = eseckv1alpha1.AddToScheme(scheme)

	// Create test ResourceTemplateData objects
	rtdProd := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prod-config",
			Namespace: "default",
			Labels:    map[string]string{"env": "prod"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"environment": jsonVal("production"),
				"replicas":    jsonVal(3),
				"host":        jsonVal("prod.example.com"),
			},
		},
	}

	rtdDev := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-config",
			Namespace: "default",
			Labels:    map[string]string{"env": "dev"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"environment": jsonVal("development"),
				"replicas":    jsonVal(1),
				"host":        jsonVal("dev.example.com"),
			},
		},
	}

	rtdOtherNs := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-config",
			Namespace: "other-ns",
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"value": jsonVal("from-other-namespace"),
			},
		},
	}

	// boolPtr is a helper function to create a pointer to a bool
	boolPtr := func(b bool) *bool {
		return &b
	}

	tests := []struct {
		name             string
		templateSpec     eseckv1alpha1.CommonTemplatingSpec
		body             string
		defaultNamespace string
		want             string
		wantErr          bool
	}{
		{
			name:             "no template references returns original body",
			templateSpec:     eseckv1alpha1.CommonTemplatingSpec{},
			body:             `{"static": "content"}`,
			defaultNamespace: "default",
			want:             `{"static": "content"}`,
			wantErr:          false,
		},
		{
			name: "empty references returns original body",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{},
			},
			body:             `{"static": "content"}`,
			defaultNamespace: "default",
			want:             `{"static": "content"}`,
			wantErr:          false,
		},
		{
			name: "enabled false returns original body even with references",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				Enabled: boolPtr(false),
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body:             `{"env": "{{ index .Values "default" "prod-config" "environment" }}"}`,
			defaultNamespace: "default",
			want:             `{"env": "{{ index .Values "default" "prod-config" "environment" }}"}`,
			wantErr:          false,
		},
		{
			name: "enabled true with references renders template",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				Enabled: boolPtr(true),
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body:             `{"env": "{{ index .Values "default" "prod-config" "environment" }}"}`,
			defaultNamespace: "default",
			want:             `{"env": "production"}`,
			wantErr:          false,
		},
		{
			name: "fetch and render with name reference",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body:             `{"env": "{{ index .Values "default" "prod-config" "environment" }}", "replicas": {{ index .Values "default" "prod-config" "replicas" }}}`,
			defaultNamespace: "default",
			want:             `{"env": "production", "replicas": 3}`,
			wantErr:          false,
		},
		{
			name: "fetch and render with label selector",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{LabelSelector: map[string]string{"env": "dev"}},
				},
			},
			body:             `{"host": "{{ index .Values "default" "dev-config" "host" }}"}`,
			defaultNamespace: "default",
			want:             `{"host": "dev.example.com"}`,
			wantErr:          false,
		},
		{
			name: "fetch from different namespace",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "other-config", Namespace: "other-ns"},
				},
			},
			body:             `{"value": "{{ index .Values "other-ns" "other-config" "value" }}"}`,
			defaultNamespace: "default",
			want:             `{"value": "from-other-namespace"}`,
			wantErr:          false,
		},
		{
			name: "fetch multiple resources and render",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
					{Name: "dev-config", Namespace: "default"},
				},
			},
			body:             `{"prod": "{{ index .Values "default" "prod-config" "host" }}", "dev": "{{ index .Values "default" "dev-config" "host" }}"}`,
			defaultNamespace: "default",
			want:             `{"prod": "prod.example.com", "dev": "dev.example.com"}`,
			wantErr:          false,
		},
		{
			name: "error when referenced resource not found with explicit namespace",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "nonexistent", Namespace: "default"},
				},
			},
			body:             `{"key": "value"}`,
			defaultNamespace: "default",
			want:             "",
			wantErr:          true,
		},
		{
			name: "no error when referenced resource not found without namespace",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "nonexistent"},
				},
			},
			body:             `{"static": "content"}`,
			defaultNamespace: "default",
			want:             `{"static": "content"}`,
			wantErr:          false,
		},
		{
			name: "template error with invalid syntax",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body:             `{"broken": "{{ .Values.unclosed }`,
			defaultNamespace: "default",
			want:             "",
			wantErr:          true,
		},
		{
			name: "complex template with conditionals",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body: `{
  "environment": "{{ index .Values "default" "prod-config" "environment" }}",
  {{- if gt (index .Values "default" "prod-config" "replicas" | int) 1 }}
  "ha_enabled": true
  {{- else }}
  "ha_enabled": false
  {{- end }}
}`,
			defaultNamespace: "default",
			want: `{
  "environment": "production",
  "ha_enabled": true
}`,
			wantErr: false,
		},
		{
			name: "use sprig functions in template",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "prod-config", Namespace: "default"},
				},
			},
			body:             `{"env_upper": "{{ index .Values "default" "prod-config" "environment" | upper }}"}`,
			defaultNamespace: "default",
			want:             `{"env_upper": "PRODUCTION"}`,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(rtdProd, rtdDev, rtdOtherNs).
				Build()

			got, err := FetchAndRenderTemplate(
				fakeClient,
				context.Background(),
				tt.templateSpec,
				tt.body,
				tt.defaultNamespace,
				nil, // rest.Config can be nil for basic templates
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAndRenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("FetchAndRenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchAndRenderTemplate_Integration(t *testing.T) {
	// This test simulates a more realistic scenario with nested values
	scheme := runtime.NewScheme()
	_ = eseckv1alpha1.AddToScheme(scheme)

	elasticsearchConfig := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "elasticsearch-settings",
			Namespace: "elastic-system",
			Labels:    map[string]string{"app": "elasticsearch"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"cluster": jsonVal(map[string]interface{}{
					"name":     "my-cluster",
					"replicas": 3,
				}),
				"indices": jsonVal(map[string]interface{}{
					"shards":   5,
					"replicas": 1,
				}),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(elasticsearchConfig).
		Build()

	templateSpec := eseckv1alpha1.CommonTemplatingSpec{
		References: []eseckv1alpha1.CommonTemplatingSpecReference{
			{Name: "elasticsearch-settings", Namespace: "elastic-system"},
		},
	}

	body := `{
  "settings": {
    "index": {
      "number_of_shards": {{ index .Values "elastic-system" "elasticsearch-settings" "indices" "shards" }},
      "number_of_replicas": {{ index .Values "elastic-system" "elasticsearch-settings" "indices" "replicas" }}
    }
  }
}`

	expected := `{
  "settings": {
    "index": {
      "number_of_shards": 5,
      "number_of_replicas": 1
    }
  }
}`

	got, err := FetchAndRenderTemplate(
		fakeClient,
		context.Background(),
		templateSpec,
		body,
		"default",
		nil,
	)

	if err != nil {
		t.Errorf("FetchAndRenderTemplate() unexpected error: %v", err)
		return
	}

	if got != expected {
		t.Errorf("FetchAndRenderTemplate() = %q, want %q", got, expected)
	}
}

func TestFetchAndRenderTemplate_IngestPipelineScenario(t *testing.T) {
	// Test a realistic ingest pipeline scenario
	scheme := runtime.NewScheme()
	_ = eseckv1alpha1.AddToScheme(scheme)

	pipelineConfig := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pipeline-settings",
			Namespace: "default",
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"timestamp_field": jsonVal("@timestamp"),
				"target_index":    jsonVal("logs-processed"),
				"description":     jsonVal("Process incoming log data"),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(pipelineConfig).
		Build()

	templateSpec := eseckv1alpha1.CommonTemplatingSpec{
		References: []eseckv1alpha1.CommonTemplatingSpecReference{
			{Name: "pipeline-settings", Namespace: "default"},
		},
	}

	body := `{
  "description": "{{ index .Values "default" "pipeline-settings" "description" }}",
  "processors": [
    {
      "set": {
        "field": "_index",
        "value": "{{ index .Values "default" "pipeline-settings" "target_index" }}"
      }
    },
    {
      "date": {
        "field": "{{ index .Values "default" "pipeline-settings" "timestamp_field" }}",
        "target_field": "@timestamp",
        "formats": ["ISO8601"]
      }
    }
  ]
}`

	expected := `{
  "description": "Process incoming log data",
  "processors": [
    {
      "set": {
        "field": "_index",
        "value": "logs-processed"
      }
    },
    {
      "date": {
        "field": "@timestamp",
        "target_field": "@timestamp",
        "formats": ["ISO8601"]
      }
    }
  ]
}`

	got, err := FetchAndRenderTemplate(
		fakeClient,
		context.Background(),
		templateSpec,
		body,
		"default",
		nil,
	)

	if err != nil {
		t.Errorf("FetchAndRenderTemplate() unexpected error: %v", err)
		return
	}

	if got != expected {
		t.Errorf("FetchAndRenderTemplate() = %q, want %q", got, expected)
	}
}
