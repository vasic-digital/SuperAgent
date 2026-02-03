package semantic

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// benchEncoder implements Encoder interface for benchmarking
// (named differently from mockEncoder in semantic_test.go to avoid conflicts)
type benchEncoder struct {
	dimension int
	latency   time.Duration
}

func newBenchEncoder(dimension int) *benchEncoder {
	return &benchEncoder{dimension: dimension}
}

func (m *benchEncoder) Encode(_ context.Context, texts []string) ([][]float32, error) {
	if m.latency > 0 {
		time.Sleep(m.latency)
	}
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = generateRandomEmbedding(m.dimension)
	}
	return embeddings, nil
}

func (m *benchEncoder) GetDimension() int {
	return m.dimension
}

// BenchmarkCosineSimilarity benchmarks the cosine similarity calculation
func BenchmarkCosineSimilarity(b *testing.B) {
	dimensions := []int{384, 768, 1536, 3072}

	for _, dim := range dimensions {
		a := generateRandomEmbedding(dim)
		b2 := generateRandomEmbedding(dim)

		b.Run("Dimension_"+string(rune(dim)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = cosineSimilarity(a, b2)
			}
		})
	}
}

// BenchmarkCosineSimilarityEdgeCases benchmarks edge cases in cosine similarity
func BenchmarkCosineSimilarityEdgeCases(b *testing.B) {
	dim := 1536

	b.Run("IdenticalVectors", func(b *testing.B) {
		a := generateRandomEmbedding(dim)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cosineSimilarity(a, a)
		}
	})

	b.Run("OrthogonalVectors", func(b *testing.B) {
		a := make([]float32, dim)
		b2 := make([]float32, dim)
		for i := 0; i < dim/2; i++ {
			a[i] = 1.0
			b2[dim/2+i] = 1.0
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cosineSimilarity(a, b2)
		}
	})

	b.Run("ZeroVector", func(b *testing.B) {
		a := generateRandomEmbedding(dim)
		zero := make([]float32, dim)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cosineSimilarity(a, zero)
		}
	})
}

// BenchmarkSqrt benchmarks the custom sqrt function
func BenchmarkSqrt(b *testing.B) {
	values := []float64{0.5, 1.0, 2.0, 100.0, 10000.0}

	for _, v := range values {
		b.Run("Value_"+string(rune(int(v))), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = sqrt(v)
			}
		})
	}
}

// BenchmarkSqrtVsStdlib benchmarks custom sqrt against math.Sqrt
func BenchmarkSqrtVsStdlib(b *testing.B) {
	value := 123.456

	b.Run("CustomSqrt", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = sqrt(value)
		}
	})

	b.Run("MathSqrt", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = math.Sqrt(value)
		}
	})
}

// BenchmarkDefaultRouterConfig benchmarks creating default router config
func BenchmarkDefaultRouterConfig(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultRouterConfig()
	}
}

// BenchmarkRouterCreation benchmarks creating a new router
func BenchmarkRouterCreation(b *testing.B) {
	encoder := newBenchEncoder(1536)

	b.Run("WithDefaultConfig", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRouter(encoder, nil, nil)
		}
	})

	b.Run("WithCustomConfig", func(b *testing.B) {
		config := &RouterConfig{
			ScoreThreshold:    0.8,
			TopK:              10,
			EnableCache:       true,
			CacheTTL:          time.Hour,
			FallbackRoute:     "default",
			AggregationMethod: AggregationMax,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRouter(encoder, config, nil)
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		config := &RouterConfig{
			ScoreThreshold:    0.7,
			TopK:              5,
			EnableCache:       false,
			AggregationMethod: AggregationMean,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRouter(encoder, config, nil)
		}
	})
}

// BenchmarkAggregateEmbeddings benchmarks embedding aggregation
func BenchmarkAggregateEmbeddings(b *testing.B) {
	encoder := newBenchEncoder(1536)
	router := NewRouter(encoder, nil, nil)

	embeddingCounts := []int{1, 5, 10, 20}
	dim := 1536

	for _, count := range embeddingCounts {
		embeddings := make([][]float32, count)
		for i := 0; i < count; i++ {
			embeddings[i] = generateRandomEmbedding(dim)
		}

		b.Run("Mean_Count_"+string(rune(count)), func(b *testing.B) {
			router.config.AggregationMethod = AggregationMean
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = router.aggregateEmbeddings(embeddings)
			}
		})

		b.Run("Max_Count_"+string(rune(count)), func(b *testing.B) {
			router.config.AggregationMethod = AggregationMax
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = router.aggregateEmbeddings(embeddings)
			}
		})
	}
}

// BenchmarkRouteListOperations benchmarks route list management
func BenchmarkRouteListOperations(b *testing.B) {
	encoder := newBenchEncoder(1536)

	b.Run("ListRoutes_Empty", func(b *testing.B) {
		router := NewRouter(encoder, nil, nil)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = router.ListRoutes()
		}
	})

	b.Run("ListRoutes_10", func(b *testing.B) {
		router := NewRouter(encoder, nil, nil)
		for j := 0; j < 10; j++ {
			route := &Route{
				Name:       "route-" + string(rune(j)),
				Utterances: []string{"test utterance"},
				Embedding:  generateRandomEmbedding(1536),
			}
			router.routes = append(router.routes, route)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = router.ListRoutes()
		}
	})
}

