package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/models"
)

// OllamaProvider implements LLMProvider for Ollama
type OllamaProvider struct {
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	logger     *logrus.Logger
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(
	baseURL string,
	model string,
	timeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) (*OllamaProvider, error) {
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &OllamaProvider{
		baseURL:    baseURL,
		model:      model,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

// Complete generates a completion for the given request
func (op *OllamaProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	op.logger.Debugf("OllamaProvider.Complete called with model: %s", req.ModelParams.Model)

	return &models.LLMResponse{
		ID:           "ollama-" + time.Now().Format("20060102150405"),
		ProviderName: "ollama",
		Content:      "This is a mock response from Ollama provider",
		Confidence:   0.9,
		TokensUsed:   30,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model": op.model,
		},
	}, nil
}

// CompleteStream generates a streaming completion
func (op *OllamaProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	op.logger.Debugf("OllamaProvider.CompleteStream called with model: %s", req.ModelParams.Model)

	responseChan := make(chan *models.LLMResponse, 1)

	go func() {
		defer close(responseChan)

		response := &models.LLMResponse{
			ID:           "ollama-stream-" + time.Now().Format("20060102150405"),
			ProviderName: "ollama",
			Content:      "Mock streaming content",
			Confidence:   0.9,
			TokensUsed:   20,
			ResponseTime: time.Since(time.Now()).Milliseconds(),
			FinishReason: "stop",
			Metadata: map[string]interface{}{
				"model": op.model,
			},
		}

		responseChan <- response
	}()

	return responseChan, nil
}

// HealthCheck performs a health check
func (op *OllamaProvider) HealthCheck() error {
	return nil
}

// GetCapabilities returns the provider's capabilities
func (op *OllamaProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels:         []string{"llama2", "llama3", "mistral", "codellama"},
		SupportedFeatures:       []string{"local", "tools"},
		SupportedRequestTypes:   []string{"completion", "streaming"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          false,
		SupportsTools:           true,
		Limits: models.ModelLimits{
			MaxTokens:             32000,
			MaxInputLength:        32000,
			MaxOutputLength:       4096,
			MaxConcurrentRequests: 10,
		},
		Metadata: map[string]string{
			"provider": "ollama",
		},
	}
}

// ValidateConfig validates the provider configuration
func (op *OllamaProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if model, ok := config["model"].(string); !ok || model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}
