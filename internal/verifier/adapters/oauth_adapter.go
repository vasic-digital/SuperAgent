// Package adapters provides provider adapters for LLMsVerifier integration
//
// # OAuth Token Limitations
//
// IMPORTANT: OAuth tokens from CLI tools have different access levels:
//
// ## Claude OAuth Tokens (from Claude Code CLI)
//
// OAuth tokens from Claude Code CLI (~/.claude/.credentials.json) are PRODUCT-RESTRICTED.
// They can ONLY be used with Claude Code itself - NOT with the standard Anthropic API.
// Attempting to use these tokens with api.anthropic.com will return:
//
//	"This credential is only authorized for use with Claude Code and cannot be used for other API requests."
//
// For general API access, use an API key from https://console.anthropic.com/
//
// ## Qwen OAuth Tokens (from Qwen Code CLI)
//
// OAuth tokens from Qwen Code CLI (~/.qwen/oauth_creds.json) are for the Qwen Portal only.
// They cannot be used with the DashScope API.
//
// For general API access, use a DashScope API key from https://dashscope.aliyuncs.com/
//
// ## Verification Strategy
//
// Given these limitations, OAuth providers use "trust mode" verification:
//   - Token presence and validity (expiration) is verified
//   - API access cannot be verified due to product restrictions
//   - Provider is marked as "trusted" with a default score when tokens are valid
//
// See: https://platform.claude.com/docs/en/api/overview
package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dev.helix.agent/internal/auth/oauth_credentials"
	"dev.helix.agent/internal/verifier"
	"github.com/sirupsen/logrus"
)

// OAuthAdapter handles verification for OAuth-based providers (Claude, Qwen)
type OAuthAdapter struct {
	credReader  *oauth_credentials.OAuthCredentialReader
	verifierSvc *verifier.VerificationService
	config      *OAuthAdapterConfig

	// Cached credentials
	claudeToken  string
	claudeExpiry time.Time
	qwenToken    string
	qwenExpiry   time.Time

	mu  sync.RWMutex
	log *logrus.Logger
}

// OAuthAdapterConfig contains configuration for OAuth adapter
type OAuthAdapterConfig struct {
	// Refresh thresholds
	RefreshThresholdMins int `yaml:"refresh_threshold_mins" json:"refresh_threshold_mins"`

	// Trust settings
	TrustOnVerificationFailure bool    `yaml:"trust_on_verification_failure" json:"trust_on_verification_failure"`
	DefaultScoreOnFailure      float64 `yaml:"default_score_on_failure" json:"default_score_on_failure"`

	// OAuth priority boost
	OAuthPriorityBoost float64 `yaml:"oauth_priority_boost" json:"oauth_priority_boost"`

	// Timeouts
	VerificationTimeout time.Duration `yaml:"verification_timeout" json:"verification_timeout"`
	RefreshTimeout      time.Duration `yaml:"refresh_timeout" json:"refresh_timeout"`

	// CLI refresh settings (for Qwen)
	EnableCLIRefresh  bool `yaml:"enable_cli_refresh" json:"enable_cli_refresh"`
	CLIRefreshRetries int  `yaml:"cli_refresh_retries" json:"cli_refresh_retries"`
}

// DefaultOAuthAdapterConfig returns sensible defaults
func DefaultOAuthAdapterConfig() *OAuthAdapterConfig {
	return &OAuthAdapterConfig{
		RefreshThresholdMins:       10,
		TrustOnVerificationFailure: false, // IMPORTANT: Don't trust tokens that can't make API calls
		DefaultScoreOnFailure:      7.5,
		OAuthPriorityBoost:         0.5,
		VerificationTimeout:        30 * time.Second,
		RefreshTimeout:             60 * time.Second,
		EnableCLIRefresh:           true,
		CLIRefreshRetries:          3,
	}
}

// NewOAuthAdapter creates a new OAuth adapter
func NewOAuthAdapter(verifierSvc *verifier.VerificationService, log *logrus.Logger) *OAuthAdapter {
	if log == nil {
		log = logrus.New()
	}

	return &OAuthAdapter{
		credReader:  oauth_credentials.NewOAuthCredentialReader(),
		verifierSvc: verifierSvc,
		config:      DefaultOAuthAdapterConfig(),
		log:         log,
	}
}

