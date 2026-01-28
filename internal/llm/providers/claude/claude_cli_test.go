package claude

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestClaudeCLIProvider_DefaultConfig tests default configuration
func TestClaudeCLIProvider_DefaultConfig(t *testing.T) {
	config := DefaultClaudeCLIConfig()

	// Model is empty by default - will be discovered dynamically
	assert.Equal(t, "", config.Model)
	assert.Equal(t, 120*time.Second, config.Timeout)
	assert.Equal(t, 4096, config.MaxOutputTokens)
}

// TestClaudeCLIProvider_NewProvider tests provider creation
func TestClaudeCLIProvider_NewProvider(t *testing.T) {
	config := ClaudeCLIConfig{
		Model:           "claude-opus-4-5-20251101",
		Timeout:         60 * time.Second,
		MaxOutputTokens: 2048,
	}

	provider := NewClaudeCLIProvider(config)

	assert.NotNil(t, provider)
	assert.Equal(t, "claude-opus-4-5-20251101", provider.model)
	assert.Equal(t, 60*time.Second, provider.timeout)
	assert.Equal(t, 2048, provider.maxOutputTokens)
}

// TestClaudeCLIProvider_NewProviderWithModel tests model-specific creation
func TestClaudeCLIProvider_NewProviderWithModel(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-haiku-4-5-20251001")

	assert.NotNil(t, provider)
	assert.Equal(t, "claude-haiku-4-5-20251001", provider.model)
}

// TestClaudeCLIProvider_GetName tests provider name
func TestClaudeCLIProvider_GetName(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")
	assert.Equal(t, "claude-cli", provider.GetName())
}

// TestClaudeCLIProvider_GetProviderType tests provider type
func TestClaudeCLIProvider_GetProviderType(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")
	assert.Equal(t, "claude", provider.GetProviderType())
}

// TestClaudeCLIProvider_GetCapabilities tests capabilities
func TestClaudeCLIProvider_GetCapabilities(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")
	caps := provider.GetCapabilities()

	assert.NotNil(t, caps)
	assert.True(t, caps.SupportsStreaming)
	assert.False(t, caps.SupportsTools) // CLI doesn't support tools
	assert.GreaterOrEqual(t, len(caps.SupportedModels), 5, "Should support multiple Claude models")

	// Check for specific models
	assert.Contains(t, caps.SupportedModels, "claude-opus-4-5-20251101")
	assert.Contains(t, caps.SupportedModels, "claude-sonnet-4-5-20250929")
	assert.Contains(t, caps.SupportedModels, "claude-sonnet-4-20250514")
}

// TestClaudeCLIProvider_SetModel tests model setting
func TestClaudeCLIProvider_SetModel(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")
	assert.Equal(t, "claude-sonnet-4-20250514", provider.GetCurrentModel())

	provider.SetModel("claude-opus-4-5-20251101")
	assert.Equal(t, "claude-opus-4-5-20251101", provider.GetCurrentModel())
}

// TestIsClaudeCodeInstalled tests CLI installation check
func TestIsClaudeCodeInstalled(t *testing.T) {
	// This test is informational - actual result depends on system
	installed := IsClaudeCodeInstalled()
	t.Logf("Claude Code installed: %v", installed)

	// Verify the function doesn't panic
	assert.NotPanics(t, func() {
		IsClaudeCodeInstalled()
	})
}

// TestGetClaudeCodePath tests path lookup
func TestGetClaudeCodePath(t *testing.T) {
	path, err := GetClaudeCodePath()

	// Just verify it returns proper types
	if err != nil {
		t.Logf("Claude Code not found: %v", err)
		assert.Empty(t, path)
	} else {
		t.Logf("Claude Code path: %s", path)
		assert.NotEmpty(t, path)
	}
}

// TestIsClaudeCodeAuthenticated tests auth check
func TestIsClaudeCodeAuthenticated(t *testing.T) {
	// This test is informational - actual result depends on system
	authenticated := IsClaudeCodeAuthenticated()
	t.Logf("Claude Code authenticated: %v", authenticated)

	// If CLI is not installed, should return false
	if !IsClaudeCodeInstalled() {
		assert.False(t, authenticated)
	}
}

