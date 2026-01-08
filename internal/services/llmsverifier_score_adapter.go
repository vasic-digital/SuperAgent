package services

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/helixagent/helixagent/internal/verifier"
)

// LLMsVerifierScoreAdapter implements LLMsVerifierScoreProvider interface
// It connects ProviderDiscovery to the actual LLMsVerifier scoring system
type LLMsVerifierScoreAdapter struct {
	scoringService     *verifier.ScoringService
	verificationSvc    *verifier.VerificationService
	providerScores     map[string]float64 // Cached provider scores
	modelScores        map[string]float64 // Cached model scores
	mu                 sync.RWMutex
	log                *logrus.Logger
	lastRefresh        time.Time
	refreshInterval    time.Duration
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
func (a *LLMsVerifierScoreAdapter) refreshFromScoringService(ctx context.Context) {
	// Get known models from scoring service
	knownModels := []string{
		"gemini-pro", "gemini-1.5-flash", "gemini-2.0-flash",
		"claude-3-5-sonnet-20241022", "claude-3-opus-20240229",
		"gpt-4", "gpt-4o",
		"deepseek-chat", "deepseek-coder",
		"mistral-large-latest",
		"qwen-turbo",
		"llama3.2", "llama2",
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

			// Map model to provider
			provider := mapModelToProvider(modelID)
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

// mapModelToProvider maps a model ID to its provider type
func mapModelToProvider(modelID string) string {
	modelProviderMap := map[string]string{
		"gemini-pro":                   "gemini",
		"gemini-1.5-flash":             "gemini",
		"gemini-2.0-flash":             "gemini",
		"gemini-2.0-flash-exp":         "gemini",
		"claude-3-5-sonnet-20241022":   "claude",
		"claude-3-opus-20240229":       "claude",
		"claude-3-sonnet-20240229":     "claude",
		"gpt-4":                        "openai",
		"gpt-4o":                       "openai",
		"gpt-4-turbo":                  "openai",
		"deepseek-chat":                "deepseek",
		"deepseek-coder":               "deepseek",
		"mistral-large-latest":         "mistral",
		"mistral-small-latest":         "mistral",
		"codestral-latest":             "mistral",
		"qwen-turbo":                   "qwen",
		"qwen-plus":                    "qwen",
		"llama3.2":                     "ollama",
		"llama2":                       "ollama",
		"llama-3.1-70b-versatile":      "groq",
		"llama3.1-70b":                 "cerebras",
	}

	if provider, ok := modelProviderMap[modelID]; ok {
		return provider
	}
	return ""
}

// Ensure LLMsVerifierScoreAdapter implements LLMsVerifierScoreProvider
var _ LLMsVerifierScoreProvider = (*LLMsVerifierScoreAdapter)(nil)
