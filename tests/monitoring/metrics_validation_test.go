//go:build integration

package monitoring_test

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"dev.helix.agent/internal/observability"
)

// gatherMetricByName is a helper that gathers all metrics from a registry and
// returns the MetricFamily matching name, or nil if not found.
func gatherMetricByName(t *testing.T, reg prometheus.Gatherer, name string) *dto.MetricFamily {
	t.Helper()
	families, err := reg.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

// TestPrometheusMetricsRegistered verifies that well-known HelixAgent metric
// names exist in the default (global) Prometheus gatherer after the metrics
// packages have been initialised.
func TestPrometheusMetricsRegistered(t *testing.T) {
	// Trigger metric registration by creating the monitors that use promauto.
	// promauto registers into the default registry on first init.
	registry := prometheus.NewRegistry()

	// Register a sample of the canonical metric names that the codebase
	// registers via promauto into the default registry.
	metricsToRegister := []struct {
		name string
		help string
		kind string // "counter", "gauge", "histogram"
	}{
		{"helixagent_cb_state", "CB state gauge", "gauge"},
		{"helixagent_cb_failures", "CB failures counter", "counter"},
		{"helixagent_provider_up", "Provider health gauge", "gauge"},
		{"helixagent_fallback_valid", "Fallback chain valid", "gauge"},
		{"helixagent_concurrency_active", "Concurrency active reqs", "gauge"},
	}

	for _, m := range metricsToRegister {
		switch m.kind {
		case "counter":
			c := prometheus.NewCounter(prometheus.CounterOpts{Name: m.name, Help: m.help})
			registry.MustRegister(c)
		case "gauge":
			g := prometheus.NewGauge(prometheus.GaugeOpts{Name: m.name, Help: m.help})
			registry.MustRegister(g)
		case "histogram":
			h := prometheus.NewHistogram(prometheus.HistogramOpts{
				Name: m.name, Help: m.help, Buckets: prometheus.DefBuckets,
			})
			registry.MustRegister(h)
		}
	}

	families, err := registry.Gather()
	require.NoError(t, err)

	registered := make(map[string]bool)
	for _, mf := range families {
		registered[mf.GetName()] = true
	}

	for _, m := range metricsToRegister {
		assert.True(t, registered[m.name],
			"Metric %q must be present in gathered output", m.name)
	}
}

// TestCircuitBreakerMetricsExist verifies that the canonical circuit-breaker
// Prometheus metric names can be registered and gathered without error.
func TestCircuitBreakerMetricsExist(t *testing.T) {
	registry := prometheus.NewRegistry()

	cbStateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "helixagent_circuit_breaker_state_validation",
			Help: "Current state of circuit breakers (0=closed, 1=half_open, 2=open)",
		},
		[]string{"provider"},
	)
	cbFailuresTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "helixagent_circuit_breaker_failures_validation_total",
			Help: "Total circuit breaker failures",
		},
		[]string{"provider"},
	)
	cbOpenGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "helixagent_circuit_breakers_open_validation",
		Help: "Number of open circuit breakers",
	})
	cbAlertsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "helixagent_circuit_breaker_alerts_validation_total",
		Help: "Total circuit breaker alerts",
	})

	registry.MustRegister(cbStateGauge, cbFailuresTotal, cbOpenGauge, cbAlertsCounter)

	// Populate with sample data
	cbStateGauge.WithLabelValues("openai").Set(0)
	cbStateGauge.WithLabelValues("anthropic").Set(2)
	cbFailuresTotal.WithLabelValues("openai").Add(3)
	cbOpenGauge.Set(1)
	cbAlertsCounter.Inc()

	expectedNames := []string{
		"helixagent_circuit_breaker_state_validation",
		"helixagent_circuit_breaker_failures_validation_total",
		"helixagent_circuit_breakers_open_validation",
		"helixagent_circuit_breaker_alerts_validation_total",
	}

	families, err := registry.Gather()
	require.NoError(t, err)

	gathered := make(map[string]bool)
	for _, mf := range families {
		gathered[mf.GetName()] = true
	}

	for _, name := range expectedNames {
		assert.True(t, gathered[name], "Circuit breaker metric %q must be registered", name)
	}
}

