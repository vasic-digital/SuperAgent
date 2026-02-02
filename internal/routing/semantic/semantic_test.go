package semantic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock encoder for testing
type mockEncoder struct {
	dimension   int
	encodeFunc  func(ctx context.Context, texts []string) ([][]float32, error)
	encodeCount int
}

func (e *mockEncoder) Encode(ctx context.Context, texts []string) ([][]float32, error) {
	e.encodeCount++
	if e.encodeFunc != nil {
		return e.encodeFunc(ctx, texts)
	}

	// Generate simple embeddings based on text length
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding := make([]float32, e.dimension)
		for j := 0; j < e.dimension; j++ {
			// Create unique embeddings based on text
			embedding[j] = float32(len(text)+i+j) * 0.01
		}
		results[i] = embedding
	}
	return results, nil
}

func (e *mockEncoder) GetDimension() int {
	return e.dimension
}

func newMockEncoder() *mockEncoder {
	return &mockEncoder{dimension: 128}
}

// Tests for ModelTier constants
func TestModelTier(t *testing.T) {
	assert.Equal(t, ModelTier("simple"), ModelTierSimple)
	assert.Equal(t, ModelTier("standard"), ModelTierStandard)
	assert.Equal(t, ModelTier("complex"), ModelTierComplex)
}

// Tests for AggregationMethod constants
func TestAggregationMethod(t *testing.T) {
	assert.Equal(t, AggregationMethod("mean"), AggregationMean)
	assert.Equal(t, AggregationMethod("max"), AggregationMax)
}

// Tests for DefaultRouterConfig
func TestDefaultRouterConfig(t *testing.T) {
	config := DefaultRouterConfig()

	assert.Equal(t, 0.7, config.ScoreThreshold)
	assert.Equal(t, 5, config.TopK)
	assert.True(t, config.EnableCache)
	assert.Equal(t, 30*time.Minute, config.CacheTTL)
	assert.Empty(t, config.FallbackRoute)
	assert.Equal(t, AggregationMean, config.AggregationMethod)
}

// Tests for NewRouter
func TestNewRouter(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		encoder := newMockEncoder()
		router := NewRouter(encoder, nil, nil)

		assert.NotNil(t, router)
		assert.NotNil(t, router.config)
		assert.NotNil(t, router.logger)
		assert.NotNil(t, router.cache)
		assert.Equal(t, encoder, router.encoder)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		encoder := newMockEncoder()
		config := &RouterConfig{
			ScoreThreshold: 0.8,
			TopK:           3,
			EnableCache:    false,
			CacheTTL:       10 * time.Minute,
		}
		logger := logrus.New()

		router := NewRouter(encoder, config, logger)

		assert.NotNil(t, router)
		assert.Equal(t, 0.8, router.config.ScoreThreshold)
		assert.Equal(t, 3, router.config.TopK)
		assert.Nil(t, router.cache)
		assert.Equal(t, logger, router.logger)
	})
}

// Tests for Route struct
func TestRouteStruct(t *testing.T) {
	handler := func(ctx context.Context, query string) (*RouteResult, error) {
		return &RouteResult{Content: "test"}, nil
	}

	route := &Route{
		Name:        "test_route",
		Description: "A test route",
		Utterances:  []string{"hello", "hi there"},
		Handler:     handler,
		Metadata:    map[string]interface{}{"key": "value"},
		ModelTier:   ModelTierSimple,
		Embedding:   []float32{0.1, 0.2, 0.3},
		Score:       0.85,
	}

	assert.Equal(t, "test_route", route.Name)
	assert.Equal(t, "A test route", route.Description)
	assert.Len(t, route.Utterances, 2)
	assert.NotNil(t, route.Handler)
	assert.Equal(t, ModelTierSimple, route.ModelTier)
	assert.Equal(t, 0.85, route.Score)
}

// Tests for RouteResult struct
func TestRouteResult(t *testing.T) {
	result := &RouteResult{
		Content:  "Test content",
		Model:    "gpt-4",
		Metadata: map[string]interface{}{"key": "value"},
		CacheKey: "cache-123",
		Latency:  100 * time.Millisecond,
	}

	assert.Equal(t, "Test content", result.Content)
	assert.Equal(t, "gpt-4", result.Model)
	assert.Equal(t, "cache-123", result.CacheKey)
	assert.Equal(t, 100*time.Millisecond, result.Latency)
}

