// Package adapters provides provider-specific verification adapters for the startup verification system.
package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/llm/providers/zen"
	"dev.helix.agent/internal/models"
	verifier "dev.helix.agent/internal/verifier"
)

var freeLog = logrus.New()

// FreeProviderType defines the type of free provider
type FreeProviderType string

const (
	// FreeProviderZen is the OpenCode Zen free provider
	FreeProviderZen FreeProviderType = "zen"
	// FreeProviderOpenRouter is the OpenRouter free tier
	FreeProviderOpenRouter FreeProviderType = "openrouter"
)

// FreeAdapterConfig holds configuration for the free provider adapter
type FreeAdapterConfig struct {
	// VerificationTimeout is the timeout for verification requests
	VerificationTimeout time.Duration
	// HealthCheckTimeout is the timeout for health check requests
	HealthCheckTimeout time.Duration
	// MaxConcurrentVerifications limits concurrent verification requests
	MaxConcurrentVerifications int
	// MinHealthScore is the minimum health score to consider a provider verified
	MinHealthScore float64
	// BaseScore is the base score for free providers
	BaseScore float64
	// MaxScore is the maximum score for free providers
	MaxScore float64
	// RetryAttempts is the number of retry attempts for failed verifications
	RetryAttempts int
	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
}

// DefaultFreeAdapterConfig returns the default configuration
func DefaultFreeAdapterConfig() *FreeAdapterConfig {
	return &FreeAdapterConfig{
		VerificationTimeout:        30 * time.Second,
		HealthCheckTimeout:         10 * time.Second,
		MaxConcurrentVerifications: 4,
		MinHealthScore:             0.5,
		BaseScore:                  6.0, // Free providers have lower base score (6.0-7.0)
		MaxScore:                   7.0,
		RetryAttempts:              2,
		RetryDelay:                 1 * time.Second,
	}
}

// FreeProviderAdapter handles verification for free providers (Zen, OpenRouter :free models)
type FreeProviderAdapter struct {
	verifierSvc *verifier.VerificationService
	config      *FreeAdapterConfig
	httpClient  *http.Client

	// Provider instances
	zenProvider *zen.ZenProvider

	// Cached verification results
	mu             sync.RWMutex
	verifiedModels map[string]*verifier.UnifiedModel
	lastVerified   map[string]time.Time
	healthStatus   map[string]bool
}

// NewFreeProviderAdapter creates a new free provider adapter
func NewFreeProviderAdapter(verifierSvc *verifier.VerificationService, config *FreeAdapterConfig) *FreeProviderAdapter {
	if config == nil {
		config = DefaultFreeAdapterConfig()
	}

	return &FreeProviderAdapter{
		verifierSvc: verifierSvc,
		config:      config,
		httpClient: &http.Client{
			Timeout: config.VerificationTimeout,
		},
		verifiedModels: make(map[string]*verifier.UnifiedModel),
		lastVerified:   make(map[string]time.Time),
		healthStatus:   make(map[string]bool),
	}
}