// NewOAuthAdapterWithConfig creates a new OAuth adapter with custom config
func NewOAuthAdapterWithConfig(verifierSvc *verifier.VerificationService, config *OAuthAdapterConfig, log *logrus.Logger) *OAuthAdapter {
	adapter := NewOAuthAdapter(verifierSvc, log)
	if config != nil {
		adapter.config = config
	}
	return adapter
}

// VerifyClaudeOAuth verifies Claude provider using OAuth credentials
//
// IMPORTANT: Claude OAuth tokens from Claude Code CLI are PRODUCT-RESTRICTED.
// They can ONLY be used with Claude Code - NOT with the standard Anthropic API.
// Attempting to use these tokens with api.anthropic.com returns:
//
//	"This credential is only authorized for use with Claude Code and cannot be used for other API requests."
//
// Therefore, Claude OAuth providers are NOT verified and NOT included in the AI debate team.
// For actual API access, use an API key from console.anthropic.com
func (oa *OAuthAdapter) VerifyClaudeOAuth(ctx context.Context) (*verifier.UnifiedProvider, error) {
	oa.log.Debug("Checking Claude OAuth credentials (PRODUCT-RESTRICTED - cannot be used for API)")

	// Read credentials from ~/.claude/.credentials.json
	creds, err := oa.credReader.ReadClaudeCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude credentials: %w", err)
	}

	if creds == nil || creds.ClaudeAiOauth == nil {
		return nil, fmt.Errorf("Claude OAuth credentials not found")
	}

	// Check token expiration
	tokenExpired := oauth_credentials.IsExpired(creds.ClaudeAiOauth.ExpiresAt)

	// IMPORTANT: Claude OAuth tokens from Claude Code CLI are PRODUCT-RESTRICTED
	// They can ONLY be used with Claude Code itself, not the standard Anthropic API.
	// API calls would fail with: "This credential is only authorized for use with Claude Code"
	//
	// Since these tokens CANNOT be used for API calls, Claude OAuth providers:
	// - Are NOT marked as verified (Verified: false)
	// - Are NOT included in the AI debate team
	// - Are returned for informational purposes only
	//
	// For actual API access, users must use an API key from console.anthropic.com
	oa.log.WithFields(logrus.Fields{
		"provider":           "claude",
		"auth_type":          "oauth",
		"token_present":      true,
		"token_expired":      tokenExpired,
		"product_restricted": true,
		"verified":           false,
		"reason":             "OAuth tokens from Claude Code CLI cannot be used for API calls",
	}).Warn("Claude OAuth token found but CANNOT be used for API calls - provider NOT verified")

	// Return provider with Verified: false - it will be excluded from debate team
	return oa.createUnverifiedClaudeProvider(creds), nil
}

// createUnverifiedClaudeProvider creates a Claude provider that is NOT verified
// because OAuth tokens from Claude Code CLI are PRODUCT-RESTRICTED and cannot be used for API calls.
// This provider is returned for informational purposes only - it will NOT be included in the debate team.
func (oa *OAuthAdapter) createUnverifiedClaudeProvider(creds *oauth_credentials.ClaudeOAuthCredentials) *verifier.UnifiedProvider {
	expiresAt := time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt)

	return &verifier.UnifiedProvider{
		ID:               "claude",
		Name:             "claude",
		DisplayName:      "Claude (Anthropic) - OAuth RESTRICTED",
		Type:             "claude",
		AuthType:         verifier.AuthTypeOAuth,
		Verified:         false, // NOT verified - OAuth tokens are product-restricted
		VerifiedAt:       time.Time{},
		Score:            0, // No score - not functional
		BaseURL:          "https://api.anthropic.com/v1/messages",
		DefaultModel:     "claude-3-5-sonnet-20241022",
		Status:           verifier.StatusUnhealthy,
		ErrorMessage:     "OAuth tokens from Claude Code CLI are product-restricted and cannot be used for API calls. Use an API key from console.anthropic.com instead.",
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: false,
		Tier:             1,
		Priority:         1,
		TestResults:      map[string]bool{"oauth_product_restricted": true, "api_functional": false},
		Models: []verifier.UnifiedModel{
			{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "claude", Verified: false, Score: 0},
			{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "claude", Verified: false, Score: 0},
			{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "claude", Verified: false, Score: 0},
		},
	}
}

