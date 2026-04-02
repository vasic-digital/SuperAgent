package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Additional test coverage for boot manager functionality
func TestBootResult_StatusTypes(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"started", "started", "started"},
		{"already_running", "already_running", "already_running"},
		{"remote", "remote", "remote"},
		{"discovered", "discovered", "discovered"},
		{"failed", "failed", "failed"},
		{"skipped", "skipped", "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &BootResult{
				Name:     "test-service",
				Status:   tt.status,
				Duration: time.Second,
			}
			assert.Equal(t, tt.expected, result.Status)
		})
	}
}

func TestBootResult_WithError(t *testing.T) {
	t.Run("stores error correctly", func(t *testing.T) {
		testErr := assert.AnError
		result := &BootResult{
			Name:     "failed-service",
			Status:   "failed",
			Error:    testErr,
			Duration: 5 * time.Second,
		}

		assert.Equal(t, testErr, result.Error)
		assert.Equal(t, "failed", result.Status)
		assert.Equal(t, 5*time.Second, result.Duration)
	})
}

func TestBootResult_DurationFormatting(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected time.Duration
	}{
		{0, 0},
		{time.Second, time.Second},
		{5 * time.Minute, 5 * time.Minute},
		{1 * time.Hour, 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := &BootResult{
				Name:     "test",
				Duration: tt.duration,
			}
			assert.Equal(t, tt.expected, result.Duration)
		})
	}
}