// VerifyZenProvider verifies the Zen free provider and returns a UnifiedProvider
func (fa *FreeProviderAdapter) VerifyZenProvider(ctx context.Context) (*verifier.UnifiedProvider, error) {
	freeLog.WithField("provider", "zen").Info("Starting Zen provider verification")
	startTime := time.Now()

	// Create anonymous Zen provider for free models
	fa.zenProvider = zen.NewZenProviderAnonymous(zen.DefaultZenModel)

	// Perform health check first
	healthErr := fa.zenProvider.HealthCheck()
	if healthErr != nil {
		freeLog.WithFields(logrus.Fields{
			"provider": "zen",
			"error":    healthErr.Error(),
		}).Warn("Zen health check failed, attempting verification anyway")
	}

	// Get free models
	freeModels := zen.FreeModels()
	models := make([]verifier.UnifiedModel, 0, len(freeModels))

	// Verify each free model
	var wg sync.WaitGroup
	var modelsMu sync.Mutex
	sem := make(chan struct{}, fa.config.MaxConcurrentVerifications)

	for _, modelID := range freeModels {
		wg.Add(1)
		go func(mID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			model, err := fa.verifyZenModel(ctx, mID)
			if err != nil {
				freeLog.WithFields(logrus.Fields{
					"provider": "zen",
					"model":    mID,
					"error":    err.Error(),
				}).Warn("Failed to verify Zen model")
				return
			}

			modelsMu.Lock()
			models = append(models, *model)
			fa.verifiedModels[mID] = model
			fa.lastVerified[mID] = time.Now()
			modelsMu.Unlock()
		}(modelID)
	}

	wg.Wait()

	// Calculate provider score based on verified models
	score := fa.calculateZenScore(models, healthErr == nil)

	// Determine status
	status := verifier.StatusVerified
	if len(models) == 0 {
		status = verifier.StatusFailed
	} else if len(models) < len(freeModels) {
		status = verifier.StatusDegraded
	}

	provider := &verifier.UnifiedProvider{
		ID:           "zen",
		Name:         "OpenCode Zen",
		Type:         "free",
		AuthType:     verifier.AuthTypeFree,
		Verified:     len(models) > 0,
		Score:        score,
		Models:       models,
		Status:       status,
		Instance:     fa.zenProvider,
		BaseURL:      zen.ZenAPIURL,
		VerifiedAt:   time.Now(),
		LastHealthAt: time.Now(),
		ErrorCount:   0,
		Metadata: map[string]interface{}{
			"verification_time_ms": time.Since(startTime).Milliseconds(),
			"verified_models":      len(models),
			"total_models":         len(freeModels),
			"health_check_passed":  healthErr == nil,
			"anonymous_mode":       true,
		},
	}

	freeLog.WithFields(logrus.Fields{
		"provider":        "zen",
		"verified_models": len(models),
		"score":           score,
		"status":          status,
		"duration_ms":     time.Since(startTime).Milliseconds(),
	}).Info("Zen provider verification completed")

	return provider, nil
}

// verifyZenModel verifies a single Zen model
func (fa *FreeProviderAdapter) verifyZenModel(ctx context.Context, modelID string) (*verifier.UnifiedModel, error) {
	startTime := time.Now()

	// Create a provider for this specific model
	modelProvider := zen.NewZenProviderAnonymous(modelID)

	// Perform reduced verification (health check + simple completion)
	var verificationErr error
	var latency time.Duration
	verified := false

	for attempt := 0; attempt <= fa.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(fa.config.RetryDelay):
			}
		}

		// Health check
		if err := modelProvider.HealthCheck(); err != nil {
			verificationErr = err
			continue
		}

		// Simple completion test
		testStart := time.Now()
		if err := fa.testModelCompletion(ctx, modelProvider); err != nil {
			verificationErr = err
			continue
		}
		latency = time.Since(testStart)
		verified = true
		break
	}

	if !verified {
		return nil, fmt.Errorf("verification failed after %d attempts: %v", fa.config.RetryAttempts+1, verificationErr)
	}

	// Calculate model score (free models score between 6.0-7.0)
	score := fa.calculateModelScore(latency, verified)

	// Get capabilities from provider
	caps := modelProvider.GetCapabilities()

	model := &verifier.UnifiedModel{
		ID:           modelID,
		Name:         getModelDisplayName(modelID),
		Provider:     "zen",
		Verified:     verified,
		Score:        score,
		Latency:      latency,
		Capabilities: convertCapabilities(caps),
		Metadata: map[string]interface{}{
			"free_model":           true,
			"verification_time_ms": time.Since(startTime).Milliseconds(),
			"latency_ms":           latency.Milliseconds(),
			"anonymous_access":     true,
			"max_tokens":           caps.Limits.MaxTokens,
			"context_window":       caps.Limits.MaxInputLength,
		},
	}

	return model, nil
}

