// Package claude provides configuration builders for Claude.
package claude

import (
	"fmt"
)

// ConfigBuilder implements the ConfigBuilder interface for Claude.
type ConfigBuilder struct{}

// NewConfigBuilder creates a new Claude config builder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// Build builds a Claude configuration from a map.
func (b *ConfigBuilder) Build(config map[string]interface{}) (interface{}, error) {
	claudeConfig := &Config{
		APIKey:    getString(config, "api_key", ""),
		BaseURL:   getString(config, "base_url", "https://api.anthropic.com"),
		Version:   getString(config, "version", "2023-06-01"),
		Timeout:   getInt(config, "timeout", 30000),
		Retries:   getInt(config, "retries", 3),
		RateLimit: getInt(config, "rate_limit", 60),
	}

	if claudeConfig.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return claudeConfig, nil
}

// Validate validates a Claude configuration.
func (b *ConfigBuilder) Validate(config interface{}) error {
	cConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	if cConfig.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	return nil
}

// Merge merges two Claude configurations.
func (b *ConfigBuilder) Merge(base, override interface{}) (interface{}, error) {
	baseConfig, ok := base.(*Config)
	if !ok {
		return nil, fmt.Errorf("base config must be *Config")
	}

	overrideConfig, ok := override.(*Config)
	if !ok {
		return nil, fmt.Errorf("override config must be *Config")
	}

	merged := &Config{
		APIKey:    overrideConfig.APIKey,
		BaseURL:   overrideConfig.BaseURL,
		Version:   overrideConfig.Version,
		Timeout:   overrideConfig.Timeout,
		Retries:   overrideConfig.Retries,
		RateLimit: overrideConfig.RateLimit,
	}

	if merged.APIKey == "" {
		merged.APIKey = baseConfig.APIKey
	}
	if merged.BaseURL == "" {
		merged.BaseURL = baseConfig.BaseURL
	}
	if merged.Version == "" {
		merged.Version = baseConfig.Version
	}
	if merged.Timeout == 0 {
		merged.Timeout = baseConfig.Timeout
	}
	if merged.Retries == 0 {
		merged.Retries = baseConfig.Retries
	}
	if merged.RateLimit == 0 {
		merged.RateLimit = baseConfig.RateLimit
	}

	return merged, nil
}

// Config represents Claude-specific configuration.
type Config struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Version   string `json:"version"`
	Timeout   int    `json:"timeout"`
	Retries   int    `json:"retries"`
	RateLimit int    `json:"rate_limit"`
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
