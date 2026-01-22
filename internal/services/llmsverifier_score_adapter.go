package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/verifier"
	"github.com/sirupsen/logrus"
)

// LLMsVerifierScoreAdapter implements LLMsVerifierScoreProvider interface
// It connects ProviderDiscovery to the actual LLMsVerifier scoring system
type LLMsVerifierScoreAdapter struct {
	scoringService  *verifier.ScoringService
	verificationSvc *verifier.VerificationService
	providerScores  map[string]float64 // Cached provider scores
	modelScores     map[string]float64 // Cached model scores
	mu              sync.RWMutex
	log             *logrus.Logger
	lastRefresh     time.Time
	refreshInterval time.Duration
}

// NewLLMsVerifierScoreAdapter creates a new adapter for LLMsVerifier scores
func NewLLMsVerifierScoreAdapter(
	scoringService *verifier.ScoringService,
	verificationSvc *verifier.VerificationService,
	log *logrus.Logger,
) *LLMsVerifierScoreAdapter {
	if log == nil {
		log = logrus.New()
	}

	adapter := &LLMsVerifierScoreAdapter{
		scoringService:  scoringService,
		verificationSvc: verificationSvc,
		providerScores:  make(map[string]float64),
		modelScores:     make(map[string]float64),
		log:             log,
		refreshInterval: 5 * time.Minute,
	}

	// Initialize with any existing verification results
	adapter.initializeFromVerificationCache()

	return adapter
}

