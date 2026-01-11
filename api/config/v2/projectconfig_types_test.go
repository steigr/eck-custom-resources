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

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestProjectConfigSpec(t *testing.T) {
	spec := ProjectConfigSpec{
		Elasticsearch: ElasticsearchSpec{
			Enabled: true,
			Url:     "https://elasticsearch.example.com:9200",
		},
		Kibana: KibanaSpec{
			Enabled: true,
			Url:     "https://kibana.example.com:5601",
		},
	}

	if !spec.Elasticsearch.Enabled {
		t.Error("Expected Elasticsearch.Enabled to be true")
	}

	if spec.Elasticsearch.Url != "https://elasticsearch.example.com:9200" {
		t.Errorf("Expected Elasticsearch.Url to be 'https://elasticsearch.example.com:9200', got %q", spec.Elasticsearch.Url)
	}

	if !spec.Kibana.Enabled {
		t.Error("Expected Kibana.Enabled to be true")
	}

	if spec.Kibana.Url != "https://kibana.example.com:5601" {
		t.Errorf("Expected Kibana.Url to be 'https://kibana.example.com:5601', got %q", spec.Kibana.Url)
	}
}

func TestProjectConfigSpec_Empty(t *testing.T) {
	spec := ProjectConfigSpec{}

	if spec.Elasticsearch.Enabled {
		t.Error("Expected Elasticsearch.Enabled to be false by default")
	}

	if spec.Elasticsearch.Url != "" {
		t.Errorf("Expected Elasticsearch.Url to be empty, got %q", spec.Elasticsearch.Url)
	}

	if spec.Kibana.Enabled {
		t.Error("Expected Kibana.Enabled to be false by default")
	}

	if spec.Kibana.Url != "" {
		t.Errorf("Expected Kibana.Url to be empty, got %q", spec.Kibana.Url)
	}
}

func TestProjectConfig(t *testing.T) {
	config := ProjectConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "config.github.com/v2",
			Kind:       "ProjectConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-project-config",
			Namespace: "default",
		},
		Spec: ProjectConfigSpec{
			Elasticsearch: ElasticsearchSpec{
				Enabled: true,
				Url:     "https://elasticsearch.example.com:9200",
			},
			Kibana: KibanaSpec{
				Enabled: true,
				Url:     "https://kibana.example.com:5601",
			},
		},
	}

	if config.Name != "my-project-config" {
		t.Errorf("Expected Name to be 'my-project-config', got %q", config.Name)
	}

	if config.Namespace != "default" {
		t.Errorf("Expected Namespace to be 'default', got %q", config.Namespace)
	}

	if config.Kind != "ProjectConfig" {
		t.Errorf("Expected Kind to be 'ProjectConfig', got %q", config.Kind)
	}

	if config.APIVersion != "config.github.com/v2" {
		t.Errorf("Expected APIVersion to be 'config.github.com/v2', got %q", config.APIVersion)
	}
}

func TestProjectConfigList(t *testing.T) {
	list := ProjectConfigList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "config.github.com/v2",
			Kind:       "ProjectConfigList",
		},
		Items: []ProjectConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "config-1",
				},
				Spec: ProjectConfigSpec{
					Elasticsearch: ElasticsearchSpec{
						Url: "https://es1.example.com:9200",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "config-2",
				},
				Spec: ProjectConfigSpec{
					Elasticsearch: ElasticsearchSpec{
						Url: "https://es2.example.com:9200",
					},
				},
			},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items in list, got %d", len(list.Items))
	}

	if list.Items[0].Name != "config-1" {
		t.Errorf("Expected first item name to be 'config-1', got %q", list.Items[0].Name)
	}

	if list.Items[1].Name != "config-2" {
		t.Errorf("Expected second item name to be 'config-2', got %q", list.Items[1].Name)
	}
}

func TestProjectConfigStatus(t *testing.T) {
	status := ProjectConfigStatus{}
	// Status is currently empty, just verify the struct exists
	_ = status
}

func TestProjectConfigSpec_FullConfig(t *testing.T) {
	spec := ProjectConfigSpec{
		Elasticsearch: ElasticsearchSpec{
			Enabled: true,
			Url:     "https://elasticsearch.example.com:9200",
			Certificate: &PublicCertificate{
				SecretName:     "es-cert",
				CertificateKey: "ca.crt",
			},
			Authentication: &ElasticsearchAuthentication{
				UsernamePassword: &UsernamePasswordAuthentication{
					SecretName: "es-credentials",
					UserName:   "elastic",
				},
			},
		},
		Kibana: KibanaSpec{
			Enabled: true,
			Url:     "https://kibana.example.com:5601",
			Certificate: &PublicCertificate{
				SecretName:     "kb-cert",
				CertificateKey: "ca.crt",
			},
			Authentication: &KibanaAuthentication{
				UsernamePassword: &UsernamePasswordAuthentication{
					SecretName: "kb-credentials",
					UserName:   "kibana_system",
				},
			},
		},
	}

	// Verify Elasticsearch config
	if !spec.Elasticsearch.Enabled {
		t.Error("Expected Elasticsearch.Enabled to be true")
	}
	if spec.Elasticsearch.Certificate == nil {
		t.Error("Expected Elasticsearch.Certificate to not be nil")
	}
	if spec.Elasticsearch.Authentication == nil {
		t.Error("Expected Elasticsearch.Authentication to not be nil")
	}

	// Verify Kibana config
	if !spec.Kibana.Enabled {
		t.Error("Expected Kibana.Enabled to be true")
	}
	if spec.Kibana.Certificate == nil {
		t.Error("Expected Kibana.Certificate to not be nil")
	}
	if spec.Kibana.Authentication == nil {
		t.Error("Expected Kibana.Authentication to not be nil")
	}
}

func TestGroupVersion(t *testing.T) {
	expected := schema.GroupVersion{Group: "config.github.com", Version: "v2"}

	if GroupVersion != expected {
		t.Errorf("Expected GroupVersion to be %v, got %v", expected, GroupVersion)
	}
}

func TestSchemeBuilder(t *testing.T) {
	if SchemeBuilder == nil {
		t.Error("Expected SchemeBuilder to not be nil")
	}
}

func TestAddToScheme(t *testing.T) {
	if AddToScheme == nil {
		t.Error("Expected AddToScheme to not be nil")
	}
}
