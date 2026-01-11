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

package v2

import (
	"testing"
)

func TestKibanaSpec(t *testing.T) {
	spec := KibanaSpec{
		Enabled: true,
		Url:     "https://kibana.example.com:5601",
	}

	if !spec.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if spec.Url != "https://kibana.example.com:5601" {
		t.Errorf("Expected Url to be 'https://kibana.example.com:5601', got %q", spec.Url)
	}
}

func TestKibanaSpec_Empty(t *testing.T) {
	spec := KibanaSpec{}

	if spec.Enabled {
		t.Error("Expected Enabled to be false by default")
	}

	if spec.Url != "" {
		t.Errorf("Expected Url to be empty, got %q", spec.Url)
	}

	if spec.Certificate != nil {
		t.Error("Expected Certificate to be nil by default")
	}

	if spec.Authentication != nil {
		t.Error("Expected Authentication to be nil by default")
	}
}

func TestKibanaSpec_WithCertificate(t *testing.T) {
	spec := KibanaSpec{
		Enabled: true,
		Url:     "https://kibana.example.com:5601",
		Certificate: &PublicCertificate{
			SecretName:     "kb-cert",
			CertificateKey: "ca.crt",
		},
	}

	if spec.Certificate == nil {
		t.Fatal("Expected Certificate to not be nil")
	}

	if spec.Certificate.SecretName != "kb-cert" {
		t.Errorf("Expected Certificate.SecretName to be 'kb-cert', got %q", spec.Certificate.SecretName)
	}

	if spec.Certificate.CertificateKey != "ca.crt" {
		t.Errorf("Expected Certificate.CertificateKey to be 'ca.crt', got %q", spec.Certificate.CertificateKey)
	}
}

func TestKibanaSpec_WithUsernamePasswordAuth(t *testing.T) {
	spec := KibanaSpec{
		Enabled: true,
		Url:     "https://kibana.example.com:5601",
		Authentication: &KibanaAuthentication{
			UsernamePassword: &UsernamePasswordAuthentication{
				SecretName: "kb-credentials",
				UserName:   "kibana_system",
			},
		},
	}

	if spec.Authentication == nil {
		t.Fatal("Expected Authentication to not be nil")
	}

	if spec.Authentication.UsernamePassword == nil {
		t.Fatal("Expected Authentication.UsernamePassword to not be nil")
	}

	if spec.Authentication.UsernamePassword.SecretName != "kb-credentials" {
		t.Errorf("Expected SecretName to be 'kb-credentials', got %q", spec.Authentication.UsernamePassword.SecretName)
	}

	if spec.Authentication.UsernamePassword.UserName != "kibana_system" {
		t.Errorf("Expected UserName to be 'kibana_system', got %q", spec.Authentication.UsernamePassword.UserName)
	}
}

func TestKibanaSpec_WithAPIKeyAuth(t *testing.T) {
	spec := KibanaSpec{
		Enabled: true,
		Url:     "https://kibana.example.com:5601",
		Authentication: &KibanaAuthentication{
			APIKey: &APIKeyAuthentication{
				APIKey: "kibana-api-key",
			},
		},
	}

	if spec.Authentication == nil {
		t.Fatal("Expected Authentication to not be nil")
	}

	if spec.Authentication.APIKey == nil {
		t.Fatal("Expected Authentication.APIKey to not be nil")
	}

	if spec.Authentication.APIKey.APIKey != "kibana-api-key" {
		t.Errorf("Expected APIKey to be 'kibana-api-key', got %q", spec.Authentication.APIKey.APIKey)
	}
}

func TestKibanaAuthentication_Empty(t *testing.T) {
	auth := KibanaAuthentication{}

	if auth.UsernamePassword != nil {
		t.Error("Expected UsernamePassword to be nil by default")
	}

	if auth.APIKey != nil {
		t.Error("Expected APIKey to be nil by default")
	}
}

func TestKibanaSpec_FullConfig(t *testing.T) {
	spec := KibanaSpec{
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
	}

	// Verify all fields are set correctly
	if !spec.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if spec.Url != "https://kibana.example.com:5601" {
		t.Errorf("Unexpected Url: %q", spec.Url)
	}

	if spec.Certificate == nil || spec.Certificate.SecretName != "kb-cert" {
		t.Error("Certificate not set correctly")
	}

	if spec.Authentication == nil || spec.Authentication.UsernamePassword == nil {
		t.Error("Authentication not set correctly")
	}

	if spec.Authentication.UsernamePassword.UserName != "kibana_system" {
		t.Error("Authentication UserName not set correctly")
	}
}
