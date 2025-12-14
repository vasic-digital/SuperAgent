package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/llm"
	"github.com/superagent/superagent/internal/models"
)

// AIDebateIntegration provides integration between AI debate system and existing ensemble infrastructure
type AIDebateIntegration struct {
	config    *config.AIDebateConfig
	ensemble  *llm.EnsembleService
	providers map[string]llm.LLMProvider
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// EnsembleService represents the existing ensemble service (placeholder for actual implementation)
type EnsembleService struct {
	providers []llm.LLMProvider
	config    *config.LLMConfig
}

// NewAIDebateIntegration creates a new AI debate integration service
func NewAIDebateIntegration(cfg *config.AIDebateConfig, logger *logrus.Logger) (*AIDebateIntegration, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	integration := &AIDebateIntegration{
		config:    cfg,
		providers: make(map[string]llm.LLMProvider),
		logger:    logger,
	}

	// Initialize providers based on configuration
	if err := integration.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return integration, nil
}

// initializeProviders initializes LLM providers from debate configuration
func (i *AIDebateIntegration) initializeProviders() error {
	for _, participant := range i.config.Participants {
		if !participant.Enabled {
			continue
		}

		// Create providers for each enabled LLM in the participant's chain
		for j, llmConfig := range participant.LLMs {
			if !llmConfig.Enabled {
				continue
			}

			providerKey := fmt.Sprintf("%s_%s", participant.Name, llmConfig.Name)
			provider, err := i.createProvider(&llmConfig)
			if err != nil {
				return fmt.Errorf("failed to create provider %s: %w", providerKey, err)
			}

			i.mu.Lock()
			i.providers[providerKey] = provider
			i.mu.Unlock()

			i.logger.Infof("Initialized provider %s for participant %s (LLM %d in chain)",
				providerKey, participant.Name, j+1)
		}
	}

	return nil
}

// createProvider creates an LLM provider based on configuration
func (i *AIDebateIntegration) createProvider(llmConfig *config.LLMConfiguration) (llm.LLMProvider, error) {
	// Map to existing provider implementations
	switch llmConfig.Provider {
	case "claude":
		return &ClaudeProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	case "deepseek":
		return &DeepSeekProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	case "gemini":
		return &GeminiProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	case "qwen":
		return &QwenProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	case "zai":
		return &ZaiProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	case "ollama":
		return &OllamaProvider{
			baseURL: llmConfig.BaseURL,
			model:   llmConfig.Model,
			timeout: time.Duration(llmConfig.Timeout) * time.Millisecond,
		}, nil

	case "openrouter":
		return &OpenRouterProvider{
			apiKey:     llmConfig.APIKey,
			model:      llmConfig.Model,
			baseURL:    llmConfig.BaseURL,
			timeout:    time.Duration(llmConfig.Timeout) * time.Millisecond,
			maxRetries: llmConfig.MaxRetries,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", llmConfig.Provider)
	}
}

// ConductEnsembleDebate conducts a debate using the existing ensemble infrastructure
func (i *AIDebateIntegration) ConductEnsembleDebate(ctx context.Context, topic string, initialContext string) (*DebateResult, error) {
	if !i.config.Enabled {
		return nil, fmt.Errorf("AI debate integration is disabled")
	}

	startTime := time.Now()
	sessionID := fmt.Sprintf("ensemble_debate_%d", startTime.Unix())

	i.logger.Infof("Starting ensemble debate session %s on topic: %s", sessionID, topic)

	// Create debate request
	debateRequest := &DebateRequest{
		Topic:              topic,
		Context:            initialContext,
		MaxRounds:          i.config.MaximalRepeatRounds,
		ConsensusThreshold: i.config.ConsensusThreshold,
		Timeout:            time.Duration(i.config.MaxResponseTime) * time.Millisecond,
	}

	// Execute debate using ensemble
	result, err := i.executeEnsembleDebate(ctx, debateRequest)
	if err != nil {
		i.logger.Errorf("Ensemble debate failed: %v", err)
		return nil, fmt.Errorf("ensemble debate failed: %w", err)
	}

	i.logger.Infof("Completed ensemble debate session %s in %v with consensus: %v",
		sessionID, result.Duration, result.Consensus != nil && result.Consensus.Reached)

	return result, nil
}

// executeEnsembleDebate executes the debate using ensemble infrastructure
func (i *AIDebateIntegration) executeEnsembleDebate(ctx context.Context, request *DebateRequest) (*DebateResult, error) {
	var responses []ParticipantResponse
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Collect responses from all participants
	for _, participant := range i.config.Participants {
		if !participant.Enabled {
			continue
		}

		wg.Add(1)
		go func(p config.DebateParticipant) {
			defer wg.Done()

			response, err := i.getParticipantEnsembleResponse(ctx, p, request)
			if err != nil {
				i.logger.Errorf("Failed to get ensemble response from %s: %v", p.Name, err)
				return
			}

			mu.Lock()
			responses = append(responses, response)
			mu.Unlock()
		}(participant)
	}

	// Wait for all responses or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// All responses collected
	case <-time.After(request.Timeout):
		i.logger.Warn("Timeout waiting for ensemble responses")
	}

	if len(responses) < 2 {
		return nil, fmt.Errorf("insufficient responses for debate: got %d, need at least 2", len(responses))
	}

	// Analyze consensus and create result
	consensus := i.analyzeEnsembleConsensus(responses)
	result := i.createEnsembleResult(request, responses, consensus)

	return result, nil
}

// getParticipantEnsembleResponse gets a response from a participant using ensemble
func (i *AIDebateIntegration) getParticipantEnsembleResponse(ctx context.Context, participant config.DebateParticipant, request *DebateRequest) (ParticipantResponse, error) {
	startTime := time.Now()

	// Try primary LLM first
	primaryLLM := participant.GetPrimaryLLM()
	if primaryLLM == nil {
		return ParticipantResponse{}, fmt.Errorf("no primary LLM found for participant %s", participant.Name)
	}

	providerKey := fmt.Sprintf("%s_%s", participant.Name, primaryLLM.Name)

	i.mu.RLock()
	provider, exists := i.providers[providerKey]
	i.mu.RUnlock()

	if !exists {
		return ParticipantResponse{}, fmt.Errorf("provider %s not found", providerKey)
	}

	// Create LLM request using existing models
	llmRequest := &models.LLMRequest{
		Prompt: fmt.Sprintf("Debate Topic: %s\n\nContext: %s\n\nAs %s (%s), provide your analysis and perspective.",
			request.Topic, request.Context, participant.Name, participant.Role),
		ModelParams: models.ModelParameters{
			Temperature: primaryLLM.Temperature,
			MaxTokens:   primaryLLM.MaxTokens,
			TopP:        primaryLLM.TopP,
		},
		RequestType: "debate_participation",
	}

	// Try primary provider
	response, err := provider.Complete(llmRequest)
	if err != nil {
		i.logger.Warnf("Primary provider %s failed for participant %s: %v", providerKey, participant.Name, err)

		// Try fallback providers
		fallbackResponses, fallbackErr := i.tryFallbackProviders(ctx, participant, request)
		if fallbackErr != nil {
			return ParticipantResponse{}, fmt.Errorf("all providers failed for participant %s: %w", participant.Name, fallbackErr)
		}

		// Use best fallback response
		if len(fallbackResponses) > 0 {
			response = fallbackResponses[0]
		} else {
			return ParticipantResponse{}, fmt.Errorf("no successful fallback responses for participant %s", participant.Name)
		}
	}

	responseTime := time.Since(startTime)

	return ParticipantResponse{
		ParticipantName: participant.Name,
		LLMName:         primaryLLM.Name,
		Content:         response.Content,
		Confidence:      response.Confidence,
		QualityScore:    response.SelectionScore,
		ResponseTime:    responseTime,
		Timestamp:       startTime,
		Metadata:        response.Metadata,
	}, nil
}

// tryFallbackProviders tries fallback providers for a participant
func (i *AIDebateIntegration) tryFallbackProviders(ctx context.Context, participant config.DebateParticipant, request *DebateRequest) ([]*models.LLMResponse, error) {
	var responses []*models.LLMResponse
	var mu sync.Mutex
	var wg sync.WaitGroup

	fallbackLLMs := participant.GetFallbackLLMs()
	if len(fallbackLLMs) == 0 {
		return nil, fmt.Errorf("no fallback LLMs available for participant %s", participant.Name)
	}

	// Try each fallback LLM
	for _, fallbackLLM := range fallbackLLMs {
		wg.Add(1)
		go func(llmConfig config.LLMConfiguration) {
			defer wg.Done()

			providerKey := fmt.Sprintf("%s_%s", participant.Name, llmConfig.Name)

			i.mu.RLock()
			provider, exists := i.providers[providerKey]
			i.mu.RUnlock()

			if !exists {
				i.logger.Warnf("Fallback provider %s not found", providerKey)
				return
			}

			// Create LLM request
			llmRequest := &models.LLMRequest{
				Prompt: fmt.Sprintf("Debate Topic: %s\n\nContext: %s\n\nAs %s (%s), provide your analysis and perspective.",
					request.Topic, request.Context, participant.Name, participant.Role),
				ModelParams: models.ModelParameters{
					Temperature: llmConfig.Temperature,
					MaxTokens:   llmConfig.MaxTokens,
					TopP:        llmConfig.TopP,
				},
				RequestType: "debate_participation",
			}

			response, err := provider.Complete(llmRequest)
			if err != nil {
				i.logger.Warnf("Fallback provider %s failed: %v", providerKey, err)
				return
			}

			mu.Lock()
			responses = append(responses, response)
			mu.Unlock()
		}(fallbackLLM)
	}

	// Wait for fallback attempts
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// All fallback attempts completed
	case <-time.After(time.Duration(participant.ResponseTimeout) * time.Millisecond):
		i.logger.Warn("Timeout waiting for fallback responses")
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("all fallback providers failed for participant %s", participant.Name)
	}

	return responses, nil
}

