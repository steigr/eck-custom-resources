package template

import (
	"context"
	"testing"

	eseckv1alpha1 "eck-custom-resources/api/es.eck/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single word lowercase",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "single word uppercase",
			input: "HELLO",
			want:  "hello",
		},
		{
			name:  "kebab-case",
			input: "my-resource-name",
			want:  "myResourceName",
		},
		{
			name:  "snake_case",
			input: "my_resource_name",
			want:  "myResourceName",
		},
		{
			name:  "PascalCase",
			input: "MyResourceName",
			want:  "myresourcename",
		},
		{
			name:  "already camelCase",
			input: "myResourceName",
			want:  "myresourcename",
		},
		{
			name:  "space separated",
			input: "my resource name",
			want:  "myResourceName",
		},
		{
			name:  "mixed separators",
			input: "my-resource_name",
			want:  "myResourceName",
		},
		{
			name:  "single character",
			input: "a",
			want:  "a",
		},
		{
			name:  "uppercase single character",
			input: "A",
			want:  "a",
		},
		{
			name:  "numbers in name",
			input: "my-resource-123",
			want:  "myResource123",
		},
		{
			name:  "rtd-1 style name",
			input: "rtd-1",
			want:  "rtd1",
		},
		{
			name:  "database-config style name",
			input: "database-config",
			want:  "databaseConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCamelCase(tt.input)
			if got != tt.want {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
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
			body: `{"name": "{{ .Values.mydata.name }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata"},
					Data:       map[string]string{"name": "test-value"},
				},
			},
			want:    `{"name": "test-value"}`,
			wantErr: false,
		},
		{
			name: "multiple variables from same resource",
			body: `{"name": "{{ .Values.config.name }}", "value": "{{ .Values.config.value }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "config"},
					Data:       map[string]string{"name": "my-name", "value": "my-value"},
				},
			},
			want:    `{"name": "my-name", "value": "my-value"}`,
			wantErr: false,
		},
		{
			name: "variables from multiple resources",
			body: `{"db": "{{ .Values.database.host }}", "cache": "{{ .Values.cache.host }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database"},
					Data:       map[string]string{"host": "db.example.com"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "cache"},
					Data:       map[string]string{"host": "cache.example.com"},
				},
			},
			want:    `{"db": "db.example.com", "cache": "cache.example.com"}`,
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
			body: `{"value": "{{ default "fallback" .Values.mydata.missing }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata"},
					Data:       map[string]string{"other": "data"},
				},
			},
			want:    `{"value": "fallback"}`,
			wantErr: false,
		},
		{
			name: "using sprig functions - upper",
			body: `{"value": "{{ .Values.mydata.name | upper }}"}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata"},
					Data:       map[string]string{"name": "hello"},
				},
			},
			want:    `{"value": "HELLO"}`,
			wantErr: false,
		},
		{
			name: "using sprig functions - quote",
			body: `{"value": {{ .Values.mydata.name | quote }}}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mydata"},
					Data:       map[string]string{"name": "test"},
				},
			},
			want:    `{"value": "test"}`,
			wantErr: false,
		},
		{
			name: "conditional template",
			body: `{{- if .Values.config.enabled }}{"enabled": true}{{- else }}{"enabled": false}{{- end }}`,
			resourceTemplateDataList: []eseckv1alpha1.ResourceTemplateData{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "config"},
					Data:       map[string]string{"enabled": "true"},
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
		Data: map[string]string{"key1": "value1"},
	}
	rtd2 := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rtd-2",
			Namespace: "default",
			Labels:    map[string]string{"app": "test", "env": "dev"},
		},
		Data: map[string]string{"key2": "value2"},
	}
	rtd3 := &eseckv1alpha1.ResourceTemplateData{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rtd-3",
			Namespace: "other-ns",
			Labels:    map[string]string{"app": "other"},
		},
		Data: map[string]string{"key3": "value3"},
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
