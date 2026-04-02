package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckConfig_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"health check enabled", true},
		{"health check disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HealthCheckConfig{
				Enabled: tt.enabled,
			}
			assert.Equal(t, tt.enabled, config.Enabled)
		})
	}
}

func TestHealthCheckConfig_Intervals(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		timeout  time.Duration
	}{
		{
			name:     "standard intervals",
			interval: 30 * time.Second,
			timeout:  5 * time.Second,
		},
		{
			name:     "aggressive monitoring",
			interval: 10 * time.Second,
			timeout:  2 * time.Second,
		},
		{
			name:     "relaxed monitoring",
			interval: 5 * time.Minute,
			timeout:  30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HealthCheckConfig{
				Interval: tt.interval,
				Timeout:  tt.timeout,
			}
			assert.Equal(t, tt.interval, config.Interval)
			assert.Equal(t, tt.timeout, config.Timeout)
		})
	}
}

func TestHealthCheckConfig_FailureThreshold(t *testing.T) {
	tests := []struct {
		name             string
		failureThreshold int
	}{
		{"single failure", 1},
		{"three failures", 3},
		{"five failures", 5},
		{"ten failures", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HealthCheckConfig{
				FailureThreshold: tt.failureThreshold,
			}
			assert.Equal(t, tt.failureThreshold, config.FailureThreshold)
		})
	}
}
