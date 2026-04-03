// Package providers implements real tests for Mistral provider
package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"dev.helix.agent/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMistral_ShortRequest tests simple completion
func TestMistral_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("mistral", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "mistral-small-latest",
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello in exactly 3 words"},
		},
		MaxTokens: 50,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Contains(t, response.Content, "hello")

	t.Logf("Response: %s", response.Content)
}

// TestMistral_Streaming tests streaming response
func TestMistral_Streaming(t *testing.T) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("mistral", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "mistral-small-latest",
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 5"},
		},
		MaxTokens: 100,
	})
	require.NoError(t, err)

	var chunks int
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		chunks++
	}

	assert.Greater(t, chunks, 1)
}

// TestMistral_ToolUse tests function calling
func TestMistral_ToolUse(t *testing.T) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("mistral", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]string{"type": "string"},
				},
				"required": []string{"location"},
			},
		},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "mistral-large-latest",
		Messages: []llm.Message{
			{Role: "user", Content: "What's the weather in Paris?"},
		},
		Tools:     tools,
		MaxTokens: 100,
	})

	require.NoError(t, err)
	if len(response.ToolCalls) > 0 {
		t.Logf("Tool called: %s", response.ToolCalls[0].Name)
	}
}

// TestMistral_Models tests different Mistral models
func TestMistral_Models(t *testing.T) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	models := []string{
		"mistral-small-latest",
		"mistral-large-latest",
		"codestral-latest",
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("mistral", apiKey, logger)
	ctx := context.Background()

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			response, err := client.Chat(ctx, llm.ChatRequest{
				Model: model,
				Messages: []llm.Message{
					{Role: "user", Content: "Say 'test'"},
				},
				MaxTokens: 10,
			})

			if err != nil {
				t.Logf("Model %s failed: %v", model, err)
				return
			}

			t.Logf("%s: %s", model, response.Content)
		})
	}
}

// TestMistral_Agents tests Mistral's agent capabilities
func TestMistral_Agents(t *testing.T) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("mistral", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with agent-specific model
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "mistral-large-latest",
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful coding assistant."},
			{Role: "user", Content: "Write a Python function to calculate factorial"},
		},
		MaxTokens: 200,
		Temperature: 0.3,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Contains(t, response.Content, "def")
}
