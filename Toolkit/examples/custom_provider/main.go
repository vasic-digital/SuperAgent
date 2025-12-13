// Package main demonstrates how to create and register a custom provider.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// CustomProvider implements a custom AI provider that integrates with a hypothetical API.
type CustomProvider struct {
	name       string
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// CustomConfig represents the configuration for our custom provider.
type CustomConfig struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Timeout   int    `json:"timeout"`
	Retries   int    `json:"retries"`
	RateLimit int    `json:"rate_limit"`
}

// NewCustomProvider creates a new custom provider instance.
func NewCustomProvider(config map[string]interface{}) (toolkit.Provider, error) {
	cfg := &CustomConfig{
		Name:      getString(config, "name", "custom-provider"),
		APIKey:    getString(config, "api_key", ""),
		BaseURL:   getString(config, "base_url", "https://api.custom-provider.com"),
		Timeout:   getInt(config, "timeout", 30000),
		Retries:   getInt(config, "retries", 3),
		RateLimit: getInt(config, "rate_limit", 60),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return &CustomProvider{
		name:    cfg.Name,
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
		},
	}, nil
}

// Name returns the name of the provider.
func (p *CustomProvider) Name() string {
	return p.name
}

// Chat performs a chat completion request.
func (p *CustomProvider) Chat(ctx context.Context, req toolkit.ChatRequest) (toolkit.ChatResponse, error) {
	log.Printf("CustomProvider: Performing chat completion with model %s", req.Model)

	// Simulate API call - in real implementation, this would make HTTP request
	// For demo purposes, we'll return a mock response

	// In a real implementation, you would convert the request to your API format:
	// apiReq := map[string]interface{}{
	//     "model":       req.Model,
	//     "messages":    req.Messages,
	//     "max_tokens":  req.MaxTokens,
	//     "temperature": req.Temperature,
	//     "stream":      req.Stream,
	// }

	// Simulate API call delay
	select {
	case <-ctx.Done():
		return toolkit.ChatResponse{}, ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	// Mock response - in real implementation, parse actual API response
	response := toolkit.ChatResponse{
		ID:      "custom-chat-" + fmt.Sprintf("%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []toolkit.ChatChoice{
			{
				Index: 0,
				Message: toolkit.ChatMessage{
					Role:    "assistant",
					Content: fmt.Sprintf("Hello! This is a response from the custom %s provider. You asked about: %s", p.name, req.Messages[len(req.Messages)-1].Content),
				},
				FinishReason: "stop",
			},
		},
		Usage: toolkit.Usage{
			PromptTokens:     50,
			CompletionTokens: 30,
			TotalTokens:      80,
		},
	}

	log.Printf("CustomProvider: Chat completion successful, used %d tokens", response.Usage.TotalTokens)
	return response, nil
}

// Embed performs an embedding request.
func (p *CustomProvider) Embed(ctx context.Context, req toolkit.EmbeddingRequest) (toolkit.EmbeddingResponse, error) {
	log.Printf("CustomProvider: Performing embedding with model %s", req.Model)

	// Simulate API call
	select {
	case <-ctx.Done():
		return toolkit.EmbeddingResponse{}, ctx.Err()
	case <-time.After(150 * time.Millisecond):
	}

	// Mock embedding response
	response := toolkit.EmbeddingResponse{
		Object: "list",
		Data: []toolkit.EmbeddingData{
			{
				Object:    "embedding",
				Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5}, // Mock 5D embedding
				Index:     0,
			},
		},
		Model: req.Model,
		Usage: toolkit.Usage{
			PromptTokens:     10,
			CompletionTokens: 0,
			TotalTokens:      10,
		},
	}

	return response, nil
}

// Rerank performs a rerank request.
func (p *CustomProvider) Rerank(ctx context.Context, req toolkit.RerankRequest) (toolkit.RerankResponse, error) {
	log.Printf("CustomProvider: Performing rerank with model %s", req.Model)

	// Simulate API call
	select {
	case <-ctx.Done():
		return toolkit.RerankResponse{}, ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	// Mock rerank response
	response := toolkit.RerankResponse{
		Object: "list",
		Model:  req.Model,
		Results: []toolkit.RerankResult{
			{
				Index:    0,
				Score:    0.95,
				Document: req.Documents[0],
			},
			{
				Index:    1,
				Score:    0.87,
				Document: req.Documents[1],
			},
		},
	}

	return response, nil
}

// DiscoverModels discovers available models from the provider.
func (p *CustomProvider) DiscoverModels(ctx context.Context) ([]toolkit.ModelInfo, error) {
	log.Println("CustomProvider: Discovering models")

	// Simulate API call to discover models
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(50 * time.Millisecond):
	}

	// Return mock model information
	models := []toolkit.ModelInfo{
		{
			ID:       "custom-chat-v1",
			Name:     "Custom Chat Model v1",
			Category: toolkit.CategoryChat,
			Capabilities: toolkit.ModelCapabilities{
				SupportsChat:      true,
				SupportsEmbedding: false,
				SupportsRerank:    false,
				MaxTokens:         4096,
				ContextWindow:     4096,
				InputPricing:      0.001,
				OutputPricing:     0.002,
			},
			Provider:    p.name,
			Description: "A custom chat completion model",
		},
		{
			ID:       "custom-embed-v1",
			Name:     "Custom Embedding Model v1",
			Category: toolkit.CategoryEmbedding,
			Capabilities: toolkit.ModelCapabilities{
				SupportsChat:      false,
				SupportsEmbedding: true,
				SupportsRerank:    false,
				MaxTokens:         512,
				ContextWindow:     512,
				InputPricing:      0.0001,
				OutputPricing:     0.0001,
			},
			Provider:    p.name,
			Description: "A custom embedding model",
		},
		{
			ID:       "custom-rerank-v1",
			Name:     "Custom Rerank Model v1",
			Category: toolkit.CategoryRerank,
			Capabilities: toolkit.ModelCapabilities{
				SupportsChat:      false,
				SupportsEmbedding: false,
				SupportsRerank:    true,
				MaxTokens:         512,
				ContextWindow:     512,
				InputPricing:      0.0002,
				OutputPricing:     0.0002,
			},
			Provider:    p.name,
			Description: "A custom reranking model",
		},
	}

	return models, nil
}