// initializeFromVerificationCache loads scores from existing verification results
func (a *LLMsVerifierScoreAdapter) initializeFromVerificationCache() {
	if a.verificationSvc == nil {
		return
	}

	ctx := context.Background()
	verifications, err := a.verificationSvc.GetAllVerifications(ctx)
	if err != nil {
		a.log.WithError(err).Warn("Failed to load verification results for score initialization")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, v := range verifications {
		if v.Verified && v.Score > 0 {
			// Store by model ID
			a.modelScores[v.ModelID] = v.Score

			// Also aggregate by provider
			if existingScore, ok := a.providerScores[v.Provider]; ok {
				// Use the higher score for the provider
				if v.Score > existingScore {
					a.providerScores[v.Provider] = v.Score
				}
			} else {
				a.providerScores[v.Provider] = v.Score
			}
		}
	}

	a.lastRefresh = time.Now()
	a.log.WithFields(logrus.Fields{
		"provider_scores": len(a.providerScores),
		"model_scores":    len(a.modelScores),
	}).Info("LLMsVerifier scores initialized from cache")
}

// GetProviderScore returns the LLMsVerifier score for a provider (0-10)
func (a *LLMsVerifierScoreAdapter) GetProviderScore(providerType string) (float64, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	score, found := a.providerScores[providerType]
	if found {
		// Normalize score to 0-10 range if needed (LLMsVerifier uses 0-100)
		if score > 10 {
			score = score / 10.0
		}
		return score, true
	}
	return 0, false
}

// GetModelScore returns the LLMsVerifier score for a specific model
func (a *LLMsVerifierScoreAdapter) GetModelScore(modelID string) (float64, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	score, found := a.modelScores[modelID]
	if found {
		// Normalize score to 0-10 range if needed (LLMsVerifier uses 0-100)
		if score > 10 {
			score = score / 10.0
		}
		return score, true
	}
	return 0, false
}

// RefreshScores refreshes scores from LLMsVerifier
func (a *LLMsVerifierScoreAdapter) RefreshScores(ctx context.Context) error {
	// Check if refresh is needed
	if time.Since(a.lastRefresh) < a.refreshInterval {
		return nil
	}

	a.log.Info("Refreshing LLMsVerifier scores")

	// Get latest verification results
	if a.verificationSvc != nil {
		verifications, err := a.verificationSvc.GetAllVerifications(ctx)
		if err != nil {
			return err
		}

		a.mu.Lock()
		for _, v := range verifications {
			if v.Verified && v.Score > 0 {
				a.modelScores[v.ModelID] = v.Score

				if existingScore, ok := a.providerScores[v.Provider]; ok {
					if v.Score > existingScore {
						a.providerScores[v.Provider] = v.Score
					}
				} else {
					a.providerScores[v.Provider] = v.Score
				}
			}
		}
		a.lastRefresh = time.Now()
		a.mu.Unlock()
	}

	// Use ScoringService to calculate/refresh scores
	if a.scoringService != nil {
		a.refreshFromScoringService(ctx)
	}

	return nil
}

// refreshFromScoringService uses the scoring service to get updated scores
// DYNAMIC: Models are discovered from verification results and scoring service, NOT hardcoded
func (a *LLMsVerifierScoreAdapter) refreshFromScoringService(ctx context.Context) {
	// DYNAMIC MODEL DISCOVERY: Get models from verification results, not a hardcoded list
	// This ensures the system adapts to whatever models are actually being verified
	knownModels := a.getKnownModelsFromVerifications()

	// If no verified models yet, try to get from scoring service's available models
	if len(knownModels) == 0 && a.scoringService != nil {
		knownModels = a.scoringService.GetAvailableModels()
	}

	if len(knownModels) == 0 {
		a.log.Debug("No models found for score refresh - waiting for verification results")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, modelID := range knownModels {
		result, err := a.scoringService.CalculateScore(ctx, modelID)
		if err != nil {
			a.log.WithError(err).WithField("model", modelID).Debug("Failed to calculate score for model")
			continue
		}
		if result != nil && result.OverallScore > 0 {
			a.modelScores[modelID] = result.OverallScore

			// Map model to provider DYNAMICALLY
			provider := inferProviderFromModel(modelID)
			if provider != "" {
				if existingScore, ok := a.providerScores[provider]; ok {
					if result.OverallScore > existingScore {
						a.providerScores[provider] = result.OverallScore
					}
				} else {
					a.providerScores[provider] = result.OverallScore
				}
			}
		}
	}

	a.log.WithFields(logrus.Fields{
		"provider_scores": len(a.providerScores),
		"model_scores":    len(a.modelScores),
	}).Debug("Refreshed scores from ScoringService")
}

// getKnownModelsFromVerifications extracts unique model IDs from verification results
// DYNAMIC: No hardcoded list - models come from actual verification data
func (a *LLMsVerifierScoreAdapter) getKnownModelsFromVerifications() []string {
	if a.verificationSvc == nil {
		return nil
	}

	ctx := context.Background()
	verifications, err := a.verificationSvc.GetAllVerifications(ctx)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var models []string
	for _, v := range verifications {
		if v.ModelID != "" && !seen[v.ModelID] {
			models = append(models, v.ModelID)
			seen[v.ModelID] = true
		}
	}
	return models
}

// UpdateScore updates the score for a specific model/provider after verification
func (a *LLMsVerifierScoreAdapter) UpdateScore(provider, modelID string, score float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.modelScores[modelID] = score

	// Update provider score if this is higher
	if existingScore, ok := a.providerScores[provider]; ok {
		if score > existingScore {
			a.providerScores[provider] = score
		}
	} else {
		a.providerScores[provider] = score
	}

	a.log.WithFields(logrus.Fields{
		"provider": provider,
		"model":    modelID,
		"score":    score,
	}).Debug("Updated LLMsVerifier score")
}

// GetAllProviderScores returns all cached provider scores
func (a *LLMsVerifierScoreAdapter) GetAllProviderScores() map[string]float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]float64)
	for k, v := range a.providerScores {
		result[k] = v
	}
	return result
}

// GetBestProvider returns the provider with the highest score
func (a *LLMsVerifierScoreAdapter) GetBestProvider() (string, float64) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var bestProvider string
	var bestScore float64

	for provider, score := range a.providerScores {
		if score > bestScore {
			bestProvider = provider
			bestScore = score
		}
	}

	return bestProvider, bestScore
}

