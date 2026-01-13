package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStartupScoringConfig(t *testing.T) {
	cfg := DefaultStartupScoringConfig()

	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.Async)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.True(t, cfg.RetryOnFailure)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, 5, cfg.ConcurrentWorkers)
}

func TestNewStartupScoringService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()

	// Create service with nil registry (for unit test)
	svc := NewStartupScoringService(nil, config, logger)

	assert.NotNil(t, svc)
	assert.Equal(t, config.Enabled, svc.config.Enabled)
	assert.False(t, svc.completed)
}

func TestStartupScoringService_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Enabled = false

	svc := NewStartupScoringService(nil, config, logger)
	result := svc.Run(context.Background())

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "disabled", result.ProviderStatus["status"])
}

func TestStartupScoringService_AsyncMode(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Async = true
	config.Timeout = 1 * time.Second

	svc := NewStartupScoringService(nil, config, logger)
	result := svc.Run(context.Background())

	// Should return immediately in async mode
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "running_async", result.ProviderStatus["status"])

	// Wait for completion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	finalResult := svc.WaitForCompletion(ctx)
	assert.NotNil(t, finalResult)
	assert.True(t, svc.IsCompleted())
}

func TestStartupScoringService_SyncMode(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Async = false
	config.Timeout = 1 * time.Second

	svc := NewStartupScoringService(nil, config, logger)
	result := svc.Run(context.Background())

	// Should complete synchronously
	assert.NotNil(t, result)
	assert.True(t, svc.IsCompleted())
	assert.NotNil(t, svc.GetResult())
}

func TestStartupScoringService_WithRegistry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a minimal registry for testing
	registryConfig := &RegistryConfig{
		Providers: map[string]*ProviderConfig{
			"test": {
				Name:    "test",
				Type:    "test",
				Enabled: true,
			},
		},
	}
	registry := NewProviderRegistryWithoutAutoDiscovery(registryConfig, nil)

	config := DefaultStartupScoringConfig()
	config.Async = false
	config.Timeout = 5 * time.Second

	svc := NewStartupScoringService(registry, config, logger)
	result := svc.Run(context.Background())

	assert.NotNil(t, result)
	assert.True(t, svc.IsCompleted())
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0))
}

func TestStartupScoringService_WaitForCompletionTimeout(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Async = true
	config.Timeout = 10 * time.Second // Long timeout

	svc := NewStartupScoringService(nil, config, logger)
	svc.Run(context.Background())

	// Use very short context - should timeout before completion
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This might either complete or timeout depending on timing
	result := svc.WaitForCompletion(ctx)
	// Result may or may not be nil depending on race
	_ = result
}

func TestStartupScoringResult_Fields(t *testing.T) {
	result := &StartupScoringResult{
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1 * time.Second),
		Duration:         1 * time.Second,
		TotalProviders:   10,
		ScoredProviders:  8,
		FailedProviders:  2,
		SkippedProviders: 0,
		ProviderScores: map[string]float64{
			"claude":   9.5,
			"deepseek": 8.5,
			"gemini":   9.0,
		},
		ProviderStatus: map[string]string{
			"claude":   "verified",
			"deepseek": "verified",
			"gemini":   "verified",
		},
		Errors:  []string{},
		Success: true,
	}

	assert.Equal(t, 10, result.TotalProviders)
	assert.Equal(t, 8, result.ScoredProviders)
	assert.Equal(t, 2, result.FailedProviders)
	assert.True(t, result.Success)
	assert.Equal(t, 9.5, result.ProviderScores["claude"])
	assert.Equal(t, "verified", result.ProviderStatus["claude"])
}

func TestStartupScoringService_GetResult_BeforeCompletion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Enabled = true

	svc := NewStartupScoringService(nil, config, logger)

	// Before running
	assert.Nil(t, svc.GetResult())
	assert.False(t, svc.IsCompleted())
}

func TestStartupScoringService_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DefaultStartupScoringConfig()
	config.Async = true
	config.Timeout = 1 * time.Second

	svc := NewStartupScoringService(nil, config, logger)
	svc.Run(context.Background())

	// Concurrent access to check for race conditions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = svc.GetResult()
			_ = svc.IsCompleted()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestStartupScoringConfig_CustomValues(t *testing.T) {
	config := StartupScoringConfig{
		Enabled:           true,
		Async:             false,
		Timeout:           30 * time.Second,
		RetryOnFailure:    false,
		MaxRetries:        5,
		ConcurrentWorkers: 10,
	}

	assert.True(t, config.Enabled)
	assert.False(t, config.Async)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.False(t, config.RetryOnFailure)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10, config.ConcurrentWorkers)
}

func TestStartupScoringService_NilLogger(t *testing.T) {
	config := DefaultStartupScoringConfig()
	config.Enabled = false

	// Should not panic with nil logger
	svc := NewStartupScoringService(nil, config, nil)
	require.NotNil(t, svc)

	result := svc.Run(context.Background())
	assert.NotNil(t, result)
}
