// Package nvidia provides a Go client for the Nvidia AI API.
package nvidia

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

// Client represents a Nvidia AI API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Nvidia AI API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://integrate.api.nvidia.com/v1",
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

// doRequest performs an HTTP request to the Nvidia API.
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