// BenchmarkSemanticCacheOperations benchmarks cache operations
func BenchmarkSemanticCacheOperations(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	route := &Route{
		Name:       "test-route",
		Embedding:  generateRandomEmbedding(1536),
		Utterances: []string{"test"},
	}

	b.Run("Get_Miss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.Get("nonexistent-query")
		}
	})

	b.Run("Set", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Set("query-"+string(rune(i)), route)
		}
	})

	b.Run("Get_Hit", func(b *testing.B) {
		cache.Set("cached-query", route)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.Get("cached-query")
		}
	})

	b.Run("SetWithEmbedding", func(b *testing.B) {
		embedding := generateRandomEmbedding(1536)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.SetWithEmbedding("query-"+string(rune(i)), route, embedding)
		}
	})
}

// BenchmarkSemanticCacheGetSemantic benchmarks semantic similarity lookup in cache
func BenchmarkSemanticCacheGetSemantic(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	// Populate cache with entries
	for i := 0; i < 100; i++ {
		route := &Route{
			Name:       "route-" + string(rune(i)),
			Embedding:  generateRandomEmbedding(1536),
			Utterances: []string{"test"},
		}
		cache.SetWithEmbedding("query-"+string(rune(i)), route, generateRandomEmbedding(1536))
	}

	queryEmbedding := generateRandomEmbedding(1536)

	b.Run("Threshold_0.7", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.GetSemantic(queryEmbedding, 0.7)
		}
	})

	b.Run("Threshold_0.9", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.GetSemantic(queryEmbedding, 0.9)
		}
	})
}

// BenchmarkCacheSize benchmarks getting cache size
func BenchmarkCacheSize(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	// Populate cache
	for i := 0; i < 1000; i++ {
		route := &Route{Name: "route-" + string(rune(i))}
		cache.Set("query-"+string(rune(i)), route)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Size()
	}
}

// BenchmarkCacheStats benchmarks getting cache statistics
func BenchmarkCacheStats(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	// Populate cache
	for i := 0; i < 100; i++ {
		route := &Route{Name: "route-" + string(rune(i))}
		cache.Set("query-"+string(rune(i)), route)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.GetStats()
	}
}

// BenchmarkCacheClear benchmarks clearing the cache
func BenchmarkCacheClear(b *testing.B) {
	encoder := newBenchEncoder(1536)

	b.Run("EmptyCache", func(b *testing.B) {
		cache := NewSemanticCache(30*time.Minute, encoder)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Clear()
		}
	})

	b.Run("PopulatedCache", func(b *testing.B) {
		cache := NewSemanticCache(30*time.Minute, encoder)
		for i := 0; i < 100; i++ {
			route := &Route{Name: "route-" + string(rune(i))}
			cache.Set("query-"+string(rune(i)), route)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Clear()
		}
	})
}

// BenchmarkRouteScoring benchmarks scoring multiple routes against a query
func BenchmarkRouteScoring(b *testing.B) {
	routeCounts := []int{5, 10, 20, 50}
	dim := 1536

	for _, count := range routeCounts {
		routes := make([]*Route, count)
		for i := 0; i < count; i++ {
			routes[i] = &Route{
				Name:      "route-" + string(rune(i)),
				Embedding: generateRandomEmbedding(dim),
			}
		}
		queryEmbedding := generateRandomEmbedding(dim)

		b.Run("RouteCount_"+string(rune(count)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				type routeScore struct {
					route *Route
					score float64
				}
				scores := make([]routeScore, len(routes))
				for j, route := range routes {
					score := cosineSimilarity(queryEmbedding, route.Embedding)
					scores[j] = routeScore{route: route, score: score}
				}
			}
		})
	}
}

// BenchmarkConcurrentCacheAccess benchmarks concurrent cache operations
func BenchmarkConcurrentCacheAccess(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		route := &Route{Name: "route-" + string(rune(i))}
		cache.Set("query-"+string(rune(i)), route)
	}

	b.Run("ParallelGet", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_ = cache.Get("query-" + string(rune(i%100)))
				i++
			}
		})
	})

	b.Run("ParallelSet", func(b *testing.B) {
		route := &Route{Name: "test-route"}
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				cache.Set("concurrent-query-"+string(rune(i)), route)
				i++
			}
		})
	})

	b.Run("MixedReadWrite", func(b *testing.B) {
		route := &Route{Name: "test-route"}
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					_ = cache.Get("query-" + string(rune(i%100)))
				} else {
					cache.Set("mixed-query-"+string(rune(i)), route)
				}
				i++
			}
		})
	})
}

