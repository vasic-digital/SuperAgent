// Package observabilityadapter bridges internal/observability with the main
// application. It provides lazy-initialized tracing, metrics, and LLM
// middleware through a single ObservabilityAdapter that can be wired into
// the HelixAgent startup sequence.
package observabilityadapter

import (
	"context"
	"fmt"
	"sync"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/observability"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config holds the configuration needed to initialize the observability
// subsystem. A nil Config is valid and puts the adapter into no-op mode.
type Config struct {
	// ServiceName identifies this service in traces and metrics.
	ServiceName string
	// ServiceVersion is the running version of the service.
	ServiceVersion string
	// Environment is the deployment environment (e.g. development, production).
	Environment string
	// ExporterType selects the trace export backend.
	ExporterType observability.ExporterType
	// ExporterEndpoint is the collector/backend address (required for OTLP).
	ExporterEndpoint string
	// ExporterHeaders are extra headers sent to the exporter (e.g. auth).
	ExporterHeaders map[string]string
	// ExporterInsecure disables TLS for the exporter connection.
	ExporterInsecure bool
	// EnableContentTrace enables logging of prompt/response content in spans.
	EnableContentTrace bool
	// SampleRate controls the trace sampling ratio (0.0 - 1.0).
	SampleRate float64
}

// DefaultConfig returns a Config suitable for local development.
// Traces are not exported (ExporterNone) and content tracing is off.
func DefaultConfig() *Config {
	return &Config{
		ServiceName:        "helixagent",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		ExporterType:       observability.ExporterNone,
		EnableContentTrace: false,
		SampleRate:         1.0,
	}
}

// ObservabilityAdapter is the central entry point for wiring OpenTelemetry
// tracing and Prometheus metrics into the application. All public methods
// are safe for concurrent use.
type ObservabilityAdapter struct {
	mu   sync.RWMutex
	once sync.Once

	config *Config

	// Core components from internal/observability
	tracer         *observability.LLMTracer
	metrics        *observability.LLMMetrics
	tracerProvider *sdktrace.TracerProvider

	// Extended metrics
	mcpMetrics       *observability.MCPMetrics
	embeddingMetrics *observability.EmbeddingMetrics
	vectorDBMetrics  *observability.VectorDBMetrics
	memoryMetrics    *observability.MemoryMetrics
	streamingMetrics *observability.StreamingMetrics
	protocolMetrics  *observability.ProtocolMetrics

	// Debate tracer
	debateTracer *observability.DebateTracer

	initialized bool
	shutdownDone bool
}

// NewObservabilityAdapter creates a new adapter. The adapter is inert
// until Initialize is called.
func NewObservabilityAdapter() *ObservabilityAdapter {
	return &ObservabilityAdapter{}
}

// Initialize sets up the tracing pipeline, metrics collectors, and LLM
// middleware. A nil config is allowed and switches the adapter to no-op
// mode (all getters still return usable zero-value objects).
//
// Initialize is idempotent -- only the first call takes effect.
func (a *ObservabilityAdapter) Initialize(config *Config) error {
	var initErr error
	a.once.Do(func() {
		initErr = a.doInitialize(config)
	})
	return initErr
}

func (a *ObservabilityAdapter) doInitialize(config *Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if config == nil {
		config = DefaultConfig()
		config.ExporterType = observability.ExporterNone
	}
	a.config = config

	// --- Tracer ---
	tracerCfg := &observability.TracerConfig{
		ServiceName:        config.ServiceName,
		ServiceVersion:     config.ServiceVersion,
		Environment:        config.Environment,
		EnableContentTrace: config.EnableContentTrace,
		SampleRate:         config.SampleRate,
		ExporterType:       config.ExporterType,
		ExporterEndpoint:   config.ExporterEndpoint,
	}

	tracer, err := observability.NewLLMTracer(tracerCfg)
	if err != nil {
		return fmt.Errorf("observability adapter: init tracer: %w", err)
	}
	a.tracer = tracer

	// --- Trace Exporter / Provider ---
	// Exporter setup may fail due to OpenTelemetry semconv schema version
	// conflicts in certain dependency combinations. When that happens we
	// degrade gracefully: tracing, metrics, and middleware still work;
	// only span export is disabled.
	exporterCfg := &observability.ExporterConfig{
		Type:        config.ExporterType,
		Endpoint:    config.ExporterEndpoint,
		Headers:     config.ExporterHeaders,
		Insecure:    config.ExporterInsecure,
		ServiceName: config.ServiceName,
		Environment: config.Environment,
		Version:     config.ServiceVersion,
	}

	tp, exporterErr := observability.SetupTraceExporter(
		context.Background(), exporterCfg,
	)
	if exporterErr == nil {
		a.tracerProvider = tp
	}
	// exporterErr is intentionally not returned -- the adapter remains
	// functional without a trace exporter.

	// --- Metrics ---
	metrics, err := observability.NewLLMMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init metrics: %w", err)
	}
	a.metrics = metrics

	// --- Extended Metrics ---
	mcpM, err := observability.NewMCPMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init MCP metrics: %w", err)
	}
	a.mcpMetrics = mcpM

	embM, err := observability.NewEmbeddingMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init embedding metrics: %w", err)
	}
	a.embeddingMetrics = embM

	vdbM, err := observability.NewVectorDBMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init vectordb metrics: %w", err)
	}
	a.vectorDBMetrics = vdbM

	memM, err := observability.NewMemoryMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init memory metrics: %w", err)
	}
	a.memoryMetrics = memM

	strM, err := observability.NewStreamingMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init streaming metrics: %w", err)
	}
	a.streamingMetrics = strM

	protoM, err := observability.NewProtocolMetrics(config.ServiceName)
	if err != nil {
		return fmt.Errorf("observability adapter: init protocol metrics: %w", err)
	}
	a.protocolMetrics = protoM

	// --- Debate Tracer ---
	a.debateTracer = observability.NewDebateTracer(a.tracer, a.metrics)

	a.initialized = true
	return nil
}

