// Package openrouter provides a OpenRouter provider implementation.
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Client represents a OpenRouter API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new OpenRouter API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatCompletion performs a chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req toolkit.ChatRequest) (toolkit.ChatResponse, error) {
	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}

	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		payload["top_p"] = req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop"] = req.Stop
	}

	var response toolkit.ChatResponse
	err := c.doRequest(ctx, "POST", "/chat/completions", payload, &response)
	return response, err
}

// CreateEmbeddings performs an embedding request.
func (c *Client) CreateEmbeddings(ctx context.Context, req toolkit.EmbeddingRequest) (toolkit.EmbeddingResponse, error) {
	payload := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	if req.EncodingFormat != "" {
		payload["encoding_format"] = req.EncodingFormat
	}

	var response toolkit.EmbeddingResponse
	err := c.doRequest(ctx, "POST", "/embeddings", payload, &response)
	return response, err
}

// CreateRerank performs a rerank request.
func (c *Client) CreateRerank(ctx context.Context, req toolkit.RerankRequest) (toolkit.RerankResponse, error) {
	payload := map[string]interface{}{
		"model":            req.Model,
		"query":            req.Query,
		"documents":        req.Documents,
		"top_n":            req.TopN,
		"return_documents": req.ReturnDocs,
	}

	var response toolkit.RerankResponse
	err := c.doRequest(ctx, "POST", "/rerank", payload, &response)
	return response, err
}

// GetModels retrieves available models from the API.
func (c *Client) GetModels(ctx context.Context) ([]ModelInfo, error) {
	var response struct {
		Data []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}

	err := c.doRequest(ctx, "GET", "/models", nil, &response)
	if err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, model := range response.Data {
		models = append(models, ModelInfo{
			ID:   model.ID,
			Type: model.Type,
		})
	}

	return models, nil
}

// ModelInfo represents basic model information from the API.
type ModelInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// doRequest performs an HTTP request to the OpenRouter API.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Provider implements the Provider interface for OpenRouter.
type Provider struct {
	client    *Client
	discovery *Discovery
	config    *Config
}

