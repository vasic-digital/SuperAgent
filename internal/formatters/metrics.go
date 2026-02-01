package formatters

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Formatter metrics
	formatterRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "formatter_requests_total",
			Help: "Total number of formatter requests",
		},
		[]string{"formatter_name", "language", "success"},
	)

	formatterRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "formatter_request_duration_seconds",
			Help:    "Request duration for formatter operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"formatter_name", "language"},
	)

	formatterBytesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "formatter_bytes_processed_total",
			Help: "Total number of bytes processed by formatters",
		},
		[]string{"formatter_name", "language"},
	)

	formatterActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "formatter_active_requests",
			Help: "Number of currently active formatter requests",
		},
		[]string{"formatter_name", "language"},
	)

	formatterCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "formatter_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"formatter_name", "language"},
	)

	formatterCacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "formatter_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"formatter_name", "language"},
	)
)

// RecordRequestStart records the start of a formatter request.
// Returns a function to record the completion.
func RecordRequestStart(formatterName, language string) func(success bool, duration time.Duration, bytesProcessed int) {
	formatterActiveRequests.WithLabelValues(formatterName, language).Inc()

	return func(success bool, duration time.Duration, bytesProcessed int) {
		// Record request completion
		formatterRequestsTotal.WithLabelValues(formatterName, language, successToString(success)).Inc()
		formatterRequestDuration.WithLabelValues(formatterName, language).Observe(duration.Seconds())
		formatterBytesProcessed.WithLabelValues(formatterName, language).Add(float64(bytesProcessed))
		formatterActiveRequests.WithLabelValues(formatterName, language).Dec()
	}
}

// RecordCacheHit records a cache hit.
func RecordCacheHit(formatterName, language string) {
	formatterCacheHits.WithLabelValues(formatterName, language).Inc()
}

// RecordCacheMiss records a cache miss.
func RecordCacheMiss(formatterName, language string) {
	formatterCacheMisses.WithLabelValues(formatterName, language).Inc()
}

// successToString converts boolean success to string label.
func successToString(success bool) string {
	if success {
		return "true"
	}
	return "false"
}
