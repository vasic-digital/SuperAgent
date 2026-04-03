// Package mcp implements Model Context Protocol support
// Adapted from Snow CLI's MCP integration patterns
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransportType defines how MCP communicates
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportLocal TransportType = "local" // Alias for stdio
	TransportHTTP  TransportType = "http"
)

// ServerConfig represents an MCP server configuration
// Based on Snow CLI's mcp-config.json format
type ServerConfig struct {
	Name        string            `json:"name"`
	Type        TransportType     `json:"type,omitempty"` // stdio, local, http
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Enabled     bool              `json:"enabled"`
	Timeout     int               `json:"timeout,omitempty"` // milliseconds
	Environment map[string]string `json:"environment,omitempty"` // Alias for env
}

// Client represents an MCP client connection
type Client struct {
	logger    *zap.Logger
	config    ServerConfig
	transport Transport
	mu        sync.RWMutex
	connected bool
}

// Transport defines the interface for MCP communication
type Transport interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error)
	IsConnected() bool
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewClient creates a new MCP client
func NewClient(logger *zap.Logger, config ServerConfig) *Client {
	// Default timeout: 5 minutes
	if config.Timeout == 0 {
		config.Timeout = 300000
	}
	
	// Environment is alias for env
	if len(config.Environment) > 0 && len(config.Env) == 0 {
		config.Env = config.Environment
	}
	
	return &Client{
		logger: logger,
		config: config,
	}
}

// Connect establishes the MCP connection
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.connected {
		return nil
	}
	
	var transport Transport
	
	switch c.config.Type {
	case TransportStdio, TransportLocal:
		transport = NewStdioTransport(c.logger, c.config)
	case TransportHTTP:
		transport = NewHTTPTransport(c.logger, c.config)
	default:
		// Auto-detect based on config
		if c.config.URL != "" {
			transport = NewHTTPTransport(c.logger, c.config)
		} else if c.config.Command != "" {
			transport = NewStdioTransport(c.logger, c.config)
		} else {
			return fmt.Errorf("cannot determine transport type for MCP server: %s", c.config.Name)
		}
	}
	
	if err := transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to MCP server %s: %w", c.config.Name, err)
	}
	
	c.transport = transport
	c.connected = true
	
	c.logger.Info("Connected to MCP server", zap.String("name", c.config.Name))
	return nil
}

// Disconnect closes the MCP connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.connected || c.transport == nil {
		return nil
	}
	
	if err := c.transport.Disconnect(); err != nil {
		c.logger.Warn("Error disconnecting from MCP server", 
			zap.String("name", c.config.Name),
			zap.Error(err))
	}
	
	c.connected = false
	c.transport = nil
	
	c.logger.Info("Disconnected from MCP server", zap.String("name", c.config.Name))
	return nil
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if !c.connected || c.transport == nil {
		return false
	}
	
	return c.transport.IsConnected()
}

// Call invokes an MCP method
func (c *Client) Call(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	if !c.IsConnected() {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	
	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      uuid.New().String(),
		Method:  method,
		Params:  paramsBytes,
	}
	
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()
	
	if transport == nil {
		return nil, fmt.Errorf("MCP client not connected: %s", c.config.Name)
	}
	
	return transport.Send(ctx, request)
}

// Initialize initializes the MCP session
func (c *Client) Initialize(ctx context.Context) (*InitializeResult, error) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCapabilities{},
		ClientInfo: Implementation{
			Name:    "HelixAgent",
			Version: "1.0.0",
		},
	}
	
	resp, err := c.Call(ctx, "initialize", params)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}
	
	// Send initialized notification
	_, _ = c.Call(ctx, "notifications/initialized", nil)
	
	return &result, nil
}

// ListTools retrieves available tools from the server
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	resp, err := c.Call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	var result ListToolsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools list: %w", err)
	}
	
	return result.Tools, nil
}

// CallTool invokes a tool on the server
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*CallToolResult, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: arguments,
	}
	
	resp, err := c.Call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, resp.Error
	}
	
	var result CallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool result: %w", err)
	}
	
	return &result, nil
}

// StdioTransport implements stdio-based MCP transport
type StdioTransport struct {
	logger *zap.Logger
	config ServerConfig
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mu     sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(logger *zap.Logger, config ServerConfig) *StdioTransport {
	return &StdioTransport{
		logger: logger,
		config: config,
	}
}

// Connect starts the subprocess
func (t *StdioTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	cmd := exec.CommandContext(ctx, t.config.Command, t.config.Args...)
	
	// Set environment
	if len(t.config.Env) > 0 {
		env := os.Environ()
		for k, v := range t.config.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}
	
	t.cmd = cmd
	t.stdin = stdin
	t.stdout = stdout
	
	return nil
}

// Disconnect stops the subprocess
func (t *StdioTransport) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.stdin != nil {
		t.stdin.Close()
	}
	
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}
	
	return nil
}

// IsConnected checks if the subprocess is running
func (t *StdioTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	return t.cmd != nil && t.cmd.Process != nil && t.cmd.ProcessState == nil
}

// Send sends a request and reads the response
func (t *StdioTransport) Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.stdin == nil || t.stdout == nil {
		return nil, fmt.Errorf("stdio transport not connected")
	}
	
	// Marshal request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	// Write request with newline
	data = append(data, '\n')
	if _, err := t.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	
	// Read response (line by line)
	decoder := json.NewDecoder(t.stdout)
	var response JSONRPCResponse
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return &response, nil
}

// HTTPTransport implements HTTP-based MCP transport
type HTTPTransport struct {
	logger *zap.Logger
	config ServerConfig
	client *http.Client
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(logger *zap.Logger, config ServerConfig) *HTTPTransport {
	return &HTTPTransport{
		logger: logger,
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Millisecond,
		},
	}
}

// Connect verifies the HTTP endpoint
func (t *HTTPTransport) Connect(ctx context.Context) error {
	// HTTP is stateless, just verify we can reach the endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", t.config.URL, nil)
	if err != nil {
		return err
	}
	
	for k, v := range t.config.Headers {
		req.Header.Set(k, v)
	}
	
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to HTTP MCP server: %w", err)
	}
	defer resp.Body.Close()
	
	return nil
}

// Disconnect is a no-op for HTTP
func (t *HTTPTransport) Disconnect() error {
	return nil
}

// IsConnected always returns true for HTTP
func (t *HTTPTransport) IsConnected() bool {
	return true
}

// Send sends a request via HTTP POST
func (t *HTTPTransport) Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", t.config.URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.config.Headers {
		req.Header.Set(k, v)
	}
	
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return &response, nil
}

// Additional types for MCP

type InitializeParams struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ClientCapabilities  `json:"capabilities"`
	ClientInfo      Implementation      `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ServerCapabilities  `json:"capabilities"`
	ServerInfo      Implementation      `json:"serverInfo"`
}

type ClientCapabilities struct {
	// Define client capabilities
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Need to import bytes