// NewProvider creates a new OpenRouter provider.
func NewProvider(config map[string]interface{}) (toolkit.Provider, error) {
	builder := NewConfigBuilder()
	cfg, err := builder.Build(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	openRouterConfig, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	return &Provider{
		client:    NewClient(openRouterConfig.APIKey),
		discovery: NewDiscovery(openRouterConfig.APIKey),
		config:    openRouterConfig,
	}, nil
}

// Name returns the name of the provider.
func (p *Provider) Name() string {
	return "openrouter"
}

// Chat performs a chat completion request.
func (p *Provider) Chat(ctx context.Context, req toolkit.ChatRequest) (toolkit.ChatResponse, error) {
	log.Printf("OpenRouter: Performing chat completion with model %s", req.Model)
	return p.client.ChatCompletion(ctx, req)
}

// Embed performs an embedding request.
func (p *Provider) Embed(ctx context.Context, req toolkit.EmbeddingRequest) (toolkit.EmbeddingResponse, error) {
	log.Printf("OpenRouter: Performing embedding with model %s", req.Model)
	return p.client.CreateEmbeddings(ctx, req)
}

// Rerank performs a rerank request.
func (p *Provider) Rerank(ctx context.Context, req toolkit.RerankRequest) (toolkit.RerankResponse, error) {
	log.Printf("OpenRouter: Performing rerank with model %s", req.Model)
	return p.client.CreateRerank(ctx, req)
}

// DiscoverModels discovers available models from the provider.
func (p *Provider) DiscoverModels(ctx context.Context) ([]toolkit.ModelInfo, error) {
	log.Println("OpenRouter: Discovering models")
	return p.discovery.Discover(ctx)
}

// ValidateConfig validates the provider configuration.
func (p *Provider) ValidateConfig(config map[string]interface{}) error {
	builder := NewConfigBuilder()
	_, err := builder.Build(config)
	return err
}

// Factory function for creating OpenRouter providers.
func Factory(config map[string]interface{}) (toolkit.Provider, error) {
	return NewProvider(config)
}

// Register registers the OpenRouter provider with the registry.
func Register(registry *toolkit.ProviderFactoryRegistry) error {
	return registry.Register("openrouter", Factory)
}

// Discovery implements the ModelDiscovery interface for OpenRouter.
type Discovery struct {
	client *Client
}

// NewDiscovery creates a new OpenRouter model discovery instance.
func NewDiscovery(apiKey string) *Discovery {
	return &Discovery{
		client: NewClient(apiKey),
	}
}

// Discover discovers available models from OpenRouter.
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

// convertToModelInfo converts OpenRouter model info to toolkit ModelInfo.
func (d *Discovery) convertToModelInfo(model ModelInfo) toolkit.ModelInfo {
	capabilities := d.inferCapabilities(model.ID, model.Type)

	return toolkit.ModelInfo{
		ID:           model.ID,
		Name:         d.formatModelName(model.ID),
		Category:     d.inferCategory(model.ID, model.Type),
		Capabilities: capabilities,
		Provider:     "openrouter",
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

	// Chat capabilities - OpenRouter has many models
	specializedKeywords := []string{
		"embedding", "rerank", "tts", "speech", "audio", "video",
		"t2v", "i2v", "vl", "vision", "visual", "multimodal", "image",
	}

	isSpecialized := d.containsAny(modelLower, specializedKeywords)
	chatTypeIndicators := []string{"chat", "text", "completion", "instruction", "instruct"}
	hasChatType := d.containsAny(typeLower, chatTypeIndicators)
	chatIDIndicators := []string{
		"instruct", "chat", "gpt", "claude", "llama", "mistral", "mixtral", "gemma",
		"qwen", "deepseek", "glm", "kimi", "phi", "yi",
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

	if strings.Contains(modelLower, "anthropic") || strings.Contains(modelLower, "claude") {
		return "Anthropic Claude models via OpenRouter"
	}
	if strings.Contains(modelLower, "openai") || strings.Contains(modelLower, "gpt") {
		return "OpenAI GPT models via OpenRouter"
	}
	if strings.Contains(modelLower, "google") || strings.Contains(modelLower, "gemini") {
		return "Google Gemini models via OpenRouter"
	}
	if strings.Contains(modelLower, "meta") || strings.Contains(modelLower, "llama") {
		return "Meta Llama models via OpenRouter"
	}

	return "Model available via OpenRouter"
}

// supportsFunctionCalling checks if model supports function calling.
func (d *Discovery) supportsFunctionCalling(modelID string) bool {
	modelLower := strings.ToLower(modelID)

	supportedModels := []string{"gpt", "claude", "llama", "mistral", "gemma", "qwen"}
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

	if strings.Contains(modelLower, "claude-3.5-sonnet") {
		return 200000
	}
	if strings.Contains(modelLower, "claude-3") {
		return 200000
	}
	if strings.Contains(modelLower, "gpt-4o") {
		return 128000
	}
	if strings.Contains(modelLower, "gpt-4") {
		return 128000
	}
	if strings.Contains(modelLower, "llama-3.1-405b") {
		return 131072
	}
	if strings.Contains(modelLower, "llama-3.1") {
		return 131072
	}

	return 4096
}

// inferMaxTokens infers maximum tokens for output.
func (d *Discovery) inferMaxTokens(modelID string) int {
	modelLower := strings.ToLower(modelID)

	if strings.Contains(modelLower, "claude-3.5-sonnet") {
		return 8192
	}
	if strings.Contains(modelLower, "gpt-4o") {
		return 4096
	}
	if strings.Contains(modelLower, "llama-3.1-405b") {
		return 4096
	}

	return 4096
}

// ConfigBuilder implements the ConfigBuilder interface for OpenRouter.
type ConfigBuilder struct{}

// NewConfigBuilder creates a new OpenRouter config builder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// Build builds a OpenRouter configuration from a map.
func (b *ConfigBuilder) Build(config map[string]interface{}) (interface{}, error) {
	openRouterConfig := &Config{
		APIKey:    getString(config, "api_key", ""),
		BaseURL:   getString(config, "base_url", "https://openrouter.ai/api/v1"),
		Timeout:   getInt(config, "timeout", 30000),
		Retries:   getInt(config, "retries", 3),
		RateLimit: getInt(config, "rate_limit", 60),
	}

	if openRouterConfig.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return openRouterConfig, nil
}

// Validate validates a OpenRouter configuration.
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

// Merge merges two OpenRouter configurations.
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

// Config represents OpenRouter-specific configuration.
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
