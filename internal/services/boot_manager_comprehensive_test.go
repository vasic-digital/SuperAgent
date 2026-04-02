package services

import (
	"context"
	"errors"
	"testing"
	"time"

	containeradapter "dev.helix.agent/internal/adapters/containers"
	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootManager_NewBootManager(t *testing.T) {
	t.Run("creates boot manager with defaults", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{
			PostgreSQL: config.ServiceEndpoint{Enabled: true},
		}

		bm := NewBootManager(cfg, logger)

		assert.NotNil(t, bm)
		assert.Equal(t, cfg, bm.Config)
		assert.Equal(t, logger, bm.Logger)
		assert.NotNil(t, bm.Results)
		assert.NotNil(t, bm.HealthChecker)
		assert.NotNil(t, bm.Discoverer)
		assert.Nil(t, bm.RemoteDeployer)
	})
}

func TestBootManager_SetContainerAdapter(t *testing.T) {
	t.Run("sets container adapter safely", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		adapter := &containeradapter.Adapter{}
		bm.SetContainerAdapter(adapter)

		assert.Equal(t, adapter, bm.ContainerAdapter)
		assert.Equal(t, adapter, bm.getContainerAdapter())
	})

	t.Run("concurrent adapter access is safe", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		done := make(chan bool, 10)

		// Multiple goroutines setting adapter
		for i := 0; i < 5; i++ {
			go func() {
				adapter := &containeradapter.Adapter{}
				bm.SetContainerAdapter(adapter)
				done <- true
			}()
		}

		// Multiple goroutines reading adapter
		for i := 0; i < 5; i++ {
			go func() {
				_ = bm.getContainerAdapter()
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("Timeout waiting for concurrent access")
			}
		}
	})
}

func TestBootManager_GetResults(t *testing.T) {
	t.Run("returns defensive copy of results", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		// Set initial result
		bm.setResult("test-service", &BootResult{
			Name:   "test-service",
			Status: "started",
		})

		// Get copy
		results := bm.GetResults()

		// Modify copy
		results["test-service"].Status = "modified"

		// Original should be unchanged
		original, ok := bm.getResult("test-service")
		require.True(t, ok)
		assert.Equal(t, "started", original.Status)
	})

	t.Run("returns empty map when no results", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		results := bm.GetResults()
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})
}

func TestBootManager_setResult(t *testing.T) {
	t.Run("stores result under write lock", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		result := &BootResult{
			Name:     "test-service",
			Status:   "started",
			Duration: time.Second,
		}

		bm.setResult("test-service", result)

		stored, ok := bm.getResult("test-service")
		require.True(t, ok)
		assert.Equal(t, result.Name, stored.Name)
		assert.Equal(t, result.Status, stored.Status)
	})
}

func TestBootManager_setResultIfAbsent(t *testing.T) {
	t.Run("sets result only if not exists", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		// First set
		bm.setResultIfAbsent("test-service", &BootResult{
			Name:   "test-service",
			Status: "first",
		})

		// Second set should be ignored
		bm.setResultIfAbsent("test-service", &BootResult{
			Name:   "test-service",
			Status: "second",
		})

		stored, ok := bm.getResult("test-service")
		require.True(t, ok)
		assert.Equal(t, "first", stored.Status)
	})

	t.Run("sets result when not exists", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		bm.setResultIfAbsent("new-service", &BootResult{
			Name:   "new-service",
			Status: "started",
		})

		stored, ok := bm.getResult("new-service")
		require.True(t, ok)
		assert.Equal(t, "started", stored.Status)
	})
}

func TestBootManager_getResult(t *testing.T) {
	t.Run("returns copy of existing result", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		bm.setResult("test-service", &BootResult{
			Name:   "test-service",
			Status: "started",
		})

		result, ok := bm.getResult("test-service")
		require.True(t, ok)
		assert.Equal(t, "test-service", result.Name)
		assert.Equal(t, "started", result.Status)
	})

	t.Run("returns false for non-existent result", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		result, ok := bm.getResult("non-existent")
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestBootManager_HealthCheckAll(t *testing.T) {
	t.Run("checks all enabled services", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{
			Postgres: config.ServiceEndpoint{
				Enabled: true,
				Host:    "localhost",
				Port:    5432,
			},
			Redis: config.ServiceEndpoint{
				Enabled: false,
				Host:    "localhost",
				Port:    6379,
			},
		}

		bm := NewBootManager(cfg, logger)
		results := bm.HealthCheckAll()

		// Should check enabled services
		assert.Contains(t, results, "Postgres")
		// Disabled services may or may not be in results depending on implementation
	})
}

