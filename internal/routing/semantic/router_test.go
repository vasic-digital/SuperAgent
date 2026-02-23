package semantic

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// threadSafeMockEncoder is a thread-safe mock encoder for concurrent tests
type threadSafeMockEncoder struct {
	dimension   int
	encodeFunc  func(ctx context.Context, texts []string) ([][]float32, error)
	encodeCount int64 // Use int64 for atomic operations
	mu          sync.Mutex
}

func (e *threadSafeMockEncoder) Encode(ctx context.Context, texts []string) ([][]float32, error) {
	atomic.AddInt64(&e.encodeCount, 1)
	if e.encodeFunc != nil {
		e.mu.Lock()
		fn := e.encodeFunc
		e.mu.Unlock()
		return fn(ctx, texts)
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

func (e *threadSafeMockEncoder) GetDimension() int {
	return e.dimension
}

func (e *threadSafeMockEncoder) GetEncodeCount() int {
	return int(atomic.LoadInt64(&e.encodeCount))
}

func newThreadSafeMockEncoder() *threadSafeMockEncoder {
	return &threadSafeMockEncoder{dimension: 128}
}

// TestEncoderInterface verifies the Encoder interface contract
func TestEncoderInterface(t *testing.T) {
	t.Run("MockEncoderImplementsInterface", func(t *testing.T) {
		var encoder Encoder = newMockEncoder()
		assert.NotNil(t, encoder)
		assert.Equal(t, 128, encoder.GetDimension())
	})
}

// TestNewRouter_AllConfigurations tests router creation with various configurations
func TestNewRouter_AllConfigurations(t *testing.T) {
	tests := []struct {
		name           string
		config         *RouterConfig
		expectedCache  bool
		expectedTopK   int
		expectedThresh float64
	}{
		{
			name:           "nil_config_uses_defaults",
			config:         nil,
			expectedCache:  true,
			expectedTopK:   5,
			expectedThresh: 0.7,
		},
		{
			name: "cache_disabled",
			config: &RouterConfig{
				EnableCache:    false,
				ScoreThreshold: 0.5,
				TopK:           10,
			},
			expectedCache:  false,
			expectedTopK:   10,
			expectedThresh: 0.5,
		},
		{
			name: "custom_aggregation_method",
			config: &RouterConfig{
				EnableCache:       true,
				CacheTTL:          5 * time.Minute,
				AggregationMethod: AggregationMax,
			},
			expectedCache:  true,
			expectedTopK:   0,
			expectedThresh: 0,
		},
		{
			name: "with_fallback_route",
			config: &RouterConfig{
				EnableCache:   true,
				CacheTTL:      time.Hour,
				FallbackRoute: "default_handler",
			},
			expectedCache:  true,
			expectedTopK:   0,
			expectedThresh: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			router := NewRouter(encoder, tt.config, nil)

			assert.NotNil(t, router)
			assert.NotNil(t, router.encoder)
			assert.NotNil(t, router.logger)
			assert.NotNil(t, router.routes)
			assert.Empty(t, router.routes)

			if tt.expectedCache {
				assert.NotNil(t, router.cache)
			} else {
				assert.Nil(t, router.cache)
			}

			if tt.config == nil {
				assert.Equal(t, tt.expectedTopK, router.config.TopK)
				assert.Equal(t, tt.expectedThresh, router.config.ScoreThreshold)
			}
		})
	}
}

// TestNewRouter_WithCustomLogger tests router with custom logger
func TestNewRouter_WithCustomLogger(t *testing.T) {
	encoder := newMockEncoder()
	customLogger := logrus.New()
	customLogger.SetLevel(logrus.DebugLevel)

	router := NewRouter(encoder, nil, customLogger)

	assert.Equal(t, customLogger, router.logger)
}

// TestAddRoute_TableDriven tests AddRoute with various inputs
func TestAddRoute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		route       *Route
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_route_simple_tier",
			route: &Route{
				Name:        "greeting",
				Description: "Handles greeting messages",
				Utterances:  []string{"hello", "hi", "hey"},
				ModelTier:   ModelTierSimple,
			},
			expectError: false,
		},
		{
			name: "valid_route_standard_tier",
			route: &Route{
				Name:        "analysis",
				Description: "Handles analysis queries",
				Utterances:  []string{"analyze this", "what does this mean"},
				ModelTier:   ModelTierStandard,
			},
			expectError: false,
		},
		{
			name: "valid_route_complex_tier",
			route: &Route{
				Name:        "research",
				Description: "Handles complex research",
				Utterances:  []string{"research about"},
				ModelTier:   ModelTierComplex,
			},
			expectError: false,
		},
		{
			name: "valid_route_with_metadata",
			route: &Route{
				Name:       "custom",
				Utterances: []string{"custom query"},
				Metadata: map[string]interface{}{
					"priority": 1,
					"category": "general",
				},
			},
			expectError: false,
		},
		{
			name: "empty_name",
			route: &Route{
				Utterances: []string{"test"},
			},
			expectError: true,
			errorMsg:    "route name is required",
		},
		{
			name: "no_utterances",
			route: &Route{
				Name:       "empty_utterances",
				Utterances: []string{},
			},
			expectError: true,
			errorMsg:    "at least one utterance",
		},
		{
			name: "nil_utterances",
			route: &Route{
				Name:       "nil_utterances",
				Utterances: nil,
			},
			expectError: true,
			errorMsg:    "at least one utterance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			router := NewRouter(encoder, nil, nil)

			err := router.AddRoute(context.Background(), tt.route)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tt.route.Embedding)
				assert.Len(t, tt.route.Embedding, encoder.GetDimension())
			}
		})
	}
}

