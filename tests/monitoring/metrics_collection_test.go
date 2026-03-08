package monitoring_test

import (
	"context"
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

// TestMonitoring_Counter_Monotonic verifies that counter metrics only increase
// and never decrease. Counters in Prometheus are monotonically increasing.
func TestMonitoring_Counter_Monotonic(t *testing.T) {
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_monotonic_counter_total",
		Help: "Test counter to verify monotonic behavior",
	})
	registry.MustRegister(counter)

	// Increment the counter multiple times
	values := make([]float64, 0, 10)
	for i := 0; i < 10; i++ {
		counter.Inc()
		// Gather current value
		metricFamilies, err := registry.Gather()
		require.NoError(t, err)
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_monotonic_counter_total" {
				for _, m := range mf.GetMetric() {
					values = append(values, m.GetCounter().GetValue())
				}
			}
		}
	}

	// Verify monotonically increasing
	require.Greater(t, len(values), 1, "Should have collected multiple values")
	for i := 1; i < len(values); i++ {
		assert.GreaterOrEqual(t, values[i], values[i-1],
			"Counter value at step %d (%f) must be >= step %d (%f)",
			i, values[i], i-1, values[i-1])
	}

	// Also verify via the OTel LLM metrics counter interface
	metrics, err := observability.NewLLMMetrics("test-counter-monotonic")
	require.NoError(t, err)

	ctx := context.Background()
	for i := 0; i < 100; i++ {
		// These should not panic and only increment
		metrics.RequestsTotal.Add(ctx, 1)
		metrics.ErrorsTotal.Add(ctx, 1)
		metrics.CacheHits.Add(ctx, 1)
	}
	// Success: OTel counters do not support negative Add values at the API level
}

// TestMonitoring_Histogram_Buckets verifies that histograms have appropriate
// buckets configured for latency measurement.
func TestMonitoring_Histogram_Buckets(t *testing.T) {
	registry := prometheus.NewRegistry()

	latencyBuckets := []float64{
		0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
	}

	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_latency_seconds",
		Help:    "Test histogram for latency bucket validation",
		Buckets: latencyBuckets,
	})
	registry.MustRegister(histogram)

	// Observe values across the bucket range
	testValues := []float64{0.001, 0.008, 0.03, 0.07, 0.15, 0.3, 0.7, 1.5, 3.0, 7.0, 15.0}
	for _, v := range testValues {
		histogram.Observe(v)
	}

	// Gather and validate bucket structure
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	var found bool
	for _, mf := range metricFamilies {
		if mf.GetName() == "test_latency_seconds" {
			found = true
			for _, m := range mf.GetMetric() {
				hist := m.GetHistogram()
				require.NotNil(t, hist, "Metric must contain histogram data")

				buckets := hist.GetBucket()
				// Expected bucket count = len(latencyBuckets), Prometheus adds +Inf
				assert.GreaterOrEqual(t, len(buckets), len(latencyBuckets),
					"Histogram must have at least %d buckets", len(latencyBuckets))

				// Verify bucket upper bounds are in ascending order
				for i := 1; i < len(buckets); i++ {
					assert.Greater(t,
						buckets[i].GetUpperBound(),
						buckets[i-1].GetUpperBound(),
						"Bucket bounds must be in ascending order")
				}

				// Verify total count matches observed values
				assert.Equal(t, uint64(len(testValues)),
					hist.GetSampleCount(),
					"Histogram sample count must match observations")
			}
		}
	}
	assert.True(t, found, "Histogram metric must be present in gathered output")
}

// TestMonitoring_Gauge_Reflects_State verifies that gauge metrics correctly
// reflect the current state, going both up and down.
func TestMonitoring_Gauge_Reflects_State(t *testing.T) {
	registry := prometheus.NewRegistry()

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_active_connections",
		Help: "Test gauge for state tracking",
	})
	registry.MustRegister(gauge)

	getGaugeValue := func() float64 {
		metricFamilies, err := registry.Gather()
		require.NoError(t, err)
		for _, mf := range metricFamilies {
			if mf.GetName() == "test_active_connections" {
				for _, m := range mf.GetMetric() {
					return m.GetGauge().GetValue()
				}
			}
		}
		t.Fatal("Gauge metric not found")
		return 0
	}

	// Initial value should be 0
	assert.Equal(t, 0.0, getGaugeValue())

	// Increment
	gauge.Inc()
	gauge.Inc()
	gauge.Inc()
	assert.Equal(t, 3.0, getGaugeValue())

	// Decrement
	gauge.Dec()
	assert.Equal(t, 2.0, getGaugeValue())

	// Set absolute value
	gauge.Set(42.0)
	assert.Equal(t, 42.0, getGaugeValue())

	// Set to zero
	gauge.Set(0)
	assert.Equal(t, 0.0, getGaugeValue())

	// Also verify via OTel interface
	metrics, err := observability.NewLLMMetrics("test-gauge-state")
	require.NoError(t, err)

	ctx := context.Background()
	// UpDownCounter can go up and down
	metrics.RequestsInFlight.Add(ctx, 5)
	metrics.RequestsInFlight.Add(ctx, -3)
	// No panic means success
}