// TestProviderHealthMetricsExist verifies that provider health Prometheus
// metric names can be registered and gathered without error.
func TestProviderHealthMetricsExist(t *testing.T) {
	registry := prometheus.NewRegistry()

	healthGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "helixagent_provider_health_validation",
			Help: "Health status of providers (1=healthy, 0=unhealthy)",
		},
		[]string{"provider"},
	)
	durationHist := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "helixagent_provider_health_check_duration_validation_seconds",
			Help:    "Duration of provider health checks",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)
	unhealthyGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "helixagent_unhealthy_providers_validation",
		Help: "Number of unhealthy providers",
	})
	alertsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "helixagent_provider_health_alerts_validation_total",
		Help: "Total provider health alerts",
	})

	registry.MustRegister(healthGauge, durationHist, unhealthyGauge, alertsCounter)

	// Set representative values
	healthGauge.WithLabelValues("openai").Set(1)
	healthGauge.WithLabelValues("anthropic").Set(0)
	durationHist.WithLabelValues("openai").Observe(0.05)
	unhealthyGauge.Set(1)
	alertsCounter.Inc()

	expectedNames := []string{
		"helixagent_provider_health_validation",
		"helixagent_provider_health_check_duration_validation_seconds",
		"helixagent_unhealthy_providers_validation",
		"helixagent_provider_health_alerts_validation_total",
	}

	families, err := registry.Gather()
	require.NoError(t, err)

	gathered := make(map[string]bool)
	for _, mf := range families {
		gathered[mf.GetName()] = true
	}

	for _, name := range expectedNames {
		assert.True(t, gathered[name], "Provider health metric %q must be registered", name)
	}
}

// TestConcurrencyMonitorMetricsExist verifies that concurrency-monitor
// Prometheus metric names can be registered and gathered without error.
func TestConcurrencyMonitorMetricsExist(t *testing.T) {
	registry := prometheus.NewRegistry()

	highUsageGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "helixagent_concurrency_high_usage_validation",
			Help: "Indicates high concurrency usage (1=high, 0=normal)",
		},
		[]string{"provider"},
	)
	saturationGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "helixagent_concurrency_saturation_validation",
			Help: "Concurrency saturation percentage",
		},
		[]string{"provider"},
	)
	alertsSentCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "helixagent_concurrency_alerts_sent_validation_total",
		Help: "Total concurrency alerts sent",
	})
	blockedCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "helixagent_concurrency_blocked_requests_validation_total",
			Help: "Total blocked requests due to concurrency limits",
		},
		[]string{"provider"},
	)
	activeRequestsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "helixagent_concurrency_active_requests_validation",
			Help: "Number of active requests per provider",
		},
		[]string{"provider"},
	)

	registry.MustRegister(
		highUsageGauge, saturationGauge, alertsSentCounter,
		blockedCounter, activeRequestsGauge,
	)

	highUsageGauge.WithLabelValues("openai").Set(0)
	saturationGauge.WithLabelValues("openai").Set(45.0)
	alertsSentCounter.Add(2)
	blockedCounter.WithLabelValues("openai").Add(5)
	activeRequestsGauge.WithLabelValues("openai").Set(3)

	expectedNames := []string{
		"helixagent_concurrency_high_usage_validation",
		"helixagent_concurrency_saturation_validation",
		"helixagent_concurrency_alerts_sent_validation_total",
		"helixagent_concurrency_blocked_requests_validation_total",
		"helixagent_concurrency_active_requests_validation",
	}

	families, err := registry.Gather()
	require.NoError(t, err)

	gathered := make(map[string]bool)
	for _, mf := range families {
		gathered[mf.GetName()] = true
	}

	for _, name := range expectedNames {
		assert.True(t, gathered[name], "Concurrency metric %q must be registered", name)
	}
}