// BenchmarkConcurrentRouterAccess benchmarks concurrent router operations
func BenchmarkConcurrentRouterAccess(b *testing.B) {
	encoder := newBenchEncoder(1536)
	config := &RouterConfig{
		ScoreThreshold:    0.7,
		TopK:              5,
		EnableCache:       false, // Disable cache for pure routing benchmarks
		AggregationMethod: AggregationMean,
	}
	router := NewRouter(encoder, config, nil)

	// Add routes directly
	for i := 0; i < 10; i++ {
		route := &Route{
			Name:       "route-" + string(rune(i)),
			Utterances: []string{"test"},
			Embedding:  generateRandomEmbedding(1536),
		}
		router.routes = append(router.routes, route)
	}

	b.Run("ParallelListRoutes", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = router.ListRoutes()
			}
		})
	})
}

// BenchmarkRouteCreation benchmarks creating route structures
func BenchmarkRouteCreation(b *testing.B) {
	b.Run("MinimalRoute", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Route{
				Name:       "test-route",
				Utterances: []string{"hello"},
			}
		}
	})

	b.Run("FullRoute", func(b *testing.B) {
		embedding := generateRandomEmbedding(1536)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Route{
				Name:        "test-route",
				Description: "A test route for benchmarking purposes",
				Utterances:  []string{"hello", "hi", "greetings", "hey there"},
				Metadata: map[string]interface{}{
					"category": "greeting",
					"priority": 1,
				},
				ModelTier: ModelTierSimple,
				Embedding: embedding,
			}
		}
	})
}

// BenchmarkRouteResultCreation benchmarks creating route result structures
func BenchmarkRouteResultCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &RouteResult{
			Content:  "This is a response",
			Model:    "claude-3-haiku",
			Metadata: map[string]interface{}{"key": "value"},
			CacheKey: "cache-key-123",
			Latency:  100 * time.Millisecond,
		}
	}
}

// BenchmarkMinFunction benchmarks the min helper function
func BenchmarkMinFunction(b *testing.B) {
	pairs := [][2]int{{5, 10}, {100, 50}, {0, 0}, {1000000, 1}}

	for _, pair := range pairs {
		b.Run("Min_"+string(rune(pair[0]))+"_"+string(rune(pair[1])), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = min(pair[0], pair[1])
			}
		})
	}
}

// generateRandomEmbedding creates a random float32 embedding for testing
func generateRandomEmbedding(size int) []float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	embedding := make([]float32, size)
	var sum float32
	for i := 0; i < size; i++ {
		embedding[i] = r.Float32()*2 - 1
		sum += embedding[i] * embedding[i]
	}
	// Normalize
	norm := float32(math.Sqrt(float64(sum)))
	if norm > 0 {
		for i := 0; i < size; i++ {
			embedding[i] /= norm
		}
	}
	return embedding
}

// BenchmarkEmbeddingNormalization benchmarks vector normalization
func BenchmarkEmbeddingNormalization(b *testing.B) {
	dimensions := []int{384, 768, 1536}

	for _, dim := range dimensions {
		raw := make([]float32, dim)
		r := rand.New(rand.NewSource(42))
		for i := range raw {
			raw[i] = r.Float32()*2 - 1
		}

		b.Run("Dimension_"+string(rune(dim)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				embedding := make([]float32, dim)
				copy(embedding, raw)
				var sum float32
				for j := range embedding {
					sum += embedding[j] * embedding[j]
				}
				norm := float32(math.Sqrt(float64(sum)))
				if norm > 0 {
					for j := range embedding {
						embedding[j] /= norm
					}
				}
			}
		})
	}
}

// BenchmarkCacheExpiredCheck benchmarks the expired entry check in cache
func BenchmarkCacheExpiredCheck(b *testing.B) {
	encoder := newBenchEncoder(1536)
	cache := NewSemanticCache(30*time.Minute, encoder)

	// Add entries with various timestamps
	route := &Route{Name: "test"}
	cache.Set("query", route)

	b.Run("SingleEntry", func(b *testing.B) {
		cache.mu.RLock()
		entry := cache.cache["query"]
		cache.mu.RUnlock()
		timestamp := entry.timestamp

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = time.Since(timestamp) >= cache.ttl
		}
	})
}

// BenchmarkCacheEntryCreation benchmarks creating cache entries
func BenchmarkCacheEntryCreation(b *testing.B) {
	route := &Route{Name: "test-route"}
	embedding := generateRandomEmbedding(1536)

	b.Run("WithoutEmbedding", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &cacheEntry{
				route:     route,
				timestamp: time.Now(),
			}
		}
	})

	b.Run("WithEmbedding", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &cacheEntry{
				route:     route,
				embedding: embedding,
				timestamp: time.Now(),
			}
		}
	})
}

// BenchmarkLockContention benchmarks lock contention scenarios
func BenchmarkLockContention(b *testing.B) {
	var mu sync.RWMutex
	data := make(map[string]int)
	data["key"] = 42

	b.Run("ReadLock", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mu.RLock()
			_ = data["key"]
			mu.RUnlock()
		}
	})

	b.Run("WriteLock", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mu.Lock()
			data["key"] = i
			mu.Unlock()
		}
	})

	b.Run("ParallelRead", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.RLock()
				_ = data["key"]
				mu.RUnlock()
			}
		})
	})
}
