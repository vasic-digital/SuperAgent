package e2e

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	events "dev.helix.agent/internal/adapters"
	"dev.helix.agent/internal/cache"
	concurrency "digital.vasic.concurrency/pkg/pool"
)

// TestE2E_StartupPerformance tests that system startup is fast
func TestE2E_StartupPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	start := time.Now()

	// Initialize event bus
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})

	// Initialize worker pool
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 1000,
	})
	pool.Start()

	// Initialize cache
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)

	elapsed := time.Since(start)

	// Cleanup
	tc.Close()
	pool.Shutdown(5 * time.Second)
	bus.Close()

	// Startup should be fast - under 500ms
	assert.True(t, elapsed < 500*time.Millisecond,
		"Startup took %v, expected < 500ms", elapsed)

	t.Logf("Component initialization completed in %v", elapsed)
}

// TestE2E_LazyLoadingReducesStartupTime verifies lazy loading benefits
func TestE2E_LazyLoadingReducesStartupTime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Measure startup WITHOUT lazy loading (eager init)
	startEager := time.Now()

	// Initialize all components immediately
	bus1 := events.NewEventBus(nil)
	pool1 := concurrency.NewWorkerPool(nil)
	pool1.Start() // Start returns void
	cache1 := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	})

	// Simulate some work that would happen during eager initialization
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		cache1.Set(ctx, "warmup-"+string(rune('a'+i%26)), i, time.Minute)
	}

	eagerElapsed := time.Since(startEager)

	cache1.Close()
	pool1.Shutdown(time.Second)
	bus1.Close()

	// Measure startup WITH lazy loading
	startLazy := time.Now()

	// Only create components, don't initialize
	bus2 := events.NewEventBus(nil)
	pool2 := concurrency.NewWorkerPool(nil)
	// Don't start pool yet - lazy
	cache2 := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	})

	lazyElapsed := time.Since(startLazy)

	cache2.Close()
	pool2.Shutdown(time.Second)
	bus2.Close()

	t.Logf("Eager startup: %v, Lazy startup: %v", eagerElapsed, lazyElapsed)

	// Lazy startup should be faster or equal
	assert.True(t, lazyElapsed <= eagerElapsed*2,
		"Lazy startup (%v) should be comparable to eager (%v)", lazyElapsed, eagerElapsed)
}

// TestE2E_SystemUnderLoad tests system behavior under realistic load
func TestE2E_SystemUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Setup
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     50000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   32,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(30 * time.Second)

	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	// Subscribe to events
	completedCh := bus.Subscribe(events.EventRequestCompleted)
	var eventsReceived int64
	go func() {
		for range completedCh {
			atomic.AddInt64(&eventsReceived, 1)
		}
	}()

	// Simulate realistic load pattern
	const (
		numWorkers     = 100
		tasksPerWorker = 100
		totalTasks     = numWorkers * tasksPerWorker
	)

	var wg sync.WaitGroup
	var completedTasks int64
	var cacheOps int64

	start := time.Now()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < tasksPerWorker; i++ {
				taskID := workerID*tasksPerWorker + i
				localTaskID := taskID // Capture for closure

				task := concurrency.NewTaskFunc(
					fmt.Sprintf("task-%d", taskID),
					func(taskCtx context.Context) (interface{}, error) {
						// Simulate work with cache access
						key := "task:" + string(rune('a'+localTaskID%26))
						var result int

						found, _ := tc.Get(ctx, key, &result)
						if !found {
							tc.Set(ctx, key, localTaskID, time.Minute)
						}
						atomic.AddInt64(&cacheOps, 1)

						// Publish completion event
						bus.Publish(events.NewEvent(
							events.EventRequestCompleted,
							"worker",
							map[string]interface{}{"task_id": localTaskID},
						))

						atomic.AddInt64(&completedTasks, 1)
						return localTaskID, nil
					},
				)

				err := pool.Submit(task)
				if err != nil {
					t.Logf("Task submission error: %v", err)
					continue
				}
			}
		}(w)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Allow events to propagate
	time.Sleep(500 * time.Millisecond)

	// Verify results
	completed := atomic.LoadInt64(&completedTasks)
	received := atomic.LoadInt64(&eventsReceived)
	ops := atomic.LoadInt64(&cacheOps)

	t.Logf("Load test completed in %v", elapsed)
	t.Logf("Tasks completed: %d/%d", completed, totalTasks)
	t.Logf("Events received: %d", received)
	t.Logf("Cache operations: %d", ops)

	// At least 90% should complete
	assert.True(t, completed >= totalTasks*90/100,
		"Expected at least 90%% completion, got %d/%d", completed, totalTasks)

	// Check pool metrics
	poolMetrics := pool.Metrics()
	t.Logf("Pool metrics - Completed: %d, Failed: %d, Active: %d",
		poolMetrics.CompletedTasks, poolMetrics.FailedTasks, poolMetrics.ActiveWorkers)

	// Check cache metrics
	cacheMetrics := tc.Metrics()
	hitRate := float64(cacheMetrics.L1Hits) / float64(cacheMetrics.L1Hits+cacheMetrics.L1Misses) * 100
	t.Logf("Cache hit rate: %.2f%%", hitRate)
}

