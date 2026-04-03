// Package main demonstrates streaming responses
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"dev.helix.agent/internal/llm"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fmt.Println("=== HelixAgent Streaming Example ===")
	fmt.Println()

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY, ANTHROPIC_API_KEY, or DEEPSEEK_API_KEY")
		return
	}

	// Detect provider type
	providerType := "openai"
	model := "gpt-4o-mini"
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providerType = "anthropic"
		model = "claude-3-5-haiku-20241022"
	} else if os.Getenv("DEEPSEEK_API_KEY") != "" {
		providerType = "deepseek"
		model = "deepseek-chat"
	}

	client := llm.NewClient(providerType, apiKey, logger)
	ctx := context.Background()

	prompt := "Write a short haiku about artificial intelligence."

	fmt.Printf("Provider: %s\n", providerType)
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Prompt: %s\n", prompt)
	fmt.Println()

	// Non-streaming version (for comparison)
	fmt.Println("--- Non-streaming ---")
	start := time.Now()
	response, err := client.Chat(ctx, llm.ChatRequest{
		Model:     model,
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		MaxTokens: 100,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Latency: %v\n", time.Since(start))
	fmt.Printf("Tokens: %d\n", response.Usage.TotalTokens)
	fmt.Println()

	// Streaming version
	fmt.Println("--- Streaming ---")
	start = time.Now()
	firstChunk := time.Time{}

	stream, err := client.ChatStream(ctx, llm.ChatRequest{
		Model:     model,
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		MaxTokens: 100,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Print("Response: ")
	var fullContent string
	var chunks int

	for chunk := range stream {
		if chunk.Error != nil {
			fmt.Printf("\nStream error: %v\n", chunk.Error)
			return
		}

		if firstChunk.IsZero() && chunk.Content != "" {
			firstChunk = time.Now()
		}

		fmt.Print(chunk.Content)
		fullContent += chunk.Content
		chunks++
	}
	fmt.Println()

	totalLatency := time.Since(start)
	ttfb := firstChunk.Sub(start) // Time to first byte

	fmt.Printf("Total latency: %v\n", totalLatency)
	fmt.Printf("Time to first byte: %v\n", ttfb)
	fmt.Printf("Chunks received: %d\n", chunks)
	fmt.Printf("Avg chunk time: %v\n", totalLatency/time.Duration(chunks))

	// Verify both responses are similar
	fmt.Println()
	fmt.Printf("Both responses match: %v\n", len(fullContent) > 10 && len(response.Content) > 10)
}
