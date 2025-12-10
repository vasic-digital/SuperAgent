package llm

import (
	"context"

	"github.com/superagent/superagent/internal/llm/providers/zai"
	"github.com/superagent/superagent/internal/models"
)

// ZaiProvider wraps the complete Z.AI provider implementation
type ZaiProvider struct {
	provider *zai.ZAIProvider
}

func NewZaiProvider(apiKey, baseURL, model string) *ZaiProvider {
	return &ZaiProvider{
		provider: zai.NewZAIProvider(apiKey, baseURL, model),
	}
}

func (z *ZaiProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	return z.provider.Complete(context.Background(), req)
}

func (z *ZaiProvider) HealthCheck() error {
	return z.provider.HealthCheck()
}

func (z *ZaiProvider) GetCapabilities() *ProviderCapabilities {
	caps := z.provider.GetCapabilities()
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

func (z *ZaiProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return z.provider.ValidateConfig(config)
}
