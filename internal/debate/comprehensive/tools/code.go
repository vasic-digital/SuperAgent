package comprehensive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ToolType represents the type of tool
type ToolType string

const (
	ToolTypeCode        ToolType = "code"
	ToolTypeCommand     ToolType = "command"
	ToolTypeDatabase    ToolType = "database"
	ToolTypeAnalysis    ToolType = "analysis"
	ToolTypeSecurity    ToolType = "security"
	ToolTypePerformance ToolType = "performance"
)

// Tool is the interface all tools must implement
type Tool interface {
	// GetName returns the tool's unique name
	GetName() string

	// GetType returns the tool type
	GetType() ToolType

	// GetDescription returns a description for LLMs
	GetDescription() string

	// GetInputSchema returns the JSON schema for inputs
	GetInputSchema() map[string]interface{}

	// Execute executes the tool with given inputs
	Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error)

	// Validate validates the inputs
	Validate(inputs map[string]interface{}) error
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success   bool                   `json:"success"`
	Output    string                 `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewToolResult creates a new successful tool result
func NewToolResult(output string) *ToolResult {
	return &ToolResult{
		Success:   true,
		Output:    output,
		Data:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// NewToolError creates a new error tool result
func NewToolError(err string) *ToolResult {
	return &ToolResult{
		Success:   false,
		Error:     err,
		Data:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// ToolRegistry manages all available tools
type ToolRegistry struct {
	tools  map[string]Tool
	logger *logrus.Logger
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(logger *logrus.Logger) *ToolRegistry {
	if logger == nil {
		logger = logrus.New()
	}

	return &ToolRegistry{
		tools:  make(map[string]Tool),
		logger: logger,
	}
}

// Register registers a tool
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.GetName()] = tool
	r.logger.WithField("tool", tool.GetName()).Debug("Tool registered")
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetAll returns all registered tools
func (r *ToolRegistry) GetAll() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

// GetByType returns tools of a specific type
func (r *ToolRegistry) GetByType(toolType ToolType) []Tool {
	var result []Tool
	for _, tool := range r.tools {
		if tool.GetType() == toolType {
			result = append(result, tool)
		}
	}
	return result
}

// Execute executes a tool by name
func (r *ToolRegistry) Execute(ctx context.Context, toolName string, inputs map[string]interface{}) (*ToolResult, error) {
	tool, ok := r.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	// Validate inputs
	if err := tool.Validate(inputs); err != nil {
		return NewToolError(fmt.Sprintf("validation failed: %v", err)), nil
	}

	// Execute tool
	start := time.Now()
	result, err := tool.Execute(ctx, inputs)
	result.Duration = time.Since(start)

	if err != nil {
		return NewToolError(err.Error()), nil
	}

	return result, nil
}

// CodeTool provides file operations
type CodeTool struct {
	basePath string
	logger   *logrus.Logger
}

// NewCodeTool creates a new code tool
func NewCodeTool(basePath string, logger *logrus.Logger) *CodeTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &CodeTool{
		basePath: basePath,
		logger:   logger,
	}
}

// GetName returns the tool name
func (t *CodeTool) GetName() string {
	return "code"
}

// GetType returns the tool type
func (t *CodeTool) GetType() ToolType {
	return ToolTypeCode
}

// GetDescription returns the description
func (t *CodeTool) GetDescription() string {
	return "Read, write, update, and delete files in the codebase"
}

// GetInputSchema returns the input schema
func (t *CodeTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: read, write, update, delete",
				"enum":        []string{"read", "write", "update", "delete"},
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path relative to project root",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content for write/update actions",
			},
		},
		"required": []string{"action", "path"},
	}
}

// Validate validates inputs
func (t *CodeTool) Validate(inputs map[string]interface{}) error {
	action, ok := inputs["action"].(string)
	if !ok {
		return fmt.Errorf("action is required")
	}

	validActions := []string{"read", "write", "update", "delete"}
	found := false
	for _, a := range validActions {
		if a == action {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid action: %s", action)
	}

	path, ok := inputs["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path is required")
	}

	// Check for path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check content for write/update
	if (action == "write" || action == "update") && inputs["content"] == nil {
		return fmt.Errorf("content is required for write/update actions")
	}

	return nil
}

// Execute executes the tool
func (t *CodeTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	action := inputs["action"].(string)
	path := inputs["path"].(string)
	fullPath := filepath.Join(t.basePath, path)

	t.logger.WithFields(logrus.Fields{
		"action": action,
		"path":   path,
	}).Debug("Executing code tool")

	switch action {
	case "read":
		return t.readFile(fullPath)
	case "write":
		content := inputs["content"].(string)
		return t.writeFile(fullPath, content)
	case "update":
		content := inputs["content"].(string)
		return t.updateFile(fullPath, content)
	case "delete":
		return t.deleteFile(fullPath)
	default:
		return NewToolError(fmt.Sprintf("unknown action: %s", action)), nil
	}
}

// readFile reads a file
func (t *CodeTool) readFile(path string) (*ToolResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewToolError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	result := NewToolResult(string(data))
	result.Data["size"] = len(data)
	result.Data["path"] = path
	return result, nil
}

// writeFile writes a file
func (t *CodeTool) writeFile(path string, content string) (*ToolResult, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewToolError(fmt.Sprintf("failed to create directory: %v", err)), nil
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return NewToolError(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	result := NewToolResult("File written successfully")
	result.Data["path"] = path
	result.Data["size"] = len(content)
	return result, nil
}

// updateFile updates a file
func (t *CodeTool) updateFile(path string, content string) (*ToolResult, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewToolError("file does not exist, use write instead"), nil
	}

	// Create backup
	backupPath := path + ".backup"
	data, err := os.ReadFile(path)
	if err != nil {
		return NewToolError(fmt.Sprintf("failed to read existing file: %v", err)), nil
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return NewToolError(fmt.Sprintf("failed to create backup: %v", err)), nil
	}

	// Write new content
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		// Restore backup on failure
		os.WriteFile(path, data, 0644)
		return NewToolError(fmt.Sprintf("failed to update file: %v", err)), nil
	}

	result := NewToolResult("File updated successfully")
	result.Data["path"] = path
	result.Data["size"] = len(content)
	result.Data["backup"] = backupPath
	return result, nil
}

// deleteFile deletes a file
func (t *CodeTool) deleteFile(path string) (*ToolResult, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewToolError("file does not exist"), nil
	}

	// Move to trash instead of permanent deletion
	trashDir := filepath.Join(t.basePath, ".trash")
	os.MkdirAll(trashDir, 0755)

	trashPath := filepath.Join(trashDir, filepath.Base(path)+"."+time.Now().Format("20060102150405"))

	if err := os.Rename(path, trashPath); err != nil {
		return NewToolError(fmt.Sprintf("failed to delete file: %v", err)), nil
	}

	result := NewToolResult("File moved to trash")
	result.Data["original_path"] = path
	result.Data["trash_path"] = trashPath
	return result, nil
}

// SearchTool searches files in the codebase
type SearchTool struct {
	basePath string
	logger   *logrus.Logger
}

// NewSearchTool creates a new search tool
func NewSearchTool(basePath string, logger *logrus.Logger) *SearchTool {
	if logger == nil {
		logger = logrus.New()
	}

	return &SearchTool{
		basePath: basePath,
		logger:   logger,
	}
}

// GetName returns the tool name
func (t *SearchTool) GetName() string {
	return "search"
}

// GetType returns the tool type
func (t *SearchTool) GetType() ToolType {
	return ToolTypeCode
}

// GetDescription returns the description
func (t *SearchTool) GetDescription() string {
	return "Search for files and content in the codebase"
}

// GetInputSchema returns the input schema
func (t *SearchTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Search pattern or glob",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Search type: file, content, or grep",
				"enum":        []string{"file", "content", "grep"},
			},
		},
		"required": []string{"pattern", "type"},
	}
}

// Validate validates inputs
func (t *SearchTool) Validate(inputs map[string]interface{}) error {
	if _, ok := inputs["pattern"].(string); !ok {
		return fmt.Errorf("pattern is required")
	}
	if _, ok := inputs["type"].(string); !ok {
		return fmt.Errorf("type is required")
	}
	return nil
}

// Execute executes the tool
func (t *SearchTool) Execute(ctx context.Context, inputs map[string]interface{}) (*ToolResult, error) {
	pattern := inputs["pattern"].(string)
	searchType := inputs["type"].(string)

	var matches []string
	var err error

	switch searchType {
	case "file":
		matches, err = t.searchFiles(pattern)
	case "content":
		matches, err = t.searchContent(pattern)
	case "grep":
		matches, err = t.grep(pattern)
	default:
		return NewToolError(fmt.Sprintf("unknown search type: %s", searchType)), nil
	}

	if err != nil {
		return NewToolError(fmt.Sprintf("search failed: %v", err)), nil
	}

	result := NewToolResult(fmt.Sprintf("Found %d matches", len(matches)))
	result.Data["matches"] = matches
	result.Data["count"] = len(matches)
	return result, nil
}

// searchFiles searches for files matching pattern
func (t *SearchTool) searchFiles(pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(t.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			relPath, _ := filepath.Rel(t.basePath, path)
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// searchContent searches file content
func (t *SearchTool) searchContent(pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(t.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(data), pattern) {
			relPath, _ := filepath.Rel(t.basePath, path)
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// grep performs regex search
func (t *SearchTool) grep(pattern string) ([]string, error) {
	// Simple string search for now
	return t.searchContent(pattern)
}
