// Package adapters provides the MCP server adapter registry.
// This file registers all available MCP server adapters for HelixAgent.
package adapters

import (
	"context"
	"fmt"
	"sync"
)

// ServerInfo contains information about an MCP server.
type ServerInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}

// ToolDefinition defines an MCP tool.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block in a tool result.
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

// MCPAdapter defines the interface for MCP server adapters.
type MCPAdapter interface {
	GetServerInfo() ServerInfo
	ListTools() []ToolDefinition
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
}

// AdapterCategory categorizes MCP adapters.
type AdapterCategory string

const (
	CategoryDatabase       AdapterCategory = "database"
	CategoryStorage        AdapterCategory = "storage"
	CategoryVersionControl AdapterCategory = "version_control"
	CategoryProductivity   AdapterCategory = "productivity"
	CategoryCommunication  AdapterCategory = "communication"
	CategorySearch         AdapterCategory = "search"
	CategoryAutomation     AdapterCategory = "automation"
	CategoryInfrastructure AdapterCategory = "infrastructure"
	CategoryAnalytics      AdapterCategory = "analytics"
	CategoryAI             AdapterCategory = "ai"
	CategoryUtility        AdapterCategory = "utility"
	CategoryDesign         AdapterCategory = "design"
	CategoryCollaboration  AdapterCategory = "collaboration"
)

// AdapterMetadata contains metadata about an adapter.
type AdapterMetadata struct {
	Name        string          `json:"name"`
	Category    AdapterCategory `json:"category"`
	Description string          `json:"description"`
	AuthType    string          `json:"authType"`    // "api_key", "oauth2", "none", "token"
	DocsURL     string          `json:"docsUrl"`
	Official    bool            `json:"official"`    // Official MCP server
	Supported   bool            `json:"supported"`   // Fully supported by HelixAgent
}

// AdapterRegistry manages MCP adapters.
type AdapterRegistry struct {
	adapters map[string]MCPAdapter
	metadata map[string]AdapterMetadata
	mu       sync.RWMutex
}

// NewAdapterRegistry creates a new adapter registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]MCPAdapter),
		metadata: make(map[string]AdapterMetadata),
	}
}

// Register registers an adapter.
func (r *AdapterRegistry) Register(name string, adapter MCPAdapter, metadata AdapterMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[name] = adapter
	r.metadata[name] = metadata
}

// Get retrieves an adapter by name.
func (r *AdapterRegistry) Get(name string) (MCPAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

// GetMetadata retrieves adapter metadata.
func (r *AdapterRegistry) GetMetadata(name string) (AdapterMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metadata, ok := r.metadata[name]
	return metadata, ok
}

// List returns all registered adapter names.
func (r *AdapterRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// ListByCategory returns adapters in a category.
func (r *AdapterRegistry) ListByCategory(category AdapterCategory) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, meta := range r.metadata {
		if meta.Category == category {
			names = append(names, name)
		}
	}
	return names
}

// ListAll returns all adapter metadata.
func (r *AdapterRegistry) ListAll() []AdapterMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]AdapterMetadata, 0, len(r.metadata))
	for _, meta := range r.metadata {
		result = append(result, meta)
	}
	return result
}

// CallTool calls a tool on an adapter.
func (r *AdapterRegistry) CallTool(ctx context.Context, adapterName, toolName string, args map[string]interface{}) (*ToolResult, error) {
	adapter, ok := r.Get(adapterName)
	if !ok {
		return nil, fmt.Errorf("adapter not found: %s", adapterName)
	}
	return adapter.CallTool(ctx, toolName, args)
}

