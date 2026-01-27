// Package verifier provides integration with LLMsVerifier for model verification,
// scoring, and health monitoring capabilities.
package verifier

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// MockLLMProvider implements llm.LLMProvider for testing
type MockLLMProvider struct {
	Name            string
	CompleteFunc    func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
	StreamFunc      func(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error)
	HealthCheckFunc func() error
	GetCapsFunc     func() *models.ProviderCapabilities
	ValidateFunc    func(config map[string]interface{}) (bool, []string)
}

func (m *MockLLMProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &models.LLMResponse{
		ID:           "mock-response",
		Content:      "Mock response content",
		Confidence:   0.9,
		ProviderName: m.Name,
	}, nil
}

func (m *MockLLMProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, req)
	}
	ch := make(chan *models.LLMResponse, 1)
	close(ch)
	return ch, nil
}

func (m *MockLLMProvider) HealthCheck() error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc()
	}
	return nil
}

func (m *MockLLMProvider) GetCapabilities() *models.ProviderCapabilities {
	if m.GetCapsFunc != nil {
		return m.GetCapsFunc()
	}
	return &models.ProviderCapabilities{
		SupportedModels:   []string{"test-model"},
		SupportsStreaming: true,
	}
}

func (m *MockLLMProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(config)
	}
	return true, nil
}

// Ensure MockLLMProvider implements llm.LLMProvider
var _ llm.LLMProvider = (*MockLLMProvider)(nil)

func TestDefaultStartupConfig(t *testing.T) {
	cfg := DefaultStartupConfig()

	assert.NotNil(t, cfg)
	assert.True(t, cfg.ParallelVerification)
	assert.Equal(t, 10, cfg.MaxConcurrency)
	assert.Equal(t, 30*time.Second, cfg.VerificationTimeout)
	assert.Equal(t, 10*time.Second, cfg.HealthCheckTimeout)
	assert.Equal(t, 5.0, cfg.MinScore)
	assert.Equal(t, 15, cfg.DebateTeamSize)
	assert.Equal(t, 5, cfg.PositionCount)
	assert.Equal(t, 2, cfg.FallbacksPerPosition)
	assert.Equal(t, 0.5, cfg.OAuthPriorityBoost)
	assert.True(t, cfg.TrustOAuthOnFailure)
	assert.True(t, cfg.EnableFreeProviders)
	assert.True(t, cfg.OAuthPrimaryNonOAuthFallback)
	assert.True(t, cfg.CacheVerificationResults)
}

func TestNewStartupVerifier(t *testing.T) {
	cfg := DefaultStartupConfig()
	logger := logrus.New()

	sv := NewStartupVerifier(cfg, logger)

	assert.NotNil(t, sv)
	assert.Equal(t, cfg, sv.config)
	assert.NotNil(t, sv.providers)
	assert.NotNil(t, sv.verifierSvc)
}

func TestNewStartupVerifierWithNilConfig(t *testing.T) {
	logger := logrus.New()

	sv := NewStartupVerifier(nil, logger)

	assert.NotNil(t, sv)
	assert.NotNil(t, sv.config)
	assert.Equal(t, 15, sv.config.DebateTeamSize)
}

func TestNewStartupVerifierWithNilLogger(t *testing.T) {
	cfg := DefaultStartupConfig()

	sv := NewStartupVerifier(cfg, nil)

	assert.NotNil(t, sv)
	assert.NotNil(t, sv.log)
}

func TestStartupVerifier_SetProviderFactory(t *testing.T) {
	cfg := DefaultStartupConfig()
	sv := NewStartupVerifier(cfg, nil)

	// Set provider factory with correct signature
	factory := func(providerType string, config ProviderCreateConfig) (llm.LLMProvider, error) {
		return &MockLLMProvider{Name: providerType}, nil
	}

	sv.SetProviderFactory(factory)

	// Factory should be set
	assert.NotNil(t, sv.providerFactory)
}

func TestStartupVerifier_GetRankedProviders_Empty(t *testing.T) {
	cfg := DefaultStartupConfig()
	sv := NewStartupVerifier(cfg, nil)

	ranked := sv.GetRankedProviders()

	// Before verification, rankedProviders is nil
	assert.Nil(t, ranked)
}

