package config

import (
	"os"
	"path/filepath"
	"testing"

	appv2 "eck-custom-resources/api/config/v2"
)

func TestDefaultSpec(t *testing.T) {
	spec := defaultSpec()

	// Verify Elasticsearch defaults
	if spec.Elasticsearch.Url == "" {
		t.Error("defaultSpec() Elasticsearch.Url should not be empty")
	}
	if spec.Elasticsearch.Authentication == nil {
		t.Error("defaultSpec() Elasticsearch.Authentication should not be nil")
	}
	if spec.Elasticsearch.Authentication.UsernamePassword == nil {
		t.Error("defaultSpec() Elasticsearch.Authentication.UsernamePassword should not be nil")
	}
	if spec.Elasticsearch.Certificate == nil {
		t.Error("defaultSpec() Elasticsearch.Certificate should not be nil")
	}

	// Verify Kibana defaults
	if spec.Kibana.Url == "" {
		t.Error("defaultSpec() Kibana.Url should not be empty")
	}
	if spec.Kibana.Authentication == nil {
		t.Error("defaultSpec() Kibana.Authentication should not be nil")
	}
	if spec.Kibana.Authentication.UsernamePassword == nil {
		t.Error("defaultSpec() Kibana.Authentication.UsernamePassword should not be nil")
	}
	if spec.Kibana.Certificate == nil {
		t.Error("defaultSpec() Kibana.Certificate should not be nil")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    appv2.ProjectConfigSpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid spec",
			spec: appv2.ProjectConfigSpec{
				Elasticsearch: appv2.ElasticsearchSpec{
					Url: "https://elasticsearch.example.com",
				},
				Kibana: appv2.KibanaSpec{
					Url: "https://kibana.example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "missing elasticsearch url",
			spec: appv2.ProjectConfigSpec{
				Elasticsearch: appv2.ElasticsearchSpec{
					Url: "",
				},
				Kibana: appv2.KibanaSpec{
					Url: "https://kibana.example.com",
				},
			},
			wantErr: true,
			errMsg:  "elasticsearch.endpoint is required",
		},
		{
			name: "missing kibana url",
			spec: appv2.ProjectConfigSpec{
				Elasticsearch: appv2.ElasticsearchSpec{
					Url: "https://elasticsearch.example.com",
				},
				Kibana: appv2.KibanaSpec{
					Url: "",
				},
			},
			wantErr: true,
			errMsg:  "kibana.url is required",
		},
		{
			name: "both urls missing",
			spec: appv2.ProjectConfigSpec{
				Elasticsearch: appv2.ElasticsearchSpec{
					Url: "",
				},
				Kibana: appv2.KibanaSpec{
					Url: "",
				},
			},
			wantErr: true,
			errMsg:  "elasticsearch.endpoint is required", // First error returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoadProjectConfigSpec_EmptyPath(t *testing.T) {
	spec, err := LoadProjectConfigSpec("")

	if err != nil {
		t.Errorf("LoadProjectConfigSpec(\"\") unexpected error: %v", err)
	}

	// Should return default spec
	if spec.Elasticsearch.Url == "" {
		t.Error("LoadProjectConfigSpec(\"\") should return default spec with Elasticsearch.Url")
	}
	if spec.Kibana.Url == "" {
		t.Error("LoadProjectConfigSpec(\"\") should return default spec with Kibana.Url")
	}
}

func TestLoadProjectConfigSpec_NonExistentFile(t *testing.T) {
	_, err := LoadProjectConfigSpec("/non/existent/path/config.yaml")

	if err == nil {
		t.Error("LoadProjectConfigSpec() expected error for non-existent file, got nil")
	}
}

func TestLoadProjectConfigSpec_ValidYAML(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
elasticsearch:
  url: "https://custom-es.example.com"
  enabled: true
kibana:
  url: "https://custom-kb.example.com"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	spec, err := LoadProjectConfigSpec(configPath)

	if err != nil {
		t.Errorf("LoadProjectConfigSpec() unexpected error: %v", err)
		return
	}

	if spec.Elasticsearch.Url != "https://custom-es.example.com" {
		t.Errorf("LoadProjectConfigSpec() Elasticsearch.Url = %q, want %q",
			spec.Elasticsearch.Url, "https://custom-es.example.com")
	}

	if spec.Kibana.Url != "https://custom-kb.example.com" {
		t.Errorf("LoadProjectConfigSpec() Kibana.Url = %q, want %q",
			spec.Kibana.Url, "https://custom-kb.example.com")
	}
}

func TestLoadProjectConfigSpec_InvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `
elasticsearch:
  url: [invalid yaml structure
  this is not valid
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	_, err := LoadProjectConfigSpec(configPath)

	if err == nil {
		t.Error("LoadProjectConfigSpec() expected error for invalid YAML, got nil")
	}
}

func TestLoadProjectConfigSpec_ValidationError(t *testing.T) {
	// Create a YAML file that fails validation (empty URLs)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")

	yamlContent := `
elasticsearch:
  url: ""
kibana:
  url: ""
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	_, err := LoadProjectConfigSpec(configPath)

	if err == nil {
		t.Error("LoadProjectConfigSpec() expected validation error, got nil")
	}
}

func TestLoadProjectConfigSpec_EnvVarExpansion(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("TEST_ES_URL", "https://env-es.example.com")
	os.Setenv("TEST_KB_URL", "https://env-kb.example.com")
	defer func() {
		os.Unsetenv("TEST_ES_URL")
		os.Unsetenv("TEST_KB_URL")
	}()

	// Create a YAML file with environment variable placeholders
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "env-config.yaml")

	yamlContent := `
elasticsearch:
  url: "${TEST_ES_URL}"
kibana:
  url: "${TEST_KB_URL}"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	spec, err := LoadProjectConfigSpec(configPath)

	if err != nil {
		t.Errorf("LoadProjectConfigSpec() unexpected error: %v", err)
		return
	}

	if spec.Elasticsearch.Url != "https://env-es.example.com" {
		t.Errorf("LoadProjectConfigSpec() Elasticsearch.Url = %q, want %q (env var not expanded)",
			spec.Elasticsearch.Url, "https://env-es.example.com")
	}

	if spec.Kibana.Url != "https://env-kb.example.com" {
		t.Errorf("LoadProjectConfigSpec() Kibana.Url = %q, want %q (env var not expanded)",
			spec.Kibana.Url, "https://env-kb.example.com")
	}
}

func TestLoadProjectConfigSpec_PartialConfig(t *testing.T) {
	// Create a YAML file with only some fields - should merge with defaults
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial-config.yaml")

	yamlContent := `
elasticsearch:
  url: "https://partial-es.example.com"
kibana:
  url: "https://partial-kb.example.com"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	spec, err := LoadProjectConfigSpec(configPath)

	if err != nil {
		t.Errorf("LoadProjectConfigSpec() unexpected error: %v", err)
		return
	}

	// Custom URLs should be loaded
	if spec.Elasticsearch.Url != "https://partial-es.example.com" {
		t.Errorf("LoadProjectConfigSpec() Elasticsearch.Url = %q, want %q",
			spec.Elasticsearch.Url, "https://partial-es.example.com")
	}

	if spec.Kibana.Url != "https://partial-kb.example.com" {
		t.Errorf("LoadProjectConfigSpec() Kibana.Url = %q, want %q",
			spec.Kibana.Url, "https://partial-kb.example.com")
	}
}

func TestLoadProjectConfigSpec_FullConfig(t *testing.T) {
	// Create a complete YAML configuration
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "full-config.yaml")

	yamlContent := `
elasticsearch:
  url: "https://full-es.example.com"
  enabled: true
  authentication:
    usernamePasswordSecret:
      secretName: "my-es-secret"
      userName: "admin"
  certificate:
    secretName: "my-es-cert"
    certificateKey: "tls.crt"
kibana:
  url: "https://full-kb.example.com"
  authentication:
    usernamePasswordSecret:
      secretName: "my-kb-secret"
      userName: "kibana-user"
  certificate:
    secretName: "my-kb-cert"
    certificateKey: "tls.crt"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	spec, err := LoadProjectConfigSpec(configPath)

	if err != nil {
		t.Errorf("LoadProjectConfigSpec() unexpected error: %v", err)
		return
	}

	// Verify Elasticsearch config
	if spec.Elasticsearch.Url != "https://full-es.example.com" {
		t.Errorf("Elasticsearch.Url = %q, want %q", spec.Elasticsearch.Url, "https://full-es.example.com")
	}
	if !spec.Elasticsearch.Enabled {
		t.Error("Elasticsearch.Enabled should be true")
	}
	if spec.Elasticsearch.Authentication == nil {
		t.Fatal("Elasticsearch.Authentication should not be nil")
	}
	if spec.Elasticsearch.Authentication.UsernamePassword == nil {
		t.Fatal("Elasticsearch.Authentication.UsernamePassword should not be nil")
	}
	if spec.Elasticsearch.Authentication.UsernamePassword.SecretName != "my-es-secret" {
		t.Errorf("Elasticsearch.Authentication.UsernamePassword.SecretName = %q, want %q",
			spec.Elasticsearch.Authentication.UsernamePassword.SecretName, "my-es-secret")
	}
	if spec.Elasticsearch.Certificate == nil {
		t.Fatal("Elasticsearch.Certificate should not be nil")
	}
	if spec.Elasticsearch.Certificate.SecretName != "my-es-cert" {
		t.Errorf("Elasticsearch.Certificate.SecretName = %q, want %q",
			spec.Elasticsearch.Certificate.SecretName, "my-es-cert")
	}

	// Verify Kibana config
	if spec.Kibana.Url != "https://full-kb.example.com" {
		t.Errorf("Kibana.Url = %q, want %q", spec.Kibana.Url, "https://full-kb.example.com")
	}
	if spec.Kibana.Authentication == nil {
		t.Fatal("Kibana.Authentication should not be nil")
	}
	if spec.Kibana.Authentication.UsernamePassword == nil {
		t.Fatal("Kibana.Authentication.UsernamePassword should not be nil")
	}
	if spec.Kibana.Authentication.UsernamePassword.SecretName != "my-kb-secret" {
		t.Errorf("Kibana.Authentication.UsernamePassword.SecretName = %q, want %q",
			spec.Kibana.Authentication.UsernamePassword.SecretName, "my-kb-secret")
	}
}

func TestLoadProjectConfigSpec_RelativePath(t *testing.T) {
	// Create a config in current working directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	yamlContent := `
elasticsearch:
  url: "https://relative-es.example.com"
kibana:
  url: "https://relative-kb.example.com"
`

	if err := os.WriteFile("config.yaml", []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	spec, err := LoadProjectConfigSpec("config.yaml")

	if err != nil {
		t.Errorf("LoadProjectConfigSpec() with relative path unexpected error: %v", err)
		return
	}

	if spec.Elasticsearch.Url != "https://relative-es.example.com" {
		t.Errorf("LoadProjectConfigSpec() Elasticsearch.Url = %q, want %q",
			spec.Elasticsearch.Url, "https://relative-es.example.com")
	}
}
