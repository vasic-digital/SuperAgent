package lazy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with factory", func(t *testing.T) {
		factory := func() (string, error) {
			return "test-value", nil
		}
		loader := New(factory, nil)

		require.NotNil(t, loader)
		assert.False(t, loader.IsInitialized())
	})

	t.Run("creates loader with default config", func(t *testing.T) {
		factory := func() (int, error) { return 42, nil }
		loader := New(factory, nil)

		require.NotNil(t, loader)
	})

	t.Run("creates loader with custom config", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		config := &Config{
			TTL:           time.Hour,
			MaxRetries:    3,
			RetryDelay:    time.Second,
			EnableMetrics: true,
		}
		loader := New(factory, config)

		require.NotNil(t, loader)
		assert.NotNil(t, loader.GetMetrics())
	})
}

func TestLoader_Get_Basic(t *testing.T) {
	t.Run("initializes on first get", func(t *testing.T) {
		factory := func() (string, error) {
			return "initialized", nil
		}
		loader := New(factory, nil)

		ctx := context.Background()
		val, err := loader.Get(ctx)

		require.NoError(t, err)
		assert.Equal(t, "initialized", val)
		assert.True(t, loader.IsInitialized())
	})

	t.Run("returns cached value on subsequent gets", func(t *testing.T) {
		callCount := 0
		factory := func() (int, error) {
			callCount++
			return callCount, nil
		}
		loader := New(factory, nil)

		ctx := context.Background()

		// First get - should initialize
		val1, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, val1)
		assert.Equal(t, 1, callCount)

		// Second get - should return cached
		val2, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, val2)
		assert.Equal(t, 1, callCount) // Factory not called again
	})

	t.Run("handles complex types", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		factory := func() (*TestStruct, error) {
			return &TestStruct{Name: "test", Value: 42}, nil
		}
		loader := New(factory, nil)

		ctx := context.Background()
		val, err := loader.Get(ctx)

		require.NoError(t, err)
		assert.Equal(t, "test", val.Name)
		assert.Equal(t, 42, val.Value)
	})
}

func TestLoader_Get_Concurrent(t *testing.T) {
	t.Run("handles concurrent initialization", func(t *testing.T) {
		callCount := 0
		var mu sync.Mutex

		factory := func() (int, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Simulate slow initialization
			return 42, nil
		}
		loader := New(factory, nil)

		ctx := context.Background()
		var wg sync.WaitGroup
		errChan := make(chan error, 100)

		// Launch 100 concurrent Get calls
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, err := loader.Get(ctx)
				if err != nil {
					errChan <- err
					return
				}
				if val != 42 {
					errChan <- errors.New("unexpected value")
				}
			}()
		}

		wg.Wait()
		close(errChan)

		// No errors should occur
		for err := range errChan {
			t.Errorf("Concurrent get error: %v", err)
		}

		// Factory should only be called once
		assert.Equal(t, 1, callCount, "Factory should be called exactly once")
		assert.True(t, loader.IsInitialized())
	})

	t.Run("handles concurrent access after initialization", func(t *testing.T) {
		factory := func() (string, error) { return "cached", nil }
		loader := New(factory, nil)

		// Initialize first
		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.NoError(t, err)

		// Concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, err := loader.Get(ctx)
				assert.NoError(t, err)
				assert.Equal(t, "cached", val)
			}()
		}

		wg.Wait()
	})
}

func TestLoader_IsInitialized(t *testing.T) {
	t.Run("returns false before initialization", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, nil)

		assert.False(t, loader.IsInitialized())
	})

	t.Run("returns true after successful initialization", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, nil)

		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.NoError(t, err)

		assert.True(t, loader.IsInitialized())
	})

	t.Run("returns false after initialization error", func(t *testing.T) {
		factory := func() (string, error) { return "", errors.New("init error") }
		loader := New(factory, nil)

		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.Error(t, err)

		assert.False(t, loader.IsInitialized())
	})
}

func TestLoader_Reset(t *testing.T) {
	t.Run("clears initialized state", func(t *testing.T) {
		factory := func() (string, error) { return "value", nil }
		loader := New(factory, nil)

		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.True(t, loader.IsInitialized())

		loader.Reset()
		assert.False(t, loader.IsInitialized())
	})

	t.Run("allows re-initialization after reset", func(t *testing.T) {
		callCount := 0
		factory := func() (int, error) {
			callCount++
			return callCount, nil
		}
		loader := New(factory, nil)

		ctx := context.Background()

		// First initialization
		val1, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, val1)

		// Reset
		loader.Reset()

		// Re-initialization
		val2, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, val2)
	})

	t.Run("handles reset of zero value type", func(t *testing.T) {
		factory := func() (int, error) { return 42, nil }
		loader := New(factory, nil)

		loader.Reset()
		assert.False(t, loader.IsInitialized())
	})
}

