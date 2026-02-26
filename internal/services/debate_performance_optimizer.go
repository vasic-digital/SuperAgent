package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
)

type DebateOptimizationConfig struct {
	EnableParallelExecution   bool          `json:"enable_parallel_execution"`
	EnableResponseCaching     bool          `json:"enable_response_caching"`
	EnableEarlyTermination    bool          `json:"enable_early_termination"`
	EnableStreaming           bool          `json:"enable_streaming"`
	EnableSmartFallback       bool          `json:"enable_smart_fallback"`
	MaxParallelRequests       int           `json:"max_parallel_requests"`
	CacheTTL                  time.Duration `json:"cache_ttl"`
	EarlyTerminationThreshold float64       `json:"early_termination_threshold"`
	RequestTimeout            time.Duration `json:"request_timeout"`
	FallbackTimeout           time.Duration `json:"fallback_timeout"`
}

func DefaultDebateOptimizationConfig() DebateOptimizationConfig {
	return DebateOptimizationConfig{
		EnableParallelExecution:   true,
		EnableResponseCaching:     true,
		EnableEarlyTermination:    true,
		EnableStreaming:           true,
		EnableSmartFallback:       true,
		MaxParallelRequests:       3,
		CacheTTL:                  5 * time.Minute,
		EarlyTerminationThreshold: 0.95,
		RequestTimeout:            60 * time.Second,
		FallbackTimeout:           30 * time.Second,
	}
}

type CachedResponse struct {
	Response  *models.LLMResponse
	Timestamp time.Time
	Model     string
	Provider  string
}

type DebatePerformanceOptimizer struct {
	config    DebateOptimizationConfig
	cache     sync.Map
	logger    *logrus.Logger
	registry  *ProviderRegistry
	semaphore chan struct{}
	mu        sync.RWMutex
	stats     *OptimizationStats
}

type OptimizationStats struct {
	CacheHits          int64 `json:"cache_hits"`
	CacheMisses        int64 `json:"cache_misses"`
	ParallelRequests   int64 `json:"parallel_requests"`
	EarlyTerminations  int64 `json:"early_terminations"`
	FallbacksTriggered int64 `json:"fallbacks_triggered"`
	TotalRequests      int64 `json:"total_requests"`
	AverageLatencyMs   int64 `json:"average_latency_ms"`
}

func NewDebatePerformanceOptimizer(config DebateOptimizationConfig, registry *ProviderRegistry, logger *logrus.Logger) *DebatePerformanceOptimizer {
	if logger == nil {
		logger = logrus.New()
	}

	return &DebatePerformanceOptimizer{
		config:    config,
		logger:    logger,
		registry:  registry,
		semaphore: make(chan struct{}, config.MaxParallelRequests),
		stats:     &OptimizationStats{},
	}
}

func (dpo *DebatePerformanceOptimizer) ExecuteWithOptimization(
	ctx context.Context,
	member *DebateTeamMember,
	prompt string,
	previousResponses map[DebateTeamPosition]string,
) (*models.LLMResponse, error) {
	startTime := time.Now()
	dpo.mu.Lock()
	dpo.stats.TotalRequests++
	dpo.mu.Unlock()

	if dpo.config.EnableResponseCaching {
		if cached := dpo.getCachedResponse(prompt, member.ModelName); cached != nil {
			dpo.mu.Lock()
			dpo.stats.CacheHits++
			dpo.mu.Unlock()
			dpo.logger.WithFields(logrus.Fields{
				"model":     member.ModelName,
				"provider":  member.ProviderName,
				"cache_hit": true,
			}).Debug("Debate response served from cache")
			return cached.Response, nil
		}
		dpo.mu.Lock()
		dpo.stats.CacheMisses++
		dpo.mu.Unlock()
	}

	var response *models.LLMResponse
	var err error

	if dpo.config.EnableSmartFallback && member.Fallbacks != nil && len(member.Fallbacks) > 0 {
		response, err = dpo.executeWithSmartFallback(ctx, member, prompt)
	} else {
		response, err = dpo.executeSingle(ctx, member, prompt)
	}

	if err == nil && dpo.config.EnableResponseCaching {
		dpo.cacheResponse(prompt, member.ModelName, member.ProviderName, response)
	}

	dpo.mu.Lock()
	latency := time.Since(startTime).Milliseconds()
	dpo.stats.AverageLatencyMs = (dpo.stats.AverageLatencyMs + latency) / 2
	dpo.mu.Unlock()

	return response, err
}

func (dpo *DebatePerformanceOptimizer) ExecuteParallel(
	ctx context.Context,
	members []*DebateTeamMember,
	prompt string,
) map[DebateTeamPosition]*models.LLMResponse {
	results := make(map[DebateTeamPosition]*models.LLMResponse)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, member := range members {
		wg.Add(1)
		go func(m *DebateTeamMember) {
			defer wg.Done()

			dpo.semaphore <- struct{}{}
			defer func() { <-dpo.semaphore }()

			dpo.mu.Lock()
			dpo.stats.ParallelRequests++
			dpo.mu.Unlock()

			resp, err := dpo.ExecuteWithOptimization(ctx, m, prompt, nil)
			if err == nil && resp != nil {
				mu.Lock()
				results[m.Position] = resp
				mu.Unlock()
			}
		}(member)
	}

	wg.Wait()
	return results
}

