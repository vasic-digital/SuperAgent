package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/models"
)

// ZaiProvider implements LLMProvider for Zai
type ZaiProvider struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	logger     *logrus.Logger
}

// NewZaiProvider creates a new Zai provider
func NewZaiProvider(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) (*ZaiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if baseURL == "" {
		baseURL = "https://api.zai.com"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &ZaiProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

// Complete generates a completion for the given request
func (zp *ZaiProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	zp.logger.Debugf("ZaiProvider.Complete called with model: %s", req.ModelParams.Model)

	return &models.LLMResponse{
		ID:           "zai-" + time.Now().Format("20060102150405"),
		ProviderName: "zai",
		Content:      "This is a mock response from Zai provider",
		Confidence:   0.9,
		TokensUsed:   30,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model": zp.model,
		},
	}, nil
}

// CompleteStream generates a streaming completion
func (zp *ZaiProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	zp.logger.Debugf("ZaiProvider.CompleteStream called with model: %s", req.ModelParams.Model)

	responseChan := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(responseChan)

		response := &models.LLMResponse{
			ID:           "zai-stream-" + time.Now().Format("20060102150405"),
			ProviderName: "zai",
			Content:      "Mock streaming content",
			Confidence:   0.9,
			TokensUsed:   20,
			ResponseTime: time.Since(time.Now()).Milliseconds(),
			FinishReason: "stop",
			Metadata: map[string]interface{}{
				"model": zp.model,
			},
		}

		responseChan <- response
	}()

	return responseChan, nil
}

// HealthCheck performs a health check
func (zp *ZaiProvider) HealthCheck() error {
	return nil
}

// GetCapabilities returns the provider's capabilities
func (zp *ZaiProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{"zai-pro", "zai-turbo"},
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
			MaxConcurrentRequests: 50,
		},
		Metadata: map[string]string{
			"provider": "zai",
		},
	}
}

// ValidateConfig validates the provider configuration
func (zp *ZaiProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		errors = append(errors, "api_key is required")
	}

	if model, ok := config["model"].(string); !ok || model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}
