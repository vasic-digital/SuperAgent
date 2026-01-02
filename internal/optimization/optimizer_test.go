package optimization

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superagent/superagent/internal/optimization/outlines"
	"github.com/superagent/superagent/internal/optimization/streaming"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.True(t, config.SemanticCache.Enabled)
	assert.Equal(t, 0.85, config.SemanticCache.SimilarityThreshold)
	assert.Equal(t, 10000, config.SemanticCache.MaxEntries)
	assert.Equal(t, 24*time.Hour, config.SemanticCache.TTL)

	assert.True(t, config.StructuredOutput.Enabled)
	assert.True(t, config.StructuredOutput.StrictMode)
	assert.Equal(t, 3, config.StructuredOutput.MaxRetries)

	assert.True(t, config.Streaming.Enabled)
	assert.Equal(t, "word", config.Streaming.BufferType)

	assert.Equal(t, "skip", config.Fallback.OnServiceUnavailable)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
		check  func(*testing.T, *Config)
	}{
		{
			name: "fix invalid similarity threshold",
			modify: func(c *Config) {
				c.SemanticCache.SimilarityThreshold = 2.0
			},
			check: func(t *testing.T, c *Config) {
				assert.Equal(t, 0.85, c.SemanticCache.SimilarityThreshold)
			},
		},
		{
			name: "fix negative similarity threshold",
			modify: func(c *Config) {
				c.SemanticCache.SimilarityThreshold = -0.5
			},
			check: func(t *testing.T, c *Config) {
				assert.Equal(t, 0.85, c.SemanticCache.SimilarityThreshold)
			},
		},
		{
			name: "fix zero max entries",
			modify: func(c *Config) {
				c.SemanticCache.MaxEntries = 0
			},
			check: func(t *testing.T, c *Config) {
				assert.Equal(t, 10000, c.SemanticCache.MaxEntries)
			},
		},
		{
			name: "fix zero max retries",
			modify: func(c *Config) {
				c.StructuredOutput.MaxRetries = 0
			},
			check: func(t *testing.T, c *Config) {
				assert.Equal(t, 3, c.StructuredOutput.MaxRetries)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.modify(config)
			err := config.Validate()
			require.NoError(t, err)
			tt.check(t, config)
		})
	}
}

func TestNewService(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		svc, err := NewService(nil)
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.NotNil(t, svc.config)
		assert.True(t, svc.config.Enabled)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			Enabled: true,
			SemanticCache: SemanticCacheConfig{
				Enabled:             true,
				SimilarityThreshold: 0.9,
				MaxEntries:          5000,
				TTL:                 12 * time.Hour,
			},
			Streaming: StreamingConfig{
				Enabled:    true,
				BufferType: "sentence",
			},
		}

		svc, err := NewService(config)
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, 0.9, svc.config.SemanticCache.SimilarityThreshold)
	})

	t.Run("initializes semantic cache when enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)
		assert.NotNil(t, svc.semanticCache)
	})

	t.Run("does not initialize semantic cache when disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)
		assert.Nil(t, svc.semanticCache)
	})

	t.Run("initializes enhanced streamer when enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.Streaming.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)
		assert.NotNil(t, svc.enhancedStreamer)
	})
}

func TestServiceOptimizeRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("returns original prompt when no optimizations apply", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = false
		config.LlamaIndex.Enabled = false
		config.LangChain.Enabled = false
		config.SGLang.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		result, err := svc.OptimizeRequest(ctx, "Hello world", nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello world", result.OriginalPrompt)
		assert.Equal(t, "Hello world", result.OptimizedPrompt)
		assert.False(t, result.CacheHit)
	})

	t.Run("returns cache hit when found", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = true
		// Disable external services
		config.LlamaIndex.Enabled = false
		config.LangChain.Enabled = false
		config.SGLang.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		// First, add an entry to the cache
		embedding := []float64{1.0, 0.0, 0.0}
		_, err = svc.semanticCache.Set(ctx, "test query", "cached response", embedding, nil)
		require.NoError(t, err)

		// Now request with the same embedding
		result, err := svc.OptimizeRequest(ctx, "test query", embedding)
		require.NoError(t, err)
		assert.True(t, result.CacheHit)
		assert.Equal(t, "cached response", result.CachedResponse)
	})

	t.Run("increments cache miss counter", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = true
		config.LlamaIndex.Enabled = false
		config.LangChain.Enabled = false
		config.SGLang.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		initialMisses := svc.cacheMisses

		// Request with an embedding that won't match anything
		embedding := []float64{0.5, 0.5, 0.0}
		_, err = svc.OptimizeRequest(ctx, "new query", embedding)
		require.NoError(t, err)

		assert.Equal(t, initialMisses+1, svc.cacheMisses)
	})
}

