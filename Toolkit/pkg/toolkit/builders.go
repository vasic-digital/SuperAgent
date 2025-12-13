// Package toolkit provides configuration builders for agents and providers.
package toolkit

import (
	"fmt"
)

// ConfigBuilderFunc defines a function type for building configurations.
type ConfigBuilderFunc func(config map[string]interface{}) (interface{}, error)

// GenericConfigBuilder provides a generic way to build configurations for different agent types.
type GenericConfigBuilder struct {
	builders map[string]ConfigBuilderFunc
}

// NewGenericConfigBuilder creates a new GenericConfigBuilder.
func NewGenericConfigBuilder() *GenericConfigBuilder {
	return &GenericConfigBuilder{
		builders: make(map[string]ConfigBuilderFunc),
	}
}

// Register registers a builder function for a specific agent type.
func (b *GenericConfigBuilder) Register(agentType string, builder ConfigBuilderFunc) {
	b.builders[agentType] = builder
}

// Build builds a configuration for the specified agent type.
func (b *GenericConfigBuilder) Build(agentType string, config map[string]interface{}) (interface{}, error) {
	builder, ok := b.builders[agentType]
	if !ok {
		return nil, fmt.Errorf("no builder registered for agent type %s", agentType)
	}
	return builder(config)
}

// Validate validates a configuration.
func (b *GenericConfigBuilder) Validate(agentType string, config interface{}) error {
	// Basic validation - can be extended
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// Merge merges two configurations with override taking precedence.
func (b *GenericConfigBuilder) Merge(base, override interface{}) (interface{}, error) {
	if base == nil {
		return override, nil
	}
	if override == nil {
		return base, nil
	}

	baseMap, ok := base.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("base config must be a map")
	}

	overrideMap, ok := override.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("override config must be a map")
	}

	merged := make(map[string]interface{})

	// Copy base
	for k, v := range baseMap {
		merged[k] = v
	}

	// Override with override values
	for k, v := range overrideMap {
		if subMap, ok := v.(map[string]interface{}); ok {
			if baseSubMap, baseOk := merged[k].(map[string]interface{}); baseOk {
				// Recursively merge nested maps
				mergedSubMap, err := b.Merge(baseSubMap, subMap)
				if err != nil {
					return nil, err
				}
				merged[k] = mergedSubMap
			} else {
				merged[k] = subMap
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

// DefaultAgentConfigBuilder provides a default builder for common agent configurations.
type DefaultAgentConfigBuilder struct{}

// NewDefaultAgentConfigBuilder creates a new DefaultAgentConfigBuilder.
func NewDefaultAgentConfigBuilder() *DefaultAgentConfigBuilder {
	return &DefaultAgentConfigBuilder{}
}

// Build builds a default agent configuration.
func (b *DefaultAgentConfigBuilder) Build(config map[string]interface{}) (interface{}, error) {
	// Extract common fields
	agentConfig := &AgentConfig{
		Name:        getString(config, "name", "default-agent"),
		Description: getString(config, "description", ""),
		Version:     getString(config, "version", "1.0.0"),
		Provider:    getString(config, "provider", ""),
		Model:       getString(config, "model", ""),
		MaxTokens:   getInt(config, "max_tokens", 4096),
		Temperature: getFloat64(config, "temperature", 0.7),
		Timeout:     getInt(config, "timeout", 30000), // 30 seconds
		Retries:     getInt(config, "retries", 3),
	}

	// Validate required fields
	if agentConfig.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if agentConfig.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	return agentConfig, nil
}

// AgentConfig represents a common agent configuration.
type AgentConfig struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"` // milliseconds
	Retries     int     `json:"retries"`
}

// Validate validates the agent configuration.
func (c *AgentConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if c.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	return nil
}

// Helper functions for type-safe config extraction
func getString(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

func getFloat64(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultValue
}

// ProviderConfigBuilder provides a builder for provider configurations.
type ProviderConfigBuilder struct{}

// NewProviderConfigBuilder creates a new ProviderConfigBuilder.
func NewProviderConfigBuilder() *ProviderConfigBuilder {
	return &ProviderConfigBuilder{}
}

// Build builds a provider configuration.
func (b *ProviderConfigBuilder) Build(config map[string]interface{}) (interface{}, error) {
	providerConfig := &ProviderConfig{
		Name:      getString(config, "name", ""),
		APIKey:    getString(config, "api_key", ""),
		BaseURL:   getString(config, "base_url", ""),
		Timeout:   getInt(config, "timeout", 30000),
		Retries:   getInt(config, "retries", 3),
		RateLimit: getInt(config, "rate_limit", 60), // requests per minute
	}

	if providerConfig.Name == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	if providerConfig.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return providerConfig, nil
}

// ProviderConfig represents a common provider configuration.
type ProviderConfig struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Timeout   int    `json:"timeout"`
	Retries   int    `json:"retries"`
	RateLimit int    `json:"rate_limit"`
}

// Validate validates the provider configuration.
func (c *ProviderConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	return nil
}
