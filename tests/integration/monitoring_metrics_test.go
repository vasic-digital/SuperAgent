package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/observability"
)

// TestPrometheusMetricsRegistered verifies that Prometheus metric families
// are properly registered. We create a fresh registry, register known
// metrics, and verify they appear in the gathered output.
func TestPrometheusMetricsRegistered(t *testing.T) {
	registry := prometheus.NewRegistry()

	// Register representative Prometheus metrics matching the project's
	// metrics/collector.go definitions.
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	providerLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_provider_latency_seconds",
			Help:    "LLM provider latency in seconds",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"provider", "model"},
	)

	cacheHits := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits",
		},
		[]string{"cache_type"},
	)

	cacheMisses := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses",
		},
		[]string{"cache_type"},
	)

	registry.MustRegister(requestDuration)
	registry.MustRegister(providerLatency)
	registry.MustRegister(cacheHits)
	registry.MustRegister(cacheMisses)

	// Emit at least one observation so the metrics appear when gathered.
	requestDuration.WithLabelValues("GET", "/v1/chat/completions", "200").Observe(0.150)
	providerLatency.WithLabelValues("openai", "gpt-4").Observe(1.5)
	cacheHits.WithLabelValues("response").Inc()
	cacheMisses.WithLabelValues("response").Inc()

	// Gather all metric families from the registry.
	families, err := registry.Gather()
	require.NoError(t, err, "Gathering metrics must not fail")
	require.Greater(t, len(families), 0, "At least one metric family expected")

	familyNames := make(map[string]bool)
	for _, f := range families {
		familyNames[f.GetName()] = true
	}

	// Verify each registered metric appears.
	expectedMetrics := []string{
		"http_request_duration_seconds",
		"llm_provider_latency_seconds",
		"cache_hits_total",
		"cache_misses_total",
	}
	for _, name := range expectedMetrics {
		assert.True(t, familyNames[name], "Metric family %q must be registered", name)
	}
}

// TestPrometheusMetricsRegistered_OTelLLM verifies that the OTel-based
// LLMMetrics from internal/observability can be instantiated and that
// recording operations do not panic.
func TestPrometheusMetricsRegistered_OTelLLM(t *testing.T) {
	// NewLLMMetrics creates OTel metric instruments (counters, histograms, gauges).
	metrics, err := observability.NewLLMMetrics("test-prometheus-metrics")
	require.NoError(t, err, "NewLLMMetrics must succeed")
	require.NotNil(t, metrics)

	// Verify the core metric instruments are non-nil.
	assert.NotNil(t, metrics.RequestsTotal, "RequestsTotal counter must be initialized")
	assert.NotNil(t, metrics.RequestDuration, "RequestDuration histogram must be initialized")
	assert.NotNil(t, metrics.InputTokens, "InputTokens counter must be initialized")
	assert.NotNil(t, metrics.OutputTokens, "OutputTokens counter must be initialized")
	assert.NotNil(t, metrics.TotalTokens, "TotalTokens counter must be initialized")
	assert.NotNil(t, metrics.ErrorsTotal, "ErrorsTotal counter must be initialized")
	assert.NotNil(t, metrics.CacheHits, "CacheHits counter must be initialized")
	assert.NotNil(t, metrics.CacheMisses, "CacheMisses counter must be initialized")
	assert.NotNil(t, metrics.DebateRounds, "DebateRounds counter must be initialized")
	assert.NotNil(t, metrics.RAGRetrievals, "RAGRetrievals counter must be initialized")

	// Verify extended metric types (MCPMetrics, EmbeddingMetrics, etc.).
	mcpMetrics, err := observability.NewMCPMetrics("test-mcp")
	require.NoError(t, err)
	assert.NotNil(t, mcpMetrics.ToolCallsTotal)
	assert.NotNil(t, mcpMetrics.ToolDuration)

	embMetrics, err := observability.NewEmbeddingMetrics("test-embedding")
	require.NoError(t, err)
	assert.NotNil(t, embMetrics.RequestsTotal)
	assert.NotNil(t, embMetrics.LatencySeconds)

	vecMetrics, err := observability.NewVectorDBMetrics("test-vectordb")
	require.NoError(t, err)
	assert.NotNil(t, vecMetrics.OperationsTotal)

	memMetrics, err := observability.NewMemoryMetrics("test-memory")
	require.NoError(t, err)
	assert.NotNil(t, memMetrics.OperationsTotal)

	streamMetrics, err := observability.NewStreamingMetrics("test-streaming")
	require.NoError(t, err)
	assert.NotNil(t, streamMetrics.ChunksTotal)

	protoMetrics, err := observability.NewProtocolMetrics("test-protocol")
	require.NoError(t, err)
	assert.NotNil(t, protoMetrics.RequestsTotal)
}

