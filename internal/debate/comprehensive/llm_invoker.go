package comprehensive

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// ProviderLLMInvoker implements LLMInvoker using a real LLM provider.
type ProviderLLMInvoker struct {
	provider llm.LLMProvider
	model    string
}

// NewProviderLLMInvoker creates an invoker backed by a real LLM provider.
func NewProviderLLMInvoker(provider llm.LLMProvider, model string) *ProviderLLMInvoker {
	return &ProviderLLMInvoker{provider: provider, model: model}
}

// Invoke calls the real LLM provider and returns the response content,
// confidence score, and any error.
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

// InvokerRegistry maps provider/model keys to LLMInvoker instances.
type InvokerRegistry struct {
	mu       sync.RWMutex
	invokers map[string]LLMInvoker // key: "provider/model"
	fallback LLMInvoker            // default invoker if no specific one found
}

// globalInvokerRegistry is the package-level registry used by agents
// when no invoker is set directly on the BaseAgent.
var globalInvokerRegistry = &InvokerRegistry{
	invokers: make(map[string]LLMInvoker),
}

// RegisterInvoker registers an LLMInvoker for a specific provider/model pair.
func RegisterInvoker(provider, model string, invoker LLMInvoker) {
	globalInvokerRegistry.mu.Lock()
	defer globalInvokerRegistry.mu.Unlock()
	key := provider + "/" + model
	globalInvokerRegistry.invokers[key] = invoker
}

// SetFallbackInvoker sets a fallback invoker used when no provider-specific
// invoker is found.
func SetFallbackInvoker(invoker LLMInvoker) {
	globalInvokerRegistry.mu.Lock()
	defer globalInvokerRegistry.mu.Unlock()
	globalInvokerRegistry.fallback = invoker
}

// GetInvoker retrieves the invoker for a provider/model pair, falling back
// to the default invoker if none is registered.
func GetInvoker(provider, model string) LLMInvoker {
	globalInvokerRegistry.mu.RLock()
	defer globalInvokerRegistry.mu.RUnlock()
	key := provider + "/" + model
	if inv, ok := globalInvokerRegistry.invokers[key]; ok {
		return inv
	}
	return globalInvokerRegistry.fallback
}

// ClearInvokerRegistry removes all registered invokers. Primarily useful
// for testing.
func ClearInvokerRegistry() {
	globalInvokerRegistry.mu.Lock()
	defer globalInvokerRegistry.mu.Unlock()
	globalInvokerRegistry.invokers = make(map[string]LLMInvoker)
	globalInvokerRegistry.fallback = nil
}
