// Package anthropic_cu provides Anthropic Computer Use integration
package anthropic_cu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

const apiURL = "https://api.anthropic.com/v1/messages"

// Provider implements LLMProvider with computer use capabilities
type Provider struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

// Config holds provider configuration
type Config struct {
	APIKey    string
	Model     string // "claude-3-5-sonnet-20241022"
	MaxTokens int
}

// NewProvider creates a new Anthropic Computer Use provider
func NewProvider(config Config) *Provider {
	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	return &Provider{
		apiKey:    config.APIKey,
		model:     config.Model,
		maxTokens: config.MaxTokens,
		client:    &http.Client{Timeout: 120 * time.Second},
	}
}

// Complete implements LLMProvider
func (p *Provider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
	}

	// Add computer use tools
	tools := []map[string]interface{}{
		{
			"type":              "computer_20241022",
			"name":              "computer",
			"display_width_px":  1024,
			"display_height_px": 768,
			"display_number":    1,
		},
		{
			"type": "text_editor_20241022",
			"name": "str_replace_editor",
		},
		{
			"type": "bash_20241022",
			"name": "bash",
		},
	}

	body := map[string]interface{}{
		"model":      p.model,
		"messages":   messages,
		"max_tokens": p.maxTokens,
		"tools":      tools,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type  string                 `json:"type"`
			Text  string                 `json:"text,omitempty"`
			ID    string                 `json:"id,omitempty"`
			Name  string                 `json:"name,omitempty"`
			Input map[string]interface{} `json:"input,omitempty"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text content and tool uses
	var content string
	var toolCalls []models.ToolCall

	for _, item := range result.Content {
		switch item.Type {
		case "text":
			content += item.Text
		case "tool_use":
			argsJSON, _ := json.Marshal(item.Input)
			toolCalls = append(toolCalls, models.ToolCall{
				ID:   item.ID,
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      item.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	return &models.LLMResponse{
		Content:      content,
		TokensUsed:   result.Usage.InputTokens + result.Usage.OutputTokens,
		ToolCalls:    toolCalls,
		ProviderID:   "anthropic-cu",
		ProviderName: "Anthropic Computer Use",
	}, nil
}

// CompleteStream implements LLMProvider
func (p *Provider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *models.LLMResponse, 1)
	ch <- resp
	close(ch)
	return ch, nil
}

// HealthCheck implements LLMProvider - checks connectivity to Anthropic API
func (p *Provider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make a minimal request to check API connectivity
	req := &models.LLMRequest{
		Messages: []models.Message{{Role: "user", Content: "Hi"}},
	}

	_, err := p.Complete(ctx, req)
	return err
}

// GetCapabilities implements LLMProvider - returns provider capabilities
func (p *Provider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels: []string{
			"claude-3-5-sonnet-20241022",
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
		},
		SupportedFeatures: []string{
			"chat", "streaming", "tools", "vision", "computer_use",
			"text_editor", "bash_execution",
		},
		SupportedRequestTypes:   []string{"chat", "completion"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          true,
		SupportsTools:           true,
		SupportsSearch:          false,
		SupportsReasoning:       true,
		SupportsCodeCompletion:  true,
		SupportsCodeAnalysis:    true,
		SupportsRefactoring:     true,
		Limits: models.ModelLimits{
			MaxTokens:             200000,
			MaxInputLength:        200000,
			MaxOutputLength:       8192,
			MaxConcurrentRequests: 50,
		},
		Metadata: map[string]string{
			"provider": "anthropic-cu",
			"type":     "computer_use",
		},
	}
}

// ValidateConfig implements LLMProvider - validates provider configuration
func (p *Provider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if p.apiKey == "" {
		errors = append(errors, "API key is required")
	}
	if p.model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}

// Ensure interface is implemented
var _ llm.LLMProvider = (*Provider)(nil)
