package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// Package-level metrics (registered once)
var (
	cmMetricsOnce             sync.Once
	cmHighConcurrencyGauge    *prometheus.GaugeVec
	cmSaturationGauge         *prometheus.GaugeVec
	cmConcurrencyAlertsTotal  prometheus.Counter
	cmBlockedRequestsCounter  *prometheus.CounterVec
	cmHighConcurrencyDuration *prometheus.HistogramVec
)

func initCMMetrics() {
	cmMetricsOnce.Do(func() {
		cmHighConcurrencyGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_high_usage",
				Help: "Indicates high concurrency usage (1=high, 0=normal)",
			},
			[]string{"provider"},
		)

		cmSaturationGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "helixagent_concurrency_saturation",
				Help: "Concurrency saturation percentage (0-100)",
			},
			[]string{"provider"},
		)

		cmConcurrencyAlertsTotal = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_alerts_sent_total",
				Help: "Total number of concurrency alerts sent to listeners",
			},
		)

		cmBlockedRequestsCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "helixagent_concurrency_blocked_requests_total",
				Help: "Total number of requests blocked due to concurrency limits",
			},
			[]string{"provider"},
		)

		cmHighConcurrencyDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "helixagent_high_concurrency_duration_seconds",
				Help:    "Duration of high concurrency periods",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"provider"},
		)
	})
}

// ConcurrencyMonitor monitors concurrency usage and provides alerts
type ConcurrencyMonitor struct {
	mu                sync.RWMutex
	registry          *ProviderRegistry
	logger            *logrus.Logger
	checkInterval     time.Duration
	alertThreshold    float64 // Saturation percentage to trigger warning alert (0-100)
	criticalThreshold float64 // Saturation percentage to trigger critical alert (0-100)
	listeners         []ConcurrencyAlertListener
	stopCh            chan struct{}
	running           bool

	// Track high concurrency periods
	highConcurrencyStart map[string]time.Time
	highConcurrencyState map[string]bool
}

// ConcurrencyAlertListener is called when concurrency alerts occur
type ConcurrencyAlertListener func(alert ConcurrencyAlert)

// ConcurrencyAlert represents a concurrency alert
type ConcurrencyAlert struct {
	Type            string                       `json:"type"`
	Provider        string                       `json:"provider,omitempty"`
	Message         string                       `json:"message"`
	Timestamp       time.Time                    `json:"timestamp"`
	Saturation      float64                      `json:"saturation,omitempty"` // Percentage (0-100)
	ActiveRequests  int64                        `json:"active_requests,omitempty"`
	TotalPermits    int64                        `json:"total_permits,omitempty"`
	Available       int64                        `json:"available,omitempty"`
	AllStats        map[string]*ConcurrencyStats `json:"all_stats,omitempty"`
	Severity        AlertSeverity                `json:"severity,omitempty"`
	EscalationLevel int                          `json:"escalation_level,omitempty"`
}

// ConcurrencyMonitorConfig configures the monitor
type ConcurrencyMonitorConfig struct {
	CheckInterval     time.Duration
	AlertThreshold    float64 // Saturation percentage (0-100) for warning alerts
	CriticalThreshold float64 // Saturation percentage (0-100) for critical alerts
}

// DefaultConcurrencyMonitorConfig returns default configuration
func DefaultConcurrencyMonitorConfig() ConcurrencyMonitorConfig {
	return ConcurrencyMonitorConfig{
		CheckInterval:     15 * time.Second,
		AlertThreshold:    80.0, // Warning when saturation >= 80%
		CriticalThreshold: 95.0, // Critical when saturation >= 95%
	}
}

// NewConcurrencyMonitor creates a new concurrency monitor
func NewConcurrencyMonitor(registry *ProviderRegistry, logger *logrus.Logger, config ConcurrencyMonitorConfig) *ConcurrencyMonitor {
	// Initialize package-level metrics (idempotent)
	initCMMetrics()

	return &ConcurrencyMonitor{
		registry:             registry,
		logger:               logger,
		checkInterval:        config.CheckInterval,
		alertThreshold:       config.AlertThreshold,
		criticalThreshold:    config.CriticalThreshold,
		listeners:            make([]ConcurrencyAlertListener, 0),
		stopCh:               make(chan struct{}),
		highConcurrencyStart: make(map[string]time.Time),
		highConcurrencyState: make(map[string]bool),
	}
}

// AddAlertListener adds a listener for alerts
func (cm *ConcurrencyMonitor) AddAlertListener(listener ConcurrencyAlertListener) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.listeners = append(cm.listeners, listener)
}