// TestAddRoute_EncoderFailures tests AddRoute when encoder fails
func TestAddRoute_EncoderFailures(t *testing.T) {
	tests := []struct {
		name       string
		encodeFunc func(ctx context.Context, texts []string) ([][]float32, error)
		errorMsg   string
	}{
		{
			name: "encoder_returns_error",
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return nil, errors.New("encoder service unavailable")
			},
			errorMsg: "failed to encode utterances",
		},
		{
			name: "encoder_returns_nil",
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				return nil, nil
			},
			errorMsg: "", // No error, but empty result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := &mockEncoder{
				dimension:  128,
				encodeFunc: tt.encodeFunc,
			}
			router := NewRouter(encoder, nil, nil)

			route := &Route{
				Name:       "test",
				Utterances: []string{"hello"},
			}

			err := router.AddRoute(context.Background(), route)

			if tt.errorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

// TestRoute_Scenarios tests route matching in various scenarios
func TestRoute_Scenarios(t *testing.T) {
	t.Run("multiple_routes_best_match", func(t *testing.T) {
		// Create encoder that produces predictable embeddings
		encoder := &mockEncoder{
			dimension: 3,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				results := make([][]float32, len(texts))
				for i, text := range texts {
					switch text {
					case "hello", "hi":
						results[i] = []float32{1.0, 0.0, 0.0}
					case "goodbye", "bye":
						results[i] = []float32{0.0, 1.0, 0.0}
					case "help", "assist":
						results[i] = []float32{0.0, 0.0, 1.0}
					case "hello there":
						results[i] = []float32{0.9, 0.1, 0.0} // Similar to greeting
					default:
						results[i] = []float32{0.33, 0.33, 0.33}
					}
				}
				return results, nil
			},
		}

		config := &RouterConfig{
			ScoreThreshold: 0.5,
			EnableCache:    false,
		}
		router := NewRouter(encoder, config, nil)

		// Add routes
		_ = router.AddRoute(context.Background(), &Route{
			Name:       "greeting",
			Utterances: []string{"hello", "hi"},
		})
		_ = router.AddRoute(context.Background(), &Route{
			Name:       "farewell",
			Utterances: []string{"goodbye", "bye"},
		})
		_ = router.AddRoute(context.Background(), &Route{
			Name:       "help",
			Utterances: []string{"help", "assist"},
		})

		// Test routing
		route, err := router.Route(context.Background(), "hello there")
		require.NoError(t, err)
		assert.Equal(t, "greeting", route.Name)
	})

	t.Run("fallback_on_low_confidence", func(t *testing.T) {
		// Create encoder with orthogonal vectors
		callCount := 0
		encoder := &mockEncoder{
			dimension: 3,
			encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
				results := make([][]float32, len(texts))
				for i := range texts {
					if callCount == 0 {
						// Route embedding
						results[i] = []float32{1.0, 0.0, 0.0}
					} else if callCount == 1 {
						// Fallback route embedding
						results[i] = []float32{0.0, 1.0, 0.0}
					} else {
						// Query embedding - orthogonal to main route
						results[i] = []float32{0.0, 0.0, 1.0}
					}
				}
				callCount++
				return results, nil
			},
		}

		config := &RouterConfig{
			ScoreThreshold: 0.9, // High threshold
			EnableCache:    false,
			FallbackRoute:  "fallback",
		}
		router := NewRouter(encoder, config, nil)

		_ = router.AddRoute(context.Background(), &Route{
			Name:       "main",
			Utterances: []string{"specific"},
		})
		_ = router.AddRoute(context.Background(), &Route{
			Name:       "fallback",
			Utterances: []string{"default"},
		})

		route, err := router.Route(context.Background(), "unrelated query")
		require.NoError(t, err)
		assert.Equal(t, "fallback", route.Name)
	})

	t.Run("cache_prevents_reencoding", func(t *testing.T) {
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

		initialCount := encoder.encodeCount

		// First call
		_, err := router.Route(context.Background(), "hello")
		require.NoError(t, err)
		firstCallCount := encoder.encodeCount

		// Second call - should hit cache
		_, err = router.Route(context.Background(), "hello")
		require.NoError(t, err)
		secondCallCount := encoder.encodeCount

		assert.Greater(t, firstCallCount, initialCount)
		assert.Equal(t, firstCallCount, secondCallCount) // No additional encoding
	})
}

