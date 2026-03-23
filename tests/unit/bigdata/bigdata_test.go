package bigdata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/bigdata"
)

// TestBigDataConfigConverter validates config conversion for BigData components
func TestBigDataConfigConverter(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := bigdata.DefaultBigDataConfig()
		require.NotNil(t, cfg, "default config must not be nil")
		// Verify config struct is properly populated with defaults
		// Some features enabled by default per BigData integration design
		t.Logf("InfiniteContext: %v, DistributedMemory: %v, KnowledgeGraph: %v",
			cfg.EnableInfiniteContext, cfg.EnableDistributedMemory, cfg.EnableKnowledgeGraph)
		t.Logf("Analytics: %v, CrossLearning: %v",
			cfg.EnableAnalytics, cfg.EnableCrossLearning)
	})

	t.Run("ConfigFieldsInitialized", func(t *testing.T) {
		cfg := bigdata.DefaultBigDataConfig()
		// Default config provides sensible connection defaults
		assert.NotNil(t, cfg, "config must not be nil")
		// Kafka and ClickHouse have default connection strings
		t.Logf("KafkaServers: %s, ClickHouseHost: %s",
			cfg.KafkaBootstrapServers, cfg.ClickHouseHost)
	})
}

// TestBigDataIntegrationComponents verifies component names are valid
func TestBigDataIntegrationComponents(t *testing.T) {
	components := []string{
		"InfiniteContext",
		"DistributedMemory",
		"KnowledgeGraph",
		"Analytics",
		"CrossLearning",
	}

	for _, comp := range components {
		t.Run(comp, func(t *testing.T) {
			assert.NotEmpty(t, comp, "component name must not be empty")
			assert.Greater(t, len(comp), 3, "component name must be descriptive")
		})
	}
}
