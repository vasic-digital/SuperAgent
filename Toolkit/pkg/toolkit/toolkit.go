// Package toolkit provides the main entry point for the AI toolkit library.
// It initializes all built-in providers and agents, and provides convenience
// functions for common operations.
package toolkit

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Toolkit represents the main toolkit instance that manages providers and agents.
type Toolkit struct {
	providerRegistry        *ProviderRegistry
	agentRegistry           *AgentRegistry
	providerFactoryRegistry *ProviderFactoryRegistry
	agentFactoryRegistry    *AgentFactoryRegistry
	configBuilder           *GenericConfigBuilder
	modelRegistry           *ModelRegistry
	logger                  *log.Logger
	mu                      sync.RWMutex
}

// NewToolkit creates a new toolkit instance with all built-in providers and agents registered.
func NewToolkit() *Toolkit {
	tk := &Toolkit{
		providerRegistry:        NewProviderRegistry(),
		agentRegistry:           NewAgentRegistry(),
		providerFactoryRegistry: NewProviderFactoryRegistry(),
		agentFactoryRegistry:    NewAgentFactoryRegistry(),
		configBuilder:           NewGenericConfigBuilder(),
		modelRegistry:           NewModelRegistry(),
		logger:                  log.Default(),
	}

	// Auto-register built-in providers
	tk.registerBuiltInProviders()

	// Auto-register built-in agents
	tk.registerBuiltInAgents()

	// Register default config builders
	tk.registerConfigBuilders()

	return tk
}

// SetLogger sets the logger for the toolkit.
func (tk *Toolkit) SetLogger(logger *log.Logger) {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.logger = logger
}

// GetProvider returns a provider by name.
func (tk *Toolkit) GetProvider(name string) (Provider, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	provider, exists := tk.providerRegistry.Get(name)
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetAgent returns an agent by name.
func (tk *Toolkit) GetAgent(name string) (Agent, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	agent, exists := tk.agentRegistry.Get(name)
	if !exists {
		return nil, fmt.Errorf("agent %s not found", name)
	}
	return agent, nil
}

// CreateProvider creates a provider using the factory registry.
func (tk *Toolkit) CreateProvider(name string, config map[string]interface{}) (Provider, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.providerFactoryRegistry.Create(name, config)
}

// CreateAgent creates an agent using the factory registry.
func (tk *Toolkit) CreateAgent(name string, config map[string]interface{}) (Agent, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.agentFactoryRegistry.Create(name, config)
}

// ListProviders returns a list of all registered provider factory names.
func (tk *Toolkit) ListProviders() []string {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.providerFactoryRegistry.List()
}

// ListAgents returns a list of all registered agent factory names.
func (tk *Toolkit) ListAgents() []string {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.agentFactoryRegistry.List()
}

// GetProviderFactoryRegistry returns the provider factory registry.
func (tk *Toolkit) GetProviderFactoryRegistry() *ProviderFactoryRegistry {
	return tk.providerFactoryRegistry
}

// GetAgentFactoryRegistry returns the agent factory registry.
func (tk *Toolkit) GetAgentFactoryRegistry() *AgentFactoryRegistry {
	return tk.agentFactoryRegistry
}

// RegisterProvider registers a custom provider.
func (tk *Toolkit) RegisterProvider(name string, provider Provider) error {
	tk.mu.Lock()
	defer tk.mu.Unlock()

	return tk.providerRegistry.Register(name, provider)
}

// RegisterAgent registers a custom agent.
func (tk *Toolkit) RegisterAgent(name string, agent Agent) error {
	tk.mu.Lock()
	defer tk.mu.Unlock()

	return tk.agentRegistry.Register(name, agent)
}

// RegisterProviderFactory registers a provider factory.
func (tk *Toolkit) RegisterProviderFactory(name string, factory ProviderFactory) error {
	tk.mu.Lock()
	defer tk.mu.Unlock()

	return tk.providerFactoryRegistry.Register(name, factory)
}

// RegisterAgentFactory registers an agent factory.
func (tk *Toolkit) RegisterAgentFactory(name string, factory AgentFactory) error {
	tk.mu.Lock()
	defer tk.mu.Unlock()

	return tk.agentFactoryRegistry.Register(name, factory)
}

// BuildConfig builds a configuration for the specified type.
func (tk *Toolkit) BuildConfig(configType string, config map[string]interface{}) (interface{}, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.configBuilder.Build(configType, config)
}

// ExecuteTask executes a task using the specified agent.
func (tk *Toolkit) ExecuteTask(ctx context.Context, agentName, task string, config interface{}) (string, error) {
	agent, err := tk.GetAgent(agentName)
	if err != nil {
		return "", fmt.Errorf("failed to get agent: %w", err)
	}

	tk.logger.Printf("Executing task with agent %s: %s", agentName, task)
	result, err := agent.Execute(ctx, task, config)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
	}

	return result, nil
}