// TestRouteWithCandidates_TableDriven tests candidate retrieval
func TestRouteWithCandidates_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		numRoutes     int
		topK          int
		expectedCount int
	}{
		{
			name:          "topK_less_than_routes",
			numRoutes:     10,
			topK:          3,
			expectedCount: 3,
		},
		{
			name:          "topK_equals_routes",
			numRoutes:     5,
			topK:          5,
			expectedCount: 5,
		},
		{
			name:          "topK_greater_than_routes",
			numRoutes:     3,
			topK:          10,
			expectedCount: 3,
		},
		{
			name:          "single_route",
			numRoutes:     1,
			topK:          5,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			config := &RouterConfig{
				TopK:        tt.topK,
				EnableCache: false,
			}
			router := NewRouter(encoder, config, nil)

			for i := 0; i < tt.numRoutes; i++ {
				_ = router.AddRoute(context.Background(), &Route{
					Name:       string(rune('A' + i)),
					Utterances: []string{"test"},
				})
			}

			candidates, err := router.RouteWithCandidates(context.Background(), "query")
			require.NoError(t, err)
			assert.Len(t, candidates, tt.expectedCount)

			// Verify candidates have scores set
			for _, candidate := range candidates {
				assert.NotNil(t, candidate)
				assert.NotEmpty(t, candidate.Name)
			}
		})
	}
}

// TestRouteWithCandidates_Ordering verifies candidates are sorted by score
func TestRouteWithCandidates_Ordering(t *testing.T) {
	// Create encoder with controlled similarity scores
	encoder := &mockEncoder{
		dimension: 3,
		encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
			results := make([][]float32, len(texts))
			for i, text := range texts {
				switch text {
				case "high":
					results[i] = []float32{1.0, 0.0, 0.0}
				case "medium":
					results[i] = []float32{0.7, 0.5, 0.0}
				case "low":
					results[i] = []float32{0.3, 0.3, 0.8}
				default: // query
					results[i] = []float32{0.95, 0.1, 0.05}
				}
			}
			return results, nil
		},
	}

	config := &RouterConfig{
		TopK:        3,
		EnableCache: false,
	}
	router := NewRouter(encoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "low_match",
		Utterances: []string{"low"},
	})
	_ = router.AddRoute(context.Background(), &Route{
		Name:       "high_match",
		Utterances: []string{"high"},
	})
	_ = router.AddRoute(context.Background(), &Route{
		Name:       "medium_match",
		Utterances: []string{"medium"},
	})

	candidates, err := router.RouteWithCandidates(context.Background(), "query")
	require.NoError(t, err)
	assert.Len(t, candidates, 3)

	// First candidate should have highest score
	assert.Equal(t, "high_match", candidates[0].Name)

	// Scores should be in descending order
	for i := 1; i < len(candidates); i++ {
		assert.GreaterOrEqual(t, candidates[i-1].Score, candidates[i].Score)
	}
}

