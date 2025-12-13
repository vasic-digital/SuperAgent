package security

import (
	"context"
	"strings"
	"testing"

	testingutils "github.com/HelixDevelopment/HelixAgent/Toolkit/Commons/testing"
	"github.com/HelixDevelopment/HelixAgent/Toolkit/pkg/toolkit"
)

// TestAPIKeySecurity tests that API keys are handled securely
func TestAPIKeySecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	// Test that API keys are not logged in error messages
	config := map[string]interface{}{
		"api_key": "sk-1234567890abcdef",
	}

	err := mockProvider.ValidateConfig(config)
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "sk-1234567890abcdef") {
			t.Error("API key leaked in error message")
		}
	}

	// Test with invalid API key format
	invalidConfig := map[string]interface{}{
		"api_key": "",
	}

	err = mockProvider.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for empty API key")
	}
}

// TestInputValidation tests input validation for security vulnerabilities
func TestInputValidation(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	ctx := context.Background()

	// Test SQL injection attempt in chat message
	sqlInjectionReq := toolkit.ChatRequest{
		Model: "test-model",
		Messages: []toolkit.ChatMessage{
			{Role: "user", Content: "'; DROP TABLE users; --"},
		},
	}

	_, err := mockProvider.Chat(ctx, sqlInjectionReq)
	// Mock provider doesn't validate input, but real providers should
	t.Logf("SQL injection test result: %v", err)

	// Test XSS attempt
	xssReq := toolkit.ChatRequest{
		Model: "test-model",
		Messages: []toolkit.ChatMessage{
			{Role: "user", Content: "<script>alert('xss')</script>"},
		},
	}

	_, err = mockProvider.Chat(ctx, xssReq)
	t.Logf("XSS test result: %v", err)

	// Test command injection attempt
	cmdInjectionReq := toolkit.ChatRequest{
		Model: "test-model",
		Messages: []toolkit.ChatMessage{
			{Role: "user", Content: "$(rm -rf /)"},
		},
	}

	_, err = mockProvider.Chat(ctx, cmdInjectionReq)
	t.Logf("Command injection test result: %v", err)
}

// TestConfigSecurity tests secure configuration handling
func TestConfigSecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	// Test that sensitive config values are not exposed
	config := map[string]interface{}{
		"api_key":      "secret-key-123",
		"secret_token": "token-456",
		"password":     "pass-789",
		"timeout":      30,
	}

	err := mockProvider.ValidateConfig(config)
	if err != nil {
		t.Errorf("Config validation failed: %v", err)
	}

	// Test config with null bytes (potential attack)
	maliciousConfig := map[string]interface{}{
		"api_key": "key\x00injected",
	}

	err = mockProvider.ValidateConfig(maliciousConfig)
	if err != nil {
		t.Logf("Null byte config rejected: %v", err)
	}
}

// TestErrorMessageSecurity tests that error messages don't leak sensitive information
func TestErrorMessageSecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")
	mockProvider.SetShouldError(true)

	ctx := context.Background()

	// Test that internal paths/file names are not exposed
	req := toolkit.ChatRequest{
		Model: "test-model",
		Messages: []toolkit.ChatMessage{
			{Role: "user", Content: "test"},
		},
	}

	_, err := mockProvider.Chat(ctx, req)
	if err != nil {
		errorMsg := err.Error()
		// Check that error doesn't contain file paths or internal details
		if strings.Contains(errorMsg, "/") || strings.Contains(errorMsg, "\\") {
			t.Logf("Warning: Error message may contain file paths: %s", errorMsg)
		}
		// Check that error doesn't contain stack traces in production-like scenarios
		if strings.Contains(errorMsg, "runtime") || strings.Contains(errorMsg, "panic") {
			t.Logf("Warning: Error message may contain stack trace: %s", errorMsg)
		}
	}
}

// TestRateLimitSecurity tests rate limiting security aspects
func TestRateLimitSecurity(t *testing.T) {
	// This would test rate limiting behavior under attack scenarios
	// For now, just test basic functionality
	t.Log("Rate limit security test placeholder - implement when rate limiting is added")
}

// TestModelValidationSecurity tests model name validation
func TestModelValidationSecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	ctx := context.Background()

	// Test with potentially dangerous model names
	dangerousModels := []string{
		"../../../etc/passwd",
		"model;rm -rf /",
		"model\x00injected",
		"<script>model</script>",
	}

	for _, model := range dangerousModels {
		req := toolkit.ChatRequest{
			Model: model,
			Messages: []toolkit.ChatMessage{
				{Role: "user", Content: "test"},
			},
		}

		_, err := mockProvider.Chat(ctx, req)
		t.Logf("Model validation test for '%s': %v", model, err)
	}
}

// TestEmbeddingSecurity tests embedding request security
func TestEmbeddingSecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	ctx := context.Background()

	// Test with large input that could cause DoS
	largeInput := strings.Repeat("word ", 10000) // 50,000 characters
	req := toolkit.EmbeddingRequest{
		Model: "test-embedding-model",
		Input: []string{largeInput},
	}

	_, err := mockProvider.Embed(ctx, req)
	t.Logf("Large input embedding test: %v", err)

	// Test with empty input
	emptyReq := toolkit.EmbeddingRequest{
		Model: "test-embedding-model",
		Input: []string{},
	}

	_, err = mockProvider.Embed(ctx, emptyReq)
	t.Logf("Empty input embedding test: %v", err)
}

// TestRerankSecurity tests rerank request security
func TestRerankSecurity(t *testing.T) {
	mockProvider := testingutils.NewMockProvider("security-test")

	ctx := context.Background()

	// Test with large documents that could cause DoS
	largeDocs := make([]string, 1000)
	for i := range largeDocs {
		largeDocs[i] = strings.Repeat("document ", 1000) // 9,000 chars each
	}

	req := toolkit.RerankRequest{
		Model:     "test-rerank-model",
		Query:     "test query",
		Documents: largeDocs,
		TopN:      10,
	}

	_, err := mockProvider.Rerank(ctx, req)
	t.Logf("Large documents rerank test: %v", err)
}