func TestServiceOptimizeResponse(t *testing.T) {
	ctx := context.Background()

	t.Run("caches response when semantic cache enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		embedding := []float64{1.0, 0.0, 0.0}
		result, err := svc.OptimizeResponse(ctx, "test response", embedding, "test query", nil)
		require.NoError(t, err)
		assert.True(t, result.Cached)
		assert.Equal(t, "test response", result.Content)
	})

	t.Run("validates structured output when schema provided", func(t *testing.T) {
		config := DefaultConfig()
		config.StructuredOutput.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		schema := outlines.StringSchema()
		result, err := svc.OptimizeResponse(ctx, `"hello"`, nil, "test", schema)
		require.NoError(t, err)
		assert.NotNil(t, result.ValidationResult)
		assert.True(t, result.ValidationResult.Valid)
	})

	t.Run("does not validate when schema is nil", func(t *testing.T) {
		config := DefaultConfig()
		config.StructuredOutput.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		result, err := svc.OptimizeResponse(ctx, "plain text", nil, "test", nil)
		require.NoError(t, err)
		assert.Nil(t, result.ValidationResult)
	})
}

func TestServiceStreamEnhanced(t *testing.T) {
	ctx := context.Background()

	t.Run("returns enhanced stream when streaming enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.Streaming.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		in := make(chan *streaming.StreamChunk, 3)
		in <- &streaming.StreamChunk{Content: "Hello ", Index: 0}
		in <- &streaming.StreamChunk{Content: "world", Index: 1}
		in <- &streaming.StreamChunk{Content: "", Index: 2, Done: true}
		close(in)

		out, getResult := svc.StreamEnhanced(ctx, in, nil)

		// Consume output
		for range out {
		}

		result := getResult()
		assert.NotNil(t, result)
		assert.Contains(t, result.FullContent, "Hello")
	})

	t.Run("returns passthrough when streaming disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.Streaming.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		in := make(chan *streaming.StreamChunk, 2)
		in <- &streaming.StreamChunk{Content: "Test", Index: 0}
		in <- &streaming.StreamChunk{Content: "", Index: 1, Done: true}
		close(in)

		out, getResult := svc.StreamEnhanced(ctx, in, nil)

		// Consume output
		for range out {
		}

		result := getResult()
		assert.NotNil(t, result)
	})
}

func TestServiceGenerateStructured(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when structured output disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.StructuredOutput.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		schema := outlines.StringSchema()
		generator := func(prompt string) (string, error) {
			return `"test"`, nil
		}

		_, err = svc.GenerateStructured(ctx, "test", schema, generator)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("generates and validates structured output", func(t *testing.T) {
		config := DefaultConfig()
		config.StructuredOutput.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		schema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
			"name": outlines.StringSchema(),
			"age":  outlines.IntegerSchema(),
		}, "name", "age")

		generator := func(prompt string) (string, error) {
			return `{"name": "John", "age": 30}`, nil
		}

		result, err := svc.GenerateStructured(ctx, "Generate a person", schema, generator)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Contains(t, result.Content, "John")
	})

	t.Run("returns validation errors for invalid output", func(t *testing.T) {
		config := DefaultConfig()
		config.StructuredOutput.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		schema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
			"name": outlines.StringSchema(),
		}, "name")

		generator := func(prompt string) (string, error) {
			return `{"invalid": "missing name"}`, nil
		}

		result, err := svc.GenerateStructured(ctx, "Generate a person", schema, generator)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})
}

func TestServiceGetCacheStats(t *testing.T) {
	t.Run("returns stats when cache enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = true

		svc, err := NewService(config)
		require.NoError(t, err)

		stats := svc.GetCacheStats()
		assert.True(t, stats["enabled"].(bool))
		assert.Equal(t, int64(0), stats["hits"])
		assert.Equal(t, int64(0), stats["misses"])
	})

	t.Run("returns disabled when cache not enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.SemanticCache.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		stats := svc.GetCacheStats()
		assert.False(t, stats["enabled"].(bool))
	})
}

func TestServiceGetServiceStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("returns empty map when no services configured", func(t *testing.T) {
		config := DefaultConfig()
		config.SGLang.Enabled = false
		config.LlamaIndex.Enabled = false
		config.LangChain.Enabled = false
		config.Guidance.Enabled = false
		config.LMQL.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		status := svc.GetServiceStatus(ctx)
		assert.Empty(t, status)
	})
}

func TestServiceIsServiceAvailable(t *testing.T) {
	t.Run("returns true for unknown service", func(t *testing.T) {
		config := DefaultConfig()
		svc, err := NewService(config)
		require.NoError(t, err)

		// Unknown services are assumed available
		assert.True(t, svc.isServiceAvailable("unknown"))
	})

	t.Run("returns false when service marked unavailable", func(t *testing.T) {
		config := DefaultConfig()
		svc, err := NewService(config)
		require.NoError(t, err)

		svc.markServiceUnavailable("test-service")

		assert.False(t, svc.isServiceAvailable("test-service"))
	})
}

func TestIsComplexTask(t *testing.T) {
	tests := []struct {
		prompt   string
		expected bool
	}{
		{"Hello", false},
		{"What is 2+2?", false},
		// Prompts must be > 100 chars for indicator matching
		{"Please implement a function that calculates fibonacci numbers. I need this to be efficient and handle large numbers step by step", true},
		{"Create a new user authentication system with login and registration functionality. This should include password hashing and session management", true},
		{"Build a REST API with CRUD operations for a product catalog. The API should support filtering, pagination, and sorting capabilities", true},
		{string(make([]byte, 600)), true}, // Long prompt
	}

	for _, tt := range tests {
		t.Run(tt.prompt[:min(30, len(tt.prompt))], func(t *testing.T) {
			result := isComplexTask(tt.prompt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceExternalServicesUnavailable(t *testing.T) {
	ctx := context.Background()

	t.Run("DecomposeTask returns error when langchain unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.LangChain.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.DecomposeTask(ctx, "test task")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("RunReActAgent returns error when langchain unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.LangChain.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.RunReActAgent(ctx, "test goal", []string{"search"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("QueryDocuments returns error when llamaindex unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.LlamaIndex.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.QueryDocuments(ctx, "test query", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("GenerateConstrained returns error when lmql unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.LMQL.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.GenerateConstrained(ctx, "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("SelectFromOptions returns error when guidance unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.Guidance.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.SelectFromOptions(ctx, "choose", []string{"a", "b"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("CreateSession returns error when sglang unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.SGLang.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		err = svc.CreateSession(ctx, "session1", "system prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("ContinueSession returns error when sglang unavailable", func(t *testing.T) {
		config := DefaultConfig()
		config.SGLang.Enabled = false

		svc, err := NewService(config)
		require.NoError(t, err)

		_, err = svc.ContinueSession(ctx, "session1", "hello")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})
}

func TestServiceMarkServiceUnavailable(t *testing.T) {
	config := DefaultConfig()
	svc, err := NewService(config)
	require.NoError(t, err)

	svc.markServiceUnavailable("test-service")

	svc.mu.RLock()
	status := svc.serviceStatus["test-service"]
	until := svc.unavailableUntil["test-service"]
	svc.mu.RUnlock()

	assert.False(t, status)
	assert.True(t, until.After(time.Now()))
}

// Benchmark tests
func BenchmarkSemanticCacheLookup(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()
	config.SemanticCache.Enabled = true
	config.LlamaIndex.Enabled = false
	config.LangChain.Enabled = false
	config.SGLang.Enabled = false

	svc, _ := NewService(config)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		embedding := make([]float64, 128)
		embedding[i%128] = 1.0
		svc.semanticCache.Set(ctx, "query", "response", embedding, nil)
	}

	embedding := []float64{1.0}
	for i := 1; i < 128; i++ {
		embedding = append(embedding, 0.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.OptimizeRequest(ctx, "test query", embedding)
	}
}

func BenchmarkStreamEnhanced(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()
	config.Streaming.Enabled = true

	svc, _ := NewService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in := make(chan *streaming.StreamChunk, 10)
		go func() {
			for j := 0; j < 10; j++ {
				in <- &streaming.StreamChunk{Content: "word ", Index: j}
			}
			in <- &streaming.StreamChunk{Done: true, Index: 10}
			close(in)
		}()

		out, _ := svc.StreamEnhanced(ctx, in, nil)
		for range out {
		}
	}
}
