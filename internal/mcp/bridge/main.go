// Package bridge main - Entry point for the MCP SSE Bridge
package bridge

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// Main is the entry point for the bridge when run as a standalone binary
func Main() {
	// Parse configuration from environment
	config := DefaultConfig()

	// Port configuration
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	// MCP command configuration
	config.MCPCommand = os.Getenv("MCP_COMMAND")
	if config.MCPCommand == "" {
		fmt.Fprintln(os.Stderr, "Error: MCP_COMMAND environment variable is required")
		fmt.Fprintln(os.Stderr, "Example: MCP_COMMAND=\"npx @modelcontextprotocol/server-filesystem /allowed/path\"")
		os.Exit(1)
	}

	// Create bridge
	bridge := New(config)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
	}()

	// Start the bridge
	fmt.Println("Starting MCP SSE Bridge...")
	if err := bridge.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Bridge error: %v\n", err)
		os.Exit(1)
	}
}