// Tests for AddRoute
func TestRouter_AddRoute(t *testing.T) {
	encoder := newMockEncoder()
	router := NewRouter(encoder, nil, nil)

	t.Run("ValidRoute", func(t *testing.T) {
		route := &Route{
			Name:       "greeting",
			Utterances: []string{"hello", "hi", "hey"},
			ModelTier:  ModelTierSimple,
		}

		err := router.AddRoute(context.Background(), route)
		require.NoError(t, err)
		assert.NotNil(t, route.Embedding)
		assert.Len(t, router.ListRoutes(), 1)
	})

	t.Run("EmptyName", func(t *testing.T) {
		route := &Route{
			Utterances: []string{"test"},
		}

		err := router.AddRoute(context.Background(), route)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "route name is required")
	})

	t.Run("NoUtterances", func(t *testing.T) {
		route := &Route{
			Name: "empty_route",
		}

		err := router.AddRoute(context.Background(), route)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one utterance")
	})

	t.Run("EncoderError", func(t *testing.T) {
		failingEncoder := &mockEncoder{
			dimension: 128,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return nil, errors.New("encoding failed")
			},
		}
		router := NewRouter(failingEncoder, nil, nil)

		route := &Route{
			Name:       "test",
			Utterances: []string{"test"},
		}

		err := router.AddRoute(context.Background(), route)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to encode utterances")
	})
}

// Tests for Route (query routing)
func TestRouter_Route(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{
		ScoreThreshold: 0.0, // Low threshold to ensure matches
		TopK:           5,
		EnableCache:    true,
		CacheTTL:       time.Minute,
	}
	router := NewRouter(encoder, config, nil)

	// Add test routes
	routes := []*Route{
		{Name: "greeting", Utterances: []string{"hello", "hi", "hey"}, ModelTier: ModelTierSimple},
		{Name: "farewell", Utterances: []string{"goodbye", "bye", "see you"}, ModelTier: ModelTierSimple},
		{Name: "question", Utterances: []string{"what is", "how to", "why"}, ModelTier: ModelTierStandard},
	}

	for _, route := range routes {
		err := router.AddRoute(context.Background(), route)
		require.NoError(t, err)
	}

	t.Run("MatchesRoute", func(t *testing.T) {
		route, err := router.Route(context.Background(), "hello there")
		require.NoError(t, err)
		assert.NotNil(t, route)
	})

	t.Run("CacheHit", func(t *testing.T) {
		// First call
		route1, err := router.Route(context.Background(), "hello there")
		require.NoError(t, err)

		// Second call should hit cache
		route2, err := router.Route(context.Background(), "hello there")
		require.NoError(t, err)

		assert.Equal(t, route1.Name, route2.Name)
	})

	t.Run("NoRoutes", func(t *testing.T) {
		emptyRouter := NewRouter(encoder, nil, nil)
		_, err := emptyRouter.Route(context.Background(), "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no routes configured")
	})

	t.Run("EncoderError", func(t *testing.T) {
		failingEncoder := &mockEncoder{
			dimension: 128,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return nil, errors.New("encoding failed")
			},
		}
		router := NewRouter(failingEncoder, &RouterConfig{EnableCache: false, ScoreThreshold: 0.0}, nil)
		router.routes = []*Route{{Name: "test", Embedding: make([]float32, 128)}}

		_, err := router.Route(context.Background(), "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to encode query")
	})

	t.Run("EmptyEmbedding", func(t *testing.T) {
		emptyEncoder := &mockEncoder{
			dimension: 128,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return [][]float32{}, nil
			},
		}
		router := NewRouter(emptyEncoder, &RouterConfig{EnableCache: false, ScoreThreshold: 0.0}, nil)
		router.routes = []*Route{{Name: "test", Embedding: make([]float32, 128)}}

		_, err := router.Route(context.Background(), "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no embedding generated")
	})
}

// Tests for Route with fallback
func TestRouter_Route_Fallback(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{
		ScoreThreshold: 0.99, // Very high threshold
		EnableCache:    false,
		FallbackRoute:  "fallback",
	}
	router := NewRouter(encoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "fallback",
		Utterances: []string{"default"},
	})

	route, err := router.Route(context.Background(), "some random query")
	require.NoError(t, err)
	assert.Equal(t, "fallback", route.Name)
}

