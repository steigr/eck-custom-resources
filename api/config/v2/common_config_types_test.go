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
)

func TestUsernamePasswordAuthentication(t *testing.T) {
	auth := UsernamePasswordAuthentication{
		SecretName: "my-secret",
		UserName:   "admin",
	}

	if auth.SecretName != "my-secret" {
		t.Errorf("Expected SecretName to be 'my-secret', got %q", auth.SecretName)
	}

	if auth.UserName != "admin" {
		t.Errorf("Expected UserName to be 'admin', got %q", auth.UserName)
	}
}

func TestUsernamePasswordAuthentication_Empty(t *testing.T) {
	auth := UsernamePasswordAuthentication{}

	if auth.SecretName != "" {
		t.Errorf("Expected SecretName to be empty, got %q", auth.SecretName)
	}

	if auth.UserName != "" {
		t.Errorf("Expected UserName to be empty, got %q", auth.UserName)
	}
}

func TestPublicCertificate(t *testing.T) {
	cert := PublicCertificate{
		SecretName:     "my-cert-secret",
		CertificateKey: "tls.crt",
	}

	if cert.SecretName != "my-cert-secret" {
		t.Errorf("Expected SecretName to be 'my-cert-secret', got %q", cert.SecretName)
	}

	if cert.CertificateKey != "tls.crt" {
		t.Errorf("Expected CertificateKey to be 'tls.crt', got %q", cert.CertificateKey)
	}
}

func TestPublicCertificate_Empty(t *testing.T) {
	cert := PublicCertificate{}

	if cert.SecretName != "" {
		t.Errorf("Expected SecretName to be empty, got %q", cert.SecretName)
	}

	if cert.CertificateKey != "" {
		t.Errorf("Expected CertificateKey to be empty, got %q", cert.CertificateKey)
	}
}

func TestAPIKeyAuthentication(t *testing.T) {
	auth := APIKeyAuthentication{
		APIKey: "my-api-key-value",
	}

	if auth.APIKey != "my-api-key-value" {
		t.Errorf("Expected APIKey to be 'my-api-key-value', got %q", auth.APIKey)
	}
}

func TestAPIKeyAuthentication_Empty(t *testing.T) {
	auth := APIKeyAuthentication{}

	if auth.APIKey != "" {
		t.Errorf("Expected APIKey to be empty, got %q", auth.APIKey)
	}
}
