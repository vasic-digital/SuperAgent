package observability

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// LLMMetrics provides comprehensive metrics for LLM operations
type LLMMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	// Request metrics
	RequestsTotal    metric.Int64Counter
	RequestsInFlight metric.Int64UpDownCounter
	RequestDuration  metric.Float64Histogram
	RequestSize      metric.Int64Histogram

	// Token metrics
	InputTokens  metric.Int64Counter
	OutputTokens metric.Int64Counter
	TotalTokens  metric.Int64Counter

	// Cost metrics
	TotalCost      metric.Float64Counter
	CostPerRequest metric.Float64Histogram

	// Error metrics
	ErrorsTotal     metric.Int64Counter
	TimeoutsTotal   metric.Int64Counter
	RateLimitsTotal metric.Int64Counter

	// Cache metrics
	CacheHits    metric.Int64Counter
	CacheMisses  metric.Int64Counter
	CacheLatency metric.Float64Histogram

	// Provider metrics
	ProviderHealth  metric.Int64Gauge
	ProviderScore   metric.Float64Gauge
	ProviderLatency metric.Float64Histogram

	// Debate metrics
	DebateRounds       metric.Int64Counter
	DebateConsensus    metric.Float64Histogram
	DebateParticipants metric.Int64Histogram

	// RAG metrics
	RAGRetrievals     metric.Int64Counter
	RAGResultCount    metric.Int64Histogram
	RAGLatency        metric.Float64Histogram
	RAGRelevanceScore metric.Float64Histogram
}

