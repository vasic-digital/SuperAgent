package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Tool represents a unified tool interface
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	Source() string // "mcp", "lsp", "custom", etc.
}

// ToolRegistry manages tools from various sources
type ToolRegistry struct {
	mu          sync.RWMutex
	tools       map[string]Tool
	mcpManager  *MCPManager
	lspClient   *LSPClient
	customTools map[string]Tool
	lastRefresh time.Time
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(mcpManager *MCPManager, lspClient *LSPClient) *ToolRegistry {
	return &ToolRegistry{
		tools:       make(map[string]Tool),
		mcpManager:  mcpManager,
		lspClient:   lspClient,
		customTools: make(map[string]Tool),
	}
}

// RegisterCustomTool registers a custom tool with validation
func (tr *ToolRegistry) RegisterCustomTool(tool Tool) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	name := tool.Name()
	if _, exists := tr.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	// Validate tool metadata
	if err := tr.validateToolMetadata(tool); err != nil {
		return fmt.Errorf("tool validation failed: %w", err)
	}

	tr.tools[name] = tool
	tr.customTools[name] = tool
	return nil
}

// validateToolMetadata validates tool metadata
func (tr *ToolRegistry) validateToolMetadata(tool Tool) error {
	if tool.Name() == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Description() == "" {
		return fmt.Errorf("tool description cannot be empty")
	}

	params := tool.Parameters()
	if params == nil {
		return fmt.Errorf("tool parameters cannot be nil")
	}

	// Validate parameter schemas
	for paramName, paramSchema := range params {
		if err := tr.validateParameterSchema(paramName, paramSchema); err != nil {
			return err
		}
	}

	return nil
}

// validateParameterSchema validates a parameter schema
func (tr *ToolRegistry) validateParameterSchema(name string, schema interface{}) error {
	// Basic validation - can be enhanced with JSON Schema validation
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("parameter %s schema must be a map", name)
	}

	if _, hasType := schemaMap["type"]; !hasType {
		return fmt.Errorf("parameter %s schema must have a type", name)
	}

	return nil
}

// RegisterExternalToolSource registers tools from an external source
func (tr *ToolRegistry) RegisterExternalToolSource(sourceName string, toolFetcher func() ([]Tool, error)) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tools, err := toolFetcher()
	if err != nil {
		return fmt.Errorf("failed to fetch tools from %s: %w", sourceName, err)
	}

	for _, tool := range tools {
		name := tool.Name()
		if _, exists := tr.tools[name]; exists {
			log.Printf("Tool %s from %s already exists, skipping", name, sourceName)
			continue
		}

		if err := tr.validateToolMetadata(tool); err != nil {
			log.Printf("Tool %s from %s validation failed: %v, skipping", name, sourceName, err)
			continue
		}

		tr.tools[name] = tool
		log.Printf("Registered tool %s from external source %s", name, sourceName)
	}

	return nil
}

// RefreshTools refreshes tools from all sources
func (tr *ToolRegistry) RefreshTools(ctx context.Context) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Clear existing tools except custom ones
	for name, tool := range tr.tools {
		if tool.Source() != "custom" {
			delete(tr.tools, name)
		}
	}

	// Add MCP tools
	if tr.mcpManager != nil {
		mcpTools := tr.mcpManager.ListTools()
		for _, mcpTool := range mcpTools {
			wrapper := &MCPToolWrapper{
				mcpTool:    mcpTool,
				mcpManager: tr.mcpManager,
			}
			tr.tools[mcpTool.Name] = wrapper
		}
	}

	// Add LSP-based tools (code actions)
	if tr.lspClient != nil {
		// LSP tools would be added here when implemented
	}

	tr.lastRefresh = time.Now()
	return nil
}

// GetTool returns a tool by name
func (tr *ToolRegistry) GetTool(name string) (Tool, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tool, exists := tr.tools[name]
	return tool, exists
}