func TestLoader_Close(t *testing.T) {
	t.Run("closes loader without error", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, nil)

		err := loader.Close()
		assert.NoError(t, err)
	})

	t.Run("prevents initialization after close", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, nil)

		err := loader.Close()
		require.NoError(t, err)

		ctx := context.Background()
		_, err = loader.Get(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loader closed")
	})
}

func TestLoader_GetMetrics(t *testing.T) {
	t.Run("returns nil when metrics disabled", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, &Config{EnableMetrics: false})

		metrics := loader.GetMetrics()
		assert.Nil(t, metrics)
	})

	t.Run("returns metrics when enabled", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, &Config{EnableMetrics: true})

		metrics := loader.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("tracks initialization count", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, &Config{EnableMetrics: true})

		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.NoError(t, err)

		metrics := loader.GetMetrics()
		assert.Equal(t, int64(1), metrics.InitCount)
	})

	t.Run("tracks access count", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, &Config{EnableMetrics: true})

		ctx := context.Background()
		// First call initializes (accessCount not incremented on init path)
		_, _ = loader.Get(ctx)
		// Subsequent calls increment accessCount
		_, _ = loader.Get(ctx)
		_, _ = loader.Get(ctx)

		metrics := loader.GetMetrics()
		assert.Equal(t, int64(2), metrics.AccessCount)
	})

	t.Run("tracks initialization errors", func(t *testing.T) {
		factory := func() (string, error) { return "", errors.New("init error") }
		loader := New(factory, &Config{EnableMetrics: true})

		ctx := context.Background()
		_, _ = loader.Get(ctx)

		metrics := loader.GetMetrics()
		assert.Equal(t, int64(1), metrics.InitErrors)
	})

	t.Run("calculates average init time", func(t *testing.T) {
		factory := func() (string, error) {
			time.Sleep(5 * time.Millisecond)
			return "test", nil
		}
		loader := New(factory, &Config{EnableMetrics: true})

		ctx := context.Background()
		_, err := loader.Get(ctx)
		require.NoError(t, err)

		metrics := loader.GetMetrics()
		assert.Greater(t, metrics.AvgInitTime, time.Duration(0))
	})

	t.Run("returns independent copy of metrics", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, &Config{EnableMetrics: true})

		ctx := context.Background()
		_, _ = loader.Get(ctx)

		metrics1 := loader.GetMetrics()
		metrics2 := loader.GetMetrics()

		// Modifying one should not affect the other
		metrics1.InitCount = 999
		assert.Equal(t, int64(1), metrics2.InitCount)
	})
}

