//go:build performance
// +build performance

// Package performance contains benchmark and load tests for critical components.
package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"dev.helix.agent/internal/cache"
	"dev.helix.agent/internal/concurrency"
	"dev.helix.agent/internal/events"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =============================================================================
// CACHE BENCHMARKS
// =============================================================================

// BenchmarkCache_Get benchmarks cache read operations
func BenchmarkCache_Get(b *testing.B) {
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		tc.Set(ctx, "key:"+string(rune('0'+i%10)), i, time.Minute, "benchmark")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var result int
		i := 0
		for pb.Next() {
			tc.Get(ctx, "key:"+string(rune('0'+i%10)), &result)
			i++
		}
	})
}

// BenchmarkCache_Set benchmarks cache write operations
func BenchmarkCache_Set(b *testing.B) {
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 100000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tc.Set(ctx, "key:"+string(rune('0'+i%10)), i, time.Minute, "benchmark")
			i++
		}
	})
}

// BenchmarkCache_GetSet_Mixed benchmarks mixed read/write operations
func BenchmarkCache_GetSet_Mixed(b *testing.B) {
	cacheConfig := &cache.TieredCacheConfig{
		L1MaxSize: 10000,
		L1TTL:     time.Minute,
		EnableL1:  true,
		EnableL2:  false,
	}
	tc := cache.NewTieredCache(nil, cacheConfig)
	defer tc.Close()

	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		tc.Set(ctx, "key:"+string(rune('0'+i%10)), i, time.Minute, "benchmark")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var result int
		i := 0
		for pb.Next() {
			if i%5 == 0 { // 20% writes
				tc.Set(ctx, "key:"+string(rune('0'+i%10)), i, time.Minute, "benchmark")
			} else { // 80% reads
				tc.Get(ctx, "key:"+string(rune('0'+i%10)), &result)
			}
			i++
		}
	})
}

// =============================================================================
// EVENT BUS BENCHMARKS
// =============================================================================

// BenchmarkEventBus_Publish benchmarks event publishing
func BenchmarkEventBus_Publish(b *testing.B) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	// Subscribe to consume events
	ch := bus.Subscribe(events.EventRequestCompleted)
	go func() {
		for range ch {
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(events.NewEvent(
				events.EventRequestCompleted,
				"benchmark",
				map[string]interface{}{"id": 1},
			))
		}
	})
}

// BenchmarkEventBus_PubSub benchmarks full publish-subscribe cycle
func BenchmarkEventBus_PubSub(b *testing.B) {
	bus := events.NewEventBus(&events.BusConfig{
		BufferSize:     10000,
		PublishTimeout: 100 * time.Millisecond,
	})
	defer bus.Close()

	var received int64
	ch := bus.Subscribe(events.EventRequestCompleted)
	go func() {
		for range ch {
			atomic.AddInt64(&received, 1)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(events.NewEvent(
			events.EventRequestCompleted,
			"benchmark",
			map[string]interface{}{"id": i},
		))
	}
}

// =============================================================================
// WORKER POOL BENCHMARKS
// =============================================================================

// BenchmarkWorkerPool_Submit benchmarks task submission
func BenchmarkWorkerPool_Submit(b *testing.B) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   16,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := concurrency.NewTaskFunc("task", func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
		pool.Submit(task)
	}
}

// BenchmarkWorkerPool_SubmitAndWait benchmarks task execution
func BenchmarkWorkerPool_SubmitAndWait(b *testing.B) {
	pool := concurrency.NewWorkerPool(&concurrency.PoolConfig{
		Workers:   16,
		QueueSize: 10000,
	})
	pool.Start()
	defer pool.Shutdown(5 * time.Second)

	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		task := concurrency.NewTaskFunc("task", func(ctx context.Context) (interface{}, error) {
			defer wg.Done()
			return nil, nil
		})
		pool.Submit(task)
	}
	wg.Wait()
}

// =============================================================================
// HTTP HANDLER BENCHMARKS
// =============================================================================