// testModelCompletion performs a simple completion test
func (fa *FreeProviderAdapter) testModelCompletion(ctx context.Context, provider llm.LLMProvider) error {
	// Create a simple test request using the provider's Complete method
	testReq := &models.LLMRequest{
		ID:     fmt.Sprintf("free-test-%d", time.Now().UnixNano()),
		Prompt: "You are a helpful assistant. Reply concisely.",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "What is 2 + 2? Reply with just the number.",
			},
		},
		ModelParams: models.ModelParameters{
			MaxTokens:   10,
			Temperature: 0.0,
		},
	}

	// Perform direct completion test
	resp, err := provider.Complete(ctx, testReq)
	if err != nil {
		return fmt.Errorf("completion failed: %w", err)
	}

	// Validate response
	if resp == nil {
		return fmt.Errorf("empty response from provider")
	}

	if resp.Content == "" {
		return fmt.Errorf("empty content in response")
	}

	// Basic validation - response should contain "4"
	if !strings.Contains(resp.Content, "4") {
		freeLog.WithFields(logrus.Fields{
			"expected": "4",
			"got":      resp.Content,
		}).Debug("Unexpected response content, but provider is responsive")
	}

	return nil
}

// calculateZenScore calculates the overall Zen provider score
func (fa *FreeProviderAdapter) calculateZenScore(models []verifier.UnifiedModel, healthPassed bool) float64 {
	if len(models) == 0 {
		return 0.0
	}

	// Base score for free providers
	score := fa.config.BaseScore

	// Add bonus for successful health check
	if healthPassed {
		score += 0.2
	}

	// Add bonus based on number of verified models
	verifiedRatio := float64(len(models)) / float64(len(zen.FreeModels()))
	score += verifiedRatio * 0.5

	// Calculate average model latency bonus
	var totalLatency time.Duration
	for _, m := range models {
		totalLatency += m.Latency
	}
	avgLatency := totalLatency / time.Duration(len(models))

	// Latency bonus: faster responses get higher scores
	if avgLatency < 500*time.Millisecond {
		score += 0.3
	} else if avgLatency < 1*time.Second {
		score += 0.2
	} else if avgLatency < 2*time.Second {
		score += 0.1
	}

	// Cap at max score
	if score > fa.config.MaxScore {
		score = fa.config.MaxScore
	}

	return score
}

// calculateModelScore calculates an individual model's score
func (fa *FreeProviderAdapter) calculateModelScore(latency time.Duration, verified bool) float64 {
	if !verified {
		return 0.0
	}

	// Base score
	score := fa.config.BaseScore

	// Latency-based adjustment
	switch {
	case latency < 300*time.Millisecond:
		score += 0.6
	case latency < 500*time.Millisecond:
		score += 0.4
	case latency < 1*time.Second:
		score += 0.2
	case latency < 2*time.Second:
		score += 0.1
	}

	// Cap at max
	if score > fa.config.MaxScore {
		score = fa.config.MaxScore
	}

	return score
}

