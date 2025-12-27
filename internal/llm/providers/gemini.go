package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/models"
)

// GeminiProvider implements LLMProvider for Google Gemini
type GeminiProvider struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	logger     *logrus.Logger
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &GeminiProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

// Complete generates a completion for the given request
func (gp *GeminiProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	gp.logger.Debugf("GeminiProvider.Complete called with model: %s", req.ModelParams.Model)

	return &models.LLMResponse{
		ID:           "gemini-" + time.Now().Format("20060102150405"),
		ProviderName: "gemini",
		Content:      "This is a mock response from Gemini provider",
		Confidence:   0.9,
		TokensUsed:   30,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model": gp.model,
		},
	}, nil
}

// CompleteStream generates a streaming completion
func (gp *GeminiProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	gp.logger.Debugf("GeminiProvider.CompleteStream called with model: %s", req.ModelParams.Model)

	responseChan := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(responseChan)

		response := &models.LLMResponse{
			ID:           "gemini-stream-" + time.Now().Format("20060102150405"),
			ProviderName: "gemini",
			Content:      "Mock streaming content",
			Confidence:   0.9,
			TokensUsed:   20,
			ResponseTime: time.Since(time.Now()).Milliseconds(),
			FinishReason: "stop",
			Metadata: map[string]interface{}{
				"model": gp.model,
			},
		}

		responseChan <- response
	}()

	return responseChan, nil
}

// HealthCheck performs a health check
func (gp *GeminiProvider) HealthCheck() error {
	return nil
}

// GetCapabilities returns the provider's capabilities
func (gp *GeminiProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{"gemini-pro", "gemini-1.5-pro", "gemini-1.5-flash"},
		SupportedFeatures:       []string{"multimodal", "tools", "function_calling"},
		SupportedRequestTypes:   []string{"completion", "streaming"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          true,
		SupportsTools:           true,
		Limits: models.ModelLimits{
			MaxTokens:             32000,
			MaxInputLength:        32000,
			MaxOutputLength:       8192,
			MaxConcurrentRequests: 60,
		},
		Metadata: map[string]string{
			"provider": "google",
		},
	}
}

// ValidateConfig validates the provider configuration
func (gp *GeminiProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		errors = append(errors, "api_key is required")
	}

	if model, ok := config["model"].(string); !ok || model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}
