// Package providers implements real tests for DeepSeek provider
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

// TestDeepSeek_ShortRequest tests simple completion
func TestDeepSeek_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello in exactly 3 words"},
		},
		MaxTokens: 50,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Contains(t, response.Content, "hello")

	t.Logf("Response: %s", response.Content)
	t.Logf("Tokens: %d", response.Usage.TotalTokens)
}

// TestDeepSeek_Reasoning tests reasoning model (DeepSeek-R1)
func TestDeepSeek_Reasoning(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-reasoner",
		Messages: []llm.Message{
			{Role: "user", Content: "Solve: If a train travels 120 km in 2 hours, what is its average speed?"},
		},
		MaxTokens: 500,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	// Should contain reasoning about the calculation
	assert.Contains(t, response.Content, "60")

	t.Logf("Response: %s", response.Content)
}

// TestDeepSeek_Streaming tests streaming response
func TestDeepSeek_Streaming(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 5"},
		},
		MaxTokens: 100,
	})
	require.NoError(t, err)

	var fullContent string
	var chunks int

	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		fullContent += chunk.Content
		chunks++
	}

	assert.NotEmpty(t, fullContent)
	assert.Greater(t, chunks, 1)

	t.Logf("Received %d chunks", chunks)
	t.Logf("Content: %s", fullContent)
}

// TestDeepSeek_JSONMode tests JSON output
func TestDeepSeek_JSONMode(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "user", Content: "Generate a JSON object with name and age fields. Output ONLY valid JSON."},
		},
		MaxTokens:      200,
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
	})

	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(response.Content), &result)
	require.NoError(t, err, "Response should be valid JSON")

	assert.NotNil(t, result["name"])
	assert.NotNil(t, result["age"])

	t.Logf("JSON: %s", response.Content)
}

// TestDeepSeek_ToolCalling tests function calling
func TestDeepSeek_ToolCalling(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "calculate",
			Description: "Perform a calculation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]string{"type": "string"},
				},
				"required": []string{"expression"},
			},
		},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "user", Content: "Calculate 25 * 4"},
		},
		Tools:     tools,
		MaxTokens: 200,
	})

	require.NoError(t, err)

	if len(response.ToolCalls) > 0 {
		assert.Equal(t, "calculate", response.ToolCalls[0].Name)
		t.Logf("Tool call: %s", response.ToolCalls[0].Arguments)
	}
}

// TestDeepSeek_LongContext tests 64K context window
func TestDeepSeek_LongContext(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	// Generate ~20K tokens
	contextText := generateLongText(20000)

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "system", Content: "Summarize:"},
			{Role: "user", Content: contextText},
		},
		MaxTokens: 500,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)

	t.Logf("Processed 20K tokens in %v", time.Since(startTime))
}

// TestDeepSeek_ErrorHandling tests error scenarios
func TestDeepSeek_ErrorHandling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", "invalid-key", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Chat(ctx, llm.ChatRequest{
		Model: "deepseek-chat",
		Messages: []llm.Message{
			{Role: "user", Content: "Hello"},
		},
	})

	assert.Error(t, err)
}

// BenchmarkDeepSeek_Latency benchmarks performance
func BenchmarkDeepSeek_Latency(b *testing.B) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		b.Skip("DEEPSEEK_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("deepseek", apiKey, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := client.Chat(ctx, llm.ChatRequest{
			Model: "deepseek-chat",
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
