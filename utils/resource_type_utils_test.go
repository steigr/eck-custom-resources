package utils

import (
	"context"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Test types for GetRegisteredGVKsInGroupWithTemplatingSpec
type TestResourceWithTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpecWithTemplate `json:"spec,omitempty"`
}

type TestSpecWithTemplate struct {
	Template TestTemplateSpec `json:"template,omitempty"`
	Body     string           `json:"body,omitempty"`
}

type TestTemplateSpec struct {
	References []TestReference `json:"references,omitempty"`
}

type TestReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func (t *TestResourceWithTemplate) DeepCopyObject() runtime.Object {
	return &TestResourceWithTemplate{
		TypeMeta:   t.TypeMeta,
		ObjectMeta: *t.ObjectMeta.DeepCopy(),
		Spec:       t.Spec,
	}
}

type TestResourceWithoutTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpecWithoutTemplate `json:"spec,omitempty"`
}

type TestSpecWithoutTemplate struct {
	Body string `json:"body,omitempty"`
}

func (t *TestResourceWithoutTemplate) DeepCopyObject() runtime.Object {
	return &TestResourceWithoutTemplate{
		TypeMeta:   t.TypeMeta,
		ObjectMeta: *t.ObjectMeta.DeepCopy(),
		Spec:       t.Spec,
	}
}

type TestResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestResourceWithTemplate `json:"items"`
}

func (t *TestResourceList) DeepCopyObject() runtime.Object {
	return &TestResourceList{
		TypeMeta: t.TypeMeta,
		ListMeta: *t.ListMeta.DeepCopy(),
	}
}

