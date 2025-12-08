package deepseek

import (
	"github.com/superagent/superagent/internal/llm"
	"github.com/superagent/superagent/internal/models"
)

// DeepSeekProvider is a minimal stub implementing the LLMProvider interface.
type DeepSeekProvider struct{}

func (d *DeepSeekProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	return nil, nil
}

func (d *DeepSeekProvider) HealthCheck() error { return nil }

func (d *DeepSeekProvider) GetCapabilities() *llm.ProviderCapabilities {
	return &llm.ProviderCapabilities{
		SupportedModels:         []string{"deepseek"},
		SupportedFeatures:       []string{"streaming"},
		SupportedRequestTypes:   []string{"code_generation"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: false,
		SupportsVision:          false,
		Metadata:                map[string]string{},
	}
}

func (d *DeepSeekProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}
