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

	// Health check
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/", handleHealth)

	log.Printf("Mock LLM Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /v1/completions")
	log.Printf("  POST /v1/chat/completions")
	log.Printf("  GET  /v1/models")
	log.Printf("  POST /v1/embeddings")
	log.Printf("  GET  /health")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