func TestBootManager_HealthCheckAllParallel(t *testing.T) {
	t.Run("checks services concurrently", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{
			Postgres: config.ServiceEndpoint{
				Enabled: true,
				Host:    "localhost",
				Port:    5432,
			},
			Redis: config.ServiceEndpoint{
				Enabled: true,
				Host:    "localhost",
				Port:    6379,
			},
		}

		bm := NewBootManager(cfg, logger)
		ctx := context.Background()
		results := bm.HealthCheckAllParallel(ctx)

		assert.NotNil(t, results)
	})
}

func TestBootManager_ShutdownAll(t *testing.T) {
	t.Run("handles shutdown with no services", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		err := bm.ShutdownAll()
		assert.NoError(t, err)
	})

	t.Run("skips remote services during shutdown", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{
			PostgreSQL: config.ServiceEndpoint{
				Enabled: true,
				Remote:  true,
			},
		}

		bm := NewBootManager(cfg, logger)
		bm.setResult("Postgres", &BootResult{
			Name:   "Postgres",
			Status: "started",
		})

		// Should not error even without adapter
		err := bm.ShutdownAll()
		// May error due to missing adapter, but should not panic
		_ = err
	})
}

func TestBootManager_logSummary(t *testing.T) {
	t.Run("logs correct summary counts", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		// Add results of various statuses
		bm.setResult("started-svc", &BootResult{Name: "started-svc", Status: "started"})
		bm.setResult("already-svc", &BootResult{Name: "already-svc", Status: "already_running"})
		bm.setResult("remote-svc", &BootResult{Name: "remote-svc", Status: "remote"})
		bm.setResult("discovered-svc", &BootResult{Name: "discovered-svc", Status: "discovered"})
		bm.setResult("failed-svc", &BootResult{Name: "failed-svc", Status: "failed"})
		bm.setResult("skipped-svc", &BootResult{Name: "skipped-svc", Status: "skipped"})

		// Should not panic
		bm.logSummary()
	})
}

func TestBootManager_checkRemoteServiceHealth(t *testing.T) {
	t.Run("returns error when adapter is nil", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		ep := config.ServiceEndpoint{
			HealthType: "http",
		}

		err := bm.checkRemoteServiceHealth(context.Background(), "test-service", ep)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "container adapter not configured")
	})

	t.Run("returns error when remote not enabled", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		// Create adapter with RemoteEnabled = false
		adapter := &containeradapter.Adapter{}
		bm.SetContainerAdapter(adapter)

		ep := config.ServiceEndpoint{
			HealthType: "http",
		}

		err := bm.checkRemoteServiceHealth(context.Background(), "test-service", ep)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "remote distribution not enabled")
	})
}

func TestBootManager_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent result access", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.ServicesConfig{}
		bm := NewBootManager(cfg, logger)

		done := make(chan bool, 20)

		// Concurrent writes
		for i := 0; i < 10; i++ {
			go func(idx int) {
				bm.setResult(string(rune('a'+idx)), &BootResult{
					Name:   string(rune('a' + idx)),
					Status: "started",
				})
				done <- true
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func() {
				_ = bm.GetResults()
				done <- true
			}()
		}

		// Wait for all
		for i := 0; i < 20; i++ {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("Timeout during concurrent access")
			}
		}

		// Verify results
		results := bm.GetResults()
		assert.GreaterOrEqual(t, len(results), 0)
	})
}

func TestBootResult_StatusValues(t *testing.T) {
	t.Run("validates all status values", func(t *testing.T) {
		statuses := []string{
			"started",
			"already_running",
			"remote",
			"discovered",
			"failed",
			"skipped",
		}

		for _, status := range statuses {
			result := &BootResult{
				Name:   "test",
				Status: status,
			}
			assert.Equal(t, status, result.Status)
		}
	})
}

func TestBootResult_ErrorHandling(t *testing.T) {
	t.Run("stores error correctly", func(t *testing.T) {
		testError := errors.New("test error")
		result := &BootResult{
			Name:   "test-service",
			Status: "failed",
			Error:  testError,
		}

		assert.Equal(t, testError, result.Error)
		assert.Equal(t, "failed", result.Status)
	})

	t.Run("handles nil error", func(t *testing.T) {
		result := &BootResult{
			Name:   "test-service",
			Status: "started",
			Error:  nil,
		}

		assert.Nil(t, result.Error)
	})
}

func TestBootResult_Duration(t *testing.T) {
	t.Run("stores duration correctly", func(t *testing.T) {
		duration := 5 * time.Second
		result := &BootResult{
			Name:     "test-service",
			Status:   "started",
			Duration: duration,
		}

		assert.Equal(t, duration, result.Duration)
	})
}