// Start starts the monitoring loop
func (cm *ConcurrencyMonitor) Start(ctx context.Context) {
	cm.mu.Lock()
	if cm.running {
		cm.mu.Unlock()
		return
	}
	cm.running = true
	cm.stopCh = make(chan struct{})
	cm.mu.Unlock()

	cm.logger.Info("Concurrency monitor started")

	ticker := time.NewTicker(cm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Concurrency monitor stopped (context cancelled)")
			return
		case <-cm.stopCh:
			cm.logger.Info("Concurrency monitor stopped")
			return
		case <-ticker.C:
			cm.checkConcurrency()
		}
	}
}

// Stop stops the monitoring loop
func (cm *ConcurrencyMonitor) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.running {
		close(cm.stopCh)
		cm.running = false
	}
}

// checkConcurrency checks all providers' concurrency usage and sends alerts if needed
func (cm *ConcurrencyMonitor) checkConcurrency() {
	if cm.registry == nil {
		return
	}

	stats := cm.registry.GetAllConcurrencyStats()
	highSaturationCount := 0
	criticalSaturationCount := 0
	maxSaturation := 0.0
	totalProviders := len(stats)

	for providerID, stat := range stats {
		// Skip providers without semaphores
		if !stat.HasSemaphore || stat.TotalPermits == 0 {
			continue
		}

		// Calculate saturation percentage
		saturation := float64(0)
		if stat.TotalPermits > 0 {
			saturation = float64(stat.AcquiredPermits) / float64(stat.TotalPermits) * 100.0
		}

		// Track max saturation and critical counts
		if saturation > maxSaturation {
			maxSaturation = saturation
		}
		if saturation >= cm.criticalThreshold {
			criticalSaturationCount++
		}

		// Update Prometheus metrics
		cmSaturationGauge.WithLabelValues(providerID).Set(saturation)

		// Check if saturation is high
		isHigh := saturation >= cm.alertThreshold
		if isHigh {
			highSaturationCount++
			cmHighConcurrencyGauge.WithLabelValues(providerID).Set(1)

			// Track high concurrency duration
			cm.mu.Lock()
			if !cm.highConcurrencyState[providerID] {
				// Transition to high concurrency
				cm.highConcurrencyState[providerID] = true
				cm.highConcurrencyStart[providerID] = time.Now()
			}
			cm.mu.Unlock()
		} else {
			cmHighConcurrencyGauge.WithLabelValues(providerID).Set(0)

			// End high concurrency period if needed
			cm.mu.Lock()
			if cm.highConcurrencyState[providerID] {
				// Transition out of high concurrency
				duration := time.Since(cm.highConcurrencyStart[providerID])
				cmHighConcurrencyDuration.WithLabelValues(providerID).Observe(duration.Seconds())
				cm.highConcurrencyState[providerID] = false
				delete(cm.highConcurrencyStart, providerID)

				cm.logger.WithFields(logrus.Fields{
					"provider":   providerID,
					"duration":   duration,
					"saturation": saturation,
				}).Info("High concurrency period ended")
			}
			cm.mu.Unlock()
		}

		// Log warning for high saturation
		if isHigh {
			cm.logger.WithFields(logrus.Fields{
				"provider":   providerID,
				"saturation": fmt.Sprintf("%.1f%%", saturation),
				"acquired":   stat.AcquiredPermits,
				"total":      stat.TotalPermits,
				"available":  stat.AvailablePermits,
				"active":     stat.ActiveRequests,
			}).Warn("High concurrency saturation detected")
		}
	}

	// Send alert if any provider has high saturation
	if highSaturationCount > 0 {
		// Determine severity based on critical saturation count
		severity := SeverityWarning
		if criticalSaturationCount > 0 {
			severity = SeverityCritical
		}
		cm.sendAlert(ConcurrencyAlert{
			Type:            "high_saturation",
			Message:         fmt.Sprintf("%d provider(s) have high concurrency saturation (%d critical)", highSaturationCount, criticalSaturationCount),
			Timestamp:       time.Now(),
			Saturation:      maxSaturation,
			AllStats:        stats,
			Severity:        severity,
			EscalationLevel: 0,
		})
	}

	cm.logger.WithFields(logrus.Fields{
		"total":           totalProviders,
		"high_saturation": highSaturationCount,
		"threshold":       fmt.Sprintf("%.1f%%", cm.alertThreshold),
	}).Debug("Concurrency check completed")
}