func TestHasTemplateInSpec(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want bool
	}{
		{
			name: "struct with Template in Spec",
			typ:  reflect.TypeOf(TestResourceWithTemplate{}),
			want: true,
		},
		{
			name: "pointer to struct with Template in Spec",
			typ:  reflect.TypeOf(&TestResourceWithTemplate{}),
			want: true,
		},
		{
			name: "struct without Template in Spec",
			typ:  reflect.TypeOf(TestResourceWithoutTemplate{}),
			want: false,
		},
		{
			name: "pointer to struct without Template in Spec",
			typ:  reflect.TypeOf(&TestResourceWithoutTemplate{}),
			want: false,
		},
		{
			name: "non-struct type",
			typ:  reflect.TypeOf("string"),
			want: false,
		},
		{
			name: "struct without Spec field",
			typ:  reflect.TypeOf(struct{ Name string }{}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasTemplateInSpec(tt.typ)
			if got != tt.want {
				t.Errorf("hasTemplateInSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRegisteredGVKsInGroupWithTemplatingSpec(t *testing.T) {
	scheme := runtime.NewScheme()

	// Register test types
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "test.group", Version: "v1", Kind: "WithTemplate"},
		&TestResourceWithTemplate{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "test.group", Version: "v1", Kind: "WithTemplateList"},
		&TestResourceList{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "test.group", Version: "v1", Kind: "WithoutTemplate"},
		&TestResourceWithoutTemplate{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "other.group", Version: "v1", Kind: "OtherResource"},
		&TestResourceWithTemplate{},
	)

	tests := []struct {
		name      string
		group     string
		wantKinds []string
	}{
		{
			name:      "returns only types with Template in test.group",
			group:     "test.group",
			wantKinds: []string{"WithTemplate"},
		},
		{
			name:      "returns types from other.group",
			group:     "other.group",
			wantKinds: []string{"OtherResource"},
		},
		{
			name:      "returns empty for non-existent group",
			group:     "nonexistent.group",
			wantKinds: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRegisteredGVKsInGroupWithTemplatingSpec(scheme, tt.group)

			gotKinds := make([]string, len(got))
			for i, gvk := range got {
				gotKinds[i] = gvk.Kind
			}

			if len(gotKinds) != len(tt.wantKinds) {
				t.Errorf("GetRegisteredGVKsInGroupWithTemplatingSpec() got %d kinds, want %d. Got: %v, Want: %v",
					len(gotKinds), len(tt.wantKinds), gotKinds, tt.wantKinds)
				return
			}

			wantMap := make(map[string]bool)
			for _, kind := range tt.wantKinds {
				wantMap[kind] = true
			}
			for _, kind := range gotKinds {
				if !wantMap[kind] {
					t.Errorf("GetRegisteredGVKsInGroupWithTemplatingSpec() got unexpected kind %q", kind)
				}
			}
		})
	}
}

func TestReferencesResourceTemplateData(t *testing.T) {
	tests := []struct {
		name                          string
		resource                      unstructured.Unstructured
		resourceTemplateDataName      string
		resourceTemplateDataNamespace string
		want                          bool
	}{
		{
			name: "matches by name with explicit namespace",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{
								map[string]interface{}{
									"name":      "my-template-data",
									"namespace": "template-ns",
								},
							},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "template-ns",
			want:                          true,
		},
		{
			name: "matches by name with implicit namespace (uses resource namespace)",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{
								map[string]interface{}{
									"name": "my-template-data",
								},
							},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          true,
		},
		{
			name: "does not match - different name",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{
								map[string]interface{}{
									"name":      "other-template-data",
									"namespace": "default",
								},
							},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
		{
			name: "does not match - different namespace",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{
								map[string]interface{}{
									"name":      "my-template-data",
									"namespace": "other-ns",
								},
							},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
		{
			name: "matches one of multiple references",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{
								map[string]interface{}{
									"name":      "first-template",
									"namespace": "default",
								},
								map[string]interface{}{
									"name":      "my-template-data",
									"namespace": "default",
								},
								map[string]interface{}{
									"name":      "third-template",
									"namespace": "default",
								},
							},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          true,
		},
		{
			name: "no spec field",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
		{
			name: "no template field",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"body": "some content",
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
		{
			name: "no references field",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
		{
			name: "empty references",
			resource: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-resource",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"references": []interface{}{},
						},
					},
				},
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			want:                          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := referencesResourceTemplateData(tt.resource, tt.resourceTemplateDataName, tt.resourceTemplateDataNamespace)
			if got != tt.want {
				t.Errorf("referencesResourceTemplateData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListResourcesReferencingResourceTemplateData(t *testing.T) {
	scheme := runtime.NewScheme()

	// Create test resources
	resource1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.group/v1",
			"kind":       "TestResource",
			"metadata": map[string]interface{}{
				"name":      "resource-1",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"references": []interface{}{
						map[string]interface{}{
							"name":      "my-template-data",
							"namespace": "default",
						},
					},
				},
			},
		},
	}

	resource2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.group/v1",
			"kind":       "TestResource",
			"metadata": map[string]interface{}{
				"name":      "resource-2",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"references": []interface{}{
						map[string]interface{}{
							"name":      "other-template-data",
							"namespace": "default",
						},
					},
				},
			},
		},
	}

	resource3 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.group/v1",
			"kind":       "TestResource",
			"metadata": map[string]interface{}{
				"name":      "resource-3",
				"namespace": "other-ns",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"references": []interface{}{
						map[string]interface{}{
							"name": "my-template-data",
							// No namespace specified, should use resource namespace
						},
					},
				},
			},
		},
	}

	resource4 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "test.group/v1",
			"kind":       "TestResource",
			"metadata": map[string]interface{}{
				"name":      "resource-4",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"body": "no template references",
			},
		},
	}

	tests := []struct {
		name                          string
		gvk                           schema.GroupVersionKind
		resourceTemplateDataName      string
		resourceTemplateDataNamespace string
		wantNames                     []string
		wantErr                       bool
	}{
		{
			name: "find resources referencing template data in default namespace",
			gvk: schema.GroupVersionKind{
				Group:   "test.group",
				Version: "v1",
				Kind:    "TestResource",
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "default",
			wantNames:                     []string{"resource-1"},
			wantErr:                       false,
		},
		{
			name: "find resources referencing template data in other-ns namespace",
			gvk: schema.GroupVersionKind{
				Group:   "test.group",
				Version: "v1",
				Kind:    "TestResource",
			},
			resourceTemplateDataName:      "my-template-data",
			resourceTemplateDataNamespace: "other-ns",
			wantNames:                     []string{"resource-3"},
			wantErr:                       false,
		},
		{
			name: "no resources match",
			gvk: schema.GroupVersionKind{
				Group:   "test.group",
				Version: "v1",
				Kind:    "TestResource",
			},
			resourceTemplateDataName:      "nonexistent-template",
			resourceTemplateDataNamespace: "default",
			wantNames:                     []string{},
			wantErr:                       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test resources
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(resource1, resource2, resource3, resource4).
				Build()

			got, err := ListResourcesReferencingResourceTemplateData(
				fakeClient,
				context.Background(),
				tt.gvk,
				tt.resourceTemplateDataName,
				tt.resourceTemplateDataNamespace,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListResourcesReferencingResourceTemplateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			gotNames := make([]string, len(got))
			for i, r := range got {
				gotNames[i] = r.GetName()
			}

			if len(gotNames) != len(tt.wantNames) {
				t.Errorf("ListResourcesReferencingResourceTemplateData() got %d resources, want %d. Got: %v, Want: %v",
					len(gotNames), len(tt.wantNames), gotNames, tt.wantNames)
				return
			}

			wantMap := make(map[string]bool)
			for _, name := range tt.wantNames {
				wantMap[name] = true
			}
			for _, name := range gotNames {
				if !wantMap[name] {
					t.Errorf("ListResourcesReferencingResourceTemplateData() got unexpected resource %q", name)
				}
			}
		})
	}
}
