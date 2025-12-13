// Package toolkit provides generic interfaces for AI providers and agents.
package toolkit

import (
	"context"
)

// Provider defines the interface that all AI providers must implement.
type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// Chat performs a chat completion request.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// Embed performs an embedding request.
	Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResponse, error)

	// Rerank performs a rerank request.
	Rerank(ctx context.Context, req RerankRequest) (RerankResponse, error)

	// DiscoverModels discovers available models from the provider.
	DiscoverModels(ctx context.Context) ([]ModelInfo, error)

	// ValidateConfig validates the provider configuration.
	ValidateConfig(config map[string]interface{}) error
}

// ModelDiscovery defines the interface for model discovery services.
type ModelDiscovery interface {
	// Discover discovers available models.
	Discover(ctx context.Context) ([]ModelInfo, error)
}

// ConfigBuilder defines the interface for building configurations.
type ConfigBuilder interface {
	// Build builds a configuration from a map.
	Build(config map[string]interface{}) (interface{}, error)

	// Validate validates a configuration.
	Validate(config interface{}) error

	// Merge merges two configurations.
	Merge(base, override interface{}) (interface{}, error)
}

// Agent defines the interface that all coding agents must implement.
type Agent interface {
	// Name returns the name of the agent.
	Name() string

	// Execute executes a task with the given configuration.
	Execute(ctx context.Context, task string, config interface{}) (string, error)

	// ValidateConfig validates the agent configuration.
	ValidateConfig(config interface{}) error

	// Capabilities returns the capabilities of the agent.
	Capabilities() []string
}

// TaskExecutor defines the interface for task execution.
type TaskExecutor interface {
	// ExecuteTask executes a specific task.
	ExecuteTask(ctx context.Context, task interface{}) (interface{}, error)
}

// ConfigValidator defines the interface for configuration validation.
type ConfigValidator interface {
	// Validate validates a configuration.
	Validate(config interface{}) error
}

// ModelManager defines the interface for managing models.
type ModelManager interface {
	// GetModel retrieves model information by ID.
	GetModel(id string) (ModelInfo, error)

	// ListModels lists all available models.
	ListModels() []ModelInfo

	// RegisterModel registers a new model.
	RegisterModel(model ModelInfo) error

	// UpdateModel updates an existing model.
	UpdateModel(model ModelInfo) error
}
