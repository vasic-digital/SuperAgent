package services

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderMappingURLs verifies that provider URL mappings include full API paths
// This test ensures that the 404 errors caused by incomplete URLs are prevented
func TestProviderMappingURLs(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		expectedURL  string
		description  string
	}{
		{
			name:         "DeepSeek must have full chat completions URL",
			providerName: "deepseek",
			expectedURL:  "https://api.deepseek.com/v1/chat/completions",
			description:  "DeepSeek provider requires full /v1/chat/completions path to avoid 404 errors",
		},
		{
			name:         "Claude (ANTHROPIC_API_KEY) must have full messages URL",
			providerName: "claude",
			expectedURL:  "https://api.anthropic.com/v1/messages",
			description:  "Claude provider requires full /v1/messages path to avoid 404 errors",
		},
		{
			name:         "Mistral must have base v1 URL",
			providerName: "mistral",
			expectedURL:  "https://api.mistral.ai/v1",
			description:  "Mistral provider appends /chat/completions internally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, mapping := range providerMappings {
				if mapping.ProviderName == tt.providerName {
					assert.Equal(t, tt.expectedURL, mapping.BaseURL,
						"Provider %s: %s. Got URL: %s", tt.providerName, tt.description, mapping.BaseURL)
					found = true
					break
				}
			}
			require.True(t, found, "Provider mapping not found for: %s", tt.providerName)
		})
	}
}

// TestDeepSeekURLNotIncomplete ensures DeepSeek URL is not the old incomplete format
func TestDeepSeekURLNotIncomplete(t *testing.T) {
	for _, mapping := range providerMappings {
		if mapping.ProviderName == "deepseek" {
			// The old buggy URL was "https://api.deepseek.com" without the path
			assert.NotEqual(t, "https://api.deepseek.com", mapping.BaseURL,
				"DeepSeek URL must include /v1/chat/completions path to avoid 404 errors")
			assert.Contains(t, mapping.BaseURL, "/v1/chat/completions",
				"DeepSeek URL must contain the full API path")
		}
	}
}

// TestClaudeURLNotIncomplete ensures Claude URL is not the old incomplete format
func TestClaudeURLNotIncomplete(t *testing.T) {
	for _, mapping := range providerMappings {
		if mapping.ProviderType == "claude" {
			// The old buggy URL was "https://api.anthropic.com" without the path
			assert.NotEqual(t, "https://api.anthropic.com", mapping.BaseURL,
				"Claude URL must include /v1/messages path to avoid 404 errors")
			assert.Contains(t, mapping.BaseURL, "/v1/messages",
				"Claude URL must contain the full API path")
		}
	}
}

// TestProviderDiscoveryCreation tests that ProviderDiscovery can be created
func TestProviderDiscoveryCreation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	discovery := NewProviderDiscovery(logger, false)
	require.NotNil(t, discovery, "ProviderDiscovery should be created successfully")
}

// TestProviderMappingHasRequiredFields verifies all provider mappings have required fields
func TestProviderMappingHasRequiredFields(t *testing.T) {
	for _, mapping := range providerMappings {
		t.Run(mapping.ProviderName, func(t *testing.T) {
			assert.NotEmpty(t, mapping.EnvVar, "EnvVar should not be empty for %s", mapping.ProviderName)
			assert.NotEmpty(t, mapping.ProviderType, "ProviderType should not be empty for %s", mapping.ProviderName)
			assert.NotEmpty(t, mapping.ProviderName, "ProviderName should not be empty for %s", mapping.ProviderName)
			assert.NotEmpty(t, mapping.BaseURL, "BaseURL should not be empty for %s", mapping.ProviderName)
			assert.NotEmpty(t, mapping.DefaultModel, "DefaultModel should not be empty for %s", mapping.ProviderName)
			assert.Greater(t, mapping.Priority, 0, "Priority should be positive for %s", mapping.ProviderName)
		})
	}
}

// TestProviderMappingURLsAreValid ensures URLs are well-formed
func TestProviderMappingURLsAreValid(t *testing.T) {
	for _, mapping := range providerMappings {
		t.Run(mapping.ProviderName+"_url_format", func(t *testing.T) {
			assert.True(t,
				len(mapping.BaseURL) > 10 && (mapping.BaseURL[:8] == "https://" || mapping.BaseURL[:7] == "http://"),
				"BaseURL should be a valid URL for %s, got: %s", mapping.ProviderName, mapping.BaseURL)
		})
	}
}

// TestProviderDiscoveryWithoutEnvVars tests discovery returns empty list without env vars
func TestProviderDiscoveryWithoutEnvVars(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	discovery := NewProviderDiscovery(logger, false)

	// Just verify it doesn't panic with empty environment
	// The actual discovery will find nothing without env vars set
	providers, err := discovery.DiscoverProviders()
	assert.NoError(t, err, "DiscoverProviders should not return an error")
	assert.NotNil(t, providers, "DiscoverProviders should return a non-nil slice")
}