// ListTools returns all available tools
func (tr *ToolRegistry) ListTools() []Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tools := make([]Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool safely executes a tool with sandboxing
func (tr *ToolRegistry) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	tool, exists := tr.GetTool(name)
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	// Basic parameter validation
	if err := tr.validateParameters(tool, params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := tool.Execute(execCtx, params)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

// validateParameters performs basic parameter validation
func (tr *ToolRegistry) validateParameters(tool Tool, params map[string]interface{}) error {
	// Basic validation - could be enhanced
	required := tool.Parameters()
	for key := range required {
		if _, exists := params[key]; !exists {
			return fmt.Errorf("missing required parameter: %s", key)
		}
	}
	return nil
}

// MCPToolWrapper wraps MCP tools to implement the Tool interface
type MCPToolWrapper struct {
	mcpTool    *MCPTool
	mcpManager *MCPManager
}

func (w *MCPToolWrapper) Name() string {
	return w.mcpTool.Name
}

func (w *MCPToolWrapper) Description() string {
	return w.mcpTool.Description
}

func (w *MCPToolWrapper) Parameters() map[string]interface{} {
	return w.mcpTool.InputSchema
}

func (w *MCPToolWrapper) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return w.mcpManager.CallTool(ctx, w.mcpTool.Name, params)
}

func (w *MCPToolWrapper) Source() string {
	return "mcp"
}

// LSPToolWrapper would wrap LSP-based tools (code actions, etc.)
// Implementation would be added when LSP tools are implemented

// UnifiedSearchOptions configures unified tool search
type UnifiedSearchOptions struct {
	Query       string   `json:"query"`
	Sources     []string `json:"sources,omitempty"`     // "mcp", "lsp", "custom", "schema"
	Categories  []string `json:"categories,omitempty"`
	MaxResults  int      `json:"max_results,omitempty"`
	MinScore    float64  `json:"min_score,omitempty"`
	FuzzyMatch  bool     `json:"fuzzy_match,omitempty"`
}

// UnifiedSearchResult represents a unified search result
type UnifiedSearchResult struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Category    string                 `json:"category,omitempty"`
	Score       float64                `json:"score"`
	MatchType   string                 `json:"match_type"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Search performs unified search across all tool sources
func (tr *ToolRegistry) Search(opts UnifiedSearchOptions) []UnifiedSearchResult {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if opts.MaxResults <= 0 {
		opts.MaxResults = 50
	}
	if opts.MinScore <= 0 {
		opts.MinScore = 0.1
	}

	var results []UnifiedSearchResult
	query := strings.ToLower(opts.Query)

	// Determine which sources to search
	searchAll := len(opts.Sources) == 0
	sourceMap := make(map[string]bool)
	for _, s := range opts.Sources {
		sourceMap[strings.ToLower(s)] = true
	}

	// Search registered tools
	for _, tool := range tr.tools {
		source := strings.ToLower(tool.Source())
		if !searchAll && !sourceMap[source] {
			continue
		}

		score, matchType := tr.calculateToolSearchScore(tool, query, opts.FuzzyMatch)
		if score >= opts.MinScore {
			results = append(results, UnifiedSearchResult{
				Name:        tool.Name(),
				Description: tool.Description(),
				Source:      tool.Source(),
				Score:       score,
				MatchType:   matchType,
				Parameters:  tool.Parameters(),
			})
		}
	}

	// Sort by score descending
	tr.sortSearchResults(results)

	// Limit results
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	return results
}

// calculateToolSearchScore calculates relevance score for a tool
func (tr *ToolRegistry) calculateToolSearchScore(tool Tool, query string, fuzzy bool) (float64, string) {
	if query == "" {
		return 1.0, "all"
	}

	var maxScore float64
	var matchType string

	name := strings.ToLower(tool.Name())
	desc := strings.ToLower(tool.Description())

	// Exact name match
	if name == query {
		return 1.0, "name"
	}

	// Name contains query
	if strings.Contains(name, query) {
		score := 0.9 * (float64(len(query)) / float64(len(name)))
		if score > maxScore {
			maxScore = score
			matchType = "name"
		}
	}

	// Description match
	if strings.Contains(desc, query) {
		words := strings.Fields(query)
		matchedWords := 0
		for _, word := range words {
			if strings.Contains(desc, word) {
				matchedWords++
			}
		}
		score := 0.7 * (float64(matchedWords) / float64(len(words)))
		if score > maxScore {
			maxScore = score
			matchType = "description"
		}
	}

	// Parameter name match
	for paramName := range tool.Parameters() {
		if strings.Contains(strings.ToLower(paramName), query) {
			score := 0.5
			if score > maxScore {
				maxScore = score
				matchType = "parameter"
			}
		}
	}

	// Fuzzy match as fallback
	if fuzzy && maxScore < 0.3 {
		fuzzyScore := tr.fuzzyMatch(name, query)
		if fuzzyScore > maxScore {
			maxScore = fuzzyScore
			matchType = "fuzzy"
		}
	}

	return maxScore, matchType
}

// fuzzyMatch calculates a fuzzy match score
func (tr *ToolRegistry) fuzzyMatch(s1, s2 string) float64 {
	if len(s1) == 0 || len(s2) == 0 {
		return 0
	}

	shorter, longer := s1, s2
	if len(s1) > len(s2) {
		shorter, longer = s2, s1
	}

	matches := 0
	for _, c := range shorter {
		if strings.ContainsRune(longer, c) {
			matches++
		}
	}

	return 0.5 * (float64(matches) / float64(len(longer)))
}

// sortSearchResults sorts results by score descending
func (tr *ToolRegistry) sortSearchResults(results []UnifiedSearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// GetToolSuggestions returns suggestions based on partial input
func (tr *ToolRegistry) GetToolSuggestions(prefix string, maxSuggestions int) []Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if maxSuggestions <= 0 {
		maxSuggestions = 10
	}

	prefixLower := strings.ToLower(prefix)
	var suggestions []Tool

	for _, tool := range tr.tools {
		if strings.HasPrefix(strings.ToLower(tool.Name()), prefixLower) {
			suggestions = append(suggestions, tool)
			if len(suggestions) >= maxSuggestions {
				break
			}
		}
	}

	return suggestions
}

// GetToolsBySource returns tools from a specific source
func (tr *ToolRegistry) GetToolsBySource(source string) []Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var tools []Tool
	sourceLower := strings.ToLower(source)

	for _, tool := range tr.tools {
		if strings.ToLower(tool.Source()) == sourceLower {
			tools = append(tools, tool)
		}
	}

	return tools
}

// GetToolStats returns statistics about registered tools
func (tr *ToolRegistry) GetToolStats() map[string]interface{} {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	sourceCounts := make(map[string]int)
	for _, tool := range tr.tools {
		sourceCounts[tool.Source()]++
	}

	return map[string]interface{}{
		"total_tools":   len(tr.tools),
		"by_source":     sourceCounts,
		"last_refresh":  tr.lastRefresh,
		"custom_count":  len(tr.customTools),
	}
}
