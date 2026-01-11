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
)

func TestAPIKeyAuthentication_DeepCopy(t *testing.T) {
	original := &APIKeyAuthentication{
		APIKey: "test-api-key",
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.APIKey != original.APIKey {
		t.Errorf("Expected APIKey %q, got %q", original.APIKey, copy.APIKey)
	}

	// Modify copy and verify original is unchanged
	copy.APIKey = "modified"
	if original.APIKey == copy.APIKey {
		t.Error("Modifying copy should not affect original")
	}
}

func TestAPIKeyAuthentication_DeepCopyNil(t *testing.T) {
	var original *APIKeyAuthentication
	copy := original.DeepCopy()

	if copy != nil {
		t.Error("DeepCopy of nil should return nil")
	}
}

func TestUsernamePasswordAuthentication_DeepCopy(t *testing.T) {
	original := &UsernamePasswordAuthentication{
		SecretName: "test-secret",
		UserName:   "test-user",
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.SecretName != original.SecretName {
		t.Errorf("Expected SecretName %q, got %q", original.SecretName, copy.SecretName)
	}

	if copy.UserName != original.UserName {
		t.Errorf("Expected UserName %q, got %q", original.UserName, copy.UserName)
	}
}

func TestPublicCertificate_DeepCopy(t *testing.T) {
	original := &PublicCertificate{
		SecretName:     "cert-secret",
		CertificateKey: "tls.crt",
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.SecretName != original.SecretName {
		t.Errorf("Expected SecretName %q, got %q", original.SecretName, copy.SecretName)
	}

	if copy.CertificateKey != original.CertificateKey {
		t.Errorf("Expected CertificateKey %q, got %q", original.CertificateKey, copy.CertificateKey)
	}
}

func TestElasticsearchAuthentication_DeepCopy(t *testing.T) {
	original := &ElasticsearchAuthentication{
		UsernamePassword: &UsernamePasswordAuthentication{
			SecretName: "es-secret",
			UserName:   "elastic",
		},
		APIKey: &APIKeyAuthentication{
			APIKey: "es-api-key",
		},
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.UsernamePassword == original.UsernamePassword {
		t.Error("DeepCopy should deep copy UsernamePassword")
	}

	if copy.APIKey == original.APIKey {
		t.Error("DeepCopy should deep copy APIKey")
	}

	if copy.UsernamePassword.SecretName != original.UsernamePassword.SecretName {
		t.Error("UsernamePassword.SecretName should be equal")
	}

	if copy.APIKey.APIKey != original.APIKey.APIKey {
		t.Error("APIKey.APIKey should be equal")
	}
}

func TestElasticsearchSpec_DeepCopy(t *testing.T) {
	original := &ElasticsearchSpec{
		Enabled: true,
		Url:     "https://es.example.com:9200",
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
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.Enabled != original.Enabled {
		t.Error("Enabled should be equal")
	}

	if copy.Url != original.Url {
		t.Error("Url should be equal")
	}

	if copy.Certificate == original.Certificate {
		t.Error("DeepCopy should deep copy Certificate")
	}

	if copy.Authentication == original.Authentication {
		t.Error("DeepCopy should deep copy Authentication")
	}
}

func TestKibanaAuthentication_DeepCopy(t *testing.T) {
	original := &KibanaAuthentication{
		UsernamePassword: &UsernamePasswordAuthentication{
			SecretName: "kb-secret",
			UserName:   "kibana_system",
		},
		APIKey: &APIKeyAuthentication{
			APIKey: "kb-api-key",
		},
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.UsernamePassword == original.UsernamePassword {
		t.Error("DeepCopy should deep copy UsernamePassword")
	}

	if copy.APIKey == original.APIKey {
		t.Error("DeepCopy should deep copy APIKey")
	}
}

func TestKibanaSpec_DeepCopy(t *testing.T) {
	original := &KibanaSpec{
		Enabled: true,
		Url:     "https://kb.example.com:5601",
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
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.Enabled != original.Enabled {
		t.Error("Enabled should be equal")
	}

	if copy.Url != original.Url {
		t.Error("Url should be equal")
	}

	if copy.Certificate == original.Certificate {
		t.Error("DeepCopy should deep copy Certificate")
	}

	if copy.Authentication == original.Authentication {
		t.Error("DeepCopy should deep copy Authentication")
	}
}

func TestProjectConfigSpec_DeepCopy(t *testing.T) {
	original := &ProjectConfigSpec{
		Elasticsearch: ElasticsearchSpec{
			Enabled: true,
			Url:     "https://es.example.com:9200",
		},
		Kibana: KibanaSpec{
			Enabled: true,
			Url:     "https://kb.example.com:5601",
		},
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.Elasticsearch.Url != original.Elasticsearch.Url {
		t.Error("Elasticsearch.Url should be equal")
	}

	if copy.Kibana.Url != original.Kibana.Url {
		t.Error("Kibana.Url should be equal")
	}
}

func TestProjectConfigStatus_DeepCopy(t *testing.T) {
	original := &ProjectConfigStatus{}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}
}

func TestProjectConfig_DeepCopy(t *testing.T) {
	original := &ProjectConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProjectConfig",
			APIVersion: "config.github.com/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: ProjectConfigSpec{
			Elasticsearch: ElasticsearchSpec{
				Enabled: true,
				Url:     "https://es.example.com:9200",
			},
		},
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if copy.Name != original.Name {
		t.Error("Name should be equal")
	}

	if copy.Spec.Elasticsearch.Url != original.Spec.Elasticsearch.Url {
		t.Error("Spec.Elasticsearch.Url should be equal")
	}
}

func TestProjectConfig_DeepCopyObject(t *testing.T) {
	original := &ProjectConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-config",
		},
	}

	copy := original.DeepCopyObject()

	if copy == nil {
		t.Error("DeepCopyObject should not return nil")
	}

	projectConfigCopy, ok := copy.(*ProjectConfig)
	if !ok {
		t.Error("DeepCopyObject should return *ProjectConfig")
	}

	if projectConfigCopy.Name != original.Name {
		t.Error("Name should be equal")
	}
}

func TestProjectConfigList_DeepCopy(t *testing.T) {
	original := &ProjectConfigList{
		Items: []ProjectConfig{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "config-1"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "config-2"},
			},
		},
	}

	copy := original.DeepCopy()

	if copy == original {
		t.Error("DeepCopy should return a new object, not the same pointer")
	}

	if len(copy.Items) != len(original.Items) {
		t.Errorf("Expected %d items, got %d", len(original.Items), len(copy.Items))
	}

	// Verify items are also deep copied
	if &copy.Items[0] == &original.Items[0] {
		t.Error("Items should be deep copied")
	}
}

func TestProjectConfigList_DeepCopyObject(t *testing.T) {
	original := &ProjectConfigList{
		Items: []ProjectConfig{
			{ObjectMeta: metav1.ObjectMeta{Name: "config-1"}},
		},
	}

	copy := original.DeepCopyObject()

	if copy == nil {
		t.Error("DeepCopyObject should not return nil")
	}

	listCopy, ok := copy.(*ProjectConfigList)
	if !ok {
		t.Error("DeepCopyObject should return *ProjectConfigList")
	}

	if len(listCopy.Items) != len(original.Items) {
		t.Error("Items length should be equal")
	}
}
