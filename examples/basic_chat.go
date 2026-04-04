//go:build ignore

// Package main demonstrates basic chat with multiple providers
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"dev.helix.agent/internal/llm"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fmt.Println("=== HelixAgent Basic Chat Example ===")
	fmt.Println()

	// Check available providers
	providers := detectAvailableProviders()
	if len(providers) == 0 {
		fmt.Println("No API keys configured!")
		fmt.Println("Please set at least one of:")
		fmt.Println("  - OPENAI_API_KEY")
		fmt.Println("  - ANTHROPIC_API_KEY")
		fmt.Println("  - DEEPSEEK_API_KEY")
		fmt.Println("  - GROQ_API_KEY")
		return
	}

	fmt.Printf("Available providers: %v\n", providers)
	fmt.Println()

	// Use first available provider
	provider := providers[0]
	fmt.Printf("Using provider: %s\n", provider.Name)
	fmt.Println("Type 'quit' to exit")
	fmt.Println("Type '/switch <provider>' to change provider")
	fmt.Println()

	client := llm.NewClient(provider.Type, provider.Key, logger)

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		if strings.HasPrefix(input, "/switch ") {
			newProvider := strings.TrimPrefix(input, "/switch ")
			if p := findProvider(providers, newProvider); p != nil {
				provider = *p
				client = llm.NewClient(provider.Type, provider.Key, logger)
				fmt.Printf("Switched to: %s\n", provider.Name)
			} else {
				fmt.Printf("Unknown provider: %s\n", newProvider)
			}
			continue
		}

		if input == "" {
			continue
		}

		// Get response
		start := time.Now()
		response, err := client.Chat(ctx, llm.ChatRequest{
			Model:       provider.DefaultModel,
			Messages:    []llm.Message{{Role: "user", Content: input}},
			MaxTokens:   500,
			Temperature: 0.7,
		})

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		latency := time.Since(start)
		fmt.Printf("Assistant: %s\n", response.Content)
		fmt.Printf("          [%.2fs | %d tokens]\n", latency.Seconds(), response.Usage.TotalTokens)
		fmt.Println()
	}
}

type ProviderInfo struct {
	Name         string
	Type         string
	Key          string
	DefaultModel string
}

func detectAvailableProviders() []ProviderInfo {
	var providers []ProviderInfo

	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "OpenAI",
			Type:         "openai",
			Key:          key,
			DefaultModel: "gpt-4o-mini",
		})
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Anthropic",
			Type:         "anthropic",
			Key:          key,
			DefaultModel: "claude-3-5-haiku-20241022",
		})
	}
	if key := os.Getenv("DEEPSEEK_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "DeepSeek",
			Type:         "deepseek",
			Key:          key,
			DefaultModel: "deepseek-chat",
		})
	}
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Groq",
			Type:         "groq",
			Key:          key,
			DefaultModel: "llama-3.1-8b-instant",
		})
	}
	if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Mistral",
			Type:         "mistral",
			Key:          key,
			DefaultModel: "mistral-small-latest",
		})
	}
	if key := os.Getenv("COHERE_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Cohere",
			Type:         "cohere",
			Key:          key,
			DefaultModel: "command-r",
		})
	}
	if key := os.Getenv("PERPLEXITY_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Perplexity",
			Type:         "perplexity",
			Key:          key,
			DefaultModel: "sonar",
		})
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		providers = append(providers, ProviderInfo{
			Name:         "Gemini",
			Type:         "gemini",
			Key:          key,
			DefaultModel: "gemini-2.0-flash-exp",
		})
	}

	return providers
}

func findProvider(providers []ProviderInfo, name string) *ProviderInfo {
	for _, p := range providers {
		if strings.EqualFold(p.Name, name) {
			return &p
		}
	}
	return nil
}
