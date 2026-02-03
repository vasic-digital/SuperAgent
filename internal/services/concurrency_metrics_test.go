package services

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// resetConcurrencyMetrics resets the package-level metrics for testing.
// This function should only be used in tests.
// CONCURRENCY FIX: Use mutex to protect metric variable access during reset
func resetConcurrencyMetrics() {
	// Acquire write lock to prevent other goroutines from accessing metrics during reset
	concurrencyMetricsMu.Lock()
	defer concurrencyMetricsMu.Unlock()

	// Unregister all metrics from the default registry first
	if concurrencyActiveRequestsGauge != nil {
		prometheus.Unregister(concurrencyActiveRequestsGauge)
	}
	if concurrencySemaphoreAvailableGauge != nil {
		prometheus.Unregister(concurrencySemaphoreAvailableGauge)
	}
	if concurrencySemaphoreTotalGauge != nil {
		prometheus.Unregister(concurrencySemaphoreTotalGauge)
	}
	if concurrencySemaphoreAcquiredGauge != nil {
		prometheus.Unregister(concurrencySemaphoreAcquiredGauge)
	}
	if concurrencyAcquisitionTimeoutsCounter != nil {
		prometheus.Unregister(concurrencyAcquisitionTimeoutsCounter)
	}
	if concurrencyAcquisitionErrorsCounter != nil {
		prometheus.Unregister(concurrencyAcquisitionErrorsCounter)
	}
	if concurrencyAlertDeliveryTotal != nil {
		prometheus.Unregister(concurrencyAlertDeliveryTotal)
	}
	if concurrencyAlertDeliveryErrorsTotal != nil {
		prometheus.Unregister(concurrencyAlertDeliveryErrorsTotal)
	}
	if concurrencyAlertRetryAttemptsTotal != nil {
		prometheus.Unregister(concurrencyAlertRetryAttemptsTotal)
	}
	if concurrencyAlertRetrySuccessTotal != nil {
		prometheus.Unregister(concurrencyAlertRetrySuccessTotal)
	}
	if concurrencyAlertRetryQueueSize != nil {
		prometheus.Unregister(concurrencyAlertRetryQueueSize)
	}
	if concurrencyAlertDeadLetterQueueSize != nil {
		prometheus.Unregister(concurrencyAlertDeadLetterQueueSize)
	}
	if concurrencyAlertTotal != nil {
		prometheus.Unregister(concurrencyAlertTotal)
	}
	if concurrencyAlertThresholdBreaches != nil {
		prometheus.Unregister(concurrencyAlertThresholdBreaches)
	}
	if concurrencyAlertCircuitBreakerState != nil {
		prometheus.Unregister(concurrencyAlertCircuitBreakerState)
	}
	if concurrencyAlertRateLimitHits != nil {
		prometheus.Unregister(concurrencyAlertRateLimitHits)
	}
	if concurrencyAlertEscalationLevel != nil {
		prometheus.Unregister(concurrencyAlertEscalationLevel)
	}

	// Reset the sync.Once so metrics can be re-initialized
	concurrencyMetricsOnce = sync.Once{}

	// Set all metric variables to nil so they can be recreated
	concurrencyActiveRequestsGauge = nil
	concurrencySemaphoreAvailableGauge = nil
	concurrencySemaphoreTotalGauge = nil
	concurrencySemaphoreAcquiredGauge = nil
	concurrencyAcquisitionTimeoutsCounter = nil
	concurrencyAcquisitionErrorsCounter = nil

	concurrencyAlertDeliveryTotal = nil
	concurrencyAlertDeliveryErrorsTotal = nil
	concurrencyAlertRetryAttemptsTotal = nil
	concurrencyAlertRetrySuccessTotal = nil
	concurrencyAlertRetryQueueSize = nil
	concurrencyAlertDeadLetterQueueSize = nil

	concurrencyAlertTotal = nil
	concurrencyAlertThresholdBreaches = nil
	concurrencyAlertCircuitBreakerState = nil
	concurrencyAlertRateLimitHits = nil
	concurrencyAlertEscalationLevel = nil
}

// TestConcurrencyMetricsInitialization tests that metrics can be initialized
func TestConcurrencyMetricsInitialization(t *testing.T) {
	// Save original registry
	originalRegistry := prometheus.DefaultRegisterer
	// Replace with a new registry for testing
	testRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = testRegistry
	defer func() {
		prometheus.DefaultRegisterer = originalRegistry
	}()

	// Reset metrics to ensure fresh initialization
	resetConcurrencyMetrics()

	// Call a metric function to trigger initialization
	RecordAlertHandled("test_type", "test_provider", "info")

	// Verify metric was registered by attempting to gather metrics
	metrics, err := testRegistry.Gather()
	assert.NoError(t, err, "Should be able to gather metrics")

	// Find our metric
	found := false
	for _, mf := range metrics {
		if mf.GetName() == "helixagent_concurrency_alerts_total" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have registered helixagent_concurrency_alerts_total metric")
}

// TestConcurrencyMetricsFunctions tests that all metric functions work without panicking
func TestConcurrencyMetricsFunctions(t *testing.T) {
	// Save original registry
	originalRegistry := prometheus.DefaultRegisterer
	testRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = testRegistry
	defer func() {
		prometheus.DefaultRegisterer = originalRegistry
	}()

	// Reset metrics
	resetConcurrencyMetrics()

	// Test each metric function
	RecordAlertHandled("test_type", "test_provider", "info")
	RecordThresholdBreach("warning", "logging", "test_provider")
	UpdateCircuitBreakerState("webhook", 2) // open
	RecordRateLimitHit("email")
	UpdateEscalationLevel("high_saturation", "test_provider", "key123", 3)

	// Verify all metrics were registered
	metrics, err := testRegistry.Gather()
	assert.NoError(t, err)

	metricNames := make(map[string]bool)
	for _, mf := range metrics {
		metricNames[mf.GetName()] = true
	}

	// Check expected metrics exist
	expectedMetrics := []string{
		"helixagent_concurrency_alerts_total",
		"helixagent_concurrency_alert_threshold_breaches_total",
		"helixagent_concurrency_alert_circuit_breaker_state",
		"helixagent_concurrency_alert_rate_limit_hits_total",
		"helixagent_concurrency_alert_escalation_level",
	}

	for _, expected := range expectedMetrics {
		assert.True(t, metricNames[expected], "Metric %s should be registered", expected)
	}
}
