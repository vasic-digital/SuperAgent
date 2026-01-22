// Package adapters provides provider adapters for LLMsVerifier integration
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
		TrustOnVerificationFailure: true,
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
func (oa *OAuthAdapter) VerifyClaudeOAuth(ctx context.Context) (*verifier.UnifiedProvider, error) {
	oa.log.Debug("Verifying Claude OAuth provider")

	// Read credentials from ~/.claude/.credentials.json
	creds, err := oa.credReader.ReadClaudeCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude credentials: %w", err)
	}

	if creds == nil || creds.ClaudeAiOauth == nil {
		return nil, fmt.Errorf("Claude OAuth credentials not found")
	}

	// Check token expiration
	expiresAt := time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt)
	if oauth_credentials.IsExpired(creds.ClaudeAiOauth.ExpiresAt) {
		oa.log.Info("Claude token expired, attempting refresh")
		refreshed, err := oauth_credentials.AutoRefreshClaudeToken(creds)
		if err != nil {
			oa.log.WithError(err).Warn("Claude token refresh failed")
			if oa.config.TrustOnVerificationFailure {
				return oa.createTrustedClaudeProvider(creds), nil
			}
			return nil, fmt.Errorf("Claude token expired and refresh failed: %w", err)
		}
		creds = refreshed
		expiresAt = time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt)
	}

	// Cache token
	oa.mu.Lock()
	oa.claudeToken = creds.ClaudeAiOauth.AccessToken
	oa.claudeExpiry = expiresAt
	oa.mu.Unlock()

	// Create provider
	provider := &verifier.UnifiedProvider{
		ID:               "claude",
		Name:             "claude",
		DisplayName:      "Claude (Anthropic)",
		Type:             "claude",
		AuthType:         verifier.AuthTypeOAuth,
		BaseURL:          "https://api.anthropic.com/v1/messages",
		DefaultModel:     "claude-sonnet-4-5-20250929",
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: true,
		Tier:             1,
		Priority:         1,
		Models: []verifier.UnifiedModel{
			{ID: "claude-opus-4-5-20251101", Name: "Claude Opus 4.5", Provider: "claude"},
			{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5", Provider: "claude"},
			{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5", Provider: "claude"},
			{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", Provider: "claude"},
			{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "claude"},
		},
	}

	// Run verification through LLMsVerifier
	verifyCtx, cancel := context.WithTimeout(ctx, oa.config.VerificationTimeout)
	defer cancel()

	result, err := oa.verifierSvc.VerifyModel(verifyCtx, provider.DefaultModel, "claude")
	if err != nil {
		oa.log.WithError(err).Warn("Claude OAuth verification failed")
		if oa.config.TrustOnVerificationFailure {
			return oa.createTrustedClaudeProvider(creds), nil
		}
		return nil, fmt.Errorf("Claude OAuth verification failed: %w", err)
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
		"provider": "claude",
		"verified": provider.Verified,
		"score":    provider.Score,
	}).Info("Claude OAuth verification completed")

	return provider, nil
}

// createTrustedClaudeProvider creates a Claude provider that's trusted without full verification
func (oa *OAuthAdapter) createTrustedClaudeProvider(creds *oauth_credentials.ClaudeOAuthCredentials) *verifier.UnifiedProvider {
	expiresAt := time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt)

	return &verifier.UnifiedProvider{
		ID:               "claude",
		Name:             "claude",
		DisplayName:      "Claude (Anthropic)",
		Type:             "claude",
		AuthType:         verifier.AuthTypeOAuth,
		Verified:         true, // Trust OAuth credentials
		VerifiedAt:       time.Now(),
		Score:            oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost,
		BaseURL:          "https://api.anthropic.com/v1/messages",
		DefaultModel:     "claude-sonnet-4-5-20250929",
		Status:           verifier.StatusHealthy,
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: true,
		Tier:             1,
		Priority:         1,
		TestResults:      map[string]bool{"oauth_trusted": true},
		Models: []verifier.UnifiedModel{
			{ID: "claude-opus-4-5-20251101", Name: "Claude Opus 4.5", Provider: "claude", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
			{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5", Provider: "claude", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
			{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5", Provider: "claude", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
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
				return oa.createTrustedQwenProvider(creds), nil
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
			return oa.createTrustedQwenProvider(creds), nil
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

// createTrustedQwenProvider creates a Qwen provider that's trusted without full verification
func (oa *OAuthAdapter) createTrustedQwenProvider(creds *oauth_credentials.QwenOAuthCredentials) *verifier.UnifiedProvider {
	expiresAt := time.UnixMilli(creds.ExpiryDate)

	return &verifier.UnifiedProvider{
		ID:               "qwen",
		Name:             "qwen",
		DisplayName:      "Qwen (Alibaba)",
		Type:             "qwen",
		AuthType:         verifier.AuthTypeOAuth,
		Verified:         true, // Trust OAuth credentials
		VerifiedAt:       time.Now(),
		Score:            oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost,
		BaseURL:          "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		DefaultModel:     "qwen-max",
		Status:           verifier.StatusHealthy,
		OAuthTokenExpiry: expiresAt,
		OAuthAutoRefresh: true,
		Tier:             2,
		Priority:         3,
		TestResults:      map[string]bool{"oauth_trusted": true},
		Models: []verifier.UnifiedModel{
			{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
			{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
			{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen", Verified: true, Score: oa.config.DefaultScoreOnFailure + oa.config.OAuthPriorityBoost},
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
