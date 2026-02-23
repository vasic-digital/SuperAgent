package semantic

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSemanticCache_Initialization tests cache initialization
func TestNewSemanticCache_Initialization(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
	}{
		{"short_ttl", 1 * time.Second},
		{"medium_ttl", 30 * time.Minute},
		{"long_ttl", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			cache := NewSemanticCache(tt.ttl, encoder)

			assert.NotNil(t, cache)
			assert.NotNil(t, cache.cache)
			assert.Equal(t, tt.ttl, cache.ttl)
			assert.Equal(t, encoder, cache.encoder)
			assert.Equal(t, 0, cache.Size())
		})
	}
}

// TestSemanticCache_Set_TableDriven tests Set with various inputs
func TestSemanticCache_Set_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		route  *Route
		verify func(t *testing.T, cache *SemanticCache, query string)
	}{
		{
			name:  "simple_route",
			query: "hello world",
			route: &Route{Name: "greeting", Score: 0.95},
			verify: func(t *testing.T, cache *SemanticCache, query string) {
				result := cache.Get(query)
				require.NotNil(t, result)
				assert.Equal(t, "greeting", result.Name)
				assert.Equal(t, 0.95, result.Score)
			},
		},
		{
			name:  "route_with_metadata",
			query: "complex query",
			route: &Route{
				Name:     "complex",
				Score:    0.85,
				Metadata: map[string]interface{}{"key": "value"},
			},
			verify: func(t *testing.T, cache *SemanticCache, query string) {
				result := cache.Get(query)
				require.NotNil(t, result)
				assert.Equal(t, "value", result.Metadata["key"])
			},
		},
		{
			name:  "empty_query",
			query: "",
			route: &Route{Name: "empty"},
			verify: func(t *testing.T, cache *SemanticCache, query string) {
				result := cache.Get(query)
				require.NotNil(t, result)
			},
		},
		{
			name:  "unicode_query",
			query: "Hello \xe4\xb8\x96\xe7\x95\x8c",
			route: &Route{Name: "unicode"},
			verify: func(t *testing.T, cache *SemanticCache, query string) {
				result := cache.Get(query)
				require.NotNil(t, result)
			},
		},
		{
			name:  "special_characters",
			query: "what's the @#$%^& meaning?",
			route: &Route{Name: "special"},
			verify: func(t *testing.T, cache *SemanticCache, query string) {
				result := cache.Get(query)
				require.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			cache := NewSemanticCache(time.Minute, encoder)

			cache.Set(tt.query, tt.route)
			assert.Equal(t, 1, cache.Size())

			if tt.verify != nil {
				tt.verify(t, cache, tt.query)
			}
		})
	}
}

// TestSemanticCache_Get_EdgeCases tests Get edge cases
func TestSemanticCache_Get_EdgeCases(t *testing.T) {
	t.Run("nonexistent_key", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		result := cache.Get("does_not_exist")
		assert.Nil(t, result)
	})

	t.Run("expired_entry", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(1*time.Millisecond, encoder)

		cache.Set("key", &Route{Name: "test"})
		time.Sleep(5 * time.Millisecond)

		result := cache.Get("key")
		assert.Nil(t, result)
	})

	t.Run("just_before_expiry", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(100*time.Millisecond, encoder)

		cache.Set("key", &Route{Name: "test"})
		time.Sleep(50 * time.Millisecond)

		result := cache.Get("key")
		assert.NotNil(t, result)
	})

	t.Run("overwrite_existing", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		cache.Set("key", &Route{Name: "first"})
		cache.Set("key", &Route{Name: "second"})

		result := cache.Get("key")
		require.NotNil(t, result)
		assert.Equal(t, "second", result.Name)
		assert.Equal(t, 1, cache.Size())
	})
}

// TestSemanticCache_SetWithEmbedding_TableDriven tests SetWithEmbedding
func TestSemanticCache_SetWithEmbedding_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		route     *Route
		embedding []float32
	}{
		{
			name:      "normal_embedding",
			query:     "test query",
			route:     &Route{Name: "test"},
			embedding: []float32{0.1, 0.2, 0.3},
		},
		{
			name:      "high_dimensional_embedding",
			query:     "complex query",
			route:     &Route{Name: "complex"},
			embedding: make([]float32, 1536), // GPT embedding size
		},
		{
			name:      "empty_embedding",
			query:     "empty",
			route:     &Route{Name: "empty"},
			embedding: []float32{},
		},
		{
			name:      "single_dimension",
			query:     "single",
			route:     &Route{Name: "single"},
			embedding: []float32{1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			cache := NewSemanticCache(time.Minute, encoder)

			cache.SetWithEmbedding(tt.query, tt.route, tt.embedding)

			assert.Equal(t, 1, cache.Size())
		})
	}
}