// ValidateConfig validates the provider configuration.
func (p *CustomProvider) ValidateConfig(config map[string]interface{}) error {
	cfg := &CustomConfig{
		Name:    getString(config, "name", "custom-provider"),
		APIKey:  getString(config, "api_key", ""),
		BaseURL: getString(config, "base_url", "https://api.custom-provider.com"),
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("base_url cannot be empty")
	}

	return nil
}

// Helper functions
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

func main() {
	fmt.Println("=== Custom Provider Example ===")

	// Initialize the toolkit
	tk := toolkit.NewToolkit()
	tk.SetLogger(log.Default())

	// Create custom provider configuration
	config := map[string]interface{}{
		"name":       "my-custom-provider",
		"api_key":    "custom-api-key-12345",
		"base_url":   "https://api.custom-provider.com/v1",
		"timeout":    30000,
		"retries":    3,
		"rate_limit": 60,
	}

	// Create the custom provider
	customProvider, err := NewCustomProvider(config)
	if err != nil {
		log.Fatalf("Failed to create custom provider: %v", err)
	}

	// Register the custom provider with the toolkit
	if err := tk.RegisterProvider("my-custom-provider", customProvider); err != nil {
		log.Fatalf("Failed to register custom provider: %v", err)
	}

	fmt.Println("✓ Custom provider registered successfully")

	// Test the provider functionality
	ctx := context.Background()

	// Test model discovery
	fmt.Println("\n--- Testing Model Discovery ---")
	models, err := customProvider.DiscoverModels(ctx)
	if err != nil {
		log.Fatalf("Failed to discover models: %v", err)
	}

	fmt.Printf("Discovered %d models:\n", len(models))
	for _, model := range models {
		fmt.Printf("  - %s (%s): %s\n", model.Name, model.ID, model.Description)
	}

	// Test chat completion
	fmt.Println("\n--- Testing Chat Completion ---")
	chatReq := toolkit.ChatRequest{
		Model: "custom-chat-v1",
		Messages: []toolkit.ChatMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "Hello! Can you tell me about custom AI providers?",
			},
		},
		MaxTokens:   200,
		Temperature: 0.7,
	}

	chatResp, err := customProvider.Chat(ctx, chatReq)
	if err != nil {
		log.Fatalf("Failed to perform chat completion: %v", err)
	}

	fmt.Printf("Chat Response: %s\n", chatResp.Choices[0].Message.Content)
	fmt.Printf("Tokens used: %d\n", chatResp.Usage.TotalTokens)

	// Test embedding
	fmt.Println("\n--- Testing Embedding ---")
	embedReq := toolkit.EmbeddingRequest{
		Model: "custom-embed-v1",
		Input: []string{"Hello world", "This is a test"},
	}

	embedResp, err := customProvider.Embed(ctx, embedReq)
	if err != nil {
		log.Fatalf("Failed to perform embedding: %v", err)
	}

	fmt.Printf("Embedding dimensions: %d\n", len(embedResp.Data[0].Embedding))
	fmt.Printf("Tokens used: %d\n", embedResp.Usage.TotalTokens)

	// Test reranking
	fmt.Println("\n--- Testing Reranking ---")
	rerankReq := toolkit.RerankRequest{
		Model: "custom-rerank-v1",
		Query: "artificial intelligence",
		Documents: []string{
			"AI is transforming technology",
			"The weather is nice today",
			"Machine learning is a subset of AI",
		},
		TopN: 2,
	}

	rerankResp, err := customProvider.Rerank(ctx, rerankReq)
	if err != nil {
		log.Fatalf("Failed to perform reranking: %v", err)
	}

	fmt.Println("Reranking results:")
	for _, result := range rerankResp.Results {
		fmt.Printf("  Score: %.2f - %s\n", result.Score, result.Document)
	}

	// Test configuration validation
	fmt.Println("\n--- Testing Configuration Validation ---")
	validConfig := map[string]interface{}{
		"name":    "test-provider",
		"api_key": "test-key",
	}
	if err := customProvider.ValidateConfig(validConfig); err != nil {
		log.Fatalf("Valid config failed validation: %v", err)
	}
	fmt.Println("✓ Valid configuration accepted")

	invalidConfig := map[string]interface{}{
		"name": "test-provider",
		// Missing api_key
	}
	if err := customProvider.ValidateConfig(invalidConfig); err == nil {
		log.Fatalf("Invalid config should have failed validation")
	}
	fmt.Println("✓ Invalid configuration rejected")

	// Demonstrate JSON serialization
	fmt.Println("\n--- Configuration Serialization ---")
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	fmt.Printf("Provider configuration:\n%s\n", string(configJSON))

	fmt.Println("\n=== Custom Provider Example Complete ===")
	fmt.Println("This example demonstrates how to:")
	fmt.Println("1. Create a custom provider implementing the Provider interface")
	fmt.Println("2. Register it with the toolkit")
	fmt.Println("3. Use it for various AI operations")
	fmt.Println("4. Handle configuration validation")
	fmt.Println("5. Serialize configurations")
}
