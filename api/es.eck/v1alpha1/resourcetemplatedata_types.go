/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceTemplateDataSpec defines the desired state of ResourceTemplateData
type ResourceTemplateDataSpec struct {
	// +optional
	TargetConfig CommonElasticsearchConfig `json:"targetInstance,omitempty"`
}

// ResourceTemplateDataStatus defines the observed state of ResourceTemplateData.
type ResourceTemplateDataStatus struct {
	// +kubebuilder:validation:Format=int64
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ResourceTemplateData is the Schema for the resourcetemplatedata API
type ResourceTemplateData struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ResourceTemplateData
	// +required
	Spec ResourceTemplateDataSpec `json:"spec"`

	// data provides key-value pairs to be used
	// +required
	Data map[string]string `json:"data"`

	// status defines the observed state of ResourceTemplateData
	// +optional
	Status ResourceTemplateDataStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ResourceTemplateDataList contains a list of ResourceTemplateData
type ResourceTemplateDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ResourceTemplateData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceTemplateData{}, &ResourceTemplateDataList{})
}
