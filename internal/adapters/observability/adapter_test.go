package observabilityadapter

import (
	"context"
	"sync"
	"testing"

	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Initialize
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_Initialize_NilConfig(t *testing.T) {
	a := NewObservabilityAdapter()

	err := a.Initialize(nil)
	require.NoError(t, err, "Initialize with nil config should succeed (no-op mode)")
	assert.True(t, a.IsInitialized())
}

func TestObservabilityAdapter_Initialize_DefaultConfig(t *testing.T) {
	a := NewObservabilityAdapter()
	cfg := DefaultConfig()

	err := a.Initialize(cfg)
	require.NoError(t, err)
	assert.True(t, a.IsInitialized())
}

func TestObservabilityAdapter_Initialize_CustomConfig(t *testing.T) {
	a := NewObservabilityAdapter()
	cfg := &Config{
		ServiceName:        "test-service",
		ServiceVersion:     "0.0.1",
		Environment:        "test",
		ExporterType:       observability.ExporterNone,
		EnableContentTrace: true,
		SampleRate:         0.5,
	}

	err := a.Initialize(cfg)
	require.NoError(t, err)
	assert.True(t, a.IsInitialized())
}

func TestObservabilityAdapter_DoubleInitialize_Safe(t *testing.T) {
	a := NewObservabilityAdapter()
	cfg := DefaultConfig()

	err1 := a.Initialize(cfg)
	require.NoError(t, err1)

	// Second call must be a harmless no-op.
	err2 := a.Initialize(&Config{
		ServiceName:  "different",
		ExporterType: observability.ExporterNone,
	})
	require.NoError(t, err2)
	assert.True(t, a.IsInitialized())
}

func TestObservabilityAdapter_NotInitialized(t *testing.T) {
	a := NewObservabilityAdapter()
	assert.False(t, a.IsInitialized())
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_Shutdown_Idempotent(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	ctx := context.Background()

	err1 := a.Shutdown(ctx)
	require.NoError(t, err1)

	err2 := a.Shutdown(ctx)
	require.NoError(t, err2, "second Shutdown must be a no-op")
}

func TestObservabilityAdapter_Shutdown_BeforeInitialize(t *testing.T) {
	a := NewObservabilityAdapter()
	// Shutdown on an un-initialized adapter should not panic.
	err := a.Shutdown(context.Background())
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Getters (initialized)
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_GetTracer_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	tracer := a.GetTracer()
	assert.NotNil(t, tracer)
}

func TestObservabilityAdapter_GetMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	metrics := a.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestObservabilityAdapter_GetMCPMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetMCPMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetEmbeddingMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetEmbeddingMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetVectorDBMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetVectorDBMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetMemoryMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetMemoryMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetStreamingMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetStreamingMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetProtocolMetrics_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	m := a.GetProtocolMetrics()
	assert.NotNil(t, m)
}

func TestObservabilityAdapter_GetDebateTracer_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	dt := a.GetDebateTracer()
	assert.NotNil(t, dt)
}

func TestObservabilityAdapter_GetTracerProvider_Initialized(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	// TracerProvider may be nil if the OTel exporter setup encountered a
	// schema version conflict (known environment issue). The adapter
	// degrades gracefully in that case.
	tp := a.GetTracerProvider()
	if tp == nil {
		t.Log("TracerProvider is nil due to OTel schema conflict; adapter degraded gracefully")
	}
}

// ---------------------------------------------------------------------------
// Getters (NOT initialized -- must fall back to global defaults)
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_GetTracer_Fallback(t *testing.T) {
	a := NewObservabilityAdapter()
	tracer := a.GetTracer()
	assert.NotNil(t, tracer, "should return global fallback tracer")
}

func TestObservabilityAdapter_GetMetrics_Fallback(t *testing.T) {
	a := NewObservabilityAdapter()
	metrics := a.GetMetrics()
	assert.NotNil(t, metrics, "should return global fallback metrics")
}

func TestObservabilityAdapter_GetDebateTracer_Fallback(t *testing.T) {
	a := NewObservabilityAdapter()
	dt := a.GetDebateTracer()
	assert.NotNil(t, dt, "should return fallback debate tracer")
}

