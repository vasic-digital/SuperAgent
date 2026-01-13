package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dev.helix.agent/internal/cache"
	"dev.helix.agent/internal/concurrency"
	"dev.helix.agent/internal/events"
)

// TestWorkerPoolWithEventBus tests worker pool integration with event bus
func TestWorkerPoolWithEventBus(t *testing.T) {
	// Setup event bus
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     1000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	// Setup worker pool
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 100,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	// Subscribe to task completion events
	completionCh := bus.Subscribe(events.EventRequestCompleted)

	var taskCompleted int64
	var eventReceived int64

	// Consume events
	go func() {
		for range completionCh {
			atomic.AddInt64(&eventReceived, 1)
		}
	}()

	// Submit tasks that publish events
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		idx := i
		task := concurrency.NewTaskFunc("task-"+string(rune('0'+idx%10)), func(ctx context.Context) (interface{}, error) {
			defer wg.Done()
			atomic.AddInt64(&taskCompleted, 1)

			// Publish completion event
			bus.Publish(events.NewEvent(
				events.EventRequestCompleted,
				"worker-pool",
				map[string]interface{}{"task_id": idx},
			))

			return nil, nil
		})
		pool.Submit(task)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow events to propagate

	assert.Equal(t, int64(50), atomic.LoadInt64(&taskCompleted))
	assert.True(t, atomic.LoadInt64(&eventReceived) >= 45) // Allow some tolerance
}

// TestCacheWithEventBusInvalidation tests cache invalidation via events
func TestCacheWithEventBusInvalidation(t *testing.T) {
	// Setup event bus
	bus := events.NewEventBus(nil)
	defer bus.Close()

	// Setup cache
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	// Populate cache with provider responses
	for i := 0; i < 10; i++ {
		tc.Set(ctx, "provider:claude:response:"+string(rune('0'+i)), i*100, time.Minute, "provider:claude")
	}

	// Verify cache is populated
	var result int
	found, _ := tc.Get(ctx, "provider:claude:response:5", &result)
	assert.True(t, found)
	assert.Equal(t, 500, result)

	// Subscribe to provider health events
	healthCh := bus.Subscribe(events.EventProviderHealthChanged)

	// Setup invalidation on health change
	go func() {
		for event := range healthCh {
			payload, ok := event.Payload.(map[string]interface{})
			if !ok {
				continue
			}
			if provider, ok := payload["provider"].(string); ok {
				if healthy, ok := payload["healthy"].(bool); ok && !healthy {
					// Invalidate all cache entries for unhealthy provider
					tc.InvalidateByTag(ctx, "provider:"+provider)
				}
			}
		}
	}()

	// Publish provider unhealthy event
	bus.Publish(events.NewEvent(
		events.EventProviderHealthChanged,
		"health-monitor",
		map[string]interface{}{
			"provider": "claude",
			"healthy":  false,
		},
	))

	// Wait for invalidation
	time.Sleep(100 * time.Millisecond)

	// Cache should be invalidated
	found, _ = tc.Get(ctx, "provider:claude:response:5", &result)
	assert.False(t, found, "Cache should be invalidated after provider unhealthy event")
}

// TestWorkerPoolWithCache tests worker pool tasks using cache
func TestWorkerPoolWithCache(t *testing.T) {
	// Setup cache
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 1000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	// Setup worker pool
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 100,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	ctx := context.Background()

	// Submit tasks that read/write to cache
	var wg sync.WaitGroup
	var cacheHits int64
	var cacheMisses int64

	// First wave - cache misses and sets
	for i := 0; i < 26; i++ {
		wg.Add(1)
		key := string(rune('a' + i))
		value := i * 10
		task := concurrency.NewTaskFunc("first-"+key, func(taskCtx context.Context) (interface{}, error) {
			defer wg.Done()

			var result int
			found, _ := tc.Get(ctx, key, &result)
			if found {
				atomic.AddInt64(&cacheHits, 1)
			} else {
				atomic.AddInt64(&cacheMisses, 1)
				tc.Set(ctx, key, value, time.Minute)
			}

			return nil, nil
		})
		pool.Submit(task)
	}
	wg.Wait()

	// Second wave - cache hits
	for i := 0; i < 26; i++ {
		wg.Add(1)
		key := string(rune('a' + i))
		task := concurrency.NewTaskFunc("second-"+key, func(taskCtx context.Context) (interface{}, error) {
			defer wg.Done()

			var result int
			found, _ := tc.Get(ctx, key, &result)
			if found {
				atomic.AddInt64(&cacheHits, 1)
			} else {
				atomic.AddInt64(&cacheMisses, 1)
			}

			return nil, nil
		})
		pool.Submit(task)
	}
	wg.Wait()

	// Should have cache hits from second wave
	assert.True(t, atomic.LoadInt64(&cacheHits) > 0)
	t.Logf("Cache hits: %d, misses: %d", atomic.LoadInt64(&cacheHits), atomic.LoadInt64(&cacheMisses))
}