// setupBenchmarkRouter creates a router for benchmarking
func setupBenchmarkRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.POST("/v1/chat/completions", func(c *gin.Context) {
		var req map[string]interface{}
		c.BindJSON(&req)
		c.JSON(http.StatusOK, gin.H{
			"id":      "chatcmpl-bench",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Benchmark response",
					},
					"finish_reason": "stop",
				},
			},
		})
	})

	r.GET("/v1/models", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"object": "list",
			"data": []map[string]interface{}{
				{"id": "model-1", "object": "model"},
				{"id": "model-2", "object": "model"},
			},
		})
	})

	return r
}

// BenchmarkHTTP_HealthCheck benchmarks health endpoint
func BenchmarkHTTP_HealthCheck(b *testing.B) {
	router := setupBenchmarkRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/health")
			if err == nil {
				resp.Body.Close()
			}
		}
	})
}

// BenchmarkHTTP_ChatCompletion benchmarks chat completion endpoint
func BenchmarkHTTP_ChatCompletion(b *testing.B) {
	router := setupBenchmarkRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}
	reqBody := map[string]interface{}{
		"model": "test-model",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Post(
				server.URL+"/v1/chat/completions",
				"application/json",
				bytes.NewReader(jsonBody),
			)
			if err == nil {
				resp.Body.Close()
			}
		}
	})
}

// =============================================================================
// LOAD TESTS
// =============================================================================

// TestLoadTest_ConcurrentRequests tests system under concurrent load
func TestLoadTest_ConcurrentRequests(t *testing.T) {
	router := setupBenchmarkRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	const (
		numClients      = 50
		requestsPerSec  = 100
		testDurationSec = 5
	)

	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(testDurationSec)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Calculate interval: we want requestsPerSec total, so each client does requestsPerSec/numClients
			// Interval = numClients * Second / requestsPerSec
			ticker := time.NewTicker(time.Duration(numClients) * time.Second / time.Duration(requestsPerSec))
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					atomic.AddInt64(&totalRequests, 1)
					resp, err := client.Get(server.URL + "/health")
					if err != nil {
						atomic.AddInt64(&failedRequests, 1)
						continue
					}
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&successRequests, 1)
					} else {
						atomic.AddInt64(&failedRequests, 1)
					}
				}
			}
		}()
	}

	wg.Wait()

	total := atomic.LoadInt64(&totalRequests)
	success := atomic.LoadInt64(&successRequests)
	failed := atomic.LoadInt64(&failedRequests)

	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", total)
	t.Logf("  Successful: %d (%.2f%%)", success, float64(success)/float64(total)*100)
	t.Logf("  Failed: %d (%.2f%%)", failed, float64(failed)/float64(total)*100)
	t.Logf("  Requests/sec: %.2f", float64(total)/float64(testDurationSec))

	// Success rate should be > 99%
	successRate := float64(success) / float64(total)
	assert.Greater(t, successRate, 0.99, "Success rate should be > 99%%")
}

// TestLoadTest_BurstTraffic tests system under burst traffic
func TestLoadTest_BurstTraffic(t *testing.T) {
	router := setupBenchmarkRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	const burstSize = 1000

	var (
		successful int64
		failed     int64
	)

	var wg sync.WaitGroup
	start := time.Now()

	// Send burst of requests
	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Get(server.URL + "/health")
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&successful, 1)
			} else {
				atomic.AddInt64(&failed, 1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Burst Test Results:")
	t.Logf("  Burst Size: %d", burstSize)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Successful: %d", successful)
	t.Logf("  Failed: %d", failed)
	t.Logf("  Requests/sec: %.2f", float64(burstSize)/duration.Seconds())

	// Most requests should succeed
	assert.Greater(t, atomic.LoadInt64(&successful), int64(burstSize*95/100), "At least 95%% should succeed")
}

// =============================================================================
// MEMORY ALLOCATION TESTS
// =============================================================================

// BenchmarkAllocation_JSONMarshal benchmarks JSON marshaling allocations
func BenchmarkAllocation_JSONMarshal(b *testing.B) {
	data := map[string]interface{}{
		"id":      "test-id",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "This is a test response with some content.",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

// BenchmarkAllocation_EventCreation benchmarks event creation allocations
func BenchmarkAllocation_EventCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = events.NewEvent(
			events.EventRequestCompleted,
			"benchmark",
			map[string]interface{}{"id": i, "status": "completed"},
		)
	}
}
