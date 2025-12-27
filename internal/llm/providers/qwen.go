package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/models"
)

// QwenProvider implements LLMProvider for Alibaba Qwen
type QwenProvider struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	logger     *logrus.Logger
}

// NewQwenProvider creates a new Qwen provider
func NewQwenProvider(
	apiKey string,
	baseURL string,
	model string,
	timeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) (*QwenProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &QwenProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

// Complete generates a completion for the given request
func (qp *QwenProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	qp.logger.Debugf("QwenProvider.Complete called with model: %s", req.ModelParams.Model)

	return &models.LLMResponse{
		ID:           "qwen-" + time.Now().Format("20060102150405"),
		ProviderName: "qwen",
		Content:      "This is a mock response from Qwen provider",
		Confidence:   0.9,
		TokensUsed:   30,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model": qp.model,
		},
	}, nil
}

// CompleteStream generates a streaming completion
func (qp *QwenProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	qp.logger.Debugf("QwenProvider.CompleteStream called with model: %s", req.ModelParams.Model)

	responseChan := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(responseChan)

		response := &models.LLMResponse{
			ID:           "qwen-stream-" + time.Now().Format("20060102150405"),
			ProviderName: "qwen",
			Content:      "Mock streaming content",
			Confidence:   0.9,
			TokensUsed:   20,
			ResponseTime: time.Since(time.Now()).Milliseconds(),
			FinishReason: "stop",
			Metadata: map[string]interface{}{
				"model": qp.model,
			},
		}

		responseChan <- response
	}()

	return responseChan, nil
}

// HealthCheck performs a health check
func (qp *QwenProvider) HealthCheck() error {
	return nil
}

// GetCapabilities returns the provider's capabilities
func (qp *QwenProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{"qwen-turbo", "qwen-plus", "qwen-max"},
		SupportedFeatures:       []string{"long_context", "tools"},
		SupportedRequestTypes:   []string{"completion", "streaming"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          true,
		SupportsTools:           true,
		Limits: models.ModelLimits{
			MaxTokens:             32000,
			MaxInputLength:        32000,
			MaxOutputLength:       4096,
			MaxConcurrentRequests: 50,
		},
		Metadata: map[string]string{
			"provider": "alibaba",
		},
	}
}

// ValidateConfig validates the provider configuration
func (qp *QwenProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		errors = append(errors, "api_key is required")
	}

	if model, ok := config["model"].(string); !ok || model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}