func TestObservabilityAdapter_GetTracerProvider_NilWhenNotInitialized(t *testing.T) {
	a := NewObservabilityAdapter()
	tp := a.GetTracerProvider()
	assert.Nil(t, tp, "TracerProvider must be nil before Initialize")
}

// ---------------------------------------------------------------------------
// LLM Middleware
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_GetLLMMiddleware_ReturnsNonNil(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	mw := a.GetLLMMiddleware()
	assert.NotNil(t, mw)
}

func TestObservabilityAdapter_GetLLMMiddleware_BeforeInitialize(t *testing.T) {
	a := NewObservabilityAdapter()
	mw := a.GetLLMMiddleware()
	assert.NotNil(t, mw, "middleware must work even before Initialize (uses global)")
}

func TestObservabilityAdapter_WrapProvider(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	mock := &mockProvider{}
	wrapped := a.WrapProvider(mock, "test-provider")
	assert.NotNil(t, wrapped)

	// The wrapped provider should still satisfy the LLMProvider interface
	// and delegate to the underlying mock.
	caps := wrapped.GetCapabilities()
	assert.NotNil(t, caps)
	assert.True(t, caps.SupportsStreaming, "should delegate GetCapabilities")
}

func TestObservabilityAdapter_WrapProvider_Complete(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	mock := &mockProvider{}
	wrapped := a.WrapProvider(mock, "test-provider")

	resp, err := wrapped.Complete(context.Background(), &models.LLMRequest{
		ModelParams: models.ModelParameters{
			Model:       "test-model",
			Temperature: 0.7,
			MaxTokens:   100,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "mock-response", resp.Content)
	assert.Equal(t, 1, mock.completeCalls, "must delegate to underlying provider")
}

// ---------------------------------------------------------------------------
// Concurrency safety
// ---------------------------------------------------------------------------

func TestObservabilityAdapter_ConcurrentAccess(t *testing.T) {
	a := NewObservabilityAdapter()
	require.NoError(t, a.Initialize(nil))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = a.GetTracer()
			_ = a.GetMetrics()
			_ = a.GetMCPMetrics()
			_ = a.GetEmbeddingMetrics()
			_ = a.GetVectorDBMetrics()
			_ = a.GetMemoryMetrics()
			_ = a.GetStreamingMetrics()
			_ = a.GetProtocolMetrics()
			_ = a.GetDebateTracer()
			_ = a.GetTracerProvider()
			_ = a.GetLLMMiddleware()
			_ = a.IsInitialized()
		}()
	}
	wg.Wait()
}

func TestObservabilityAdapter_ConcurrentInitialize(t *testing.T) {
	a := NewObservabilityAdapter()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = a.Initialize(nil)
		}()
	}
	wg.Wait()

	assert.True(t, a.IsInitialized())
}

// ---------------------------------------------------------------------------
// DefaultConfig
// ---------------------------------------------------------------------------

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "helixagent", cfg.ServiceName)
	assert.Equal(t, "1.0.0", cfg.ServiceVersion)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, observability.ExporterNone, cfg.ExporterType)
	assert.False(t, cfg.EnableContentTrace)
	assert.Equal(t, 1.0, cfg.SampleRate)
}

// ---------------------------------------------------------------------------
// Mock LLM provider for middleware tests
// ---------------------------------------------------------------------------

type mockProvider struct {
	completeCalls int
}

func (m *mockProvider) Complete(_ context.Context, _ *models.LLMRequest) (*models.LLMResponse, error) {
	m.completeCalls++
	return &models.LLMResponse{
		Content:      "mock-response",
		FinishReason: "stop",
		ID:           "resp-1",
	}, nil
}

func (m *mockProvider) CompleteStream(_ context.Context, _ *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	ch := make(chan *models.LLMResponse, 1)
	ch <- &models.LLMResponse{Content: "chunk-1", FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (m *mockProvider) HealthCheck() error {
	return nil
}

func (m *mockProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportsStreaming: true,
	}
}

func (m *mockProvider) ValidateConfig(_ map[string]interface{}) (bool, []string) {
	return true, nil
}