// TestMonitoring_Labels_Consistent verifies that metric labels follow
// consistent naming conventions (snake_case, no spaces).
func TestMonitoring_Labels_Consistent(t *testing.T) {
	registry := prometheus.NewRegistry()

	// Register metrics with various label configurations
	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_labeled_requests_total",
			Help: "Test counter with labels",
		},
		[]string{"provider", "model", "status_code"},
	)
	registry.MustRegister(counterVec)

	histogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_labeled_latency_seconds",
			Help:    "Test histogram with labels",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "operation_type"},
	)
	registry.MustRegister(histogramVec)

	// Record with various label values
	counterVec.WithLabelValues("openai", "gpt-4", "200").Inc()
	counterVec.WithLabelValues("anthropic", "claude-3", "500").Inc()
	histogramVec.WithLabelValues("openai", "complete").Observe(0.5)
	histogramVec.WithLabelValues("anthropic", "stream").Observe(1.2)

	// Gather all metrics and validate label naming
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		name := mf.GetName()
		if !strings.HasPrefix(name, "test_labeled_") {
			continue
		}

		// Metric name should be snake_case
		assert.NotContains(t, name, " ",
			"Metric name %q must not contain spaces", name)
		assert.NotContains(t, name, "-",
			"Metric name %q should use underscores, not hyphens", name)

		for _, m := range mf.GetMetric() {
			for _, label := range m.GetLabel() {
				labelName := label.GetName()
				// Label name should be snake_case
				assert.NotContains(t, labelName, " ",
					"Label name %q must not contain spaces", labelName)
				assert.NotContains(t, labelName, "-",
					"Label name %q should use underscores, not hyphens",
					labelName)

				// Label value should not be empty
				assert.NotEmpty(t, label.GetValue(),
					"Label %q should have a non-empty value", labelName)
			}
		}
	}
}

// TestMonitoring_MetricRegistration_NoDuplicates verifies that attempting
// to register duplicate metrics either safely returns the existing one or
// is properly handled.
func TestMonitoring_MetricRegistration_NoDuplicates(t *testing.T) {
	t.Run("PrometheusRegistryRejectsDuplicates", func(t *testing.T) {
		registry := prometheus.NewRegistry()

		counter1 := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_dup_counter_total",
			Help: "First registration",
		})
		err := registry.Register(counter1)
		require.NoError(t, err, "First registration should succeed")

		counter2 := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_dup_counter_total",
			Help: "Duplicate registration",
		})
		err = registry.Register(counter2)
		assert.Error(t, err,
			"Duplicate registration must return an error")
		assert.Contains(t, err.Error(), "test_dup_counter_total",
			"Error should reference the duplicate metric name")
	})

	t.Run("OTelMetricsHandleDuplicateNames", func(t *testing.T) {
		// OTel meters with the same service name should return
		// the same instruments for the same metric names.
		metrics1, err := observability.NewLLMMetrics("test-dup-otel")
		require.NoError(t, err)

		metrics2, err := observability.NewLLMMetrics("test-dup-otel")
		require.NoError(t, err)

		// Both should be usable without panic
		ctx := context.Background()
		metrics1.RequestsTotal.Add(ctx, 1)
		metrics2.RequestsTotal.Add(ctx, 1)
		metrics1.RecordRequest(ctx, "p", "m", time.Millisecond, 1, 1, 0, nil)
		metrics2.RecordRequest(ctx, "p", "m", time.Millisecond, 1, 1, 0, nil)
	})

	t.Run("ExtendedMetricsInitializeSafely", func(t *testing.T) {
		// NewMCPMetrics, NewEmbeddingMetrics, etc. should not fail
		mcpMetrics, err := observability.NewMCPMetrics("test-dup-mcp")
		require.NoError(t, err)
		assert.NotNil(t, mcpMetrics)

		embMetrics, err := observability.NewEmbeddingMetrics("test-dup-emb")
		require.NoError(t, err)
		assert.NotNil(t, embMetrics)

		vecMetrics, err := observability.NewVectorDBMetrics("test-dup-vec")
		require.NoError(t, err)
		assert.NotNil(t, vecMetrics)

		memMetrics, err := observability.NewMemoryMetrics("test-dup-mem")
		require.NoError(t, err)
		assert.NotNil(t, memMetrics)

		streamMetrics, err := observability.NewStreamingMetrics("test-dup-stream")
		require.NoError(t, err)
		assert.NotNil(t, streamMetrics)

		protoMetrics, err := observability.NewProtocolMetrics("test-dup-proto")
		require.NoError(t, err)
		assert.NotNil(t, protoMetrics)
	})

	t.Run("PrometheusExpositionAfterRegistration", func(t *testing.T) {
		// Validate that after registration the /metrics endpoint serves
		// the expected metrics without duplicates.
		registry := prometheus.NewRegistry()

		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_unique_counter_total",
			Help: "Unique counter",
		})
		registry.MustRegister(counter)
		counter.Add(5)

		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		body := w.Body.String()
		// Count occurrences of the metric name in TYPE lines
		typeCount := strings.Count(body, "# TYPE test_unique_counter_total")
		assert.Equal(t, 1, typeCount,
			"Each metric should have exactly one TYPE declaration")
	})
}
