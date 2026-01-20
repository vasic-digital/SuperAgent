package observability

import (
	"context"
	"time"

	"dev.helix.agent/internal/llm"
)

// TracedProvider wraps an LLM provider with OpenTelemetry tracing
type TracedProvider struct {
	provider llm.LLMProvider
	tracer   *LLMTracer
	name     string
}

// NewTracedProvider creates a traced wrapper around an LLM provider
func NewTracedProvider(provider llm.LLMProvider, tracer *LLMTracer) *TracedProvider {
	name := "unknown"
	if provider != nil {
		name = provider.GetName()
	}
	return &TracedProvider{
		provider: provider,
		tracer:   tracer,
		name:     name,
	}
}

// Complete performs a completion with tracing
func (p *TracedProvider) Complete(ctx context.Context, request *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	startTime := time.Now()

	// Start trace
	params := &LLMRequestParams{
		Provider:    p.name,
		Model:       request.Model,
		Temperature: request.Temperature,
		MaxTokens:   request.MaxTokens,
	}

	// Count messages
	for _, msg := range request.Messages {
		params.InputTokens += estimateTokens(msg.Content)
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

		// Handle tool calls if present
		if len(response.ToolCalls) > 0 {
			toolNames := make([]string, len(response.ToolCalls))
			for i, tc := range response.ToolCalls {
				toolNames[i] = tc.Function.Name
			}
			respParams.ToolCalls = toolNames
		}
	}

	p.tracer.EndLLMRequest(ctx, span, respParams, startTime)

	return response, err
}

// CompleteStream performs streaming completion with tracing
func (p *TracedProvider) CompleteStream(ctx context.Context, request *llm.CompletionRequest) (<-chan *llm.StreamChunk, error) {
	startTime := time.Now()

	params := &LLMRequestParams{
		Provider:    p.name,
		Model:       request.Model,
		Temperature: request.Temperature,
		MaxTokens:   request.MaxTokens,
		Stream:      true,
	}

	for _, msg := range request.Messages {
		params.InputTokens += estimateTokens(msg.Content)
	}

	ctx, span := p.tracer.StartLLMRequest(ctx, params)

	// Call underlying provider
	chunks, err := p.provider.CompleteStream(ctx, request)
	if err != nil {
		p.tracer.EndLLMRequest(ctx, span, &LLMResponseParams{Error: err}, startTime)
		return nil, err
	}

	// Wrap the channel to track completion
	tracedChunks := make(chan *llm.StreamChunk, 100)
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

// GetName returns the provider name
func (p *TracedProvider) GetName() string {
	return p.provider.GetName()
}

// GetCapabilities returns provider capabilities
func (p *TracedProvider) GetCapabilities() *llm.ProviderCapabilities {
	return p.provider.GetCapabilities()
}

// HealthCheck performs a health check with tracing
func (p *TracedProvider) HealthCheck(ctx context.Context) error {
	return p.provider.HealthCheck(ctx)
}

// ValidateConfig validates the provider configuration
func (p *TracedProvider) ValidateConfig() error {
	return p.provider.ValidateConfig()
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
	return NewTracedProvider(provider, r.tracer)
}

// GetProviderByModel returns a traced provider for a model
func (r *TracedProviderRegistry) GetProviderByModel(model string) llm.LLMProvider {
	provider := r.registry.GetProviderByModel(model)
	if provider == nil {
		return nil
	}
	return NewTracedProvider(provider, r.tracer)
}

// GetHealthyProviders returns traced healthy providers
func (r *TracedProviderRegistry) GetHealthyProviders() []llm.LLMProvider {
	providers := r.registry.GetHealthyProviders()
	traced := make([]llm.LLMProvider, len(providers))
	for i, p := range providers {
		traced[i] = NewTracedProvider(p, r.tracer)
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
	params.Metadata = map[string]interface{}{
		"debate_id":    debateID,
		"round":        round,
		"participants": participants,
	}

	ctx, span := dt.tracer.StartLLMRequest(ctx, params)

	return ctx, func(responses map[string]string, consensusReached bool) {
		respParams := &LLMResponseParams{}
		respParams.Metadata = map[string]interface{}{
			"responses_count":   len(responses),
			"consensus_reached": consensusReached,
		}

		// Count total tokens
		totalTokens := 0
		for _, resp := range responses {
			totalTokens += estimateTokens(resp)
		}
		respParams.OutputTokens = totalTokens

		dt.tracer.EndLLMRequest(ctx, span, respParams, startTime)

		// Record debate metrics
		if dt.metrics != nil {
			dt.metrics.RecordDebateRound(ctx, debateID, round, len(participants), consensusReached)
		}
	}
}

// TraceDebateComplete traces a complete debate
func (dt *DebateTracer) TraceDebateComplete(ctx context.Context, debateID string, topic string) (context.Context, func(result string, rounds int, participants []string)) {
	startTime := time.Now()

	params := &LLMRequestParams{
		Provider:    "debate",
		Model:       "ensemble",
		InputTokens: estimateTokens(topic),
	}
	params.Metadata = map[string]interface{}{
		"debate_id": debateID,
		"topic":     truncateStr(topic, 100),
	}

	ctx, span := dt.tracer.StartLLMRequest(ctx, params)

	return ctx, func(result string, rounds int, participants []string) {
		respParams := &LLMResponseParams{
			OutputTokens: estimateTokens(result),
		}
		respParams.Metadata = map[string]interface{}{
			"total_rounds":  rounds,
			"participants":  participants,
			"result_length": len(result),
		}

		dt.tracer.EndLLMRequest(ctx, span, respParams, startTime)

		// Record complete debate metrics
		if dt.metrics != nil {
			dt.metrics.RecordDebateComplete(ctx, debateID, rounds, len(participants), time.Since(startTime))
		}
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