// TestEventBusMultipleSubscribers tests multiple subscribers receiving events
func TestEventBusMultipleSubscribers(t *testing.T) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	const numSubscribers = 10
	const numEvents = 100

	var received [numSubscribers]int64
	var wg sync.WaitGroup

	// Create subscribers
	channels := make([]<-chan *events.Event, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		channels[i] = bus.Subscribe(events.EventCacheHit)
	}

	// Start receivers
	for i := 0; i < numSubscribers; i++ {
		idx := i
		wg.Add(1)
		go func(ch <-chan *events.Event) {
			defer wg.Done()
			for {
				select {
				case _, ok := <-ch:
					if !ok {
						return
					}
					atomic.AddInt64(&received[idx], 1)
				case <-time.After(2 * time.Second):
					return
				}
			}
		}(channels[i])
	}

	// Publish events
	for i := 0; i < numEvents; i++ {
		bus.Publish(events.NewEvent(events.EventCacheHit, "test", i))
	}

	// Wait for delivery
	time.Sleep(500 * time.Millisecond)
	bus.Close()
	wg.Wait()

	// Each subscriber should receive most events
	for i := 0; i < numSubscribers; i++ {
		count := atomic.LoadInt64(&received[i])
		assert.True(t, count >= numEvents/2, "Subscriber %d received only %d events", i, count)
	}
}

// TestConcurrentCacheAccess tests concurrent cache operations
func TestConcurrentCacheAccess(t *testing.T) {
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	var operations int64

	// Concurrent operations
	for i := 0; i < 100; i++ {
		wg.Add(3)

		// Writer
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := string(rune('a' + (idx+j)%26))
				tc.Set(ctx, key, idx*j, time.Minute)
				atomic.AddInt64(&operations, 1)
			}
		}(i)

		// Reader
		go func(idx int) {
			defer wg.Done()
			var result int
			for j := 0; j < 100; j++ {
				key := string(rune('a' + (idx+j)%26))
				tc.Get(ctx, key, &result)
				atomic.AddInt64(&operations, 1)
			}
		}(i)

		// Deleter
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := string(rune('a' + (idx+j)%26))
				tc.Delete(ctx, key)
				atomic.AddInt64(&operations, 1)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("Completed %d concurrent cache operations", atomic.LoadInt64(&operations))
}

// TestWorkerPoolScaling tests that worker pool scales under load
func TestWorkerPoolScaling(t *testing.T) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   16,
		QueueSize: 1000,
	})
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	var wg sync.WaitGroup
	var completed int64

	// Submit many tasks
	for i := 0; i < 500; i++ {
		wg.Add(1)
		idx := i
		task := concurrency.NewTaskFunc("scale-"+string(rune('0'+idx%10)), func(ctx context.Context) (interface{}, error) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&completed, 1)
			return nil, nil
		})
		pool.Submit(task)
	}

	// Check metrics during execution
	time.Sleep(50 * time.Millisecond)
	metrics := pool.Metrics()
	t.Logf("Workers active during load: %d", metrics.ActiveWorkers)

	wg.Wait()
	assert.Equal(t, int64(500), atomic.LoadInt64(&completed))
}