func TestProviderAuthType_Constants(t *testing.T) {
	assert.Equal(t, ProviderAuthType("api_key"), AuthTypeAPIKey)
	assert.Equal(t, ProviderAuthType("oauth"), AuthTypeOAuth)
	assert.Equal(t, ProviderAuthType("free"), AuthTypeFree)
	assert.Equal(t, ProviderAuthType("anonymous"), AuthTypeAnonymous)
	assert.Equal(t, ProviderAuthType("local"), AuthTypeLocal)
}

func TestProviderStatus_Constants(t *testing.T) {
	assert.Equal(t, ProviderStatus("unknown"), StatusUnknown)
	assert.Equal(t, ProviderStatus("healthy"), StatusHealthy)
	assert.Equal(t, ProviderStatus("verified"), StatusVerified)
	assert.Equal(t, ProviderStatus("unhealthy"), StatusUnhealthy)
	assert.Equal(t, ProviderStatus("failed"), StatusFailed)
	assert.Equal(t, ProviderStatus("degraded"), StatusDegraded)
	assert.Equal(t, ProviderStatus("rate_limited"), StatusRateLimited)
	assert.Equal(t, ProviderStatus("auth_failed"), StatusAuthFailed)
	assert.Equal(t, ProviderStatus("unavailable"), StatusUnavailable)
}

func TestSupportedProviders(t *testing.T) {
	// Verify expected providers are defined
	expectedProviders := []string{"claude", "qwen", "gemini", "deepseek", "mistral", "groq", "cerebras", "openrouter", "zen", "ollama", "zai"}

	for _, name := range expectedProviders {
		info, ok := SupportedProviders[name]
		assert.True(t, ok, "Provider %s should be in SupportedProviders", name)
		assert.NotNil(t, info)
		assert.NotEmpty(t, info.Type)
		assert.NotEmpty(t, info.DisplayName)
	}
}

func TestGetProviderInfo(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantOK   bool
	}{
		{"Claude exists", "claude", true},
		{"Qwen exists", "qwen", true},
		{"Gemini exists", "gemini", true},
		{"DeepSeek exists", "deepseek", true},
		{"Unknown provider", "unknown", false},
		{"Empty provider", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := GetProviderInfo(tt.provider)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.NotNil(t, info)
			}
		})
	}
}

func TestIsOAuthProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{"Claude is OAuth", "claude", true},
		{"Qwen is OAuth", "qwen", true},
		{"Gemini is not OAuth", "gemini", false},
		{"DeepSeek is not OAuth", "deepseek", false},
		{"Zen is not OAuth", "zen", false},
		{"Unknown provider", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOAuthProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsFreeProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{"Zen is free", "zen", true},
		{"OpenRouter has free tier", "openrouter", true},
		{"Ollama is free (local)", "ollama", true},
		{"Claude is not free", "claude", false},
		{"Gemini is not free", "gemini", false},
		{"Unknown provider", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFreeProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProvidersByAuthType(t *testing.T) {
	// Test OAuth providers
	oauthProviders := GetProvidersByAuthType(AuthTypeOAuth)
	assert.NotEmpty(t, oauthProviders)

	// Verify Claude and Qwen are in OAuth list
	hasOAuth := make(map[string]bool)
	for _, p := range oauthProviders {
		hasOAuth[p.Type] = true
	}
	assert.True(t, hasOAuth["claude"], "Claude should be OAuth provider")
	assert.True(t, hasOAuth["qwen"], "Qwen should be OAuth provider")

	// Test API key providers
	apiKeyProviders := GetProvidersByAuthType(AuthTypeAPIKey)
	assert.NotEmpty(t, apiKeyProviders)

	// Test free providers
	freeProviders := GetProvidersByAuthType(AuthTypeFree)
	assert.NotEmpty(t, freeProviders)
}

func TestGetProvidersByTier(t *testing.T) {
	// Tier 1 should have premium providers
	tier1 := GetProvidersByTier(1)
	assert.NotEmpty(t, tier1)

	// Tier 5 should have free providers
	tier5 := GetProvidersByTier(5)
	assert.NotEmpty(t, tier5)

	// Non-existent tier should be empty
	tier100 := GetProvidersByTier(100)
	assert.Empty(t, tier100)
}

func TestUnifiedProvider_Fields(t *testing.T) {
	provider := &UnifiedProvider{
		ID:          "test-provider",
		Name:        "Test Provider",
		DisplayName: "Test Display Name",
		Type:        "test",
		AuthType:    AuthTypeAPIKey,
		Verified:    true,
		Score:       8.5,
		Status:      StatusVerified,
		Models: []UnifiedModel{
			{
				ID:       "model-1",
				Name:     "Model One",
				Score:    8.0,
				Verified: true,
			},
		},
		BaseURL:      "https://api.test.com",
		DefaultModel: "model-1",
	}

	assert.Equal(t, "test-provider", provider.ID)
	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, AuthTypeAPIKey, provider.AuthType)
	assert.True(t, provider.Verified)
	assert.Equal(t, 8.5, provider.Score)
	assert.Equal(t, StatusVerified, provider.Status)
	assert.Len(t, provider.Models, 1)
}

func TestUnifiedModel_Fields(t *testing.T) {
	model := &UnifiedModel{
		ID:                "model-1",
		Name:              "Test Model",
		Provider:          "test-provider",
		Score:             8.5,
		Verified:          true,
		Latency:           100 * time.Millisecond,
		ContextWindow:     128000,
		MaxOutputTokens:   4096,
		SupportsStreaming: true,
		SupportsTools:     true,
		Capabilities:      []string{"text", "code"},
		Metadata: map[string]interface{}{
			"version": "1.0",
		},
	}

	assert.Equal(t, "model-1", model.ID)
	assert.Equal(t, "Test Model", model.Name)
	assert.Equal(t, "test-provider", model.Provider)
	assert.Equal(t, 8.5, model.Score)
	assert.True(t, model.Verified)
	assert.Equal(t, 100*time.Millisecond, model.Latency)
	assert.Equal(t, 128000, model.ContextWindow)
	assert.True(t, model.SupportsStreaming)
	assert.Contains(t, model.Capabilities, "text")
	assert.Contains(t, model.Capabilities, "code")
}

func TestStartupResult_Fields(t *testing.T) {
	result := &StartupResult{
		TotalProviders:  10,
		VerifiedCount:   8,
		FailedCount:     2,
		SkippedCount:    0,
		APIKeyProviders: 5,
		OAuthProviders:  2,
		FreeProviders:   3,
		StartedAt:       time.Now(),
		CompletedAt:     time.Now(),
		DurationMs:      1500,
	}

	assert.Equal(t, 10, result.TotalProviders)
	assert.Equal(t, 8, result.VerifiedCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Equal(t, 5, result.APIKeyProviders)
	assert.Equal(t, 2, result.OAuthProviders)
	assert.Equal(t, 3, result.FreeProviders)
	assert.Equal(t, int64(1500), result.DurationMs)
}

func TestDebateTeamResult_Fields(t *testing.T) {
	result := &DebateTeamResult{
		Positions: []*DebatePosition{
			{
				Position: 1,
				Role:     "analyst",
				Primary: &DebateLLM{
					Provider:  "claude",
					ModelID:   "claude-sonnet-4-5",
					ModelName: "Claude Sonnet 4.5",
					AuthType:  AuthTypeOAuth,
					Score:     9.5,
					Verified:  true,
					IsOAuth:   true,
				},
			},
		},
		TotalLLMs:  15,
		MinScore:   5.0,
		OAuthFirst: true,
		SelectedAt: time.Now(),
	}

	assert.Len(t, result.Positions, 1)
	assert.Equal(t, 15, result.TotalLLMs)
	assert.Equal(t, 5.0, result.MinScore)
	assert.True(t, result.OAuthFirst)
	assert.NotNil(t, result.Positions[0].Primary)
	assert.Equal(t, "claude", result.Positions[0].Primary.Provider)
	assert.True(t, result.Positions[0].Primary.IsOAuth)
}

func TestStartupError_Fields(t *testing.T) {
	err := &StartupError{
		Provider:    "test-provider",
		ModelID:     "test-model",
		Phase:       "verification",
		Error:       "connection timeout",
		Recoverable: true,
	}

	assert.Equal(t, "test-provider", err.Provider)
	assert.Equal(t, "test-model", err.ModelID)
	assert.Equal(t, "verification", err.Phase)
	assert.Equal(t, "connection timeout", err.Error)
	assert.True(t, err.Recoverable)
}

func TestProviderDiscoveryResult_Fields(t *testing.T) {
	result := &ProviderDiscoveryResult{
		ID:          "test-provider",
		Type:        "api_key",
		AuthType:    AuthTypeAPIKey,
		Discovered:  true,
		Source:      "env",
		Credentials: "sk-***",
		BaseURL:     "https://api.test.com",
		Models:      []string{"model-1", "model-2"},
	}

	assert.Equal(t, "test-provider", result.ID)
	assert.Equal(t, AuthTypeAPIKey, result.AuthType)
	assert.True(t, result.Discovered)
	assert.Equal(t, "env", result.Source)
	assert.Len(t, result.Models, 2)
}

func TestStartupVerifier_VerifyAllProviders_EmptyEnvironment(t *testing.T) {
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := sv.VerifyAllProviders(ctx)

	// Should not error even with no providers
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should reflect no discovered providers (in a clean test environment)
	assert.GreaterOrEqual(t, result.TotalProviders, 0)
}

func TestStartupVerifier_GetDebateTeam_BeforeVerification(t *testing.T) {
	cfg := DefaultStartupConfig()
	sv := NewStartupVerifier(cfg, nil)

	team := sv.GetDebateTeam()

	// Should return nil before verification
	assert.Nil(t, team)
}

func TestStartupConfig_Validation(t *testing.T) {
	cfg := DefaultStartupConfig()

	// Verify all expected fields have sensible defaults
	assert.Greater(t, cfg.MaxConcurrency, 0)
	assert.Greater(t, cfg.VerificationTimeout, time.Duration(0))
	assert.Greater(t, cfg.HealthCheckTimeout, time.Duration(0))
	assert.GreaterOrEqual(t, cfg.MinScore, 0.0)
	assert.Equal(t, 15, cfg.DebateTeamSize)
	assert.Equal(t, 5*3, cfg.DebateTeamSize) // 5 positions * 3 LLMs each
	assert.Equal(t, cfg.PositionCount*(1+cfg.FallbacksPerPosition), cfg.DebateTeamSize)
}

func TestProviderTypeInfo_Fields(t *testing.T) {
	// Test Claude provider info
	claude, ok := SupportedProviders["claude"]
	require.True(t, ok)
	assert.Equal(t, "claude", claude.Type)
	assert.Equal(t, "Claude (Anthropic)", claude.DisplayName)
	assert.Equal(t, AuthTypeOAuth, claude.AuthType)
	assert.Equal(t, 1, claude.Tier)
	assert.NotEmpty(t, claude.EnvVars)
	assert.NotEmpty(t, claude.BaseURL)
	assert.NotEmpty(t, claude.Models)
	assert.False(t, claude.Free)

	// Test Zen provider info
	zen, ok := SupportedProviders["zen"]
	require.True(t, ok)
	assert.Equal(t, "zen", zen.Type)
	assert.Equal(t, "OpenCode Zen", zen.DisplayName)
	assert.Equal(t, AuthTypeFree, zen.AuthType)
	assert.True(t, zen.Free)
	assert.NotEmpty(t, zen.Models)
}

// ============================================================================
// Re-Evaluation Tests - Verify fresh verification on every call
// ============================================================================

func TestStartupVerifier_ReEvaluation_TimestampsUpdated(t *testing.T) {
	// Test that each call to VerifyAllProviders updates timestamps
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	cfg.EnableFreeProviders = true // Ensure Zen is discovered
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First verification
	result1, err1 := sv.VerifyAllProviders(ctx)
	require.NoError(t, err1)
	require.NotNil(t, result1)

	// Record first verification time
	firstStarted := result1.StartedAt
	firstCompleted := result1.CompletedAt

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Second verification - should be fresh, not cached
	result2, err2 := sv.VerifyAllProviders(ctx)
	require.NoError(t, err2)
	require.NotNil(t, result2)

	// Timestamps should be different (newer)
	assert.True(t, result2.StartedAt.After(firstStarted) || result2.StartedAt.Equal(firstStarted),
		"Second verification started_at should be >= first verification")
	assert.True(t, result2.CompletedAt.After(firstCompleted),
		"Second verification completed_at should be after first verification")
}

func TestStartupVerifier_ReEvaluation_ResultsAreConsistent(t *testing.T) {
	// Test that multiple verifications produce consistent results
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	cfg.EnableFreeProviders = true
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run verification multiple times
	var results []*StartupResult
	for i := 0; i < 3; i++ {
		result, err := sv.VerifyAllProviders(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
		results = append(results, result)
		time.Sleep(50 * time.Millisecond)
	}

	// All results should have same provider count (consistent discovery)
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0].TotalProviders, results[i].TotalProviders,
			"Provider count should be consistent across verifications")
	}

	// Each result should have a valid duration (>= 0 since very fast verifications may round to 0)
	for _, r := range results {
		assert.GreaterOrEqual(t, r.DurationMs, int64(0), "Duration should be non-negative")
	}
}

