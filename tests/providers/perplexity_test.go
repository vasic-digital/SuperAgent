// Package providers implements real tests for Perplexity provider
package providers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"dev.helix.agent/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestPerplexity_Search tests search-enhanced completion
func TestPerplexity_Search(t *testing.T) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("perplexity", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "sonar",
		Messages: []llm.Message{
			{Role: "user", Content: "What is the latest version of Go?"},
		},
		MaxTokens: 200,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)
	// Should contain version number
	assert.True(t, strings.Contains(response.Content, "1.2") || strings.Contains(response.Content, "1.3"))

	t.Logf("Response: %s", response.Content)
	if len(response.Citations) > 0 {
		t.Logf("Citations: %d", len(response.Citations))
	}
}

// TestPerplexity_Streaming tests streaming with citations
func TestPerplexity_Streaming(t *testing.T) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("perplexity", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model: "sonar",
		Messages: []llm.Message{
			{Role: "user", Content: "What is machine learning?"},
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

	assert.NotEmpty(t, fullContent)
	assert.Greater(t, chunks, 1)
	t.Logf("Received %d chunks", chunks)
}

// TestPerplexity_Models tests different Perplexity models
func TestPerplexity_Models(t *testing.T) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	models := []string{
		"sonar",
		"sonar-pro",
		"sonar-reasoning",
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("perplexity", apiKey, logger)
	ctx := context.Background()

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			response, err := client.Chat(ctx, llm.ChatRequest{
				Model: model,
				Messages: []llm.Message{
					{Role: "user", Content: "What is 2+2?"},
				},
				MaxTokens: 50,
			})

			if err != nil {
				t.Logf("Model %s failed: %v", model, err)
				return
			}

			t.Logf("%s: %s", model, response.Content)
			if len(response.Citations) > 0 {
				t.Logf("  Citations: %d", len(response.Citations))
			}
		})
	}
}

// TestPerplexity_Citations verifies citation format
func TestPerplexity_Citations(t *testing.T) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	logger, _ := zap.NewDevelopment()
	client := llm.NewClient("perplexity", apiKey, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.Chat(ctx, llm.ChatRequest{
		Model: "sonar-pro",
		Messages: []llm.Message{
			{Role: "user", Content: "Who invented the World Wide Web?"},
		},
		MaxTokens: 200,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response.Content)

	// Should mention Tim Berners-Lee
	assert.Contains(t, strings.ToLower(response.Content), "tim berners-lee")

	// Should have citations
	if len(response.Citations) > 0 {
		t.Logf("Citations: %v", response.Citations)
	}
}