// analyzeEnsembleConsensus analyzes consensus among ensemble responses
func (i *AIDebateIntegration) analyzeEnsembleConsensus(responses []ParticipantResponse) *ConsensusResult {
	if len(responses) == 0 {
		return &ConsensusResult{
			Reached:        false,
			ConsensusLevel: 0.0,
			AgreementScore: 0.0,
			QualityScore:   0.0,
			Summary:        "No responses available for consensus analysis",
		}
	}

	// Calculate average scores
	avgConfidence := 0.0
	avgQuality := 0.0
	for _, response := range responses {
		avgConfidence += response.Confidence
		avgQuality += response.QualityScore
	}
	avgConfidence /= float64(len(responses))
	avgQuality /= float64(len(responses))

	// Calculate consensus level
	consensusLevel := (avgConfidence + avgQuality) / 2.0
	consensusReached := consensusLevel >= i.config.ConsensusThreshold

	summary := fmt.Sprintf("Ensemble analysis of %d responses: avg confidence %.2f, avg quality %.2f",
		len(responses), avgConfidence, avgQuality)

	return &ConsensusResult{
		Reached:        consensusReached,
		ConsensusLevel: consensusLevel,
		AgreementScore: avgConfidence,
		QualityScore:   avgQuality,
		Summary:        summary,
		KeyPoints:      i.extractEnsembleKeyPoints(responses),
	}
}