// Tests for Route without matching threshold
func TestRouter_Route_NoMatch(t *testing.T) {
	// Create an encoder that produces orthogonal vectors for different inputs
	// to ensure low cosine similarity
	callCount := 0
	orthogonalEncoder := &mockEncoder{
		dimension: 128,
		encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
			results := make([][]float32, len(texts))
			for i := range texts {
				embedding := make([]float32, 128)
				// First call (AddRoute with "hello") - vector in one direction
				// Second call (Route with query) - orthogonal vector
				if callCount == 0 {
					// First dimension hot for route embedding
					embedding[0] = 1.0
				} else {
					// Different dimension hot for query embedding - orthogonal
					embedding[64] = 1.0
				}
				results[i] = embedding
			}
			callCount++
			return results, nil
		},
	}

	config := &RouterConfig{
		ScoreThreshold: 0.5, // Even modest threshold won't match orthogonal vectors
		EnableCache:    false,
	}
	router := NewRouter(orthogonalEncoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	_, err := router.Route(context.Background(), "completely different query")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no route matched")
}

// Tests for RouteWithCandidates
func TestRouter_RouteWithCandidates(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{
		TopK:        3,
		EnableCache: false,
	}
	router := NewRouter(encoder, config, nil)

	// Add test routes
	for i := 0; i < 5; i++ {
		_ = router.AddRoute(context.Background(), &Route{
			Name:       string(rune('A' + i)),
			Utterances: []string{"test"},
		})
	}

	t.Run("ReturnsTopK", func(t *testing.T) {
		candidates, err := router.RouteWithCandidates(context.Background(), "test query")
		require.NoError(t, err)
		assert.Len(t, candidates, 3)
	})

	t.Run("NoRoutes", func(t *testing.T) {
		emptyRouter := NewRouter(encoder, nil, nil)
		_, err := emptyRouter.RouteWithCandidates(context.Background(), "test")
		require.Error(t, err)
	})
}

// Tests for RemoveRoute
func TestRouter_RemoveRoute(t *testing.T) {
	encoder := newMockEncoder()
	router := NewRouter(encoder, nil, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test1",
		Utterances: []string{"hello"},
	})
	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test2",
		Utterances: []string{"bye"},
	})

	assert.Len(t, router.ListRoutes(), 2)

	router.RemoveRoute("test1")
	assert.Len(t, router.ListRoutes(), 1)
	assert.Equal(t, "test2", router.ListRoutes()[0].Name)

	// Remove non-existent (should not panic)
	router.RemoveRoute("nonexistent")
	assert.Len(t, router.ListRoutes(), 1)
}

// Tests for ListRoutes
func TestRouter_ListRoutes(t *testing.T) {
	encoder := newMockEncoder()
	router := NewRouter(encoder, nil, nil)

	assert.Empty(t, router.ListRoutes())

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	routes := router.ListRoutes()
	assert.Len(t, routes, 1)
	assert.Equal(t, "test", routes[0].Name)
}

