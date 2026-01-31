package bigdata

import (
	"context"
	"fmt"
	"time"

	"dev.helix.agent/internal/analytics"
	"dev.helix.agent/internal/conversation"
	"dev.helix.agent/internal/knowledge"
	"dev.helix.agent/internal/learning"
	"dev.helix.agent/internal/memory"
	"dev.helix.agent/pkg/messaging"
	"github.com/sirupsen/logrus"
)

// BigDataIntegration manages all big data components
type BigDataIntegration struct {
	// Core components
	infiniteContext     *conversation.InfiniteContextEngine
	distributedMemory   *memory.DistributedMemoryManager
	graphStreaming      *knowledge.StreamingKnowledgeGraph
	clickhouseAnalytics *analytics.ClickHouseAnalytics
	crossSessionLearner *learning.CrossSessionLearner

	// Messaging
	kafkaBroker messaging.MessageBroker

	// Configuration
	config *IntegrationConfig
	logger *logrus.Logger

	// State
	isRunning bool
}

// IntegrationConfig holds configuration for big data integration
type IntegrationConfig struct {
	// Enable/disable individual components
	EnableInfiniteContext   bool
	EnableDistributedMemory bool
	EnableKnowledgeGraph    bool
	EnableAnalytics         bool
	EnableCrossLearning     bool

	// Kafka configuration
	KafkaBootstrapServers string
	KafkaConsumerGroup    string

	// ClickHouse configuration
	ClickHouseHost     string
	ClickHousePort     int
	ClickHouseDatabase string
	ClickHouseUser     string
	ClickHousePassword string

	// Neo4j configuration
	Neo4jURI      string
	Neo4jUsername string
	Neo4jPassword string
	Neo4jDatabase string

	// Context engine configuration
	ContextCacheSize       int
	ContextCacheTTL        time.Duration
	ContextCompressionType string

	// Learning configuration
	LearningMinConfidence float64
	LearningMinFrequency  int
}

// DefaultIntegrationConfig returns default configuration
func DefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		// Enable all components by default
		EnableInfiniteContext:   true,
		EnableDistributedMemory: true,
		EnableKnowledgeGraph:    true,
		EnableAnalytics:         true,
		EnableCrossLearning:     true,

		// Kafka defaults
		KafkaBootstrapServers: "localhost:9092",
		KafkaConsumerGroup:    "helixagent-bigdata",

		// ClickHouse defaults
		ClickHouseHost:     "localhost",
		ClickHousePort:     9000,
		ClickHouseDatabase: "helixagent_analytics",
		ClickHouseUser:     "default",
		ClickHousePassword: "",

		// Neo4j defaults
		Neo4jURI:      "bolt://localhost:7687",
		Neo4jUsername: "neo4j",
		Neo4jPassword: "helixagent123",
		Neo4jDatabase: "helixagent",

		// Context engine defaults
		ContextCacheSize:       100,
		ContextCacheTTL:        30 * time.Minute,
		ContextCompressionType: "hybrid",

		// Learning defaults
		LearningMinConfidence: 0.7,
		LearningMinFrequency:  3,
	}
}

// NewBigDataIntegration creates a new big data integration
func NewBigDataIntegration(
	config *IntegrationConfig,
	kafkaBroker messaging.MessageBroker,
	logger *logrus.Logger,
) (*BigDataIntegration, error) {
	if config == nil {
		config = DefaultIntegrationConfig()
	}

	integration := &BigDataIntegration{
		config:      config,
		kafkaBroker: kafkaBroker,
		logger:      logger,
		isRunning:   false,
	}

	return integration, nil
}

