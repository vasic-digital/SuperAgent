// Package main provides the HelixAgent MCP server for OpenCode integration.
//
// This server implements the Model Context Protocol (MCP) for OpenCode,
// providing access to HelixAgent's AI Debate Ensemble, background tasks,
// RAG, and memory systems.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an outgoing MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HelixAgentConfig holds the HelixAgent connection configuration
type HelixAgentConfig struct {
	Endpoint     string
	PreferHTTP3  bool
	EnableTOON   bool
	EnableBrotli bool
}

// MCPServer handles MCP protocol communication
type MCPServer struct {
	config      HelixAgentConfig
	httpClient  *http.Client
	contentType string
	encoding    string
}

// Tool definitions
var tools = []map[string]interface{}{
	{
		"name":        "helixagent_debate",
		"description": "Start an AI debate with 15 LLMs (5 positions x 3 LLMs each) to reach consensus on a topic",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "The topic or question to debate",
				},
				"rounds": map[string]interface{}{
					"type":        "integer",
					"description": "Number of debate rounds (default: 3)",
					"default":     3,
				},
				"enable_multi_pass_validation": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable multi-pass validation for higher quality responses",
					"default":     true,
				},
			},
			"required": []string{"topic"},
		},
	},
	{
		"name":        "helixagent_ensemble",
		"description": "Get a response from the AI Debate Ensemble (single query, confidence-weighted voting)",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The prompt to send to the ensemble",
				},
				"temperature": map[string]interface{}{
					"type":        "number",
					"description": "Sampling temperature (0-1)",
					"default":     0.7,
				},
			},
			"required": []string{"prompt"},
		},
	},
	{
		"name":        "helixagent_task",
		"description": "Create a background task for long-running operations",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The command to execute",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Task description",
				},
				"timeout": map[string]interface{}{
					"type":        "integer",
					"description": "Timeout in seconds (default: 300)",
					"default":     300,
				},
			},
			"required": []string{"command"},
		},
	},
	{
		"name":        "helixagent_rag",
		"description": "Perform a hybrid RAG query (dense + sparse retrieval with reranking)",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The query to search for",
				},
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "The collection to search in",
				},
				"top_k": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to return (default: 5)",
					"default":     5,
				},
			},
			"required": []string{"query"},
		},
	},
	{
		"name":        "helixagent_memory",
		"description": "Access the Mem0-style memory system for storing and retrieving memories",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"add", "search", "get", "delete"},
					"description": "The memory action to perform",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Memory content (for add action)",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (for search action)",
				},
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (for get/delete actions)",
				},
			},
			"required": []string{"action"},
		},
	},
}

func main() {
	// Load configuration from environment
	config := HelixAgentConfig{
		Endpoint:     getEnv("HELIXAGENT_ENDPOINT", "https://localhost:7061"),
		PreferHTTP3:  getEnvBool("HELIXAGENT_PREFER_HTTP3", true),
		EnableTOON:   getEnvBool("HELIXAGENT_ENABLE_TOON", true),
		EnableBrotli: getEnvBool("HELIXAGENT_ENABLE_BROTLI", true),
	}

	server := NewMCPServer(config)

	// Run MCP server on stdio
	server.Run(os.Stdin, os.Stdout)
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(config HelixAgentConfig) *MCPServer {
	contentType := "application/json"
	if config.EnableTOON {
		contentType = "application/toon+json"
	}

	encoding := "gzip"
	if config.EnableBrotli {
		encoding = "br, gzip"
	}

	return &MCPServer{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		contentType: contentType,
		encoding:    encoding,
	}
}

// Run starts the MCP server, reading from stdin and writing to stdout
func (s *MCPServer) Run(reader io.Reader, writer io.Writer) {
	scanner := bufio.NewScanner(reader)
	encoder := json.NewEncoder(writer)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeError(encoder, nil, -32700, "Parse error")
			continue
		}

		resp := s.handleRequest(&req)
		encoder.Encode(resp)
	}
}

// handleRequest processes an MCP request and returns a response
func (s *MCPServer) handleRequest(req *MCPRequest) *MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleListTools(req)
	case "tools/call":
		return s.handleCallTool(req)
	case "notifications/initialized":
		return nil // No response needed for notifications
	default:
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(req *MCPRequest) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "helixagent-mcp",
				"version": "1.0.0",
			},
		},
	}
}