func (dpo *DebatePerformanceOptimizer) executeSingle(
	ctx context.Context,
	member *DebateTeamMember,
	prompt string,
) (*models.LLMResponse, error) {
	provider := member.Provider
	if provider == nil {
		var err error
		provider, err = dpo.registry.GetProvider(member.ProviderName)
		if err != nil {
			return nil, err
		}
	}

	req := &models.LLMRequest{
		Messages: []models.Message{
			{Role: "user", Content: prompt},
		},
		ModelParams: models.ModelParameters{
			Model: member.ModelName,
		},
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dpo.config.RequestTimeout)
	defer cancel()

	return provider.Complete(timeoutCtx, req)
}

func (dpo *DebatePerformanceOptimizer) executeWithSmartFallback(
	ctx context.Context,
	member *DebateTeamMember,
	prompt string,
) (*models.LLMResponse, error) {
	response, err := dpo.executeSingle(ctx, member, prompt)
	if err == nil {
		return response, nil
	}

	dpo.logger.WithFields(logrus.Fields{
		"position": member.Position,
		"model":    member.ModelName,
		"error":    err.Error(),
	}).Warn("Primary provider failed, trying fallbacks")

	dpo.mu.Lock()
	dpo.stats.FallbacksTriggered++
	dpo.mu.Unlock()

	for i, fallback := range member.Fallbacks {
		if fallback == nil || !fallback.IsActive {
			continue
		}

		dpo.logger.WithFields(logrus.Fields{
			"position":       member.Position,
			"fallback_index": i,
			"model":          fallback.ModelName,
			"provider":       fallback.ProviderName,
		}).Info("Trying fallback provider")

		timeoutCtx, cancel := context.WithTimeout(ctx, dpo.config.FallbackTimeout)
		response, fallbackErr := dpo.executeSingle(timeoutCtx, fallback, prompt)
		cancel()

		if fallbackErr == nil {
			return response, nil
		}

		dpo.logger.WithFields(logrus.Fields{
			"position":       member.Position,
			"fallback_index": i,
			"error":          fallbackErr.Error(),
		}).Warn("Fallback provider failed")
	}

	return nil, err
}

func (dpo *DebatePerformanceOptimizer) ShouldTerminateEarly(responses map[DebateTeamPosition]string) bool {
	if !dpo.config.EnableEarlyTermination || len(responses) < 3 {
		return false
	}

	consensus := dpo.calculateConsensus(responses)
	if consensus >= dpo.config.EarlyTerminationThreshold {
		dpo.mu.Lock()
		dpo.stats.EarlyTerminations++
		dpo.mu.Unlock()

		dpo.logger.WithFields(logrus.Fields{
			"consensus": consensus,
			"threshold": dpo.config.EarlyTerminationThreshold,
			"responses": len(responses),
		}).Info("Early termination triggered due to high consensus")

		return true
	}
	return false
}

func (dpo *DebatePerformanceOptimizer) calculateConsensus(responses map[DebateTeamPosition]string) float64 {
	if len(responses) < 2 {
		return 0.0
	}

	keyPoints := make(map[string]int)
	totalPoints := 0

	for _, resp := range responses {
		words := tokenize(resp)
		for _, word := range words {
			if len(word) > 4 {
				keyPoints[word]++
				totalPoints++
			}
		}
	}

	if totalPoints == 0 {
		return 0.0
	}

	agreementCount := 0
	for _, count := range keyPoints {
		if count > 1 {
			agreementCount += count
		}
	}

	return float64(agreementCount) / float64(totalPoints)
}

func (dpo *DebatePerformanceOptimizer) getCachedResponse(prompt, model string) *CachedResponse {
	cacheKey := dpo.generateCacheKey(prompt, model)
	if cached, ok := dpo.cache.Load(cacheKey); ok {
		cachedResp := cached.(*CachedResponse)
		if time.Since(cachedResp.Timestamp) < dpo.config.CacheTTL {
			return cachedResp
		}
		dpo.cache.Delete(cacheKey)
	}
	return nil
}

func (dpo *DebatePerformanceOptimizer) cacheResponse(prompt, model, provider string, response *models.LLMResponse) {
	cacheKey := dpo.generateCacheKey(prompt, model)
	dpo.cache.Store(cacheKey, &CachedResponse{
		Response:  response,
		Timestamp: time.Now(),
		Model:     model,
		Provider:  provider,
	})
}

func (dpo *DebatePerformanceOptimizer) generateCacheKey(prompt, model string) string {
	return model + ":" + hashString(prompt)
}

func (dpo *DebatePerformanceOptimizer) GetStats() *OptimizationStats {
	dpo.mu.RLock()
	defer dpo.mu.RUnlock()
	return dpo.stats
}

func (dpo *DebatePerformanceOptimizer) ClearCache() {
	dpo.cache = sync.Map{}
	dpo.logger.Info("Debate response cache cleared")
}

func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	result := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		if word != "" {
			result = append(result, word)
		}
	}
	return result
}

func hashString(s string) string {
	h := uint32(2166136261)
	for _, c := range s {
		h ^= uint32(c)
		h *= 16777619
	}
	return fmt.Sprintf("%08x", h)
}