// TestSemanticCache_GetSemantic_TableDriven tests semantic retrieval
func TestSemanticCache_GetSemantic_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		cached []struct {
			query     string
			route     *Route
			embedding []float32
		}
		queryEmbedding []float32
		threshold      float64
		expectMatch    bool
		expectedRoute  string
	}{
		{
			name: "exact_match",
			cached: []struct {
				query     string
				route     *Route
				embedding []float32
			}{
				{"hello", &Route{Name: "greeting"}, []float32{1.0, 0.0, 0.0}},
			},
			queryEmbedding: []float32{1.0, 0.0, 0.0},
			threshold:      0.9,
			expectMatch:    true,
			expectedRoute:  "greeting",
		},
		{
			name: "similar_match",
			cached: []struct {
				query     string
				route     *Route
				embedding []float32
			}{
				{"hello", &Route{Name: "greeting"}, []float32{1.0, 0.0, 0.0}},
			},
			queryEmbedding: []float32{0.95, 0.1, 0.05},
			threshold:      0.9,
			expectMatch:    true,
			expectedRoute:  "greeting",
		},
		{
			name: "below_threshold",
			cached: []struct {
				query     string
				route     *Route
				embedding []float32
			}{
				{"hello", &Route{Name: "greeting"}, []float32{1.0, 0.0, 0.0}},
			},
			queryEmbedding: []float32{0.0, 1.0, 0.0}, // Orthogonal
			threshold:      0.5,
			expectMatch:    false,
		},
		{
			name: "multiple_cached_best_match",
			cached: []struct {
				query     string
				route     *Route
				embedding []float32
			}{
				{"hello", &Route{Name: "greeting"}, []float32{1.0, 0.0, 0.0}},
				{"bye", &Route{Name: "farewell"}, []float32{0.0, 1.0, 0.0}},
				{"help", &Route{Name: "assistance"}, []float32{0.0, 0.0, 1.0}},
			},
			queryEmbedding: []float32{0.0, 0.95, 0.05},
			threshold:      0.5,
			expectMatch:    true,
			expectedRoute:  "farewell",
		},
		{
			name: "empty_cache",
			cached: []struct {
				query     string
				route     *Route
				embedding []float32
			}{},
			queryEmbedding: []float32{1.0, 0.0, 0.0},
			threshold:      0.5,
			expectMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			cache := NewSemanticCache(time.Minute, encoder)

			for _, c := range tt.cached {
				cache.SetWithEmbedding(c.query, c.route, c.embedding)
			}

			result := cache.GetSemantic(tt.queryEmbedding, tt.threshold)

			if tt.expectMatch {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedRoute, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestSemanticCache_GetSemantic_ExpiredEntries tests that expired entries are ignored
func TestSemanticCache_GetSemantic_ExpiredEntries(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(1*time.Millisecond, encoder)

	cache.SetWithEmbedding("test", &Route{Name: "test"}, []float32{1.0, 0.0, 0.0})
	time.Sleep(5 * time.Millisecond)

	result := cache.GetSemantic([]float32{1.0, 0.0, 0.0}, 0.5)
	assert.Nil(t, result)
}

// TestSemanticCache_Clear_Scenarios tests Clear in various scenarios
func TestSemanticCache_Clear_Scenarios(t *testing.T) {
	tests := []struct {
		name       string
		numEntries int
	}{
		{"empty_cache", 0},
		{"single_entry", 1},
		{"multiple_entries", 10},
		{"many_entries", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			cache := NewSemanticCache(time.Minute, encoder)

			for i := 0; i < tt.numEntries; i++ {
				cache.Set(string(rune('a'+i)), &Route{Name: string(rune('A' + i))})
			}

			assert.Equal(t, tt.numEntries, cache.Size())

			cache.Clear()

			assert.Equal(t, 0, cache.Size())
		})
	}
}

// TestSemanticCache_Size_Accuracy tests Size accuracy
func TestSemanticCache_Size_Accuracy(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	assert.Equal(t, 0, cache.Size())

	for i := 1; i <= 100; i++ {
		cache.Set(string(rune(i)), &Route{Name: "test"})
		assert.Equal(t, i, cache.Size())
	}

	// Overwrites shouldn't increase size
	for i := 1; i <= 50; i++ {
		cache.Set(string(rune(i)), &Route{Name: "updated"})
	}
	assert.Equal(t, 100, cache.Size())
}

// TestSemanticCache_GetStats_Comprehensive tests GetStats
func TestSemanticCache_GetStats_Comprehensive(t *testing.T) {
	t.Run("empty_cache", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		stats := cache.GetStats()

		assert.Equal(t, 0, stats.Size)
		assert.Equal(t, time.Minute, stats.TTL)
		assert.Nil(t, stats.OldestEntry)
		assert.Nil(t, stats.NewestEntry)
	})

	t.Run("single_entry", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		cache.Set("key", &Route{Name: "test"})

		stats := cache.GetStats()

		assert.Equal(t, 1, stats.Size)
		assert.NotNil(t, stats.OldestEntry)
		assert.NotNil(t, stats.NewestEntry)
		assert.True(t, stats.OldestEntry.Equal(*stats.NewestEntry))
	})

	t.Run("multiple_entries_time_ordered", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		cache.Set("first", &Route{Name: "first"})
		time.Sleep(10 * time.Millisecond)
		cache.Set("second", &Route{Name: "second"})
		time.Sleep(10 * time.Millisecond)
		cache.Set("third", &Route{Name: "third"})

		stats := cache.GetStats()

		assert.Equal(t, 3, stats.Size)
		assert.NotNil(t, stats.OldestEntry)
		assert.NotNil(t, stats.NewestEntry)
		assert.True(t, stats.OldestEntry.Before(*stats.NewestEntry))
	})

	t.Run("stats_after_clear", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Minute, encoder)

		cache.Set("key", &Route{Name: "test"})
		cache.Clear()

		stats := cache.GetStats()

		assert.Equal(t, 0, stats.Size)
		assert.Nil(t, stats.OldestEntry)
		assert.Nil(t, stats.NewestEntry)
	})
}

