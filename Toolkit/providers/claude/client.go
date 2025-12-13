// Package claude provides a Go client for the Anthropic Claude API.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/superagent/toolkit/pkg/toolkit"
)

// Client represents a Claude API client.
type Client struct {
	apiKey     string
	baseURL    string
	version    string
	httpClient *http.Client
}

// NewClient creates a new Claude API client.
func NewClient(apiKey, version string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com",
		version: version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatCompletion performs a chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req toolkit.ChatRequest) (toolkit.ChatResponse, error) {
	// Convert toolkit messages to Claude format
	var messages []map[string]interface{}
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	payload := map[string]interface{}{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": req.MaxTokens,
	}

	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		payload["top_p"] = req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop_sequences"] = req.Stop
	}

	var response toolkit.ChatResponse
	err := c.doRequest(ctx, "POST", "/v1/messages", payload, &response)
	return response, err
}

// CreateEmbeddings performs an embedding request.
// Note: Claude doesn't have native embeddings, this would need to be handled differently
func (c *Client) CreateEmbeddings(ctx context.Context, req toolkit.EmbeddingRequest) (toolkit.EmbeddingResponse, error) {
	return toolkit.EmbeddingResponse{}, fmt.Errorf("Claude does not support embeddings directly")
}

// CreateRerank performs a rerank request.
// Note: Claude doesn't have native reranking, this would need to be handled differently
func (c *Client) CreateRerank(ctx context.Context, req toolkit.RerankRequest) (toolkit.RerankResponse, error) {
	return toolkit.RerankResponse{}, fmt.Errorf("Claude does not support reranking directly")
}

// GetModels retrieves available models from the API.
// Note: Anthropic's API may not have a models endpoint, this is a placeholder
func (c *Client) GetModels(ctx context.Context) ([]ModelInfo, error) {
	// Claude models are typically known, not fetched from API
	models := []ModelInfo{
		{ID: "claude-3-opus-20240229", Type: "chat"},
		{ID: "claude-3-sonnet-20240229", Type: "chat"},
		{ID: "claude-3-haiku-20240307", Type: "chat"},
		{ID: "claude-3-5-sonnet-20240620", Type: "chat"},
	}
	return models, nil
}

// ModelInfo represents basic model information.
type ModelInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// doRequest performs an HTTP request to the Claude API.
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

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", c.version)
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
