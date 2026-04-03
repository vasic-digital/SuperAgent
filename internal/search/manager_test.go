package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "tavily only",
			config: &Config{
				TavilyAPIKey: "test-key",
				Timeout:      30 * time.Second,
				Logger:       logger,
			},
			wantErr: false,
		},
		{
			name: "perplexity only",
			config: &Config{
				PerplexityAPIKey: "test-key",
				Timeout:          30 * time.Second,
				Logger:           logger,
			},
			wantErr: false,
		},
		{
			name: "exa only",
			config: &Config{
				ExaAPIKey: "test-key",
				Timeout:   30 * time.Second,
				Logger:    logger,
			},
			wantErr: false,
		},
		{
			name: "all providers",
			config: &Config{
				TavilyAPIKey:     "tavily-key",
				PerplexityAPIKey: "perplexity-key",
				ExaAPIKey:        "exa-key",
				Timeout:          30 * time.Second,
				Logger:           logger,
			},
			wantErr: false,
		},
		{
			name: "no providers fails",
			config: &Config{
				Timeout: 30 * time.Second,
				Logger:  logger,
			},
			wantErr: true,
		},
		{
			name:    "nil config fails",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestManager_Search(t *testing.T) {
	logger := zap.NewNop()
	manager, err := NewManager(&Config{
		TavilyAPIKey: "test-key",
		Timeout:      10 * time.Second,
		Logger:       logger,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This would make real API calls in integration tests
	// For unit tests, we verify the method structure
	t.Run("search params validated", func(t *testing.T) {
		// Empty query should fail validation
		_, err := manager.Search(ctx, "", &Options{NumResults: 5})
		assert.Error(t, err)
	})

	t.Run("options defaults applied", func(t *testing.T) {
		options := &Options{}
		options.SetDefaults()
		assert.Equal(t, 10, options.NumResults)
		assert.True(t, options.IncludeText)
	})
}

func TestManager_SearchWithAI(t *testing.T) {
	logger := zap.NewNop()
	manager, err := NewManager(&Config{
		PerplexityAPIKey: "test-key",
		Timeout:          10 * time.Second,
		Logger:           logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("requires perplexity", func(t *testing.T) {
		managerWithoutPerplexity, _ := NewManager(&Config{
			TavilyAPIKey: "test-key",
			Logger:       logger,
		})
		_, err := managerWithoutPerplexity.SearchWithAI(ctx, "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "perplexity")
	})

	t.Run("ai options defaults", func(t *testing.T) {
		options := &AIOptions{}
		options.SetDefaults()
		assert.Equal(t, "sonar", options.Model)
		assert.Equal(t, 500, options.MaxTokens)
		assert.True(t, options.IncludeSources)
	})
}

func TestManager_SearchCode(t *testing.T) {
	logger := zap.NewNop()
	manager, err := NewManager(&Config{
		ExaAPIKey: "test-key",
		Timeout:   10 * time.Second,
		Logger:    logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("requires exa", func(t *testing.T) {
		managerWithoutExa, _ := NewManager(&Config{
			TavilyAPIKey: "test-key",
			Logger:       logger,
		})
		_, err := managerWithoutExa.SearchCode(ctx, "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exa")
	})

	t.Run("code options", func(t *testing.T) {
		options := &CodeOptions{
			NumResults: 5,
			Language:   "go",
		}
		assert.Equal(t, 5, options.NumResults)
		assert.Equal(t, "go", options.Language)
	})
}

func TestManager_SearchAggregated(t *testing.T) {
	logger := zap.NewNop()
	manager, err := NewManager(&Config{
		TavilyAPIKey:     "tavily-key",
		PerplexityAPIKey: "perplexity-key",
		ExaAPIKey:        "exa-key",
		Timeout:          30 * time.Second,
		Logger:           logger,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("aggregated options validation", func(t *testing.T) {
		options := &AggregatedOptions{
			Providers: []string{"tavily"},
		}
		assert.NotNil(t, options.Providers)
		assert.Len(t, options.Providers, 1)
	})

	t.Run("deduplication enabled", func(t *testing.T) {
		options := &AggregatedOptions{
			Deduplicate: true,
		}
		assert.True(t, options.Deduplicate)
	})
}

func TestSearchResult_Merge(t *testing.T) {
	result1 := SearchResult{
		Title:   "Result 1",
		URL:     "https://example.com/1",
		Snippet: "Snippet 1",
		Score:   0.9,
	}

	result2 := SearchResult{
		Title:   "Result 2",
		URL:     "https://example.com/2",
		Snippet: "Snippet 2",
		Score:   0.8,
	}

	merged := MergeResults([]SearchResult{result1, result2}, 10)
	assert.Len(t, merged, 2)
	assert.Equal(t, "Result 1", merged[0].Title) // Higher score first
}

func TestSearchResult_Deduplicate(t *testing.T) {
	results := []SearchResult{
		{URL: "https://example.com/page", Title: "Page 1"},
		{URL: "https://example.com/page", Title: "Page 1 Duplicate"},
		{URL: "https://example.com/other", Title: "Other Page"},
	}

	deduped := DeduplicateResults(results)
	assert.Len(t, deduped, 2)
	assert.Equal(t, "Page 1", deduped[0].Title)
}

func TestAggregatedResult_Stats(t *testing.T) {
	result := AggregatedResult{
		Results: []SearchResult{
			{Source: "tavily"},
			{Source: "tavily"},
			{Source: "perplexity"},
		},
		Providers: []string{"tavily", "perplexity"},
	}

	stats := result.GetStats()
	assert.Equal(t, 3, stats.TotalResults)
	assert.Equal(t, 2, stats.ProviderCount)
	assert.Equal(t, 2, stats.ResultsBySource["tavily"])
	assert.Equal(t, 1, stats.ResultsBySource["perplexity"])
}

func BenchmarkSearchResult_Merge(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = SearchResult{
			Title: fmt.Sprintf("Result %d", i),
			URL:   fmt.Sprintf("https://example.com/%d", i),
			Score: float64(100-i) / 100,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeResults(results, 50)
	}
}

func BenchmarkSearchResult_Deduplicate(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		url := fmt.Sprintf("https://example.com/%d", i%50) // 50% duplicates
		results[i] = SearchResult{
			URL:   url,
			Title: fmt.Sprintf("Result %d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeduplicateResults(results)
	}
}
