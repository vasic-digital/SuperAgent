package observability

import (
	"context"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// TracedProvider wraps an LLM provider with OpenTelemetry tracing
type TracedProvider struct {
	provider llm.LLMProvider
	tracer   *LLMTracer
	name     string
}

// NewTracedProvider creates a traced wrapper around an LLM provider
func NewTracedProvider(provider llm.LLMProvider, tracer *LLMTracer, name string) *TracedProvider {
	return &TracedProvider{
		provider: provider,
		tracer:   tracer,
		name:     name,
	}
}

// Complete performs a completion with tracing
func (p *TracedProvider) Complete(ctx context.Context, request *models.LLMRequest) (*models.LLMResponse, error) {
	startTime := time.Now()

	// Start trace
	params := &LLMRequestParams{
		Provider:    p.name,
		Model:       request.ModelParams.Model,
		Temperature: request.ModelParams.Temperature,
		MaxTokens:   request.ModelParams.MaxTokens,
	}

	ctx, span := p.tracer.StartLLMRequest(ctx, params)

	// Call underlying provider
	response, err := p.provider.Complete(ctx, request)

	// End trace
	respParams := &LLMResponseParams{
		Error: err,
	}

	if response != nil {
		respParams.OutputTokens = estimateTokens(response.Content)
		respParams.FinishReason = response.FinishReason
		respParams.ResponseID = response.ID
	}

	p.tracer.EndLLMRequest(ctx, span, respParams, startTime)

	return response, err
}

// CompleteStream performs streaming completion with tracing
func (p *TracedProvider) CompleteStream(ctx context.Context, request *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	startTime := time.Now()

	params := &LLMRequestParams{
		Provider:    p.name,
		Model:       request.ModelParams.Model,
		Temperature: request.ModelParams.Temperature,
		MaxTokens:   request.ModelParams.MaxTokens,
	}

	ctx, span := p.tracer.StartLLMRequest(ctx, params)

	// Call underlying provider
	chunks, err := p.provider.CompleteStream(ctx, request)
	if err != nil {
		p.tracer.EndLLMRequest(ctx, span, &LLMResponseParams{Error: err}, startTime)
		return nil, err
	}

	// Wrap the channel to track completion
	tracedChunks := make(chan *models.LLMResponse, 100)
	go func() {
		defer close(tracedChunks)

		var totalContent string
		var finishReason string

		for chunk := range chunks {
			tracedChunks <- chunk
			if chunk.Content != "" {
				totalContent += chunk.Content
			}
			if chunk.FinishReason != "" {
				finishReason = chunk.FinishReason
			}
		}

		// End trace when stream completes
		p.tracer.EndLLMRequest(ctx, span, &LLMResponseParams{
			OutputTokens: estimateTokens(totalContent),
			FinishReason: finishReason,
		}, startTime)
	}()

	return tracedChunks, nil
}

// GetCapabilities returns provider capabilities
func (p *TracedProvider) GetCapabilities() *models.ProviderCapabilities {
	return p.provider.GetCapabilities()
}

// HealthCheck performs a health check with tracing
func (p *TracedProvider) HealthCheck() error {
	return p.provider.HealthCheck()
}

// ValidateConfig validates the provider configuration
func (p *TracedProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return p.provider.ValidateConfig(config)
}

// estimateTokens provides a rough token estimate (4 chars per token)
func estimateTokens(text string) int {
	return len(text) / 4
}

// TracedProviderRegistry wraps a provider registry with tracing
type TracedProviderRegistry struct {
	registry ProviderRegistry
	tracer   *LLMTracer
}

// ProviderRegistry interface for the provider registry
type ProviderRegistry interface {
	GetProvider(name string) llm.LLMProvider
	GetProviderByModel(model string) llm.LLMProvider
	GetHealthyProviders() []llm.LLMProvider
	ListProviders() []string
}

// NewTracedProviderRegistry creates a traced provider registry
func NewTracedProviderRegistry(registry ProviderRegistry, tracer *LLMTracer) *TracedProviderRegistry {
	return &TracedProviderRegistry{
		registry: registry,
		tracer:   tracer,
	}
}

// GetProvider returns a traced provider by name
func (r *TracedProviderRegistry) GetProvider(name string) llm.LLMProvider {
	provider := r.registry.GetProvider(name)
	if provider == nil {
		return nil
	}
	return NewTracedProvider(provider, r.tracer, name)
}

// GetProviderByModel returns a traced provider for a model
func (r *TracedProviderRegistry) GetProviderByModel(model string) llm.LLMProvider {
	provider := r.registry.GetProviderByModel(model)
	if provider == nil {
		return nil
	}
	return NewTracedProvider(provider, r.tracer, "model-"+model)
}

// GetHealthyProviders returns traced healthy providers
func (r *TracedProviderRegistry) GetHealthyProviders() []llm.LLMProvider {
	providers := r.registry.GetHealthyProviders()
	traced := make([]llm.LLMProvider, len(providers))
	for i, p := range providers {
		traced[i] = NewTracedProvider(p, r.tracer, r.registry.ListProviders()[i])
	}
	return traced
}

// ListProviders lists provider names
func (r *TracedProviderRegistry) ListProviders() []string {
	return r.registry.ListProviders()
}

// DebateTracer provides specialized tracing for AI debates
type DebateTracer struct {
	tracer  *LLMTracer
	metrics *LLMMetrics
}

// NewDebateTracer creates a new debate tracer
func NewDebateTracer(tracer *LLMTracer, metrics *LLMMetrics) *DebateTracer {
	return &DebateTracer{
		tracer:  tracer,
		metrics: metrics,
	}
}

// TraceDebateRound traces a single debate round
func (dt *DebateTracer) TraceDebateRound(ctx context.Context, debateID string, round int, participants []string) (context.Context, func(responses map[string]string, consensusReached bool)) {
	startTime := time.Now()

	params := &LLMRequestParams{
		Provider: "debate",
		Model:    "ensemble",
	}

	ctx, span := dt.tracer.StartLLMRequest(ctx, params)

	return ctx, func(responses map[string]string, consensusReached bool) {
		respParams := &LLMResponseParams{}

		// Count total tokens
		totalTokens := 0
		for _, resp := range responses {
			totalTokens += estimateTokens(resp)
		}
		respParams.OutputTokens = totalTokens

		dt.tracer.EndLLMRequest(ctx, span, respParams, startTime)

		// Record debate metrics
		if dt.metrics != nil {
			consensusScore := 0.0
			if consensusReached {
				consensusScore = 1.0
			}
			dt.metrics.RecordDebateRound(ctx, len(participants), consensusScore)
		}
	}
}

// TraceDebateComplete traces a complete debate
func (dt *DebateTracer) TraceDebateComplete(ctx context.Context, debateID string, topic string) (context.Context, func(result string, rounds int, participants []string)) {
	startTime := time.Now()

	params := &LLMRequestParams{
		Provider: "debate",
		Model:    "ensemble",
	}

	ctx, span := dt.tracer.StartLLMRequest(ctx, params)

	return ctx, func(result string, rounds int, participants []string) {
		respParams := &LLMResponseParams{
			OutputTokens: estimateTokens(result),
		}

		dt.tracer.EndLLMRequest(ctx, span, respParams, startTime)

		// Record metrics for the final debate round
		if dt.metrics != nil {
			dt.metrics.RecordDebateRound(ctx, len(participants), 1.0)
		}
		_ = startTime // silence unused warning
	}
}
