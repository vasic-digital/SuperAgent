package formatters

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthChecker performs health checks on formatters
type HealthChecker struct {
	registry *FormatterRegistry
	logger   *logrus.Logger
	timeout  time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(registry *FormatterRegistry, logger *logrus.Logger, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		logger:   logger,
		timeout:  timeout,
	}
}

// CheckAll performs health checks on all formatters
func (h *HealthChecker) CheckAll(ctx context.Context) *HealthReport {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	results := h.registry.HealthCheckAll(ctx)

	healthy := 0
	unhealthy := 0
	formatterResults := make([]FormatterHealthResult, 0, len(results))

	for name, err := range results {
		result := FormatterHealthResult{
			Name:    name,
			Healthy: err == nil,
		}

		if err != nil {
			result.Error = err.Error()
			unhealthy++
		} else {
			healthy++
		}

		formatterResults = append(formatterResults, result)
	}

	return &HealthReport{
		TotalFormatters:  len(results),
		HealthyCount:     healthy,
		UnhealthyCount:   unhealthy,
		FormatterResults: formatterResults,
		Timestamp:        time.Now(),
	}
}

// Check performs a health check on a single formatter
func (h *HealthChecker) Check(ctx context.Context, name string) (*FormatterHealthResult, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	formatter, err := h.registry.Get(name)
	if err != nil {
		return nil, fmt.Errorf("formatter not found: %w", err)
	}

	start := time.Now()
	err = formatter.HealthCheck(ctx)
	duration := time.Since(start)

	result := &FormatterHealthResult{
		Name:     name,
		Healthy:  err == nil,
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// HealthReport represents the health status of all formatters
type HealthReport struct {
	TotalFormatters  int
	HealthyCount     int
	UnhealthyCount   int
	FormatterResults []FormatterHealthResult
	Timestamp        time.Time
}

// FormatterHealthResult represents the health status of a single formatter
type FormatterHealthResult struct {
	Name     string
	Healthy  bool
	Error    string
	Duration time.Duration
}

// IsHealthy returns whether all formatters are healthy
func (h *HealthReport) IsHealthy() bool {
	return h.UnhealthyCount == 0
}

// HealthPercentage returns the percentage of healthy formatters
func (h *HealthReport) HealthPercentage() float64 {
	if h.TotalFormatters == 0 {
		return 100.0
	}
	return (float64(h.HealthyCount) / float64(h.TotalFormatters)) * 100.0
}
