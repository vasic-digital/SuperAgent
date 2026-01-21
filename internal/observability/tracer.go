// Package observability provides OpenTelemetry-based tracing and metrics for LLM operations.
// It follows OpenTelemetry semantic conventions for GenAI and LLM systems.
package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// LLM Semantic Convention attributes (OpenTelemetry GenAI conventions)
const (
	// Core LLM attributes
	AttrLLMSystem        = "gen_ai.system"
	AttrLLMProvider      = "gen_ai.request.model"
	AttrLLMModel         = "gen_ai.request.model"
	AttrLLMTemperature   = "gen_ai.request.temperature"
	AttrLLMMaxTokens = "gen_ai.request.max_tokens" // #nosec G101 - OpenTelemetry attribute name, not credentials
	AttrLLMTopP          = "gen_ai.request.top_p"
	AttrLLMStopSequences = "gen_ai.request.stop_sequences"

	// Token usage - OpenTelemetry attribute names for token counting metrics, not credentials
	AttrLLMInputTokens  = "gen_ai.usage.input_tokens"  // #nosec G101
	AttrLLMOutputTokens = "gen_ai.usage.output_tokens" // #nosec G101
	AttrLLMTotalTokens  = "gen_ai.usage.total_tokens"  // #nosec G101

	// Response attributes
	AttrLLMFinishReason = "gen_ai.response.finish_reason"
	AttrLLMResponseID   = "gen_ai.response.id"

	// Content (optional, may be sensitive)
	AttrLLMSystemPrompt = "gen_ai.prompt.system"
	AttrLLMUserPrompt   = "gen_ai.prompt.user"
	AttrLLMAssistant    = "gen_ai.completion"

	// HelixAgent-specific
	AttrHelixRequestID   = "helix.request.id"
	AttrHelixSessionID   = "helix.session.id"
	AttrHelixUserID      = "helix.user.id"
	AttrHelixEnsemble    = "helix.ensemble.enabled"
	AttrHelixDebate      = "helix.debate.enabled"
	AttrHelixCacheHit    = "helix.cache.hit"
	AttrHelixProviderScore = "helix.provider.score"
)

// TracerConfig configures the LLM tracer
type TracerConfig struct {
	ServiceName        string
	ServiceVersion     string
	Environment        string
	EnableContentTrace bool // Whether to include prompt/response content in traces
	SampleRate         float64
	ExporterEndpoint   string
	ExporterType       ExporterType
}

// ExporterType defines the type of trace exporter
type ExporterType string

const (
	ExporterOTLP    ExporterType = "otlp"
	ExporterJaeger  ExporterType = "jaeger"
	ExporterZipkin  ExporterType = "zipkin"
	ExporterConsole ExporterType = "console"
	ExporterNone    ExporterType = "none"
)

// DefaultTracerConfig returns default configuration
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:        "helixagent",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		EnableContentTrace: false, // Disabled by default for privacy
		SampleRate:         1.0,
		ExporterType:       ExporterNone,
	}
}

// LLMTracer provides tracing for LLM operations
type LLMTracer struct {
	tracer      trace.Tracer
	meter       metric.Meter
	config      *TracerConfig
	mu          sync.RWMutex
	initialized bool

	// Metrics
	requestCounter    metric.Int64Counter
	tokenCounter      metric.Int64Counter
	latencyHistogram  metric.Float64Histogram
	errorCounter      metric.Int64Counter
	cacheHitCounter   metric.Int64Counter
	costCounter       metric.Float64Counter
}

// NewLLMTracer creates a new LLM tracer
func NewLLMTracer(config *TracerConfig) (*LLMTracer, error) {
	if config == nil {
		config = DefaultTracerConfig()
	}

	tracer := otel.Tracer(
		config.ServiceName,
		trace.WithInstrumentationVersion(config.ServiceVersion),
	)

	meter := otel.Meter(
		config.ServiceName,
		metric.WithInstrumentationVersion(config.ServiceVersion),
	)

	t := &LLMTracer{
		tracer: tracer,
		meter:  meter,
		config: config,
	}

	if err := t.initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	t.initialized = true
	return t, nil
}