// TestClaudeCLIProvider_IsCLIAvailable tests availability check
func TestClaudeCLIProvider_IsCLIAvailable(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")

	available := provider.IsCLIAvailable()
	t.Logf("CLI available: %v", available)

	if !available {
		err := provider.GetCLIError()
		t.Logf("CLI error: %v", err)
		assert.NotNil(t, err)
	}

	// Calling multiple times should return same result (sync.Once)
	available2 := provider.IsCLIAvailable()
	assert.Equal(t, available, available2)
}

// TestClaudeCLIProvider_ValidateConfig tests config validation
func TestClaudeCLIProvider_ValidateConfig(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")

	valid, errs := provider.ValidateConfig(nil)
	if IsClaudeCodeInstalled() && IsClaudeCodeAuthenticated() {
		assert.True(t, valid)
		assert.Empty(t, errs)
	} else {
		assert.False(t, valid)
		assert.NotEmpty(t, errs)
	}
}

// TestClaudeCLIProvider_Complete_NoPrompt tests error on empty prompt
func TestClaudeCLIProvider_Complete_NoPrompt(t *testing.T) {
	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")

	// Skip if CLI not available
	if !provider.IsCLIAvailable() {
		t.Skip("Claude CLI not available")
	}

	ctx := context.Background()
	resp, err := provider.Complete(ctx, &models.LLMRequest{
		Prompt:   "",
		Messages: nil,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no prompt")
}

// TestClaudeCLIProvider_Complete_CLIUnavailable tests behavior when CLI unavailable
func TestClaudeCLIProvider_Complete_CLIUnavailable(t *testing.T) {
	// Create provider with invalid path
	provider := &ClaudeCLIProvider{
		model:        "claude-sonnet-4-20250514",
		cliAvailable: false,
		cliCheckErr:  exec.ErrNotFound,
	}

	ctx := context.Background()
	resp, err := provider.Complete(ctx, &models.LLMRequest{
		Prompt: "Hello",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not available")
}

// TestClaudeCLIProvider_HealthCheck_CLIUnavailable tests health check when CLI unavailable
func TestClaudeCLIProvider_HealthCheck_CLIUnavailable(t *testing.T) {
	provider := &ClaudeCLIProvider{
		model:        "claude-sonnet-4-20250514",
		cliAvailable: false,
		cliCheckErr:  exec.ErrNotFound,
	}

	err := provider.HealthCheck()

	assert.Error(t, err)
}

// Integration test - only runs if Claude CLI is installed and authenticated
func TestClaudeCLIProvider_Integration_Complete(t *testing.T) {
	if !IsClaudeCodeInstalled() {
		t.Skip("Claude Code CLI not installed")
	}
	if !IsClaudeCodeAuthenticated() {
		t.Skip("Claude Code CLI not authenticated")
	}

	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := provider.Complete(ctx, &models.LLMRequest{
		Prompt: "Reply with exactly one word: hello",
		ModelParams: models.ModelParameters{
			MaxTokens: 10,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
	assert.Equal(t, "claude-cli", resp.ProviderName)
	t.Logf("Response: %s", resp.Content)
}

// Integration test for health check
func TestClaudeCLIProvider_Integration_HealthCheck(t *testing.T) {
	if !IsClaudeCodeInstalled() {
		t.Skip("Claude Code CLI not installed")
	}
	if !IsClaudeCodeAuthenticated() {
		t.Skip("Claude Code CLI not authenticated")
	}

	provider := NewClaudeCLIProviderWithModel("claude-sonnet-4-20250514")

	err := provider.HealthCheck()

	assert.NoError(t, err)
}

// TestCanUseClaudeOAuth tests the OAuth check function
func TestCanUseClaudeOAuth(t *testing.T) {
	// This test is informational
	canUse := CanUseClaudeOAuth()
	t.Logf("Can use Claude OAuth: %v", canUse)

	// The function should not panic
	assert.NotPanics(t, func() {
		CanUseClaudeOAuth()
	})
}
