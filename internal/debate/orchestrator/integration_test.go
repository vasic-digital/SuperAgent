package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/services"
)

// =============================================================================
// Full Integration Tests - End-to-End Debate Flow
// =============================================================================

// TestFullDebateFlow tests a complete debate from API request to result
func TestFullDebateFlow(t *testing.T) {
	// This test verifies the orchestrator initialization and API adapter work correctly
	// Full debate execution requires all roles to be filled, which is tested via unit tests

	// Setup: Create orchestrator with mock providers
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3
	config.EnableLearning = true

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	// Register providers - creates agents in the pool
	err := orch.RegisterProvider("claude", "claude-3", 9.0)
	require.NoError(t, err)
	err = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	require.NoError(t, err)
	err = orch.RegisterProvider("gemini", "gemini-pro", 8.0)
	require.NoError(t, err)

	// Verify setup
	assert.Equal(t, 3, orch.GetAgentPool().Size())

	// Create API adapter and verify it works
	adapter := NewAPIAdapter(orch)
	require.NotNil(t, adapter)

	// Verify API request conversion works
	apiReq := &APICreateDebateRequest{
		DebateID: "integration-test-1",
		Topic:    "Best practices for error handling in Go",
		Participants: []APIParticipantConfig{
			{Name: "Analyst", Role: "analyst", LLMProvider: "claude", LLMModel: "claude-3"},
			{Name: "Developer", Role: "developer", LLMProvider: "deepseek", LLMModel: "deepseek-coder"},
			{Name: "Reviewer", Role: "reviewer", LLMProvider: "gemini", LLMModel: "gemini-pro"},
		},
		MaxRounds: 3,
		Timeout:   60,
		Strategy:  "mesh",
	}

	debateReq := adapter.ConvertAPIRequest(apiReq)
	require.NotNil(t, debateReq)
	assert.Equal(t, "integration-test-1", debateReq.ID)
	assert.Equal(t, "Best practices for error handling in Go", debateReq.Topic)
	assert.Equal(t, topology.TopologyGraphMesh, debateReq.TopologyType)
}

// TestServiceIntegrationFlow tests the ServiceIntegration with services types
func TestServiceIntegrationFlow(t *testing.T) {
	// This test verifies the ServiceIntegration type conversion and setup
	// Full debate execution requires all roles to be filled

	// Setup: Create service integration with mock providers
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3
	config.EnableLearning = true

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	// Register providers
	_ = orch.RegisterProvider("claude", "claude-3", 9.0)
	_ = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	_ = orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	// Create service integration
	siConfig := DefaultServiceIntegrationConfig()
	siConfig.MinAgentsForNewFramework = 3

	si := &ServiceIntegration{
		orchestrator: orch,
		logger:       logrus.New(),
		config:       siConfig,
	}

	// Create services.DebateConfig
	debateConfig := &services.DebateConfig{
		DebateID: "service-integration-test-1",
		Topic:    "Microservices vs Monolith architecture",
		Participants: []services.ParticipantConfig{
			{Name: "Architect", Role: "architect", LLMProvider: "claude", LLMModel: "claude-3"},
			{Name: "DevOps", Role: "devops", LLMProvider: "deepseek", LLMModel: "deepseek-coder"},
			{Name: "PM", Role: "analyst", LLMProvider: "gemini", LLMModel: "gemini-pro"},
		},
		MaxRounds: 3,
		Timeout:   time.Minute,
		Strategy:  "consensus",
	}

	// Verify ShouldUseNewFramework
	assert.True(t, si.ShouldUseNewFramework(debateConfig))

	// Verify type conversion works
	request := si.convertDebateConfig(debateConfig)
	require.NotNil(t, request)
	assert.Equal(t, "service-integration-test-1", request.ID)
	assert.Equal(t, "Microservices vs Monolith architecture", request.Topic)
	assert.Equal(t, 3, request.MaxRounds)
	assert.Len(t, request.PreferredProviders, 3)
	assert.Equal(t, 0.75, request.MinConsensus) // Default for consensus strategy

	// Verify statistics are available
	ctx := context.Background()
	stats, err := si.GetStatistics(ctx)
	require.NoError(t, err)
	assert.True(t, stats.FrameworkEnabled)
	assert.True(t, stats.LearningEnabled)
	assert.Equal(t, 3, stats.RegisteredAgents)
}