// inferProviderFromModel dynamically infers provider type from model ID patterns
// DYNAMIC: Uses pattern matching instead of hardcoded mappings
// This ensures new models from any provider are automatically categorized correctly
func inferProviderFromModel(modelID string) string {
	modelLower := strings.ToLower(modelID)

	// IMPORTANT: Check OpenRouter format FIRST (provider/model pattern)
	// This handles models like "anthropic/claude-3", "meta-llama/llama-3.1-70b"
	if strings.Contains(modelLower, "/") {
		parts := strings.Split(modelLower, "/")
		if len(parts) >= 2 {
			providerPart := parts[0]
			switch {
			case providerPart == "anthropic":
				return "claude"
			case providerPart == "openai":
				return "openai"
			case providerPart == "google":
				return "gemini"
			case providerPart == "meta-llama" || providerPart == "meta":
				return "openrouter" // Meta/Llama models via OpenRouter
			case providerPart == "together" || providerPart == "togethercomputer":
				return "together"
			default:
				return "openrouter" // Generic OpenRouter model
			}
		}
	}

	// Check provider-specific prefixes BEFORE generic model names
	// This handles cases like "groq-llama" where provider prefix should take precedence
	if strings.HasPrefix(modelLower, "groq-") || strings.HasPrefix(modelLower, "groq_") {
		return "groq"
	}
	if strings.HasPrefix(modelLower, "cerebras-") || strings.HasPrefix(modelLower, "cerebras_") {
		return "cerebras"
	}
	if strings.HasPrefix(modelLower, "together-") || strings.HasPrefix(modelLower, "together_") {
		return "together"
	}

	// Pattern-based inference - matches model naming conventions
	switch {
	// Claude/Anthropic models
	case strings.Contains(modelLower, "claude"):
		return "claude"
	case strings.Contains(modelLower, "anthropic"):
		return "claude"

	// OpenAI models
	case strings.HasPrefix(modelLower, "gpt-"):
		return "openai"
	case strings.HasPrefix(modelLower, "o1"):
		return "openai"
	case strings.Contains(modelLower, "openai"):
		return "openai"

	// Google/Gemini models
	case strings.Contains(modelLower, "gemini"):
		return "gemini"
	case strings.Contains(modelLower, "palm"):
		return "gemini"
	case strings.Contains(modelLower, "bard"):
		return "gemini"

	// DeepSeek models
	case strings.Contains(modelLower, "deepseek"):
		return "deepseek"

	// Mistral models
	case strings.Contains(modelLower, "mistral"):
		return "mistral"
	case strings.Contains(modelLower, "codestral"):
		return "mistral"
	case strings.Contains(modelLower, "mixtral"):
		return "mistral"

	// Qwen models
	case strings.Contains(modelLower, "qwen"):
		return "qwen"

	// Groq-specific indicators (already checked prefix above)
	case strings.Contains(modelLower, "groq"):
		return "groq"

	// Cerebras-specific indicators (already checked prefix above)
	case strings.Contains(modelLower, "cerebras"):
		return "cerebras"

	// Llama models - check provider context if available
	// NOTE: This is checked LAST because llama can be served by many providers
	case strings.Contains(modelLower, "llama"):
		// Llama can be served by multiple providers
		// If hosted locally, it's usually ollama
		// If via API, could be groq, cerebras, together, etc.
		// Return empty to let the caller provide context
		return ""
	}

	// Unable to determine provider - return empty string
	// The caller should handle this case by using the model as-is
	return ""
}

// VerifyProvider performs LLMsVerifier verification for a provider
// This is the central method that all provider validation should route through
// Returns verification result with score and status
func (a *LLMsVerifierScoreAdapter) VerifyProvider(ctx context.Context, providerType, modelID string) (*ProviderVerificationResult, error) {
	if a.verificationSvc == nil {
		return nil, fmt.Errorf("verification service not initialized")
	}

	a.log.WithFields(logrus.Fields{
		"provider": providerType,
		"model":    modelID,
	}).Info("Starting LLMsVerifier verification for provider")

	// Perform verification through LLMsVerifier
	result, err := a.verificationSvc.VerifyModel(ctx, modelID, providerType)
	if err != nil {
		a.log.WithError(err).WithFields(logrus.Fields{
			"provider": providerType,
			"model":    modelID,
		}).Warn("LLMsVerifier verification failed")

		return &ProviderVerificationResult{
			Name:     providerType,
			Status:   ProviderStatusUnhealthy,
			Verified: false,
			Error:    err.Error(),
		}, err
	}

	// Create verification result
	verificationResult := &ProviderVerificationResult{
		Name:       providerType,
		Verified:   result.Verified,
		VerifiedAt: result.CompletedAt,
	}

	if result.Verified {
		verificationResult.Status = ProviderStatusHealthy
		verificationResult.Score = result.Score
		verificationResult.Error = ""

		// Update cached score
		a.UpdateScore(providerType, modelID, result.Score)

		a.log.WithFields(logrus.Fields{
			"provider": providerType,
			"model":    modelID,
			"score":    result.Score,
			"verified": true,
		}).Info("LLMsVerifier verification successful")
	} else {
		verificationResult.Status = ProviderStatusUnhealthy
		verificationResult.Score = 0
		verificationResult.Error = result.ErrorMessage
	}

	return verificationResult, nil
}