// TestEventBusWithWorkerPoolTasks tests event-driven task submission
func TestEventBusWithWorkerPoolTasks(t *testing.T) {
	bus := events.NewEventBus(nil)
	defer bus.Close()

	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   8,
		QueueSize: 100,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	// Subscribe to events
	taskCh := bus.Subscribe(events.EventRequestReceived)

	var tasksExecuted int64
	var wg sync.WaitGroup

	// Event consumer that submits tasks to worker pool
	go func() {
		for event := range taskCh {
			wg.Add(1)
			payload := event.Payload
			task := concurrency.NewTaskFunc("event-task", func(ctx context.Context) (interface{}, error) {
				defer wg.Done()
				atomic.AddInt64(&tasksExecuted, 1)
				return payload, nil
			})
			pool.Submit(task)
		}
	}()

	// Publish events
	for i := 0; i < 50; i++ {
		bus.Publish(events.NewEvent(events.EventRequestReceived, "test", i))
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)
	bus.Close()
	wg.Wait()

	assert.True(t, atomic.LoadInt64(&tasksExecuted) >= 40)
}

// TestCacheMetricsCollection tests cache metrics aggregation
func TestCacheMetricsCollection(t *testing.T) {
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 100,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	pc := cache.NewProviderCache(tc, nil)
	mc := cache.NewMCPServerCache(tc, nil)

	ctx := context.Background()

	// Generate cache operations
	for i := 0; i < 10; i++ {
		tc.Set(ctx, "key-"+string(rune('0'+i)), i, time.Minute)
	}

	var result int
	for i := 0; i < 20; i++ {
		tc.Get(ctx, "key-"+string(rune('0'+i%10)), &result) // Some hits
		tc.Get(ctx, "missing-"+string(rune('0'+i)), &result) // Some misses
	}

	// Collect metrics
	collector := cache.NewCacheMetricsCollector(tc, pc, mc, nil)
	metrics := collector.Collect()

	assert.NotNil(t, metrics)
	assert.True(t, metrics.TotalHits+metrics.TotalMisses > 0)

	summary := collector.Summary()
	assert.NotNil(t, summary)
	t.Logf("Cache hit rate: %.2f%%", summary.HitRate)
}

// TestFullIntegration tests all components working together
func TestFullIntegration(t *testing.T) {
	// Setup all components
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   16,
		QueueSize: 1000,
	})
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()
	var totalOperations int64
	var cacheHits int64
	var wg sync.WaitGroup

	// Subscribe to events
	hitCh := bus.Subscribe(events.EventCacheHit)
	missCh := bus.Subscribe(events.EventCacheMiss)

	// Event counters
	go func() {
		for range hitCh {
			atomic.AddInt64(&cacheHits, 1)
		}
	}()
	go func() {
		for range missCh {
		}
	}()

	// Submit tasks that use cache and emit events
	for i := 0; i < 100; i++ {
		wg.Add(1)
		idx := i
		task := concurrency.NewTaskFunc("full-"+string(rune('0'+idx%10)), func(taskCtx context.Context) (interface{}, error) {
			defer wg.Done()

			key := "item:" + string(rune('a'+idx%26))
			var result int

			found, _ := tc.Get(ctx, key, &result)
			if found {
				bus.Publish(events.NewEvent(events.EventCacheHit, "worker", key))
			} else {
				bus.Publish(events.NewEvent(events.EventCacheMiss, "worker", key))
				tc.Set(ctx, key, idx*10, time.Minute)
			}

			atomic.AddInt64(&totalOperations, 1)
			return nil, nil
		})
		pool.Submit(task)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int64(100), atomic.LoadInt64(&totalOperations))
	t.Logf("Total operations: %d, Cache hits: %d",
		atomic.LoadInt64(&totalOperations), atomic.LoadInt64(&cacheHits))

	// Check pool metrics
	poolMetrics := pool.Metrics()
	t.Logf("Pool completed tasks: %d", poolMetrics.CompletedTasks)

	// Check cache metrics
	cacheMetrics := tc.Metrics()
	assert.True(t, cacheMetrics.L1Hits+cacheMetrics.L1Misses >= 100)
}

// BenchmarkIntegration_WorkerPoolWithCache benchmarks integrated worker pool and cache
func BenchmarkIntegration_WorkerPoolWithCache(b *testing.B) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   32,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(10 * time.Second)

	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		tc.Set(ctx, "bench-"+string(rune('a'+i%26)), i, time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(1)

			key := "bench-" + string(rune('a'+i%26))
			task := concurrency.NewTaskFunc("bench-task", func(taskCtx context.Context) (interface{}, error) {
				defer wg.Done()
				var result int
				tc.Get(ctx, key, &result)
				return result, nil
			})

			pool.Submit(task)
			wg.Wait()
			i++
		}
	})
}
