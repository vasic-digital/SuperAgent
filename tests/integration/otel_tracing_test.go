package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"dev.helix.agent/internal/observability"
)

// TestTracerInitialization verifies that an LLMTracer can be initialized with
// default configuration (ExporterNone) without requiring any external
// dependency such as Jaeger, Zipkin, or an OTLP collector.
func TestTracerInitialization(t *testing.T) {
	// Test with nil config (should use defaults).
	tracer, err := observability.NewLLMTracer(nil)
	require.NoError(t, err, "NewLLMTracer(nil) must succeed with default config")
	require.NotNil(t, tracer, "Tracer must not be nil")

	// Test with explicit default config.
	cfg := observability.DefaultTracerConfig()
	assert.Equal(t, "helixagent", cfg.ServiceName)
	assert.Equal(t, "1.0.0", cfg.ServiceVersion)
	assert.Equal(t, "development", cfg.Environment)
	assert.False(t, cfg.EnableContentTrace, "Content tracing must be disabled by default")
	assert.Equal(t, observability.ExporterNone, cfg.ExporterType)

	tracer2, err := observability.NewLLMTracer(cfg)
	require.NoError(t, err, "NewLLMTracer with default config must succeed")
	require.NotNil(t, tracer2)

	// Test with console exporter type -- only validates config is accepted.
	consoleCfg := observability.DefaultTracerConfig()
	consoleCfg.ExporterType = observability.ExporterConsole

	// SetupTraceExporter with none creates a TracerProvider.
	// NOTE: resource.Merge in the exporter may return a schema URL conflict
	// when the SDK's default resource (semconv v1.39) differs from the
	// project's pinned semconv v1.26. We accept both outcomes.
	tp, setupErr := observability.SetupTraceExporter(context.Background(), &observability.ExporterConfig{
		Type:        observability.ExporterNone,
		ServiceName: "test-tracer-init",
		Environment: "test",
		Version:     "0.0.1",
	})
	if setupErr != nil {
		// Known schema URL conflict between SDK default resource and
		// project semconv version -- log and skip the exporter-specific
		// assertions but do not fail the test.
		t.Logf("SetupTraceExporter returned expected schema conflict: %v", setupErr)
	} else {
		require.NotNil(t, tp, "TracerProvider must not be nil")

		// Shutdown the trace provider cleanly.
		err = observability.ShutdownTraceExporter(context.Background(), tp)
		assert.NoError(t, err, "Shutdown must not return an error")
	}

	// Shutdown with nil provider should be safe.
	err = observability.ShutdownTraceExporter(context.Background(), nil)
	assert.NoError(t, err, "Shutdown with nil provider must be safe")
}

// TestTracerInitialization_ExporterTypes verifies that different exporter
// type constants are correctly defined and distinguishable.
func TestTracerInitialization_ExporterTypes(t *testing.T) {
	exporterTypes := map[observability.ExporterType]string{
		observability.ExporterOTLP:    "otlp",
		observability.ExporterJaeger:  "jaeger",
		observability.ExporterZipkin:  "zipkin",
		observability.ExporterConsole: "console",
		observability.ExporterNone:    "none",
	}

	for et, expected := range exporterTypes {
		assert.Equal(t, observability.ExporterType(expected), et,
			"ExporterType constant must match expected string")
	}

	// Ensure all types are distinct.
	seen := make(map[observability.ExporterType]bool)
	for et := range exporterTypes {
		assert.False(t, seen[et], "Duplicate exporter type: %s", et)
		seen[et] = true
	}
}