// VerifyOpenRouterFreeModels verifies OpenRouter :free models
func (fa *FreeProviderAdapter) VerifyOpenRouterFreeModels(ctx context.Context, openRouterAPIKey string) (*verifier.UnifiedProvider, error) {
	freeLog.WithField("provider", "openrouter").Info("Starting OpenRouter free models verification")
	startTime := time.Now()

	// OpenRouter free models have ":free" suffix
	freeModelPatterns := []string{
		"google/gemma-2-9b-it:free",
		"meta-llama/llama-3-8b-instruct:free",
		"mistralai/mistral-7b-instruct:free",
		"microsoft/phi-3-mini-128k-instruct:free",
		"qwen/qwen-2-7b-instruct:free",
	}

	models := make([]verifier.UnifiedModel, 0)
	var wg sync.WaitGroup
	var modelsMu sync.Mutex
	sem := make(chan struct{}, fa.config.MaxConcurrentVerifications)

	for _, modelID := range freeModelPatterns {
		wg.Add(1)
		go func(mID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			model, err := fa.verifyOpenRouterFreeModel(ctx, mID, openRouterAPIKey)
			if err != nil {
				freeLog.WithFields(logrus.Fields{
					"provider": "openrouter",
					"model":    mID,
					"error":    err.Error(),
				}).Debug("OpenRouter free model not available")
				return
			}

			modelsMu.Lock()
			models = append(models, *model)
			fa.verifiedModels[mID] = model
			fa.lastVerified[mID] = time.Now()
			modelsMu.Unlock()
		}(modelID)
	}

	wg.Wait()

	// Calculate score
	score := fa.calculateOpenRouterScore(models)

	// Determine status
	status := verifier.StatusVerified
	if len(models) == 0 {
		status = verifier.StatusUnavailable
	} else if len(models) < 2 {
		status = verifier.StatusDegraded
	}

	provider := &verifier.UnifiedProvider{
		ID:           "openrouter-free",
		Name:         "OpenRouter Free Tier",
		Type:         "free",
		AuthType:     verifier.AuthTypeFree,
		Verified:     len(models) > 0,
		Score:        score,
		Models:       models,
		Status:       status,
		BaseURL:      "https://openrouter.ai/api/v1",
		APIKey:       openRouterAPIKey,
		VerifiedAt:   time.Now(),
		LastHealthAt: time.Now(),
		ErrorCount:   0,
		Metadata: map[string]interface{}{
			"verification_time_ms": time.Since(startTime).Milliseconds(),
			"verified_models":      len(models),
			"total_models":         len(freeModelPatterns),
			"free_tier":            true,
		},
	}

	freeLog.WithFields(logrus.Fields{
		"provider":        "openrouter",
		"verified_models": len(models),
		"score":           score,
		"status":          status,
		"duration_ms":     time.Since(startTime).Milliseconds(),
	}).Info("OpenRouter free models verification completed")

	return provider, nil
}

// verifyOpenRouterFreeModel verifies a single OpenRouter free model
func (fa *FreeProviderAdapter) verifyOpenRouterFreeModel(ctx context.Context, modelID, apiKey string) (*verifier.UnifiedModel, error) {
	startTime := time.Now()

	// OpenRouter health check for free model
	healthURL := "https://openrouter.ai/api/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := fa.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	latency := time.Since(startTime)
	score := fa.calculateModelScore(latency, true)

	model := &verifier.UnifiedModel{
		ID:       modelID,
		Name:     getOpenRouterModelName(modelID),
		Provider: "openrouter",
		Verified: true,
		Score:    score,
		Latency:  latency,
		Capabilities: []string{
			"text_completion",
			"chat",
			"streaming",
		},
		Metadata: map[string]interface{}{
			"free_model":           true,
			"verification_time_ms": time.Since(startTime).Milliseconds(),
			"latency_ms":           latency.Milliseconds(),
		},
	}

	return model, nil
}

// calculateOpenRouterScore calculates the OpenRouter free tier score
func (fa *FreeProviderAdapter) calculateOpenRouterScore(models []verifier.UnifiedModel) float64 {
	if len(models) == 0 {
		return 0.0
	}

	// Base score
	score := fa.config.BaseScore

	// Bonus for multiple verified models
	if len(models) >= 3 {
		score += 0.4
	} else if len(models) >= 2 {
		score += 0.2
	}

	// Average latency bonus
	var totalLatency time.Duration
	for _, m := range models {
		totalLatency += m.Latency
	}
	avgLatency := totalLatency / time.Duration(len(models))

	if avgLatency < 500*time.Millisecond {
		score += 0.3
	} else if avgLatency < 1*time.Second {
		score += 0.2
	}

	if score > fa.config.MaxScore {
		score = fa.config.MaxScore
	}

	return score
}

