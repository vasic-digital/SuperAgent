package llm

import (
	"github.com/superagent/superagent/internal/models"
	"time"
)

// DeepSeekProvider implemented in the llm package for MVP ensemble
type DeepSeekProvider struct{}

func (d *DeepSeekProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	resp := &models.LLMResponse{
		ID:           "rsp-deepseek-1",
		RequestID:    req.ID,
		ProviderName: "DeepSeek",
		Content:      "stub completion from DeepSeek",
		Confidence:   0.9,
		TokensUsed:   10,
		ResponseTime: 10,
		FinishReason: "stop",
		Metadata:     map[string]interface{}{},
		CreatedAt:    time.Now(),
	}
	return resp, nil
}

func (d *DeepSeekProvider) HealthCheck() error { return nil }

func (d *DeepSeekProvider) GetCapabilities() *ProviderCapabilities {
	return &ProviderCapabilities{
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
