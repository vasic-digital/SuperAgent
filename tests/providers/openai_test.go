// Package providers implements real tests for OpenAI provider
package providers

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"dev.helix.agent/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestOpenAI_ShortRequest tests simple completion
func TestOpenAI_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello in exactly 3 words"},
		},
		MaxTokens: 50,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Contains(t, response.Content, "hello")
	assert.Greater(t, response.Usage.TotalTokens, 0)

	t.Logf("Response: %s", response.Content)
	t.Logf("Tokens: %d", response.Usage.TotalTokens)
}

// TestOpenAI_Streaming tests streaming response
func TestOpenAI_Streaming(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 5"},
		},
		MaxTokens: 100,
	})
	require.NoError(t, err)

	var fullContent string
	var chunks int
	startTime := time.Now()

	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		fullContent += chunk.Content
		chunks++
	}

	duration := time.Since(startTime)

	assert.NotEmpty(t, fullContent)
	assert.Greater(t, chunks, 1) // Should have multiple chunks
	assert.Less(t, duration, 30*time.Second)

	t.Logf("Received %d chunks in %v", chunks, duration)
	t.Logf("Content: %s", fullContent)
}

// TestOpenAI_JSONMode tests JSON mode
func TestOpenAI_JSONMode(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{Role: "user", Content: "Generate a JSON object with name and age"},
		},
		MaxTokens:      200,
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
	})

	require.NoError(t, err)

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(response.Content), &result)
	require.NoError(t, err, "Response should be valid JSON")

	assert.NotNil(t, result["name"], "Should have name field")
	assert.NotNil(t, result["age"], "Should have age field")

	t.Logf("JSON Response: %s", response.Content)
}

// TestOpenAI_ToolCalling tests function calling
func TestOpenAI_ToolCalling(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get the current weather",
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
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{Role: "user", Content: "What's the weather in San Francisco?"},
		},
		Tools:     tools,
		MaxTokens: 200,
	})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.ToolCalls), 1, "Should have tool calls")
	assert.Equal(t, "get_weather", response.ToolCalls[0].Name)

	t.Logf("Tool call: %s", response.ToolCalls[0].Name)
	t.Logf("Arguments: %s", response.ToolCalls[0].Arguments)
}

// TestOpenAI_LongContext tests context handling
func TestOpenAI_LongContext(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Generate ~10K tokens of context
	contextText := generateLongText(10000)

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gpt-4o",
		Messages: []llm.Message{
			{Role: "system", Content: "Summarize the following text:"},
			{Role: "user", Content: contextText},
		},
		MaxTokens: 500,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Greater(t, response.Usage.InputTokens, 5000, "Should have processed large input")

	t.Logf("Input tokens: %d", response.Usage.InputTokens)
	t.Logf("Output tokens: %d", response.Usage.OutputTokens)
}

// TestOpenAI_Vision tests vision capabilities
func TestOpenAI_Vision(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Base64 encoded image (1x1 red pixel as example)
	imageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gpt-4o",
		Messages: []llm.Message{
			{
				Role: "user",
				Content: "What color is this image?",
				Images: []string{imageBase64},
			},
		},
		MaxTokens: 100,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	// Should mention red
	assert.Contains(t, response.Content, "red")

	t.Logf("Vision response: %s", response.Content)
}

// TestOpenAI_ErrorHandling tests error scenarios
func TestOpenAI_ErrorHandling(t *testing.T) {
	// Test with invalid API key
	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", "invalid-key", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Chat(ctx, llm.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{Role: "user", Content: "Hello"},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication")
}

// TestOpenAI_Performance benchmarks performance
func BenchmarkOpenAI_Latency(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("openai", apiKey, logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := client.Chat(ctx, llm.ChatRequest{
			Model: "gpt-4o-mini",
			Messages: []llm.Message{
				{Role: "user", Content: "Say hello"},
			},
			MaxTokens: 50,
		})
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(time.Since(start).Milliseconds()), "ms/latency")
	}
}

// generateLongText generates text of approximately n tokens
func generateLongText(tokens int) string {
	// Approximate 0.75 tokens per word
	words := tokens / 3 // Each word about 3 chars on average

	paragraph := "This is a comprehensive analysis of artificial intelligence. "
	paragraph += "Machine learning has revolutionized many industries. "
	paragraph += "Deep learning models achieve remarkable results. "

	var result string
	for len(result) < words*4 {
		result += paragraph
	}

	return result
}
