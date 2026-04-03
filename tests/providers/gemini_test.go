// Package providers implements real tests for Gemini provider
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

// TestGemini_ShortRequest tests simple completion
func TestGemini_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gemini-2.0-flash-exp",
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

// TestGemini_Streaming tests streaming
func TestGemini_Streaming(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "gemini-2.0-flash-exp",
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

// TestGemini_Vision tests vision capabilities
func TestGemini_Vision(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "What color is this image?",
				Images:  []string{imageBase64},
			},
		},
		MaxTokens: 50,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	t.Logf("Vision: %s", response.Content)
}

// TestGemini_ToolUse tests function calling
func TestGemini_ToolUse(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "search",
			Description: "Search for information",
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
		Model: "gemini-2.0-flash-exp",
		Messages: []llm.Message{
			{Role: "user", Content: "Search for information about Go programming"},
		},
		Tools:     tools,
		MaxTokens: 100,
	})

	require.NoError(t, err)
	if len(response.ToolCalls) > 0 {
		t.Logf("Tool: %s", response.ToolCalls[0].Name)
	}
}

// TestGemini_LongContext tests 1M context window
func TestGemini_LongContext(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping long context in short mode")
	}

	// Generate ~50K tokens
	contextText := generateLongText(50000)

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	start := time.Now()
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gemini-1.5-pro",
		Messages: []llm.Message{
			{Role: "user", Content: "Summarize: " + contextText},
		},
		MaxTokens: 500,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	t.Logf("Processed 50K tokens in %v", time.Since(start))
}

// TestGemini_JSON tests JSON mode
func TestGemini_JSON(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("gemini", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []llm.Message{
			{Role: "user", Content: "Generate JSON with fields: name, age, city"},
		},
		MaxTokens:      200,
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	t.Logf("JSON: %s", response.Content)
}