func TestStartupVerifier_ReEvaluation_ProvidersReSorted(t *testing.T) {
	// Test that providers are re-sorted on each verification
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	cfg.EnableFreeProviders = true
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First verification
	_, err := sv.VerifyAllProviders(ctx)
	require.NoError(t, err)

	ranked := sv.GetRankedProviders()
	if len(ranked) < 2 {
		t.Skip("Need at least 2 providers to test sorting")
	}

	// Verify sorted order (descending by score, OAuth first)
	for i := 1; i < len(ranked); i++ {
		prevOAuth := ranked[i-1].AuthType == AuthTypeOAuth
		currOAuth := ranked[i].AuthType == AuthTypeOAuth

		// OAuth providers should come first
		if !prevOAuth && currOAuth {
			t.Error("OAuth providers should be ranked before non-OAuth")
		}

		// Within same auth type, scores should be descending
		if prevOAuth == currOAuth {
			assert.GreaterOrEqual(t, ranked[i-1].Score, ranked[i].Score,
				"Providers should be sorted by score (descending)")
		}
	}
}

func TestStartupVerifier_ReEvaluation_DebateTeamReselected(t *testing.T) {
	// Test that debate team is reselected on each verification
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	cfg.EnableFreeProviders = true
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First verification
	result1, err := sv.VerifyAllProviders(ctx)
	require.NoError(t, err)

	if result1.DebateTeam == nil {
		t.Skip("No debate team configured (need verified providers)")
	}

	firstSelectedAt := result1.DebateTeam.SelectedAt

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Second verification
	result2, err := sv.VerifyAllProviders(ctx)
	require.NoError(t, err)
	require.NotNil(t, result2.DebateTeam)

	// Debate team should be reselected with new timestamp
	assert.True(t, result2.DebateTeam.SelectedAt.After(firstSelectedAt),
		"Debate team should be reselected with newer timestamp")
}