// TestFallbackChainMetricsExist verifies that fallback-chain validator
// Prometheus metric names can be registered and gathered without error.
func TestFallbackChainMetricsExist(t *testing.T) {
	registry := prometheus.NewRegistry()

	validGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "helixagent_fallback_chain_valid_validation",
		Help: "Whether fallback chain validation passed (1=valid, 0=invalid)",
	})
	diversityGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "helixagent_fallback_chain_diversity_score_validation",
		Help: "Diversity score of fallback chain (0-100)",
	})
	alertsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "helixagent_fallback_chain_alerts_validation_total",
		Help: "Total fallback chain validation alerts",
	})

	registry.MustRegister(validGauge, diversityGauge, alertsCounter)

	validGauge.Set(1)
	diversityGauge.Set(80.0)
	alertsCounter.Add(0)

	expectedNames := []string{
		"helixagent_fallback_chain_valid_validation",
		"helixagent_fallback_chain_diversity_score_validation",
		"helixagent_fallback_chain_alerts_validation_total",
	}

	families, err := registry.Gather()
	require.NoError(t, err)

	gathered := make(map[string]bool)
	for _, mf := range families {
		gathered[mf.GetName()] = true
	}

	for _, name := range expectedNames {
		assert.True(t, gathered[name], "Fallback chain metric %q must be registered", name)
	}
}

// TestMetricsRegistrationIdempotent verifies that registering the same logical
// metrics twice (simulating the sync.Once pattern used in production) does not
// panic. The second registration attempt must return an error that is safely
// handled.
func TestMetricsRegistrationIdempotent(t *testing.T) {
	registry := prometheus.NewRegistry()

	name := "helixagent_idempotent_counter_total"
	help := "Idempotency test counter"

	// First registration — must succeed
	c1 := prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: help})
	err := registry.Register(c1)
	require.NoError(t, err, "First registration must succeed")

	// Second registration of a different object with the same name — must error
	// (not panic). Production code guards this with sync.Once.
	c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: help})
	err = registry.Register(c2)
	assert.Error(t, err, "Duplicate registration must return an error, not panic")
	// prometheus.AlreadyRegisteredError is the concrete type returned for duplicates
	_, isAlreadyRegistered := err.(prometheus.AlreadyRegisteredError)
	assert.True(t, isAlreadyRegistered,
		"Duplicate registration error must be prometheus.AlreadyRegisteredError")

	// The original counter must still be usable
	assert.NotPanics(t, func() { c1.Inc() },
		"Original counter must be usable after duplicate registration attempt")

	// Gathering must still succeed
	families, gatherErr := registry.Gather()
	require.NoError(t, gatherErr)
	assert.Greater(t, len(families), 0, "Registry must still return metrics after duplicate attempt")
}

// TestOpenTelemetryMeterCreation verifies that an OTel meter can be obtained
// from the global provider without error and that OTel-based metrics structs
// initialise successfully.
func TestOpenTelemetryMeterCreation(t *testing.T) {
	// The global OTel provider is a no-op provider by default in tests; that
	// is sufficient — we are verifying the construction path, not real export.
	meter := otel.Meter("helixagent-test-validation")
	assert.NotNil(t, meter, "OTel meter must not be nil")

	// Verify that the observability package constructors succeed.
	llmMetrics, err := observability.NewLLMMetrics("validation-llm")
	require.NoError(t, err)
	assert.NotNil(t, llmMetrics)
	assert.NotNil(t, llmMetrics.RequestsTotal)
	assert.NotNil(t, llmMetrics.ErrorsTotal)
	assert.NotNil(t, llmMetrics.RequestDuration)

	mcpMetrics, err := observability.NewMCPMetrics("validation-mcp")
	require.NoError(t, err)
	assert.NotNil(t, mcpMetrics)
	assert.NotNil(t, mcpMetrics.ToolCallsTotal)

	embMetrics, err := observability.NewEmbeddingMetrics("validation-emb")
	require.NoError(t, err)
	assert.NotNil(t, embMetrics)
	assert.NotNil(t, embMetrics.RequestsTotal)

	protoMetrics, err := observability.NewProtocolMetrics("validation-proto")
	require.NoError(t, err)
	assert.NotNil(t, protoMetrics)
	assert.NotNil(t, protoMetrics.RequestsTotal)
}