// handleListTools handles the tools/list request
func (s *MCPServer) handleListTools(req *MCPRequest) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleCallTool handles tool invocations
func (s *MCPServer) handleCallTool(req *MCPRequest) *MCPResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return s.errorResponse(req.ID, -32602, "Invalid params")
	}

	toolName, _ := params["name"].(string)
	toolArgs, _ := params["arguments"].(map[string]interface{})

	var result interface{}
	var err error

	switch toolName {
	case "helixagent_debate":
		result, err = s.callDebate(toolArgs)
	case "helixagent_ensemble":
		result, err = s.callEnsemble(toolArgs)
	case "helixagent_task":
		result, err = s.callTask(toolArgs)
	case "helixagent_rag":
		result, err = s.callRAG(toolArgs)
	case "helixagent_memory":
		result, err = s.callMemory(toolArgs)
	default:
		return s.errorResponse(req.ID, -32602, fmt.Sprintf("Unknown tool: %s", toolName))
	}

	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": formatResult(result),
				},
			},
		},
	}
}

// callDebate invokes the AI debate endpoint
func (s *MCPServer) callDebate(args map[string]interface{}) (interface{}, error) {
	topic, _ := args["topic"].(string)
	rounds := 3
	if r, ok := args["rounds"].(float64); ok {
		rounds = int(r)
	}
	enableMultiPass := true
	if mp, ok := args["enable_multi_pass_validation"].(bool); ok {
		enableMultiPass = mp
	}

	body := map[string]interface{}{
		"topic":                        topic,
		"rounds":                       rounds,
		"enable_multi_pass_validation": enableMultiPass,
	}

	return s.doRequest("POST", "/v1/debates", body)
}

// callEnsemble invokes the ensemble endpoint
func (s *MCPServer) callEnsemble(args map[string]interface{}) (interface{}, error) {
	prompt, _ := args["prompt"].(string)
	temperature := 0.7
	if t, ok := args["temperature"].(float64); ok {
		temperature = t
	}

	body := map[string]interface{}{
		"model": "helix-debate-ensemble",
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": temperature,
	}

	return s.doRequest("POST", "/v1/chat/completions", body)
}

// callTask creates a background task
func (s *MCPServer) callTask(args map[string]interface{}) (interface{}, error) {
	command, _ := args["command"].(string)
	description, _ := args["description"].(string)
	timeout := 300
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	body := map[string]interface{}{
		"command":     command,
		"description": description,
		"timeout":     timeout,
	}

	return s.doRequest("POST", "/v1/tasks", body)
}

// callRAG invokes the RAG endpoint
func (s *MCPServer) callRAG(args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	collection, _ := args["collection"].(string)
	topK := 5
	if k, ok := args["top_k"].(float64); ok {
		topK = int(k)
	}

	body := map[string]interface{}{
		"query":      query,
		"collection": collection,
		"top_k":      topK,
	}

	return s.doRequest("POST", "/v1/rag/query", body)
}

// callMemory invokes the memory endpoint
func (s *MCPServer) callMemory(args map[string]interface{}) (interface{}, error) {
	action, _ := args["action"].(string)

	var method, path string
	var body map[string]interface{}

	switch action {
	case "add":
		method = "POST"
		path = "/v1/memory"
		body = map[string]interface{}{
			"content": args["content"],
		}
	case "search":
		method = "POST"
		path = "/v1/memory/search"
		body = map[string]interface{}{
			"query": args["query"],
		}
	case "get":
		method = "GET"
		memoryID, _ := args["memory_id"].(string)
		path = fmt.Sprintf("/v1/memory/%s", memoryID)
	case "delete":
		method = "DELETE"
		memoryID, _ := args["memory_id"].(string)
		path = fmt.Sprintf("/v1/memory/%s", memoryID)
	default:
		return nil, fmt.Errorf("unknown memory action: %s", action)
	}

	return s.doRequest(method, path, body)
}

// doRequest performs an HTTP request to HelixAgent
func (s *MCPServer) doRequest(method, path string, body interface{}) (interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, s.config.Endpoint+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", s.contentType)
	req.Header.Set("Accept", s.contentType)
	req.Header.Set("Accept-Encoding", s.encoding)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}

// errorResponse creates an error response
func (s *MCPServer) errorResponse(id interface{}, code int, message string) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}

// writeError writes an error response
func (s *MCPServer) writeError(encoder *json.Encoder, id interface{}, code int, message string) {
	encoder.Encode(s.errorResponse(id, code, message))
}

// formatResult formats a result for display
func formatResult(result interface{}) string {
	switch v := result.(type) {
	case string:
		return v
	case map[string]interface{}:
		formatted, _ := json.MarshalIndent(v, "", "  ")
		return string(formatted)
	default:
		formatted, _ := json.Marshal(v)
		return string(formatted)
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}