// NewLLMMetrics creates a new metrics collector
func NewLLMMetrics(serviceName string) (*LLMMetrics, error) {
	meter := otel.Meter(serviceName)
	m := &LLMMetrics{meter: meter}

	var err error

	// Request metrics
	m.RequestsTotal, err = meter.Int64Counter("llm_requests_total",
		metric.WithDescription("Total number of LLM requests"))
	if err != nil {
		return nil, err
	}

	m.RequestsInFlight, err = meter.Int64UpDownCounter("llm_requests_in_flight",
		metric.WithDescription("Number of requests currently being processed"))
	if err != nil {
		return nil, err
	}

	m.RequestDuration, err = meter.Float64Histogram("llm_request_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.RequestSize, err = meter.Int64Histogram("llm_request_size_bytes",
		metric.WithDescription("Request size in bytes"),
		metric.WithUnit("By"))
	if err != nil {
		return nil, err
	}

	// Token metrics
	m.InputTokens, err = meter.Int64Counter("llm_input_tokens_total",
		metric.WithDescription("Total input tokens"))
	if err != nil {
		return nil, err
	}

	m.OutputTokens, err = meter.Int64Counter("llm_output_tokens_total",
		metric.WithDescription("Total output tokens"))
	if err != nil {
		return nil, err
	}

	m.TotalTokens, err = meter.Int64Counter("llm_tokens_total",
		metric.WithDescription("Total tokens (input + output)"))
	if err != nil {
		return nil, err
	}

	// Cost metrics
	m.TotalCost, err = meter.Float64Counter("llm_cost_total_usd",
		metric.WithDescription("Total cost in USD"),
		metric.WithUnit("USD"))
	if err != nil {
		return nil, err
	}

	m.CostPerRequest, err = meter.Float64Histogram("llm_cost_per_request_usd",
		metric.WithDescription("Cost per request in USD"),
		metric.WithUnit("USD"))
	if err != nil {
		return nil, err
	}

	// Error metrics
	m.ErrorsTotal, err = meter.Int64Counter("llm_errors_total",
		metric.WithDescription("Total errors"))
	if err != nil {
		return nil, err
	}

	m.TimeoutsTotal, err = meter.Int64Counter("llm_timeouts_total",
		metric.WithDescription("Total timeouts"))
	if err != nil {
		return nil, err
	}

	m.RateLimitsTotal, err = meter.Int64Counter("llm_rate_limits_total",
		metric.WithDescription("Total rate limit hits"))
	if err != nil {
		return nil, err
	}

	// Cache metrics
	m.CacheHits, err = meter.Int64Counter("llm_cache_hits_total",
		metric.WithDescription("Cache hits"))
	if err != nil {
		return nil, err
	}

	m.CacheMisses, err = meter.Int64Counter("llm_cache_misses_total",
		metric.WithDescription("Cache misses"))
	if err != nil {
		return nil, err
	}

	m.CacheLatency, err = meter.Float64Histogram("llm_cache_latency_seconds",
		metric.WithDescription("Cache operation latency"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	// Provider metrics
	m.ProviderLatency, err = meter.Float64Histogram("llm_provider_latency_seconds",
		metric.WithDescription("Provider response latency"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	// Debate metrics
	m.DebateRounds, err = meter.Int64Counter("llm_debate_rounds_total",
		metric.WithDescription("Total debate rounds"))
	if err != nil {
		return nil, err
	}

	m.DebateConsensus, err = meter.Float64Histogram("llm_debate_consensus_score",
		metric.WithDescription("Debate consensus score (0-1)"))
	if err != nil {
		return nil, err
	}

	m.DebateParticipants, err = meter.Int64Histogram("llm_debate_participants",
		metric.WithDescription("Number of debate participants"))
	if err != nil {
		return nil, err
	}

	// RAG metrics
	m.RAGRetrievals, err = meter.Int64Counter("rag_retrievals_total",
		metric.WithDescription("Total RAG retrievals"))
	if err != nil {
		return nil, err
	}

	m.RAGResultCount, err = meter.Int64Histogram("rag_result_count",
		metric.WithDescription("Number of RAG results"))
	if err != nil {
		return nil, err
	}

	m.RAGLatency, err = meter.Float64Histogram("rag_latency_seconds",
		metric.WithDescription("RAG retrieval latency"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}

	m.RAGRelevanceScore, err = meter.Float64Histogram("rag_relevance_score",
		metric.WithDescription("RAG result relevance score"))
	if err != nil {
		return nil, err
	}

	return m, nil
}

// RecordRequest records a completed LLM request
func (m *LLMMetrics) RecordRequest(ctx context.Context, provider, model string, duration time.Duration, inputTokens, outputTokens int, cost float64, err error) {
	attrs := metric.WithAttributes(
		attribute.String("provider", provider),
		attribute.String("model", model),
	)

	m.RequestsTotal.Add(ctx, 1, attrs)
	m.RequestDuration.Record(ctx, duration.Seconds(), attrs)
	m.InputTokens.Add(ctx, int64(inputTokens), attrs)
	m.OutputTokens.Add(ctx, int64(outputTokens), attrs)
	m.TotalTokens.Add(ctx, int64(inputTokens+outputTokens), attrs)

	if cost > 0 {
		m.TotalCost.Add(ctx, cost, attrs)
		m.CostPerRequest.Record(ctx, cost, attrs)
	}

	if err != nil {
		m.ErrorsTotal.Add(ctx, 1, attrs)
	}
}

// RecordCacheHit records a cache hit
func (m *LLMMetrics) RecordCacheHit(ctx context.Context, latency time.Duration) {
	m.CacheHits.Add(ctx, 1)
	m.CacheLatency.Record(ctx, latency.Seconds())
}

// RecordCacheMiss records a cache miss
func (m *LLMMetrics) RecordCacheMiss(ctx context.Context) {
	m.CacheMisses.Add(ctx, 1)
}

// RecordDebateRound records a debate round
func (m *LLMMetrics) RecordDebateRound(ctx context.Context, participants int, consensusScore float64) {
	m.DebateRounds.Add(ctx, 1)
	m.DebateParticipants.Record(ctx, int64(participants))
	m.DebateConsensus.Record(ctx, consensusScore)
}

// RecordRAGRetrieval records a RAG retrieval operation
func (m *LLMMetrics) RecordRAGRetrieval(ctx context.Context, resultCount int, latency time.Duration, avgRelevance float64) {
	m.RAGRetrievals.Add(ctx, 1)
	m.RAGResultCount.Record(ctx, int64(resultCount))
	m.RAGLatency.Record(ctx, latency.Seconds())
	m.RAGRelevanceScore.Record(ctx, avgRelevance)
}

// Global metrics instance
var (
	globalMetrics *LLMMetrics
	metricsOnce   sync.Once
)

// InitGlobalMetrics initializes the global metrics
func InitGlobalMetrics(serviceName string) error {
	var initErr error
	metricsOnce.Do(func() {
		globalMetrics, initErr = NewLLMMetrics(serviceName)
	})
	return initErr
}

// GetMetrics returns the global metrics
func GetMetrics() *LLMMetrics {
	if globalMetrics == nil {
		globalMetrics, _ = NewLLMMetrics("helixagent")
	}
	return globalMetrics
}
