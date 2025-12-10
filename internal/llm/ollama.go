package llm

import (
	"context"

	"github.com/superagent/superagent/internal/llm/providers/ollama"
	"github.com/superagent/superagent/internal/models"
)

// OllamaProvider wraps the complete Ollama provider implementation
type OllamaProvider struct {
	provider *ollama.OllamaProvider
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	return &OllamaProvider{
		provider: ollama.NewOllamaProvider(baseURL, model),
	}
}

func (o *OllamaProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	return o.provider.Complete(context.Background(), req)
}

func (o *OllamaProvider) HealthCheck() error {
	return o.provider.HealthCheck()
}

func (o *OllamaProvider) GetCapabilities() *ProviderCapabilities {
	caps := o.provider.GetCapabilities()
	return &ProviderCapabilities{
		SupportedModels:         caps.SupportedModels,
		SupportedFeatures:       caps.SupportedFeatures,
		SupportedRequestTypes:   caps.SupportedRequestTypes,
		SupportsStreaming:       caps.SupportsStreaming,
		SupportsFunctionCalling: caps.SupportsFunctionCalling,
		SupportsVision:          caps.SupportsVision,
		Metadata:                caps.Metadata,
	}
}

func (o *OllamaProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return o.provider.ValidateConfig(config)
}
