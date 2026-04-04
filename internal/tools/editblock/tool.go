// Package editblock provides edit block functionality as a tool
package editblock

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Tool provides edit block operations as a tool
type Tool struct {
	basePath string
	logger   *logrus.Logger
}

// ToolResult represents the result of tool execution
type ToolResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"file_path,omitempty"`
	Diff     string `json:"diff,omitempty"`
	Error    string `json:"error,omitempty"`
}

// NewTool creates a new edit block tool
func NewTool(basePath string, logger *logrus.Logger) *Tool {
	if logger == nil {
		logger = logrus.New()
	}
	return &Tool{
		basePath: basePath,
		logger:   logger,
	}
}

// Name returns the tool name
func (t *Tool) Name() string {
	return "EditBlock"
}

// Description returns the tool description
func (t *Tool) Description() string {
	return "Apply search/replace edit blocks to files using Aider-style format"
}

// Schema returns the tool schema
func (t *Tool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to edit",
			},
			"search": map[string]interface{}{
				"type":        "string",
				"description": "Text to search for (exact match required)",
			},
			"replace": map[string]interface{}{
				"type":        "string",
				"description": "Text to replace with",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "Preview changes without applying",
				"default":     false,
			},
		},
		"required": []string{"file_path", "search", "replace"},
	}
}

// Execute runs the edit block tool
func (t *Tool) Execute(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	// Extract parameters
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return &ToolResult{
			Success: false,
			Error:   "file_path is required",
		}, nil
	}

	search, ok := input["search"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "search is required",
		}, nil
	}

	replace, ok := input["replace"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "replace is required",
		}, nil
	}

	dryRun := false
	if dr, ok := input["dry_run"].(bool); ok {
		dryRun = dr
	}

	block := EditBlock{
		FilePath: filePath,
		Search:   search,
		Replace:  replace,
	}

	if err := block.Validate(); err != nil {
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Generate diff preview
	diff, err := block.Diff(t.basePath)
	if err != nil {
		return &ToolResult{
			Success:  false,
			FilePath: filePath,
			Error:    fmt.Sprintf("failed to generate diff: %v", err),
		}, nil
	}

	if dryRun {
		return &ToolResult{
			Success:  true,
			FilePath: filePath,
			Diff:     diff,
		}, nil
	}

	// Apply the edit
	result, err := block.Apply(t.basePath)
	if err != nil {
		return &ToolResult{
			Success:  false,
			FilePath: filePath,
			Error:    err.Error(),
		}, nil
	}

	if !result.Success {
		return &ToolResult{
			Success:  false,
			FilePath: filePath,
			Error:    result.Error,
		}, nil
	}

	t.logger.WithFields(logrus.Fields{
		"file":   filePath,
		"lines":  fmt.Sprintf("%d-%d", result.LineStart, result.LineEnd),
		"action": "edit_block_applied",
	}).Info("Edit block applied successfully")

	return &ToolResult{
		Success:  true,
		FilePath: filePath,
		Diff:     diff,
	}, nil
}

// Apply applies a search/replace edit directly
func (t *Tool) Apply(ctx context.Context, filePath, search, replace string) (*ToolResult, error) {
	return t.Execute(ctx, map[string]interface{}{
		"file_path": filePath,
		"search":    search,
		"replace":   replace,
		"dry_run":   false,
	})
}

// Preview generates a diff preview without applying
func (t *Tool) Preview(ctx context.Context, filePath, search, replace string) (*ToolResult, error) {
	return t.Execute(ctx, map[string]interface{}{
		"file_path": filePath,
		"search":    search,
		"replace":   replace,
		"dry_run":   true,
	})
}

// ApplyFromText parses and applies edit blocks from Aider-style text
func (t *Tool) ApplyFromText(ctx context.Context, filePath, text string) ([]ToolResult, error) {
	blocks := ParseEditBlocks(text)
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no edit blocks found in text")
	}

	results := make([]ToolResult, 0, len(blocks))
	for _, block := range blocks {
		block.FilePath = filePath
		result, err := t.Execute(ctx, map[string]interface{}{
			"file_path": block.FilePath,
			"search":    block.Search,
			"replace":   block.Replace,
		})
		if err != nil {
			return results, err
		}
		results = append(results, *result)
	}

	return results, nil
}

// BatchApply applies multiple edit blocks
func (t *Tool) BatchApply(ctx context.Context, operations []map[string]string) ([]ToolResult, error) {
	results := make([]ToolResult, 0, len(operations))

	for _, op := range operations {
		result, err := t.Execute(ctx, map[string]interface{}{
			"file_path": op["file_path"],
			"search":    op["search"],
			"replace":   op["replace"],
		})
		if err != nil {
			return results, err
		}
		results = append(results, *result)
	}

	return results, nil
}

// FindSimilar finds similar blocks of code using fuzzy matching
func (t *Tool) FindSimilar(filePath, search string, threshold float64) ([]string, error) {
	fullPath := t.basePath + "/" + filePath
	lines, err := ReadFileLines(fullPath)
	if err != nil {
		return nil, err
	}

	searchLines := strings.Split(search, "\n")
	if len(searchLines) == 0 {
		return nil, fmt.Errorf("empty search pattern")
	}

	var matches []string
	for i := 0; i <= len(lines)-len(searchLines); i++ {
		similarity := calculateSimilarity(lines[i:i+len(searchLines)], searchLines)
		if similarity >= threshold {
			matches = append(matches, strings.Join(lines[i:i+len(searchLines)], "\n"))
		}
	}

	return matches, nil
}

// calculateSimilarity calculates Jaccard similarity between two line blocks
func calculateSimilarity(a, b []string) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	matches := 0
	for i := range a {
		if normalizeWhitespace(a[i]) == normalizeWhitespace(b[i]) {
			matches++
		}
	}

	return float64(matches) / float64(len(a))
}
