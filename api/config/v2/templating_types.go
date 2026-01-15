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

package v2

// CommonTemplatingSpec defines the templating configuration for resources
type CommonTemplatingSpec struct {
	// Enabled indicates if templating is active. Defaults to true.
	// +optional
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// References to ResourceTemplateData objects
	// +optional
	References []CommonTemplatingSpecReference `json:"references,omitempty"`
}

// CommonTemplatingSpecReference defines a reference to a ResourceTemplateData object
type CommonTemplatingSpecReference struct {
	// Name of the ResourceTemplateData object
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace of the ResourceTemplateData object
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// LabelSelector to select ResourceTemplateData objects
	// +optional
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
}

// IsEnabled returns true if templating is enabled (defaults to true if not set)
func (c *CommonTemplatingSpec) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}