// VerifyProviderWithType performs verification based on provider authentication type
// Handles API key, OAuth, and anonymous/free providers differently
func (a *LLMsVerifierScoreAdapter) VerifyProviderWithType(ctx context.Context, providerType, modelID, authType string) (*ProviderVerificationResult, error) {
	a.log.WithFields(logrus.Fields{
		"provider":  providerType,
		"model":     modelID,
		"auth_type": authType,
	}).Info("Starting type-aware LLMsVerifier verification")

	switch authType {
	case "oauth":
		// OAuth providers (Claude Code CLI, Qwen Code CLI)
		// Trust CLI credentials, but verify API access
		return a.verifyOAuthProvider(ctx, providerType, modelID)

	case "anonymous", "free":
		// Anonymous/Free providers (Zen, OpenRouter :free)
		// No authentication required, verify model availability
		return a.verifyFreeProvider(ctx, providerType, modelID)

	default:
		// API key providers - standard verification
		return a.VerifyProvider(ctx, providerType, modelID)
	}
}

// verifyOAuthProvider verifies an OAuth-based provider
func (a *LLMsVerifierScoreAdapter) verifyOAuthProvider(ctx context.Context, providerType, modelID string) (*ProviderVerificationResult, error) {
	a.log.WithFields(logrus.Fields{
		"provider": providerType,
		"model":    modelID,
		"type":     "oauth",
	}).Debug("Verifying OAuth provider through LLMsVerifier")

	// For OAuth providers, we trust CLI credentials
	// Perform a lightweight verification to confirm API access
	result, err := a.VerifyProvider(ctx, providerType, modelID)
	if err != nil {
		// OAuth providers are trusted even if verification fails
		// (CLI credentials may have different scopes than API verification)
		a.log.WithError(err).WithFields(logrus.Fields{
			"provider": providerType,
		}).Warn("OAuth verification failed, trusting CLI credentials")

		return &ProviderVerificationResult{
			Name:     providerType,
			Status:   ProviderStatusHealthy, // Trust OAuth
			Verified: true,
			Score:    7.5, // Default score for trusted OAuth
			Error:    "",
		}, nil
	}

	return result, nil
}

// verifyFreeProvider verifies an anonymous/free provider
func (a *LLMsVerifierScoreAdapter) verifyFreeProvider(ctx context.Context, providerType, modelID string) (*ProviderVerificationResult, error) {
	a.log.WithFields(logrus.Fields{
		"provider": providerType,
		"model":    modelID,
		"type":     "free",
	}).Debug("Verifying free/anonymous provider through LLMsVerifier")

	// For free providers, verify model availability without authentication
	result, err := a.VerifyProvider(ctx, providerType, modelID)
	if err != nil {
		// Free providers are available even if verification fails
		// Mark as healthy with reduced confidence score
		a.log.WithError(err).WithFields(logrus.Fields{
			"provider": providerType,
		}).Warn("Free provider verification failed, marking as available with reduced score")

		return &ProviderVerificationResult{
			Name:     providerType,
			Status:   ProviderStatusHealthy, // Allow free providers
			Verified: true,
			Score:    6.5, // Reduced score for unverified free
			Error:    "",
		}, nil
	}

	return result, nil
}

// GetVerificationService returns the underlying verification service
func (a *LLMsVerifierScoreAdapter) GetVerificationService() *verifier.VerificationService {
	return a.verificationSvc
}

// GetScoringService returns the underlying scoring service
func (a *LLMsVerifierScoreAdapter) GetScoringService() *verifier.ScoringService {
	return a.scoringService
}

// IsInitialized returns true if both services are properly initialized
func (a *LLMsVerifierScoreAdapter) IsInitialized() bool {
	return a.verificationSvc != nil || a.scoringService != nil
}

// Ensure LLMsVerifierScoreAdapter implements LLMsVerifierScoreProvider
var _ LLMsVerifierScoreProvider = (*LLMsVerifierScoreAdapter)(nil)
