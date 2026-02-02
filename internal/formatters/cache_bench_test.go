package formatters

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func newBenchCache(maxSize int) *FormatterCache {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // suppress logs during benchmarks

	cfg := &CacheConfig{
		TTL:         10 * time.Minute,
		MaxSize:     maxSize,
		CleanupFreq: 10 * time.Minute, // long interval to avoid interference
	}
	return NewFormatterCache(cfg, logger)
}

func benchFormatRequest(i int) *FormatRequest {
	return &FormatRequest{
		Content:  fmt.Sprintf("func main() { fmt.Println(%d) }", i),
		Language: "go",
		FilePath: fmt.Sprintf("file_%d.go", i),
	}
}

func benchFormatResult() *FormatResult {
	return &FormatResult{
		Content:       "func main() { fmt.Println(0) }",
		Changed:       true,
		FormatterName: "gofmt",
		Success:       true,
	}
}

// BenchmarkFormatCacheGet measures the throughput of cache lookups (hits).
func BenchmarkFormatCacheGet(b *testing.B) {
	cache := newBenchCache(10000)
	defer cache.Stop()

	req := benchFormatRequest(0)
	cache.Set(req, benchFormatResult())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get(req)
	}
}

// BenchmarkFormatCacheSet measures the throughput of inserting new entries
// into a cache that is not at capacity.
func BenchmarkFormatCacheSet(b *testing.B) {
	cache := newBenchCache(b.N + 1)
	defer cache.Stop()

	requests := make([]*FormatRequest, b.N)
	for i := range requests {
		requests[i] = benchFormatRequest(i)
	}
	result := benchFormatResult()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(requests[i], result)
	}
}

// BenchmarkFormatCacheSetAtCapacity measures the throughput of inserting
// entries when the cache is full, forcing eviction on every Set call.
func BenchmarkFormatCacheSetAtCapacity(b *testing.B) {
	const capacity = 100
	cache := newBenchCache(capacity)
	defer cache.Stop()

	// Fill cache to capacity.
	result := benchFormatResult()
	for i := 0; i < capacity; i++ {
		cache.Set(benchFormatRequest(i), result)
	}

	// Prepare unique requests that will not hit existing keys.
	requests := make([]*FormatRequest, b.N)
	for i := range requests {
		requests[i] = benchFormatRequest(capacity + i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(requests[i], result)
	}
}

// BenchmarkFormatCacheConcurrentReadWrite measures concurrent mixed
// read and write throughput using b.RunParallel.
func BenchmarkFormatCacheConcurrentReadWrite(b *testing.B) {
	cache := newBenchCache(10000)
	defer cache.Stop()

	// Pre-populate with some entries for reads to hit.
	result := benchFormatResult()
	for i := 0; i < 500; i++ {
		cache.Set(benchFormatRequest(i), result)
	}

	var counter uint64
	var mu sync.Mutex

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine gets a local counter to generate unique keys.
		mu.Lock()
		counter++
		localID := counter
		mu.Unlock()

		i := int(localID) * 1_000_000
		for pb.Next() {
			if i%2 == 0 {
				// Write with a unique key.
				cache.Set(benchFormatRequest(i), result)
			} else {
				// Read an existing key.
				cache.Get(benchFormatRequest(i % 500))
			}
			i++
		}
	})
}