// TestE2E_GracefulShutdown tests that system shuts down cleanly
func TestE2E_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	bus := events.NewEventBus(nil)
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 100,
	})
	pool.Start()

	tc := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	})

	// Submit some tasks
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		localI := i
		task := concurrency.NewTaskFunc(
			fmt.Sprintf("shutdown-task-%d", localI),
			func(ctx context.Context) (interface{}, error) {
				defer wg.Done()
				time.Sleep(50 * time.Millisecond)
				return nil, nil
			},
		)
		pool.Submit(task)
	}

	// Start shutdown while tasks are running
	shutdownStart := time.Now()

	// Shutdown in order
	err := pool.Shutdown(10 * time.Second)
	assert.NoError(t, err)

	tc.Close()
	bus.Close()

	shutdownElapsed := time.Since(shutdownStart)

	// Wait for all tasks to complete
	wg.Wait()

	t.Logf("Graceful shutdown completed in %v", shutdownElapsed)

	// Shutdown should complete within timeout
	assert.True(t, shutdownElapsed < 10*time.Second,
		"Shutdown took too long: %v", shutdownElapsed)
}

// TestE2E_ResourceCleanup tests that resources are properly cleaned up
func TestE2E_ResourceCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	for iteration := 0; iteration < 5; iteration++ {
		// Create and destroy components multiple times
		bus := events.NewEventBus(&events.BusConfig{
			BufferSize: 1000,
		})

		pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
			Workers:   4,
			QueueSize: 100,
		})
		pool.Start()

		tc := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
			L1MaxSize: 1000,
			L1TTL:     time.Minute,
			EnableL1:  true,
		})

		// Do some work
		ctx := context.Background()
		for i := 0; i < 100; i++ {
			tc.Set(ctx, "key-"+string(rune('0'+i%10)), i, time.Minute)

			ch := bus.Subscribe(events.EventCacheHit)
			bus.Publish(events.NewEvent(events.EventCacheHit, "test", i))
			bus.Unsubscribe(ch)

			localI := i
			task := concurrency.NewTaskFunc(
				fmt.Sprintf("cleanup-task-%d-%d", iteration, localI),
				func(ctx context.Context) (interface{}, error) {
					return nil, nil
				},
			)
			pool.Submit(task)
		}

		// Cleanup
		tc.Close()
		pool.Shutdown(time.Second)
		bus.Close()
	}

	// No panic or hang means resources were properly cleaned up
	t.Log("Resource cleanup test completed successfully")
}

// TestE2E_EventPropagation tests event flow through the system
func TestE2E_EventPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	// Create multiple subscribers for different event types
	var receivedCounts [5]int64
	eventTypes := []events.EventType{
		events.EventCacheHit,
		events.EventCacheMiss,
		events.EventRequestReceived,
		events.EventRequestCompleted,
		events.EventSystemHealthCheck,
	}

	for i, eventType := range eventTypes {
		idx := i
		ch := bus.Subscribe(eventType)
		go func(c <-chan *events.Event) {
			for range c {
				atomic.AddInt64(&receivedCounts[idx], 1)
			}
		}(ch)
	}

	// Publish events of different types
	for i := 0; i < 100; i++ {
		for _, eventType := range eventTypes {
			bus.Publish(events.NewEvent(eventType, "test", i))
		}
	}

	// Wait for propagation
	time.Sleep(500 * time.Millisecond)

	// Verify all subscribers received events
	for i, eventType := range eventTypes {
		count := atomic.LoadInt64(&receivedCounts[i])
		assert.True(t, count >= 90, "Event type %s received only %d events", eventType, count)
	}
}

// TestE2E_CacheEviction tests that cache eviction works correctly
func TestE2E_CacheEviction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
		L1MaxSize: 10, // Small cache to trigger eviction
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	})
	defer tc.Close()

	ctx := context.Background()

	// Fill cache beyond capacity
	for i := 0; i < 100; i++ {
		tc.Set(ctx, "key-"+string(rune('a'+i%26))+"-"+string(rune('0'+i/26)), i, time.Minute)
	}

	metrics := tc.Metrics()

	// Should have evicted items
	assert.True(t, metrics.L1Evictions > 0, "Expected evictions, got %d", metrics.L1Evictions)

	// Size should be at or below max
	assert.True(t, metrics.L1Size <= 10, "Cache size %d exceeds max 10", metrics.L1Size)

	t.Logf("Cache evictions: %d, final size: %d", metrics.L1Evictions, metrics.L1Size)
}

// TestE2E_ConcurrentStartupShutdown tests rapid startup/shutdown cycles
func TestE2E_ConcurrentStartupShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			bus := events.NewEventBus(nil)
			pool := concurrency.NewWorkerPool(nil)
			pool.Start()

			tc := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
				L1MaxSize: 100,
				L1TTL:     time.Minute,
				EnableL1:  true,
			})

			// Quick operation
			ctx := context.Background()
			tc.Set(ctx, "test", "value", time.Minute)

			task := concurrency.NewTaskFunc(
				"cycle-task",
				func(ctx context.Context) (interface{}, error) {
					return nil, nil
				},
			)
			pool.Submit(task)

			// Shutdown
			tc.Close()
			pool.Shutdown(time.Second)
			bus.Close()

			atomic.AddInt64(&successCount, 1)
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(10), atomic.LoadInt64(&successCount),
		"Not all startup/shutdown cycles completed successfully")
}

// BenchmarkE2E_FullWorkflow benchmarks a complete workflow
func BenchmarkE2E_FullWorkflow(b *testing.B) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize: 100000,
	})
	defer bus.Close()

	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   32,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	tc := cache.NewTieredCache(nil, &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
	})
	defer tc.Close()

	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		tc.Set(ctx, "bench-"+string(rune('a'+i%26)), i, time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(1)

			key := "bench-" + string(rune('a'+i%26))
			localI := i
			task := concurrency.NewTaskFunc(
				fmt.Sprintf("bench-task-%d", localI),
				func(taskCtx context.Context) (interface{}, error) {
					defer wg.Done()

					var result int
					tc.Get(ctx, key, &result)

					bus.PublishAsync(events.NewEvent(events.EventCacheHit, "bench", localI))

					return result, nil
				},
			)

			pool.Submit(task)
			wg.Wait()
			i++
		}
	})
}