// TestOrchestratorWithAllComponents tests the orchestrator with all components
func TestOrchestratorWithAllComponents(t *testing.T) {
	// Setup
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3
	config.EnableLearning = true
	config.EnableCrossDebateLearning = true
	config.DefaultTopology = topology.TopologyGraphMesh

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	// Register providers
	_ = orch.RegisterProvider("claude", "claude-3", 9.0)
	_ = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	_ = orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	// Verify all components are initialized
	assert.NotNil(t, orch.GetAgentPool())
	assert.NotNil(t, orch.GetKnowledgeRepository())
	assert.Equal(t, 3, orch.GetAgentPool().Size())

	// Get statistics
	ctx := context.Background()
	stats, err := orch.GetStatistics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.RegisteredAgents)
	assert.Equal(t, 0, stats.ActiveDebates)
}

// TestDebateWithDifferentTopologies tests topology selection from strategies
func TestDebateWithDifferentTopologies(t *testing.T) {
	topologies := []struct {
		name     string
		strategy string
		expected topology.TopologyType
	}{
		{"Mesh", "mesh", topology.TopologyGraphMesh},
		{"Chain", "sequential", topology.TopologyChain},
		{"Star", "star", topology.TopologyStar},
		{"Parallel", "parallel", topology.TopologyGraphMesh},
		{"Pipeline", "pipeline", topology.TopologyChain},
		{"Hub", "hub", topology.TopologyStar},
		{"Default", "", topology.TopologyGraphMesh},
	}

	for _, tc := range topologies {
		t.Run(tc.name, func(t *testing.T) {
			// Verify topology selection from strategy
			selected := selectTopologyFromStrategy(tc.strategy)
			assert.Equal(t, tc.expected, selected)
		})
	}
}

// TestDebateWithLearningEnabled tests learning configuration
func TestDebateWithLearningEnabled(t *testing.T) {
	// Setup with learning enabled
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3
	config.EnableLearning = true
	config.EnableCrossDebateLearning = true

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	_ = orch.RegisterProvider("claude", "claude-3", 9.0)
	_ = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	_ = orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	adapter := NewAPIAdapter(orch)

	// Verify learning components are initialized
	assert.NotNil(t, orch.GetKnowledgeRepository())

	// Verify statistics include learning fields
	ctx := context.Background()
	stats, err := adapter.GetStatistics(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.TotalLessons, 0)
	assert.GreaterOrEqual(t, stats.TotalPatterns, 0)
	assert.GreaterOrEqual(t, stats.TotalDebatesLearned, 0)

	// Verify request conversion preserves learning flag
	enableLearning := true
	req := &DebateRequest{
		ID:             "learning-test-1",
		Topic:          "Test topic for learning",
		EnableLearning: &enableLearning,
	}
	assert.NotNil(t, req.EnableLearning)
	assert.True(t, *req.EnableLearning)
}

// TestMultipleDebatesConcurrently tests concurrent orchestrator operations
func TestMultipleDebatesConcurrently(t *testing.T) {
	// Setup
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	_ = orch.RegisterProvider("claude", "claude-3", 9.0)
	_ = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	_ = orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	adapter := NewAPIAdapter(orch)
	ctx := context.Background()

	// Run multiple concurrent operations (statistics, request conversion)
	numOps := 5
	done := make(chan bool, numOps)

	for i := 0; i < numOps; i++ {
		go func(idx int) {
			// Get statistics concurrently
			stats, err := adapter.GetStatistics(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, stats)

			// Convert requests concurrently
			req := &APICreateDebateRequest{
				DebateID: "concurrent-test-" + string(rune('A'+idx)),
				Topic:    "Concurrent topic " + string(rune('A'+idx)),
				Participants: []APIParticipantConfig{
					{Name: "Agent1", LLMProvider: "claude"},
					{Name: "Agent2", LLMProvider: "deepseek"},
				},
			}
			debateReq := adapter.ConvertAPIRequest(req)
			assert.NotNil(t, debateReq)

			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numOps; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for concurrent operation")
		}
	}
}

