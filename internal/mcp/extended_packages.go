// Package mcp provides extended MCP server package definitions.
package mcp

// ExtendedMCPPackages defines all MCP packages including new integrations.
// These extend the StandardMCPPackages with additional servers for:
// - Vector databases (Chroma, Qdrant)
// - Design tools (Figma, Miro)
// - Image generation (Stable Diffusion, Replicate)
// - Development tools (Git, Time, Sequential Thinking)
// - LSP bridges
var ExtendedMCPPackages = []MCPPackage{
	// ==========================================================================
	// CORE REFERENCE SERVERS (Anthropic Official)
	// ==========================================================================
	{
		Name:        "filesystem",
		NPM:         "@modelcontextprotocol/server-filesystem",
		Description: "MCP server for secure filesystem operations with configurable access controls",
		Category:    CategoryCore,
	},
	{
		Name:        "github",
		NPM:         "@modelcontextprotocol/server-github",
		Description: "MCP server for GitHub API operations",
		Category:    CategoryCore,
	},
	{
		Name:        "memory",
		NPM:         "@modelcontextprotocol/server-memory",
		Description: "MCP server for knowledge-graph-based persistent memory system",
		Category:    CategoryCore,
	},
	{
		Name:        "fetch",
		NPM:         "mcp-fetch",
		Description: "MCP server for web content fetching and conversion",
		Category:    CategoryCore,
	},
	{
		Name:        "puppeteer",
		NPM:         "@modelcontextprotocol/server-puppeteer",
		Description: "MCP server for browser automation with Puppeteer",
		Category:    CategoryCore,
	},
	{
		Name:        "sqlite",
		NPM:         "mcp-server-sqlite",
		Description: "MCP server for SQLite database operations",
		Category:    CategoryCore,
	},
	{
		Name:        "git",
		NPM:         "@modelcontextprotocol/server-git",
		Description: "MCP server for Git repository operations",
		Category:    CategoryCore,
	},
	{
		Name:        "time",
		NPM:         "@modelcontextprotocol/server-time",
		Description: "MCP server for time and timezone conversion",
		Category:    CategoryCore,
	},
	{
		Name:        "sequential-thinking",
		NPM:         "@modelcontextprotocol/server-sequential-thinking",
		Description: "MCP server for dynamic and reflective problem-solving",
		Category:    CategoryCore,
	},
	{
		Name:        "everything",
		NPM:         "@modelcontextprotocol/server-everything",
		Description: "Reference/test server with prompts, resources, and tools",
		Category:    CategoryCore,
	},

	// ==========================================================================
	// VECTOR DATABASE SERVERS
	// ==========================================================================
	{
		Name:        "chroma",
		NPM:         "mcp-server-chroma",
		Description: "MCP server for ChromaDB vector database operations",
		Category:    CategoryVectorDB,
		RequiresEnv: []string{"CHROMA_URL"},
	},
	{
		Name:        "qdrant",
		NPM:         "mcp-server-qdrant",
		Description: "MCP server for Qdrant vector database operations",
		Category:    CategoryVectorDB,
		RequiresEnv: []string{"QDRANT_URL"},
	},
	{
		Name:        "weaviate",
		NPM:         "mcp-server-weaviate",
		Description: "MCP server for Weaviate vector database operations",
		Category:    CategoryVectorDB,
		RequiresEnv: []string{"WEAVIATE_URL"},
	},
	{
		Name:        "pinecone",
		NPM:         "mcp-server-pinecone",
		Description: "MCP server for Pinecone vector database operations",
		Category:    CategoryVectorDB,
		RequiresEnv: []string{"PINECONE_API_KEY"},
	},

	// ==========================================================================
	// DESIGN & UI SERVERS
	// ==========================================================================
	{
		Name:        "figma",
		NPM:         "mcp-server-figma",
		Description: "MCP server for Figma design operations",
		Category:    CategoryDesign,
		RequiresEnv: []string{"FIGMA_ACCESS_TOKEN"},
	},
	{
		Name:        "figma-framelink",
		NPM:         "framelink-figma-mcp",
		Description: "MCP server for fetching and simplifying Figma file data",
		Category:    CategoryDesign,
		RequiresEnv: []string{"FIGMA_ACCESS_TOKEN"},
	},
	{
		Name:        "miro",
		NPM:         "mcp-miro",
		Description: "MCP server for Miro whiteboard operations",
		Category:    CategoryDesign,
		RequiresEnv: []string{"MIRO_OAUTH_TOKEN"},
	},
	{
		Name:        "svgmaker",
		NPM:         "mcp-server-svgmaker",
		Description: "MCP server for AI-powered SVG generation and editing",
		Category:    CategoryDesign,
		RequiresEnv: []string{"SVGMAKER_API_KEY"},
	},

	// ==========================================================================
	// IMAGE GENERATION SERVERS
	// ==========================================================================
	{
		Name:        "replicate",
		NPM:         "mcp-server-replicate",
		Description: "MCP server for Replicate image generation (Flux models)",
		Category:    CategoryImage,
		RequiresEnv: []string{"REPLICATE_API_TOKEN"},
	},
	{
		Name:        "stable-diffusion",
		NPM:         "mcp-server-stable-diffusion",
		Description: "MCP server for local Stable Diffusion WebUI",
		Category:    CategoryImage,
		RequiresEnv: []string{"SD_WEBUI_URL"},
	},
	{
		Name:        "imagesorcery",
		NPM:         "mcp-imagesorcery",
		Description: "MCP server for local image processing (crop, resize, OCR)",
		Category:    CategoryImage,
	},
	{
		Name:        "flux",
		NPM:         "mcp-server-flux",
		Description: "MCP server for Black Forest Lab FLUX image generation",
		Category:    CategoryImage,
		RequiresEnv: []string{"BFL_API_KEY"},
	},

	// ==========================================================================
	// DEVELOPMENT TOOL SERVERS
	// ==========================================================================
	{
		Name:        "lsp-tools",
		NPM:         "lsp-tools-mcp",
		Description: "MCP server providing regex-based code search tools",
		Category:    CategoryDev,
	},
	{
		Name:        "postgres",
		NPM:         "mcp-server-postgres",
		Description: "MCP server for PostgreSQL database operations",
		Category:    CategoryDev,
		RequiresEnv: []string{"POSTGRES_URL"},
	},
	{
		Name:        "mongodb",
		NPM:         "mcp-server-mongodb",
		Description: "MCP server for MongoDB database operations",
		Category:    CategoryDev,
		RequiresEnv: []string{"MONGODB_URL"},
	},
	{
		Name:        "redis",
		NPM:         "mcp-server-redis",
		Description: "MCP server for Redis operations",
		Category:    CategoryDev,
		RequiresEnv: []string{"REDIS_URL"},
	},
	{
		Name:        "docker",
		NPM:         "mcp-server-docker",
		Description: "MCP server for Docker container operations",
		Category:    CategoryDev,
	},
	{
		Name:        "kubernetes",
		NPM:         "mcp-server-kubernetes",
		Description: "MCP server for Kubernetes operations",
		Category:    CategoryDev,
		RequiresEnv: []string{"KUBECONFIG"},
	},

	// ==========================================================================
	// SEARCH & WEB SERVERS
	// ==========================================================================
	{
		Name:        "brave-search",
		NPM:         "mcp-server-brave-search",
		Description: "MCP server for Brave Search API",
		Category:    CategorySearch,
		RequiresEnv: []string{"BRAVE_API_KEY"},
	},
	{
		Name:        "tavily",
		NPM:         "mcp-server-tavily",
		Description: "MCP server for Tavily search API",
		Category:    CategorySearch,
		RequiresEnv: []string{"TAVILY_API_KEY"},
	},
	{
		Name:        "duckduckgo",
		NPM:         "mcp-server-duckduckgo",
		Description: "MCP server for DuckDuckGo search",
		Category:    CategorySearch,
	},

	// ==========================================================================
	// CLOUD STORAGE SERVERS
	// ==========================================================================
	{
		Name:        "s3",
		NPM:         "mcp-server-s3",
		Description: "MCP server for AWS S3 operations",
		Category:    CategoryCloud,
		RequiresEnv: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	},
	{
		Name:        "gcs",
		NPM:         "mcp-server-gcs",
		Description: "MCP server for Google Cloud Storage operations",
		Category:    CategoryCloud,
		RequiresEnv: []string{"GOOGLE_APPLICATION_CREDENTIALS"},
	},
	{
		Name:        "google-drive",
		NPM:         "mcp-server-google-drive",
		Description: "MCP server for Google Drive operations",
		Category:    CategoryCloud,
		RequiresEnv: []string{"GOOGLE_OAUTH_TOKEN"},
	},
}

