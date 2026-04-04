package codebase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewIndexer(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIndexConfig()

	indexer := NewIndexer(logger, "/tmp/test", config)
	assert.NotNil(t, indexer)
	assert.Equal(t, "/tmp/test", indexer.rootPath)
	assert.Equal(t, config, indexer.config)
}

func TestIndexer_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultIndexConfig()
	indexer := NewIndexer(logger, "/tmp/test", config)

	stats := indexer.GetStats()
	assert.NotNil(t, stats)
}

func TestDefaultIndexConfig(t *testing.T) {
	config := DefaultIndexConfig()
	assert.NotEmpty(t, config.IncludePatterns)
	assert.NotEmpty(t, config.ExcludePatterns)
	assert.Greater(t, config.ChunkSize, 0)
	assert.Greater(t, config.MaxFileSize, int64(0))
}
