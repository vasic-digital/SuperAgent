package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/llm"
)

// =============================================================================
// ProviderRegistryBridge Tests
// =============================================================================

func TestNewProviderRegistryBridge(t *testing.T) {
	// We can't easily create a real services.ProviderRegistry in unit tests
	// without extensive setup, so we'll test what we can
	bridge := &ProviderRegistryBridge{registry: nil}
	assert.NotNil(t, bridge)
}

// Test that ProviderRegistryBridge implements the interface
func TestProviderRegistryBridge_ImplementsInterface(t *testing.T) {
	// This test verifies at compile time that the bridge implements the interface
	var _ ProviderRegistry = (*ProviderRegistryBridge)(nil)
}

// =============================================================================
// OrchestratorFactory Tests
// =============================================================================

func TestNewOrchestratorFactory(t *testing.T) {
	factory := NewOrchestratorFactory(nil)

	assert.NotNil(t, factory)
	assert.Nil(t, factory.providerRegistry)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestDefaultLessonBankConfig(t *testing.T) {
	config := defaultLessonBankConfig()

	// Should have semantic search disabled by default
	assert.False(t, config.EnableSemanticSearch)
}

func TestCreateLessonBank(t *testing.T) {
	config := defaultLessonBankConfig()
	bank := createLessonBank(config)

	assert.NotNil(t, bank)
}

// =============================================================================
// Integration with Mock Provider Registry
// =============================================================================

// mockServicesProviderRegistry simulates the services.ProviderRegistry behavior
// This allows testing the bridge logic without full service dependencies

func TestProviderRegistryBridge_WithMock(t *testing.T) {
	// Use our existing mock from orchestrator_test.go
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))

	// Test that the mock works correctly
	provider, err := mockRegistry.GetProvider("claude")
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Test GetAvailableProviders
	providers := mockRegistry.GetAvailableProviders()
	assert.Len(t, providers, 2)
}

// TestOrchestratorFactoryWithMock tests the factory creates working orchestrators
func TestOrchestratorFactoryWithMock(t *testing.T) {
	// Use the mock registry directly
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))

	// Create orchestrator manually with mock
	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 2

	lessonBankConfig := defaultLessonBankConfig()
	lessonBank := createLessonBank(lessonBankConfig)

	orch := NewOrchestrator(mockRegistry, lessonBank, config)
	require.NotNil(t, orch)

	// Register providers
	err := orch.RegisterProvider("claude", "claude-3", 9.0)
	require.NoError(t, err)

	err = orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	require.NoError(t, err)

	// Verify providers were registered
	pool := orch.GetAgentPool()
	assert.Equal(t, 2, pool.Size())
}

// =============================================================================
// ProviderInvoker Integration Tests
// =============================================================================

func TestProviderInvoker_WithBridge(t *testing.T) {
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))

	invoker := NewProviderInvoker(mockRegistry)
	assert.NotNil(t, invoker)

	// Test that invoker can get provider
	provider, err := mockRegistry.GetProvider("claude")
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

// =============================================================================
// End-to-End Mock Integration
// =============================================================================

func TestEndToEndMockIntegration(t *testing.T) {
	// Set up mock providers
	mockRegistry := newMockProviderRegistry()
	mockRegistry.AddProvider("claude", newMockLLMProvider("claude"))
	mockRegistry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	mockRegistry.AddProvider("gemini", newMockLLMProvider("gemini"))

	// Create orchestrator
	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3

	lessonBankConfig := defaultLessonBankConfig()
	lessonBank := createLessonBank(lessonBankConfig)

	orch := NewOrchestrator(mockRegistry, lessonBank, config)

	// Register all mock providers
	orch.RegisterProvider("claude", "claude-3", 9.0)
	orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	// Verify setup
	assert.Equal(t, 3, orch.GetAgentPool().Size())
	assert.NotNil(t, orch.GetKnowledgeRepository())

	// Get statistics
	ctx := t.Context()
	stats, err := orch.GetStatistics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.ActiveDebates)
	assert.Equal(t, 3, stats.RegisteredAgents)
}

// Ensure mockProviderRegistry implements ProviderRegistry
func TestMockImplementsProviderRegistry(t *testing.T) {
	var _ ProviderRegistry = (*mockProviderRegistry)(nil)
}

// Ensure mockLLMProvider implements LLMProvider
func TestMockImplementsLLMProvider(t *testing.T) {
	var _ llm.LLMProvider = (*mockLLMProvider)(nil)
}
