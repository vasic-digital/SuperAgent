package monitoring_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/handlers"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestMonitoring_HealthEndpoint_Structure validates that the health response
// from the monitoring status endpoint has the expected top-level fields.
func TestMonitoring_HealthEndpoint_Structure(t *testing.T) {
	// Create monitoring handler with nil monitors (safe default)
	handler := handlers.NewMonitoringHandler(nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/monitoring/status", nil)

	handler.GetOverallStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify the "healthy" field exists and is a boolean
	healthyVal, exists := response["healthy"]
	assert.True(t, exists, "Response must contain 'healthy' field")
	_, isBool := healthyVal.(bool)
	assert.True(t, isBool, "'healthy' must be a boolean value")

	// Verify optional sub-status fields are typed correctly when present
	for _, field := range []string{
		"circuit_breakers",
		"oauth_tokens",
		"provider_health",
		"fallback_chain",
		"concurrency",
		"concurrency_alerts",
	} {
		if val, ok := response[field]; ok && val != nil {
			_, isMap := val.(map[string]interface{})
			assert.True(t, isMap, "Field %q must be a JSON object when present", field)
		}
	}
}

// TestMonitoring_CircuitBreakerState_Reporting validates that the circuit
// breaker status endpoint returns the correct service-unavailable response
// when no monitor is configured, and that the response has the expected
// error structure.
func TestMonitoring_CircuitBreakerState_Reporting(t *testing.T) {
	handler := handlers.NewMonitoringHandler(nil, nil, nil, nil, nil, nil)

	t.Run("ReturnsServiceUnavailableWhenNilMonitor", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodGet,
			"/v1/monitoring/circuit-breakers",
			nil,
		)

		handler.GetCircuitBreakerStatus(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)
		assert.Contains(t, body, "error", "Error response must have 'error' field")
	})

	t.Run("ResetAllReturnsServiceUnavailableWhenNilMonitor", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/v1/monitoring/circuit-breakers/reset-all",
			nil,
		)

		handler.ResetAllCircuitBreakers(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestMonitoring_ProviderHealth_Fields validates that the provider health
// response structure from the health handler contains the expected fields
// for latency, status, and error tracking.
func TestMonitoring_ProviderHealth_Fields(t *testing.T) {
	// We test the response type structure by marshaling/unmarshaling
	// the ProviderHealthResponse struct directly.
	sample := handlers.ProviderHealthResponse{
		ProviderID:    "openai",
		ProviderName:  "OpenAI",
		Healthy:       true,
		CircuitState:  "closed",
		FailureCount:  2,
		SuccessCount:  98,
		AvgResponseMs: 150,
		UptimePercent: 98.0,
		LastSuccessAt: "2026-03-08T10:00:00Z",
		LastFailureAt: "2026-03-08T09:00:00Z",
		LastCheckedAt: "2026-03-08T10:05:00Z",
	}

	data, err := json.Marshal(sample)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	expectedFields := []string{
		"provider_id",
		"provider_name",
		"healthy",
		"circuit_state",
		"failure_count",
		"success_count",
		"avg_response_ms",
		"uptime_percent",
		"last_success_at",
		"last_failure_at",
		"last_checked_at",
	}

	for _, field := range expectedFields {
		_, exists := decoded[field]
		assert.True(t, exists,
			"ProviderHealthResponse must contain field %q", field)
	}

	// Verify numeric fields are numbers (JSON float64)
	for _, numField := range []string{
		"failure_count",
		"success_count",
		"avg_response_ms",
		"uptime_percent",
	} {
		_, isFloat := decoded[numField].(float64)
		assert.True(t, isFloat,
			"Field %q must be a numeric value", numField)
	}
}

// TestMonitoring_StatusEndpoint_Format validates the /v1/monitoring/status
// response is valid JSON with a top-level object and the "healthy" boolean.
func TestMonitoring_StatusEndpoint_Format(t *testing.T) {
	handler := handlers.NewMonitoringHandler(nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/v1/monitoring/status",
		nil,
	)

	handler.GetOverallStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Content-Type should be JSON
	contentType := w.Header().Get("Content-Type")
	assert.True(t,
		strings.Contains(contentType, "application/json"),
		"Content-Type must be application/json, got %q", contentType)

	// Response must be a valid JSON object
	var parsed handlers.OverallMonitoringStatus
	err := json.Unmarshal(w.Body.Bytes(), &parsed)
	require.NoError(t, err, "Response must be valid JSON")

	// When all monitors are nil, status should be healthy
	assert.True(t, parsed.Healthy,
		"System with no monitors should report healthy")
}

// TestMonitoring_MetricsEndpoint_Prometheus validates that the /metrics
// endpoint returns data in Prometheus exposition format.
func TestMonitoring_MetricsEndpoint_Prometheus(t *testing.T) {
	// Create a dedicated registry to avoid global state conflicts
	registry := prometheus.NewRegistry()

	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "monitoring_test_requests_total",
		Help: "Test counter for monitoring endpoint validation",
	})
	registry.MustRegister(testCounter)
	testCounter.Inc()

	testHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "monitoring_test_latency_seconds",
		Help:    "Test histogram for monitoring endpoint validation",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
	})
	registry.MustRegister(testHistogram)
	testHistogram.Observe(0.05)
	testHistogram.Observe(0.15)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()

	// Prometheus exposition format markers
	assert.Contains(t, body, "# HELP",
		"Prometheus output must contain HELP comments")
	assert.Contains(t, body, "# TYPE",
		"Prometheus output must contain TYPE declarations")

	// Validate our test metrics appear
	assert.Contains(t, body, "monitoring_test_requests_total",
		"Counter metric must appear in output")
	assert.Contains(t, body, "monitoring_test_latency_seconds",
		"Histogram metric must appear in output")

	// Validate histogram has bucket entries
	assert.Contains(t, body, "monitoring_test_latency_seconds_bucket",
		"Histogram must include bucket lines")
	assert.Contains(t, body, "monitoring_test_latency_seconds_sum",
		"Histogram must include _sum")
	assert.Contains(t, body, "monitoring_test_latency_seconds_count",
		"Histogram must include _count")
}