// IsInitialized reports whether Initialize completed successfully.
func (a *ObservabilityAdapter) IsInitialized() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.initialized
}

// Shutdown gracefully flushes and shuts down the trace exporter. It is
// safe to call multiple times; only the first call performs work.
func (a *ObservabilityAdapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.shutdownDone {
		return nil
	}
	a.shutdownDone = true

	if a.tracerProvider != nil {
		if err := observability.ShutdownTraceExporter(ctx, a.tracerProvider); err != nil {
			return fmt.Errorf("observability adapter: shutdown exporter: %w", err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Getters — each returns the underlying component or a safe fallback.
// ---------------------------------------------------------------------------

// GetTracer returns the LLM tracer. If the adapter has not been initialized
// it returns the global default tracer from the observability package.
func (a *ObservabilityAdapter) GetTracer() *observability.LLMTracer {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.tracer != nil {
		return a.tracer
	}
	return observability.GetTracer()
}

// GetMetrics returns the LLM metrics collector. If the adapter has not
// been initialized it returns the global default metrics.
func (a *ObservabilityAdapter) GetMetrics() *observability.LLMMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.metrics != nil {
		return a.metrics
	}
	return observability.GetMetrics()
}

// GetMCPMetrics returns the MCP tool-call metrics collector.
func (a *ObservabilityAdapter) GetMCPMetrics() *observability.MCPMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.mcpMetrics != nil {
		return a.mcpMetrics
	}
	return observability.GetMCPMetrics()
}

// GetEmbeddingMetrics returns the embedding metrics collector.
func (a *ObservabilityAdapter) GetEmbeddingMetrics() *observability.EmbeddingMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.embeddingMetrics != nil {
		return a.embeddingMetrics
	}
	return observability.GetEmbeddingMetrics()
}

// GetVectorDBMetrics returns the vector database metrics collector.
func (a *ObservabilityAdapter) GetVectorDBMetrics() *observability.VectorDBMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.vectorDBMetrics != nil {
		return a.vectorDBMetrics
	}
	return observability.GetVectorDBMetrics()
}

// GetMemoryMetrics returns the memory subsystem metrics collector.
func (a *ObservabilityAdapter) GetMemoryMetrics() *observability.MemoryMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.memoryMetrics != nil {
		return a.memoryMetrics
	}
	return observability.GetMemoryMetrics()
}

// GetStreamingMetrics returns the streaming metrics collector.
func (a *ObservabilityAdapter) GetStreamingMetrics() *observability.StreamingMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.streamingMetrics != nil {
		return a.streamingMetrics
	}
	return observability.GetStreamingMetrics()
}

// GetProtocolMetrics returns the protocol metrics collector.
func (a *ObservabilityAdapter) GetProtocolMetrics() *observability.ProtocolMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.protocolMetrics != nil {
		return a.protocolMetrics
	}
	return observability.GetProtocolMetrics()
}

// GetDebateTracer returns a specialized tracer for AI debate rounds.
func (a *ObservabilityAdapter) GetDebateTracer() *observability.DebateTracer {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.debateTracer != nil {
		return a.debateTracer
	}
	// Build a fallback from global singletons.
	return observability.NewDebateTracer(
		observability.GetTracer(),
		observability.GetMetrics(),
	)
}

// GetTracerProvider returns the underlying OpenTelemetry TracerProvider.
// Returns nil if the adapter was not initialized.
func (a *ObservabilityAdapter) GetTracerProvider() *sdktrace.TracerProvider {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tracerProvider
}

// ---------------------------------------------------------------------------
// LLM Middleware
// ---------------------------------------------------------------------------

// GetLLMMiddleware returns a function that wraps an LLM provider with
// OpenTelemetry tracing. The wrapper records span data for every
// Complete and CompleteStream call.
func (a *ObservabilityAdapter) GetLLMMiddleware() func(provider llm.LLMProvider, name string) llm.LLMProvider {
	tracer := a.GetTracer()
	return func(provider llm.LLMProvider, name string) llm.LLMProvider {
		return observability.NewTracedProvider(provider, tracer, name)
	}
}

// WrapProvider is a convenience helper that wraps a single provider with
// tracing middleware.
func (a *ObservabilityAdapter) WrapProvider(
	provider llm.LLMProvider,
	name string,
) llm.LLMProvider {
	return a.GetLLMMiddleware()(provider, name)
}