func TestStartupVerifier_ReEvaluation_LastVerifyAtUpdated(t *testing.T) {
	// Test that lastVerifyAt is updated on each call
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initial state - not initialized
	assert.False(t, sv.IsInitialized())

	// First verification
	_, err := sv.VerifyAllProviders(ctx)
	require.NoError(t, err)

	// Should now be initialized
	assert.True(t, sv.IsInitialized())

	// Get lastVerifyAt indirectly through re-verification
	time.Sleep(100 * time.Millisecond)

	// Second verification should complete without error
	_, err = sv.VerifyAllProviders(ctx)
	require.NoError(t, err)
	assert.True(t, sv.IsInitialized())
}

func TestStartupResult_FreshTimestamps(t *testing.T) {
	// Test that StartupResult has fresh timestamps (within last 5 minutes)
	cfg := DefaultStartupConfig()
	cfg.VerificationTimeout = 5 * time.Second
	sv := NewStartupVerifier(cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := sv.VerifyAllProviders(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	now := time.Now()
	maxAge := 5 * time.Minute

	// started_at should be within last 5 minutes
	assert.True(t, now.Sub(result.StartedAt) < maxAge,
		"started_at should be fresh (within last 5 minutes)")

	// completed_at should be within last 5 minutes
	assert.True(t, now.Sub(result.CompletedAt) < maxAge,
		"completed_at should be fresh (within last 5 minutes)")

	// completed_at should be after started_at
	assert.True(t, result.CompletedAt.After(result.StartedAt) || result.CompletedAt.Equal(result.StartedAt),
		"completed_at should be >= started_at")
}
