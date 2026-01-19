package modelsdev

import (
	"time"
)

// Model represents a model from Models.dev API for caching purposes
type Model struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Provider      string             `json:"provider"`
	DisplayName   string             `json:"display_name,omitempty"`
	Description   string             `json:"description,omitempty"`
	ContextWindow int                `json:"context_window,omitempty"`
	MaxTokens     int                `json:"max_tokens,omitempty"`
	Pricing       *Pricing           `json:"pricing,omitempty"`
	Capabilities  *ModelCapabilities `json:"capabilities,omitempty"`
	Performance   *ModelPerformance  `json:"performance,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Categories    []string           `json:"categories,omitempty"`
	Family        string             `json:"family,omitempty"`
	Version       string             `json:"version,omitempty"`
	ReleaseDate   *time.Time         `json:"release_date,omitempty"`
	Deprecated    bool               `json:"deprecated,omitempty"`
	Metadata      map[string]any     `json:"metadata,omitempty"`
}

// Provider represents a provider from Models.dev API for caching purposes
type Provider struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Website     string   `json:"website,omitempty"`
	APIDocsURL  string   `json:"api_docs_url,omitempty"`
	APIBase     string   `json:"api_base,omitempty"`
	ModelsCount int      `json:"models_count,omitempty"`
	Features    []string `json:"features,omitempty"`
	Models      []string `json:"models,omitempty"`
}

// Pricing represents the pricing information for a model
type Pricing struct {
	InputCost       float64 `json:"input_cost"`         // Cost per million input tokens
	OutputCost      float64 `json:"output_cost"`        // Cost per million output tokens
	Currency        string  `json:"currency,omitempty"` // Default: USD
	Unit            string  `json:"unit,omitempty"`     // "tokens", "characters", etc.
	CachedInputCost float64 `json:"cached_input_cost,omitempty"`
}

// ModelFilters represents filters for listing models
type ModelFilters struct {
	Provider     string   `json:"provider,omitempty"`
	Family       string   `json:"family,omitempty"`
	Category     string   `json:"category,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	MinContext   int      `json:"min_context,omitempty"`
	MaxContext   int      `json:"max_context,omitempty"`
	HasVision    *bool    `json:"has_vision,omitempty"`
	HasStreaming *bool    `json:"has_streaming,omitempty"`
	HasTools     *bool    `json:"has_tools,omitempty"`
	Page         int      `json:"page,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`    // "name", "popularity", "benchmark_score"
	SortOrder    string   `json:"sort_order,omitempty"` // "asc", "desc"
}

// CachedModel wraps a Model with cache metadata
type CachedModel struct {
	Model     *Model    `json:"model"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
	HitCount  int64     `json:"hit_count"`
}

// CachedProvider wraps a Provider with cache metadata
type CachedProvider struct {
	Provider  *Provider `json:"provider"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
	HitCount  int64     `json:"hit_count"`
}

// RefreshResult represents the result of a cache refresh operation
type RefreshResult struct {
	ModelsRefreshed    int           `json:"models_refreshed"`
	ProvidersRefreshed int           `json:"providers_refreshed"`
	Errors             []string      `json:"errors,omitempty"`
	Duration           time.Duration `json:"duration"`
	Success            bool          `json:"success"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	ModelCount       int       `json:"model_count"`
	ProviderCount    int       `json:"provider_count"`
	TotalHits        int64     `json:"total_hits"`
	TotalMisses      int64     `json:"total_misses"`
	HitRate          float64   `json:"hit_rate"`
	LastRefresh      time.Time `json:"last_refresh"`
	OldestEntry      time.Time `json:"oldest_entry"`
	MemoryUsageBytes int64     `json:"memory_usage_bytes"`
}

// CacheConfig represents configuration for the cache
type CacheConfig struct {
	ModelTTL        time.Duration `json:"model_ttl"`
	ProviderTTL     time.Duration `json:"provider_ttl"`
	MaxModels       int           `json:"max_models"`
	MaxProviders    int           `json:"max_providers"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// ServiceConfig represents configuration for the ModelsDevService
type ServiceConfig struct {
	Client          ClientConfig  `json:"client"`
	Cache           CacheConfig   `json:"cache"`
	RefreshOnStart  bool          `json:"refresh_on_start"`
	AutoRefresh     bool          `json:"auto_refresh"`
	RefreshInterval time.Duration `json:"refresh_interval"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		ModelTTL:        1 * time.Hour,
		ProviderTTL:     2 * time.Hour,
		MaxModels:       5000,
		MaxProviders:    100,
		CleanupInterval: 10 * time.Minute,
	}
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Client:          DefaultClientConfig(),
		Cache:           DefaultCacheConfig(),
		RefreshOnStart:  true,
		AutoRefresh:     true,
		RefreshInterval: 24 * time.Hour,
	}
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		BaseURL:   DefaultBaseURL,
		Timeout:   DefaultTimeout,
		UserAgent: DefaultUserAgent,
	}
}

// ToCapabilitiesList converts ModelCapabilities to a list of capability strings
func (c *ModelCapabilities) ToCapabilitiesList() []string {
	if c == nil {
		return nil
	}

	caps := make([]string, 0)
	if c.Vision {
		caps = append(caps, "vision")
	}
	if c.FunctionCalling {
		caps = append(caps, "function_calling")
	}
	if c.Streaming {
		caps = append(caps, "streaming")
	}
	if c.JSONMode {
		caps = append(caps, "json_mode")
	}
	if c.ImageGeneration {
		caps = append(caps, "image_generation")
	}
	if c.Audio {
		caps = append(caps, "audio")
	}
	if c.CodeGeneration {
		caps = append(caps, "code_generation")
	}
	if c.Reasoning {
		caps = append(caps, "reasoning")
	}
	if c.ToolUse {
		caps = append(caps, "tool_use")
	}
	return caps
}

// HasCapability checks if the model has a specific capability
func (c *ModelCapabilities) HasCapability(capability string) bool {
	if c == nil {
		return false
	}

	switch capability {
	case "vision":
		return c.Vision
	case "function_calling":
		return c.FunctionCalling
	case "streaming":
		return c.Streaming
	case "json_mode":
		return c.JSONMode
	case "image_generation":
		return c.ImageGeneration
	case "audio":
		return c.Audio
	case "code_generation":
		return c.CodeGeneration
	case "reasoning":
		return c.Reasoning
	case "tool_use":
		return c.ToolUse
	default:
		return false
	}
}

// GetDisplayName returns the display name or falls back to name
func (m *Model) GetDisplayName() string {
	if m.DisplayName != "" {
		return m.DisplayName
	}
	return m.Name
}

// GetDisplayName returns the display name or falls back to name
func (p *Provider) GetDisplayName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}

// IsExpired checks if the cached model has expired
func (cm *CachedModel) IsExpired() bool {
	return time.Now().After(cm.ExpiresAt)
}

// IsExpired checks if the cached provider has expired
func (cp *CachedProvider) IsExpired() bool {
	return time.Now().After(cp.ExpiresAt)
}
