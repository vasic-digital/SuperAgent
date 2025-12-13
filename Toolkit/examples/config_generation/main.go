// Package main demonstrates configuration generation for multiple agents and providers.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/superagent/toolkit/pkg/toolkit"
)

func main() {
	fmt.Println("=== AI Toolkit Configuration Generation Example ===")

	// Initialize the toolkit
	tk := toolkit.NewToolkit()
	tk.SetLogger(log.Default())

	// Define multiple agent configurations
	agentConfigs := []map[string]interface{}{
		{
			"name":        "code-assistant",
			"description": "Specialized in code generation and review",
			"provider":    "siliconflow",
			"model":       "deepseek-coder",
			"max_tokens":  8192,
			"temperature": 0.3,
			"timeout":     60000,
			"retries":     5,
		},
		{
			"name":        "creative-writer",
			"description": "Creative writing and content generation",
			"provider":    "openrouter",
			"model":       "anthropic/claude-3-haiku",
			"max_tokens":  4096,
			"temperature": 0.8,
			"timeout":     45000,
			"retries":     3,
		},
		{
			"name":        "data-analyst",
			"description": "Data analysis and visualization",
			"provider":    "nvidia",
			"model":       "meta/llama-3.1-70b-instruct",
			"max_tokens":  16384,
			"temperature": 0.1,
			"timeout":     120000,
			"retries":     3,
		},
	}

	// Define provider configurations
	providerConfigs := []map[string]interface{}{
		{
			"name":       "siliconflow",
			"api_key":    os.Getenv("SILICONFLOW_API_KEY"), // Load from environment
			"base_url":   "https://api.siliconflow.com",
			"timeout":    30000,
			"retries":    3,
			"rate_limit": 60,
		},
		{
			"name":       "openrouter",
			"api_key":    os.Getenv("OPENROUTER_API_KEY"),
			"base_url":   "https://openrouter.ai/api/v1",
			"timeout":    30000,
			"retries":    3,
			"rate_limit": 50,
		},
		{
			"name":       "nvidia",
			"api_key":    os.Getenv("NVIDIA_API_KEY"),
			"base_url":   "https://integrate.api.nvidia.com/v1",
			"timeout":    30000,
			"retries":    3,
			"rate_limit": 100,
		},
	}

	// Generate and validate configurations
	fmt.Println("Generating agent configurations...")

	var validAgents []toolkit.AgentConfig
	for i, config := range agentConfigs {
		fmt.Printf("\n--- Agent %d: %s ---\n", i+1, config["name"])

		// Build configuration
		builtConfig, err := tk.BuildConfig("agent", config)
		if err != nil {
			log.Printf("Failed to build agent config: %v", err)
			continue
		}

		agentCfg, ok := builtConfig.(*toolkit.AgentConfig)
		if !ok {
			log.Println("Invalid config type")
			continue
		}

		// Validate configuration
		if err := tk.ValidateAgentConfig("generic", agentCfg); err != nil {
			log.Printf("Agent config validation failed: %v", err)
			continue
		}

		fmt.Printf("✓ Agent configuration validated: %s\n", agentCfg.Name)
		validAgents = append(validAgents, *agentCfg)
	}

	fmt.Println("\nGenerating provider configurations...")

	var validProviders []toolkit.ProviderConfig
	for i, config := range providerConfigs {
		fmt.Printf("\n--- Provider %d: %s ---\n", i+1, config["name"])

		// Build configuration
		builtConfig, err := tk.BuildConfig("provider", config)
		if err != nil {
			log.Printf("Failed to build provider config: %v", err)
			continue
		}

		providerCfg, ok := builtConfig.(*toolkit.ProviderConfig)
		if !ok {
			log.Println("Invalid config type")
			continue
		}

		// Validate configuration (skip if API key is missing)
		if providerCfg.APIKey == "" {
			fmt.Printf("⚠ Skipping validation for %s (API key not set)\n", providerCfg.Name)
		} else {
			if err := tk.ValidateProviderConfig(providerCfg.Name, config); err != nil {
				log.Printf("Provider config validation failed: %v", err)
				continue
			}
			fmt.Printf("✓ Provider configuration validated: %s\n", providerCfg.Name)
		}

		validProviders = append(validProviders, *providerCfg)
	}

	// Generate configuration files
	fmt.Println("\n=== Generating Configuration Files ===")

	// Create configs directory
	if err := os.MkdirAll("configs", 0755); err != nil {
		log.Printf("Failed to create configs directory: %v", err)
		return
	}

	// Generate agents.json
	agentsFile := "configs/agents.json"
	agentsData := map[string]interface{}{
		"agents": validAgents,
		"metadata": map[string]interface{}{
			"version":      "1.0",
			"generated_at": "2024-01-01T00:00:00Z",
			"description":  "Generated agent configurations",
		},
	}

	if err := writeJSONFile(agentsFile, agentsData); err != nil {
		log.Printf("Failed to write agents config: %v", err)
	} else {
		fmt.Printf("✓ Generated %s\n", agentsFile)
	}

	// Generate providers.json
	providersFile := "configs/providers.json"
	providersData := map[string]interface{}{
		"providers": validProviders,
		"metadata": map[string]interface{}{
			"version":      "1.0",
			"generated_at": "2024-01-01T00:00:00Z",
			"description":  "Generated provider configurations",
		},
	}

	if err := writeJSONFile(providersFile, providersData); err != nil {
		log.Printf("Failed to write providers config: %v", err)
	} else {
		fmt.Printf("✓ Generated %s\n", providersFile)
	}

	// Generate deployment configuration
	deploymentFile := "configs/deployment.json"
	deploymentData := map[string]interface{}{
		"deployment": map[string]interface{}{
			"name":    "multi-agent-system",
			"version": "1.0.0",
			"agents": []map[string]interface{}{
				{
					"name":     "code-assistant",
					"enabled":  true,
					"replicas": 2,
					"resources": map[string]interface{}{
						"cpu":    "500m",
						"memory": "1Gi",
					},
				},
				{
					"name":     "creative-writer",
					"enabled":  true,
					"replicas": 1,
					"resources": map[string]interface{}{
						"cpu":    "250m",
						"memory": "512Mi",
					},
				},
				{
					"name":     "data-analyst",
					"enabled":  true,
					"replicas": 1,
					"resources": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "2Gi",
					},
				},
			},
			"providers": []map[string]interface{}{
				{
					"name":     "siliconflow",
					"enabled":  true,
					"priority": 1,
				},
				{
					"name":     "openrouter",
					"enabled":  true,
					"priority": 2,
				},
				{
					"name":     "nvidia",
					"enabled":  true,
					"priority": 3,
				},
			},
		},
		"metadata": map[string]interface{}{
			"version":      "1.0",
			"generated_at": "2024-01-01T00:00:00Z",
			"description":  "Deployment configuration for multi-agent system",
		},
	}

	if err := writeJSONFile(deploymentFile, deploymentData); err != nil {
		log.Printf("Failed to write deployment config: %v", err)
	} else {
		fmt.Printf("✓ Generated %s\n", deploymentFile)
	}

	// Display summary
	fmt.Println("\n=== Configuration Generation Summary ===")
	fmt.Printf("Total agents configured: %d\n", len(validAgents))
	fmt.Printf("Total providers configured: %d\n", len(validProviders))
	fmt.Printf("Configuration files generated in: ./configs/\n")

	fmt.Println("\n=== Next Steps ===")
	fmt.Println("1. Set your API keys as environment variables:")
	fmt.Println("   export SILICONFLOW_API_KEY=your_key_here")
	fmt.Println("   export OPENROUTER_API_KEY=your_key_here")
	fmt.Println("   export NVIDIA_API_KEY=your_key_here")
	fmt.Println("2. Review and customize the generated configuration files")
	fmt.Println("3. Run integration tests to validate your setup")
}

// writeJSONFile writes data to a JSON file with proper formatting.
func writeJSONFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