// Initialize initializes all enabled big data components
func (bdi *BigDataIntegration) Initialize(ctx context.Context) error {
	bdi.logger.Info("Initializing big data integration...")

	// Initialize Infinite Context Engine
	if bdi.config.EnableInfiniteContext {
		if err := bdi.initializeInfiniteContext(ctx); err != nil {
			bdi.logger.WithError(err).Error("Failed to initialize infinite context engine")
			return fmt.Errorf("infinite context initialization failed: %w", err)
		}
		bdi.logger.Info("✓ Infinite context engine initialized")
	}

	// Initialize Distributed Memory
	if bdi.config.EnableDistributedMemory {
		if err := bdi.initializeDistributedMemory(ctx); err != nil {
			bdi.logger.WithError(err).Error("Failed to initialize distributed memory")
			return fmt.Errorf("distributed memory initialization failed: %w", err)
		}
		bdi.logger.Info("✓ Distributed memory initialized")
	}

	// Initialize Knowledge Graph Streaming
	if bdi.config.EnableKnowledgeGraph {
		if err := bdi.initializeKnowledgeGraph(ctx); err != nil {
			bdi.logger.WithError(err).Error("Failed to initialize knowledge graph streaming")
			return fmt.Errorf("knowledge graph initialization failed: %w", err)
		}
		bdi.logger.Info("✓ Knowledge graph streaming initialized")
	}

	// Initialize ClickHouse Analytics
	if bdi.config.EnableAnalytics {
		if err := bdi.initializeAnalytics(ctx); err != nil {
			bdi.logger.WithError(err).Error("Failed to initialize analytics")
			return fmt.Errorf("analytics initialization failed: %w", err)
		}
		bdi.logger.Info("✓ ClickHouse analytics initialized")
	}

	// Initialize Cross-Session Learning
	if bdi.config.EnableCrossLearning {
		if err := bdi.initializeCrossLearning(ctx); err != nil {
			bdi.logger.WithError(err).Error("Failed to initialize cross-session learning")
			return fmt.Errorf("cross-session learning initialization failed: %w", err)
		}
		bdi.logger.Info("✓ Cross-session learning initialized")
	}

	bdi.logger.Info("Big data integration initialized successfully")
	return nil
}

// initializeInfiniteContext initializes the infinite context engine
func (bdi *BigDataIntegration) initializeInfiniteContext(ctx context.Context) error {
	config := conversation.InfiniteContextConfig{
		CacheSize:       bdi.config.ContextCacheSize,
		CacheTTL:        bdi.config.ContextCacheTTL,
		CompressionType: bdi.config.ContextCompressionType,
	}

	engine, err := conversation.NewInfiniteContextEngine(
		config,
		bdi.kafkaBroker,
		bdi.logger,
	)
	if err != nil {
		return err
	}

	bdi.infiniteContext = engine
	return nil
}

// initializeDistributedMemory initializes the distributed memory manager
func (bdi *BigDataIntegration) initializeDistributedMemory(ctx context.Context) error {
	config := memory.DistributedMemoryConfig{
		NodeID:           fmt.Sprintf("node-%d", time.Now().Unix()),
		EventTopic:       "helixagent.memory.events",
		SnapshotTopic:    "helixagent.memory.snapshots",
		ConflictTopic:    "helixagent.memory.conflicts",
		CRDTStrategy:     "merge_all",
		SnapshotInterval: 5 * time.Minute,
	}

	manager, err := memory.NewDistributedMemoryManager(
		config,
		bdi.kafkaBroker,
		bdi.logger,
	)
	if err != nil {
		return err
	}

	// Start event consumer
	if err := manager.StartEventConsumer(ctx); err != nil {
		return err
	}

	bdi.distributedMemory = manager
	return nil
}

// initializeKnowledgeGraph initializes the knowledge graph streaming
func (bdi *BigDataIntegration) initializeKnowledgeGraph(ctx context.Context) error {
	config := knowledge.GraphStreamingConfig{
		Neo4jURI:          bdi.config.Neo4jURI,
		Neo4jUsername:     bdi.config.Neo4jUsername,
		Neo4jPassword:     bdi.config.Neo4jPassword,
		Neo4jDatabase:     bdi.config.Neo4jDatabase,
		EntityTopic:       "helixagent.entities.updates",
		RelationshipTopic: "helixagent.relationships.updates",
	}

	graph, err := knowledge.NewStreamingKnowledgeGraph(
		config,
		bdi.kafkaBroker,
		bdi.logger,
	)
	if err != nil {
		return err
	}

	// Start streaming updates
	if err := graph.StartStreaming(ctx); err != nil {
		return err
	}

	bdi.graphStreaming = graph
	return nil
}

// initializeAnalytics initializes the ClickHouse analytics
func (bdi *BigDataIntegration) initializeAnalytics(ctx context.Context) error {
	config := analytics.ClickHouseConfig{
		Host:     bdi.config.ClickHouseHost,
		Port:     bdi.config.ClickHousePort,
		Database: bdi.config.ClickHouseDatabase,
		Username: bdi.config.ClickHouseUser,
		Password: bdi.config.ClickHousePassword,
	}

	client, err := analytics.NewClickHouseAnalytics(
		config,
		bdi.logger,
	)
	if err != nil {
		return err
	}

	// Test connection
	if err := client.HealthCheck(ctx); err != nil {
		return fmt.Errorf("clickhouse health check failed: %w", err)
	}

	bdi.clickhouseAnalytics = client
	return nil
}

