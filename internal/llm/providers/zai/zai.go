package zai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/superagent/superagent/internal/models"
)

// ZAIProvider implements the LLMProvider interface for Z.AI
type ZAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// ZAIRequest represents a request to the Z.AI API
type ZAIRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt,omitempty"`
	Messages    []ZAIMessage           `json:"messages,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stop        []string               `json:"stop,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ZAIMessage represents a message in the Z.AI API format
type ZAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ZAIResponse represents a response from the Z.AI API
type ZAIResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []ZAIChoice `json:"choices"`
	Usage   ZAIUsage    `json:"usage"`
}

// ZAIChoice represents a choice in the Z.AI response
type ZAIChoice struct {
	Index        int        `json:"index"`
	Text         string     `json:"text,omitempty"`
	Message      ZAIMessage `json:"message,omitempty"`
	FinishReason string     `json:"finish_reason"`
}

// ZAIUsage represents token usage in the Z.AI response
type ZAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ZAIError represents an error from the Z.AI API
type ZAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// NewZAIProvider creates a new Z.AI provider instance
func NewZAIProvider(apiKey, baseURL, model string) *ZAIProvider {
	if baseURL == "" {
		baseURL = "https://api.z.ai/v1"
	}
	if model == "" {
		model = "z-ai-base"
	}

	return &ZAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Complete implements the LLMProvider interface
func (z *ZAIProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	// Convert internal request to Z.AI format
	zaiReq := z.convertToZAIRequest(req)

	// Make API call
	resp, err := z.makeRequest(ctx, zaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to complete request: %w", err)
	}

	// Convert response back to internal format
	return z.convertFromZAIResponse(resp, req.ID)
}

// CompleteStream implements streaming completion for Z.AI
func (z *ZAIProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	responseChan := make(chan *models.LLMResponse, 10)

	go func() {
		defer close(responseChan)

		// For now, simulate streaming by getting the complete response and sending it in chunks
		// In a full implementation, this would use Z.AI's actual streaming API

		response, err := z.Complete(ctx, req)
		if err != nil {
			// For streaming, we just close the channel on error
			// In a real implementation, you might want to send an error response
			return
		}

		// Simulate streaming by breaking the response into chunks
		content := response.Content
		chunkSize := 50 // characters per chunk

		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}

			chunk := content[i:end]

			streamResponse := &models.LLMResponse{
				ID:           fmt.Sprintf("%s-chunk-%d", response.ID, i/chunkSize),
				ProviderID:   response.ProviderID,
				ProviderName: response.ProviderName,
				Content:      chunk,
				Confidence:   response.Confidence,
				TokensUsed:   response.TokensUsed / (len(content)/chunkSize + 1), // Approximate token distribution
				ResponseTime: response.ResponseTime / int64(len(content)/chunkSize+1),
				FinishReason: func() string {
					if end >= len(content) {
						return "stop"
					}
					return ""
				}(),
				CreatedAt: time.Now(),
			}

			select {
			case responseChan <- streamResponse:
			case <-ctx.Done():
				return
			}

			// Small delay to simulate streaming
			time.Sleep(50 * time.Millisecond)
		}
	}()

	return responseChan, nil
}

// HealthCheck implements health checking for the Z.AI provider
func (z *ZAIProvider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Simple health check - try to get models list or basic endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", z.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+z.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetCapabilities returns the capabilities of the Z.AI provider
func (z *ZAIProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels: []string{
			"z-ai-base",
			"z-ai-pro",
			"z-ai-enterprise",
		},
		SupportedFeatures: []string{
			"text_completion",
			"chat",
		},
		SupportedRequestTypes: []string{
			"text_completion",
			"chat",
		},
		SupportsStreaming:       false, // Not implemented yet
		SupportsFunctionCalling: false,
		SupportsVision:          false,
		Limits: models.ModelLimits{
			MaxTokens:             4096,
			MaxInputLength:        8192,
			MaxOutputLength:       4096,
			MaxConcurrentRequests: 10,
		},
		Metadata: map[string]string{
			"provider":     "Z.AI",
			"model_family": "Z.AI",
			"api_version":  "v1",
		},
	}
}

// ValidateConfig validates the provider configuration
func (z *ZAIProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if z.apiKey == "" {
		errors = append(errors, "API key is required")
	}

	if z.baseURL == "" {
		errors = append(errors, "base URL is required")
	}

	if z.model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}

// convertToZAIRequest converts internal request format to Z.AI API format
func (z *ZAIProvider) convertToZAIRequest(req *models.LLMRequest) *ZAIRequest {
	zaiReq := &ZAIRequest{
		Model:       z.model,
		Stream:      false,
		Temperature: req.ModelParams.Temperature,
		MaxTokens:   req.ModelParams.MaxTokens,
		TopP:        req.ModelParams.TopP,
		Stop:        req.ModelParams.StopSequences,
		Parameters:  make(map[string]interface{}),
	}

	// Handle different request types
	if len(req.Messages) > 0 {
		// Chat format
		messages := make([]ZAIMessage, 0, len(req.Messages))
		for _, msg := range req.Messages {
			messages = append(messages, ZAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
		zaiReq.Messages = messages
	} else {
		// Completion format
		zaiReq.Prompt = req.Prompt
	}

	return zaiReq
}

// convertFromZAIResponse converts Z.AI API response to internal format
func (z *ZAIProvider) convertFromZAIResponse(resp *ZAIResponse, requestID string) (*models.LLMResponse, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from Z.AI API")
	}

	choice := resp.Choices[0]

	// Extract content from either text or message field
	var content string
	if choice.Text != "" {
		content = choice.Text
	} else if choice.Message.Content != "" {
		content = choice.Message.Content
	} else {
		return nil, fmt.Errorf("no content found in Z.AI response")
	}

	return &models.LLMResponse{
		ID:           resp.ID,
		RequestID:    requestID,
		ProviderID:   "zai",
		ProviderName: "Z.AI",
		Content:      content,
		Confidence:   0.80, // Z.AI doesn't provide confidence scores
		TokensUsed:   resp.Usage.TotalTokens,
		ResponseTime: time.Now().UnixMilli() - (resp.Created * 1000),
		FinishReason: choice.FinishReason,
		Metadata: map[string]interface{}{
			"model":             resp.Model,
			"object":            resp.Object,
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
		},
		CreatedAt: time.Now(),
	}, nil
}

// makeRequest sends a request to the Z.AI API
func (z *ZAIProvider) makeRequest(ctx context.Context, req *ZAIRequest) (*ZAIResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine endpoint based on request type
	endpoint := "/completions"
	if len(req.Messages) > 0 {
		endpoint = "/chat/completions"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", z.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var zaiErr ZAIError
		if err := json.Unmarshal(body, &zaiErr); err == nil && zaiErr.Error.Message != "" {
			return nil, fmt.Errorf("Z.AI API error: %s (%s)", zaiErr.Error.Message, zaiErr.Error.Type)
		}
		return nil, fmt.Errorf("Z.AI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var zaiResp ZAIResponse
	if err := json.Unmarshal(body, &zaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &zaiResp, nil
}