// extractEnsembleKeyPoints extracts key points from ensemble responses
func (i *AIDebateIntegration) extractEnsembleKeyPoints(responses []ParticipantResponse) []string {
	var keyPoints []string
	pointFrequency := make(map[string]int)

	// Simple keyword extraction from responses
	for _, response := range responses {
		keywords := extractKeywords(response.Content)
		for _, keyword := range keywords {
			pointFrequency[keyword]++
		}
	}

	// Get most frequent points
	for point, count := range pointFrequency {
		if count >= len(responses)/2 { // Point mentioned by at least half
			keyPoints = append(keyPoints, point)
		}
	}

	return keyPoints
}

// createEnsembleResult creates the final ensemble debate result
func (i *AIDebateIntegration) createEnsembleResult(request *DebateRequest, responses []ParticipantResponse, consensus *ConsensusResult) *DebateResult {
	// Find best response
	bestResponse := responses[0]
	bestScore := bestResponse.Confidence + bestResponse.QualityScore

	for _, response := range responses[1:] {
		score := response.Confidence + response.QualityScore
		if score > bestScore {
			bestScore = score
			bestResponse = response
		}
	}

	// Calculate quality metrics
	qualityMetrics := make(map[string]float64)
	qualityMetrics["avg_confidence"] = 0.0
	qualityMetrics["avg_quality"] = 0.0
	qualityMetrics["response_time_avg"] = 0.0

	for _, response := range responses {
		qualityMetrics["avg_confidence"] += response.Confidence
		qualityMetrics["avg_quality"] += response.QualityScore
		qualityMetrics["response_time_avg"] += float64(response.ResponseTime.Milliseconds())
	}

	qualityMetrics["avg_confidence"] /= float64(len(responses))
	qualityMetrics["avg_quality"] /= float64(len(responses))
	qualityMetrics["response_time_avg"] /= float64(len(responses))

	return &DebateResult{
		SessionID:       fmt.Sprintf("ensemble_%d", time.Now().Unix()),
		Topic:           request.Topic,
		Consensus:       consensus,
		BestResponse:    bestResponse,
		AllResponses:    responses,
		Duration:        time.Duration(qualityMetrics["response_time_avg"]) * time.Millisecond,
		RoundsConducted: 1, // Ensemble is single-round
		FinalScore:      bestScore,
		QualityMetrics:  qualityMetrics,
		Recommendations: i.generateEnsembleRecommendations(consensus, responses),
	}
}