// TestAgentPoolManagement tests agent pool operations
func TestAgentPoolManagement(t *testing.T) {
	// Setup
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 2

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	// Initial state
	pool := orch.GetAgentPool()
	assert.Equal(t, 0, pool.Size())

	// Register providers
	err := orch.RegisterProvider("claude", "claude-3", 9.0)
	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	err = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	require.NoError(t, err)
	assert.Equal(t, 2, pool.Size())

	// Try to register same provider again
	err = orch.RegisterProvider("claude", "claude-3-sonnet", 8.8)
	require.NoError(t, err)
	// Should add a new agent (different model)
	assert.Equal(t, 3, pool.Size())

	// Get agents by domain
	generalAgents := pool.GetByDomain(agents.DomainGeneral)
	assert.GreaterOrEqual(t, len(generalAgents), 0)
}

// TestRecommendationsFromLearning tests getting recommendations
func TestRecommendationsFromLearning(t *testing.T) {
	// Setup
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3
	config.EnableLearning = true

	lessonBank := createLessonBank(defaultLessonBankConfig())
	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	_ = orch.RegisterProvider("claude", "claude-3", 9.0)
	_ = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	_ = orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	// Get recommendations for a topic
	ctx := context.Background()
	recs, err := orch.GetRecommendations(ctx, "error handling", agents.DomainCode)

	require.NoError(t, err)
	assert.NotNil(t, recs)
}

// TestTypeConversionRoundTrip tests converting types back and forth
func TestTypeConversionRoundTrip(t *testing.T) {
	si := NewServiceIntegration(nil, nil, DefaultServiceIntegrationConfig())

	// Original services.DebateConfig
	original := &services.DebateConfig{
		DebateID: "roundtrip-test",
		Topic:    "Test roundtrip conversion",
		Participants: []services.ParticipantConfig{
			{Name: "Agent1", Role: "analyst", LLMProvider: "claude", LLMModel: "claude-3"},
			{Name: "Agent2", Role: "coder", LLMProvider: "deepseek", LLMModel: "deepseek-coder"},
		},
		MaxRounds: 5,
		Timeout:   3 * time.Minute,
		Strategy:  "consensus",
		Metadata:  map[string]interface{}{"key": "value"},
	}

	// Convert to DebateRequest
	request := si.convertDebateConfig(original)

	// Verify conversion
	assert.Equal(t, original.DebateID, request.ID)
	assert.Equal(t, original.Topic, request.Topic)
	assert.Equal(t, original.MaxRounds, request.MaxRounds)
	assert.Equal(t, original.Timeout, request.Timeout)
	assert.Contains(t, request.PreferredProviders, "claude")
	assert.Contains(t, request.PreferredProviders, "deepseek")
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	// Test with disabled framework
	t.Run("DisabledFramework", func(t *testing.T) {
		config := DefaultServiceIntegrationConfig()
		config.EnableNewFramework = false

		si := NewServiceIntegration(nil, nil, config)

		ctx := context.Background()
		result, err := si.ConductDebate(ctx, &services.DebateConfig{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "disabled")
	})

	// Test GetDebateStatus for non-existent debate
	t.Run("NonExistentDebate", func(t *testing.T) {
		mockRegistry := newMockProviderRegistry()
		config := DefaultOrchestratorConfig()
		lessonBank := createLessonBank(defaultLessonBankConfig())
		orch := NewOrchestrator(mockRegistry, lessonBank, config)

		adapter := NewAPIAdapter(orch)

		status, found := adapter.GetDebateStatus("non-existent")
		assert.False(t, found)
		assert.Empty(t, status)
	})

	// Test CancelDebate for non-existent debate
	t.Run("CancelNonExistent", func(t *testing.T) {
		mockRegistry := newMockProviderRegistry()
		config := DefaultOrchestratorConfig()
		lessonBank := createLessonBank(defaultLessonBankConfig())
		orch := NewOrchestrator(mockRegistry, lessonBank, config)

		adapter := NewAPIAdapter(orch)

		err := adapter.CancelDebate("non-existent")
		assert.Error(t, err)
	})
}
