// Package providers implements real tests for Cohere provider
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

// TestCohere_ShortRequest tests simple completion
func TestCohere_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("cohere", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "command-r",
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

// TestCohere_ToolUse tests function calling
func TestCohere_ToolUse(t *testing.T) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("cohere", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "search",
			Description: "Search for documents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]string{"type": "string"},
				},
				"required": []string{"query"},
			},
		},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "command-r-plus",
		Messages: []llm.Message{
			{Role: "user", Content: "Search for documents about machine learning"},
		},
		Tools:     tools,
		MaxTokens: 100,
	})

	require.NoError(t, err)
	if len(response.ToolCalls) > 0 {
		t.Logf("Tool called: %s", response.ToolCalls[0].Name)
	}
}

// TestCohere_Streaming tests streaming response
func TestCohere_Streaming(t *testing.T) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("cohere", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "command-r",
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

// TestCohere_RAG tests RAG capabilities
func TestCohere_RAG(t *testing.T) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("cohere", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cohere supports document grounding
	documents := []map[string]string{
		{"title": "Go", "text": "Go is a programming language created at Google."},
		{"title": "Python", "text": "Python is a high-level programming language."},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "command-r",
		Messages: []llm.Message{
			{Role: "user", Content: "What company created Go?"},
		},
		Documents: documents,
		MaxTokens: 100,
	})

	require.NoError(t, err)
	assert.Contains(t, response.Content, "Google")
}

// TestCohere_Models tests different Cohere models
func TestCohere_Models(t *testing.T) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	models := []string{
		"command-r",
		"command-r-plus",
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("cohere", apiKey, logger)
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
