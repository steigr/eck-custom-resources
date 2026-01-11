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

func TestCommonTemplatingSpec(t *testing.T) {
	spec := CommonTemplatingSpec{
		References: []CommonTemplatingSpecReference{
			{
				Name:      "template-1",
				Namespace: "default",
			},
			{
				Name:      "template-2",
				Namespace: "other-ns",
			},
		},
	}

	if len(spec.References) != 2 {
		t.Errorf("Expected 2 references, got %d", len(spec.References))
	}

	if spec.References[0].Name != "template-1" {
		t.Errorf("Expected first reference name to be 'template-1', got %q", spec.References[0].Name)
	}

	if spec.References[1].Namespace != "other-ns" {
		t.Errorf("Expected second reference namespace to be 'other-ns', got %q", spec.References[1].Namespace)
	}
}

func TestCommonTemplatingSpec_Empty(t *testing.T) {
	spec := CommonTemplatingSpec{}

	if spec.References != nil && len(spec.References) != 0 {
		t.Errorf("Expected References to be empty, got %d", len(spec.References))
	}
}

func TestCommonTemplatingSpecReference(t *testing.T) {
	ref := CommonTemplatingSpecReference{
		Name:      "my-template",
		Namespace: "my-namespace",
		LabelSelector: map[string]string{
			"app":  "my-app",
			"tier": "backend",
		},
	}

	if ref.Name != "my-template" {
		t.Errorf("Expected Name to be 'my-template', got %q", ref.Name)
	}

	if ref.Namespace != "my-namespace" {
		t.Errorf("Expected Namespace to be 'my-namespace', got %q", ref.Namespace)
	}

	if len(ref.LabelSelector) != 2 {
		t.Errorf("Expected 2 label selectors, got %d", len(ref.LabelSelector))
	}

	if ref.LabelSelector["app"] != "my-app" {
		t.Errorf("Expected LabelSelector['app'] to be 'my-app', got %q", ref.LabelSelector["app"])
	}
}

func TestCommonTemplatingSpecReference_NameOnly(t *testing.T) {
	ref := CommonTemplatingSpecReference{
		Name: "template-by-name",
	}

	if ref.Name != "template-by-name" {
		t.Errorf("Expected Name to be 'template-by-name', got %q", ref.Name)
	}

	if ref.Namespace != "" {
		t.Errorf("Expected Namespace to be empty, got %q", ref.Namespace)
	}

	if ref.LabelSelector != nil && len(ref.LabelSelector) != 0 {
		t.Errorf("Expected LabelSelector to be empty")
	}
}

func TestCommonTemplatingSpecReference_LabelSelectorOnly(t *testing.T) {
	ref := CommonTemplatingSpecReference{
		LabelSelector: map[string]string{
			"environment": "production",
		},
	}

	if ref.Name != "" {
		t.Errorf("Expected Name to be empty, got %q", ref.Name)
	}

	if len(ref.LabelSelector) != 1 {
		t.Errorf("Expected 1 label selector, got %d", len(ref.LabelSelector))
	}

	if ref.LabelSelector["environment"] != "production" {
		t.Errorf("Expected LabelSelector['environment'] to be 'production', got %q", ref.LabelSelector["environment"])
	}
}