// sendAlert sends an alert to all listeners
func (cm *ConcurrencyMonitor) sendAlert(alert ConcurrencyAlert) {
	cmConcurrencyAlertsTotal.Inc()

	cm.mu.RLock()
	listeners := cm.listeners
	cm.mu.RUnlock()

	for _, listener := range listeners {
		go listener(alert)
	}

	cm.logger.WithFields(logrus.Fields{
		"type":       alert.Type,
		"message":    alert.Message,
		"count":      len(alert.AllStats),
		"severity":   alert.Severity,
		"saturation": alert.Saturation,
	}).Error("Concurrency alert triggered")
}

// GetStatus returns the current concurrency status of all providers
func (cm *ConcurrencyMonitor) GetStatus() ConcurrencyStatus {
	if cm.registry == nil {
		return ConcurrencyStatus{
			Healthy:   true,
			Providers: make(map[string]ConcurrencyProviderStatus),
		}
	}

	stats := cm.registry.GetAllConcurrencyStats()
	providers := make(map[string]ConcurrencyProviderStatus)
	highSaturationCount := 0
	totalWithSemaphores := 0

	for providerID, stat := range stats {
		status := ConcurrencyProviderStatus{
			Provider:          stat.Provider,
			HasSemaphore:      stat.HasSemaphore,
			TotalPermits:      stat.TotalPermits,
			AcquiredPermits:   stat.AcquiredPermits,
			ActiveRequests:    stat.ActiveRequests,
			AvailablePermits:  stat.AvailablePermits,
			SemaphoreExists:   stat.SemaphoreExists,
			SemaphoreCapacity: stat.SemaphoreCapacity,
		}

		// Calculate saturation
		if stat.HasSemaphore && stat.TotalPermits > 0 {
			status.Saturation = float64(stat.AcquiredPermits) / float64(stat.TotalPermits) * 100.0
			totalWithSemaphores++
			if status.Saturation >= cm.alertThreshold {
				status.HighSaturation = true
				highSaturationCount++
			}
		}

		providers[providerID] = status
	}

	return ConcurrencyStatus{
		Healthy:             highSaturationCount == 0,
		HighSaturationCount: highSaturationCount,
		TotalWithSemaphores: totalWithSemaphores,
		TotalProviders:      len(stats),
		AlertThreshold:      cm.alertThreshold,
		Providers:           providers,
		CheckedAt:           time.Now(),
	}
}

// ConcurrencyStatus represents the overall concurrency status
type ConcurrencyStatus struct {
	Healthy             bool                                 `json:"healthy"`
	HighSaturationCount int                                  `json:"high_saturation_count"`
	TotalWithSemaphores int                                  `json:"total_with_semaphores"`
	TotalProviders      int                                  `json:"total_providers"`
	AlertThreshold      float64                              `json:"alert_threshold"`
	Providers           map[string]ConcurrencyProviderStatus `json:"providers"`
	CheckedAt           time.Time                            `json:"checked_at"`
}

// ConcurrencyProviderStatus represents a single provider's concurrency status
type ConcurrencyProviderStatus struct {
	Provider          string  `json:"provider"`
	HasSemaphore      bool    `json:"has_semaphore"`
	TotalPermits      int64   `json:"total_permits"`
	AcquiredPermits   int64   `json:"acquired_permits"`
	ActiveRequests    int64   `json:"active_requests"`
	AvailablePermits  int64   `json:"available_permits"`
	SemaphoreExists   bool    `json:"semaphore_exists"`
	SemaphoreCapacity int64   `json:"semaphore_capacity"`
	Saturation        float64 `json:"saturation"`
	HighSaturation    bool    `json:"high_saturation"`
}

// RecordBlockedRequest records a request that was blocked due to concurrency limits
func (cm *ConcurrencyMonitor) RecordBlockedRequest(provider string) {
	initCMMetrics()
	cmBlockedRequestsCounter.WithLabelValues(provider).Inc()
}

// ResetHighConcurrencyTracking resets high concurrency tracking for a provider
func (cm *ConcurrencyMonitor) ResetHighConcurrencyTracking(provider string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.highConcurrencyStart, provider)
	delete(cm.highConcurrencyState, provider)
	cmHighConcurrencyGauge.WithLabelValues(provider).Set(0)
}

// ResetAllHighConcurrencyTracking resets high concurrency tracking for all providers
func (cm *ConcurrencyMonitor) ResetAllHighConcurrencyTracking() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.highConcurrencyStart = make(map[string]time.Time)
	cm.highConcurrencyState = make(map[string]bool)

	// Reset all gauge values to 0
	if cmHighConcurrencyGauge != nil {
		cmHighConcurrencyGauge.Reset()
	}
}