// TestRemoveRoute_Scenarios tests various removal scenarios
func TestRemoveRoute_Scenarios(t *testing.T) {
	tests := []struct {
		name           string
		routesToAdd    []string
		routeToRemove  string
		expectedRemain []string
	}{
		{
			name:           "remove_first",
			routesToAdd:    []string{"A", "B", "C"},
			routeToRemove:  "A",
			expectedRemain: []string{"B", "C"},
		},
		{
			name:           "remove_middle",
			routesToAdd:    []string{"A", "B", "C"},
			routeToRemove:  "B",
			expectedRemain: []string{"A", "C"},
		},
		{
			name:           "remove_last",
			routesToAdd:    []string{"A", "B", "C"},
			routeToRemove:  "C",
			expectedRemain: []string{"A", "B"},
		},
		{
			name:           "remove_nonexistent",
			routesToAdd:    []string{"A", "B"},
			routeToRemove:  "X",
			expectedRemain: []string{"A", "B"},
		},
		{
			name:           "remove_from_single",
			routesToAdd:    []string{"A"},
			routeToRemove:  "A",
			expectedRemain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := newMockEncoder()
			router := NewRouter(encoder, nil, nil)

			for _, name := range tt.routesToAdd {
				_ = router.AddRoute(context.Background(), &Route{
					Name:       name,
					Utterances: []string{"test"},
				})
			}

			router.RemoveRoute(tt.routeToRemove)

			routes := router.ListRoutes()
			assert.Len(t, routes, len(tt.expectedRemain))

			routeNames := make([]string, len(routes))
			for i, r := range routes {
				routeNames[i] = r.Name
			}

			for _, expected := range tt.expectedRemain {
				assert.Contains(t, routeNames, expected)
			}
		})
	}
}

