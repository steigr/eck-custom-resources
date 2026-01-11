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

// Tests for Dashboard
func TestDashboardSpec(t *testing.T) {
	space := "my-space"
	spec := DashboardSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Space: &space,
			Body:  `{"title": "Test Dashboard"}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}

	if spec.Space == nil || *spec.Space != "my-space" {
		t.Error("Expected Space to be 'my-space'")
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestDashboard(t *testing.T) {
	dashboard := Dashboard{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dashboard",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dashboard",
			Namespace: "default",
		},
		Spec: DashboardSpec{
			SavedObject: SavedObject{
				Body: `{"title": "Test"}`,
			},
		},
	}

	if dashboard.Name != "test-dashboard" {
		t.Errorf("Expected Name to be 'test-dashboard', got %q", dashboard.Name)
	}

	if dashboard.Kind != "Dashboard" {
		t.Errorf("Expected Kind to be 'Dashboard', got %q", dashboard.Kind)
	}
}

func TestDashboardList(t *testing.T) {
	list := DashboardList{
		Items: []Dashboard{
			{ObjectMeta: metav1.ObjectMeta{Name: "dash-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "dash-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}

// Tests for DataView
func TestDataViewSpec(t *testing.T) {
	spec := DataViewSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Body: `{"title": "logs-*", "timeFieldName": "@timestamp"}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestDataView(t *testing.T) {
	dataView := DataView{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataView",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataview",
			Namespace: "default",
		},
	}

	if dataView.Name != "test-dataview" {
		t.Errorf("Expected Name to be 'test-dataview', got %q", dataView.Name)
	}

	if dataView.Kind != "DataView" {
		t.Errorf("Expected Kind to be 'DataView', got %q", dataView.Kind)
	}
}

func TestDataViewList(t *testing.T) {
	list := DataViewList{
		Items: []DataView{
			{ObjectMeta: metav1.ObjectMeta{Name: "dv-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for Space
func TestSpaceSpec(t *testing.T) {
	spec := SpaceSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		Body: `{"name": "My Space", "description": "Test space"}`,
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestSpace(t *testing.T) {
	space := Space{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Space",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-space",
			Namespace: "default",
		},
	}

	if space.Name != "test-space" {
		t.Errorf("Expected Name to be 'test-space', got %q", space.Name)
	}

	if space.Kind != "Space" {
		t.Errorf("Expected Kind to be 'Space', got %q", space.Kind)
	}
}

func TestSpaceList(t *testing.T) {
	list := SpaceList{
		Items: []Space{
			{ObjectMeta: metav1.ObjectMeta{Name: "space-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for Visualization
func TestVisualizationSpec(t *testing.T) {
	spec := VisualizationSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Body: `{"title": "My Visualization", "visState": "{}"}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestVisualization(t *testing.T) {
	viz := Visualization{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Visualization",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-viz",
			Namespace: "default",
		},
	}

	if viz.Name != "test-viz" {
		t.Errorf("Expected Name to be 'test-viz', got %q", viz.Name)
	}

	if viz.Kind != "Visualization" {
		t.Errorf("Expected Kind to be 'Visualization', got %q", viz.Kind)
	}
}

func TestVisualizationList(t *testing.T) {
	list := VisualizationList{
		Items: []Visualization{
			{ObjectMeta: metav1.ObjectMeta{Name: "viz-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for Lens
func TestLensSpec(t *testing.T) {
	spec := LensSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Body: `{"title": "My Lens", "visualizationType": "lnsXY"}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}
}

func TestLens(t *testing.T) {
	lens := Lens{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Lens",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lens",
			Namespace: "default",
		},
	}

	if lens.Name != "test-lens" {
		t.Errorf("Expected Name to be 'test-lens', got %q", lens.Name)
	}

	if lens.Kind != "Lens" {
		t.Errorf("Expected Kind to be 'Lens', got %q", lens.Kind)
	}
}

func TestLensList(t *testing.T) {
	list := LensList{
		Items: []Lens{
			{ObjectMeta: metav1.ObjectMeta{Name: "lens-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for SavedSearch
func TestSavedSearchSpec(t *testing.T) {
	spec := SavedSearchSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Body: `{"title": "My Search", "columns": ["_source"]}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}
}

func TestSavedSearch(t *testing.T) {
	search := SavedSearch{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SavedSearch",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-search",
			Namespace: "default",
		},
	}

	if search.Name != "test-search" {
		t.Errorf("Expected Name to be 'test-search', got %q", search.Name)
	}

	if search.Kind != "SavedSearch" {
		t.Errorf("Expected Kind to be 'SavedSearch', got %q", search.Kind)
	}
}

func TestSavedSearchList(t *testing.T) {
	list := SavedSearchList{
		Items: []SavedSearch{
			{ObjectMeta: metav1.ObjectMeta{Name: "search-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for IndexPattern
func TestIndexPatternSpec(t *testing.T) {
	spec := IndexPatternSpec{
		TargetConfig: CommonKibanaConfig{
			KibanaInstance: "my-kibana",
		},
		SavedObject: SavedObject{
			Body: `{"title": "logs-*", "timeFieldName": "@timestamp"}`,
		},
	}

	if spec.TargetConfig.KibanaInstance != "my-kibana" {
		t.Errorf("Expected KibanaInstance to be 'my-kibana', got %q", spec.TargetConfig.KibanaInstance)
	}
}

func TestIndexPattern(t *testing.T) {
	pattern := IndexPattern{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IndexPattern",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pattern",
			Namespace: "default",
		},
	}

	if pattern.Name != "test-pattern" {
		t.Errorf("Expected Name to be 'test-pattern', got %q", pattern.Name)
	}

	if pattern.Kind != "IndexPattern" {
		t.Errorf("Expected Kind to be 'IndexPattern', got %q", pattern.Kind)
	}
}

func TestIndexPatternList(t *testing.T) {
	list := IndexPatternList{
		Items: []IndexPattern{
			{ObjectMeta: metav1.ObjectMeta{Name: "pattern-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for KibanaInstance
func TestKibanaInstance(t *testing.T) {
	instance := KibanaInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KibanaInstance",
			APIVersion: "kibana.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-kibana",
			Namespace: "default",
		},
	}

	if instance.Name != "my-kibana" {
		t.Errorf("Expected Name to be 'my-kibana', got %q", instance.Name)
	}

	if instance.Kind != "KibanaInstance" {
		t.Errorf("Expected Kind to be 'KibanaInstance', got %q", instance.Kind)
	}
}

func TestKibanaInstanceList(t *testing.T) {
	list := KibanaInstanceList{
		Items: []KibanaInstance{
			{ObjectMeta: metav1.ObjectMeta{Name: "kb-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "kb-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}
