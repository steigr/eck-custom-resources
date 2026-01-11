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

func TestCommonElasticsearchConfig(t *testing.T) {
	config := CommonElasticsearchConfig{
		ElasticsearchInstance:          "my-elasticsearch",
		ElasticsearchInstanceNamespace: "elastic-system",
	}

	if config.ElasticsearchInstance != "my-elasticsearch" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-elasticsearch', got %q", config.ElasticsearchInstance)
	}

	if config.ElasticsearchInstanceNamespace != "elastic-system" {
		t.Errorf("Expected ElasticsearchInstanceNamespace to be 'elastic-system', got %q", config.ElasticsearchInstanceNamespace)
	}
}

func TestCommonElasticsearchConfig_Empty(t *testing.T) {
	config := CommonElasticsearchConfig{}

	if config.ElasticsearchInstance != "" {
		t.Errorf("Expected ElasticsearchInstance to be empty, got %q", config.ElasticsearchInstance)
	}

	if config.ElasticsearchInstanceNamespace != "" {
		t.Errorf("Expected ElasticsearchInstanceNamespace to be empty, got %q", config.ElasticsearchInstanceNamespace)
	}
}

func TestUpdateMode_Constants(t *testing.T) {
	if UpdateModeOverwrite != "Overwrite" {
		t.Errorf("Expected UpdateModeOverwrite to be 'Overwrite', got %q", UpdateModeOverwrite)
	}

	if UpdateModeBlock != "Block" {
		t.Errorf("Expected UpdateModeBlock to be 'Block', got %q", UpdateModeBlock)
	}
}

func TestUpdatePolicySpec(t *testing.T) {
	policy := UpdatePolicySpec{
		UpdateMode: UpdateModeOverwrite,
	}

	if policy.UpdateMode != UpdateModeOverwrite {
		t.Errorf("Expected UpdateMode to be UpdateModeOverwrite, got %q", policy.UpdateMode)
	}
}

func TestUpdatePolicySpec_Block(t *testing.T) {
	policy := UpdatePolicySpec{
		UpdateMode: UpdateModeBlock,
	}

	if policy.UpdateMode != UpdateModeBlock {
		t.Errorf("Expected UpdateMode to be UpdateModeBlock, got %q", policy.UpdateMode)
	}
}

func TestUpdatePolicySpec_Empty(t *testing.T) {
	policy := UpdatePolicySpec{}

	if policy.UpdateMode != "" {
		t.Errorf("Expected UpdateMode to be empty by default, got %q", policy.UpdateMode)
	}
}