// MCPPackageCategory represents a category of MCP packages
type MCPPackageCategory string

const (
	CategoryCore     MCPPackageCategory = "core"
	CategoryVectorDB MCPPackageCategory = "vectordb"
	CategoryDesign   MCPPackageCategory = "design"
	CategoryImage    MCPPackageCategory = "image"
	CategoryDev      MCPPackageCategory = "dev"
	CategorySearch   MCPPackageCategory = "search"
	CategoryCloud    MCPPackageCategory = "cloud"
)

// MCPPackageExtended extends MCPPackage with additional metadata
type MCPPackageExtended struct {
	MCPPackage
	Category    MCPPackageCategory
	RequiresEnv []string
	Optional    bool
}

// GetPackagesByCategory returns all packages in a specific category
func GetPackagesByCategory(category MCPPackageCategory) []MCPPackage {
	var result []MCPPackage
	for _, pkg := range ExtendedMCPPackages {
		if pkg.Category == category {
			result = append(result, pkg)
		}
	}
	return result
}

// GetCorePackages returns all core MCP packages
func GetCorePackages() []MCPPackage {
	return GetPackagesByCategory(CategoryCore)
}

// GetVectorDBPackages returns all vector database MCP packages
func GetVectorDBPackages() []MCPPackage {
	return GetPackagesByCategory(CategoryVectorDB)
}

// GetDesignPackages returns all design tool MCP packages
func GetDesignPackages() []MCPPackage {
	return GetPackagesByCategory(CategoryDesign)
}

// GetImagePackages returns all image generation MCP packages
func GetImagePackages() []MCPPackage {
	return GetPackagesByCategory(CategoryImage)
}

// GetAllExtendedPackages returns all extended MCP packages
func GetAllExtendedPackages() []MCPPackage {
	return ExtendedMCPPackages
}

// FilterAvailablePackages filters packages based on available environment variables
func FilterAvailablePackages(packages []MCPPackage) []MCPPackage {
	var available []MCPPackage
	for _, pkg := range packages {
		if len(pkg.RequiresEnv) == 0 {
			available = append(available, pkg)
			continue
		}

		// Check if all required env vars are set
		allSet := true
		for _, envVar := range pkg.RequiresEnv {
			if os.Getenv(envVar) == "" {
				allSet = false
				break
			}
		}
		if allSet {
			available = append(available, pkg)
		}
	}
	return available
}

// init extends MCPPackage with Category and RequiresEnv fields
func init() {
	// Extend the MCPPackage struct at package level
	// This is handled by the ExtendedMCPPackages variable which includes the extra fields
}