func TestLoader_Get_InitializationError(t *testing.T) {
	t.Run("returns error from factory", func(t *testing.T) {
		expectedErr := errors.New("initialization failed")
		factory := func() (string, error) { return "", expectedErr }
		loader := New(factory, nil)

		ctx := context.Background()
		val, err := loader.Get(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initialization failed")
		assert.Equal(t, "", val)
	})

	t.Run("retries on subsequent calls after error", func(t *testing.T) {
		callCount := 0
		factory := func() (string, error) {
			callCount++
			return "", errors.New("error")
		}
		loader := New(factory, nil)

		ctx := context.Background()

		_, err1 := loader.Get(ctx)
		_, err2 := loader.Get(ctx)

		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Equal(t, 2, callCount, "Factory should be called again after error")
	})
}

func TestLoader_Get_ContextCancellation(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		factory := func() (string, error) {
			time.Sleep(5 * time.Second)
			return "test", nil
		}
		loader := New(factory, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		val, err := loader.Get(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
		assert.Equal(t, "", val)
	})

	t.Run("allows retry after context cancellation", func(t *testing.T) {
		callCount := 0
		factory := func() (string, error) {
			callCount++
			if callCount == 1 {
				time.Sleep(5 * time.Second)
			}
			return "success", nil
		}
		loader := New(factory, nil)

		// First attempt with short timeout
		ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
		_, _ = loader.Get(ctx1)
		cancel1()

		// Should be able to retry
		ctx2 := context.Background()
		val, err := loader.Get(ctx2)

		assert.NoError(t, err)
		assert.Equal(t, "success", val)
		assert.Equal(t, 2, callCount)
	})
}

func TestLoader_Warmup(t *testing.T) {
	t.Run("pre-initializes loader", func(t *testing.T) {
		factory := func() (string, error) { return "warmed", nil }
		loader := New(factory, nil)

		ctx := context.Background()
		err := loader.Warmup(ctx)

		require.NoError(t, err)
		assert.True(t, loader.IsInitialized())

		// Get should return immediately
		val, err := loader.Get(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "warmed", val)
	})

	t.Run("returns error if initialization fails", func(t *testing.T) {
		factory := func() (string, error) { return "", errors.New("warmup error") }
		loader := New(factory, nil)

		ctx := context.Background()
		err := loader.Warmup(ctx)

		assert.Error(t, err)
		assert.False(t, loader.IsInitialized())
	})
}

func TestLoader_WaitFor(t *testing.T) {
	t.Run("returns immediately if initialized", func(t *testing.T) {
		factory := func() (string, error) { return "test", nil }
		loader := New(factory, nil)

		ctx := context.Background()
		_, _ = loader.Get(ctx)

		err := loader.WaitFor(ctx)
		assert.NoError(t, err)
	})

	t.Run("waits for initialization", func(t *testing.T) {
		factory := func() (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "test", nil
		}
		loader := New(factory, nil)

		// Start initialization in background
		go func() {
			ctx := context.Background()
			_, _ = loader.Get(ctx)
		}()

		// Wait for it
		ctx := context.Background()
		err := loader.WaitFor(ctx)

		assert.NoError(t, err)
		assert.True(t, loader.IsInitialized())
	})

	t.Run("returns context error if timeout", func(t *testing.T) {
		factory := func() (string, error) {
			time.Sleep(5 * time.Second)
			return "test", nil
		}
		loader := New(factory, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := loader.WaitFor(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestLoader_TTLExpiration(t *testing.T) {
	t.Run("expires after TTL", func(t *testing.T) {
		callCount := 0
		factory := func() (string, error) {
			callCount++
			return "value", nil
		}
		loader := New(factory, &Config{TTL: 100 * time.Millisecond})

		ctx := context.Background()

		// First get
		val1, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "value", val1)
		assert.Equal(t, 1, callCount)

		// Wait for TTL to expire
		time.Sleep(150 * time.Millisecond)

		// Should re-initialize
		val2, err := loader.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "value", val2)
		assert.Equal(t, 2, callCount)
	})

	t.Run("no expiration when TTL is zero", func(t *testing.T) {
		callCount := 0
		factory := func() (string, error) {
			callCount++
			return "value", nil
		}
		loader := New(factory, &Config{TTL: 0})

		ctx := context.Background()

		_, _ = loader.Get(ctx)
		time.Sleep(100 * time.Millisecond)
		_, _ = loader.Get(ctx)

		assert.Equal(t, 1, callCount)
	})
}

func TestRegistry(t *testing.T) {
	t.Run("NewRegistry creates empty registry", func(t *testing.T) {
		registry := NewRegistry()
		assert.NotNil(t, registry)

		loader, ok := registry.Get("test")
		assert.False(t, ok)
		assert.Nil(t, loader)
	})

	t.Run("Register adds loader", func(t *testing.T) {
		registry := NewRegistry()
		loader := New(func() (string, error) { return "test", nil }, nil)

		registry.Register("my-loader", loader)

		retrieved, ok := registry.Get("my-loader")
		assert.True(t, ok)
		assert.Equal(t, loader, retrieved)
	})

	t.Run("Register overwrites existing", func(t *testing.T) {
		registry := NewRegistry()
		loader1 := New(func() (string, error) { return "first", nil }, nil)
		loader2 := New(func() (string, error) { return "second", nil }, nil)

		registry.Register("test", loader1)
		registry.Register("test", loader2)

		retrieved, _ := registry.Get("test")
		assert.Equal(t, loader2, retrieved)
	})

	t.Run("handles multiple loaders", func(t *testing.T) {
		registry := NewRegistry()

		loader1 := New(func() (int, error) { return 1, nil }, nil)
		loader2 := New(func() (int, error) { return 2, nil }, nil)
		loader3 := New(func() (int, error) { return 3, nil }, nil)

		registry.Register("loader1", loader1)
		registry.Register("loader2", loader2)
		registry.Register("loader3", loader3)

		l1, ok1 := registry.Get("loader1")
		l2, ok2 := registry.Get("loader2")
		l3, ok3 := registry.Get("loader3")

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.True(t, ok3)
		assert.Equal(t, loader1, l1)
		assert.Equal(t, loader2, l2)
		assert.Equal(t, loader3, l3)
	})

	t.Run("is safe for concurrent access", func(t *testing.T) {
		registry := NewRegistry()

		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				loader := New(func() (int, error) { return n, nil }, nil)
				registry.Register(fmt.Sprintf("loader-%d", n), loader)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				registry.Get(fmt.Sprintf("loader-%d", n))
			}(i)
		}

		wg.Wait()

		// All loaders should be registered
		for i := 0; i < 100; i++ {
			_, ok := registry.Get(fmt.Sprintf("loader-%d", i))
			assert.True(t, ok, "Loader %d should be registered", i)
		}
	})
}

func TestRegistry_InitializeAll(t *testing.T) {
	t.Run("initializes all loaders", func(t *testing.T) {
		registry := NewRegistry()

		// Note: InitializeAll only works with *Loader[interface{}] due to type assertion
		loader1 := New(func() (interface{}, error) { return "v1", nil }, nil)
		loader2 := New(func() (interface{}, error) { return "v2", nil }, nil)

		registry.Register("loader1", loader1)
		registry.Register("loader2", loader2)

		ctx := context.Background()
		err := registry.InitializeAll(ctx)

		assert.NoError(t, err)
		assert.True(t, loader1.IsInitialized())
		assert.True(t, loader2.IsInitialized())
	})

	t.Run("returns error if any loader fails", func(t *testing.T) {
		registry := NewRegistry()

		// Note: InitializeAll only works with *Loader[interface{}] due to type assertion
		loader1 := New(func() (interface{}, error) { return "v1", nil }, nil)
		loader2 := New(func() (interface{}, error) { return nil, errors.New("fail") }, nil)

		registry.Register("loader1", loader1)
		registry.Register("loader2", loader2)

		ctx := context.Background()
		err := registry.InitializeAll(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize")
	})

	t.Run("initializes concurrently", func(t *testing.T) {
		registry := NewRegistry()

		start := time.Now()

		// Create loaders with delays
		for i := 0; i < 5; i++ {
			loader := New(func() (interface{}, error) {
				time.Sleep(50 * time.Millisecond)
				return "value", nil
			}, nil)
			registry.Register(fmt.Sprintf("loader-%d", i), loader)
		}

		ctx := context.Background()
		err := registry.InitializeAll(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		// Should be much faster than sequential (250ms)
		assert.Less(t, elapsed, 200*time.Millisecond)
	})
}

func TestRegistry_CloseAll(t *testing.T) {
	t.Run("closes all loaders", func(t *testing.T) {
		registry := NewRegistry()

		// Note: CloseAll only works with *Loader[interface{}] due to type assertion
		loader1 := New(func() (interface{}, error) { return "v1", nil }, nil)
		loader2 := New(func() (interface{}, error) { return "v2", nil }, nil)

		registry.Register("loader1", loader1)
		registry.Register("loader2", loader2)

		err := registry.CloseAll()

		assert.NoError(t, err)

		// Verify closed by trying to get
		ctx := context.Background()
		_, err = loader1.Get(ctx)
		assert.Error(t, err)
	})

	t.Run("handles empty registry", func(t *testing.T) {
		registry := NewRegistry()

		err := registry.CloseAll()

		assert.NoError(t, err)
	})
}

// Benchmarks
func BenchmarkLoader_Get(b *testing.B) {
	factory := func() (string, error) { return "test", nil }
	loader := New(factory, nil)
	ctx := context.Background()

	// Initialize first
	_, _ = loader.Get(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.Get(ctx)
	}
}

func BenchmarkLoader_Get_Concurrent(b *testing.B) {
	factory := func() (string, error) { return "test", nil }
	loader := New(factory, nil)
	ctx := context.Background()

	// Initialize first
	_, _ = loader.Get(ctx)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = loader.Get(ctx)
		}
	})
}

func BenchmarkRegistry_Register(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := New(func() (int, error) { return i, nil }, nil)
		registry.Register(fmt.Sprintf("loader-%d", i), loader)
	}
}
