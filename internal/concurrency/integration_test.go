package concurrency

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for concurrency utilities
// These tests verify real-world usage scenarios

func TestSemaphore_Integration(t *testing.T) {
	t.Run("bounded concurrent access to resource", func(t *testing.T) {
		// Simulate a database connection pool with max 3 connections
		sem := NewSemaphore(3)
		defer sem.Close()

		var activeConnections int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Try to use 10 connections simultaneously
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Acquire connection
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				err := sem.Acquire(ctx)
				require.NoError(t, err)

				// Use connection
				mu.Lock()
				activeConnections++
				current := activeConnections
				mu.Unlock()

				// Verify we never exceed limit
				assert.LessOrEqual(t, current, 3, "Active connections exceeded limit")

				// Simulate work
				time.Sleep(10 * time.Millisecond)

				mu.Lock()
				activeConnections--
				mu.Unlock()

				// Release connection
				sem.Release()
			}(i)
		}

		wg.Wait()
	})

	t.Run("rate limiting API calls", func(t *testing.T) {
		// Allow max 5 concurrent API calls
		sem := NewSemaphore(5)
		defer sem.Close()

		var completed int
		var mu sync.Mutex
		var wg sync.WaitGroup

		start := time.Now()

		// Launch 20 API calls
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				err := sem.Acquire(context.Background())
				require.NoError(t, err)

				// Simulate API call
				time.Sleep(20 * time.Millisecond)

				mu.Lock()
				completed++
				mu.Unlock()

				sem.Release()
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// With 5 concurrent slots and 20 calls at 20ms each,
		// should take at least 80ms (4 batches)
		assert.GreaterOrEqual(t, duration, 80*time.Millisecond)
		assert.Equal(t, 20, completed)
	})
}

func TestRateLimiter_Integration(t *testing.T) {
	t.Run("API rate limiting", func(t *testing.T) {
		// Limit to 10 requests per second
		rl := NewRateLimiter(10)
		defer rl.Stop()

		var successCount int
		var mu sync.Mutex
		var wg sync.WaitGroup

		start := time.Now()

		// Try to make 50 requests as fast as possible
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := rl.Acquire(context.Background())
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		// Should take at least 4 seconds for 50 requests at 10/sec
		assert.GreaterOrEqual(t, duration, 4*time.Second)
		assert.Equal(t, 50, successCount)
	})

	t.Run("rate limit with timeout", func(t *testing.T) {
		// Very restrictive: 1 request per second
		rl := NewRateLimiter(1)
		defer rl.Stop()

		// First request should succeed immediately
		err := rl.Acquire(context.Background())
		require.NoError(t, err)

		// Second request with short timeout should fail
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = rl.Acquire(ctx)
		assert.Error(t, err)
	})
}

func TestResourcePool_Integration(t *testing.T) {
	t.Run("database connection pool simulation", func(t *testing.T) {
		connectionID := 0
		var mu sync.Mutex

		factory := func() (interface{}, error) {
			mu.Lock()
			defer mu.Unlock()
			connectionID++
			return map[string]interface{}{
				"id":     connectionID,
				"status": "ready",
			}, nil
		}

		// Pool of 3 connections
		pool, err := NewResourcePool(3, factory)
		require.NoError(t, err)
		defer pool.Close()

		// Acquire all connections
		var connections []interface{}
		for i := 0; i < 3; i++ {
			conn, err := pool.Acquire(context.Background())
			require.NoError(t, err)
			connections = append(connections, conn)
		}

		// Verify each connection is unique
		ids := make(map[int]bool)
		for _, conn := range connections {
			c := conn.(map[string]interface{})
			id := c["id"].(int)
			assert.False(t, ids[id], "Duplicate connection ID")
			ids[id] = true
		}

		// Release all connections
		for _, conn := range connections {
			err := pool.Release(conn)
			require.NoError(t, err)
		}

		// Should be able to acquire again
		conn, err := pool.Acquire(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, conn)
	})
}

func TestAsyncProcessor_Integration(t *testing.T) {
	t.Run("background job processing", func(t *testing.T) {
		// 2 workers, queue size 10
		processor := NewAsyncProcessor(2, 10)
		defer processor.Stop()

		var results []int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Submit 10 jobs
		for i := 0; i < 10; i++ {
			wg.Add(1)
			jobID := i
			submitted := processor.Submit(func() {
				defer wg.Done()
				// Simulate work
				time.Sleep(5 * time.Millisecond)
				mu.Lock()
				results = append(results, jobID)
				mu.Unlock()
			})
			assert.True(t, submitted)
		}

		wg.Wait()

		mu.Lock()
		assert.Len(t, results, 10)
		mu.Unlock()
	})

	t.Run("graceful shutdown", func(t *testing.T) {
		processor := NewAsyncProcessor(1, 5)

		var completed int
		var mu sync.Mutex

		// Submit slow jobs
		for i := 0; i < 3; i++ {
			processor.Submit(func() {
				time.Sleep(50 * time.Millisecond)
				mu.Lock()
				completed++
				mu.Unlock()
			})
		}

		// Give jobs time to start
		time.Sleep(10 * time.Millisecond)

		// Stop processor - should wait for running jobs
		processor.Stop()

		mu.Lock()
		// Should have completed running jobs
		assert.GreaterOrEqual(t, completed, 1)
		mu.Unlock()
	})
}

