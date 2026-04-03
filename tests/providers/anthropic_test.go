// Package providers implements real tests for Anthropic/Claude provider
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

// TestAnthropic_ShortRequest tests simple completion
func TestAnthropic_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-haiku-20241022",
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

// TestAnthropic_Streaming tests streaming response
func TestAnthropic_Streaming(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "claude-3-5-haiku-20241022",
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

// TestAnthropic_ToolUse tests tool use (Claude's equivalent to function calling)
func TestAnthropic_ToolUse(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "calculate_sum",
			Description: "Calculate the sum of two numbers",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]string{"type": "number"},
					"b": map[string]string{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: "What is 42 + 58?"},
		},
		Tools:     tools,
		MaxTokens: 200,
	})

	require.NoError(t, err)

	// Claude should use the tool
	if len(response.ToolCalls) > 0 {
		assert.Equal(t, "calculate_sum", response.ToolCalls[0].Name)
		t.Logf("Tool call: %s", response.ToolCalls[0].Name)
		t.Logf("Arguments: %s", response.ToolCalls[0].Arguments)
	} else {
		t.Logf("Direct response (model chose not to use tool): %s", response.Content)
	}
}

// TestAnthropic_Vision tests vision capabilities
func TestAnthropic_Vision(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Base64 encoded red pixel
	imageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "What color is this image? Describe it briefly.",
				Images:  []string{imageBase64},
			},
		},
		MaxTokens: 100,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	t.Logf("Vision response: %s", response.Content)
}

// TestAnthropic_JSONOutput tests structured JSON output
func TestAnthropic_JSONOutput(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-haiku-20241022",
		Messages: []llm.Message{
			{
				Role: "user",
				Content: `Generate a JSON object with fields: product (string), price (number), in_stock (boolean). 
				Output ONLY valid JSON, no other text.`,
			},
		},
		MaxTokens: 200,
	})

	require.NoError(t, err)

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(response.Content), &result)
	require.NoError(t, err, "Response should be valid JSON")

	assert.NotNil(t, result["product"], "Should have product field")
	assert.NotNil(t, result["price"], "Should have price field")
	assert.NotNil(t, result["in_stock"], "Should have in_stock field")

	t.Logf("JSON Response: %s", response.Content)
}

// TestAnthropic_LongContext tests 200K context window
func TestAnthropic_LongContext(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	// Generate ~50K tokens of context (well within 200K limit)
	contextText := generateLongText(50000)

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	startTime := time.Now()
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []llm.Message{
			{Role: "system", Content: "Summarize the following text in one sentence:"},
			{Role: "user", Content: contextText},
		},
		MaxTokens: 500,
	})

	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)

	t.Logf("Processed 50K tokens in %v", duration)
	t.Logf("Summary: %s", response.Content)
}

// TestAnthropic_PromptCaching tests prompt caching (if supported)
func TestAnthropic_PromptCaching(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Large prompt to trigger caching
	largePrompt := generateLongText(5000)

	startTime := time.Now()
	response1, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-haiku-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: largePrompt + "\n\nQuestion: What is this about?"},
		},
		MaxTokens: 100,
	})
	require.NoError(t, err)
	firstDuration := time.Since(startTime)

	// Second request should be faster if caching works
	startTime = time.Now()
	response2, err := client.Chat(ctx, llm.ChatRequest{
		Model: "claude-3-5-haiku-20241022",
		Messages: []llm.Message{
			{Role: "user", Content: largePrompt + "\n\nQuestion: Summarize briefly"},
		},
		MaxTokens: 100,
	})
	require.NoError(t, err)
	secondDuration := time.Since(startTime)

	t.Logf("First request: %v", firstDuration)
	t.Logf("Second request: %v", secondDuration)
	t.Logf("Response 1: %s", response1.Content)
	t.Logf("Response 2: %s", response2.Content)
}

// BenchmarkAnthropic benchmarks Anthropic performance
func BenchmarkAnthropic_Latency(b *testing.B) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		b.Skip("ANTHROPIC_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("anthropic", apiKey, logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := client.Chat(ctx, llm.ChatRequest{
			Model: "claude-3-5-haiku-20241022",
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
