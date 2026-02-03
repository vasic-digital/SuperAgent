// Package mcp provides adapter types to bridge between HelixAgent's
// internal MCP types and the extracted digital.vasic.mcp module.
// This allows HelixAgent to use the generic MCP module while maintaining
// its existing API contracts.
package mcp

import (
	"context"
	"time"

	extclient "digital.vasic.mcp/pkg/client"
	extprotocol "digital.vasic.mcp/pkg/protocol"
	extregistry "digital.vasic.mcp/pkg/registry"
)

// Type re-exports from the extracted module for convenience

// Protocol types
type (
	Tool            = extprotocol.Tool
	ToolResult      = extprotocol.ToolResult
	Resource        = extprotocol.Resource
	ResourceContent = extprotocol.ResourceContent
	Prompt          = extprotocol.Prompt
	PromptMessage   = extprotocol.PromptMessage
	Request         = extprotocol.Request
	Response        = extprotocol.Response
	ContentBlock    = extprotocol.ContentBlock
)

// Client types
type (
	ClientConfig  = extclient.Config
	TransportType = extclient.TransportType
)

// Registry types
type Adapter = extregistry.Adapter

// Transport type constants
const (
	TransportStdio = extclient.TransportStdio
	TransportHTTP  = extclient.TransportHTTP
)

// ClientAdapter wraps the extracted MCP client
type ClientAdapter struct {
	client extclient.Client
	config extclient.Config
}

// NewClientAdapter creates a new adapter wrapping an external MCP client
func NewClientAdapter(config extclient.Config) (*ClientAdapter, error) {
	var client extclient.Client
	var err error

	switch config.Transport {
	case extclient.TransportStdio:
		client, err = extclient.NewStdioClient(config)
	case extclient.TransportHTTP:
		client, err = extclient.NewHTTPClient(config)
	default:
		client, err = extclient.NewStdioClient(config)
	}

	if err != nil {
		return nil, err
	}

	return &ClientAdapter{
		client: client,
		config: config,
	}, nil
}

// Initialize performs the MCP initialization handshake
func (a *ClientAdapter) Initialize(ctx context.Context) (*extprotocol.InitializeResult, error) {
	return a.client.Initialize(ctx)
}

// ListTools returns the tools available on the MCP server
func (a *ClientAdapter) ListTools(ctx context.Context) ([]extprotocol.Tool, error) {
	return a.client.ListTools(ctx)
}

// CallTool invokes a tool on the MCP server
func (a *ClientAdapter) CallTool(
	ctx context.Context,
	name string,
	args map[string]interface{},
) (*extprotocol.ToolResult, error) {
	return a.client.CallTool(ctx, name, args)
}

// ListResources returns the resources available on the MCP server
func (a *ClientAdapter) ListResources(ctx context.Context) ([]extprotocol.Resource, error) {
	return a.client.ListResources(ctx)
}

// ReadResource reads a resource from the MCP server
func (a *ClientAdapter) ReadResource(
	ctx context.Context,
	uri string,
) (*extprotocol.ResourceContent, error) {
	return a.client.ReadResource(ctx, uri)
}

// ListPrompts returns the prompts available on the MCP server
func (a *ClientAdapter) ListPrompts(ctx context.Context) ([]extprotocol.Prompt, error) {
	return a.client.ListPrompts(ctx)
}

// GetPrompt retrieves a prompt with the given arguments
func (a *ClientAdapter) GetPrompt(
	ctx context.Context,
	name string,
	args map[string]string,
) ([]extprotocol.PromptMessage, error) {
	return a.client.GetPrompt(ctx, name, args)
}

// Close shuts down the client and cleans up resources
func (a *ClientAdapter) Close() error {
	return a.client.Close()
}

// RegistryAdapter wraps the extracted MCP registry
type RegistryAdapter struct {
	registry *extregistry.Registry
}

// NewRegistryAdapter creates a new adapter wrapping the external registry
func NewRegistryAdapter() *RegistryAdapter {
	return &RegistryAdapter{
		registry: extregistry.New(),
	}
}

// Register registers an adapter
func (a *RegistryAdapter) Register(adapter extregistry.Adapter) error {
	return a.registry.Register(adapter)
}

// Get returns an adapter by name
func (a *RegistryAdapter) Get(name string) (extregistry.Adapter, bool) {
	return a.registry.Get(name)
}

// List returns all registered adapter names
func (a *RegistryAdapter) List() []string {
	return a.registry.List()
}

// Unregister removes an adapter
func (a *RegistryAdapter) Unregister(name string) error {
	return a.registry.Unregister(name)
}

// StartAll starts all registered adapters
func (a *RegistryAdapter) StartAll(ctx context.Context) error {
	return a.registry.StartAll(ctx)
}

// StopAll stops all registered adapters
func (a *RegistryAdapter) StopAll(ctx context.Context) error {
	return a.registry.StopAll(ctx)
}

// HealthCheckAll checks the health of all adapters
func (a *RegistryAdapter) HealthCheckAll(ctx context.Context) map[string]error {
	return a.registry.HealthCheckAll(ctx)
}

// DefaultClientConfig returns a default MCP client configuration
func DefaultClientConfig() ClientConfig {
	return extclient.DefaultConfig()
}

// ServerConfig for internal use (maps to HelixAgent's existing config)
type ServerConfig struct {
	Name    string
	Command []string
	Env     map[string]string
	Timeout time.Duration
}

// ToExternalConfig converts internal server config to external client config
func (c *ServerConfig) ToExternalConfig() extclient.Config {
	cfg := extclient.DefaultConfig()
	cfg.Transport = extclient.TransportStdio
	cfg.ClientName = c.Name
	if len(c.Command) > 0 {
		cfg.ServerCommand = c.Command[0]
		if len(c.Command) > 1 {
			cfg.ServerArgs = c.Command[1:]
		}
	}
	cfg.ServerEnv = c.Env
	if c.Timeout > 0 {
		cfg.Timeout = c.Timeout
	}
	return cfg
}
