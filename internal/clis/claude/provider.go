// Package claude provides LLM provider implementation for Claude Code.
package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/clis/claude/api"
	"dev.helix.agent/internal/clis/claude/strategy"
)

// Provider implements the LLM provider interface for Claude Code
 type Provider struct {
	name       string
	strategy   strategy.Strategy
	integration *Integration
	config     *Config
}

// NewProvider creates a new Claude Code provider
 func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	integration, err := NewIntegration(config)
	if err != nil {
		return nil, fmt.Errorf("create integration: %w", err)
	}
	
	return &Provider{
		name:        fmt.Sprintf("claude_code_%s", config.StrategyType),
		strategy:    integration.strategy,
		integration: integration,
		config:      config,
	}, nil
}

// Name returns the provider name
 func (p *Provider) Name() string {
	return p.name
}

// Complete sends a completion request
func (p *Provider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if !p.integration.IsStarted() {
		if err := p.integration.Start(ctx); err != nil {
			return nil, fmt.Errorf("start integration: %w", err)
		}
	}
	
	// Convert LLM request to Claude API request
	claudeReq := &api.MessageRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages:  convertMessages(req.Messages),
	}
	
	if req.Temperature != nil {
		claudeReq.Temperature = req.Temperature
	}
	
	resp, err := p.integration.CreateMessage(ctx, claudeReq)
	if err != nil {
		return nil, err
	}
	
	return convertResponse(resp), nil
}

// CompleteStream sends a streaming completion request
func (p *Provider) CompleteStream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.CompletionChunk, error) {
	if !p.integration.IsStarted() {
		if err := p.integration.Start(ctx); err != nil {
			return nil, fmt.Errorf("start integration: %w", err)
		}
	}
	
	claudeReq := &api.MessageRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages:  convertMessages(req.Messages),
		Stream:    true,
	}
	
	events, errs := p.integration.CreateMessageStream(ctx, claudeReq)
	
	chunks := make(chan llm.CompletionChunk)
	
	go func() {
		defer close(chunks)
		
		for {
			select {
			case event, ok := <-events:
				if !ok {
					return
				}
				
				chunk := convertEventToChunk(event)
				if chunk != nil {
					select {
					case chunks <- *chunk:
					case <-ctx.Done():
						return
					}
				}
				
			case err := <-errs:
				if err != nil {
					select {
					case chunks <- llm.CompletionChunk{Error: err}:
					case <-ctx.Done():
					}
					return
				}
				
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, nil
}

// Health checks if the provider is healthy
func (p *Provider) Health(ctx context.Context) error {
	return p.integration.HealthCheck(ctx)
}

// Models returns available models
func (p *Provider) Models(ctx context.Context) ([]llm.Model, error) {
	models := []llm.Model{
		{
			ID:       api.ModelClaudeOpus4_1,
			Name:     "Claude Opus 4.1",
			Provider: p.name,
			Capabilities: llm.Capabilities{
				Chat:           true,
				Streaming:      true,
				FunctionCall:   true,
				Vision:         true,
				CodeGeneration: true,
			},
		},
		{
			ID:       api.ModelClaudeSonnet4_5,
			Name:     "Claude Sonnet 4.5",
			Provider: p.name,
			Capabilities: llm.Capabilities{
				Chat:           true,
				Streaming:      true,
				FunctionCall:   true,
				Vision:         true,
				CodeGeneration: true,
			},
		},
		{
			ID:       api.ModelClaudeSonnet4_6,
			Name:     "Claude Sonnet 4.6",
			Provider: p.name,
			Capabilities: llm.Capabilities{
				Chat:           true,
				Streaming:      true,
				FunctionCall:   true,
				Vision:         true,
				CodeGeneration: true,
			},
		},
		{
			ID:       api.ModelClaudeHaiku4_5,
			Name:     "Claude Haiku 4.5",
			Provider: p.name,
			Capabilities: llm.Capabilities{
				Chat:           true,
				Streaming:      true,
				FunctionCall:   true,
				Vision:         true,
				CodeGeneration: true,
			},
		},
	}
	
	return models, nil
}

// GetIntegration returns the underlying integration
func (p *Provider) GetIntegration() *Integration {
	return p.integration
}

// GetStrategy returns the current strategy
func (p *Provider) GetStrategy() strategy.Strategy {
	return p.strategy
}

// convertMessages converts LLM messages to Claude API messages
func convertMessages(msgs []llm.Message) []api.Message {
	var result []api.Message
	
	for _, msg := range msgs {
		content := msg.Content
		
		// Handle structured content
		if msg.Role == llm.RoleAssistant && len(msg.ToolCalls) > 0 {
			var blocks []api.ContentBlock
			
			// Add text content if present
			if msg.Content != "" {
				blocks = append(blocks, api.ContentBlock{
					Type: "text",
					Text: msg.Content,
				})
			}
			
			// Add tool calls
			for _, tc := range msg.ToolCalls {
				inputJSON, _ := json.Marshal(tc.Arguments)
				blocks = append(blocks, api.ContentBlock{
					Type: "tool_use",
					ID:   tc.ID,
					Name: tc.Name,
					Input: inputJSON,
				})
			}
			
			content = blocks
		}
		
		if msg.Role == llm.RoleTool {
			// Tool result
			content = []api.ContentBlock{
				{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
					IsError:   msg.IsError,
				},
			}
		}
		
		result = append(result, api.Message{
			Role:    string(msg.Role),
			Content: content,
		})
	}
	
	return result
}

// convertResponse converts Claude API response to LLM response
func convertResponse(resp *api.MessageResponse) *llm.CompletionResponse {
	var content string
	var toolCalls []llm.ToolCall
	
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			var args map[string]interface{}
			json.Unmarshal(block.Input, &args)
			toolCalls = append(toolCalls, llm.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: args,
			})
		}
	}
	
	return &llm.CompletionResponse{
		ID:        resp.ID,
		Content:   content,
		ToolCalls: toolCalls,
		Usage: llm.Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
	}
}

// convertEventToChunk converts a stream event to a completion chunk
func convertEventToChunk(event api.StreamEvent) *llm.CompletionChunk {
	switch event.Type {
	case "content_block_delta":
		if event.Delta != nil && event.Delta.Type == "text_delta" {
			return &llm.CompletionChunk{
				Content: event.Delta.Text,
			}
		}
		
	case "message_stop":
		return &llm.CompletionChunk{
			IsFinished: true,
		}
		
	case "message_delta":
		if event.Delta != nil && event.Delta.StopReason != "" {
			return &llm.CompletionChunk{
				IsFinished: true,
			}
		}
	}
	
	return nil
}


