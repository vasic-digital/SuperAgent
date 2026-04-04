// Package bash_providers provides integration of bash scripts as MCP tools.
// This package scans, parses, and registers bash tools from CLI agents.
package bash_providers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"dev.helix.agent/internal/services"
)

// BashTool represents a bash script tool
type BashTool struct {
	Name        string
	Description string
	ScriptPath  string
	Parameters  []Parameter
	EnvVars     []EnvVar
}

// Parameter represents a tool parameter
type Parameter struct {
	Name        string
	Description string
	Required    bool
	Type        string
}

// EnvVar represents an environment variable requirement
type EnvVar struct {
	Name     string
	Required bool
	Default  string
}

// Registry manages bash tools
type Registry struct {
	tools      map[string]*BashTool
	toolsDir   string
	argcBinary string
}

// NewRegistry creates a new bash tool registry
func NewRegistry(toolsDir string) *Registry {
	return &Registry{
		tools:      make(map[string]*BashTool),
		toolsDir:   toolsDir,
		argcBinary: "argc",
	}
}

// Discover scans the tools directory for bash scripts
func (r *Registry) Discover() error {
	return filepath.Walk(r.toolsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.sh files
		if info.IsDir() || !strings.HasSuffix(path, ".sh") {
			return nil
		}

		// Skip guard operation and utility scripts
		if strings.Contains(path, "guard_operation.sh") ||
			strings.Contains(path, "README.md") {
			return nil
		}

		// Parse the tool
		tool, err := r.parseTool(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		r.tools[tool.Name] = tool
		return nil
	})
}

// parseTool extracts metadata from a bash script
func (r *Registry) parseTool(path string) (*BashTool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tool := &BashTool{
		ScriptPath: path,
		Name:       strings.TrimSuffix(filepath.Base(path), ".sh"),
	}

	scanner := bufio.NewScanner(file)
	paramRegex := regexp.MustCompile(`^#\s*@option\s+(--(\w+))(!)?\s*(.*)$`)
	envRegex := regexp.MustCompile(`^#\s*@env\s+(\w+)(!)?\s*(?:\[=([^\]]*)\])?\s*(.*)$`)
	descRegex := regexp.MustCompile(`^#\s*@describe\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse description
		if matches := descRegex.FindStringSubmatch(line); matches != nil {
			tool.Description = strings.TrimSpace(matches[1])
			continue
		}

		// Parse parameters
		if matches := paramRegex.FindStringSubmatch(line); matches != nil {
			param := Parameter{
				Name:        matches[2],
				Required:    matches[3] == "!",
				Description: strings.TrimSpace(matches[4]),
				Type:        "string",
			}
			tool.Parameters = append(tool.Parameters, param)
			continue
		}

		// Parse environment variables
		if matches := envRegex.FindStringSubmatch(line); matches != nil {
			env := EnvVar{
				Name:     matches[1],
				Required: matches[2] == "!",
				Default:  matches[3],
			}
			tool.EnvVars = append(tool.EnvVars, env)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tool, nil
}

// Execute runs a bash tool
func (r *Registry) Execute(ctx context.Context, toolName string, args map[string]interface{}) (*services.ToolCallResult, error) {
	tool, ok := r.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	// Build command arguments
	cmdArgs := []string{}
	for _, param := range tool.Parameters {
		value, ok := args[param.Name]
		if !ok {
			if param.Required {
				return nil, fmt.Errorf("missing required parameter: %s", param.Name)
			}
			continue
		}
		cmdArgs = append(cmdArgs, fmt.Sprintf("--%s", param.Name), fmt.Sprintf("%v", value))
	}

	// Set environment variables
	env := os.Environ()
	for _, envVar := range tool.EnvVars {
		value := os.Getenv(envVar.Name)
		if value == "" && envVar.Default != "" {
			value = envVar.Default
		}
		if value == "" && envVar.Required {
			return nil, fmt.Errorf("missing required environment variable: %s", envVar.Name)
		}
		if value != "" {
			env = append(env, fmt.Sprintf("%s=%s", envVar.Name, value))
		}
	}

	// Execute the script
	cmd := exec.CommandContext(ctx, tool.ScriptPath, cmdArgs...)
	cmd.Env = env
	cmd.Dir = filepath.Dir(tool.ScriptPath)

	output, err := cmd.CombinedOutput()

	result := &services.ToolCallResult{
		Content: []services.Content{
			{
				Type: "text",
				Text: string(output),
			},
		},
		IsError: err != nil,
	}

	if err != nil {
		result.Content = append(result.Content, services.Content{
			Type: "text",
			Text: fmt.Sprintf("Error: %v", err),
		})
	}

	return result, nil
}

// ToMCPTool converts a BashTool to MCP Tool format
func (t *BashTool) ToMCPTool() services.MCPToolDefinition {
	properties := make(map[string]interface{})
	required := []string{}

	for _, param := range t.Parameters {
		properties[param.Name] = map[string]string{
			"type":        param.Type,
			"description": param.Description,
		}
		if param.Required {
			required = append(required, param.Name)
		}
	}

	return services.MCPToolDefinition{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: services.ToolInputSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		},
	}
}

// List returns all registered tools
func (r *Registry) List() []*BashTool {
	result := make([]*BashTool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// Get returns a specific tool
func (r *Registry) Get(name string) (*BashTool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}
