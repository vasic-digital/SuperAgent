// Package search provides the search service
package search

import (
	"context"
	"fmt"
	"time"

	containeradapter "dev.helix.agent/internal/adapters/containers"
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
	Enabled          bool
	EmbedderType     string // "openai" or "local"
	OpenAIKey        string
	OpenAIModel      string
	VectorStoreType  string // "chroma" or "qdrant"
	ChromaHost       string
	ChromaPort       int
	QdrantHost       string
	QdrantPort       int
	CollectionName   string
	IndexerConfig    indexer.Config
	ContainerAdapter *containeradapter.Adapter // Container adapter for orchestration
	ComposeFile      string                    // Docker compose file path for container management
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
		ComposeFile:     "docker-compose.yml",
	}
}

// NewService creates a new search service
func NewService(config ServiceConfig, logger *logrus.Logger) (*Service, error) {
	if !config.Enabled {
		return &Service{config: config, logger: logger}, nil
	}

	// Ensure vector store containers are running via container adapter
	if config.ContainerAdapter != nil {
		if err := ensureVectorStoreContainer(context.Background(), config, logger); err != nil {
			logger.WithError(err).Warn("Failed to ensure vector store container, continuing anyway")
		}
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

// ensureVectorStoreContainer ensures the vector store container is running
func ensureVectorStoreContainer(ctx context.Context, config ServiceConfig, logger *logrus.Logger) error {
	if config.ContainerAdapter == nil {
		return fmt.Errorf("container adapter not available")
	}

	var serviceName string
	switch config.VectorStoreType {
	case "chroma":
		serviceName = "chromadb"
	case "qdrant":
		serviceName = "qdrant"
	default:
		return fmt.Errorf("unknown vector store type: %s", config.VectorStoreType)
	}

	logger.WithField("service", serviceName).Info("Ensuring vector store container is running")

	// Check if service is already running via container adapter
	composeFile := config.ComposeFile
	if composeFile == "" {
		composeFile = "docker-compose.yml"
	}

	// Check current status of compose services
	statuses, err := config.ContainerAdapter.ComposeStatus(ctx, composeFile)
	if err != nil {
		logger.WithError(err).Warn("Failed to get compose status, attempting ComposeUp")
		return startVectorStoreContainer(ctx, config.ContainerAdapter, composeFile, serviceName, logger)
	}

	// Check if the specific service is running
	for _, status := range statuses {
		if status.Name == serviceName {
			if status.State == "running" {
				logger.WithField("service", serviceName).Info("Vector store container is already running")
				return nil
			}
			// Service exists but is not running
			logger.WithFields(logrus.Fields{
				"service": status.Name,
				"state":   status.State,
			}).Info("Vector store container found but not running, starting it")
			return startVectorStoreContainer(ctx, config.ContainerAdapter, composeFile, serviceName, logger)
		}
	}

	// Service not found in compose status, try to start it
	logger.WithField("service", serviceName).Info("Vector store container not found, starting via ComposeUp")
	return startVectorStoreContainer(ctx, config.ContainerAdapter, composeFile, serviceName, logger)
}

// startVectorStoreContainer starts the vector store container using ComposeUp
func startVectorStoreContainer(ctx context.Context, adapter *containeradapter.Adapter, composeFile, serviceName string, logger *logrus.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	if err := adapter.ComposeUp(ctx, composeFile, "default"); err != nil {
		return fmt.Errorf("failed to start %s container: %w", serviceName, err)
	}

	logger.WithField("service", serviceName).Info("Vector store container started successfully")
	return nil
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
