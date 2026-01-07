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

// jsonValue is a helper function to create apiextensionsv1.JSON from any value
func jsonValue(v interface{}) apiextensionsv1.JSON {
	data, _ := json.Marshal(v)
	return apiextensionsv1.JSON{Raw: data}
}

func TestHasTemplateReferences(t *testing.T) {
	tests := []struct {
		name         string
		templateSpec eseckv1alpha1.CommonTemplatingSpec
		want         bool
	}{
		{
			name:         "empty template spec",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{},
			want:         false,
		},
		{
			name: "nil references",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: nil,
			},
			want: false,
		},
		{
			name: "empty references slice",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{},
			},
			want: false,
		},
		{
			name: "with name reference",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "my-template-data"},
				},
			},
			want: true,
		},
		{
			name: "with label selector reference",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{LabelSelector: map[string]string{"app": "test"}},
				},
			},
			want: true,
		},
		{
			name: "with multiple references",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "template-data-1"},
					{Name: "template-data-2", Namespace: "other-ns"},
					{LabelSelector: map[string]string{"env": "prod"}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasTemplateReferences(tt.templateSpec)
			if got != tt.want {
				t.Errorf("HasTemplateReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderBody(t *testing.T) {
	tests := []struct {
		name                     string
		body                     string
		resourceTemplateDataList []eseckv1alpha1.ResourceTemplateData
		want                     string
		wantErr                  bool
	}{
		{
			name: "simple variable substitution",
			body: `{"name": "{{ .Values.default.mydata.name }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"name": jsonValue("test-value"),
						},
					},
				},
			},
			want:    `{"name": "test-value"}`,
			wantErr: false,
		},
		{
			name: "multiple variables from same resource",
			body: `{"name": "{{ .Values.default.config.name }}", "value": "{{ .Values.default.config.value }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "config", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"name":  jsonValue("my-name"),
							"value": jsonValue("my-value"),
						},
					},
				},
			},
			want:    `{"name": "my-name", "value": "my-value"}`,
			wantErr: false,
		},
		{
			name: "variables from multiple resources in same namespace",
			body: `{"db": "{{ .Values.default.database.host }}", "cache": "{{ .Values.default.cache.host }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"host": jsonValue("db.example.com"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "cache", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"host": jsonValue("cache.example.com"),
						},
					},
				},
			},
			want:    `{"db": "db.example.com", "cache": "cache.example.com"}`,
			wantErr: false,
		},
		{
			name: "variables from multiple namespaces",
			body: `{"db": "{{ .Values.production.database.host }}", "cache": "{{ .Values.staging.cache.host }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database", Namespace: "production"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"host": jsonValue("db.prod.example.com"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "cache", Namespace: "staging"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"host": jsonValue("cache.staging.example.com"),
						},
					},
				},
			},
			want:    `{"db": "db.prod.example.com", "cache": "cache.staging.example.com"}`,
			wantErr: false,
		},
		{
			name:                     "empty resource list with no variables",
			body:                     `{"static": "value"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{},
			want:                     `{"static": "value"}`,
			wantErr:                  false,
		},
		{
			name: "using sprig functions - default",
			body: `{"value": "{{ default "fallback" .Values.default.mydata.missing }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"other": jsonValue("data"),
						},
					},
				},
			},
			want:    `{"value": "fallback"}`,
			wantErr: false,
		},
		{
			name: "using sprig functions - upper",
			body: `{"value": "{{ .Values.default.mydata.name | upper }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"name": jsonValue("hello"),
						},
					},
				},
			},
			want:    `{"value": "HELLO"}`,
			wantErr: false,
		},
		{
			name: "using sprig functions - quote",
			body: `{"value": {{ .Values.default.mydata.name | quote }}}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"name": jsonValue("test"),
						},
					},
				},
			},
			want:    `{"value": "test"}`,
			wantErr: false,
		},
		{
			name: "conditional template",
			body: `{{- if .Values.default.config.enabled }}{"enabled": true}{{- else }}{"enabled": false}{{- end }}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "config", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"enabled": jsonValue(true),
						},
					},
				},
			},
			want:    `{"enabled": true}`,
			wantErr: false,
		},
		{
			name:                     "invalid template syntax",
			body:                     `{"value": "{{ .unclosed }`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{},
			want:                     "",
			wantErr:                  true,
		},
		{
			name: "nested values",
			body: `{"nested": "{{ .Values.default.mydata.nested.deep.value }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"nested": jsonValue(map[string]interface{}{
								"deep": map[string]interface{}{
									"value": "deeply-nested",
								},
							}),
						},
					},
				},
			},
			want:    `{"nested": "deeply-nested"}`,
			wantErr: false,
		},
		{
			name: "array values",
			body: `{{- range .Values.default.mydata.items }}{{ . }}{{- end }}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"items": jsonValue([]string{"a", "b", "c"}),
						},
					},
				},
			},
			want:    `abc`,
			wantErr: false,
		},
		{
			name: "numeric values",
			body: `{"port": {{ .Values.default.mydata.port }}, "replicas": {{ .Values.default.mydata.replicas }}}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata", Namespace: "default"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"port":     jsonValue(8080),
							"replicas": jsonValue(3),
						},
					},
				},
			},
			want:    `{"port": 8080, "replicas": 3}`,
			wantErr: false,
		},
		{
			name: "kebab-case namespace and name used as-is",
			body: `{"value": "{{ index .Values "my-namespace" "my-resource" "key" }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "my-resource", Namespace: "my-namespace"},
					Spec: eseckv1alpha1.ResourceTemplateDataSpec{
						Values: map[string]apiextensionsv1.JSON{
							"key": jsonValue("original-keys"),
						},
					},
				},
			},
			want:    `{"value": "original-keys"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass nil for rest.Config since we're not using lookup functions in these tests
			got, err := RenderBody(tt.body, tt.resourceTemplateDataList, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RenderBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderBodyWithValues(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		values  map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple map values",
			body:    `{"key": "{{ .Values.key }}"}`,
			values:  map[string]interface{}{"key": "value"},
			want:    `{"key": "value"}`,
			wantErr: false,
		},
		{
			name: "nested map values",
			body: `{"nested": "{{ .Values.outer.inner }}"}`,
			values: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "nested-value",
				},
			},
			want:    `{"nested": "nested-value"}`,
			wantErr: false,
		},
		{
			name:    "empty values",
			body:    `{"static": "content"}`,
			values:  map[string]interface{}{},
			want:    `{"static": "content"}`,
			wantErr: false,
		},
		{
			name:    "nil values",
			body:    `{"static": "content"}`,
			values:  nil,
			want:    `{"static": "content"}`,
			wantErr: false,
		},
		{
			name: "range over slice",
			body: `{{- range .Values.items }}{{ . }}{{- end }}`,
			values: map[string]interface{}{
				"items": []string{"a", "b", "c"},
			},
			want:    `abc`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderBodyWithValues(tt.body, tt.values, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderBodyWithValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RenderBodyWithValues() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchResourceTemplateData(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = eseckv1alpha1.AddToScheme(scheme)

	// Create test ResourceTemplateData objects
	rtd1 := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rtd-1",
			Namespace: "default",
			Labels:    map[string]string{"app": "test", "env": "prod"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"key1": jsonValue("value1"),
			},
		},
	}
	rtd2 := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rtd-2",
			Namespace: "default",
			Labels:    map[string]string{"app": "test", "env": "dev"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"key2": jsonValue("value2"),
			},
		},
	}
	rtd3 := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rtd-3",
			Namespace: "other-ns",
			Labels:    map[string]string{"app": "other"},
		},
		Spec: eseckv1alpha1.ResourceTemplateDataSpec{
			Values: map[string]apiextensionsv1.JSON{
				"key3": jsonValue("value3"),
			},
		},
	}

	tests := []struct {
		name             string
		templateSpec     eseckv1alpha1.CommonTemplatingSpec
		defaultNamespace string
		wantNames        []string
		wantErr          bool
	}{
		{
			name: "fetch by name without namespace searches cluster-wide",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "rtd-1"},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-1"},
			wantErr:          false,
		},
		{
			name: "fetch by name in specific namespace",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "rtd-3", Namespace: "other-ns"},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-3"},
			wantErr:          false,
		},
		{
			name: "fetch by label selector without namespace searches cluster-wide",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{LabelSelector: map[string]string{"app": "test"}},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-1", "rtd-2"},
			wantErr:          false,
		},
		{
			name: "fetch by label selector with namespace",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{LabelSelector: map[string]string{"app": "test"}, Namespace: "default"},
				},
			},
			defaultNamespace: "other-ns",
			wantNames:        []string{"rtd-1", "rtd-2"},
			wantErr:          false,
		},
		{
			name: "fetch by label selector - specific label cluster-wide",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{LabelSelector: map[string]string{"env": "prod"}},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-1"},
			wantErr:          false,
		},
		{
			name: "fetch multiple by name",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "rtd-1"},
					{Name: "rtd-2"},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-1", "rtd-2"},
			wantErr:          false,
		},
		{
			name: "fetch non-existent resource without namespace returns empty",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "non-existent"},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{},
			wantErr:          false,
		},
		{
			name: "fetch non-existent resource with explicit namespace returns error",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "non-existent", Namespace: "default"},
				},
			},
			defaultNamespace: "default",
			wantNames:        nil,
			wantErr:          true,
		},
		{
			name: "empty references",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{},
			},
			defaultNamespace: "default",
			wantNames:        nil,
			wantErr:          false,
		},
		{
			name: "deduplicate when same resource matched by name and selector",
			templateSpec: eseckv1alpha1.CommonTemplatingSpec{
				References: []eseckv1alpha1.CommonTemplatingSpecReference{
					{Name: "rtd-1"},
					{LabelSelector: map[string]string{"app": "test"}},
				},
			},
			defaultNamespace: "default",
			wantNames:        []string{"rtd-1", "rtd-2"},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test objects
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(rtd1, rtd2, rtd3).
				Build()

			got, err := FetchResourceTemplateData(fakeClient, context.Background(), tt.templateSpec, tt.defaultNamespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchResourceTemplateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that we got the expected resources
			gotNames := make([]string, len(got))
			for i, rtd := range got {
				gotNames[i] = rtd.Name
			}

			if len(gotNames) != len(tt.wantNames) {
				t.Errorf("FetchResourceTemplateData() got %d resources, want %d. Got: %v, Want: %v",
					len(gotNames), len(tt.wantNames), gotNames, tt.wantNames)
				return
			}

			// Check all expected names are present (order may vary for label selectors)
			wantMap := make(map[string]bool)
			for _, name := range tt.wantNames {
				wantMap[name] = true
			}
			for _, name := range gotNames {
				if !wantMap[name] {
					t.Errorf("FetchResourceTemplateData() got unexpected resource %q", name)
				}
			}
		})
	}
}
