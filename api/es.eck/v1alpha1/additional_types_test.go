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

// Tests for ElasticsearchRole
func TestElasticsearchRoleSpec(t *testing.T) {
	spec := ElasticsearchRoleSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"cluster": ["monitor"], "indices": []}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestElasticsearchRole(t *testing.T) {
	role := ElasticsearchRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ElasticsearchRole",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-role",
			Namespace: "default",
		},
		Spec: ElasticsearchRoleSpec{
			Body: `{"cluster": [], "indices": []}`,
		},
	}

	if role.Name != "test-role" {
		t.Errorf("Expected Name to be 'test-role', got %q", role.Name)
	}

	if role.Kind != "ElasticsearchRole" {
		t.Errorf("Expected Kind to be 'ElasticsearchRole', got %q", role.Kind)
	}
}

func TestElasticsearchRoleList(t *testing.T) {
	list := ElasticsearchRoleList{
		Items: []ElasticsearchRole{
			{ObjectMeta: metav1.ObjectMeta{Name: "role-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "role-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}

// Tests for ElasticsearchUser
func TestElasticsearchUserSpec(t *testing.T) {
	spec := ElasticsearchUserSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"roles": ["viewer"]}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestElasticsearchUser(t *testing.T) {
	user := ElasticsearchUser{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ElasticsearchUser",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-user",
			Namespace: "default",
		},
		Spec: ElasticsearchUserSpec{
			Body: `{"roles": []}`,
		},
	}

	if user.Name != "test-user" {
		t.Errorf("Expected Name to be 'test-user', got %q", user.Name)
	}

	if user.Kind != "ElasticsearchUser" {
		t.Errorf("Expected Kind to be 'ElasticsearchUser', got %q", user.Kind)
	}
}

func TestElasticsearchUserList(t *testing.T) {
	list := ElasticsearchUserList{
		Items: []ElasticsearchUser{
			{ObjectMeta: metav1.ObjectMeta{Name: "user-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for SnapshotRepository
func TestSnapshotRepositorySpec(t *testing.T) {
	spec := SnapshotRepositorySpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"type": "fs", "settings": {"location": "/backup"}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestSnapshotRepository(t *testing.T) {
	repo := SnapshotRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SnapshotRepository",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
	}

	if repo.Name != "test-repo" {
		t.Errorf("Expected Name to be 'test-repo', got %q", repo.Name)
	}
}

func TestSnapshotRepositoryList(t *testing.T) {
	list := SnapshotRepositoryList{
		Items: []SnapshotRepository{
			{ObjectMeta: metav1.ObjectMeta{Name: "repo-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for SnapshotLifecyclePolicy
func TestSnapshotLifecyclePolicySpec(t *testing.T) {
	spec := SnapshotLifecyclePolicySpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"schedule": "0 30 1 * * ?", "name": "<snap-{now/d}>", "repository": "my-repo"}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}
}

func TestSnapshotLifecyclePolicy(t *testing.T) {
	slp := SnapshotLifecyclePolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SnapshotLifecyclePolicy",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-slp",
			Namespace: "default",
		},
	}

	if slp.Name != "test-slp" {
		t.Errorf("Expected Name to be 'test-slp', got %q", slp.Name)
	}
}

func TestSnapshotLifecyclePolicyList(t *testing.T) {
	list := SnapshotLifecyclePolicyList{
		Items: []SnapshotLifecyclePolicy{
			{ObjectMeta: metav1.ObjectMeta{Name: "slp-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for ElasticsearchInstance
func TestElasticsearchInstance(t *testing.T) {
	instance := ElasticsearchInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ElasticsearchInstance",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-es",
			Namespace: "default",
		},
	}

	if instance.Name != "my-es" {
		t.Errorf("Expected Name to be 'my-es', got %q", instance.Name)
	}

	if instance.Kind != "ElasticsearchInstance" {
		t.Errorf("Expected Kind to be 'ElasticsearchInstance', got %q", instance.Kind)
	}
}

func TestElasticsearchInstanceList(t *testing.T) {
	list := ElasticsearchInstanceList{
		Items: []ElasticsearchInstance{
			{ObjectMeta: metav1.ObjectMeta{Name: "es-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "es-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}

// Tests for ElasticsearchApikey
func TestElasticsearchApikeySpec(t *testing.T) {
	spec := ElasticsearchApikeySpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"name": "my-api-key", "role_descriptors": {}}`,
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}
}

func TestElasticsearchApikey(t *testing.T) {
	apiKey := ElasticsearchApikey{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ElasticsearchApikey",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-api-key",
			Namespace: "default",
		},
	}

	if apiKey.Name != "my-api-key" {
		t.Errorf("Expected Name to be 'my-api-key', got %q", apiKey.Name)
	}
}

func TestElasticsearchApikeyList(t *testing.T) {
	list := ElasticsearchApikeyList{
		Items: []ElasticsearchApikey{
			{ObjectMeta: metav1.ObjectMeta{Name: "key-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

// Tests for ResourceTemplateData
func TestResourceTemplateDataSpec(t *testing.T) {
	spec := ResourceTemplateDataSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}
}

func TestResourceTemplateData(t *testing.T) {
	rtd := ResourceTemplateData{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceTemplateData",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-template-data",
			Namespace: "default",
		},
	}

	if rtd.Name != "my-template-data" {
		t.Errorf("Expected Name to be 'my-template-data', got %q", rtd.Name)
	}
}

func TestResourceTemplateDataList(t *testing.T) {
	list := ResourceTemplateDataList{
		Items: []ResourceTemplateData{
			{ObjectMeta: metav1.ObjectMeta{Name: "rtd-1"}},
		},
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}
