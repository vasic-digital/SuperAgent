// Package debate_integration provides the ProviderLLMInvoker that bridges
// HelixAgent's LLM providers with the comprehensive debate system's LLMInvoker interface.
package debate_integration

import (
	"context"
	"fmt"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// ProviderLLMInvoker implements comprehensive.LLMInvoker using a real LLM provider.
// This bridges the gap between HelixAgent's provider system and the debate engine.
type ProviderLLMInvoker struct {
	provider llm.LLMProvider
	model    string
}

// NewProviderLLMInvoker creates an invoker backed by a real LLM provider.
func NewProviderLLMInvoker(provider llm.LLMProvider, model string) *ProviderLLMInvoker {
	return &ProviderLLMInvoker{provider: provider, model: model}
}

// Invoke calls the real LLM provider and returns content, confidence, error.
// This implements the comprehensive.LLMInvoker interface.
func (i *ProviderLLMInvoker) Invoke(
	ctx context.Context,
	systemPrompt, userPrompt string,
) (string, float64, error) {
	if i.provider == nil {
		return "", 0, fmt.Errorf("provider is nil")
	}

	req := &models.LLMRequest{
		ID: fmt.Sprintf("debate-%d", time.Now().UnixNano()),
		Messages: []models.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ModelParams: models.ModelParameters{
			Model:       i.model,
			MaxTokens:   500,
			Temperature: 0.7,
		},
	}

	resp, err := i.provider.Complete(ctx, req)
	if err != nil {
		return "", 0, fmt.Errorf("LLM call failed: %w", err)
	}

	return resp.Content, resp.Confidence, nil
}