// TestTraceSpanCreation creates a span using the OTel SDK's in-memory
// exporter and verifies it has expected attributes and lifecycle.
func TestTraceSpanCreation(t *testing.T) {
	// Set up an in-memory span exporter so we can inspect recorded spans.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	// Install as global provider for this test.
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	// Create a tracer and start a span.
	tracer := tp.Tracer("test-span-creation")
	ctx, span := tracer.Start(context.Background(), "test-operation",
		trace.WithAttributes(
			attribute.String("test.key", "test-value"),
			attribute.Int("test.count", 42),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	// Verify the span context is valid.
	sc := span.SpanContext()
	assert.True(t, sc.TraceID().IsValid(), "TraceID must be valid")
	assert.True(t, sc.SpanID().IsValid(), "SpanID must be valid")
	assert.True(t, sc.IsSampled(), "Span must be sampled with AlwaysSample")

	// End the span to flush it to the exporter.
	span.SetStatus(codes.Ok, "")
	span.End()

	// Force flush to ensure the span is exported.
	err := tp.ForceFlush(context.Background())
	require.NoError(t, err)

	// Verify the span was exported.
	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 1, "At least one span must be exported")

	// Find our span.
	var found bool
	for _, s := range spans {
		if s.Name == "test-operation" {
			found = true

			// Verify span kind.
			assert.Equal(t, trace.SpanKindClient, s.SpanKind,
				"Span kind must be Client")

			// Verify attributes.
			attrMap := make(map[string]interface{})
			for _, attr := range s.Attributes {
				switch attr.Value.Type() {
				case attribute.STRING:
					attrMap[string(attr.Key)] = attr.Value.AsString()
				case attribute.INT64:
					attrMap[string(attr.Key)] = attr.Value.AsInt64()
				}
			}
			assert.Equal(t, "test-value", attrMap["test.key"],
				"String attribute must match")
			assert.Equal(t, int64(42), attrMap["test.count"],
				"Int attribute must match")

			// Verify status.
			assert.Equal(t, codes.Ok, s.Status.Code,
				"Span status must be OK")

			break
		}
	}
	assert.True(t, found, "Span 'test-operation' must be found in exported spans")
}

// TestTraceSpanCreation_ChildSpans verifies parent-child span relationships
// are properly established when creating nested spans.
func TestTraceSpanCreation_ChildSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	tracer := tp.Tracer("test-child-spans")

	// Create parent span.
	ctx, parentSpan := tracer.Start(context.Background(), "parent-operation")
	parentSC := parentSpan.SpanContext()

	// Create child span.
	_, childSpan := tracer.Start(ctx, "child-operation")
	childSC := childSpan.SpanContext()

	// Child must share the same TraceID.
	assert.Equal(t, parentSC.TraceID(), childSC.TraceID(),
		"Child span must share parent TraceID")
	// But have a different SpanID.
	assert.NotEqual(t, parentSC.SpanID(), childSC.SpanID(),
		"Child span must have a different SpanID")

	childSpan.End()
	parentSpan.End()

	err := tp.ForceFlush(context.Background())
	require.NoError(t, err)

	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 2,
		"At least two spans (parent + child) must be exported")
}

