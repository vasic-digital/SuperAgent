// Package llm provides LLM provider abstractions and ensemble orchestration for HelixAgent.
//
// This package is the core of HelixAgent's multi-provider LLM support, enabling
// intelligent aggregation of responses from multiple language models.
//
// # Core Components
//
//   - LLMProvider: Interface that all provider implementations must satisfy
//   - Ensemble: Multi-model response aggregation with voting strategies
//   - CircuitBreaker: Fault tolerance for provider failures
//
// # Supported Providers
//
// The providers/ subdirectory contains implementations for:
//   - Claude (Anthropic)
//   - DeepSeek
//   - Gemini (Google)
//   - Mistral
//   - OpenRouter
//   - Qwen (Alibaba)
//   - ZAI
//   - Zen (OpenCode)
//   - Cerebras
//   - Ollama (deprecated, score: 5.0)
//
// # Provider Interface
//
// All providers implement the LLMProvider interface:
//
//	type LLMProvider interface {
//	    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
//	    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan *StreamChunk, error)
//	    HealthCheck(ctx context.Context) error
//	    GetCapabilities() *ProviderCapabilities
//	    ValidateConfig() error
//	}
//
// # Ensemble Orchestration
//
// The ensemble system aggregates responses from multiple providers using
// configurable voting strategies:
//
//   - Confidence-weighted voting
//   - Majority vote
//   - Best-of-N selection
//
// # Circuit Breaker Pattern
//
// The circuit breaker prevents cascading failures when providers are unhealthy:
//
//   - Closed: Normal operation, requests pass through
//   - Open: Provider is failing, requests are short-circuited
//   - Half-Open: Testing if provider has recovered
//
// # Example Usage
//
//	provider := providers.NewClaude(config)
//	if err := provider.ValidateConfig(); err != nil {
//	    log.Fatal(err)
//	}
//
//	request := &llm.CompletionRequest{
//	    Prompt: "Hello, world!",
//	    MaxTokens: 100,
//	}
//
//	response, err := provider.Complete(ctx, request)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(response.Text)
//
// # Configuration
//
// Providers are configured via environment variables:
//
//	CLAUDE_API_KEY      - Anthropic Claude API key
//	DEEPSEEK_API_KEY    - DeepSeek API key
//	GEMINI_API_KEY      - Google Gemini API key
//	MISTRAL_API_KEY     - Mistral API key
//	OPENROUTER_API_KEY  - OpenRouter API key
//	QWEN_API_KEY        - Qwen/DashScope API key
//	ZAI_API_KEY         - ZAI API key
//	OPENCODE_API_KEY    - Zen (OpenCode) API key
//	CEREBRAS_API_KEY    - Cerebras API key
//
// See the providers/ subdirectory for provider-specific documentation.
package llm