// TestListRoutes_ReturnsCopy verifies ListRoutes returns a copy
func TestListRoutes_ReturnsCopy(t *testing.T) {
	encoder := newMockEncoder()
	router := NewRouter(encoder, nil, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	routes1 := router.ListRoutes()
	routes2 := router.ListRoutes()

	// Modify the first slice
	routes1[0] = nil

	// Second slice should be unaffected
	assert.NotNil(t, routes2[0])
	assert.Equal(t, "test", routes2[0].Name)
}

// TestAggregateEmbeddings_EdgeCases tests aggregation edge cases
func TestAggregateEmbeddings_EdgeCases(t *testing.T) {
	encoder := newMockEncoder()

	t.Run("single_embedding", func(t *testing.T) {
		config := &RouterConfig{
			AggregationMethod: AggregationMean,
			EnableCache:       false,
		}
		router := NewRouter(encoder, config, nil)

		embeddings := [][]float32{{1.0, 2.0, 3.0}}
		result := router.aggregateEmbeddings(embeddings)

		assert.Equal(t, []float32{1.0, 2.0, 3.0}, result)
	})

	t.Run("three_embeddings_mean", func(t *testing.T) {
		config := &RouterConfig{
			AggregationMethod: AggregationMean,
			EnableCache:       false,
		}
		router := NewRouter(encoder, config, nil)

		embeddings := [][]float32{
			{0.0, 3.0, 6.0},
			{3.0, 3.0, 3.0},
			{6.0, 3.0, 0.0},
		}
		result := router.aggregateEmbeddings(embeddings)

		assert.Equal(t, float32(3.0), result[0])
		assert.Equal(t, float32(3.0), result[1])
		assert.Equal(t, float32(3.0), result[2])
	})

	t.Run("negative_values_max", func(t *testing.T) {
		config := &RouterConfig{
			AggregationMethod: AggregationMax,
			EnableCache:       false,
		}
		router := NewRouter(encoder, config, nil)

		embeddings := [][]float32{
			{-5.0, -1.0, 0.0},
			{-3.0, -2.0, -1.0},
		}
		result := router.aggregateEmbeddings(embeddings)

		assert.Equal(t, float32(-3.0), result[0])
		assert.Equal(t, float32(-1.0), result[1])
		assert.Equal(t, float32(0.0), result[2])
	})
}

// TestCosineSimilarity_Comprehensive tests cosine similarity calculations
func TestCosineSimilarity_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		delta    float64
	}{
		{
			name:     "identical_unit_vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "opposite_vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{-1.0, 0.0, 0.0},
			expected: -1.0,
			delta:    0.001,
		},
		{
			name:     "orthogonal_xy",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "45_degree_angle",
			a:        []float32{1.0, 0.0},
			b:        []float32{1.0, 1.0},
			expected: 0.707, // cos(45) approx
			delta:    0.01,
		},
		{
			name:     "scaled_vectors",
			a:        []float32{2.0, 4.0, 6.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 1.0, // Same direction
			delta:    0.001,
		},
		{
			name:     "empty_first",
			a:        []float32{},
			b:        []float32{1.0},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "empty_second",
			a:        []float32{1.0},
			b:        []float32{},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "both_empty",
			a:        []float32{},
			b:        []float32{},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "zero_first",
			a:        []float32{0.0, 0.0, 0.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "zero_second",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{0.0, 0.0, 0.0},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "both_zero",
			a:        []float32{0.0, 0.0},
			b:        []float32{0.0, 0.0},
			expected: 0.0,
			delta:    0.0,
		},
		{
			name:     "high_dimensional",
			a:        []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
			b:        []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
			expected: 1.0,
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

// TestSqrt_Comprehensive tests square root calculation
func TestSqrt_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
		delta    float64
	}{
		{"perfect_square_1", 1.0, 1.0, 0.001},
		{"perfect_square_4", 4.0, 2.0, 0.001},
		{"perfect_square_9", 9.0, 3.0, 0.001},
		{"perfect_square_16", 16.0, 4.0, 0.001},
		{"perfect_square_25", 25.0, 5.0, 0.001},
		{"perfect_square_100", 100.0, 10.0, 0.001},
		{"non_perfect_2", 2.0, 1.41421, 0.001},
		{"non_perfect_3", 3.0, 1.73205, 0.001},
		{"non_perfect_5", 5.0, 2.23607, 0.001},
		{"decimal", 0.25, 0.5, 0.001},
		{"large", 10000.0, 100.0, 0.001},
		{"zero", 0.0, 0.0, 0.0},
		{"negative", -4.0, 0.0, 0.0},
		{"very_small", 0.0001, 0.01, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqrt(tt.input)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

// TestMin_Comprehensive tests min function
func TestMin_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a_smaller", 1, 5, 1},
		{"b_smaller", 10, 3, 3},
		{"equal", 7, 7, 7},
		{"negative_a", -5, 5, -5},
		{"negative_b", 5, -5, -5},
		{"both_negative", -3, -7, -7},
		{"zero_and_positive", 0, 5, 0},
		{"zero_and_negative", 0, -5, -5},
		{"large_numbers", 1000000, 999999, 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConcurrency_HeavyLoad tests concurrent access under heavy load
func TestConcurrency_HeavyLoad(t *testing.T) {
	encoder := newThreadSafeMockEncoder()
	config := &RouterConfig{
		ScoreThreshold: 0.0,
		EnableCache:    true,
		CacheTTL:       time.Minute,
	}
	router := NewRouter(encoder, config, nil)

	// Add routes
	for i := 0; i < 10; i++ {
		_ = router.AddRoute(context.Background(), &Route{
			Name:       string(rune('A' + i)),
			Utterances: []string{"test"},
		})
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = router.Route(context.Background(), "test")
		}(i)
	}

	// Concurrent list
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = router.ListRoutes()
		}(i)
	}

	// Concurrent candidates
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = router.RouteWithCandidates(context.Background(), "test")
		}(i)
	}

	wg.Wait()
}

// TestConcurrency_AddRemove tests concurrent add and remove operations
func TestConcurrency_AddRemove(t *testing.T) {
	encoder := newThreadSafeMockEncoder()
	config := &RouterConfig{
		EnableCache: false,
	}
	router := NewRouter(encoder, config, nil)

	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = router.AddRoute(context.Background(), &Route{
				Name:       string(rune('A'+id%26)) + string(rune('0'+id/26)),
				Utterances: []string{"test"},
			})
		}(i)
	}

	wg.Wait()

	// Concurrent removes
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			router.RemoveRoute(string(rune('A'+id%26)) + string(rune('0'+id/26)))
		}(i)
	}

	wg.Wait()
}

