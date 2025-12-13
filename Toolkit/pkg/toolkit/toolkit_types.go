// Package toolkit provides generic types and interfaces for AI provider toolkits.
package toolkit

import (
	"time"
)

// ModelCapabilities represents the capabilities of an AI model.
type ModelCapabilities struct {
	SupportsChat      bool    `json:"supports_chat"`
	SupportsEmbedding bool    `json:"supports_embedding"`
	SupportsRerank    bool    `json:"supports_rerank"`
	SupportsVision    bool    `json:"supports_vision"`
	SupportsAudio     bool    `json:"supports_audio"`
	SupportsVideo     bool    `json:"supports_video"`
	MaxTokens         int     `json:"max_tokens"`
	ContextWindow     int     `json:"context_window"`
	InputPricing      float64 `json:"input_pricing,omitempty"`
	OutputPricing     float64 `json:"output_pricing,omitempty"`
	ImageSupport      bool    `json:"image_support,omitempty"`
	FunctionCalling   bool    `json:"function_calling,omitempty"`
}

// ModelCategory represents the category of an AI model.
type ModelCategory string

const (
	CategoryChat       ModelCategory = "chat"
	CategoryEmbedding  ModelCategory = "embedding"
	CategoryRerank     ModelCategory = "rerank"
	CategoryImage      ModelCategory = "image"
	CategoryMultimodal ModelCategory = "multimodal"
)

// ModelInfo contains information about a specific model.
type ModelInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Category     ModelCategory     `json:"category"`
	Capabilities ModelCapabilities `json:"capabilities"`
	Provider     string            `json:"provider"`
	Description  string            `json:"description,omitempty"`
	CreatedAt    *time.Time        `json:"created_at,omitempty"`
	UpdatedAt    *time.Time        `json:"updated_at,omitempty"`
}

// ModelRegistry manages a collection of model information.
type ModelRegistry struct {
	Models map[string]ModelInfo `json:"models"`
}

// NewModelRegistry creates a new ModelRegistry.
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		Models: make(map[string]ModelInfo),
	}
}

// Register adds a model to the registry.
func (r *ModelRegistry) Register(model ModelInfo) {
	r.Models[model.ID] = model
}

// Get retrieves a model by ID.
func (r *ModelRegistry) Get(id string) (ModelInfo, bool) {
	model, ok := r.Models[id]
	return model, ok
}

// List returns all registered models.
func (r *ModelRegistry) List() []ModelInfo {
	var models []ModelInfo
	for _, model := range r.Models {
		models = append(models, model)
	}
	return models
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role      string    `json:"role"` // "user", "assistant", "system", etc.
	Content   string    `json:"content"`
	Name      string    `json:"name,omitempty"` // For function calls
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ChatRequest represents a request to generate chat completions.
type ChatRequest struct {
	Model            string         `json:"model"`
	Messages         []ChatMessage  `json:"messages"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	TopK             int            `json:"top_k,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int `json:"logit_bias,omitempty"`
	User             string         `json:"user,omitempty"`
}

// ChatChoice represents a single choice in a chat response.
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatResponse represents the response from a chat completion request.
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest represents a request to generate embeddings.
type EmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	User           string   `json:"user,omitempty"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	Dimensions     int      `json:"dimensions,omitempty"`
}

// EmbeddingData represents a single embedding vector.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse represents the response from an embedding request.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  Usage           `json:"usage"`
}

// RerankRequest represents a request to rerank documents.
type RerankRequest struct {
	Model      string   `json:"model"`
	Query      string   `json:"query"`
	Documents  []string `json:"documents"`
	TopN       int      `json:"top_n,omitempty"`
	ReturnDocs bool     `json:"return_documents,omitempty"`
}

// RerankResult represents a single rerank result.
type RerankResult struct {
	Index    int     `json:"index"`
	Score    float64 `json:"score"`
	Document string  `json:"document,omitempty"`
}

// RerankResponse represents the response from a rerank request.
type RerankResponse struct {
	Object  string         `json:"object"`
	Model   string         `json:"model"`
	Results []RerankResult `json:"results"`
}

// APIError represents an error returned by an API.
type APIError struct {
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message"`
	Type      string      `json:"type,omitempty"`
	Param     string      `json:"param,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

// Error implements the error interface.
func (e APIError) Error() string {
	return e.Message
}