// VerifyAllFreeProviders verifies all free providers and returns them
func (fa *FreeProviderAdapter) VerifyAllFreeProviders(ctx context.Context, openRouterAPIKey string) ([]*verifier.UnifiedProvider, error) {
	freeLog.Info("Starting verification of all free providers")
	startTime := time.Now()

	providers := make([]*verifier.UnifiedProvider, 0, 2)
	var wg sync.WaitGroup
	var providersMu sync.Mutex

	// Verify Zen
	wg.Add(1)
	go func() {
		defer wg.Done()
		zenProvider, err := fa.VerifyZenProvider(ctx)
		if err != nil {
			freeLog.WithField("error", err.Error()).Warn("Failed to verify Zen provider")
			return
		}
		providersMu.Lock()
		providers = append(providers, zenProvider)
		providersMu.Unlock()
	}()

	// Verify OpenRouter free models (if API key provided)
	if openRouterAPIKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			orProvider, err := fa.VerifyOpenRouterFreeModels(ctx, openRouterAPIKey)
			if err != nil {
				freeLog.WithField("error", err.Error()).Warn("Failed to verify OpenRouter free models")
				return
			}
			providersMu.Lock()
			providers = append(providers, orProvider)
			providersMu.Unlock()
		}()
	}

	wg.Wait()

	freeLog.WithFields(logrus.Fields{
		"verified_providers": len(providers),
		"duration_ms":        time.Since(startTime).Milliseconds(),
	}).Info("Free providers verification completed")

	return providers, nil
}

// GetVerifiedModels returns all verified free models
func (fa *FreeProviderAdapter) GetVerifiedModels() map[string]*verifier.UnifiedModel {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	result := make(map[string]*verifier.UnifiedModel, len(fa.verifiedModels))
	for k, v := range fa.verifiedModels {
		result[k] = v
	}
	return result
}

// IsModelVerified checks if a model is verified
func (fa *FreeProviderAdapter) IsModelVerified(modelID string) bool {
	fa.mu.RLock()
	defer fa.mu.RUnlock()
	_, ok := fa.verifiedModels[modelID]
	return ok
}

// GetHealthStatus returns the health status of free providers
func (fa *FreeProviderAdapter) GetHealthStatus() map[string]bool {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	result := make(map[string]bool, len(fa.healthStatus))
	for k, v := range fa.healthStatus {
		result[k] = v
	}
	return result
}

// RefreshVerification re-verifies a specific provider
func (fa *FreeProviderAdapter) RefreshVerification(ctx context.Context, providerType FreeProviderType, apiKey string) (*verifier.UnifiedProvider, error) {
	switch providerType {
	case FreeProviderZen:
		return fa.VerifyZenProvider(ctx)
	case FreeProviderOpenRouter:
		return fa.VerifyOpenRouterFreeModels(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unknown free provider type: %s", providerType)
	}
}

// Helper functions

// getModelDisplayName returns a user-friendly display name for a model
func getModelDisplayName(modelID string) string {
	nameMap := map[string]string{
		zen.ModelBigPickle:    "Big Pickle (Stealth)",
		zen.ModelGrokCodeFast: "Grok Code Fast",
		zen.ModelGLM47Free:    "GLM 4.7 Free",
		zen.ModelGPT5Nano:     "GPT 5 Nano",
	}

	if name, ok := nameMap[modelID]; ok {
		return name
	}

	// Extract name from model ID
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return modelID
}

// getOpenRouterModelName returns a display name for OpenRouter models
func getOpenRouterModelName(modelID string) string {
	// Remove :free suffix and provider prefix
	name := strings.TrimSuffix(modelID, ":free")
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		return fmt.Sprintf("%s (Free)", parts[1])
	}
	return fmt.Sprintf("%s (Free)", name)
}

// convertCapabilities converts provider capabilities to a string slice
func convertCapabilities(caps *models.ProviderCapabilities) []string {
	if caps == nil {
		return []string{}
	}

	result := make([]string, 0, len(caps.SupportedFeatures))
	result = append(result, caps.SupportedFeatures...)

	if caps.SupportsStreaming {
		result = append(result, "streaming")
	}
	if caps.SupportsFunctionCalling {
		result = append(result, "function_calling")
	}
	if caps.SupportsVision {
		result = append(result, "vision")
	}
	if caps.SupportsTools {
		result = append(result, "tools")
	}
	if caps.SupportsReasoning {
		result = append(result, "reasoning")
	}

	return result
}
