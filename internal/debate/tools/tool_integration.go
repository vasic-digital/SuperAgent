// Package tools provides tool integration for debate agents.
// Enables agents to invoke MCP, ACP, LSP, RAG, formatters, and other services.
package tools

import (
	"context"
	"fmt"
)

// ToolIntegration provides access to all HelixAgent services for debate agents.
type ToolIntegration struct {
	mcpClient         MCPClient
	acpClient         ACPClient
	lspClient         LSPClient
	embeddingClient   EmbeddingClient
	ragClient         RAGClient
	formatterRegistry FormatterRegistry
	enabled           bool
}

// NewToolIntegration creates a tool integration instance.
func NewToolIntegration(
	mcpClient MCPClient,
	acpClient ACPClient,
	lspClient LSPClient,
	embeddingClient EmbeddingClient,
	ragClient RAGClient,
	formatterRegistry FormatterRegistry,
) *ToolIntegration {
	return &ToolIntegration{
		mcpClient:         mcpClient,
		acpClient:         acpClient,
		lspClient:         lspClient,
		embeddingClient:   embeddingClient,
		ragClient:         ragClient,
		formatterRegistry: formatterRegistry,
		enabled:           true,
	}
}

// MCPClient provides access to Model Context Protocol servers.
type MCPClient interface {
	// CallTool invokes an MCP tool by name
	CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)

	// ListTools returns available MCP tools
	ListTools(ctx context.Context) ([]ToolInfo, error)

	// GetResource retrieves an MCP resource
	GetResource(ctx context.Context, uri string) (interface{}, error)
}

// ACPClient provides access to Agent Communication Protocol.
type ACPClient interface {
	// SendMessage sends a message via ACP
	SendMessage(ctx context.Context, target string, message interface{}) (interface{}, error)

	// Subscribe subscribes to ACP events
	Subscribe(ctx context.Context, eventType string) (<-chan interface{}, error)
}

// LSPClient provides access to Language Server Protocol.
type LSPClient interface {
	// GetDefinition gets symbol definition
	GetDefinition(ctx context.Context, file string, line int, char int) (interface{}, error)

	// GetReferences finds all references to a symbol
	GetReferences(ctx context.Context, file string, line int, char int) ([]interface{}, error)

	// GetDiagnostics gets code diagnostics
	GetDiagnostics(ctx context.Context, file string) ([]interface{}, error)

	// Format formats code
	Format(ctx context.Context, file string, content string) (string, error)
}

// EmbeddingClient provides access to embedding services.
type EmbeddingClient interface {
	// Embed generates embeddings for text
	Embed(ctx context.Context, text string) ([]float64, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)

	// Similarity calculates similarity between texts
	Similarity(ctx context.Context, text1, text2 string) (float64, error)
}

// RAGClient provides access to RAG (Retrieval-Augmented Generation) services.
type RAGClient interface {
	// Search performs semantic search
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)

	// Store stores documents for retrieval
	Store(ctx context.Context, docs []Document) error

	// Rerank reranks search results
	Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)
}

// FormatterRegistry provides access to code formatters.
type FormatterRegistry interface {
	// Format formats code
	Format(ctx context.Context, language string, code string) (string, error)

	// ListFormatters returns available formatters
	ListFormatters() []string
}

// ToolInfo describes an available tool.
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Category    string                 `json:"category"` // "mcp", "lsp", "rag", etc.
}

// SearchResult represents a RAG search result.
type SearchResult struct {
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Document represents a document for RAG storage.
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// InvokeMCPTool invokes an MCP tool.
func (t *ToolIntegration) InvokeMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	if !t.enabled || t.mcpClient == nil {
		return nil, fmt.Errorf("MCP integration not available")
	}

	return t.mcpClient.CallTool(ctx, toolName, args)
}

// QueryRAG performs RAG semantic search.
func (t *ToolIntegration) QueryRAG(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if !t.enabled || t.ragClient == nil {
		return nil, fmt.Errorf("RAG integration not available")
	}

	return t.ragClient.Search(ctx, query, limit)
}

// GetCodeDefinition gets symbol definition via LSP.
func (t *ToolIntegration) GetCodeDefinition(ctx context.Context, file string, line, char int) (interface{}, error) {
	if !t.enabled || t.lspClient == nil {
		return nil, fmt.Errorf("LSP integration not available")
	}

	return t.lspClient.GetDefinition(ctx, file, line, char)
}

// GenerateEmbedding generates text embeddings.
func (t *ToolIntegration) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if !t.enabled || t.embeddingClient == nil {
		return nil, fmt.Errorf("embedding integration not available")
	}

	return t.embeddingClient.Embed(ctx, text)
}

// FormatCode formats code using available formatters.
func (t *ToolIntegration) FormatCode(ctx context.Context, language, code string) (string, error) {
	if !t.enabled || t.formatterRegistry == nil {
		return code, fmt.Errorf("formatter integration not available")
	}

	return t.formatterRegistry.Format(ctx, language, code)
}

// ListAvailableTools returns all available tools across all integrations.
func (t *ToolIntegration) ListAvailableTools(ctx context.Context) ([]ToolInfo, error) {
	tools := make([]ToolInfo, 0)

	// MCP tools
	if t.mcpClient != nil {
		mcpTools, err := t.mcpClient.ListTools(ctx)
		if err == nil {
			tools = append(tools, mcpTools...)
		}
	}

	// LSP capabilities
	if t.lspClient != nil {
		tools = append(tools, ToolInfo{
			Name:        "lsp_get_definition",
			Description: "Get symbol definition",
			Category:    "lsp",
		}, ToolInfo{
			Name:        "lsp_get_references",
			Description: "Find all references",
			Category:    "lsp",
		}, ToolInfo{
			Name:        "lsp_get_diagnostics",
			Description: "Get code diagnostics",
			Category:    "lsp",
		})
	}

	// RAG capabilities
	if t.ragClient != nil {
		tools = append(tools, ToolInfo{
			Name:        "rag_search",
			Description: "Semantic search over knowledge base",
			Category:    "rag",
		})
	}

	// Embedding capabilities
	if t.embeddingClient != nil {
		tools = append(tools, ToolInfo{
			Name:        "embed_text",
			Description: "Generate text embeddings",
			Category:    "embedding",
		})
	}

	// Formatter capabilities
	if t.formatterRegistry != nil {
		formatters := t.formatterRegistry.ListFormatters()
		for _, formatter := range formatters {
			tools = append(tools, ToolInfo{
				Name:        fmt.Sprintf("format_%s", formatter),
				Description: fmt.Sprintf("Format %s code", formatter),
				Category:    "formatter",
			})
		}
	}

	return tools, nil
}

// Enable enables tool integration.
func (t *ToolIntegration) Enable() {
	t.enabled = true
}

// Disable disables tool integration.
func (t *ToolIntegration) Disable() {
	t.enabled = false
}

// IsEnabled returns whether tool integration is enabled.
func (t *ToolIntegration) IsEnabled() bool {
	return t.enabled
}