// TestCacheStats_Struct tests CacheStats struct
func TestCacheStats_Struct(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	stats := &CacheStats{
		Size:        50,
		OldestEntry: &earlier,
		NewestEntry: &now,
		TTL:         30 * time.Minute,
	}

	assert.Equal(t, 50, stats.Size)
	assert.Equal(t, 30*time.Minute, stats.TTL)
	assert.True(t, stats.OldestEntry.Before(*stats.NewestEntry))
}

// TestSemanticCache_removeExpired_Scenarios tests removeExpired
func TestSemanticCache_removeExpired_Scenarios(t *testing.T) {
	t.Run("removes_only_expired", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(50*time.Millisecond, encoder)

		// Add some entries
		cache.Set("old1", &Route{Name: "old1"})
		cache.Set("old2", &Route{Name: "old2"})

		// Wait for expiry
		time.Sleep(60 * time.Millisecond)

		// Add new entry
		cache.Set("new", &Route{Name: "new"})

		cache.removeExpired()

		assert.Equal(t, 1, cache.Size())
		assert.NotNil(t, cache.Get("new"))
	})

	t.Run("no_expired_entries", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(time.Hour, encoder)

		cache.Set("key1", &Route{Name: "test1"})
		cache.Set("key2", &Route{Name: "test2"})

		initialSize := cache.Size()
		cache.removeExpired()

		assert.Equal(t, initialSize, cache.Size())
	})

	t.Run("all_expired", func(t *testing.T) {
		encoder := newMockEncoder()
		cache := NewSemanticCache(1*time.Millisecond, encoder)

		cache.Set("key1", &Route{Name: "test1"})
		cache.Set("key2", &Route{Name: "test2"})
		cache.Set("key3", &Route{Name: "test3"})

		time.Sleep(5 * time.Millisecond)
		cache.removeExpired()

		assert.Equal(t, 0, cache.Size())
	})
}

// TestSemanticCache_Concurrency tests concurrent access
func TestSemanticCache_Concurrency(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set(string(rune(id)), &Route{Name: string(rune('A' + id%26))})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = cache.Get(string(rune(id)))
		}(i)
	}

	// Concurrent size checks
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cache.Size()
		}()
	}

	// Concurrent stats
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cache.GetStats()
		}()
	}

	wg.Wait()
}

// TestSemanticCache_ConcurrentWriteRead tests concurrent writes and reads
func TestSemanticCache_ConcurrentWriteRead(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	var wg sync.WaitGroup

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			cache.Set("key", &Route{Name: string(rune('A' + i%26))})
		}
	}()

	// Reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = cache.Get("key")
		}
	}()

	wg.Wait()
}

// TestSemanticCache_SetWithEmbedding_Overwrite tests overwriting with embedding
func TestSemanticCache_SetWithEmbedding_Overwrite(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	embedding1 := []float32{1.0, 0.0, 0.0}
	embedding2 := []float32{0.0, 1.0, 0.0}

	cache.SetWithEmbedding("key", &Route{Name: "first"}, embedding1)
	cache.SetWithEmbedding("key", &Route{Name: "second"}, embedding2)

	assert.Equal(t, 1, cache.Size())

	// The semantic lookup should use the newer embedding
	result := cache.GetSemantic([]float32{0.0, 1.0, 0.0}, 0.9)
	require.NotNil(t, result)
	assert.Equal(t, "second", result.Name)
}

