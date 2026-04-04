// Package sandbox provides a tool for sandboxed execution
package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Tool provides sandboxed command execution as a tool
type Tool struct {
	logger *logrus.Logger
}

// NewTool creates a new sandbox tool
func NewTool(logger *logrus.Logger) *Tool {
	return &Tool{
		logger: logger,
	}
}

// Name returns the tool name
func (t *Tool) Name() string {
	return "Sandbox"
}

// Description returns the tool description
func (t *Tool) Description() string {
	return "Execute commands in a secure sandboxed environment"
}

// Schema returns the tool schema
func (t *Tool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Command arguments",
			},
			"image": map[string]interface{}{
				"type":        "string",
				"description": "Container image to use (default: alpine:latest)",
				"default":     "alpine:latest",
			},
			"enable_network": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable network access",
				"default":     false,
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds",
				"default":     60,
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory inside container",
				"default":     "/workspace",
			},
			"memory_limit": map[string]interface{}{
				"type":        "string",
				"description": "Memory limit (e.g., 512m, 1g)",
				"default":     "512m",
			},
		},
		"required": []string{"command"},
	}
}

// Execute runs the sandbox tool
func (t *Tool) Execute(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	// Extract parameters
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command is required")
	}

	var args []string
	if rawArgs, ok := input["args"].([]interface{}); ok {
		for _, arg := range rawArgs {
			if str, ok := arg.(string); ok {
				args = append(args, str)
			}
		}
	}

	// Build full command
	fullCommand := append([]string{command}, args...)

	// Get configuration
	image := "alpine:latest"
	if img, ok := input["image"].(string); ok && img != "" {
		image = img
	}

	enableNetwork := false
	if en, ok := input["enable_network"].(bool); ok {
		enableNetwork = en
	}

	timeout := 60
	if to, ok := input["timeout"].(float64); ok {
		timeout = int(to)
	}

	workingDir := "/workspace"
	if wd, ok := input["working_dir"].(string); ok && wd != "" {
		workingDir = wd
	}

	memoryLimit := "512m"
	if ml, ok := input["memory_limit"].(string); ok && ml != "" {
		memoryLimit = ml
	}

	// Create sandbox configuration
	config := Config{
		Runtime:       RuntimeDocker,
		EnableNetwork: enableNetwork,
		WorkingDir:    workingDir,
		MemoryLimit:   memoryLimit,
		Timeout:       time.Duration(timeout) * time.Second,
		EnvVars:       make(map[string]string),
	}

	// Create sandbox
	sandbox, err := NewSandbox(config, image)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create sandbox: %v", err),
		}, nil
	}

	// Execute command
	result, err := sandbox.Execute(ctx, fullCommand)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Execution failed: %v", err),
		}, nil
	}

	return &ToolResult{
		Success:  result.ExitCode == 0,
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Metadata: map[string]interface{}{
			"duration_ms": result.Duration.Milliseconds(),
			"image":       image,
			"sandboxed":   true,
		},
	}, nil
}

// ToolResult represents the result of tool execution
type ToolResult struct {
	Success  bool                   `json:"success"`
	ExitCode int                    `json:"exit_code"`
	Stdout   string                 `json:"stdout"`
	Stderr   string                 `json:"stderr,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