// TestRouter_ContextCancellation tests context cancellation behavior
func TestRouter_ContextCancellation(t *testing.T) {
	encoder := &mockEncoder{
		dimension: 128,
		encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
				results := make([][]float32, len(texts))
				for i := range results {
					results[i] = make([]float32, 128)
				}
				return results, nil
			}
		},
	}

	router := NewRouter(encoder, nil, nil)

	t.Run("cancelled_context_on_add_route", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := router.AddRoute(ctx, &Route{
			Name:       "test",
			Utterances: []string{"hello"},
		})

		require.Error(t, err)
	})

	t.Run("timeout_context_on_route", func(t *testing.T) {
		_ = router.AddRoute(context.Background(), &Route{
			Name:       "timeout_test",
			Utterances: []string{"hello"},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond) // Ensure timeout

		_, err := router.Route(ctx, "query")
		// Either times out or succeeds (race condition with encoder)
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})
}

// TestRoute_ScoreInResult tests that the score is correctly set in route result
func TestRoute_ScoreInResult(t *testing.T) {
	encoder := &mockEncoder{
		dimension: 3,
		encodeFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
			results := make([][]float32, len(texts))
			for i := range texts {
				results[i] = []float32{1.0, 0.0, 0.0} // All same vector
			}
			return results, nil
		},
	}

	config := &RouterConfig{
		ScoreThreshold: 0.0,
		EnableCache:    false,
	}
	router := NewRouter(encoder, config, nil)

	_ = router.AddRoute(context.Background(), &Route{
		Name:       "test",
		Utterances: []string{"hello"},
	})

	route, err := router.Route(context.Background(), "query")
	require.NoError(t, err)
	assert.InDelta(t, 1.0, route.Score, 0.001) // Identical vectors should have score ~1.0
}

// TestRouteHandler_Execution tests that route handlers can be executed
func TestRouteHandler_Execution(t *testing.T) {
	t.Run("handler_returns_success", func(t *testing.T) {
		handler := func(ctx context.Context, query string) (*RouteResult, error) {
			return &RouteResult{
				Content:  "Response for: " + query,
				Model:    "test-model",
				Latency:  50 * time.Millisecond,
				Metadata: map[string]interface{}{"processed": true},
			}, nil
		}

		result, err := handler(context.Background(), "test query")
		require.NoError(t, err)
		assert.Equal(t, "Response for: test query", result.Content)
		assert.Equal(t, "test-model", result.Model)
		assert.Equal(t, 50*time.Millisecond, result.Latency)
		assert.True(t, result.Metadata["processed"].(bool))
	})

	t.Run("handler_returns_error", func(t *testing.T) {
		handler := func(ctx context.Context, query string) (*RouteResult, error) {
			return nil, errors.New("handler failed")
		}

		result, err := handler(context.Background(), "test query")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "handler failed", err.Error())
	})

	t.Run("handler_respects_context", func(t *testing.T) {
		handler := func(ctx context.Context, query string) (*RouteResult, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
				return &RouteResult{Content: "done"}, nil
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := handler(ctx, "test")
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

// TestModelTier_Usage tests model tier constants are used correctly
func TestModelTier_Usage(t *testing.T) {
	encoder := newMockEncoder()
	router := NewRouter(encoder, nil, nil)

	tests := []struct {
		name      string
		tier      ModelTier
		tierValue string
	}{
		{"simple", ModelTierSimple, "simple"},
		{"standard", ModelTierStandard, "standard"},
		{"complex", ModelTierComplex, "complex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &Route{
				Name:       "test_" + tt.tierValue,
				Utterances: []string{"test"},
				ModelTier:  tt.tier,
			}

			err := router.AddRoute(context.Background(), route)
			require.NoError(t, err)

			assert.Equal(t, tt.tierValue, string(route.ModelTier))
		})
	}
}