// initializeCrossLearning initializes the cross-session learner
func (bdi *BigDataIntegration) initializeCrossLearning(ctx context.Context) error {
	config := learning.CrossSessionConfig{
		CompletedTopic: "helixagent.conversations.completed",
		InsightsTopic:  "helixagent.learning.insights",
		MinConfidence:  bdi.config.LearningMinConfidence,
		MinFrequency:   bdi.config.LearningMinFrequency,
	}

	learner := learning.NewCrossSessionLearner(
		config,
		bdi.kafkaBroker,
		bdi.logger,
	)

	// Start learning
	if err := learner.StartLearning(ctx); err != nil {
		return err
	}

	bdi.crossSessionLearner = learner
	return nil
}

// Start starts all big data components
func (bdi *BigDataIntegration) Start(ctx context.Context) error {
	if bdi.isRunning {
		return fmt.Errorf("big data integration is already running")
	}

	bdi.logger.Info("Starting big data integration...")

	// All components are already started during initialization
	// This method is provided for consistency

	bdi.isRunning = true
	bdi.logger.Info("Big data integration started successfully")
	return nil
}

// Stop stops all big data components
func (bdi *BigDataIntegration) Stop(ctx context.Context) error {
	if !bdi.isRunning {
		return nil
	}

	bdi.logger.Info("Stopping big data integration...")

	// Stop distributed memory
	if bdi.distributedMemory != nil {
		if err := bdi.distributedMemory.Stop(ctx); err != nil {
			bdi.logger.WithError(err).Warn("Error stopping distributed memory")
		}
	}

	// Stop knowledge graph streaming
	if bdi.graphStreaming != nil {
		if err := bdi.graphStreaming.Stop(ctx); err != nil {
			bdi.logger.WithError(err).Warn("Error stopping knowledge graph streaming")
		}
	}

	// Close ClickHouse connection
	if bdi.clickhouseAnalytics != nil {
		if err := bdi.clickhouseAnalytics.Close(); err != nil {
			bdi.logger.WithError(err).Warn("Error closing ClickHouse connection")
		}
	}

	bdi.isRunning = false
	bdi.logger.Info("Big data integration stopped")
	return nil
}

// GetInfiniteContext returns the infinite context engine
func (bdi *BigDataIntegration) GetInfiniteContext() *conversation.InfiniteContextEngine {
	return bdi.infiniteContext
}

// GetDistributedMemory returns the distributed memory manager
func (bdi *BigDataIntegration) GetDistributedMemory() *memory.DistributedMemoryManager {
	return bdi.distributedMemory
}

// GetKnowledgeGraph returns the knowledge graph streaming
func (bdi *BigDataIntegration) GetKnowledgeGraph() *knowledge.StreamingKnowledgeGraph {
	return bdi.graphStreaming
}

// GetAnalytics returns the ClickHouse analytics client
func (bdi *BigDataIntegration) GetAnalytics() *analytics.ClickHouseAnalytics {
	return bdi.clickhouseAnalytics
}

// GetCrossLearner returns the cross-session learner
func (bdi *BigDataIntegration) GetCrossLearner() *learning.CrossSessionLearner {
	return bdi.crossSessionLearner
}

// IsRunning returns whether the integration is running
func (bdi *BigDataIntegration) IsRunning() bool {
	return bdi.isRunning
}

// HealthCheck checks the health of all components
func (bdi *BigDataIntegration) HealthCheck(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// Check infinite context
	if bdi.infiniteContext != nil {
		health["infinite_context"] = "healthy"
	} else if bdi.config.EnableInfiniteContext {
		health["infinite_context"] = "not_initialized"
	} else {
		health["infinite_context"] = "disabled"
	}

	// Check distributed memory
	if bdi.distributedMemory != nil {
		health["distributed_memory"] = "healthy"
	} else if bdi.config.EnableDistributedMemory {
		health["distributed_memory"] = "not_initialized"
	} else {
		health["distributed_memory"] = "disabled"
	}

	// Check knowledge graph
	if bdi.graphStreaming != nil {
		health["knowledge_graph"] = "healthy"
	} else if bdi.config.EnableKnowledgeGraph {
		health["knowledge_graph"] = "not_initialized"
	} else {
		health["knowledge_graph"] = "disabled"
	}

	// Check analytics
	if bdi.clickhouseAnalytics != nil {
		if err := bdi.clickhouseAnalytics.HealthCheck(ctx); err != nil {
			health["analytics"] = "unhealthy"
		} else {
			health["analytics"] = "healthy"
		}
	} else if bdi.config.EnableAnalytics {
		health["analytics"] = "not_initialized"
	} else {
		health["analytics"] = "disabled"
	}

	// Check cross-session learning
	if bdi.crossSessionLearner != nil {
		health["cross_learning"] = "healthy"
	} else if bdi.config.EnableCrossLearning {
		health["cross_learning"] = "not_initialized"
	} else {
		health["cross_learning"] = "disabled"
	}

	return health
}