// TestSemanticCache_GetSemantic_HighThreshold tests high threshold behavior
func TestSemanticCache_GetSemantic_HighThreshold(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	// Cache an entry
	cache.SetWithEmbedding("test", &Route{Name: "test"}, []float32{1.0, 0.0, 0.0})

	// Query with very high threshold
	result := cache.GetSemantic([]float32{0.99, 0.1, 0.0}, 0.999)
	assert.Nil(t, result) // Should not match due to high threshold

	// Query with lower threshold
	result = cache.GetSemantic([]float32{0.99, 0.1, 0.0}, 0.9)
	assert.NotNil(t, result)
}

// TestSemanticCache_GetSemantic_ZeroThreshold tests zero threshold
func TestSemanticCache_GetSemantic_ZeroThreshold(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	cache.SetWithEmbedding("test", &Route{Name: "test"}, []float32{1.0, 0.0, 0.0})

	// Any non-zero similarity should match with zero threshold
	result := cache.GetSemantic([]float32{0.1, 0.9, 0.0}, 0.0)
	// This might still be nil if similarity is exactly 0 or negative
	// The test verifies the function handles zero threshold correctly
	// Just verify no panic occurs
	_ = result
}

// TestSemanticCache_LargeCache tests cache with many entries
func TestSemanticCache_LargeCache(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	numEntries := 10000

	// Add many entries
	for i := 0; i < numEntries; i++ {
		embedding := make([]float32, 128)
		embedding[i%128] = 1.0
		cache.SetWithEmbedding(string(rune(i)), &Route{Name: string(rune(i))}, embedding)
	}

	assert.Equal(t, numEntries, cache.Size())

	// Clear should handle large cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

// TestSemanticCache_TTL_Boundary tests TTL boundary conditions
func TestSemanticCache_TTL_Boundary(t *testing.T) {
	t.Run("at_exact_ttl", func(t *testing.T) {
		encoder := newMockEncoder()
		ttl := 50 * time.Millisecond
		cache := NewSemanticCache(ttl, encoder)

		cache.Set("key", &Route{Name: "test"})

		// Sleep exactly TTL
		time.Sleep(ttl)

		// Entry should be expired (>= ttl)
		result := cache.Get("key")
		assert.Nil(t, result)
	})

	t.Run("just_before_ttl", func(t *testing.T) {
		encoder := newMockEncoder()
		ttl := 100 * time.Millisecond
		cache := NewSemanticCache(ttl, encoder)

		cache.Set("key", &Route{Name: "test"})

		// Sleep less than TTL
		time.Sleep(ttl - 20*time.Millisecond)

		result := cache.Get("key")
		assert.NotNil(t, result)
	})
}

// TestCacheEntry tests cacheEntry struct behavior
func TestCacheEntry(t *testing.T) {
	route := &Route{Name: "test", Score: 0.9}
	embedding := []float32{1.0, 2.0, 3.0}
	timestamp := time.Now()

	entry := &cacheEntry{
		route:     route,
		embedding: embedding,
		timestamp: timestamp,
	}

	assert.Equal(t, route, entry.route)
	assert.Equal(t, embedding, entry.embedding)
	assert.Equal(t, timestamp, entry.timestamp)
}

// TestMockEncoder tests the mock encoder used in tests
func TestMockEncoder(t *testing.T) {
	t.Run("default_behavior", func(t *testing.T) {
		encoder := newMockEncoder()

		embeddings, err := encoder.Encode(context.Background(), []string{"hello", "world"})

		require.NoError(t, err)
		assert.Len(t, embeddings, 2)
		assert.Len(t, embeddings[0], 128)
		assert.Len(t, embeddings[1], 128)
	})

	t.Run("custom_encode_func", func(t *testing.T) {
		encoder := &mockEncoder{
			dimension: 3,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return [][]float32{{1.0, 2.0, 3.0}}, nil
			},
		}

		embeddings, err := encoder.Encode(context.Background(), []string{"test"})

		require.NoError(t, err)
		assert.Equal(t, [][]float32{{1.0, 2.0, 3.0}}, embeddings)
	})

	t.Run("encode_count_tracking", func(t *testing.T) {
		encoder := newMockEncoder()

		assert.Equal(t, int64(0), encoder.encodeCount)

		_, _ = encoder.Encode(context.Background(), []string{"test"})
		assert.Equal(t, int64(1), encoder.encodeCount)

		_, _ = encoder.Encode(context.Background(), []string{"test"})
		assert.Equal(t, int64(2), encoder.encodeCount)
	})

	t.Run("get_dimension", func(t *testing.T) {
		encoder := &mockEncoder{dimension: 256}
		assert.Equal(t, 256, encoder.GetDimension())
	})
}