func (t *LLMTracer) initMetrics() error {
	var err error

	t.requestCounter, err = t.meter.Int64Counter(
		"llm.requests.total",
		metric.WithDescription("Total number of LLM requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	t.tokenCounter, err = t.meter.Int64Counter(
		"llm.tokens.total",
		metric.WithDescription("Total tokens processed"),
		metric.WithUnit("{token}"),
	)
	if err != nil {
		return err
	}

	t.latencyHistogram, err = t.meter.Float64Histogram(
		"llm.request.duration",
		metric.WithDescription("LLM request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	t.errorCounter, err = t.meter.Int64Counter(
		"llm.errors.total",
		metric.WithDescription("Total LLM errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	t.cacheHitCounter, err = t.meter.Int64Counter(
		"llm.cache.hits",
		metric.WithDescription("Cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return err
	}

	t.costCounter, err = t.meter.Float64Counter(
		"llm.cost.total",
		metric.WithDescription("Total LLM cost"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		return err
	}

	return nil
}

// LLMRequestParams contains parameters for tracing an LLM request
type LLMRequestParams struct {
	Provider      string
	Model         string
	RequestID     string
	SessionID     string
	UserID        string
	Temperature   float64
	MaxTokens     int
	TopP          float64
	StopSequences []string
	SystemPrompt  string
	UserPrompt    string
	Ensemble      bool
	Debate        bool
}

// LLMResponseParams contains parameters for completing an LLM trace
type LLMResponseParams struct {
	InputTokens   int
	OutputTokens  int
	FinishReason  string
	ResponseID    string
	Content       string
	CacheHit      bool
	ProviderScore float64
	CostUSD       float64
	Error         error
}

// StartLLMRequest starts a new span for an LLM request
func (t *LLMTracer) StartLLMRequest(ctx context.Context, params *LLMRequestParams) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String(AttrLLMSystem, params.Provider),
		attribute.String(AttrLLMProvider, params.Provider),
		attribute.String(AttrLLMModel, params.Model),
		attribute.String(AttrHelixRequestID, params.RequestID),
	}

	if params.SessionID != "" {
		attrs = append(attrs, attribute.String(AttrHelixSessionID, params.SessionID))
	}
	if params.UserID != "" {
		attrs = append(attrs, attribute.String(AttrHelixUserID, params.UserID))
	}
	if params.Temperature > 0 {
		attrs = append(attrs, attribute.Float64(AttrLLMTemperature, params.Temperature))
	}
	if params.MaxTokens > 0 {
		attrs = append(attrs, attribute.Int(AttrLLMMaxTokens, params.MaxTokens))
	}
	if params.TopP > 0 {
		attrs = append(attrs, attribute.Float64(AttrLLMTopP, params.TopP))
	}
	if len(params.StopSequences) > 0 {
		attrs = append(attrs, attribute.StringSlice(AttrLLMStopSequences, params.StopSequences))
	}

	attrs = append(attrs, attribute.Bool(AttrHelixEnsemble, params.Ensemble))
	attrs = append(attrs, attribute.Bool(AttrHelixDebate, params.Debate))

	// Optionally include content (may be sensitive)
	if t.config.EnableContentTrace {
		if params.SystemPrompt != "" {
			attrs = append(attrs, attribute.String(AttrLLMSystemPrompt, params.SystemPrompt))
		}
		if params.UserPrompt != "" {
			attrs = append(attrs, attribute.String(AttrLLMUserPrompt, params.UserPrompt))
		}
	}

	ctx, span := t.tracer.Start(ctx, "llm.completion",
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	// Record request metric
	t.requestCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", params.Provider),
		attribute.String("model", params.Model),
	))

	return ctx, span
}

// EndLLMRequest completes an LLM request span with response data
func (t *LLMTracer) EndLLMRequest(ctx context.Context, span trace.Span, params *LLMResponseParams, startTime time.Time) {
	duration := time.Since(startTime).Seconds()

	attrs := []attribute.KeyValue{
		attribute.Int(AttrLLMInputTokens, params.InputTokens),
		attribute.Int(AttrLLMOutputTokens, params.OutputTokens),
		attribute.Int(AttrLLMTotalTokens, params.InputTokens+params.OutputTokens),
		attribute.String(AttrLLMFinishReason, params.FinishReason),
		attribute.Bool(AttrHelixCacheHit, params.CacheHit),
	}

	if params.ResponseID != "" {
		attrs = append(attrs, attribute.String(AttrLLMResponseID, params.ResponseID))
	}
	if params.ProviderScore > 0 {
		attrs = append(attrs, attribute.Float64(AttrHelixProviderScore, params.ProviderScore))
	}

	// Optionally include response content
	if t.config.EnableContentTrace && params.Content != "" {
		attrs = append(attrs, attribute.String(AttrLLMAssistant, params.Content))
	}

	span.SetAttributes(attrs...)

	// Record metrics
	providerAttr := attribute.String("provider", span.SpanContext().TraceID().String())

	t.tokenCounter.Add(ctx, int64(params.InputTokens+params.OutputTokens), metric.WithAttributes(providerAttr))
	t.latencyHistogram.Record(ctx, duration, metric.WithAttributes(providerAttr))

	if params.CacheHit {
		t.cacheHitCounter.Add(ctx, 1, metric.WithAttributes(providerAttr))
	}

	if params.CostUSD > 0 {
		t.costCounter.Add(ctx, params.CostUSD, metric.WithAttributes(providerAttr))
	}

	if params.Error != nil {
		span.RecordError(params.Error)
		span.SetStatus(codes.Error, params.Error.Error())
		t.errorCounter.Add(ctx, 1, metric.WithAttributes(providerAttr))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}

// StartEnsembleRequest starts a span for ensemble operations
func (t *LLMTracer) StartEnsembleRequest(ctx context.Context, strategy string, providerCount int) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "llm.ensemble",
		trace.WithAttributes(
			attribute.String("ensemble.strategy", strategy),
			attribute.Int("ensemble.provider_count", providerCount),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartDebateRound starts a span for a debate round
func (t *LLMTracer) StartDebateRound(ctx context.Context, round int, topic string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "llm.debate.round",
		trace.WithAttributes(
			attribute.Int("debate.round", round),
			attribute.String("debate.topic", topic),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// StartRAGRetrieval starts a span for RAG retrieval
func (t *LLMTracer) StartRAGRetrieval(ctx context.Context, query string, topK int) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "rag.retrieval",
		trace.WithAttributes(
			attribute.Int("rag.top_k", topK),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// RecordRAGRetrievalResult records RAG retrieval results
func (t *LLMTracer) RecordRAGRetrievalResult(span trace.Span, resultCount int, latencyMs float64) {
	span.SetAttributes(
		attribute.Int("rag.result_count", resultCount),
		attribute.Float64("rag.latency_ms", latencyMs),
	)
}

// StartToolExecution starts a span for tool execution
func (t *LLMTracer) StartToolExecution(ctx context.Context, toolName string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "tool.execution",
		trace.WithAttributes(
			attribute.String("tool.name", toolName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
}

// Global tracer instance
var (
	globalTracer *LLMTracer
	tracerOnce   sync.Once
)

// InitGlobalTracer initializes the global tracer
func InitGlobalTracer(config *TracerConfig) error {
	var initErr error
	tracerOnce.Do(func() {
		globalTracer, initErr = NewLLMTracer(config)
	})
	return initErr
}

// GetTracer returns the global tracer
func GetTracer() *LLMTracer {
	if globalTracer == nil {
		// Initialize with defaults if not set
		globalTracer, _ = NewLLMTracer(nil)
	}
	return globalTracer
}