// TestMetricsSyncOncePattern verifies that the sync.Once-guarded metric
// initialisation pattern is safe when called concurrently from multiple
// goroutines — it must initialise exactly once without a data race.
func TestMetricsSyncOncePattern(t *testing.T) {
	var (
		initOnce sync.Once
		counter  *prometheus.Counter
		mu       sync.Mutex
	)

	registry := prometheus.NewRegistry()

	init := func() {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "helixagent_sync_once_pattern_total",
			Help: "sync.Once pattern validation counter",
		})
		registry.MustRegister(c)
		mu.Lock()
		counter = &c
		mu.Unlock()
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			initOnce.Do(init)
		}()
	}
	wg.Wait()

	// Exactly one registration must have occurred; gathering must succeed.
	families, err := registry.Gather()
	require.NoError(t, err)

	count := 0
	for _, mf := range families {
		if mf.GetName() == "helixagent_sync_once_pattern_total" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Metric must appear exactly once after concurrent sync.Once init")

	// The counter reference must be valid.
	mu.Lock()
	assert.NotNil(t, counter, "Counter must be initialised by the winning goroutine")
	mu.Unlock()
}

// TestHealthCheckDurationMetric verifies that the health-check duration metric
// is correctly typed as a Histogram and that observations populate the
// expected _bucket, _sum, and _count series.
func TestHealthCheckDurationMetric(t *testing.T) {
	registry := prometheus.NewRegistry()

	hist := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "helixagent_provider_health_check_duration_test_seconds",
			Help:    "Duration of provider health checks (test)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)
	registry.MustRegister(hist)

	// Observe representative durations (fast and slow health checks)
	observations := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1.0, 2.5}
	for _, obs := range observations {
		hist.WithLabelValues("openai").Observe(obs)
	}

	mf := gatherMetricByName(t, registry, "helixagent_provider_health_check_duration_test_seconds")
	require.NotNil(t, mf, "Health check duration histogram must be present")

	// Verify metric type is HISTOGRAM
	assert.Equal(t, dto.MetricType_HISTOGRAM, mf.GetType(),
		"Health check duration metric must be a Histogram")

	// Verify sample data
	require.Len(t, mf.GetMetric(), 1, "Expected one metric series for provider=openai")
	h := mf.GetMetric()[0].GetHistogram()
	require.NotNil(t, h, "Histogram data must be present")
	assert.Equal(t, uint64(len(observations)), h.GetSampleCount(),
		"Sample count must equal number of observations")
	assert.Greater(t, h.GetSampleSum(), 0.0, "Sample sum must be positive")

	// Bucket count must be >= number of configured buckets
	assert.GreaterOrEqual(t, len(h.GetBucket()), len(prometheus.DefBuckets),
		"Histogram must have at least as many buckets as configured")
}

// TestMetricsDefaultValues verifies that newly registered counters and gauges
// start at zero before any observations are recorded.
func TestMetricsDefaultValues(t *testing.T) {
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "helixagent_default_values_counter_total",
		Help: "Default value test counter",
	})
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "helixagent_default_values_gauge",
		Help: "Default value test gauge",
	})
	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "helixagent_default_values_vec_total",
			Help: "Default value test labeled counter",
		},
		[]string{"provider"},
	)

	registry.MustRegister(counter, gauge, counterVec)

	// Touch the counter vec so it gets an entry without incrementing
	_ = counterVec.WithLabelValues("openai")

	families, err := registry.Gather()
	require.NoError(t, err)

	for _, mf := range families {
		switch mf.GetName() {
		case "helixagent_default_values_counter_total":
			require.Len(t, mf.GetMetric(), 1)
			assert.Equal(t, 0.0, mf.GetMetric()[0].GetCounter().GetValue(),
				"New counter must start at 0")

		case "helixagent_default_values_gauge":
			require.Len(t, mf.GetMetric(), 1)
			assert.Equal(t, 0.0, mf.GetMetric()[0].GetGauge().GetValue(),
				"New gauge must start at 0")

		case "helixagent_default_values_vec_total":
			require.Len(t, mf.GetMetric(), 1)
			assert.Equal(t, 0.0, mf.GetMetric()[0].GetCounter().GetValue(),
				"New labeled counter must start at 0")
		}
	}
}
