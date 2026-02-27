package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Metrics collector for HelixAgent
type Collector struct {
	// Request metrics
	RequestDuration *prometheus.HistogramVec
	RequestCount    *prometheus.CounterVec
	RequestSize     *prometheus.HistogramVec
	ResponseSize    *prometheus.HistogramVec

	// Provider metrics
	ProviderLatency *prometheus.HistogramVec
	ProviderErrors  *prometheus.CounterVec
	ProviderTokens  *prometheus.CounterVec

	// Debate metrics
	DebateDuration  *prometheus.HistogramVec
	DebateRounds    *prometheus.HistogramVec
	DebateConsensus *prometheus.CounterVec

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
	CacheSize   prometheus.Gauge

	// Resource metrics
	MemoryUsage    prometheus.Gauge
	GoroutineCount prometheus.Gauge
	CPUUsage       prometheus.Gauge
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	c := &Collector{
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint", "status"},
		),

		ProviderLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "llm_provider_latency_seconds",
				Help:    "LLM provider latency in seconds",
				Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"provider", "model"},
		),

		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total cache hits",
			},
			[]string{"cache_type"},
		),

		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total cache misses",
			},
			[]string{"cache_type"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(c.RequestDuration)
	prometheus.MustRegister(c.ProviderLatency)
	prometheus.MustRegister(c.CacheHits)
	prometheus.MustRegister(c.CacheMisses)

	return c
}

// Handler returns HTTP handler for metrics
func (c *Collector) Handler() http.Handler {
	return promhttp.Handler()
}
