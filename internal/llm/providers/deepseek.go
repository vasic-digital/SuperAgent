package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/models"
)

// DeepSeekProvider implements LLMProvider for DeepSeek
type DeepSeekProvider struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	logger     *logrus.Logger
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) (*DeepSeekProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &DeepSeekProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

// Complete generates a completion for the given request
func (dsp *DeepSeekProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	dsp.logger.Debugf("DeepSeekProvider.Complete called with model: %s", req.ModelParams.Model)

	// TODO: Implement actual API call to DeepSeek
	return &models.LLMResponse{
		ID:           "deepseek-" + time.Now().Format("20060102150405"),
		ProviderName: "deepseek",
		Content:      "This is a mock response from DeepSeek provider",
		Confidence:   0.9,
		TokensUsed:   30,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model": dsp.model,
		},
	}, nil
}

// CompleteStream generates a streaming completion
func (dsp *DeepSeekProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	dsp.logger.Debugf("DeepSeekProvider.CompleteStream called with model: %s", req.ModelParams.Model)

	responseChan := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(responseChan)

		response := &models.LLMResponse{
			ID:           "deepseek-stream-" + time.Now().Format("20060102150405"),
			ProviderName: "deepseek",
			Content:      "Mock streaming content",
			Confidence:   0.9,
			TokensUsed:   20,
			ResponseTime: time.Since(time.Now()).Milliseconds(),
			FinishReason: "stop",
			Metadata: map[string]interface{}{
				"model": dsp.model,
			},
		}

		responseChan <- response
	}()

	return responseChan, nil
}

// HealthCheck performs a health check
func (dsp *DeepSeekProvider) HealthCheck() error {
	return nil
}

// GetCapabilities returns the provider's capabilities
func (dsp *DeepSeekProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{"deepseek-chat", "deepseek-coder"},
		SupportedFeatures:       []string{"long_context", "tools"},
		SupportedRequestTypes:   []string{"completion", "streaming"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          false,
		SupportsTools:           true,
		Limits: models.ModelLimits{
			MaxTokens:             128000,
			MaxInputLength:        128000,
			MaxOutputLength:       4096,
			MaxConcurrentRequests: 100,
		},
		Metadata: map[string]string{
			"provider": "deepseek",
		},
	}
}

// ValidateConfig validates the provider configuration
func (dsp *DeepSeekProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		errors = append(errors, "api_key is required")
	}

	if model, ok := config["model"].(string); !ok || model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}
