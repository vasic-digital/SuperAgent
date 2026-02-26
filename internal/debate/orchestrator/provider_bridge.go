// Package orchestrator provides the debate orchestrator that bridges
// the new debate framework with existing HelixAgent services.
package orchestrator

import (
	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/services"
)

// =============================================================================
// Provider Bridge - Adapts services.ProviderRegistry to orchestrator interface
// =============================================================================

// ProviderRegistryBridge adapts the existing services.ProviderRegistry
// to the orchestrator.ProviderRegistry interface.
type ProviderRegistryBridge struct {
	registry *services.ProviderRegistry
}

// NewProviderRegistryBridge creates a new bridge to the services provider registry.
func NewProviderRegistryBridge(registry *services.ProviderRegistry) *ProviderRegistryBridge {
	return &ProviderRegistryBridge{registry: registry}
}

// GetProvider implements ProviderRegistry interface.
func (b *ProviderRegistryBridge) GetProvider(name string) (llm.LLMProvider, error) {
	return b.registry.GetProvider(name)
}

// GetAvailableProviders implements ProviderRegistry interface.
func (b *ProviderRegistryBridge) GetAvailableProviders() []string {
	return b.registry.ListProviders()
}

// GetProvidersByScore returns providers ordered by LLMsVerifier score.
func (b *ProviderRegistryBridge) GetProvidersByScore() []string {
	return b.registry.ListProvidersOrderedByScore()
}

// GetProviderScore returns the LLMsVerifier score for a provider.
func (b *ProviderRegistryBridge) GetProviderScore(name string) float64 {
	// Score is stored in the verification result, not config
	health := b.registry.GetProviderHealth(name)
	if health != nil && health.Score > 0 {
		return health.Score
	}

	// Fallback to config weight
	config, err := b.registry.GetProviderConfig(name)
	if err != nil || config == nil {
		return 0.0
	}
	return config.Weight
}

// GetProviderModels returns available model names for a provider.
func (b *ProviderRegistryBridge) GetProviderModels(name string) []string {
	config, err := b.registry.GetProviderConfig(name)
	if err != nil || config == nil {
		return nil
	}

	// Extract model names from ModelConfig slice
	models := make([]string, 0, len(config.Models))
	for _, m := range config.Models {
		if m.Enabled && m.Name != "" {
			models = append(models, m.Name)
		} else if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	return models
}

// IsProviderHealthy checks if a provider is healthy.
func (b *ProviderRegistryBridge) IsProviderHealthy(name string) bool {
	health := b.registry.GetProviderHealth(name)
	if health == nil {
		return false
	}
	return health.Verified && health.Status == services.ProviderStatusHealthy
}

// GetAllProviderScores returns a map of provider names to their scores.
func (b *ProviderRegistryBridge) GetAllProviderScores() map[string]float64 {
	scores := make(map[string]float64)
	for _, name := range b.registry.ListProviders() {
		scores[name] = b.GetProviderScore(name)
	}
	return scores
}

// Ensure ProviderRegistryBridge implements ProviderRegistry
var _ ProviderRegistry = (*ProviderRegistryBridge)(nil)

// =============================================================================
// Orchestrator Factory - Creates orchestrator with existing services
// =============================================================================

// OrchestratorFactory creates orchestrators integrated with existing services.
type OrchestratorFactory struct {
	providerRegistry *services.ProviderRegistry
}

// NewOrchestratorFactory creates a new orchestrator factory.
func NewOrchestratorFactory(providerRegistry *services.ProviderRegistry) *OrchestratorFactory {
	return &OrchestratorFactory{
		providerRegistry: providerRegistry,
	}
}

// CreateOrchestrator creates a new orchestrator integrated with services.
// lessonBank can be nil for fresh debates without learning history.
func (f *OrchestratorFactory) CreateOrchestrator(config OrchestratorConfig) *Orchestrator {
	bridge := NewProviderRegistryBridge(f.providerRegistry)

	// Create a fresh lesson bank if none provided
	// In production, this would be retrieved from a persistent store
	lessonBankConfig := defaultLessonBankConfig()
	lessonBank := createLessonBank(lessonBankConfig)

	orch := NewOrchestrator(bridge, lessonBank, config)

	// Auto-register verified providers
	f.registerVerifiedProviders(orch)

	return orch
}

// CreateOrchestratorWithDefaults creates an orchestrator with default config.
func (f *OrchestratorFactory) CreateOrchestratorWithDefaults() *Orchestrator {
	return f.CreateOrchestrator(DefaultOrchestratorConfig())
}

// registerVerifiedProviders registers all available providers from the registry.
// Providers are registered even if not yet verified, but unhealthy ones are skipped.
// This ensures the debate system has agents available immediately on startup.
func (f *OrchestratorFactory) registerVerifiedProviders(orch *Orchestrator) {
	if f.providerRegistry == nil {
		return
	}

	providers := f.providerRegistry.ListProvidersOrderedByScore()

	for _, name := range providers {
		config, err := f.providerRegistry.GetProviderConfig(name)
		if err != nil || config == nil {
			continue
		}

		// Check health status - only skip if explicitly unhealthy (not just unverified)
		health := f.providerRegistry.GetProviderHealth(name)
		if health != nil && health.Status == services.ProviderStatusUnhealthy {
			continue // Skip explicitly unhealthy providers
		}

		// Get score from health verification, or use default/config weight
		score := 5.0 // Default minimum score
		if health != nil && health.Score > 0 {
			score = health.Score
		} else if config.Weight > 0 {
			score = config.Weight
		}

		// Register each enabled model for this provider
		for _, model := range config.Models {
			if !model.Enabled {
				continue
			}

			modelName := model.Name
			if modelName == "" {
				modelName = model.ID
			}

			if modelName != "" {
				_ = orch.RegisterProvider(name, modelName, score) //nolint:errcheck
			}
		}
	}
}

// =============================================================================
// Helper functions for lesson bank integration
// =============================================================================

// defaultLessonBankConfig returns a default lesson bank configuration.
func defaultLessonBankConfig() debate.LessonBankConfig {
	config := debate.DefaultLessonBankConfig()
	config.EnableSemanticSearch = false // Disabled by default until embeddings configured
	return config
}

// createLessonBank creates a new lesson bank with the given config.
func createLessonBank(config debate.LessonBankConfig) *debate.LessonBank {
	return debate.NewLessonBank(config, nil, nil)
}