func TestLazyLoader_Integration(t *testing.T) {
	t.Run("expensive resource initialization", func(t *testing.T) {
		loadCount := 0
		var mu sync.Mutex

		loader := func() (interface{}, error) {
			mu.Lock()
			defer mu.Unlock()
			loadCount++
			// Simulate expensive initialization
			time.Sleep(100 * time.Millisecond)
			return "expensive-resource", nil
		}

		lazyLoader := NewLazyLoader(loader)

		// Multiple goroutines try to get resource simultaneously
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, err := lazyLoader.Get()
				require.NoError(t, err)
				assert.Equal(t, "expensive-resource", val)
			}()
		}

		wg.Wait()

		// Should only load once despite 10 concurrent requests
		mu.Lock()
		assert.Equal(t, 1, loadCount)
		mu.Unlock()
	})

	t.Run("cache after load", func(t *testing.T) {
		loadCount := 0
		loader := func() (interface{}, error) {
			loadCount++
			return "cached-value", nil
		}

		lazyLoader := NewLazyLoader(loader)

		// First get
		val, err := lazyLoader.Get()
		require.NoError(t, err)
		assert.Equal(t, "cached-value", val)

		// Subsequent gets should use cache
		for i := 0; i < 5; i++ {
			val, err := lazyLoader.Get()
			require.NoError(t, err)
			assert.Equal(t, "cached-value", val)
		}

		assert.Equal(t, 1, loadCount)
	})
}

func TestNonBlockingCache_Integration(t *testing.T) {
	t.Run("high concurrency cache access", func(t *testing.T) {
		cache := NewNonBlockingCache(time.Minute)

		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Set(string(rune('a'+id%26)), id)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Get(string(rune('a' + id%26)))
			}(i)
		}

		wg.Wait()

		// Should have 26 unique keys
		assert.LessOrEqual(t, cache.Len(), 26)
	})

	t.Run("cache consistency under load", func(t *testing.T) {
		cache := NewNonBlockingCache(time.Minute)

		// Write value
		cache.Set("counter", 0)

		var wg sync.WaitGroup
		var mu sync.Mutex

		// Increment counter 1000 times
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mu.Lock()
				val, _ := cache.Get("counter")
				count := val.(int)
				cache.Set("counter", count+1)
				mu.Unlock()
			}()
		}

		wg.Wait()

		val, _ := cache.Get("counter")
		assert.Equal(t, 1000, val.(int))
	})
}

func TestBackgroundTask_Integration(t *testing.T) {
	t.Run("periodic health check simulation", func(t *testing.T) {
		checkCount := 0
		var mu sync.Mutex

		task := NewBackgroundTask(func(ctx context.Context) {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					mu.Lock()
					checkCount++
					mu.Unlock()
				case <-ctx.Done():
					return
				}
			}
		})

		task.Start()

		// Let it run for 50ms
		time.Sleep(50 * time.Millisecond)

		task.Stop()

		mu.Lock()
		// Should have checked approximately 5 times
		assert.GreaterOrEqual(t, checkCount, 3)
		mu.Unlock()
	})

	t.Run("cleanup on stop", func(t *testing.T) {
		cleanedUp := false

		task := NewBackgroundTask(func(ctx context.Context) {
			<-ctx.Done()
			// Cleanup
			cleanedUp = true
		})

		task.Start()
		time.Sleep(10 * time.Millisecond)
		task.Stop()

		assert.True(t, cleanedUp)
	})
}

func TestConcurrencyUtilities_Together(t *testing.T) {
	t.Run("complete request processing pipeline", func(t *testing.T) {
		// Rate limiter: 100 req/sec
		rl := NewRateLimiter(100)
		defer rl.Stop()

		// Semaphore: max 10 concurrent DB operations
		sem := NewSemaphore(10)
		defer sem.Close()

		// Resource pool: 5 connections
		pool, err := NewResourcePool(5, func() (interface{}, error) {
			return "db-connection", nil
		})
		require.NoError(t, err)
		defer pool.Close()

		// Async processor for background tasks
		processor := NewAsyncProcessor(3, 20)
		defer processor.Stop()

		var successCount int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Simulate 50 requests
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 1. Rate limiting
				if err := rl.Acquire(context.Background()); err != nil {
					return
				}

				// 2. Semaphore for DB access
				if err := sem.Acquire(context.Background()); err != nil {
					return
				}

				// 3. Get DB connection from pool
				conn, err := pool.Acquire(context.Background())
				if err != nil {
					sem.Release()
					return
				}

				// Simulate DB operation
				time.Sleep(5 * time.Millisecond)
				_ = conn

				// 4. Release resources
				pool.Release(conn)
				sem.Release()

				// 5. Background task
				processor.Submit(func() {
					// Log analytics
					time.Sleep(1 * time.Millisecond)
				})

				mu.Lock()
				successCount++
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		time.Sleep(50 * time.Millisecond) // Wait for background tasks

		mu.Lock()
		assert.GreaterOrEqual(t, successCount, 40) // At least 80% success
		mu.Unlock()
	})
}
