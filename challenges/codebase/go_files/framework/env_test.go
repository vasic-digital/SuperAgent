package framework

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvLoader_Load(t *testing.T) {
	// Create temp directory and .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `# Test environment
ANTHROPIC_API_KEY=sk-ant-test-key-12345
OPENAI_API_KEY=sk-test-openai-key
SIMPLE_VALUE=hello
QUOTED_VALUE="quoted string"
SINGLE_QUOTED='single quoted'
`

	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test .env: %v", err)
	}

	loader := NewEnvLoader()
	if err := loader.Load(envPath); err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"ANTHROPIC_API_KEY", "sk-ant-test-key-12345"},
		{"OPENAI_API_KEY", "sk-test-openai-key"},
		{"SIMPLE_VALUE", "hello"},
		{"QUOTED_VALUE", "quoted string"},
		{"SINGLE_QUOTED", "single quoted"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := loader.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Get(%s) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestEnvLoader_LoadNonExistent(t *testing.T) {
	loader := NewEnvLoader()
	err := loader.Load("/nonexistent/path/.env")

	// Should not error - just use system env
	if err != nil {
		t.Errorf("Load() for nonexistent file should not error: %v", err)
	}

	if !loader.IsLoaded() {
		t.Error("IsLoaded() should be true after Load()")
	}
}

func TestEnvLoader_GetRequired(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["EXISTS"] = "value"

	// Existing key
	val, err := loader.GetRequired("EXISTS")
	if err != nil {
		t.Errorf("GetRequired() failed for existing key: %v", err)
	}
	if val != "value" {
		t.Errorf("GetRequired() = %q, want %q", val, "value")
	}

	// Non-existing key
	_, err = loader.GetRequired("NOT_EXISTS")
	if err == nil {
		t.Error("GetRequired() should fail for non-existing key")
	}
}

func TestEnvLoader_GetWithDefault(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["EXISTS"] = "value"

	// Existing key
	got := loader.GetWithDefault("EXISTS", "default")
	if got != "value" {
		t.Errorf("GetWithDefault() = %q, want %q", got, "value")
	}

	// Non-existing key
	got = loader.GetWithDefault("NOT_EXISTS", "default")
	if got != "default" {
		t.Errorf("GetWithDefault() = %q, want %q", got, "default")
	}
}

func TestEnvLoader_GetAPIKey(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["ANTHROPIC_API_KEY"] = "sk-ant-test"
	loader.values["OPENAI_API_KEY"] = "sk-openai-test"
	loader.values["CUSTOM_PROVIDER_API_KEY"] = "custom-key"

	tests := []struct {
		provider string
		expected string
	}{
		{"anthropic", "sk-ant-test"},
		{"ANTHROPIC", "sk-ant-test"}, // Case insensitive
		{"openai", "sk-openai-test"},
		{"custom_provider", "custom-key"}, // Fallback pattern
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := loader.GetAPIKey(tt.provider)
			if got != tt.expected {
				t.Errorf("GetAPIKey(%s) = %q, want %q", tt.provider, got, tt.expected)
			}
		})
	}
}

func TestEnvLoader_ListConfiguredProviders(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["ANTHROPIC_API_KEY"] = "sk-ant-test"
	loader.values["OPENAI_API_KEY"] = "sk-openai-test"

	providers := loader.ListConfiguredProviders()

	if len(providers) < 2 {
		t.Errorf("ListConfiguredProviders() returned %d providers, expected at least 2", len(providers))
	}

	// Check that expected providers are present
	found := make(map[string]bool)
	for _, p := range providers {
		found[p] = true
	}

	if !found["anthropic"] {
		t.Error("anthropic should be in configured providers")
	}
	if !found["openai"] {
		t.Error("openai should be in configured providers")
	}
}

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "***"},
		{"abcd", "****"},
		{"abcdefgh", "ab******"},
		{"sk-ant-api-123456789", "sk-a****************"},
		{"sk-openai-very-long-key-12345", "sk-o*************************"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactAPIKey(tt.input)
			if got != tt.expected {
				t.Errorf("RedactAPIKey(%q) = %q, want %q", tt.input, got, tt.expected)
			}

			// Verify length is preserved
			if len(got) != len(tt.input) {
				t.Errorf("RedactAPIKey(%q) length = %d, want %d", tt.input, len(got), len(tt.input))
			}
		})
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"https://api.example.com/v1?api_key=secret123",
			"https://api.example.com/v1?api_key=****",
		},
		{
			"https://api.example.com/v1?key=secret&other=value",
			"https://api.example.com/v1?key=****&other=value",
		},
		{
			"https://api.example.com/v1?token=mytoken",
			"https://api.example.com/v1?token=****",
		},
		{
			"https://api.example.com/v1", // No params to redact
			"https://api.example.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactURL(tt.input)
			if got != tt.expected {
				t.Errorf("RedactURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactHeaders(t *testing.T) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer sk-secret-token",
		"X-API-Key":     "my-api-key",
		"X-Request-ID":  "12345",
	}

	redacted := RedactHeaders(headers)

	if redacted["Content-Type"] != "application/json" {
		t.Error("Content-Type should not be redacted")
	}
	if redacted["X-Request-ID"] != "12345" {
		t.Error("X-Request-ID should not be redacted")
	}
	if redacted["Authorization"] != "****" {
		t.Errorf("Authorization should be redacted, got %q", redacted["Authorization"])
	}
	if redacted["X-API-Key"] != "****" {
		t.Errorf("X-API-Key should be redacted, got %q", redacted["X-API-Key"])
	}
}

func TestEnvLoader_ResolveTemplate(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["API_KEY"] = "secret123"
	loader.values["BASE_URL"] = "https://api.example.com"

	tests := []struct {
		template string
		expected string
	}{
		{"${API_KEY}", "secret123"},
		{"${BASE_URL}/v1", "https://api.example.com/v1"},
		{"Key: ${API_KEY}, URL: ${BASE_URL}", "Key: secret123, URL: https://api.example.com"},
		{"No variables here", "No variables here"},
		{"${NONEXISTENT}", ""}, // Missing var resolves to empty
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			got := loader.ResolveTemplate(tt.template)
			if got != tt.expected {
				t.Errorf("ResolveTemplate(%q) = %q, want %q", tt.template, got, tt.expected)
			}
		})
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	tests := []struct {
		provider string
		key      string
		wantErr  bool
	}{
		{"anthropic", "sk-ant-api-123456", false},
		{"anthropic", "invalid-key", true},
		{"openai", "sk-123456", false},
		{"openai", "invalid", true},
		{"openrouter", "sk-or-v1-123456", false},
		{"nvidia", "nvapi-123456", false},
		{"nvidia", "sk-123", true},
		{"huggingface", "hf_abcdef", false},
		{"unknown", "any-key", false}, // Unknown providers always pass
	}

	for _, tt := range tests {
		t.Run(tt.provider+"_"+tt.key, func(t *testing.T) {
			err := ValidateAPIKeyFormat(tt.provider, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKeyFormat(%s, %s) error = %v, wantErr %v", tt.provider, tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestEnvLoader_GetAllRedacted(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["ANTHROPIC_API_KEY"] = "sk-ant-secret-12345"
	loader.values["DEBUG_MODE"] = "true"
	loader.values["PASSWORD"] = "mysecret"
	loader.values["PORT"] = "7061"

	redacted := loader.GetAllRedacted()

	// API key should be redacted
	if redacted["ANTHROPIC_API_KEY"] == "sk-ant-secret-12345" {
		t.Error("ANTHROPIC_API_KEY should be redacted")
	}
	// Check that it starts with "sk-a" and rest is masked
	if !hasPrefix(redacted["ANTHROPIC_API_KEY"], "sk-a") {
		t.Errorf("ANTHROPIC_API_KEY redacted incorrectly: %q", redacted["ANTHROPIC_API_KEY"])
	}

	// Password should be redacted
	if redacted["PASSWORD"] == "mysecret" {
		t.Error("PASSWORD should be redacted")
	}

	// Non-secret values should not be redacted
	if redacted["DEBUG_MODE"] != "true" {
		t.Error("DEBUG_MODE should not be redacted")
	}
	if redacted["PORT"] != "7061" {
		t.Error("PORT should not be redacted")
	}
}

func TestEnvLoader_WriteRedactedEnv(t *testing.T) {
	loader := NewEnvLoader()
	loader.values["API_KEY"] = "sk-secret-key"
	loader.values["DEBUG"] = "true"

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, ".env.redacted")

	if err := loader.WriteRedactedEnv(outPath); err != nil {
		t.Fatalf("WriteRedactedEnv() failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Should contain redacted key
	if !contains(contentStr, "API_KEY=sk-s*********") {
		t.Error("API_KEY should be redacted in output")
	}

	// Should contain non-secret value
	if !contains(contentStr, "DEBUG=true") {
		t.Error("DEBUG should not be redacted")
	}

	// Should not contain actual secret
	if contains(contentStr, "sk-secret-key") {
		t.Error("Actual secret should not appear in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