// Tests for ClearCache
func TestRouter_ClearCache(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{
		ScoreThreshold: 0.0,
		EnableCache:    true,
		CacheTTL:       time.Minute,
	}
	router := NewRouter(encoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	// Populate cache
	_, _ = router.Route(context.Background(), "hello")
	assert.Greater(t, router.cache.Size(), 0)

	router.ClearCache()
	assert.Equal(t, 0, router.cache.Size())
}

// Tests for ClearCache without cache
func TestRouter_ClearCache_NoCache(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{EnableCache: false}
	router := NewRouter(encoder, config, nil)

	// Should not panic
	router.ClearCache()
}

// Tests for aggregateEmbeddings
func TestRouter_aggregateEmbeddings(t *testing.T) {
	encoder := newMockEncoder()

	t.Run("MeanAggregation", func(t *testing.T) {
		config := &RouterConfig{AggregationMethod: AggregationMean, EnableCache: false}
		router := NewRouter(encoder, config, nil)

		embeddings := [][]float32{
			{1.0, 2.0, 3.0},
			{3.0, 4.0, 5.0},
		}

		result := router.aggregateEmbeddings(embeddings)
		assert.Equal(t, float32(2.0), result[0])
		assert.Equal(t, float32(3.0), result[1])
		assert.Equal(t, float32(4.0), result[2])
	})

	t.Run("MaxAggregation", func(t *testing.T) {
		config := &RouterConfig{AggregationMethod: AggregationMax, EnableCache: false}
		router := NewRouter(encoder, config, nil)

		embeddings := [][]float32{
			{1.0, 4.0, 3.0},
			{3.0, 2.0, 5.0},
		}

		result := router.aggregateEmbeddings(embeddings)
		assert.Equal(t, float32(3.0), result[0])
		assert.Equal(t, float32(4.0), result[1])
		assert.Equal(t, float32(5.0), result[2])
	})

	t.Run("EmptyEmbeddings", func(t *testing.T) {
		router := NewRouter(encoder, nil, nil)
		result := router.aggregateEmbeddings([][]float32{})
		assert.Nil(t, result)
	})
}

// Tests for cosineSimilarity
func TestCosineSimilarity(t *testing.T) {
	t.Run("IdenticalVectors", func(t *testing.T) {
		a := []float32{1.0, 2.0, 3.0}
		b := []float32{1.0, 2.0, 3.0}
		result := cosineSimilarity(a, b)
		assert.InDelta(t, 1.0, result, 0.001)
	})

	t.Run("OrthogonalVectors", func(t *testing.T) {
		a := []float32{1.0, 0.0}
		b := []float32{0.0, 1.0}
		result := cosineSimilarity(a, b)
		assert.InDelta(t, 0.0, result, 0.001)
	})

	t.Run("DifferentLengths", func(t *testing.T) {
		a := []float32{1.0, 2.0}
		b := []float32{1.0, 2.0, 3.0}
		result := cosineSimilarity(a, b)
		assert.Equal(t, 0.0, result)
	})

	t.Run("EmptyVectors", func(t *testing.T) {
		result := cosineSimilarity([]float32{}, []float32{})
		assert.Equal(t, 0.0, result)
	})

	t.Run("ZeroVector", func(t *testing.T) {
		a := []float32{0.0, 0.0, 0.0}
		b := []float32{1.0, 2.0, 3.0}
		result := cosineSimilarity(a, b)
		assert.Equal(t, 0.0, result)
	})
}

// Tests for sqrt
func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{4.0, 2.0},
		{9.0, 3.0},
		{16.0, 4.0},
		{0.0, 0.0},
		{-1.0, 0.0},
		{2.0, 1.414},
	}

	for _, tt := range tests {
		result := sqrt(tt.input)
		assert.InDelta(t, tt.expected, result, 0.01)
	}
}

// Tests for min
func TestMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 5))
	assert.Equal(t, 3, min(10, 3))
	assert.Equal(t, 5, min(5, 5))
	assert.Equal(t, -5, min(-5, 5))
}

// Tests for SemanticCache
func TestNewSemanticCache(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.cache)
	assert.Equal(t, time.Minute, cache.ttl)
}

func TestSemanticCache_SetGet(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	route := &Route{Name: "test", Score: 0.9}

	t.Run("SetAndGet", func(t *testing.T) {
		cache.Set("hello", route)

		result := cache.Get("hello")
		assert.NotNil(t, result)
		assert.Equal(t, "test", result.Name)
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		result := cache.Get("nonexistent")
		assert.Nil(t, result)
	})

	t.Run("GetExpired", func(t *testing.T) {
		shortTTL := NewSemanticCache(1*time.Millisecond, encoder)
		shortTTL.Set("expiring", route)

		time.Sleep(5 * time.Millisecond)

		result := shortTTL.Get("expiring")
		assert.Nil(t, result)
	})
}

func TestSemanticCache_SetWithEmbedding(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	route := &Route{Name: "test"}
	embedding := []float32{0.1, 0.2, 0.3}

	cache.SetWithEmbedding("query", route, embedding)

	assert.Equal(t, 1, cache.Size())
}

