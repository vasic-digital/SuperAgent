// Package tools provides tool integration for debate agents.
// Enables agents to invoke MCP, ACP, LSP, RAG, formatters, and other services.
package tools

import (
	"context"
	"fmt"
	"time"
)

// ToolIntegration provides access to all HelixAgent services for debate agents.
type ToolIntegration struct {
	mcpClient         MCPClient
	acpClient         ACPClient
	lspClient         LSPClient
	embeddingClient   EmbeddingClient
	ragClient         RAGClient
	formatterRegistry FormatterRegistry
	visionClient      VisionClient
	memoryClient      HelixMemoryClient
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

// VisionClient provides access to image/screenshot analysis.
type VisionClient interface {
	// AnalyzeImage analyzes raw image data with a prompt
	AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (interface{}, error)

	// AnalyzeURL analyzes an image at a URL with a prompt
	AnalyzeURL(ctx context.Context, imageURL string, prompt string) (interface{}, error)
}

// HelixMemoryClient provides full 4-engine memory access (Mem0, Cognee, Letta, Graphiti).
type HelixMemoryClient interface {
	// StoreFact stores a fact with metadata via Mem0
	StoreFact(ctx context.Context, fact string, metadata map[string]string) error

	// RecallFacts recalls facts matching a query via Mem0
	RecallFacts(ctx context.Context, query string, limit int) ([]string, error)

	// BuildGraph builds a knowledge graph from content via Cognee
	BuildGraph(ctx context.Context, content string) error

	// QueryGraph queries the knowledge graph via Cognee
	QueryGraph(ctx context.Context, query string) ([]interface{}, error)

	// CreateAgentSession creates a stateful agent session via Letta
	CreateAgentSession(ctx context.Context, agentID string) (string, error)

	// SendAgentMessage sends a message to a Letta agent session
	SendAgentMessage(ctx context.Context, sessionID string, message string) (string, error)

	// AddTemporalEdge adds a temporal edge to the Graphiti graph
	AddTemporalEdge(ctx context.Context, from, to, relation string, timestamp time.Time) error

	// QueryTimeline queries the temporal timeline for an entity via Graphiti
	QueryTimeline(ctx context.Context, entity string) ([]interface{}, error)
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

	// Vision capabilities
	if t.visionClient != nil {
		tools = append(tools, ToolInfo{
			Name:        "vision_analyze_image",
			Description: "Analyze image data with a prompt",
			Category:    "vision",
		}, ToolInfo{
			Name:        "vision_analyze_url",
			Description: "Analyze image at URL with a prompt",
			Category:    "vision",
		})
	}

	// Memory capabilities
	if t.memoryClient != nil {
		tools = append(tools, ToolInfo{
			Name:        "memory_store_fact",
			Description: "Store a fact in HelixMemory (Mem0)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_recall",
			Description: "Recall facts from HelixMemory (Mem0)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_build_graph",
			Description: "Build knowledge graph (Cognee)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_query_graph",
			Description: "Query knowledge graph (Cognee)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_agent_session",
			Description: "Create stateful agent session (Letta)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_agent_message",
			Description: "Send message to agent session (Letta)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_temporal_edge",
			Description: "Add temporal edge to graph (Graphiti)",
			Category:    "memory",
		}, ToolInfo{
			Name:        "memory_query_timeline",
			Description: "Query temporal timeline (Graphiti)",
			Category:    "memory",
		})
	}

	return tools, nil
}

// GetVisionClient returns the vision client, or nil if not configured.
func (t *ToolIntegration) GetVisionClient() VisionClient {
	return t.visionClient
}

// GetMemoryClient returns the HelixMemory client, or nil if not configured.
func (t *ToolIntegration) GetMemoryClient() HelixMemoryClient {
	return t.memoryClient
}

// SetVisionClient sets the vision client for image analysis capabilities.
func (t *ToolIntegration) SetVisionClient(c VisionClient) {
	t.visionClient = c
}

// SetMemoryClient sets the HelixMemory client for 4-engine memory access.
func (t *ToolIntegration) SetMemoryClient(c HelixMemoryClient) {
	t.memoryClient = c
}

// AnalyzeImage analyzes an image using the vision client.
func (t *ToolIntegration) AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (interface{}, error) {
	if !t.enabled || t.visionClient == nil {
		return nil, fmt.Errorf("vision integration not available")
	}

	return t.visionClient.AnalyzeImage(ctx, imageData, prompt)
}

// AnalyzeImageURL analyzes an image at a URL using the vision client.
func (t *ToolIntegration) AnalyzeImageURL(ctx context.Context, imageURL string, prompt string) (interface{}, error) {
	if !t.enabled || t.visionClient == nil {
		return nil, fmt.Errorf("vision integration not available")
	}

	return t.visionClient.AnalyzeURL(ctx, imageURL, prompt)
}

// RecallMemory recalls facts from HelixMemory.
func (t *ToolIntegration) RecallMemory(ctx context.Context, query string, limit int) ([]string, error) {
	if !t.enabled || t.memoryClient == nil {
		return nil, fmt.Errorf("memory integration not available")
	}

	return t.memoryClient.RecallFacts(ctx, query, limit)
}

// StoreMemory stores a fact in HelixMemory.
func (t *ToolIntegration) StoreMemory(ctx context.Context, fact string, metadata map[string]string) error {
	if !t.enabled || t.memoryClient == nil {
		return fmt.Errorf("memory integration not available")
	}

	return t.memoryClient.StoreFact(ctx, fact, metadata)
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
