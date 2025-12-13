// Package nvidia provides a Nvidia provider implementation.
package nvidia

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Provider implements the Provider interface for Nvidia.
type Provider struct {
	client    *Client
	discovery *Discovery
	config    *Config
}

// NewProvider creates a new Nvidia provider.
func NewProvider(config map[string]interface{}) (toolkit.Provider, error) {
	builder := NewConfigBuilder()
	cfg, err := builder.Build(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	nvidiaConfig, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	return &Provider{
		client:    NewClient(nvidiaConfig.APIKey),
		discovery: NewDiscovery(nvidiaConfig.APIKey),
		config:    nvidiaConfig,
	}, nil
}

// Name returns the name of the provider.
func (p *Provider) Name() string {
	return "nvidia"
}

// Chat performs a chat completion request.
func (p *Provider) Chat(ctx context.Context, req toolkit.ChatRequest) (toolkit.ChatResponse, error) {
	log.Printf("Nvidia: Performing chat completion with model %s", req.Model)
	return p.client.ChatCompletion(ctx, req)
}

// Embed performs an embedding request.
func (p *Provider) Embed(ctx context.Context, req toolkit.EmbeddingRequest) (toolkit.EmbeddingResponse, error) {
	log.Printf("Nvidia: Performing embedding with model %s", req.Model)
	return p.client.CreateEmbeddings(ctx, req)
}

// Rerank performs a rerank request.
func (p *Provider) Rerank(ctx context.Context, req toolkit.RerankRequest) (toolkit.RerankResponse, error) {
	log.Printf("Nvidia: Performing rerank with model %s", req.Model)
	return p.client.CreateRerank(ctx, req)
}

// DiscoverModels discovers available models from the provider.
func (p *Provider) DiscoverModels(ctx context.Context) ([]toolkit.ModelInfo, error) {
	log.Println("Nvidia: Discovering models")
	return p.discovery.Discover(ctx)
}

// ValidateConfig validates the provider configuration.
func (p *Provider) ValidateConfig(config map[string]interface{}) error {
	builder := NewConfigBuilder()
	_, err := builder.Build(config)
	return err
}

// Factory function for creating Nvidia providers.
func Factory(config map[string]interface{}) (toolkit.Provider, error) {
	return NewProvider(config)
}

// Register registers the Nvidia provider with the registry.
func Register(registry *toolkit.ProviderFactoryRegistry) error {
	return registry.Register("nvidia", Factory)
}

// Discovery implements the ModelDiscovery interface for Nvidia.
type Discovery struct {
	client *Client
}

// NewDiscovery creates a new Nvidia model discovery instance.
func NewDiscovery(apiKey string) *Discovery {
	return &Discovery{
		client: NewClient(apiKey),
	}
}

// Discover discovers available models from Nvidia.
func (d *Discovery) Discover(ctx context.Context) ([]toolkit.ModelInfo, error) {
	models, err := d.client.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	var modelInfos []toolkit.ModelInfo
	for _, model := range models {
		modelInfo := d.convertToModelInfo(model)
		modelInfos = append(modelInfos, modelInfo)
	}

	return modelInfos, nil
}

// convertToModelInfo converts Nvidia model info to toolkit ModelInfo.
func (d *Discovery) convertToModelInfo(model ModelInfo) toolkit.ModelInfo {
	capabilities := d.inferCapabilities(model.ID, model.Type)

	return toolkit.ModelInfo{
		ID:           model.ID,
		Name:         d.formatModelName(model.ID),
		Category:     d.inferCategory(model.ID, model.Type),
		Capabilities: capabilities,
		Provider:     "nvidia",
		Description:  d.getModelDescription(model.ID),
	}
}

// inferCapabilities infers model capabilities from ID and type.
func (d *Discovery) inferCapabilities(modelID, modelType string) toolkit.ModelCapabilities {
	capabilities := toolkit.ModelCapabilities{}

	modelLower := strings.ToLower(modelID)
	typeLower := strings.ToLower(modelType)

	// Embedding capabilities
	capabilities.SupportsEmbedding = strings.Contains(typeLower, "embedding")

	// Rerank capabilities
	capabilities.SupportsRerank = strings.Contains(typeLower, "rerank")

	// Audio capabilities
	audioKeywords := []string{"tts", "audio", "speech", "voice"}
	capabilities.SupportsAudio = d.containsAny(modelLower, audioKeywords) || d.containsAny(typeLower, audioKeywords)

	// Vision capabilities
	visionKeywords := []string{"vl", "vision", "visual", "multimodal"}
	capabilities.SupportsVision = d.containsAny(modelLower, visionKeywords) || d.containsAny(typeLower, visionKeywords)

	// Chat capabilities - Nvidia specific models
	specializedKeywords := []string{
		"embedding", "rerank", "tts", "speech", "audio", "video",
		"t2v", "i2v", "vl", "vision", "visual", "multimodal", "image",
	}

	isSpecialized := d.containsAny(modelLower, specializedKeywords)
	chatTypeIndicators := []string{"chat", "text", "completion", "instruction", "instruct"}
	hasChatType := d.containsAny(typeLower, chatTypeIndicators)
	chatIDIndicators := []string{
		"instruct", "chat", "llama", "mistral", "mixtral", "gemma",
	}
	hasChatID := d.containsAny(modelLower, chatIDIndicators)

	if hasChatType && !isSpecialized {
		capabilities.SupportsChat = true
	} else if hasChatID && !isSpecialized {
		capabilities.SupportsChat = true
	} else if typeLower == "" || typeLower == "unknown" {
		capabilities.SupportsChat = false
	} else {
		capabilities.SupportsChat = false
	}

	// Function calling support
	capabilities.FunctionCalling = d.supportsFunctionCalling(modelID)

	// Context window and max tokens
	capabilities.ContextWindow = d.inferContextWindow(modelID)
	capabilities.MaxTokens = d.inferMaxTokens(modelID)

	return capabilities
}

// containsAny checks if the string contains any of the keywords.
func (d *Discovery) containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

// formatModelName formats model ID into human-readable name.
func (d *Discovery) formatModelName(modelID string) string {
	parts := strings.Split(modelID, "/")
	if len(parts) > 1 {
		name := strings.ReplaceAll(parts[1], "-", " ")
		name = strings.ReplaceAll(name, "_", " ")
		return parts[0] + " " + name
	}
	return strings.ReplaceAll(modelID, "-", " ")
}

// inferCategory infers the model category.
func (d *Discovery) inferCategory(modelID, modelType string) toolkit.ModelCategory {
	typeLower := strings.ToLower(modelType)
	modelLower := strings.ToLower(modelID)

	if strings.Contains(typeLower, "embedding") {
		return toolkit.CategoryEmbedding
	}
	if strings.Contains(typeLower, "rerank") {
		return toolkit.CategoryRerank
	}
	if strings.Contains(modelLower, "vl") || strings.Contains(modelLower, "vision") ||
		strings.Contains(modelLower, "multimodal") {
		return toolkit.CategoryMultimodal
	}
	if strings.Contains(modelLower, "image") {
		return toolkit.CategoryImage
	}

	return toolkit.CategoryChat
}

// getModelDescription returns a description for the model.
func (d *Discovery) getModelDescription(modelID string) string {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "llama") {
		return "Meta Llama models hosted on Nvidia"
	}
	if strings.Contains(modelLower, "mistral") {
		return "Mistral models hosted on Nvidia"
	}
	if strings.Contains(modelLower, "gemma") {
		return "Google Gemma models hosted on Nvidia"
	}

	return "Nvidia hosted model"
}