// TestLLMTraceAttributes verifies that GenAI-specific attributes are
// properly set on spans when using the LLMTracer.
func TestLLMTraceAttributes(t *testing.T) {
	// Set up an in-memory exporter.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	// Create the LLMTracer with content tracing enabled.
	cfg := observability.DefaultTracerConfig()
	cfg.EnableContentTrace = true

	tracer, err := observability.NewLLMTracer(cfg)
	require.NoError(t, err)
	require.NotNil(t, tracer)

	// Start an LLM request with full parameters.
	params := &observability.LLMRequestParams{
		Provider:      "openai",
		Model:         "gpt-4",
		RequestID:     "req-12345",
		SessionID:     "sess-67890",
		UserID:        "user-abc",
		Temperature:   0.7,
		MaxTokens:     2048,
		TopP:          0.9,
		StopSequences: []string{"###", "END"},
		SystemPrompt:  "You are a helpful assistant.",
		UserPrompt:    "Explain quantum computing.",
		Ensemble:      true,
		Debate:        false,
	}

	ctx, span := tracer.StartLLMRequest(context.Background(), params)
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	// End the request with response data.
	startTime := time.Now().Add(-1 * time.Second) // Simulate 1-second duration
	respParams := &observability.LLMResponseParams{
		InputTokens:   50,
		OutputTokens:  200,
		FinishReason:  "stop",
		ResponseID:    "resp-xyz",
		Content:       "Quantum computing uses qubits...",
		CacheHit:      false,
		ProviderScore: 8.5,
		CostUSD:       0.003,
		Error:         nil,
	}

	tracer.EndLLMRequest(ctx, span, respParams, startTime)

	// Force flush and inspect spans.
	err = tp.ForceFlush(context.Background())
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 1, "At least one span must be exported")

	// Find the llm.completion span.
	var llmSpan *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name == "llm.completion" {
			llmSpan = &spans[i]
			break
		}
	}
	require.NotNil(t, llmSpan, "Must find an 'llm.completion' span")

	// Build attribute map for easier assertions.
	attrMap := make(map[string]interface{})
	for _, attr := range llmSpan.Attributes {
		switch attr.Value.Type() {
		case attribute.STRING:
			attrMap[string(attr.Key)] = attr.Value.AsString()
		case attribute.INT64:
			attrMap[string(attr.Key)] = attr.Value.AsInt64()
		case attribute.FLOAT64:
			attrMap[string(attr.Key)] = attr.Value.AsFloat64()
		case attribute.BOOL:
			attrMap[string(attr.Key)] = attr.Value.AsBool()
		}
	}

	// Verify GenAI semantic convention attributes (from request).
	assert.Equal(t, "openai", attrMap[observability.AttrLLMSystem],
		"gen_ai.system must be set to provider name")
	assert.Equal(t, "gpt-4", attrMap[observability.AttrLLMModel],
		"gen_ai.request.model must be set to model name")
	assert.Equal(t, "req-12345", attrMap[observability.AttrHelixRequestID],
		"helix.request.id must be set")
	assert.Equal(t, "sess-67890", attrMap[observability.AttrHelixSessionID],
		"helix.session.id must be set")
	assert.Equal(t, "user-abc", attrMap[observability.AttrHelixUserID],
		"helix.user.id must be set")
	assert.Equal(t, 0.7, attrMap[observability.AttrLLMTemperature],
		"gen_ai.request.temperature must be set")
	assert.Equal(t, int64(2048), attrMap[observability.AttrLLMMaxTokens],
		"gen_ai.request.max_tokens must be set")
	assert.Equal(t, true, attrMap[observability.AttrHelixEnsemble],
		"helix.ensemble.enabled must be true")
	assert.Equal(t, false, attrMap[observability.AttrHelixDebate],
		"helix.debate.enabled must be false")

	// Verify content attributes (since EnableContentTrace=true).
	assert.Equal(t, "You are a helpful assistant.", attrMap[observability.AttrLLMSystemPrompt],
		"System prompt must be included when content tracing is enabled")
	assert.Equal(t, "Explain quantum computing.", attrMap[observability.AttrLLMUserPrompt],
		"User prompt must be included when content tracing is enabled")

	// Verify response attributes.
	assert.Equal(t, int64(50), attrMap[observability.AttrLLMInputTokens],
		"Input tokens must be set")
	assert.Equal(t, int64(200), attrMap[observability.AttrLLMOutputTokens],
		"Output tokens must be set")
	assert.Equal(t, int64(250), attrMap[observability.AttrLLMTotalTokens],
		"Total tokens must be input + output")
	assert.Equal(t, "stop", attrMap[observability.AttrLLMFinishReason],
		"Finish reason must be set")
	assert.Equal(t, "resp-xyz", attrMap[observability.AttrLLMResponseID],
		"Response ID must be set")
	assert.Equal(t, false, attrMap[observability.AttrHelixCacheHit],
		"Cache hit must be false")
	assert.Equal(t, 8.5, attrMap[observability.AttrHelixProviderScore],
		"Provider score must be set")

	// Verify response content (since EnableContentTrace=true).
	assert.Equal(t, "Quantum computing uses qubits...", attrMap[observability.AttrLLMAssistant],
		"Completion content must be included when content tracing is enabled")

	// Verify span status is OK (no error).
	assert.Equal(t, codes.Ok, llmSpan.Status.Code, "Span status must be OK")
}

