// Package main provides a command-line interface for the AI toolkit.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/superagent/toolkit/agents/crush"
	"github.com/superagent/toolkit/agents/generic"
	"github.com/superagent/toolkit/agents/opencode"
	"github.com/superagent/toolkit/pkg/toolkit"
	"github.com/superagent/toolkit/providers/chutes"
	"github.com/superagent/toolkit/providers/claude"
	"github.com/superagent/toolkit/providers/nvidia"
	"github.com/superagent/toolkit/providers/openrouter"
	"github.com/superagent/toolkit/providers/siliconflow"
)

var (
	// Global variables
	tk     *toolkit.Toolkit
	logger *log.Logger

	// CLI flags
	verbose bool
)

func main() {
	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Parse global flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--verbose":
			verbose = true
			args = append(args[:i], args[i+1:]...)
			i--
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		}
	}

	// Initialize toolkit
	tk = toolkit.NewToolkit()
	logger = log.New(os.Stdout, "[TOOLKIT] ", log.LstdFlags)
	if verbose {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}
	tk.SetLogger(logger)

	// Register built-in providers and agents
	if err := siliconflow.Register(tk.GetProviderFactoryRegistry()); err != nil {
		logger.Printf("Failed to register siliconflow provider: %v", err)
	}
	if err := chutes.Register(tk.GetProviderFactoryRegistry()); err != nil {
		logger.Printf("Failed to register chutes provider: %v", err)
	}
	if err := claude.Register(tk.GetProviderFactoryRegistry()); err != nil {
		logger.Printf("Failed to register claude provider: %v", err)
	}
	if err := nvidia.Register(tk.GetProviderFactoryRegistry()); err != nil {
		logger.Printf("Failed to register nvidia provider: %v", err)
	}
	if err := openrouter.Register(tk.GetProviderFactoryRegistry()); err != nil {
		logger.Printf("Failed to register openrouter provider: %v", err)
	}

	if err := opencode.Register(tk.GetAgentFactoryRegistry()); err != nil {
		logger.Printf("Failed to register opencode agent: %v", err)
	}
	if err := crush.Register(tk.GetAgentFactoryRegistry()); err != nil {
		logger.Printf("Failed to register crush agent: %v", err)
	}
	if err := generic.Register(tk.GetAgentFactoryRegistry()); err != nil {
		logger.Printf("Failed to register generic agent: %v", err)
	}

	// Execute command
	switch command {
	case "list":
		handleList(args)
	case "execute":
		handleExecute(args)
	case "discover":
		handleDiscover(args)
	case "validate":
		handleValidate(args)
	case "config":
		handleConfig(args)
	case "test":
		handleTest(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AI Toolkit CLI - Manage providers and agents")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  toolkit [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [type]          List providers, agents, or all")
	fmt.Println("  execute <agent> <task>  Execute a task with an agent")
	fmt.Println("  discover [provider]  Discover models from providers")
	fmt.Println("  validate <type> <name> <config>  Validate configuration")
	fmt.Println("  config generate <type> <name>    Generate sample config")
	fmt.Println("  test                 Run integration tests")
	fmt.Println("  help                 Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose        Verbose output")
	fmt.Println("  -h, --help           Show help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  toolkit list all")
	fmt.Println("  toolkit execute generic 'Hello world'")
	fmt.Println("  toolkit discover siliconflow")
	fmt.Println("  toolkit config generate provider siliconflow")
}

// Command handlers

func handleList(args []string) {
	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	switch listType {
	case "providers", "provider", "p":
		listProviders()
	case "agents", "agent", "a":
		listAgents()
	case "all":
		listProviders()
		fmt.Println()
		listAgents()
	default:
		fmt.Printf("Unknown type: %s. Use 'providers', 'agents', or 'all'\n", listType)
		os.Exit(1)
	}
}

func handleExecute(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: toolkit execute <agent> <task>")
		os.Exit(1)
	}

	agentName := args[0]
	task := strings.Join(args[1:], " ")

	executeTask(agentName, task)
}

func handleDiscover(args []string) {
	if len(args) == 0 {
		discoverAllModels()
	} else {
		discoverProviderModels(args[0])
	}
}

func handleValidate(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: toolkit validate <type> <name> <config-file>")
		os.Exit(1)
	}

	configType := args[0]
	name := args[1]
	configPath := args[2]

	validateConfig(configType, name, configPath)
}

func handleConfig(args []string) {
	if len(args) < 3 || args[0] != "generate" {
		fmt.Println("Usage: toolkit config generate <type> <name>")
		os.Exit(1)
	}

	configType := args[1]
	name := args[2]

	generateSampleConfig(configType, name)
}

func handleTest(args []string) {
	fmt.Println("Running integration tests...")

	// Import and run the integration test suite
	fmt.Println("Note: For full integration testing, run:")
	fmt.Println("  go test ./examples/integration_test/")
	fmt.Println("  go run ./examples/integration_test/main.go")
}

// Implementation functions

func listProviders() {
	fmt.Println("=== Available Providers ===")
	providers := tk.ListProviders()

	if len(providers) == 0 {
		fmt.Println("No providers registered")
		return
	}

	for i, provider := range providers {
		fmt.Printf("%d. %s\n", i+1, provider)
	}
}

