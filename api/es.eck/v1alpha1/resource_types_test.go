/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestComponentTemplateSpec(t *testing.T) {
	spec := ComponentTemplateSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"template": {"mappings": {}}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestComponentTemplateSpec_WithDependencies(t *testing.T) {
	spec := ComponentTemplateSpec{
		Body: `{"template": {}}`,
		Dependencies: Dependencies{
			IndexTemplates:     []string{"template-1"},
			ComponentTemplates: []string{"component-1", "component-2"},
		},
	}

	if len(spec.Dependencies.IndexTemplates) != 1 {
		t.Errorf("Expected 1 index template dependency, got %d", len(spec.Dependencies.IndexTemplates))
	}

	if len(spec.Dependencies.ComponentTemplates) != 2 {
		t.Errorf("Expected 2 component template dependencies, got %d", len(spec.Dependencies.ComponentTemplates))
	}
}

func TestComponentTemplate(t *testing.T) {
	template := ComponentTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ComponentTemplate",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-component-template",
			Namespace: "default",
		},
		Spec: ComponentTemplateSpec{
			Body: `{"template": {"mappings": {}}}`,
		},
	}

	if template.Name != "test-component-template" {
		t.Errorf("Expected Name to be 'test-component-template', got %q", template.Name)
	}

	if template.Kind != "ComponentTemplate" {
		t.Errorf("Expected Kind to be 'ComponentTemplate', got %q", template.Kind)
	}
}

func TestComponentTemplateList(t *testing.T) {
	list := ComponentTemplateList{
		Items: []ComponentTemplate{
			{ObjectMeta: metav1.ObjectMeta{Name: "template-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "template-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}

func TestComponentTemplateStatus(t *testing.T) {
	status := ComponentTemplateStatus{}
	// Status is empty, just verify struct exists
	_ = status
}

// Tests for IndexTemplate
func TestIndexTemplateSpec(t *testing.T) {
	spec := IndexTemplateSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"index_patterns": ["logs-*"], "template": {}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestIndexTemplate(t *testing.T) {
	template := IndexTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IndexTemplate",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-index-template",
			Namespace: "default",
		},
		Spec: IndexTemplateSpec{
			Body: `{"index_patterns": ["*"], "template": {}}`,
		},
	}

	if template.Name != "test-index-template" {
		t.Errorf("Expected Name to be 'test-index-template', got %q", template.Name)
	}

	if template.Kind != "IndexTemplate" {
		t.Errorf("Expected Kind to be 'IndexTemplate', got %q", template.Kind)
	}
}

func TestIndexTemplateList(t *testing.T) {
	list := IndexTemplateList{
		Items: []IndexTemplate{
			{ObjectMeta: metav1.ObjectMeta{Name: "template-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for Index
func TestIndexSpec(t *testing.T) {
	spec := IndexSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"settings": {"number_of_shards": 1}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestIndex(t *testing.T) {
	index := Index{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Index",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-index",
			Namespace: "default",
		},
		Spec: IndexSpec{
			Body: `{"settings": {}}`,
		},
	}

	if index.Name != "test-index" {
		t.Errorf("Expected Name to be 'test-index', got %q", index.Name)
	}
}

func TestIndexList(t *testing.T) {
	list := IndexList{
		Items: []Index{
			{ObjectMeta: metav1.ObjectMeta{Name: "index-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for IndexLifecyclePolicy
func TestIndexLifecyclePolicySpec(t *testing.T) {
	spec := IndexLifecyclePolicySpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"policy": {"phases": {}}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}
}

func TestIndexLifecyclePolicy(t *testing.T) {
	ilp := IndexLifecyclePolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IndexLifecyclePolicy",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ilp",
			Namespace: "default",
		},
	}

	if ilp.Name != "test-ilp" {
		t.Errorf("Expected Name to be 'test-ilp', got %q", ilp.Name)
	}
}

func TestIndexLifecyclePolicyList(t *testing.T) {
	list := IndexLifecyclePolicyList{
		Items: []IndexLifecyclePolicy{
			{ObjectMeta: metav1.ObjectMeta{Name: "ilp-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}