// supportsFunctionCalling checks if model supports function calling.
func (d *Discovery) supportsFunctionCalling(modelID string) bool {
	modelLower := strings.ToLower(modelID)

	supportedModels := []string{"llama", "mistral", "gemma"}
	for _, model := range supportedModels {
		if strings.Contains(modelLower, model) {
			return true
		}
	}

	return false
}

// inferContextWindow infers context window size.
func (d *Discovery) inferContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "llama-3.1-405b") {
		return 131072
	}
	if strings.Contains(modelLower, "llama-3.1") {
		return 131072
	}
	if strings.Contains(modelLower, "llama-3") {
		return 8192
	}
	if strings.Contains(modelLower, "mistral-7b") {
		return 32768
	}

	return 4096
}

// inferMaxTokens infers maximum tokens for output.
func (d *Discovery) inferMaxTokens(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "llama-3.1-405b") {
		return 4096
	}
	if strings.Contains(modelLower, "llama-3.1") {
		return 4096
	}

	return 4096
}

// ConfigBuilder implements the ConfigBuilder interface for Nvidia.
type ConfigBuilder struct{}

// NewConfigBuilder creates a new Nvidia config builder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// Build builds a Nvidia configuration from a map.
func (b *ConfigBuilder) Build(config map[string]interface{}) (interface{}, error) {
	nvidiaConfig := &Config{
		APIKey:    getString(config, "api_key", ""),
		BaseURL:   getString(config, "base_url", "https://integrate.api.nvidia.com/v1"),
		Timeout:   getInt(config, "timeout", 30000),
		Retries:   getInt(config, "retries", 3),
		RateLimit: getInt(config, "rate_limit", 60),
	}

	if nvidiaConfig.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return nvidiaConfig, nil
}

// Validate validates a Nvidia configuration.
func (b *ConfigBuilder) Validate(config interface{}) error {
	c, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	return nil
}

// Merge merges two Nvidia configurations.
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

// Config represents Nvidia-specific configuration.
type Config struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
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
