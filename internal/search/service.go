// Package search provides the search service
package search

import (
	"context"
	"fmt"

	"dev.helix.agent/internal/search/embedder"
	"dev.helix.agent/internal/search/indexer"
	"dev.helix.agent/internal/search/store"
	"dev.helix.agent/internal/search/types"
	"github.com/sirupsen/logrus"
)

// Service provides semantic search capabilities
type Service struct {
	Indexer  Indexer
	Searcher Searcher
	config   ServiceConfig
	logger   *logrus.Logger
}

// ServiceConfig holds service configuration
type ServiceConfig struct {
	Enabled         bool
	EmbedderType    string // "openai" or "local"
	OpenAIKey       string
	OpenAIModel     string
	VectorStoreType string // "chroma" or "qdrant"
	ChromaHost      string
	ChromaPort      int
	QdrantHost      string
	QdrantPort      int
	CollectionName  string
	IndexerConfig   indexer.Config
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Enabled:         true,
		EmbedderType:    "local",
		OpenAIModel:     "text-embedding-3-small",
		VectorStoreType: "chroma",
		ChromaHost:      "localhost",
		ChromaPort:      8000,
		QdrantHost:      "localhost",
		QdrantPort:      6333,
		CollectionName:  "code_embeddings",
		IndexerConfig:   indexer.DefaultConfig(),
	}
}

// NewService creates a new search service
func NewService(config ServiceConfig, logger *logrus.Logger) (*Service, error) {
	if !config.Enabled {
		return &Service{config: config, logger: logger}, nil
	}

	// Create embedder
	var emb types.Embedder
	var err error

	switch config.EmbedderType {
	case "openai":
		if config.OpenAIKey == "" {
			return nil, fmt.Errorf("OpenAI API key required")
		}
		emb = embedder.NewOpenAIEmbedder(config.OpenAIKey, config.OpenAIModel)
		logger.Info("Using OpenAI embedder")
	case "local":
		emb = embedder.NewLocalEmbedder(1536)
		logger.Info("Using local embedder")
	default:
		return nil, fmt.Errorf("unknown embedder type: %s", config.EmbedderType)
	}

	// Create vector store
	var vectorStore types.VectorStore

	switch config.VectorStoreType {
	case "chroma":
		vectorStore, err = store.NewChromaStore(config.ChromaHost, config.ChromaPort, config.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to create Chroma store: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"host": config.ChromaHost,
			"port": config.ChromaPort,
		}).Info("Using Chroma vector store")

	case "qdrant":
		vectorStore, err = store.NewQdrantStore(config.QdrantHost, config.QdrantPort, config.CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to create Qdrant store: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"host": config.QdrantHost,
			"port": config.QdrantPort,
		}).Info("Using Qdrant vector store")

	default:
		return nil, fmt.Errorf("unknown vector store type: %s", config.VectorStoreType)
	}

	// Create indexer
	idx := indexer.NewCodeIndexer(emb, vectorStore, config.IndexerConfig)

	// Create searcher
	searcher := NewCodeSearcher(emb, vectorStore, config.CollectionName)

	service := &Service{
		Indexer:  idx,
		Searcher: searcher,
		config:   config,
		logger:   logger,
	}

	return service, nil
}

// Initialize creates collection and optionally indexes on startup
func (s *Service) Initialize(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Search service disabled")
		return nil
	}

	// Create collection
	// (This is done automatically by the indexer or store)

	// Index on startup if configured
	if s.config.IndexerConfig.IndexOnStartup {
		s.logger.Info("Starting initial indexing...")
		result, err := s.Indexer.Index(ctx)
		if err != nil {
			return fmt.Errorf("failed to index: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"files_indexed": result.FilesIndexed,
			"duration_ms":   result.Duration.Milliseconds(),
		}).Info("Initial indexing complete")
	}

	// Start file watcher if configured
	if s.config.IndexerConfig.WatchFiles {
		watcher, err := indexer.NewFileWatcher(s.Indexer.(*indexer.CodeIndexer), s.config.IndexerConfig)
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}

		if err := watcher.Start(ctx); err != nil {
			return fmt.Errorf("failed to start file watcher: %w", err)
		}

		s.logger.Info("File watching enabled")
	}

	return nil
}

// Search performs a semantic search
func (s *Service) Search(ctx context.Context, query string, opts types.SearchOptions) ([]types.SearchResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("search service disabled")
	}
	return s.Searcher.Search(ctx, query, opts)
}

// Index triggers full reindexing
func (s *Service) Index(ctx context.Context) (*types.IndexResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("search service disabled")
	}
	return s.Indexer.Index(ctx)
}

// Health returns service health status
func (s *Service) Health() map[string]interface{} {
	if !s.config.Enabled {
		return map[string]interface{}{
			"status":  "disabled",
			"enabled": false,
		}
	}

	return map[string]interface{}{
		"status":       "healthy",
		"enabled":      true,
		"embedder":     s.config.EmbedderType,
		"vector_store": s.config.VectorStoreType,
		"collection":   s.config.CollectionName,
	}
}