// TestMetricsEndpointServesData creates a test HTTP server with a Prometheus
// promhttp handler, hits /metrics, and verifies the response contains the
// expected metric names.
func TestMetricsEndpointServesData(t *testing.T) {
	registry := prometheus.NewRegistry()

	// Register a counter and a histogram so the /metrics output is non-empty.
	reqCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "helixagent_requests_total",
			Help: "Total number of requests",
		},
		[]string{"provider"},
	)

	latencyHist := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "helixagent_request_duration_seconds",
			Help:    "Request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	registry.MustRegister(reqCounter)
	registry.MustRegister(latencyHist)

	// Record some values so they appear in the output.
	reqCounter.WithLabelValues("deepseek").Add(42)
	latencyHist.WithLabelValues("deepseek").Observe(0.25)

	// Create a test server with promhttp using our custom registry.
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	require.NoError(t, err, "GET /metrics must succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint must return 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading response body must succeed")
	bodyStr := string(body)

	// Verify the response contains expected metric names.
	assert.True(t, strings.Contains(bodyStr, "helixagent_requests_total"),
		"Metrics output must contain helixagent_requests_total")
	assert.True(t, strings.Contains(bodyStr, "helixagent_request_duration_seconds"),
		"Metrics output must contain helixagent_request_duration_seconds")

	// Verify the output contains actual values.
	assert.True(t, strings.Contains(bodyStr, "42"),
		"Metrics output must contain the counter value 42")
	assert.True(t, strings.Contains(bodyStr, "deepseek"),
		"Metrics output must contain provider label deepseek")

	// Verify Content-Type indicates text/plain exposition format.
	contentType := resp.Header.Get("Content-Type")
	assert.True(t,
		strings.Contains(contentType, "text/plain") ||
			strings.Contains(contentType, "text/openmetrics"),
		"Content-Type must be text exposition format, got: %s", contentType)
}

// TestMetricValuesChangeOnOperation verifies that incrementing a Prometheus
// counter actually changes its value in the metrics output.
func TestMetricValuesChangeOnOperation(t *testing.T) {
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_operation_counter_total",
		Help: "Counter to verify value changes on operations",
	})
	registry.MustRegister(counter)

	// Helper to read current counter value from registry.
	getCounterValue := func() float64 {
		families, err := registry.Gather()
		require.NoError(t, err)
		for _, f := range families {
			if f.GetName() == "test_operation_counter_total" {
				for _, m := range f.GetMetric() {
					return m.GetCounter().GetValue()
				}
			}
		}
		t.Fatal("counter metric not found in gathered families")
		return 0
	}

	// Increment once and verify the value is 1.
	counter.Inc()
	val1 := getCounterValue()
	assert.Equal(t, float64(1), val1, "After one Inc() counter must be 1")

	// Increment 9 more times and verify the total is 10.
	for i := 0; i < 9; i++ {
		counter.Inc()
	}
	val2 := getCounterValue()
	assert.Equal(t, float64(10), val2, "After 10 total increments counter must be 10")

	// Use Add to bump by a specific amount.
	counter.Add(25)
	val3 := getCounterValue()
	assert.Equal(t, float64(35), val3, "After adding 25 more, counter must be 35")

	// Verify via /metrics endpoint as well.
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "35"),
		"Metrics endpoint must reflect the updated counter value of 35")
}

// TestMetricValuesChangeOnOperation_OTelRecordRequest verifies that the
// OTel-based LLMMetrics RecordRequest method does not panic and can be
// called repeatedly with various parameters.
func TestMetricValuesChangeOnOperation_OTelRecordRequest(t *testing.T) {
	metrics, err := observability.NewLLMMetrics("test-record-request")
	require.NoError(t, err)

	ctx := context.Background()

	// Record a successful request.
	metrics.RecordRequest(ctx, "openai", "gpt-4", 1500*time.Millisecond, 100, 200, 0.05, nil)

	// Record a failed request.
	metrics.RecordRequest(ctx, "deepseek", "deepseek-chat", 500*time.Millisecond, 50, 0, 0, assert.AnError)

	// Record cache hit/miss.
	metrics.RecordCacheHit(ctx, 10*time.Millisecond)
	metrics.RecordCacheMiss(ctx)

	// Record debate round.
	metrics.RecordDebateRound(ctx, 5, 0.85)

	// Record RAG retrieval.
	metrics.RecordRAGRetrieval(ctx, 10, 200*time.Millisecond, 0.92)

	// If we reached this point, none of the operations panicked.
	// This validates the full metric recording interface.
}