func TestSemanticCache_GetSemantic(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	route := &Route{Name: "test"}
	embedding := []float32{0.1, 0.2, 0.3}

	cache.SetWithEmbedding("query", route, embedding)

	t.Run("FoundMatch", func(t *testing.T) {
		queryEmbedding := []float32{0.1, 0.2, 0.3}
		result := cache.GetSemantic(queryEmbedding, 0.5)
		assert.NotNil(t, result)
		assert.Equal(t, "test", result.Name)
	})

	t.Run("NoMatch", func(t *testing.T) {
		queryEmbedding := []float32{1.0, 0.0, 0.0}
		result := cache.GetSemantic(queryEmbedding, 0.99)
		assert.Nil(t, result)
	})
}

func TestSemanticCache_Clear(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	cache.Set("key1", &Route{Name: "route1"})
	cache.Set("key2", &Route{Name: "route2"})

	assert.Equal(t, 2, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
}

func TestSemanticCache_Size(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	assert.Equal(t, 0, cache.Size())

	cache.Set("key1", &Route{Name: "route1"})
	assert.Equal(t, 1, cache.Size())

	cache.Set("key2", &Route{Name: "route2"})
	assert.Equal(t, 2, cache.Size())
}

func TestSemanticCache_GetStats(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(time.Minute, encoder)

	t.Run("EmptyCache", func(t *testing.T) {
		stats := cache.GetStats()
		assert.Equal(t, 0, stats.Size)
		assert.Equal(t, time.Minute, stats.TTL)
		assert.Nil(t, stats.OldestEntry)
		assert.Nil(t, stats.NewestEntry)
	})

	t.Run("WithEntries", func(t *testing.T) {
		cache.Set("key1", &Route{Name: "route1"})
		time.Sleep(10 * time.Millisecond)
		cache.Set("key2", &Route{Name: "route2"})

		stats := cache.GetStats()
		assert.Equal(t, 2, stats.Size)
		assert.NotNil(t, stats.OldestEntry)
		assert.NotNil(t, stats.NewestEntry)
		assert.True(t, stats.OldestEntry.Before(*stats.NewestEntry))
	})
}

func TestCacheStats(t *testing.T) {
	now := time.Now()
	stats := &CacheStats{
		Size:        100,
		OldestEntry: &now,
		NewestEntry: &now,
		TTL:         30 * time.Minute,
	}

	assert.Equal(t, 100, stats.Size)
	assert.NotNil(t, stats.OldestEntry)
	assert.Equal(t, 30*time.Minute, stats.TTL)
}

func TestSemanticCache_removeExpired(t *testing.T) {
	encoder := newMockEncoder()
	cache := NewSemanticCache(1*time.Millisecond, encoder)

	cache.Set("key1", &Route{Name: "route1"})
	cache.Set("key2", &Route{Name: "route2"})

	time.Sleep(5 * time.Millisecond)

	cache.removeExpired()

	assert.Equal(t, 0, cache.Size())
}

// Test RouterConfig struct
func TestRouterConfig(t *testing.T) {
	config := &RouterConfig{
		ScoreThreshold:    0.8,
		TopK:              3,
		EnableCache:       true,
		CacheTTL:          10 * time.Minute,
		FallbackRoute:     "default",
		AggregationMethod: AggregationMax,
	}

	assert.Equal(t, 0.8, config.ScoreThreshold)
	assert.Equal(t, 3, config.TopK)
	assert.True(t, config.EnableCache)
	assert.Equal(t, 10*time.Minute, config.CacheTTL)
	assert.Equal(t, "default", config.FallbackRoute)
	assert.Equal(t, AggregationMax, config.AggregationMethod)
}

// Test RouteHandler
func TestRouteHandler(t *testing.T) {
	handler := func(ctx context.Context, query string) (*RouteResult, error) {
		return &RouteResult{
			Content: "Handled: " + query,
			Model:   "gpt-4",
			Latency: 100 * time.Millisecond,
		}, nil
	}

	result, err := handler(context.Background(), "test query")
	require.NoError(t, err)
	assert.Equal(t, "Handled: test query", result.Content)
	assert.Equal(t, "gpt-4", result.Model)
}

// Test concurrent access
func TestRouter_Concurrent(t *testing.T) {
	encoder := newMockEncoder()
	config := &RouterConfig{
		ScoreThreshold: 0.0,
		EnableCache:    true,
		CacheTTL:       time.Minute,
	}
	router := NewRouter(encoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	// Concurrent reads and writes
	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		go func() {
			_, _ = router.Route(context.Background(), "hello")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func(idx int) {
			router.ListRoutes()
			done <- true
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
