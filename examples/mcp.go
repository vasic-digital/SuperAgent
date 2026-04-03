// Package main demonstrates MCP (Model Context Protocol) usage
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"dev.helix.agent/internal/mcp"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fmt.Println("=== HelixAgent MCP Example ===")
	fmt.Println()

	// Create MCP client
	client := mcp.NewClient(&mcp.Config{
		Logger:  logger,
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// Example 1: Connect to stdio-based MCP server
	fmt.Println("--- MCP stdio Server ---")

	// Example: Connect to a filesystem MCP server
	stdioServer, err := client.ConnectStdio(ctx, "npx", []string{
		"-y", "@modelcontextprotocol/server-filesystem",
		"/tmp", // Allow access to /tmp
	})
	if err != nil {
		fmt.Printf("Could not connect to filesystem MCP: %v\n", err)
		fmt.Println("Note: Requires Node.js and npx installed")
	} else {
		defer stdioServer.Close()

		// List available tools
		tools, err := stdioServer.ListTools(ctx)
		if err != nil {
			fmt.Printf("Error listing tools: %v\n", err)
		} else {
			fmt.Printf("Available tools (%d):\n", len(tools))
			for _, tool := range tools {
				fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			}
		}

		// Call a tool
		fmt.Println("\nCalling read_file tool...")
		result, err := stdioServer.CallTool(ctx, "read_file", map[string]interface{}{
			"path": "/tmp/mcp_test.txt",
		})
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Result: %v\n", result)
		}
	}
	fmt.Println()

	// Example 2: Connect to HTTP-based MCP server
	fmt.Println("--- MCP HTTP Server ---")

	mcpURL := os.Getenv("MCP_SERVER_URL")
	if mcpURL == "" {
		mcpURL = "http://localhost:3000/mcp"
	}

	httpServer, err := client.ConnectHTTP(ctx, mcpURL)
	if err != nil {
		fmt.Printf("Could not connect to HTTP MCP at %s: %v\n", mcpURL, err)
		fmt.Println("Set MCP_SERVER_URL environment variable to test")
	} else {
		defer httpServer.Close()

		// Get server info
		info, err := httpServer.GetServerInfo(ctx)
		if err != nil {
			fmt.Printf("Error getting server info: %v\n", err)
		} else {
			fmt.Printf("Server: %s v%s\n", info.Name, info.Version)
		}

		// List and call resources
		resources, err := httpServer.ListResources(ctx)
		if err != nil {
			fmt.Printf("Error listing resources: %v\n", err)
		} else {
			fmt.Printf("Available resources (%d):\n", len(resources))
			for _, r := range resources {
				fmt.Printf("  - %s (%s)\n", r.URI, r.MIMEType)
			}
		}
	}
	fmt.Println()

	// Example 3: MCP Registry - manage multiple servers
	fmt.Println("--- MCP Registry ---")

	registry := mcp.NewRegistry(&mcp.RegistryConfig{
		Logger: logger,
	})

	// Register servers
	registry.Register("filesystem", mcp.ServerConfig{
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
	})

	registry.Register("sqlite", mcp.ServerConfig{
		Type:    "stdio",
		Command: "uvx",
		Args:    []string{"mcp-server-sqlite", "/tmp/test.db"},
	})

	fmt.Printf("Registered %d MCP servers\n", len(registry.ListServers()))

	// Initialize all servers
	start := time.Now()
	results := registry.InitializeAll(ctx)
	fmt.Printf("Initialized %d/%d servers in %v\n",
		results.Successful, results.Total, time.Since(start))

	for name, err := range results.Errors {
		fmt.Printf("  %s failed: %v\n", name, err)
	}

	// Aggregate tools from all servers
	fmt.Println("\nAggregated tools from all servers:")
	allTools := registry.GetAllTools(ctx)
	for serverName, tools := range allTools {
		fmt.Printf("  %s (%d tools):\n", serverName, len(tools))
		for _, tool := range tools[:min(3, len(tools))] {
			fmt.Printf("    - %s\n", tool.Name)
		}
		if len(tools) > 3 {
			fmt.Printf("    ... and %d more\n", len(tools)-3)
		}
	}

	// Example 4: LLM with MCP tools
	fmt.Println("\n--- LLM with MCP Tools ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Set OPENAI_API_KEY to test LLM with MCP")
	} else {
		// Create LLM client with MCP tools
		llmClient := mcp.NewLLMClient(apiKey, registry, logger)

		response, err := llmClient.Chat(ctx, "List all files in /tmp directory")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("LLM Response:\n%s\n", response.Content)
			fmt.Printf("Tools used: %v\n", response.ToolsUsed)
		}
	}

	// Cleanup
	registry.ShutdownAll(ctx)
	fmt.Println("\nMCP servers shutdown")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