// generateEnsembleRecommendations generates recommendations based on ensemble results
func (i *AIDebateIntegration) generateEnsembleRecommendations(consensus *ConsensusResult, responses []ParticipantResponse) []string {
	var recommendations []string

	if consensus.Reached {
		recommendations = append(recommendations, "Ensemble consensus successfully achieved.")
	} else {
		recommendations = append(recommendations, "Consider adjusting consensus threshold or providing more context.")
	}

	if len(responses) < 3 {
		recommendations = append(recommendations, "Consider adding more participants for diverse perspectives.")
	}

	if consensus.QualityScore < 0.7 {
		recommendations = append(recommendations, "Response quality could be improved. Consider adjusting participant parameters.")
	}

	return recommendations
}

// GetProviderCapabilities returns capabilities of all configured providers
func (i *AIDebateIntegration) GetProviderCapabilities() map[string]*models.ProviderCapabilities {
	i.mu.RLock()
	defer i.mu.RUnlock()

	capabilities := make(map[string]*models.ProviderCapabilities)

	for key, provider := range i.providers {
		caps := provider.GetCapabilities()
		capabilities[key] = caps
	}

	return capabilities
}

// GetProviderHealth returns health status of all providers
func (i *AIDebateIntegration) GetProviderHealth() map[string]string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	health := make(map[string]string)

	for key, provider := range i.providers {
		err := provider.HealthCheck()
		if err != nil {
			health[key] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			health[key] = "healthy"
		}
	}

	return health
}

// UpdateConfiguration updates the debate configuration
func (i *AIDebateIntegration) UpdateConfiguration(newConfig *config.AIDebateConfig) error {
	if newConfig == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Clear existing providers
	i.providers = make(map[string]llm.LLMProvider)

	// Update configuration
	i.config = newConfig

	// Reinitialize providers
	if err := i.initializeProviders(); err != nil {
		return fmt.Errorf("failed to reinitialize providers: %w", err)
	}

	i.logger.Info("AI debate configuration updated successfully")
	return nil
}

// Provider implementations (simplified versions that integrate with existing LLM system)