// VerifyQwenOAuth verifies Qwen provider using OAuth credentials
func (oa *OAuthAdapter) VerifyQwenOAuth(ctx context.Context) (*verifier.UnifiedProvider, error) {
	oa.log.Debug("Verifying Qwen OAuth provider")

	// Read credentials from ~/.qwen/oauth_creds.json
	creds, err := oa.credReader.ReadQwenCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to read Qwen credentials: %w", err)
	}

	if creds == nil {
		return nil, fmt.Errorf("Qwen OAuth credentials not found")
	}

	// Check token expiration
	expiresAt := time.UnixMilli(creds.ExpiryDate)
	if oauth_credentials.IsExpired(creds.ExpiryDate) {
		oa.log.Info("Qwen token expired, attempting refresh")

		// Try standard OAuth refresh first
		refreshed, err := oauth_credentials.AutoRefreshQwenToken(creds)
		if err != nil {
			// Try CLI-based refresh as fallback
			if oa.config.EnableCLIRefresh {
				oa.log.Info("Standard refresh failed, trying CLI-based refresh")
				refreshed, err = oauth_credentials.RefreshQwenTokenWithFallback(ctx, creds)
			}
		}

		if err != nil {
			oa.log.WithError(err).Warn("Qwen token refresh failed")
			if oa.config.TrustOnVerificationFailure {
				return oa.createUnverifiedQwenProvider(creds), nil
			}
			return nil, fmt.Errorf("Qwen token expired and refresh failed: %w", err)
		}
		creds = refreshed
		expiresAt = time.UnixMilli(creds.ExpiryDate)
	}

	// Cache token
	oa.mu.Lock()
	oa.qwenToken = creds.AccessToken
	oa.qwenExpiry = expiresAt
	oa.mu.Unlock()

	// Create provider
	provider := &verifier.UnifiedProvider{
		ID:               "qwen",
		Name:             "qwen",
		DisplayName:      "Qwen (Alibaba)",
		Type:             "qwen",
		AuthType:         verifier.AuthTypeOAuth,
		BaseURL:          "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		DefaultModel:     "qwen-max",
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: true,
		Tier:             2,
		Priority:         3,
		Models: []verifier.UnifiedModel{
			{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen"},
			{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen"},
			{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen"},
			{ID: "qwen-coder-turbo", Name: "Qwen Coder Turbo", Provider: "qwen"},
			{ID: "qwen-long", Name: "Qwen Long", Provider: "qwen"},
		},
	}

	// Run verification through LLMsVerifier
	verifyCtx, cancel := context.WithTimeout(ctx, oa.config.VerificationTimeout)
	defer cancel()

	result, err := oa.verifierSvc.VerifyModel(verifyCtx, provider.DefaultModel, "qwen")
	if err != nil {
		oa.log.WithError(err).Warn("Qwen OAuth verification failed")
		if oa.config.TrustOnVerificationFailure {
			return oa.createUnverifiedQwenProvider(creds), nil
		}
		return nil, fmt.Errorf("Qwen OAuth verification failed: %w", err)
	}

	provider.Verified = result.Verified
	provider.VerifiedAt = time.Now()
	provider.Score = result.OverallScore/10.0 + oa.config.OAuthPriorityBoost
	provider.ScoreSuffix = result.ScoreSuffix
	provider.CodeVisible = result.CodeVisible
	provider.TestResults = result.TestsMap

	if provider.Verified {
		provider.Status = verifier.StatusHealthy
	} else {
		provider.Status = verifier.StatusUnhealthy
		provider.ErrorMessage = result.ErrorMessage
	}

	// Update model scores
	for i := range provider.Models {
		provider.Models[i].Verified = provider.Verified
		provider.Models[i].Score = provider.Score
		provider.Models[i].VerifiedAt = provider.VerifiedAt
	}

	oa.log.WithFields(logrus.Fields{
		"provider": "qwen",
		"verified": provider.Verified,
		"score":    provider.Score,
	}).Info("Qwen OAuth verification completed")

	return provider, nil
}

// createUnverifiedQwenProvider creates a Qwen provider that is NOT verified
// because OAuth tokens from Qwen CLI are for the Qwen Portal only - NOT for DashScope API.
// This provider is returned for informational purposes only - it will NOT be included in the debate team.
func (oa *OAuthAdapter) createUnverifiedQwenProvider(creds *oauth_credentials.QwenOAuthCredentials) *verifier.UnifiedProvider {
	expiresAt := time.UnixMilli(creds.ExpiryDate)

	return &verifier.UnifiedProvider{
		ID:               "qwen",
		Name:             "qwen",
		DisplayName:      "Qwen (Alibaba) - OAuth RESTRICTED",
		Type:             "qwen",
		AuthType:         verifier.AuthTypeOAuth,
		Verified:         false, // NOT verified - OAuth tokens are for Portal only
		VerifiedAt:       time.Time{},
		Score:            0, // No score - not functional
		BaseURL:          "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		DefaultModel:     "qwen-max",
		Status:           verifier.StatusUnhealthy,
		ErrorMessage:     "OAuth tokens from Qwen CLI are for the Qwen Portal only - cannot be used for DashScope API. Use a DashScope API key instead.",
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: false,
		Tier:             2,
		Priority:         3,
		TestResults:      map[string]bool{"oauth_portal_only": true, "api_functional": false},
		Models: []verifier.UnifiedModel{
			{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen", Verified: false, Score: 0},
			{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen", Verified: false, Score: 0},
			{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen", Verified: false, Score: 0},
		},
	}
}

// RefreshTokenIfNeeded checks and refreshes tokens if needed
func (oa *OAuthAdapter) RefreshTokenIfNeeded(ctx context.Context, providerType string) error {
	threshold := time.Duration(oa.config.RefreshThresholdMins) * time.Minute

	switch providerType {
	case "claude":
		oa.mu.RLock()
		expiry := oa.claudeExpiry
		oa.mu.RUnlock()

		if time.Until(expiry) < threshold {
			creds, err := oa.credReader.ReadClaudeCredentials()
			if err != nil {
				return err
			}
			_, err = oauth_credentials.AutoRefreshClaudeToken(creds)
			return err
		}

	case "qwen":
		oa.mu.RLock()
		expiry := oa.qwenExpiry
		oa.mu.RUnlock()

		if time.Until(expiry) < threshold {
			creds, err := oa.credReader.ReadQwenCredentials()
			if err != nil {
				return err
			}
			_, err = oauth_credentials.RefreshQwenTokenWithFallback(ctx, creds)
			return err
		}
	}

	return nil
}

// GetClaudeToken returns the cached Claude OAuth token
func (oa *OAuthAdapter) GetClaudeToken() (string, time.Time) {
	oa.mu.RLock()
	defer oa.mu.RUnlock()
	return oa.claudeToken, oa.claudeExpiry
}

// GetQwenToken returns the cached Qwen OAuth token
func (oa *OAuthAdapter) GetQwenToken() (string, time.Time) {
	oa.mu.RLock()
	defer oa.mu.RUnlock()
	return oa.qwenToken, oa.qwenExpiry
}

// IsClaudeTokenValid checks if the Claude token is still valid
func (oa *OAuthAdapter) IsClaudeTokenValid() bool {
	oa.mu.RLock()
	defer oa.mu.RUnlock()
	return oa.claudeToken != "" && time.Now().Before(oa.claudeExpiry)
}

// IsQwenTokenValid checks if the Qwen token is still valid
func (oa *OAuthAdapter) IsQwenTokenValid() bool {
	oa.mu.RLock()
	defer oa.mu.RUnlock()
	return oa.qwenToken != "" && time.Now().Before(oa.qwenExpiry)
}

// StartBackgroundRefresh starts a goroutine that periodically checks and refreshes tokens
func (oa *OAuthAdapter) StartBackgroundRefresh(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Check Claude token
				if err := oa.RefreshTokenIfNeeded(ctx, "claude"); err != nil {
					oa.log.WithError(err).Warn("Background Claude token refresh failed")
				}

				// Check Qwen token
				if err := oa.RefreshTokenIfNeeded(ctx, "qwen"); err != nil {
					oa.log.WithError(err).Warn("Background Qwen token refresh failed")
				}
			}
		}
	}()
}
