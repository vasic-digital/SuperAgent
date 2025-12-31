// Package main provides a mock LLM server for testing purposes.
// This server mimics the API responses of various LLM providers
// to enable fast, reliable, and deterministic testing without
// requiring actual LLM API connections.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// CompletionRequest represents an incoming completion request
type CompletionRequest struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionResponse represents the completion response
type CompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Text         string  `json:"text,omitempty"`
	Delta        *Delta  `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason"`
}

// Delta represents a streaming delta
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelsResponse represents the models list response
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents an available model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  Usage           `json:"usage"`
}

// EmbeddingData represents embedding data
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Claude-specific request/response types
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Messages  []ClaudeMessage `json:"messages"`
	Stream    bool            `json:"stream,omitempty"`
	System    string          `json:"system,omitempty"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []ClaudeContent `json:"content"`
	Model        string          `json:"model"`
	StopReason   *string         `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
	Usage        ClaudeUsage     `json:"usage"`
}

type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Ollama-specific request/response types
type OllamaRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream,omitempty"`
	Options struct {
		Temperature float64  `json:"temperature,omitempty"`
		MaxTokens   int      `json:"num_predict,omitempty"`
	} `json:"options,omitempty"`
}

type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Gemini-specific request/response types
type GeminiRequest struct {
	Contents         []GeminiContent       `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text string `json:"text,omitempty"`
}

type GeminiGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type GeminiResponse struct {
	Candidates    []GeminiCandidate    `json:"candidates"`
	UsageMetadata *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

var mockModels = []Model{
	{ID: "mock-gpt-4", Object: "model", Created: time.Now().Unix(), OwnedBy: "mock-provider"},
	{ID: "mock-claude-3", Object: "model", Created: time.Now().Unix(), OwnedBy: "mock-provider"},
	{ID: "mock-llama-3", Object: "model", Created: time.Now().Unix(), OwnedBy: "mock-provider"},
	{ID: "mock-gemini-pro", Object: "model", Created: time.Now().Unix(), OwnedBy: "mock-provider"},
}

func generateMockResponse(prompt string, model string) string {
	// Generate deterministic mock responses based on prompt content
	lowerPrompt := strings.ToLower(prompt)

	switch {
	case strings.Contains(lowerPrompt, "hello"):
		return "Hello! I'm a mock LLM response. How can I help you today?"
	case strings.Contains(lowerPrompt, "code"):
		return "```python\ndef hello_world():\n    print('Hello, World!')\n```\nHere's a simple code example for you."
	case strings.Contains(lowerPrompt, "explain"):
		return "This is a mock explanation. The concept involves understanding the fundamental principles and applying them systematically."
	case strings.Contains(lowerPrompt, "summarize"):
		return "Summary: The main points are A, B, and C. The key takeaway is that effective summarization captures essential information concisely."
	case strings.Contains(lowerPrompt, "test"):
		return "This is a test response from the mock LLM server. All systems are functioning correctly."
	default:
		return fmt.Sprintf("Mock response for model %s: Your request has been processed. The mock server received your prompt and generated this deterministic response.", model)
	}
}

func handleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract prompt from request
	prompt := req.Prompt
	if prompt == "" && len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}

	responseContent := generateMockResponse(prompt, req.Model)

	if req.Stream {
		handleStreamingResponse(w, req, responseContent)
		return
	}

	response := CompletionResponse{
		ID:      fmt.Sprintf("mock-cmpl-%d", time.Now().UnixNano()),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Text:  responseContent,
				Message: Message{
					Role:    "assistant",
					Content: responseContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     len(strings.Fields(prompt)) * 2,
			CompletionTokens: len(strings.Fields(responseContent)) * 2,
			TotalTokens:      (len(strings.Fields(prompt)) + len(strings.Fields(responseContent))) * 2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleStreamingResponse(w http.ResponseWriter, req CompletionRequest, content string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	words := strings.Fields(content)
	for i, word := range words {
		chunk := CompletionResponse{
			ID:      fmt.Sprintf("mock-cmpl-%d", time.Now().UnixNano()),
			Object:  "text_completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []Choice{
				{
					Index: 0,
					Delta: &Delta{
						Content: word + " ",
					},
					FinishReason: "",
				},
			},
		}

		if i == len(words)-1 {
			chunk.Choices[0].FinishReason = "stop"
		}

		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		time.Sleep(50 * time.Millisecond) // Simulate streaming delay
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}

	responseContent := generateMockResponse(prompt, req.Model)

	if req.Stream {
		handleStreamingResponse(w, req, responseContent)
		return
	}

	response := CompletionResponse{
		ID:      fmt.Sprintf("mock-chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: responseContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     len(strings.Fields(prompt)) * 2,
			CompletionTokens: len(strings.Fields(responseContent)) * 2,
			TotalTokens:      (len(strings.Fields(prompt)) + len(strings.Fields(responseContent))) * 2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := ModelsResponse{
		Object: "list",
		Data:   mockModels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Generate mock embeddings (1536 dimensions like OpenAI)
	embeddings := make([]EmbeddingData, len(req.Input))
	for i := range req.Input {
		embedding := make([]float64, 1536)
		for j := range embedding {
			embedding[j] = float64(i+j) / 1536.0 // Deterministic mock embedding
		}
		embeddings[i] = EmbeddingData{
			Object:    "embedding",
			Index:     i,
			Embedding: embedding,
		}
	}

	response := EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  req.Model,
		Usage: Usage{
			PromptTokens: len(req.Input) * 10,
			TotalTokens:  len(req.Input) * 10,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOllamaGenerate handles Ollama's /api/generate endpoint
func handleOllamaGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OllamaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	responseContent := generateMockResponse(req.Prompt, req.Model)

	// Return Ollama-compatible response format
	response := OllamaResponse{
		Model:    req.Model,
		Response: responseContent,
		Done:     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGeminiGenerate handles Gemini's /v1beta/models/:model:generateContent endpoint
func handleGeminiGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GeminiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract prompt from contents
	prompt := ""
	if len(req.Contents) > 0 {
		content := req.Contents[len(req.Contents)-1]
		if len(content.Parts) > 0 {
			prompt = content.Parts[0].Text
		}
	}

	responseContent := generateMockResponse(prompt, "gemini")

	// Return Gemini-compatible response format
	response := GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Content: GeminiContent{
					Parts: []GeminiPart{
						{Text: responseContent},
					},
					Role: "model",
				},
				FinishReason: "STOP",
				Index:        0,
			},
		},
		UsageMetadata: &GeminiUsageMetadata{
			PromptTokenCount:     len(strings.Fields(prompt)) * 2,
			CandidatesTokenCount: len(strings.Fields(responseContent)) * 2,
			TotalTokenCount:      (len(strings.Fields(prompt)) + len(strings.Fields(responseContent))) * 2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleQwenGeneration handles Qwen's /services/aigc/text-generation/generation endpoint
func handleQwenGeneration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract prompt from messages
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}

	responseContent := generateMockResponse(prompt, req.Model)

	// Return OpenAI-compatible response (Qwen uses OpenAI format)
	response := CompletionResponse{
		ID:      fmt.Sprintf("qwen-cmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: responseContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     len(strings.Fields(prompt)) * 2,
			CompletionTokens: len(strings.Fields(responseContent)) * 2,
			TotalTokens:      (len(strings.Fields(prompt)) + len(strings.Fields(responseContent))) * 2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClaudeMessages handles Claude-specific /v1/messages endpoint
func handleClaudeMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClaudeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract prompt from messages
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}

	responseContent := generateMockResponse(prompt, req.Model)
	stopReason := "end_turn"

	// Return Claude-compatible response format
	response := ClaudeResponse{
		ID:         fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Type:       "message",
		Role:       "assistant",
		Model:      req.Model,
		StopReason: &stopReason,
		Content: []ClaudeContent{
			{
				Type: "text",
				Text: responseContent,
			},
		},
		Usage: ClaudeUsage{
			InputTokens:  len(strings.Fields(prompt)) * 2,
			OutputTokens: len(strings.Fields(responseContent)) * 2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "mock-llm-server",
		"version": "1.0.0",
		"models":  len(mockModels),
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/completions", handleCompletions)
	mux.HandleFunc("/v1/chat/completions", handleChatCompletions)
	mux.HandleFunc("/v1/models", handleModels)
	mux.HandleFunc("/v1/embeddings", handleEmbeddings)

	// Claude-compatible endpoint
	mux.HandleFunc("/v1/messages", handleClaudeMessages)

	// Ollama-compatible endpoint
	mux.HandleFunc("/api/generate", handleOllamaGenerate)

	// Qwen-compatible endpoint
	mux.HandleFunc("/services/aigc/text-generation/generation", handleQwenGeneration)

	// Gemini-compatible endpoint (matches /v1beta/models/*/generateContent pattern)
	mux.HandleFunc("/v1beta/models/", handleGeminiGenerate)

	// Health check
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/", handleHealth)

	log.Printf("Mock LLM Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /v1/completions")
	log.Printf("  POST /v1/chat/completions")
	log.Printf("  POST /v1/messages (Claude)")
	log.Printf("  POST /api/generate (Ollama)")
	log.Printf("  POST /services/aigc/text-generation/generation (Qwen)")
	log.Printf("  POST /v1beta/models/:model:generateContent (Gemini)")
	log.Printf("  GET  /v1/models")
	log.Printf("  POST /v1/embeddings")
	log.Printf("  GET  /health")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
