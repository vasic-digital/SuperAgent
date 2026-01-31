package bigdata

import (
	"fmt"

	"dev.helix.agent/internal/config"
)

// ConfigToIntegrationConfig converts a config.BigDataConfig to bigdata.IntegrationConfig
func ConfigToIntegrationConfig(cfg *config.BigDataConfig) *IntegrationConfig {
	if cfg == nil {
		return DefaultIntegrationConfig()
	}

	return &IntegrationConfig{
		// Enable/disable individual components
		EnableInfiniteContext:   cfg.EnableInfiniteContext,
		EnableDistributedMemory: cfg.EnableDistributedMemory,
		EnableKnowledgeGraph:    cfg.EnableKnowledgeGraph,
		EnableAnalytics:         cfg.EnableAnalytics,
		EnableCrossLearning:     cfg.EnableCrossLearning,

		// Kafka configuration
		KafkaBootstrapServers: cfg.KafkaBootstrapServers,
		KafkaConsumerGroup:    cfg.KafkaConsumerGroup,

		// ClickHouse configuration
		ClickHouseHost:     cfg.ClickHouseHost,
		ClickHousePort:     cfg.ClickHousePort,
		ClickHouseDatabase: cfg.ClickHouseDatabase,
		ClickHouseUser:     cfg.ClickHouseUser,
		ClickHousePassword: cfg.ClickHousePassword,

		// Neo4j configuration
		Neo4jURI:      cfg.Neo4jURI,
		Neo4jUsername: cfg.Neo4jUsername,
		Neo4jPassword: cfg.Neo4jPassword,
		Neo4jDatabase: cfg.Neo4jDatabase,

		// Context engine configuration
		ContextCacheSize:       cfg.ContextCacheSize,
		ContextCacheTTL:        cfg.ContextCacheTTL,
		ContextCompressionType: cfg.ContextCompressionType,

		// Learning configuration
		LearningMinConfidence: cfg.LearningMinConfidence,
		LearningMinFrequency:  cfg.LearningMinFrequency,
	}
}

// IntegrationConfigToConfig converts a bigdata.IntegrationConfig to config.BigDataConfig
func IntegrationConfigToConfig(icfg *IntegrationConfig) *config.BigDataConfig {
	if icfg == nil {
		return &config.BigDataConfig{}
	}

	return &config.BigDataConfig{
		// Enable/disable individual components
		EnableInfiniteContext:   icfg.EnableInfiniteContext,
		EnableDistributedMemory: icfg.EnableDistributedMemory,
		EnableKnowledgeGraph:    icfg.EnableKnowledgeGraph,
		EnableAnalytics:         icfg.EnableAnalytics,
		EnableCrossLearning:     icfg.EnableCrossLearning,

		// Kafka configuration
		KafkaBootstrapServers: icfg.KafkaBootstrapServers,
		KafkaConsumerGroup:    icfg.KafkaConsumerGroup,

		// ClickHouse configuration
		ClickHouseHost:     icfg.ClickHouseHost,
		ClickHousePort:     icfg.ClickHousePort,
		ClickHouseDatabase: icfg.ClickHouseDatabase,
		ClickHouseUser:     icfg.ClickHouseUser,
		ClickHousePassword: icfg.ClickHousePassword,

		// Neo4j configuration
		Neo4jURI:      icfg.Neo4jURI,
		Neo4jUsername: icfg.Neo4jUsername,
		Neo4jPassword: icfg.Neo4jPassword,
		Neo4jDatabase: icfg.Neo4jDatabase,

		// Context engine configuration
		ContextCacheSize:       icfg.ContextCacheSize,
		ContextCacheTTL:        icfg.ContextCacheTTL,
		ContextCompressionType: icfg.ContextCompressionType,

		// Learning configuration
		LearningMinConfidence: icfg.LearningMinConfidence,
		LearningMinFrequency:  icfg.LearningMinFrequency,
	}
}

// DefaultBigDataConfig returns a default config.BigDataConfig
// This is useful for testing and when config package is not available
func DefaultBigDataConfig() *config.BigDataConfig {
	icfg := DefaultIntegrationConfig()
	return IntegrationConfigToConfig(icfg)
}

// ValidateConfig validates the integration configuration
func ValidateConfig(icfg *IntegrationConfig) error {
	if icfg == nil {
		return nil
	}

	// Validate Kafka configuration if cross-learning or infinite context enabled
	if icfg.EnableCrossLearning || icfg.EnableInfiniteContext {
		if icfg.KafkaBootstrapServers == "" {
			return fmt.Errorf("Kafka bootstrap servers required when cross-learning or infinite context enabled")
		}
	}

	// Validate ClickHouse configuration if analytics enabled
	if icfg.EnableAnalytics {
		if icfg.ClickHouseHost == "" {
			return fmt.Errorf("ClickHouse host required when analytics enabled")
		}
		if icfg.ClickHousePort <= 0 {
			return fmt.Errorf("ClickHouse port must be positive")
		}
	}

	// Validate Neo4j configuration if knowledge graph enabled
	if icfg.EnableKnowledgeGraph {
		if icfg.Neo4jURI == "" {
			return fmt.Errorf("Neo4j URI required when knowledge graph enabled")
		}
	}

	// Validate context cache settings
	if icfg.ContextCacheSize < 0 {
		return fmt.Errorf("context cache size must be non-negative")
	}
	if icfg.ContextCacheTTL < 0 {
		return fmt.Errorf("context cache TTL must be non-negative")
	}

	// Validate learning confidence
	if icfg.LearningMinConfidence < 0 || icfg.LearningMinConfidence > 1 {
		return fmt.Errorf("learning min confidence must be between 0 and 1")
	}
	if icfg.LearningMinFrequency < 0 {
		return fmt.Errorf("learning min frequency must be non-negative")
	}

	return nil
}