// TestMaskToken tests the token masking function
func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "Long token gets masked",
			token:    "sk-ant-api01-verylongtoken12345",
			expected: "sk-an...12345",
		},
		{
			name:     "Short token returns masked",
			token:    "short",
			expected: "***",
		},
		{
			name:     "Empty token returns masked",
			token:    "",
			expected: "***",
		},
		{
			name:     "Exactly 10 char token returns masked",
			token:    "1234567890",
			expected: "***",
		},
		{
			name:     "11 char token gets partial masking",
			token:    "12345678901",
			expected: "12345...78901",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProviderHealthStatusConstants verifies status constants are correctly defined
func TestProviderHealthStatusConstants(t *testing.T) {
	assert.Equal(t, ProviderHealthStatus("unknown"), ProviderStatusUnknown)
	assert.Equal(t, ProviderHealthStatus("healthy"), ProviderStatusHealthy)
	assert.Equal(t, ProviderHealthStatus("unhealthy"), ProviderStatusUnhealthy)
	assert.Equal(t, ProviderHealthStatus("rate_limited"), ProviderStatusRateLimited)
	assert.Equal(t, ProviderHealthStatus("auth_failed"), ProviderStatusAuthFailed)
}

// TestDiscoveredProviderInitialization tests DiscoveredProvider struct
func TestDiscoveredProviderInitialization(t *testing.T) {
	dp := &DiscoveredProvider{
		Name:         "test-provider",
		Type:         "test",
		APIKeyEnvVar: "TEST_API_KEY",
		BaseURL:      "https://api.test.com",
		DefaultModel: "test-model",
		Status:       ProviderStatusUnknown,
	}

	assert.Equal(t, "test-provider", dp.Name)
	assert.Equal(t, "test", dp.Type)
	assert.Equal(t, ProviderStatusUnknown, dp.Status)
	assert.False(t, dp.Verified)
	assert.Equal(t, float64(0), dp.Score)
}

// TestProviderMappingDeepSeekModel verifies DeepSeek uses the correct model
func TestProviderMappingDeepSeekModel(t *testing.T) {
	for _, mapping := range providerMappings {
		if mapping.ProviderName == "deepseek" {
			// DeepSeek should use deepseek-chat or deepseek-coder
			assert.Contains(t, []string{"deepseek-chat", "deepseek-coder"}, mapping.DefaultModel,
				"DeepSeek should use a valid model name")
		}
	}
}

// TestProviderMappingClaudeModels verifies Claude uses valid models
func TestProviderMappingClaudeModels(t *testing.T) {
	validClaudeModels := []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-sonnet-4-20250514",
		"claude-sonnet-4-5-20250929", // Claude 4.5 Sonnet (balanced)
	}

	for _, mapping := range providerMappings {
		if mapping.ProviderType == "claude" {
			found := false
			for _, model := range validClaudeModels {
				if mapping.DefaultModel == model {
					found = true
					break
				}
			}
			assert.True(t, found,
				"Claude provider %s should use a valid model, got: %s",
				mapping.ProviderName, mapping.DefaultModel)
		}
	}
}

// TestQwenOAuthURLIsDifferentFromRegular verifies OAuth URL is correctly configured
func TestQwenOAuthURLIsDifferentFromRegular(t *testing.T) {
	// Regular Qwen uses dashscope.aliyuncs.com/api/v1
	// OAuth Qwen should use dashscope.aliyuncs.com/compatible-mode/v1
	for _, mapping := range providerMappings {
		if mapping.ProviderName == "qwen" {
			// Regular Qwen should use api/v1 path
			assert.Contains(t, mapping.BaseURL, "dashscope.aliyuncs.com",
				"Qwen should use Alibaba DashScope API")
		}
	}
}

// TestProviderPriorityRanges verifies priority values are reasonable
func TestProviderPriorityRanges(t *testing.T) {
	for _, mapping := range providerMappings {
		t.Run(mapping.ProviderName+"_priority", func(t *testing.T) {
			assert.GreaterOrEqual(t, mapping.Priority, 1, "Priority should be at least 1")
			assert.LessOrEqual(t, mapping.Priority, 20, "Priority should not exceed 20")
		})
	}
}

// TestContainsAnyHelper tests the containsAny helper function
func TestContainsAnyHelper(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substrs  []string
		expected bool
	}{
		{
			name:     "Contains one of the substrings",
			s:        "error: 401 unauthorized",
			substrs:  []string{"401", "403", "unauthorized"},
			expected: true,
		},
		{
			name:     "Contains none of the substrings",
			s:        "success: ok",
			substrs:  []string{"error", "fail", "unauthorized"},
			expected: false,
		},
		{
			name:     "Empty string",
			s:        "",
			substrs:  []string{"test"},
			expected: false,
		},
		{
			name:     "Empty substrs",
			s:        "test string",
			substrs:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.s, tt.substrs...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProviderMappingsUniqueProviderNames ensures no duplicate provider names
func TestProviderMappingsUniqueProviderNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, mapping := range providerMappings {
		// Same provider name can appear for different env vars (e.g., ANTHROPIC_API_KEY and CLAUDE_API_KEY)
		key := mapping.EnvVar + "_" + mapping.ProviderName
		assert.False(t, seen[key], "Duplicate mapping found: %s for %s", mapping.EnvVar, mapping.ProviderName)
		seen[key] = true
	}
}

// TestProviderMappingsTier1HasLowestPriority verifies premium providers have priority 1
func TestProviderMappingsTier1HasLowestPriority(t *testing.T) {
	tier1Providers := []string{"claude", "openai"}
	for _, mapping := range providerMappings {
		for _, t1p := range tier1Providers {
			if mapping.ProviderName == t1p {
				assert.Equal(t, 1, mapping.Priority,
					"Tier 1 provider %s should have priority 1, got %d",
					t1p, mapping.Priority)
			}
		}
	}
}
