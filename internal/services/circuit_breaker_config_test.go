package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerConfig_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"circuit breaker enabled", true},
		{"circuit breaker disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CircuitBreakerConfig{
				Enabled: tt.enabled,
			}
			assert.Equal(t, tt.enabled, config.Enabled)
		})
	}
}

func TestCircuitBreakerConfig_Thresholds(t *testing.T) {
	t.Run("sets failure and success thresholds", func(t *testing.T) {
		config := CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 5,
			SuccessThreshold: 3,
		}

		assert.Equal(t, 5, config.FailureThreshold)
		assert.Equal(t, 3, config.SuccessThreshold)
	})
}

func TestCircuitBreakerConfig_RecoveryTimeout(t *testing.T) {
	tests := []struct {
		name            string
		recoveryTimeout time.Duration
	}{
		{"30 second recovery", 30 * time.Second},
		{"1 minute recovery", 1 * time.Minute},
		{"5 minute recovery", 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CircuitBreakerConfig{
				RecoveryTimeout: tt.recoveryTimeout,
			}
			assert.Equal(t, tt.recoveryTimeout, config.RecoveryTimeout)
		})
	}
}
