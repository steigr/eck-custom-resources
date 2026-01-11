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

func TestIngestPipelineSpec(t *testing.T) {
	spec := IngestPipelineSpec{
		TargetConfig: CommonElasticsearchConfig{
			ElasticsearchInstance: "my-es",
		},
		Body: `{"description": "Test pipeline", "processors": []}`,
		UpdatePolicy: UpdatePolicySpec{
			UpdateMode: UpdateModeOverwrite,
		},
	}

	if spec.TargetConfig.ElasticsearchInstance != "my-es" {
		t.Errorf("Expected ElasticsearchInstance to be 'my-es', got %q", spec.TargetConfig.ElasticsearchInstance)
	}

	if spec.Body == "" {
		t.Error("Expected Body to not be empty")
	}

	if spec.UpdatePolicy.UpdateMode != UpdateModeOverwrite {
		t.Errorf("Expected UpdateMode to be Overwrite, got %q", spec.UpdatePolicy.UpdateMode)
	}
}

func TestIngestPipelineSpec_WithTemplate(t *testing.T) {
	spec := IngestPipelineSpec{
		Body: `{"description": "{{ .value }}", "processors": []}`,
		Template: CommonTemplatingSpec{
			References: []CommonTemplatingSpecReference{
				{
					Name:      "my-template-data",
					Namespace: "default",
				},
			},
		},
	}

	if len(spec.Template.References) != 1 {
		t.Errorf("Expected 1 template reference, got %d", len(spec.Template.References))
	}

	if spec.Template.References[0].Name != "my-template-data" {
		t.Errorf("Expected reference name to be 'my-template-data', got %q", spec.Template.References[0].Name)
	}
}

func TestIngestPipelineStatus(t *testing.T) {
	status := IngestPipelineStatus{
		ObservedGeneration: 5,
		Conditions: []metav1.Condition{
			{
				Type:   IngestPipelineConditionTypeInitialDeployment,
				Status: metav1.ConditionTrue,
				Reason: IngestPipelineReasonSucceeded,
			},
		},
	}

	if status.ObservedGeneration != 5 {
		t.Errorf("Expected ObservedGeneration to be 5, got %d", status.ObservedGeneration)
	}

	if len(status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(status.Conditions))
	}

	if status.Conditions[0].Type != IngestPipelineConditionTypeInitialDeployment {
		t.Errorf("Expected condition type to be InitialDeployment, got %q", status.Conditions[0].Type)
	}
}

func TestIngestPipelineConditionTypes(t *testing.T) {
	if IngestPipelineConditionTypeInitialDeployment != "InitialDeployment" {
		t.Errorf("Expected InitialDeployment, got %q", IngestPipelineConditionTypeInitialDeployment)
	}

	if IngestPipelineConditionTypeLastUpdate != "LastUpdate" {
		t.Errorf("Expected LastUpdate, got %q", IngestPipelineConditionTypeLastUpdate)
	}
}

func TestIngestPipelineReasons(t *testing.T) {
	if IngestPipelineReasonPending != "Pending" {
		t.Errorf("Expected Pending, got %q", IngestPipelineReasonPending)
	}

	if IngestPipelineReasonSucceeded != "Succeeded" {
		t.Errorf("Expected Succeeded, got %q", IngestPipelineReasonSucceeded)
	}

	if IngestPipelineReasonFailed != "Failed" {
		t.Errorf("Expected Failed, got %q", IngestPipelineReasonFailed)
	}

	if IngestPipelineReasonBlocked != "Blocked" {
		t.Errorf("Expected Blocked, got %q", IngestPipelineReasonBlocked)
	}
}

func TestIngestPipeline(t *testing.T) {
	pipeline := IngestPipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IngestPipeline",
			APIVersion: "es.eck.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pipeline",
			Namespace: "default",
		},
		Spec: IngestPipelineSpec{
			Body: `{"description": "Test", "processors": []}`,
		},
	}

	if pipeline.Name != "test-pipeline" {
		t.Errorf("Expected Name to be 'test-pipeline', got %q", pipeline.Name)
	}

	if pipeline.Kind != "IngestPipeline" {
		t.Errorf("Expected Kind to be 'IngestPipeline', got %q", pipeline.Kind)
	}
}

func TestIngestPipelineList(t *testing.T) {
	list := IngestPipelineList{
		Items: []IngestPipeline{
			{ObjectMeta: metav1.ObjectMeta{Name: "pipeline-1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "pipeline-2"}},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
}
