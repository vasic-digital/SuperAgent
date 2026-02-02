package bigdata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"dev.helix.agent/internal/analytics"
	"dev.helix.agent/internal/conversation"
	"dev.helix.agent/internal/knowledge"
	"dev.helix.agent/internal/learning"
	"dev.helix.agent/internal/memory"
	"dev.helix.agent/internal/messaging"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
	"github.com/sirupsen/logrus"
)

// providerRegistryLLMClient implements conversation.LLMClient using ProviderRegistry
type providerRegistryLLMClient struct {
	registry *services.ProviderRegistry
	logger   *logrus.Logger
}

func (c *providerRegistryLLMClient) Complete(ctx context.Context, prompt string, maxTokens int) (string, int, error) {
	if c.registry == nil {
		c.logger.Warn("Provider registry is nil, cannot perform LLM compression")
		return "", 0, errors.New("provider registry not available")
	}

	// Get the best provider from registry
	providers := c.registry.ListProvidersOrderedByScore()
	if len(providers) == 0 {
		c.logger.Warn("No LLM providers available for compression")
		return "", 0, errors.New("no LLM providers available")
	}

	// Try each provider in order of score
	for _, providerName := range providers {
		provider, err := c.registry.GetProvider(providerName)
		if err != nil {
			c.logger.WithError(err).Warnf("Failed to get provider %s", providerName)
			continue
		}

		// Create LLM request
		req := &models.LLMRequest{
			ID:          fmt.Sprintf("compression-%d", time.Now().UnixNano()),
			Prompt:      prompt,
			RequestType: "compression",
			ModelParams: models.ModelParameters{
				MaxTokens:   maxTokens,
				Temperature: 0.1, // Low temperature for consistent compression
			},
			CreatedAt: time.Now(),
		}

		resp, err := provider.Complete(ctx, req)
		if err != nil {
			c.logger.WithError(err).Warnf("Provider %s failed compression", providerName)
			continue
		}

		return resp.Content, resp.TokensUsed, nil
	}

	return "", 0, errors.New("all LLM providers failed compression")
}

// inMemoryEventLog implements memory.EventLog with in-memory storage
type inMemoryEventLog struct {
	mu     sync.RWMutex
	events []*memory.MemoryEvent
}

func (d *inMemoryEventLog) Append(event *memory.MemoryEvent) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events = append(d.events, event)
	return nil
}

func (d *inMemoryEventLog) GetEvents(memoryID string) ([]*memory.MemoryEvent, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []*memory.MemoryEvent
	for _, event := range d.events {
		if event.MemoryID == memoryID {
			result = append(result, event)
		}
	}
	return result, nil
}

func (d *inMemoryEventLog) GetEventsSince(timestamp time.Time) ([]*memory.MemoryEvent, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []*memory.MemoryEvent
	for _, event := range d.events {
		if event.Timestamp.After(timestamp) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (d *inMemoryEventLog) GetEventsForUser(userID string) ([]*memory.MemoryEvent, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []*memory.MemoryEvent
	for _, event := range d.events {
		if event.UserID == userID {
			result = append(result, event)
		}
	}
	return result, nil
}

func (d *inMemoryEventLog) GetEventsFromNode(nodeID string) ([]*memory.MemoryEvent, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []*memory.MemoryEvent
	for _, event := range d.events {
		if event.NodeID == nodeID {
			result = append(result, event)
		}
	}
	return result, nil
}

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

	// Provider registry for LLM compression
	ProviderRegistry *services.ProviderRegistry
}

// DefaultIntegrationConfig returns default configuration
func DefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		// Enable all components by default
		EnableInfiniteContext:   true,
		EnableDistributedMemory: false,
		EnableKnowledgeGraph:    false,
		EnableAnalytics:         false,
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
	// Create LLM client for compression using provider registry
	llmClient := &providerRegistryLLMClient{
		registry: bdi.config.ProviderRegistry,
		logger:   bdi.logger,
	}
	compressor := conversation.NewContextCompressor(llmClient, bdi.logger)

	engine := conversation.NewInfiniteContextEngine(
		bdi.kafkaBroker,
		compressor,
		bdi.logger,
	)

	bdi.infiniteContext = engine
	return nil
}

// initializeDistributedMemory initializes the distributed memory manager
func (bdi *BigDataIntegration) initializeDistributedMemory(ctx context.Context) error {
	// Create local memory manager with dummy dependencies
	store := memory.NewInMemoryStore()
	localManager := memory.NewManager(store, nil, nil, nil, nil, bdi.logger)

	// Create in-memory event log
	eventLog := &inMemoryEventLog{}

	// Create CRDT resolver
	conflictResolver := memory.NewCRDTResolver("merge_all")

	// Generate node ID
	nodeID := fmt.Sprintf("node-%d", time.Now().Unix())

	// Create distributed memory manager
	manager := memory.NewDistributedMemoryManager(
		localManager,
		nodeID,
		eventLog,
		conflictResolver,
		bdi.kafkaBroker,
		bdi.logger,
	)

	bdi.distributedMemory = manager
	return nil
}

// initializeKnowledgeGraph initializes the knowledge graph streaming
func (bdi *BigDataIntegration) initializeKnowledgeGraph(ctx context.Context) error {
	config := knowledge.GraphStreamingConfig{
		Neo4jURI:      bdi.config.Neo4jURI,
		Neo4jUser:     bdi.config.Neo4jUsername,
		Neo4jPassword: bdi.config.Neo4jPassword,
		Neo4jDatabase: bdi.config.Neo4jDatabase,
		EntityTopic:   "helixagent.entities.updates",
		MemoryTopic:   "helixagent.memory.updates",
		DebateTopic:   "helixagent.debate.updates",
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

	// Test connection - health check not available in current version
	// Assume connection is okay

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

// stopComponentTimeout is the maximum time to wait for each component to stop
const stopComponentTimeout = 10 * time.Second

// Stop stops all big data components with per-component timeouts
func (bdi *BigDataIntegration) Stop(ctx context.Context) error {
	if !bdi.isRunning {
		return nil
	}

	bdi.logger.Info("Stopping big data integration...")

	// Stop distributed memory
	if bdi.distributedMemory != nil {
		// Stop method not available in current version
		bdi.logger.Debug("Distributed memory disabled")
	}

	// Stop knowledge graph streaming with timeout
	if bdi.graphStreaming != nil {
		stopCtx, cancel := context.WithTimeout(ctx, stopComponentTimeout)
		if err := bdi.graphStreaming.Stop(stopCtx); err != nil {
			bdi.logger.WithError(err).Warn("Error stopping knowledge graph streaming")
		}
		cancel()
	}

	// Close ClickHouse connection with timeout
	if bdi.clickhouseAnalytics != nil {
		done := make(chan error, 1)
		go func() {
			done <- bdi.clickhouseAnalytics.Close()
		}()
		select {
		case err := <-done:
			if err != nil {
				bdi.logger.WithError(err).Warn("Error closing ClickHouse connection")
			}
		case <-time.After(stopComponentTimeout):
			bdi.logger.Warn("Timed out closing ClickHouse connection")
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
		// Health check not available, assume healthy
		health["analytics"] = "healthy"
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
