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

func TestSavedObject(t *testing.T) {
	space := "my-space"
	savedObj := SavedObject{
		Space: &space,
		Body:  `{"title": "Test Object"}`,
		Dependencies: []Dependency{
			{
				ObjectType: "visualization",
				Name:       "my-viz",
			},
		},
	}

	if savedObj.Space == nil || *savedObj.Space != "my-space" {
		t.Error("Expected Space to be 'my-space'")
	}

	if savedObj.Body == "" {
		t.Error("Expected Body to not be empty")
	}

	if len(savedObj.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(savedObj.Dependencies))
	}
}

func TestSavedObject_Empty(t *testing.T) {
	savedObj := SavedObject{}

	if savedObj.Space != nil {
		t.Error("Expected Space to be nil")
	}

	if savedObj.Body != "" {
		t.Errorf("Expected Body to be empty, got %q", savedObj.Body)
	}

	if savedObj.Dependencies != nil && len(savedObj.Dependencies) > 0 {
		t.Error("Expected Dependencies to be empty")
	}
}

func TestSavedObject_GetSavedObject(t *testing.T) {
	space := "test-space"
	original := SavedObject{
		Space: &space,
		Body:  `{"title": "Test"}`,
		Dependencies: []Dependency{
			{ObjectType: "dashboard", Name: "dash-1"},
		},
	}

	result := original.GetSavedObject()

	if result.Space == nil || *result.Space != "test-space" {
		t.Error("GetSavedObject should return same Space")
	}

	if result.Body != original.Body {
		t.Error("GetSavedObject should return same Body")
	}

	if len(result.Dependencies) != len(original.Dependencies) {
		t.Error("GetSavedObject should return same Dependencies")
	}
}

func TestDependency(t *testing.T) {
	space := "dep-space"
	dep := Dependency{
		ObjectType: "visualization",
		Name:       "my-visualization",
		Space:      &space,
	}

	if dep.ObjectType != "visualization" {
		t.Errorf("Expected ObjectType to be 'visualization', got %q", dep.ObjectType)
	}

	if dep.Name != "my-visualization" {
		t.Errorf("Expected Name to be 'my-visualization', got %q", dep.Name)
	}

	if dep.Space == nil || *dep.Space != "dep-space" {
		t.Error("Expected Space to be 'dep-space'")
	}
}

func TestDependency_NoSpace(t *testing.T) {
	dep := Dependency{
		ObjectType: "index-pattern",
		Name:       "logs-*",
	}

	if dep.ObjectType != "index-pattern" {
		t.Errorf("Expected ObjectType to be 'index-pattern', got %q", dep.ObjectType)
	}

	if dep.Space != nil {
		t.Error("Expected Space to be nil")
	}
}

func TestSavedObjectType(t *testing.T) {
	// Test that SavedObjectType can be used as expected
	var objType SavedObjectType = "visualization"

	if objType != "visualization" {
		t.Errorf("Expected 'visualization', got %q", objType)
	}

	// Test valid types
	validTypes := []SavedObjectType{
		"visualization",
		"dashboard",
		"search",
		"index-pattern",
		"lens",
	}

	for _, vt := range validTypes {
		if vt == "" {
			t.Errorf("SavedObjectType should not be empty")
		}
	}
}

func TestSavedObject_MultipleDependencies(t *testing.T) {
	savedObj := SavedObject{
		Body: `{"title": "Dashboard with deps"}`,
		Dependencies: []Dependency{
			{ObjectType: "visualization", Name: "viz-1"},
			{ObjectType: "visualization", Name: "viz-2"},
			{ObjectType: "index-pattern", Name: "logs-*"},
			{ObjectType: "lens", Name: "my-lens"},
		},
	}

	if len(savedObj.Dependencies) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(savedObj.Dependencies))
	}

	// Verify each dependency
	expectedTypes := []SavedObjectType{"visualization", "visualization", "index-pattern", "lens"}
	for i, dep := range savedObj.Dependencies {
		if dep.ObjectType != expectedTypes[i] {
			t.Errorf("Dependency %d: expected type %q, got %q", i, expectedTypes[i], dep.ObjectType)
		}
	}
}
