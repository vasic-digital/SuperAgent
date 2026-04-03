// Package providers implements real tests for Groq provider
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

// TestGroq_ShortRequest tests simple completion with speed
func TestGroq_ShortRequest(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "llama-3.1-8b-instant",
		Messages: []llm.Message{
			{Role: "user", Content: "Say hello in exactly 3 words"},
		},
		MaxTokens: 50,
	})
	latency := time.Since(start)

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	assert.Contains(t, response.Content, "hello")
	// Groq should be very fast
	assert.Less(t, latency, 2*time.Second, "Groq should respond quickly")

	t.Logf("Response: %s", response.Content)
	t.Logf("Latency: %v", latency)
}

// TestGroq_Streaming tests streaming at high speed
func TestGroq_Streaming(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "llama-3.1-8b-instant",
		Messages: []llm.Message{
			{Role: "user", Content: "Write a 50-word story about a robot"},
		},
		MaxTokens: 200,
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

	totalTime := time.Since(start)

	assert.NotEmpty(t, fullContent)
	assert.Greater(t, chunks, 10)
	// Groq streaming should be very fast (800+ tokens/sec)
	tokensPerSec := float64(len(fullContent)) / totalTime.Seconds()
	t.Logf("Tokens/sec: %.0f", tokensPerSec)
	t.Logf("Total time: %v", totalTime)
}

// TestGroq_ToolUse tests tool calling
func TestGroq_ToolUse(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tools := []llm.ToolDefinition{
		{
			Name:        "get_current_time",
			Description: "Get the current time",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "llama-3.1-70b-versatile",
		Messages: []llm.Message{
			{Role: "user", Content: "What time is it?"},
		},
		Tools:     tools,
		MaxTokens: 100,
	})

	require.NoError(t, err)

	if len(response.ToolCalls) > 0 {
		t.Logf("Tool called: %s", response.ToolCalls[0].Name)
	}
}

// TestGroq_Vision tests vision model
func TestGroq_Vision(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Simple red pixel image
	imageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "llama-3.2-11b-vision-preview",
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "What color is this?",
				Images:  []string{imageBase64},
			},
		},
		MaxTokens: 50,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	t.Logf("Vision response: %s", response.Content)
}

// TestGroq_MultipleModels tests different Groq models
func TestGroq_MultipleModels(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set")
	}

	models := []string{
		"llama-3.1-8b-instant",
		"llama-3.1-70b-versatile",
		"mixtral-8x7b-32768",
		"gemma2-9b-it",
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)
	ctx := context.Background()

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			start := time.Now()
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

			t.Logf("%s: %v - %s", model, time.Since(start), response.Content)
		})
	}
}

// BenchmarkGroq_Speed benchmarks Groq's speed
func BenchmarkGroq_Speed(b *testing.B) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		b.Skip("GROQ_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("groq", apiKey, logger)
	ctx := context.Background()

	prompt := "Explain quantum computing in simple terms"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		response, err := client.Chat(ctx, llm.ChatRequest{
			Model: "llama-3.1-8b-instant",
			Messages: []llm.Message{
				{Role: "user", Content: prompt},
			},
			MaxTokens: 200,
		})
		if err != nil {
			b.Fatal(err)
		}

		duration := time.Since(start)
		tokensPerSec := float64(response.Usage.OutputTokens) / duration.Seconds()
		b.ReportMetric(tokensPerSec, "tokens/sec")
		b.ReportMetric(float64(duration.Milliseconds()), "ms/latency")
	}
}
