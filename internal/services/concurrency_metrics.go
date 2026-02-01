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