type ClaudeProvider struct {
	apiKey     string
	model      string
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

func (p *ClaudeProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	// Integrate with existing Claude implementation
	// This would call the actual Claude provider from the existing system
	return &models.LLMResponse{
		Content:    fmt.Sprintf("Claude response to: %s", req.Prompt),
		Confidence: 0.85,
		TokensUsed: 150,
		Metadata:   map[string]interface{}{"provider": "claude", "model": p.model},
	}, nil
}

func (p *ClaudeProvider) HealthCheck() error {
	// Implement health check logic
	return nil
}

func (p *ClaudeProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{p.model},
		SupportedFeatures:       []string{"text_generation", "reasoning", "analysis"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: false,
		Limits: models.ModelLimits{
			MaxTokens:       4096,
			MaxInputLength:  100000,
			MaxOutputLength: 4096,
		},
	}
}

func (p *ClaudeProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	// Implement configuration validation
	return true, []string{}
}

// Similar implementations for other providers...

type DeepSeekProvider struct {
	apiKey     string
	model      string
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

func (p *DeepSeekProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	return &models.LLMResponse{
		Content:    fmt.Sprintf("DeepSeek response to: %s", req.Prompt),
		Confidence: 0.80,
		TokensUsed: 140,
		Metadata:   map[string]interface{}{"provider": "deepseek", "model": p.model},
	}, nil
}

func (p *DeepSeekProvider) HealthCheck() error {
	return nil
}

func (p *DeepSeekProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{p.model},
		SupportedFeatures:       []string{"text_generation", "code_generation", "analysis"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: false,
		Limits: models.ModelLimits{
			MaxTokens:       4096,
			MaxInputLength:  100000,
			MaxOutputLength: 4096,
		},
	}
}

func (p *DeepSeekProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, []string{}
}

// Placeholder implementations for other providers
type GeminiProvider struct {
	apiKey, model, baseURL string
	timeout                time.Duration
	maxRetries             int
}
type QwenProvider struct {
	apiKey, model, baseURL string
	timeout                time.Duration
	maxRetries             int
}
type ZaiProvider struct {
	apiKey, model, baseURL string
	timeout                time.Duration
	maxRetries             int
}
type OllamaProvider struct {
	baseURL, model string
	timeout        time.Duration
}
type OpenRouterProvider struct {
	apiKey, model, baseURL string
	timeout                time.Duration
	maxRetries             int
}

// Implement similar methods for all provider types...

// DebateRequest represents a debate request
type DebateRequest struct {
	Topic              string
	Context            string
	MaxRounds          int
	ConsensusThreshold float64
	Timeout            time.Duration
}

// DebateResult represents the result of a debate
type DebateResult struct {
	SessionID       string
	Topic           string
	Consensus       *ConsensusResult
	BestResponse    ParticipantResponse
	AllResponses    []ParticipantResponse
	Duration        time.Duration
	RoundsConducted int
	FinalScore      float64
	QualityMetrics  map[string]float64
	Recommendations []string
	FallbackUsed    bool
	CogneeEnhanced  bool
	CogneeInsights  *CogneeInsights
	MemoryUsed      bool
}

// ParticipantResponse represents a participant's response
type ParticipantResponse struct {
	ParticipantName string
	LLMName         string
	Content         string
	Confidence      float64
	QualityScore    float64
	ResponseTime    time.Duration
	RoundNumber     int
	Timestamp       time.Time
	CogneeEnhanced  bool
	Metadata        map[string]interface{}
}

// ConsensusResult represents consensus analysis
type ConsensusResult struct {
	Reached         bool
	ConsensusLevel  float64
	AgreementScore  float64
	QualityScore    float64
	Summary         string
	KeyPoints       []string
	Disagreements   []string
	Recommendations []string
	CogneeInsights  *CogneeInsights
}

// CogneeInsights represents insights from Cognee AI
type CogneeInsights struct {
	SentimentAnalysis map[string]float64
	EntityExtraction  []string
	TopicModeling     []string
	CoherenceScore    float64
	RelevanceScore    float64
	InnovationScore   float64
	Summary           string
}
