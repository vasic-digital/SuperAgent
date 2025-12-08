package llm

import (
	"github.com/superagent/superagent/internal/models"
	"time"
)

// ClaudeProvider stub
type ClaudeProvider struct{}

func (c *ClaudeProvider) Complete(req *models.LLMRequest) (*models.LLMResponse, error) {
	resp := &models.LLMResponse{
		ID:           "rsp-claude-1",
		RequestID:    req.ID,
		ProviderName: "Claude",
		Content:      "stub completion from Claude",
		Confidence:   0.85,
		TokensUsed:   42,
		ResponseTime: 42,
		FinishReason: "stop",
		Metadata:     map[string]interface{}{},
		CreatedAt:    time.Now(),
	}
	return resp, nil
}

func (c ClaudeProvider) HealthCheck() error { return nil }

func (c ClaudeProvider) GetCapabilities() *ProviderCapabilities {
	return &ProviderCapabilities{
		SupportedModels:         []string{"claude"},
		SupportedFeatures:       []string{"streaming"},
		SupportedRequestTypes:   []string{"code_generation"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: false,
		SupportsVision:          false,
		Metadata:                map[string]string{},
	}
}

func (c ClaudeProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}