func listAgents() {
	fmt.Println("=== Available Agents ===")
	agents := tk.ListAgents()

	if len(agents) == 0 {
		fmt.Println("No agents registered")
		return
	}

	for i, agent := range agents {
		fmt.Printf("%d. %s\n", i+1, agent)

		// Show capabilities if verbose
		if verbose {
			if agent, err := tk.GetAgent(agent); err == nil {
				caps := agent.Capabilities()
				if len(caps) > 0 {
					fmt.Printf("   Capabilities: %s\n", strings.Join(caps, ", "))
				}
			}
		}
	}
}

func executeTask(agentName, task string) {
	fmt.Printf("Executing task with agent '%s'...\n", agentName)
	fmt.Printf("Task: %s\n\n", task)

	ctx := context.Background()

	// Create a basic config for the agent
	config := map[string]interface{}{
		"name":        fmt.Sprintf("cli-%s", agentName),
		"description": "CLI execution",
		"provider":    "siliconflow", // Default provider
		"model":       "deepseek-chat",
		"max_tokens":  1000,
		"temperature": 0.7,
	}

	result, err := tk.ExecuteTask(ctx, agentName, task, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing task: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Result:")
	fmt.Println(result)
}

func discoverAllModels() {
	fmt.Println("Discovering models from all providers...")

	ctx := context.Background()
	models, err := tk.DiscoverModels(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering models: %v\n", err)
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Println("No models discovered")
		return
	}

	fmt.Printf("Discovered %d models:\n\n", len(models))

	// Group by provider
	byProvider := make(map[string][]toolkit.ModelInfo)
	for _, model := range models {
		byProvider[model.Provider] = append(byProvider[model.Provider], model)
	}

	for provider, providerModels := range byProvider {
		fmt.Printf("Provider: %s\n", provider)
		for _, model := range providerModels {
			fmt.Printf("  - %s (%s): %s\n", model.Name, model.ID, model.Description)
			if verbose {
				fmt.Printf("    Category: %s, Context: %d tokens\n", model.Category, model.Capabilities.ContextWindow)
			}
		}
		fmt.Println()
	}
}

func discoverProviderModels(providerName string) {
	fmt.Printf("Discovering models from provider '%s'...\n", providerName)

	provider, err := tk.GetProvider(providerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Provider not found: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	models, err := provider.DiscoverModels(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering models: %v\n", err)
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Println("No models discovered")
		return
	}

	fmt.Printf("Discovered %d models:\n\n", len(models))
	for _, model := range models {
		fmt.Printf("- %s (%s): %s\n", model.Name, model.ID, model.Description)
		if verbose {
			fmt.Printf("  Category: %s, Context: %d tokens\n", model.Category, model.Capabilities.ContextWindow)
		}
	}
}

func validateConfig(configType, name, configPath string) {
	fmt.Printf("Validating %s configuration for '%s'...\n", configType, name)

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config JSON: %v\n", err)
		os.Exit(1)
	}

	// Validate based on type
	switch configType {
	case "provider":
		if err := tk.ValidateProviderConfig(name, config); err != nil {
			fmt.Fprintf(os.Stderr, "Provider config validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Provider configuration is valid")

	case "agent":
		// Build the config first
		builtConfig, err := tk.BuildConfig("agent", config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Agent config build failed: %v\n", err)
			os.Exit(1)
		}

		if err := tk.ValidateAgentConfig(name, builtConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Agent config validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Agent configuration is valid")

	default:
		fmt.Fprintf(os.Stderr, "Unknown config type: %s\n", configType)
		os.Exit(1)
	}
}

func generateSampleConfig(configType, name string) {
	var config interface{}

	switch configType {
	case "provider":
		config = generateProviderConfig(name)
	case "agent":
		config = generateAgentConfig(name)
	default:
		fmt.Fprintf(os.Stderr, "Unknown config type: %s\n", configType)
		os.Exit(1)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	filename := fmt.Sprintf("%s-%s-config.json", configType, name)
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated sample configuration: %s\n", filename)
}

func generateProviderConfig(name string) map[string]interface{} {
	baseConfig := map[string]interface{}{
		"name":       name,
		"api_key":    "your-api-key-here",
		"base_url":   "https://api.example.com",
		"timeout":    30000,
		"retries":    3,
		"rate_limit": 60,
	}

	// Provider-specific customizations
	switch name {
	case "siliconflow":
		baseConfig["base_url"] = "https://api.siliconflow.com"
	case "claude":
		baseConfig["base_url"] = "https://api.anthropic.com"
		baseConfig["version"] = "2023-06-01"
	case "openrouter":
		baseConfig["base_url"] = "https://openrouter.ai/api/v1"
	case "nvidia":
		baseConfig["base_url"] = "https://integrate.api.nvidia.com/v1"
	}

	return baseConfig
}

func generateAgentConfig(name string) map[string]interface{} {
	baseConfig := map[string]interface{}{
		"name":        fmt.Sprintf("my-%s-agent", name),
		"description": fmt.Sprintf("A %s agent", name),
		"provider":    "siliconflow",
		"model":       "deepseek-chat",
		"max_tokens":  4096,
		"temperature": 0.7,
		"timeout":     30000,
		"retries":     3,
	}

	// Agent-specific customizations
	switch name {
	case "generic":
		baseConfig["description"] = "A general-purpose AI assistant"
	case "code-review":
		baseConfig["description"] = "Specialized in code review and analysis"
		baseConfig["model"] = "deepseek-coder"
		baseConfig["temperature"] = 0.3
		baseConfig["focus_areas"] = []string{"security", "performance", "maintainability"}
		baseConfig["language"] = "go"
	}

	return baseConfig
}
