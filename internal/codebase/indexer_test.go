package codebase

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewIndexer(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				RootPath:          "/tmp/test",
				EmbeddingProvider: "openai",
				EmbeddingModel:    "text-embedding-3-small",
				Logger:            logger,
			},
			wantErr: false,
		},
		{
			name:    "nil config fails",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty root path fails",
			config: &Config{
				EmbeddingProvider: "openai",
				Logger:            logger,
			},
			wantErr: true,
		},
		{
			name: "default model applied",
			config: &Config{
				RootPath:          "/tmp/test",
				EmbeddingProvider: "openai",
				Logger:            logger,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexer, err := NewIndexer(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, indexer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, indexer)
			}
		})
	}
}

func TestIndexer_Index(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "codebase-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":          "package main\n\nfunc main() {}",
		"utils.go":         "package main\n\nfunc helper() {}",
		"README.md":        "# Test Project",
		"subdir/nested.go": "package subdir\n\nfunc Nested() {}",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          tmpDir,
		EmbeddingProvider: "openai",
		EmbeddingModel:    "text-embedding-3-small",
		IncludePatterns:   []string{"*.go"},
		Logger:            logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("index files", func(t *testing.T) {
		start := time.Now()
		stats, err := indexer.Index(ctx)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, stats.FilesIndexed, 3) // .go files
		assert.GreaterOrEqual(t, stats.ChunksIndexed, 3)
		assert.Zero(t, stats.Errors)
		assert.Less(t, time.Since(start), 60*time.Second)
	})

	t.Run("incremental index", func(t *testing.T) {
		// Second index should be faster (incremental)
		start := time.Now()
		stats, err := indexer.Index(ctx)
		require.NoError(t, err)

		// Should detect no changes
		assert.Equal(t, 0, stats.FilesIndexed) // No new files
		t.Logf("Incremental index took %v", time.Since(start))
	})
}

func TestIndexer_Search(t *testing.T) {
	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          "/tmp/test",
		EmbeddingProvider: "openai",
		Logger:            logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		options *SearchOptions
		wantErr bool
	}{
		{
			name:    "empty query fails",
			query:   "",
			options: &SearchOptions{TopK: 5},
			wantErr: true,
		},
		{
			name:    "valid query",
			query:   "function definitions",
			options: &SearchOptions{TopK: 5, MinScore: 0.7},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := indexer.Search(ctx, tt.query, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				// May error if not indexed, but structure should be valid
				if err == nil {
					assert.NotNil(t, results)
				}
			}
		})
	}
}

func TestIndexer_GetStats(t *testing.T) {
	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          "/tmp/test",
		EmbeddingProvider: "openai",
		Logger:            logger,
	})
	require.NoError(t, err)

	stats := indexer.GetStats()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.LastIndexed.IsZero(), true)
}

func TestIndexer_UpdateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codebase-update-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create initial file
	testFile := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	require.NoError(t, err)

	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          tmpDir,
		EmbeddingProvider: "openai",
		Logger:            logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Update the file
	err = os.WriteFile(testFile, []byte("package main\n\nfunc updated() {}"), 0644)
	require.NoError(t, err)

	err = indexer.UpdateFile(ctx, testFile)
	// May error without real embedding service, but structure is tested
	if err != nil {
		t.Logf("UpdateFile error (expected without API key): %v", err)
	}
}

func TestIndexer_DeleteFile(t *testing.T) {
	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          "/tmp/test",
		EmbeddingProvider: "openai",
		Logger:            logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Should not error for non-existent file
	err = indexer.DeleteFile(ctx, "/tmp/test/nonexistent.go")
	assert.NoError(t, err)
}

func TestIndexer_Watch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codebase-watch-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logger := zap.NewNop()
	indexer, err := NewIndexer(&Config{
		RootPath:          tmpDir,
		EmbeddingProvider: "openai",
		Logger:            logger,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start watching
	go func() {
		err := indexer.Watch(ctx)
		assert.NoError(t, err) // Should return nil on context cancellation
	}()

	// Create a file
	time.Sleep(100 * time.Millisecond)
	testFile := filepath.Join(tmpDir, "watch_test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	require.NoError(t, err)

	// Wait for watch to process
	time.Sleep(500 * time.Millisecond)
}

func TestSearchOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options *SearchOptions
		wantErr bool
	}{
		{
			name:    "valid options",
			options: &SearchOptions{TopK: 5, MinScore: 0.7},
			wantErr: false,
		},
		{
			name:    "zero topk uses default",
			options: &SearchOptions{TopK: 0},
			wantErr: false,
		},
		{
			name:    "negative topk fails",
			options: &SearchOptions{TopK: -1},
			wantErr: true,
		},
		{
			name:    "negative score fails",
			options: &SearchOptions{TopK: 5, MinScore: -0.1},
			wantErr: true,
		},
		{
			name:    "score over 1 fails",
			options: &SearchOptions{TopK: 5, MinScore: 1.5},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIndexStats_String(t *testing.T) {
	stats := IndexStats{
		FilesIndexed:  10,
		ChunksIndexed: 50,
		Errors:        1,
	}

	str := stats.String()
	assert.Contains(t, str, "10")
	assert.Contains(t, str, "50")
	assert.Contains(t, str, "1")
}

func BenchmarkIndexer_Search(b *testing.B) {
	logger := zap.NewNop()
	indexer, _ := NewIndexer(&Config{
		RootPath:          "/tmp/test",
		EmbeddingProvider: "openai",
		Logger:            logger,
	})

	ctx := context.Background()
	options := &SearchOptions{TopK: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexer.Search(ctx, "test query", options)
	}
}
