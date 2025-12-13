// Package main demonstrates basic usage of the AI toolkit library.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/superagent/toolkit/pkg/toolkit"
)

func main() {
	// Initialize the toolkit
	tk := toolkit.NewToolkit()

	// Set up logging
	tk.SetLogger(log.Default())

	fmt.Println("=== AI Toolkit Basic Usage Example ===")

	// List available providers and agents
	fmt.Println("\nAvailable Providers:")
	providers := tk.ListProviders()
	for _, provider := range providers {
		fmt.Printf("  - %s\n", provider)
	}

	fmt.Println("\nAvailable Agents:")
	agents := tk.ListAgents()
	for _, agent := range agents {
		fmt.Printf("  - %s\n", agent)
	}

	// Example: Create a provider configuration
	providerConfig := map[string]interface{}{
		"name":     "siliconflow",
		"api_key":  "your-api-key-here", // Replace with actual API key
		"base_url": "https://api.siliconflow.com",
		"timeout":  30000,
		"retries":  3,
	}

	// Validate provider configuration
	fmt.Println("\n=== Validating Provider Configuration ===")
	if err := tk.ValidateProviderConfig("siliconflow", providerConfig); err != nil {
		log.Printf("Provider config validation failed: %v", err)
		fmt.Println("Note: This is expected if the provider is not fully implemented")
	} else {
		fmt.Println("Provider configuration is valid")
	}

	// Example: Create an agent configuration
	agentConfig := map[string]interface{}{
		"name":        "my-assistant",
		"description": "A helpful AI assistant",
		"provider":    "siliconflow",
		"model":       "deepseek-chat",
		"max_tokens":  4096,
		"temperature": 0.7,
		"timeout":     30000,
		"retries":     3,
	}

	// Build agent configuration
	fmt.Println("\n=== Building Agent Configuration ===")
	builtConfig, err := tk.BuildConfig("agent", agentConfig)
	if err != nil {
		log.Printf("Failed to build agent config: %v", err)
		return
	}

	agentCfg, ok := builtConfig.(*toolkit.AgentConfig)
	if !ok {
		log.Println("Invalid config type")
		return
	}

	fmt.Printf("Agent Config: %+v\n", agentCfg)

	// Validate agent configuration
	if err := tk.ValidateAgentConfig("generic", agentCfg); err != nil {
		log.Printf("Agent config validation failed: %v", err)
	} else {
		fmt.Println("Agent configuration is valid")
	}

	// Example: Execute a simple task (this would fail without proper provider setup)
	fmt.Println("\n=== Executing Task (Demo) ===")
	ctx := context.Background()
	task := "Say hello and introduce yourself in one sentence."

	// Note: This will likely fail as we don't have real providers configured
	result, err := tk.ExecuteTask(ctx, "generic", task, agentCfg)
	if err != nil {
		log.Printf("Task execution failed (expected): %v", err)
		fmt.Println("Note: Task execution requires properly configured providers")
	} else {
		fmt.Printf("Task Result: %s\n", result)
	}

	// Example: Discover models (this would also require real providers)
	fmt.Println("\n=== Discovering Models (Demo) ===")
	models, err := tk.DiscoverModels(ctx)
	if err != nil {
		log.Printf("Model discovery failed (expected): %v", err)
		fmt.Println("Note: Model discovery requires properly configured providers")
	} else {
		fmt.Printf("Discovered %d models\n", len(models))
		for _, model := range models {
			fmt.Printf("  - %s (%s): %s\n", model.Name, model.ID, model.Description)
		}
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("To run this with real providers, configure your API keys and ensure providers are properly implemented.")
}