// DiscoverModels discovers models from all registered providers.
func (tk *Toolkit) DiscoverModels(ctx context.Context) ([]ModelInfo, error) {
	tk.mu.RLock()
	providers := tk.providerRegistry.GetAll()
	tk.mu.RUnlock()

	var allModels []ModelInfo
	for name, provider := range providers {
		tk.logger.Printf("Discovering models from provider: %s", name)
		models, err := provider.DiscoverModels(ctx)
		if err != nil {
			tk.logger.Printf("Failed to discover models from provider %s: %v", name, err)
			continue
		}
		allModels = append(allModels, models...)
	}

	return allModels, nil
}

// GetModel returns model information by ID.
func (tk *Toolkit) GetModel(id string) (ModelInfo, error) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	model, exists := tk.modelRegistry.Get(id)
	if !exists {
		return ModelInfo{}, fmt.Errorf("model %s not found", id)
	}
	return model, nil
}

// ListModels returns all registered models.
func (tk *Toolkit) ListModels() []ModelInfo {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	return tk.modelRegistry.List()
}

// ValidateProviderConfig validates a provider configuration.
func (tk *Toolkit) ValidateProviderConfig(providerName string, config map[string]interface{}) error {
	provider, err := tk.CreateProvider(providerName, config)
	if err != nil {
		return fmt.Errorf("failed to create provider for validation: %w", err)
	}

	return provider.ValidateConfig(config)
}

// ValidateAgentConfig validates an agent configuration.
func (tk *Toolkit) ValidateAgentConfig(agentName string, config interface{}) error {
	agent, err := tk.GetAgent(agentName)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	return agent.ValidateConfig(config)
}

// ChatCompletion performs a chat completion using the specified provider.
func (tk *Toolkit) ChatCompletion(ctx context.Context, providerName string, req ChatRequest) (ChatResponse, error) {
	provider, err := tk.GetProvider(providerName)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to get provider: %w", err)
	}

	tk.logger.Printf("Performing chat completion with provider %s, model %s", providerName, req.Model)
	return provider.Chat(ctx, req)
}

// CreateEmbeddings creates embeddings using the specified provider.
func (tk *Toolkit) CreateEmbeddings(ctx context.Context, providerName string, req EmbeddingRequest) (EmbeddingResponse, error) {
	provider, err := tk.GetProvider(providerName)
	if err != nil {
		return EmbeddingResponse{}, fmt.Errorf("failed to get provider: %w", err)
	}

	tk.logger.Printf("Creating embeddings with provider %s, model %s", providerName, req.Model)
	return provider.Embed(ctx, req)
}

// CreateRerank performs reranking using the specified provider.
func (tk *Toolkit) CreateRerank(ctx context.Context, providerName string, req RerankRequest) (RerankResponse, error) {
	provider, err := tk.GetProvider(providerName)
	if err != nil {
		return RerankResponse{}, fmt.Errorf("failed to get provider: %w", err)
	}

	tk.logger.Printf("Performing rerank with provider %s, model %s", providerName, req.Model)
	return provider.Rerank(ctx, req)
}

// registerBuiltInProviders registers all built-in providers.
// Note: Built-in providers register themselves via init functions in their respective packages.
func (tk *Toolkit) registerBuiltInProviders() {
	// Built-in providers will register themselves when their packages are imported
	tk.logger.Println("Built-in providers will register via package init functions")
}

// registerBuiltInAgents registers all built-in agents.
// Note: Built-in agents register themselves via init functions in their respective packages.
func (tk *Toolkit) registerBuiltInAgents() {
	// Built-in agents will register themselves when their packages are imported
	tk.logger.Println("Built-in agents will register via package init functions")
}

// registerConfigBuilders registers default configuration builders.
func (tk *Toolkit) registerConfigBuilders() {
	tk.configBuilder.Register("agent", func(config map[string]interface{}) (interface{}, error) {
		builder := NewDefaultAgentConfigBuilder()
		return builder.Build(config)
	})

	tk.configBuilder.Register("provider", func(config map[string]interface{}) (interface{}, error) {
		builder := NewProviderConfigBuilder()
		return builder.Build(config)
	})
}

// registerProviderFactory is a helper to register provider factories.
func (tk *Toolkit) registerProviderFactory(name string, factory ProviderFactory) {
	if err := tk.providerFactoryRegistry.Register(name, factory); err != nil {
		tk.logger.Printf("Failed to register provider factory %s: %v", name, err)
	}
}

// registerAgentFactory is a helper to register agent factories.
func (tk *Toolkit) registerAgentFactory(name string, factory AgentFactory) {
	if err := tk.agentFactoryRegistry.Register(name, factory); err != nil {
		tk.logger.Printf("Failed to register agent factory %s: %v", name, err)
	}
}

// Global toolkit instance for convenience
var globalToolkit *Toolkit
var globalOnce sync.Once

// GetGlobalToolkit returns the global toolkit instance.
func GetGlobalToolkit() *Toolkit {
	globalOnce.Do(func() {
		globalToolkit = NewToolkit()
	})
	return globalToolkit
}
