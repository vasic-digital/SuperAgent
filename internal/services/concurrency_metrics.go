package services

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Package-level metrics (registered once)
var (
	concurrencyMetricsOnce                sync.Once
	concurrencyActiveRequestsGauge        *prometheus.GaugeVec
	concurrencySemaphoreAvailableGauge    *prometheus.GaugeVec
	concurrencySemaphoreTotalGauge        *prometheus.GaugeVec
	concurrencySemaphoreAcquiredGauge     *prometheus.GaugeVec
	concurrencyAcquisitionTimeoutsCounter *prometheus.CounterVec
	concurrencyAcquisitionErrorsCounter   *prometheus.CounterVec

	// Alert manager metrics
	concurrencyAlertDeliveryTotal       *prometheus.CounterVec
	concurrencyAlertDeliveryErrorsTotal *prometheus.CounterVec
	concurrencyAlertRetryAttemptsTotal  *prometheus.CounterVec
	concurrencyAlertRetrySuccessTotal   *prometheus.CounterVec
	concurrencyAlertRetryQueueSize      *prometheus.GaugeVec
	concurrencyAlertDeadLetterQueueSize *prometheus.GaugeVec
)

func initConcurrencyMetrics() {
	concurrencyMetricsOnce.Do(func() {
		concurrencyActiveRequestsGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_active_requests",
				Help: "Number of active requests per provider",
			},
			[]string{"provider"},
		)

		concurrencySemaphoreAvailableGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_semaphore_available",
				Help: "Number of available semaphore permits per provider",
			},
			[]string{"provider"},
		)

		concurrencySemaphoreTotalGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_semaphore_total",
				Help: "Total semaphore permits (max concurrent requests) per provider",
			},
			[]string{"provider"},
		)

		concurrencySemaphoreAcquiredGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_semaphore_acquired",
				Help: "Number of acquired semaphore permits per provider",
			},
			[]string{"provider"},
		)

		concurrencyAcquisitionTimeoutsCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_acquisition_timeouts_total",
				Help: "Total number of semaphore acquisition timeouts per provider",
			},
			[]string{"provider"},
		)

		concurrencyAcquisitionErrorsCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_acquisition_errors_total",
				Help: "Total number of semaphore acquisition errors per provider",
			},
			[]string{"provider"},
		)

		// Alert manager metrics
		concurrencyAlertDeliveryTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_alert_delivery_total",
				Help: "Total number of alert delivery attempts per channel",
			},
			[]string{"channel", "provider", "alert_type"},
		)

		concurrencyAlertDeliveryErrorsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_alert_delivery_errors_total",
				Help: "Total number of alert delivery errors per channel",
			},
			[]string{"channel", "provider", "alert_type"},
		)

		concurrencyAlertRetryAttemptsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_alert_retry_attempts_total",
				Help: "Total number of retry attempts per channel",
			},
			[]string{"channel", "provider", "alert_type"},
		)

		concurrencyAlertRetrySuccessTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_alert_retry_success_total",
				Help: "Total number of successful retries per channel",
			},
			[]string{"channel", "provider", "alert_type"},
		)

		concurrencyAlertRetryQueueSize = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_alert_retry_queue_size",
				Help: "Current size of retry queue per channel",
			},
			[]string{"channel"},
		)

		concurrencyAlertDeadLetterQueueSize = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_alert_dead_letter_queue_size",
				Help: "Current size of dead letter queue per channel",
			},
			[]string{"channel"},
		)
	})
}

// UpdateConcurrencyMetrics updates all concurrency-related metrics for a provider
// This should be called periodically or when state changes
// totalPermits: maximum concurrent requests (semaphore capacity), 0 means no semaphore
// acquiredPermits: number of permits currently acquired (should be <= totalPermits)
// activeRequests: number of active requests (may be equal to acquiredPermits if semaphore weight is 1)
func UpdateConcurrencyMetrics(provider string, totalPermits, acquiredPermits, activeRequests int64) {
	// Initialize metrics if not already done
	initConcurrencyMetrics()

	// Update gauges
	concurrencyActiveRequestsGauge.WithLabelValues(provider).Set(float64(activeRequests))
	if totalPermits > 0 {
		concurrencySemaphoreTotalGauge.WithLabelValues(provider).Set(float64(totalPermits))
		concurrencySemaphoreAcquiredGauge.WithLabelValues(provider).Set(float64(acquiredPermits))
		concurrencySemaphoreAvailableGauge.WithLabelValues(provider).Set(float64(totalPermits - acquiredPermits))
	} else {
		// No semaphore - set to zero to indicate unlimited capacity
		concurrencySemaphoreTotalGauge.WithLabelValues(provider).Set(0)
		concurrencySemaphoreAcquiredGauge.WithLabelValues(provider).Set(0)
		concurrencySemaphoreAvailableGauge.WithLabelValues(provider).Set(0)
	}
}

// RecordAcquisitionTimeout records a semaphore acquisition timeout
func RecordAcquisitionTimeout(provider string) {
	initConcurrencyMetrics()
	concurrencyAcquisitionTimeoutsCounter.WithLabelValues(provider).Inc()
}

// RecordAcquisitionError records a semaphore acquisition error (non-timeout)
func RecordAcquisitionError(provider string) {
	initConcurrencyMetrics()
	concurrencyAcquisitionErrorsCounter.WithLabelValues(provider).Inc()
}

// Alert manager metrics functions

// RecordAlertDelivery records an alert delivery attempt
func RecordAlertDelivery(channel, provider, alertType string) {
	initConcurrencyMetrics()
	concurrencyAlertDeliveryTotal.WithLabelValues(channel, provider, alertType).Inc()
}

// RecordAlertDeliveryError records an alert delivery error
func RecordAlertDeliveryError(channel, provider, alertType string) {
	initConcurrencyMetrics()
	concurrencyAlertDeliveryErrorsTotal.WithLabelValues(channel, provider, alertType).Inc()
}

// RecordRetryAttempt records a retry attempt
func RecordRetryAttempt(channel, provider, alertType string) {
	initConcurrencyMetrics()
	concurrencyAlertRetryAttemptsTotal.WithLabelValues(channel, provider, alertType).Inc()
}

// RecordRetrySuccess records a successful retry
func RecordRetrySuccess(channel, provider, alertType string) {
	initConcurrencyMetrics()
	concurrencyAlertRetrySuccessTotal.WithLabelValues(channel, provider, alertType).Inc()
}

// UpdateRetryQueueSize updates the retry queue size gauge
func UpdateRetryQueueSize(channel string, size int) {
	initConcurrencyMetrics()
	concurrencyAlertRetryQueueSize.WithLabelValues(channel).Set(float64(size))
}

// UpdateDeadLetterQueueSize updates the dead letter queue size gauge
func UpdateDeadLetterQueueSize(channel string, size int) {
	initConcurrencyMetrics()
	concurrencyAlertDeadLetterQueueSize.WithLabelValues(channel).Set(float64(size))
}
