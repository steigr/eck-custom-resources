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
)

func TestDependencies(t *testing.T) {
	deps := Dependencies{
		IndexTemplates:     []string{"template-1", "template-2"},
		ComponentTemplates: []string{"component-1"},
		Indices:            []string{"index-1", "index-2", "index-3"},
	}

	if len(deps.IndexTemplates) != 2 {
		t.Errorf("Expected 2 index templates, got %d", len(deps.IndexTemplates))
	}

	if len(deps.ComponentTemplates) != 1 {
		t.Errorf("Expected 1 component template, got %d", len(deps.ComponentTemplates))
	}

	if len(deps.Indices) != 3 {
		t.Errorf("Expected 3 indices, got %d", len(deps.Indices))
	}
}

func TestDependencies_Empty(t *testing.T) {
	deps := Dependencies{}

	if deps.IndexTemplates != nil && len(deps.IndexTemplates) > 0 {
		t.Error("Expected IndexTemplates to be empty")
	}

	if deps.ComponentTemplates != nil && len(deps.ComponentTemplates) > 0 {
		t.Error("Expected ComponentTemplates to be empty")
	}

	if deps.Indices != nil && len(deps.Indices) > 0 {
		t.Error("Expected Indices to be empty")
	}
}

func TestDependencies_IndexTemplatesOnly(t *testing.T) {
	deps := Dependencies{
		IndexTemplates: []string{"logs-template", "metrics-template"},
	}

	if len(deps.IndexTemplates) != 2 {
		t.Errorf("Expected 2 index templates, got %d", len(deps.IndexTemplates))
	}

	if deps.IndexTemplates[0] != "logs-template" {
		t.Errorf("Expected first template to be 'logs-template', got %q", deps.IndexTemplates[0])
	}
}

func TestDependencies_ComponentTemplatesOnly(t *testing.T) {
	deps := Dependencies{
		ComponentTemplates: []string{"mappings", "settings"},
	}

	if len(deps.ComponentTemplates) != 2 {
		t.Errorf("Expected 2 component templates, got %d", len(deps.ComponentTemplates))
	}
}

func TestDependencies_IndicesOnly(t *testing.T) {
	deps := Dependencies{
		Indices: []string{"logs-000001"},
	}

	if len(deps.Indices) != 1 {
		t.Errorf("Expected 1 index, got %d", len(deps.Indices))
	}

	if deps.Indices[0] != "logs-000001" {
		t.Errorf("Expected index to be 'logs-000001', got %q", deps.Indices[0])
	}
}