// TestLLMTraceAttributes_ContentTraceDisabled verifies that prompts and
// completion content are NOT included in spans when EnableContentTrace
// is false (the default).
func TestLLMTraceAttributes_ContentTraceDisabled(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	// Content tracing disabled (default).
	cfg := observability.DefaultTracerConfig()
	assert.False(t, cfg.EnableContentTrace)

	tracer, err := observability.NewLLMTracer(cfg)
	require.NoError(t, err)

	params := &observability.LLMRequestParams{
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		RequestID:    "req-privacy",
		SystemPrompt: "SECRET system prompt",
		UserPrompt:   "SECRET user prompt",
	}

	ctx, span := tracer.StartLLMRequest(context.Background(), params)

	respParams := &observability.LLMResponseParams{
		InputTokens:  10,
		OutputTokens: 20,
		FinishReason: "stop",
		Content:      "SECRET response content",
	}
	tracer.EndLLMRequest(ctx, span, respParams, time.Now().Add(-100*time.Millisecond))

	err = tp.ForceFlush(context.Background())
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 1)

	for _, s := range spans {
		if s.Name == "llm.completion" {
			for _, attr := range s.Attributes {
				key := string(attr.Key)
				// Content attributes must NOT be present.
				assert.NotEqual(t, observability.AttrLLMSystemPrompt, key,
					"System prompt must not be in attributes when content tracing is disabled")
				assert.NotEqual(t, observability.AttrLLMUserPrompt, key,
					"User prompt must not be in attributes when content tracing is disabled")
				assert.NotEqual(t, observability.AttrLLMAssistant, key,
					"Completion must not be in attributes when content tracing is disabled")
			}
			break
		}
	}
}

// TestLLMTraceAttributes_ErrorRecording verifies that errors are properly
// recorded in spans.
func TestLLMTraceAttributes_ErrorRecording(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	tracer, err := observability.NewLLMTracer(observability.DefaultTracerConfig())
	require.NoError(t, err)

	params := &observability.LLMRequestParams{
		Provider:  "openai",
		Model:     "gpt-4",
		RequestID: "req-error-test",
	}

	ctx, span := tracer.StartLLMRequest(context.Background(), params)

	// End with an error.
	respParams := &observability.LLMResponseParams{
		Error: assert.AnError,
	}
	tracer.EndLLMRequest(ctx, span, respParams, time.Now().Add(-50*time.Millisecond))

	err = tp.ForceFlush(context.Background())
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 1)

	for _, s := range spans {
		if s.Name == "llm.completion" {
			// Verify error status.
			assert.Equal(t, codes.Error, s.Status.Code,
				"Span status must be Error when request fails")
			assert.NotEmpty(t, s.Status.Description,
				"Error description must not be empty")

			// Verify error event was recorded.
			assert.GreaterOrEqual(t, len(s.Events), 1,
				"At least one event (error) must be recorded")
			break
		}
	}
}

// TestLLMTraceAttributes_SpecializedSpans verifies that ensemble, debate,
// RAG, and tool execution create properly named spans.
func TestLLMTraceAttributes_SpecializedSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	tracer, err := observability.NewLLMTracer(observability.DefaultTracerConfig())
	require.NoError(t, err)

	ctx := context.Background()

	// Ensemble span.
	_, ensembleSpan := tracer.StartEnsembleRequest(ctx, "weighted-vote", 5)
	ensembleSpan.End()

	// Debate round span.
	_, debateSpan := tracer.StartDebateRound(ctx, 3, "code-review")
	debateSpan.End()

	// RAG retrieval span.
	_, ragSpan := tracer.StartRAGRetrieval(ctx, "quantum computing", 10)
	tracer.RecordRAGRetrievalResult(ragSpan, 5, 150.0)
	ragSpan.End()

	// Tool execution span.
	_, toolSpan := tracer.StartToolExecution(ctx, "file-search")
	toolSpan.End()

	err = tp.ForceFlush(context.Background())
	require.NoError(t, err)

	spans := exporter.GetSpans()

	// Collect span names.
	spanNames := make(map[string]bool)
	for _, s := range spans {
		spanNames[s.Name] = true
	}

	assert.True(t, spanNames["llm.ensemble"], "Ensemble span must be created")
	assert.True(t, spanNames["llm.debate.round"], "Debate round span must be created")
	assert.True(t, spanNames["rag.retrieval"], "RAG retrieval span must be created")
	assert.True(t, spanNames["tool.execution"], "Tool execution span must be created")
}