// AvailableAdapters lists all available MCP server adapters with their metadata.
var AvailableAdapters = []AdapterMetadata{
	// Database
	{Name: "postgresql", Category: CategoryDatabase, Description: "PostgreSQL database operations", AuthType: "connection_string", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/postgres", Official: true, Supported: true},
	{Name: "sqlite", Category: CategoryDatabase, Description: "SQLite database operations", AuthType: "file_path", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/sqlite", Official: true, Supported: true},
	{Name: "mongodb", Category: CategoryDatabase, Description: "MongoDB NoSQL database", AuthType: "connection_string", DocsURL: "https://github.com/modelcontextprotocol/servers", Official: false, Supported: true},
	{Name: "redis", Category: CategoryDatabase, Description: "Redis cache and data store", AuthType: "connection_string", DocsURL: "https://github.com/modelcontextprotocol/servers", Official: false, Supported: true},
	{Name: "neon", Category: CategoryDatabase, Description: "Neon serverless Postgres", AuthType: "api_key", DocsURL: "https://neon.tech/docs", Official: false, Supported: true},
	{Name: "supabase", Category: CategoryDatabase, Description: "Supabase backend-as-a-service", AuthType: "api_key", DocsURL: "https://supabase.com/docs", Official: false, Supported: true},

	// Storage
	{Name: "aws-s3", Category: CategoryStorage, Description: "AWS S3 object storage", AuthType: "api_key", DocsURL: "https://docs.aws.amazon.com/s3/", Official: false, Supported: true},
	{Name: "google-drive", Category: CategoryStorage, Description: "Google Drive file storage", AuthType: "oauth2", DocsURL: "https://developers.google.com/drive", Official: false, Supported: true},
	{Name: "dropbox", Category: CategoryStorage, Description: "Dropbox file storage", AuthType: "oauth2", DocsURL: "https://www.dropbox.com/developers", Official: false, Supported: true},

	// Version Control
	{Name: "github", Category: CategoryVersionControl, Description: "GitHub repositories and issues", AuthType: "token", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/github", Official: true, Supported: true},
	{Name: "gitlab", Category: CategoryVersionControl, Description: "GitLab repositories and CI/CD", AuthType: "token", DocsURL: "https://docs.gitlab.com/ee/api/", Official: false, Supported: true},
	{Name: "bitbucket", Category: CategoryVersionControl, Description: "Bitbucket repositories", AuthType: "token", DocsURL: "https://developer.atlassian.com/bitbucket/api/2/reference/", Official: false, Supported: true},

	// Productivity
	{Name: "notion", Category: CategoryProductivity, Description: "Notion workspace management", AuthType: "api_key", DocsURL: "https://developers.notion.com/", Official: false, Supported: true},
	{Name: "linear", Category: CategoryProductivity, Description: "Linear issue tracking", AuthType: "api_key", DocsURL: "https://developers.linear.app/", Official: false, Supported: true},
	{Name: "todoist", Category: CategoryProductivity, Description: "Todoist task management", AuthType: "api_key", DocsURL: "https://developer.todoist.com/", Official: false, Supported: true},
	{Name: "obsidian", Category: CategoryProductivity, Description: "Obsidian note-taking", AuthType: "local", DocsURL: "https://obsidian.md/", Official: false, Supported: true},

	// Communication
	{Name: "slack", Category: CategoryCommunication, Description: "Slack workspace messaging", AuthType: "token", DocsURL: "https://api.slack.com/", Official: true, Supported: true},
	{Name: "discord", Category: CategoryCommunication, Description: "Discord server management", AuthType: "token", DocsURL: "https://discord.com/developers/docs", Official: false, Supported: true},
	{Name: "email", Category: CategoryCommunication, Description: "Email sending and reading", AuthType: "oauth2", DocsURL: "", Official: false, Supported: true},

	// Search
	{Name: "brave-search", Category: CategorySearch, Description: "Brave Search web search", AuthType: "api_key", DocsURL: "https://brave.com/search/api/", Official: true, Supported: true},
	{Name: "exa", Category: CategorySearch, Description: "Exa semantic search", AuthType: "api_key", DocsURL: "https://docs.exa.ai/", Official: false, Supported: true},
	{Name: "google-search", Category: CategorySearch, Description: "Google Custom Search", AuthType: "api_key", DocsURL: "https://developers.google.com/custom-search", Official: false, Supported: true},

	// Automation
	{Name: "puppeteer", Category: CategoryAutomation, Description: "Browser automation", AuthType: "none", DocsURL: "https://pptr.dev/", Official: true, Supported: true},
	{Name: "browserbase", Category: CategoryAutomation, Description: "Cloud browser automation", AuthType: "api_key", DocsURL: "https://docs.browserbase.com/", Official: false, Supported: true},
	{Name: "playwright", Category: CategoryAutomation, Description: "Cross-browser automation", AuthType: "none", DocsURL: "https://playwright.dev/", Official: false, Supported: true},

	// Infrastructure
	{Name: "docker", Category: CategoryInfrastructure, Description: "Docker container management", AuthType: "socket", DocsURL: "https://docs.docker.com/engine/api/", Official: false, Supported: true},
	{Name: "kubernetes", Category: CategoryInfrastructure, Description: "Kubernetes cluster management", AuthType: "kubeconfig", DocsURL: "https://kubernetes.io/docs/reference/using-api/", Official: false, Supported: true},
	{Name: "cloudflare", Category: CategoryInfrastructure, Description: "Cloudflare Workers and KV", AuthType: "api_key", DocsURL: "https://developers.cloudflare.com/", Official: false, Supported: true},
	{Name: "vercel", Category: CategoryInfrastructure, Description: "Vercel deployments", AuthType: "token", DocsURL: "https://vercel.com/docs/rest-api", Official: false, Supported: true},

	// Analytics & Observability
	{Name: "sentry", Category: CategoryAnalytics, Description: "Error tracking and monitoring", AuthType: "api_key", DocsURL: "https://docs.sentry.io/api/", Official: true, Supported: true},
	{Name: "axiom", Category: CategoryAnalytics, Description: "Log management and observability", AuthType: "api_key", DocsURL: "https://axiom.co/docs", Official: false, Supported: true},
	{Name: "datadog", Category: CategoryAnalytics, Description: "Monitoring and analytics", AuthType: "api_key", DocsURL: "https://docs.datadoghq.com/api/", Official: false, Supported: true},

	// AI & ML
	{Name: "everart", Category: CategoryAI, Description: "AI image generation", AuthType: "api_key", DocsURL: "https://everart.ai/", Official: true, Supported: true},
	{Name: "replicate", Category: CategoryAI, Description: "ML model hosting", AuthType: "api_key", DocsURL: "https://replicate.com/docs", Official: false, Supported: true},
	{Name: "huggingface", Category: CategoryAI, Description: "HuggingFace models and datasets", AuthType: "api_key", DocsURL: "https://huggingface.co/docs/api-inference", Official: false, Supported: true},

	// Utility
	{Name: "fetch", Category: CategoryUtility, Description: "HTTP requests and web fetching", AuthType: "none", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/fetch", Official: true, Supported: true},
	{Name: "time", Category: CategoryUtility, Description: "Time and timezone utilities", AuthType: "none", DocsURL: "", Official: true, Supported: true},
	{Name: "memory", Category: CategoryUtility, Description: "Persistent memory storage", AuthType: "none", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/memory", Official: true, Supported: true},
	{Name: "filesystem", Category: CategoryUtility, Description: "File system operations", AuthType: "none", DocsURL: "https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem", Official: true, Supported: true},
	{Name: "sequential-thinking", Category: CategoryUtility, Description: "Step-by-step reasoning", AuthType: "none", DocsURL: "", Official: true, Supported: true},
	{Name: "e2b", Category: CategoryUtility, Description: "Code sandbox execution", AuthType: "api_key", DocsURL: "https://e2b.dev/docs", Official: false, Supported: true},

	// Payments & Commerce
	{Name: "stripe", Category: CategoryProductivity, Description: "Payment processing", AuthType: "api_key", DocsURL: "https://stripe.com/docs/api", Official: false, Supported: true},
	{Name: "shopify", Category: CategoryProductivity, Description: "E-commerce platform", AuthType: "api_key", DocsURL: "https://shopify.dev/docs/api", Official: false, Supported: true},

	// Maps & Location
	{Name: "google-maps", Category: CategoryUtility, Description: "Google Maps and Places", AuthType: "api_key", DocsURL: "https://developers.google.com/maps", Official: true, Supported: true},

	// Design & UI
	{Name: "figma", Category: CategoryDesign, Description: "Figma design integration for reading and modifying design files", AuthType: "token", DocsURL: "https://www.figma.com/developers/api", Official: false, Supported: true},
	{Name: "svgmaker", Category: CategoryDesign, Description: "AI-powered SVG generation, editing, and optimization", AuthType: "api_key", DocsURL: "https://svgmaker.io/docs", Official: false, Supported: true},
	{Name: "stable-diffusion", Category: CategoryAI, Description: "Stable Diffusion WebUI integration for AI image generation", AuthType: "none", DocsURL: "https://github.com/AUTOMATIC1111/stable-diffusion-webui", Official: false, Supported: true},

	// Collaboration
	{Name: "miro", Category: CategoryCollaboration, Description: "Miro whiteboard collaboration for visual planning", AuthType: "oauth2", DocsURL: "https://developers.miro.com/docs", Official: false, Supported: true},
}

// DefaultRegistry is the default adapter registry.
var DefaultRegistry = NewAdapterRegistry()

// InitializeDefaultRegistry initializes the default registry with all adapters.
func InitializeDefaultRegistry() {
	for _, meta := range AvailableAdapters {
		DefaultRegistry.metadata[meta.Name] = meta
	}
}

// GetAdapterCount returns the total number of available adapters.
func GetAdapterCount() int {
	return len(AvailableAdapters)
}

// GetSupportedAdapters returns adapters that are fully supported.
func GetSupportedAdapters() []AdapterMetadata {
	var supported []AdapterMetadata
	for _, meta := range AvailableAdapters {
		if meta.Supported {
			supported = append(supported, meta)
		}
	}
	return supported
}

// GetOfficialAdapters returns official MCP adapters.
func GetOfficialAdapters() []AdapterMetadata {
	var official []AdapterMetadata
	for _, meta := range AvailableAdapters {
		if meta.Official {
			official = append(official, meta)
		}
	}
	return official
}

func init() {
	InitializeDefaultRegistry()
}
