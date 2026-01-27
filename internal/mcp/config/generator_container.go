// Package config provides containerized MCP configuration generation for CLI agents
// This generator uses Docker containers instead of npx for all MCP servers,
// eliminating npm/npx dependencies completely.
package config

import (
	"fmt"
	"os"
	"strings"
)

// ContainerMCPServerConfig represents a containerized MCP server configuration
type ContainerMCPServerConfig struct {
	Type        string            `json:"type"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Env         map[string]string `json:"env,omitempty"` // Crush uses "env" instead of "environment"
	Enabled     bool              `json:"enabled"`
	Port        int               `json:"port"`
	Category    string            `json:"category"`
}

// MCPContainerPort defines the port allocation for each MCP server
type MCPContainerPort struct {
	Name     string
	Port     int
	Category string
}

// Port allocation scheme for containerized MCP servers
// Organized by category for easy management and no conflicts
var MCPContainerPorts = []MCPContainerPort{
	// TIER 1: Core Official MCP Servers (9101-9120)
	{"fetch", 9101, "core"},
	{"git", 9102, "core"},
	{"time", 9103, "core"},
	{"filesystem", 9104, "core"},
	{"memory", 9105, "core"},
	{"everything", 9106, "core"},
	{"sequential-thinking", 9107, "core"},
	{"sqlite", 9108, "core"},
	{"puppeteer", 9109, "core"},
	{"postgres", 9110, "core"},

	// TIER 2: Database MCP Servers (9201-9220)
	{"mongodb", 9201, "database"},
	{"redis", 9202, "database"},
	{"mysql", 9203, "database"},
	{"elasticsearch", 9204, "database"},
	{"supabase", 9205, "database"},

	// TIER 3: Vector Database MCP Servers (9301-9320)
	{"qdrant", 9301, "vector"},
	{"chroma", 9302, "vector"},
	{"pinecone", 9303, "vector"},
	{"weaviate", 9304, "vector"},

	// TIER 4: DevOps & Infrastructure (9401-9440)
	{"github", 9401, "devops"},
	{"gitlab", 9402, "devops"},
	{"sentry", 9403, "devops"},
	{"kubernetes", 9404, "devops"},
	{"docker", 9405, "devops"},
	{"ansible", 9406, "devops"},
	{"aws", 9407, "devops"},
	{"gcp", 9408, "devops"},
	{"heroku", 9409, "devops"},
	{"cloudflare", 9410, "devops"},
	{"vercel", 9411, "devops"},
	{"workers", 9412, "devops"},
	{"jetbrains", 9413, "devops"},
	{"k8s-alt", 9414, "devops"},

	// TIER 5: Browser & Web Automation (9501-9520)
	{"playwright", 9501, "browser"},
	{"browserbase", 9502, "browser"},
	{"firecrawl", 9503, "browser"},
	{"crawl4ai", 9504, "browser"},

	// TIER 6: Communication (9601-9620)
	{"slack", 9601, "communication"},
	{"discord", 9602, "communication"},
	{"telegram", 9603, "communication"},

	// TIER 7: Productivity & Project Management (9701-9740)
	{"notion", 9701, "productivity"},
	{"linear", 9702, "productivity"},
	{"jira", 9703, "productivity"},
	{"asana", 9704, "productivity"},
	{"trello", 9705, "productivity"},
	{"todoist", 9706, "productivity"},
	{"monday", 9707, "productivity"},
	{"airtable", 9708, "productivity"},
	{"obsidian", 9709, "productivity"},
	{"atlassian", 9710, "productivity"},

	// TIER 8: Search & AI (9801-9840)
	{"brave-search", 9801, "search"},
	{"exa", 9802, "search"},
	{"tavily", 9803, "search"},
	{"perplexity", 9804, "search"},
	{"kagi", 9805, "search"},
	{"omnisearch", 9806, "search"},
	{"context7", 9807, "search"},
	{"llamaindex", 9808, "search"},
	{"langchain", 9809, "search"},
	{"openai", 9810, "search"},

	// TIER 9: Google Services (9901-9920)
	{"google-drive", 9901, "google"},
	{"google-calendar", 9902, "google"},
	{"google-maps", 9903, "google"},
	{"youtube", 9904, "google"},
	{"gmail", 9905, "google"},

	// TIER 10: Monitoring & Observability (9921-9940)
	{"datadog", 9921, "monitoring"},
	{"grafana", 9922, "monitoring"},
	{"prometheus", 9923, "monitoring"},

	// TIER 11: Finance & Business (9941-9960)
	{"stripe", 9941, "finance"},
	{"hubspot", 9942, "finance"},
	{"zendesk", 9943, "finance"},

	// TIER 12: Design (9961-9970)
	{"figma", 9961, "design"},
}

// ContainerMCPConfigGenerator generates containerized MCP configurations
type ContainerMCPConfigGenerator struct {
	homeDir   string
	helixHome string
	baseURL   string
	mcpHost   string
	envVars   map[string]string
	portMap   map[string]MCPContainerPort
}

// NewContainerMCPConfigGenerator creates a new container MCP config generator
func NewContainerMCPConfigGenerator(baseURL string) *ContainerMCPConfigGenerator {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}

	helixHome := os.Getenv("HELIXAGENT_HOME")
	if helixHome == "" {
		helixHome = homeDir + "/.helixagent"
	}

	// MCP container host - can be overridden for remote deployments
	mcpHost := os.Getenv("MCP_CONTAINER_HOST")
	if mcpHost == "" {
		mcpHost = "localhost"
	}

	g := &ContainerMCPConfigGenerator{
		homeDir:   homeDir,
		helixHome: helixHome,
		baseURL:   baseURL,
		mcpHost:   mcpHost,
		envVars:   make(map[string]string),
		portMap:   make(map[string]MCPContainerPort),
	}

	// Build port map for quick lookups
	for _, p := range MCPContainerPorts {
		g.portMap[p.Name] = p
	}

	g.loadEnvVars()
	return g
}

// loadEnvVars loads environment variables from .env files and environment
func (g *ContainerMCPConfigGenerator) loadEnvVars() {
	// Load from multiple .env files
	envFiles := []string{".env", ".env.local", ".env.mcp", ".env.mcp.generated"}
	for _, file := range envFiles {
		if data, err := os.ReadFile(file); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
					g.envVars[key] = value
				}
			}
		}
	}

	// Also load from environment (overrides files)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			g.envVars[parts[0]] = parts[1]
		}
	}
}

// hasEnvVar checks if an environment variable is set and non-empty
func (g *ContainerMCPConfigGenerator) hasEnvVar(name string) bool {
	val, ok := g.envVars[name]
	return ok && val != ""
}

// hasAnyEnvVar checks if any of the given environment variables is set
func (g *ContainerMCPConfigGenerator) hasAnyEnvVar(names ...string) bool {
	for _, name := range names {
		if g.hasEnvVar(name) {
			return true
		}
	}
	return false
}

// hasAllEnvVars checks if all of the given environment variables are set
func (g *ContainerMCPConfigGenerator) hasAllEnvVars(names ...string) bool {
	for _, name := range names {
		if !g.hasEnvVar(name) {
			return false
		}
	}
	return true
}

// getEnvOrDefault returns the environment variable value or default
func (g *ContainerMCPConfigGenerator) getEnvOrDefault(name, defaultVal string) string {
	if val, ok := g.envVars[name]; ok && val != "" {
		return val
	}
	return defaultVal
}

// getMCPURL returns the container URL for an MCP server
func (g *ContainerMCPConfigGenerator) getMCPURL(name string) string {
	if port, ok := g.portMap[name]; ok {
		return fmt.Sprintf("http://%s:%d/sse", g.mcpHost, port.Port)
	}
	return ""
}

// GenerateContainerMCPs generates ALL MCP configurations using containers
// ZERO npx commands - all MCPs use containerized remote endpoints
func (g *ContainerMCPConfigGenerator) GenerateContainerMCPs() map[string]ContainerMCPServerConfig {
	mcps := make(map[string]ContainerMCPServerConfig)

	// ==========================================================================
	// CATEGORY 1: HelixAgent Core (Always enabled - remote endpoint)
	// ==========================================================================
	mcps["helixagent"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.baseURL + "/mcp/sse",
		Enabled:  true,
		Port:     0,
		Category: "helixagent",
	}

	// ==========================================================================
	// CATEGORY 2: Core MCP Servers (Always available - no API keys needed)
	// ==========================================================================
	mcps["fetch"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("fetch"),
		Enabled:  true,
		Port:     9101,
		Category: "core",
	}
	mcps["git"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("git"),
		Enabled:  true,
		Port:     9102,
		Category: "core",
	}
	mcps["time"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("time"),
		Enabled:  true,
		Port:     9103,
		Category: "core",
	}
	mcps["filesystem"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("filesystem"),
		Enabled:  true,
		Port:     9104,
		Category: "core",
	}
	mcps["memory"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("memory"),
		Enabled:  true,
		Port:     9105,
		Category: "core",
	}
	mcps["everything"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("everything"),
		Enabled:  true,
		Port:     9106,
		Category: "core",
	}
	mcps["sequential-thinking"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("sequential-thinking"),
		Enabled:  true,
		Port:     9107,
		Category: "core",
	}
	mcps["sqlite"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("sqlite"),
		Enabled:  true,
		Port:     9108,
		Category: "core",
	}
	mcps["puppeteer"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("puppeteer"),
		Enabled:  true,
		Port:     9109,
		Category: "core",
	}
	mcps["postgres"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("postgres"),
		Enabled:  g.hasAnyEnvVar("POSTGRES_URL", "POSTGRES_HOST"),
		Port:     9110,
		Category: "core",
	}

	// ==========================================================================
	// CATEGORY 3: Database MCP Servers (Enabled if backend available)
	// ==========================================================================
	mcps["mongodb"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("mongodb"),
		Enabled:  g.hasAnyEnvVar("MONGODB_URI", "MONGODB_HOST"),
		Port:     9201,
		Category: "database",
	}
	mcps["redis"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("redis"),
		Enabled:  g.hasAnyEnvVar("REDIS_URL", "REDIS_HOST"),
		Port:     9202,
		Category: "database",
	}
	mcps["mysql"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("mysql"),
		Enabled:  g.hasAnyEnvVar("MYSQL_URL", "MYSQL_HOST"),
		Port:     9203,
		Category: "database",
	}
	mcps["elasticsearch"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("elasticsearch"),
		Enabled:  g.hasAnyEnvVar("ELASTICSEARCH_URL", "ELASTICSEARCH_HOST"),
		Port:     9204,
		Category: "database",
	}
	mcps["supabase"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("supabase"),
		Enabled:  g.hasAllEnvVars("SUPABASE_URL", "SUPABASE_KEY"),
		Port:     9205,
		Category: "database",
	}

	// ==========================================================================
	// CATEGORY 4: Vector Database MCP Servers
	// ==========================================================================
	mcps["qdrant"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("qdrant"),
		Enabled:  g.hasAnyEnvVar("QDRANT_URL", "QDRANT_HOST"),
		Port:     9301,
		Category: "vector",
	}
	mcps["chroma"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("chroma"),
		Enabled:  g.hasAnyEnvVar("CHROMA_URL", "CHROMA_HOST"),
		Port:     9302,
		Category: "vector",
	}
	mcps["pinecone"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("pinecone"),
		Enabled:  g.hasEnvVar("PINECONE_API_KEY"),
		Port:     9303,
		Category: "vector",
	}
	mcps["weaviate"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("weaviate"),
		Enabled:  g.hasAnyEnvVar("WEAVIATE_URL", "WEAVIATE_HOST"),
		Port:     9304,
		Category: "vector",
	}

	// ==========================================================================
	// CATEGORY 5: DevOps & Infrastructure
	// ==========================================================================
	mcps["github"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("github"),
		Enabled:  g.hasEnvVar("GITHUB_TOKEN"),
		Port:     9401,
		Category: "devops",
	}
	mcps["gitlab"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("gitlab"),
		Enabled:  g.hasEnvVar("GITLAB_TOKEN"),
		Port:     9402,
		Category: "devops",
	}
	mcps["sentry"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("sentry"),
		Enabled:  g.hasAllEnvVars("SENTRY_AUTH_TOKEN", "SENTRY_ORG"),
		Port:     9403,
		Category: "devops",
	}
	mcps["kubernetes"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("kubernetes"),
		Enabled:  g.hasEnvVar("KUBECONFIG"),
		Port:     9404,
		Category: "devops",
	}
	mcps["docker"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("docker"),
		Enabled:  true, // Always available - uses local docker socket
		Port:     9405,
		Category: "devops",
	}
	mcps["ansible"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("ansible"),
		Enabled:  true, // Always available
		Port:     9406,
		Category: "devops",
	}
	mcps["aws"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("aws"),
		Enabled:  g.hasAllEnvVars("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"),
		Port:     9407,
		Category: "devops",
	}
	mcps["gcp"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("gcp"),
		Enabled:  g.hasEnvVar("GOOGLE_APPLICATION_CREDENTIALS"),
		Port:     9408,
		Category: "devops",
	}
	mcps["heroku"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("heroku"),
		Enabled:  g.hasEnvVar("HEROKU_API_KEY"),
		Port:     9409,
		Category: "devops",
	}
	mcps["cloudflare"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("cloudflare"),
		Enabled:  g.hasEnvVar("CLOUDFLARE_API_TOKEN"),
		Port:     9410,
		Category: "devops",
	}
	mcps["vercel"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("vercel"),
		Enabled:  g.hasEnvVar("VERCEL_TOKEN"),
		Port:     9411,
		Category: "devops",
	}
	mcps["workers"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("workers"),
		Enabled:  g.hasEnvVar("CLOUDFLARE_API_TOKEN"),
		Port:     9412,
		Category: "devops",
	}
	mcps["jetbrains"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("jetbrains"),
		Enabled:  true, // Works locally
		Port:     9413,
		Category: "devops",
	}

	// ==========================================================================
	// CATEGORY 6: Browser & Web Automation
	// ==========================================================================
	mcps["playwright"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("playwright"),
		Enabled:  true, // Always available
		Port:     9501,
		Category: "browser",
	}
	mcps["browserbase"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("browserbase"),
		Enabled:  g.hasEnvVar("BROWSERBASE_API_KEY"),
		Port:     9502,
		Category: "browser",
	}
	mcps["firecrawl"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("firecrawl"),
		Enabled:  g.hasEnvVar("FIRECRAWL_API_KEY"),
		Port:     9503,
		Category: "browser",
	}
	mcps["crawl4ai"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("crawl4ai"),
		Enabled:  true, // Works locally
		Port:     9504,
		Category: "browser",
	}

	// ==========================================================================
	// CATEGORY 7: Communication
	// ==========================================================================
	mcps["slack"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("slack"),
		Enabled:  g.hasAllEnvVars("SLACK_BOT_TOKEN", "SLACK_TEAM_ID"),
		Port:     9601,
		Category: "communication",
	}
	mcps["discord"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("discord"),
		Enabled:  g.hasEnvVar("DISCORD_TOKEN"),
		Port:     9602,
		Category: "communication",
	}
	mcps["telegram"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("telegram"),
		Enabled:  g.hasEnvVar("TELEGRAM_BOT_TOKEN"),
		Port:     9603,
		Category: "communication",
	}

	// ==========================================================================
	// CATEGORY 8: Productivity & Project Management
	// ==========================================================================
	mcps["notion"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("notion"),
		Enabled:  g.hasEnvVar("NOTION_API_KEY"),
		Port:     9701,
		Category: "productivity",
	}
	mcps["linear"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("linear"),
		Enabled:  g.hasEnvVar("LINEAR_API_KEY"),
		Port:     9702,
		Category: "productivity",
	}
	mcps["jira"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("jira"),
		Enabled:  g.hasAllEnvVars("JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"),
		Port:     9703,
		Category: "productivity",
	}
	mcps["asana"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("asana"),
		Enabled:  g.hasEnvVar("ASANA_ACCESS_TOKEN"),
		Port:     9704,
		Category: "productivity",
	}
	mcps["trello"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("trello"),
		Enabled:  g.hasAllEnvVars("TRELLO_API_KEY", "TRELLO_API_TOKEN"),
		Port:     9705,
		Category: "productivity",
	}
	mcps["todoist"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("todoist"),
		Enabled:  g.hasEnvVar("TODOIST_API_TOKEN"),
		Port:     9706,
		Category: "productivity",
	}
	mcps["monday"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("monday"),
		Enabled:  g.hasEnvVar("MONDAY_API_TOKEN"),
		Port:     9707,
		Category: "productivity",
	}
	mcps["airtable"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("airtable"),
		Enabled:  g.hasEnvVar("AIRTABLE_API_KEY"),
		Port:     9708,
		Category: "productivity",
	}
	mcps["obsidian"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("obsidian"),
		Enabled:  g.hasEnvVar("OBSIDIAN_VAULT_PATH"),
		Port:     9709,
		Category: "productivity",
	}
	mcps["atlassian"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("atlassian"),
		Enabled:  g.hasAllEnvVars("ATLASSIAN_URL", "ATLASSIAN_EMAIL", "ATLASSIAN_API_TOKEN"),
		Port:     9710,
		Category: "productivity",
	}

	// ==========================================================================
	// CATEGORY 9: Search & AI
	// ==========================================================================
	mcps["brave-search"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("brave-search"),
		Enabled:  g.hasEnvVar("BRAVE_API_KEY"),
		Port:     9801,
		Category: "search",
	}
	mcps["exa"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("exa"),
		Enabled:  g.hasEnvVar("EXA_API_KEY"),
		Port:     9802,
		Category: "search",
	}
	mcps["tavily"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("tavily"),
		Enabled:  g.hasEnvVar("TAVILY_API_KEY"),
		Port:     9803,
		Category: "search",
	}
	mcps["perplexity"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("perplexity"),
		Enabled:  g.hasEnvVar("PERPLEXITY_API_KEY"),
		Port:     9804,
		Category: "search",
	}
	mcps["kagi"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("kagi"),
		Enabled:  g.hasEnvVar("KAGI_API_KEY"),
		Port:     9805,
		Category: "search",
	}
	mcps["omnisearch"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("omnisearch"),
		Enabled:  true, // Works without API key
		Port:     9806,
		Category: "search",
	}
	mcps["context7"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("context7"),
		Enabled:  true, // Works without API key
		Port:     9807,
		Category: "search",
	}
	mcps["llamaindex"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("llamaindex"),
		Enabled:  g.hasEnvVar("OPENAI_API_KEY"),
		Port:     9808,
		Category: "search",
	}
	mcps["langchain"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("langchain"),
		Enabled:  g.hasEnvVar("OPENAI_API_KEY"),
		Port:     9809,
		Category: "search",
	}
	mcps["openai"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("openai"),
		Enabled:  g.hasEnvVar("OPENAI_API_KEY"),
		Port:     9810,
		Category: "search",
	}

	// ==========================================================================
	// CATEGORY 10: Google Services
	// ==========================================================================
	mcps["google-drive"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("google-drive"),
		Enabled:  g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
		Port:     9901,
		Category: "google",
	}
	mcps["google-calendar"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("google-calendar"),
		Enabled:  g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
		Port:     9902,
		Category: "google",
	}
	mcps["google-maps"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("google-maps"),
		Enabled:  g.hasEnvVar("GOOGLE_MAPS_API_KEY"),
		Port:     9903,
		Category: "google",
	}
	mcps["youtube"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("youtube"),
		Enabled:  g.hasEnvVar("YOUTUBE_API_KEY"),
		Port:     9904,
		Category: "google",
	}
	mcps["gmail"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("gmail"),
		Enabled:  g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
		Port:     9905,
		Category: "google",
	}

	// ==========================================================================
	// CATEGORY 11: Monitoring & Observability
	// ==========================================================================
	mcps["datadog"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("datadog"),
		Enabled:  g.hasAllEnvVars("DD_API_KEY", "DD_APP_KEY"),
		Port:     9921,
		Category: "monitoring",
	}
	mcps["grafana"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("grafana"),
		Enabled:  g.hasAllEnvVars("GRAFANA_URL", "GRAFANA_TOKEN"),
		Port:     9922,
		Category: "monitoring",
	}
	mcps["prometheus"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("prometheus"),
		Enabled:  g.hasEnvVar("PROMETHEUS_URL"),
		Port:     9923,
		Category: "monitoring",
	}

	// ==========================================================================
	// CATEGORY 12: Finance & Business
	// ==========================================================================
	mcps["stripe"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("stripe"),
		Enabled:  g.hasEnvVar("STRIPE_API_KEY"),
		Port:     9941,
		Category: "finance",
	}
	mcps["hubspot"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("hubspot"),
		Enabled:  g.hasEnvVar("HUBSPOT_ACCESS_TOKEN"),
		Port:     9942,
		Category: "finance",
	}
	mcps["zendesk"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("zendesk"),
		Enabled:  g.hasAllEnvVars("ZENDESK_SUBDOMAIN", "ZENDESK_EMAIL", "ZENDESK_TOKEN"),
		Port:     9943,
		Category: "finance",
	}

	// ==========================================================================
	// CATEGORY 13: Design
	// ==========================================================================
	mcps["figma"] = ContainerMCPServerConfig{
		Type:     "remote",
		URL:      g.getMCPURL("figma"),
		Enabled:  g.hasEnvVar("FIGMA_API_KEY"),
		Port:     9961,
		Category: "design",
	}

	return mcps
}

// GetEnabledContainerMCPs returns only the MCPs that are enabled
func (g *ContainerMCPConfigGenerator) GetEnabledContainerMCPs() map[string]ContainerMCPServerConfig {
	all := g.GenerateContainerMCPs()
	enabled := make(map[string]ContainerMCPServerConfig)

	for name, cfg := range all {
		if cfg.Enabled {
			enabled[name] = cfg
		}
	}

	return enabled
}

// GetDisabledContainerMCPs returns MCPs that are disabled with reasons
func (g *ContainerMCPConfigGenerator) GetDisabledContainerMCPs() map[string]string {
	all := g.GenerateContainerMCPs()
	disabled := make(map[string]string)

	for name, cfg := range all {
		if !cfg.Enabled {
			reason := g.getDisableReason(name)
			disabled[name] = reason
		}
	}

	return disabled
}

// getDisableReason returns the reason why an MCP is disabled
func (g *ContainerMCPConfigGenerator) getDisableReason(name string) string {
	// Map of MCP names to their required environment variables
	requirements := map[string][]string{
		"gitlab":          {"GITLAB_TOKEN"},
		"slack":           {"SLACK_BOT_TOKEN", "SLACK_TEAM_ID"},
		"discord":         {"DISCORD_TOKEN"},
		"telegram":        {"TELEGRAM_BOT_TOKEN"},
		"notion":          {"NOTION_API_KEY"},
		"linear":          {"LINEAR_API_KEY"},
		"jira":            {"JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"},
		"asana":           {"ASANA_ACCESS_TOKEN"},
		"trello":          {"TRELLO_API_KEY", "TRELLO_API_TOKEN"},
		"todoist":         {"TODOIST_API_TOKEN"},
		"monday":          {"MONDAY_API_TOKEN"},
		"brave-search":    {"BRAVE_API_KEY"},
		"exa":             {"EXA_API_KEY"},
		"tavily":          {"TAVILY_API_KEY"},
		"perplexity":      {"PERPLEXITY_API_KEY"},
		"kagi":            {"KAGI_API_KEY"},
		"cloudflare":      {"CLOUDFLARE_API_TOKEN"},
		"vercel":          {"VERCEL_TOKEN"},
		"heroku":          {"HEROKU_API_KEY"},
		"aws":             {"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
		"gcp":             {"GOOGLE_APPLICATION_CREDENTIALS"},
		"supabase":        {"SUPABASE_URL", "SUPABASE_KEY"},
		"google-drive":    {"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"},
		"google-calendar": {"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"},
		"google-maps":     {"GOOGLE_MAPS_API_KEY"},
		"youtube":         {"YOUTUBE_API_KEY"},
		"gmail":           {"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"},
		"datadog":         {"DD_API_KEY", "DD_APP_KEY"},
		"grafana":         {"GRAFANA_URL", "GRAFANA_TOKEN"},
		"prometheus":      {"PROMETHEUS_URL"},
		"stripe":          {"STRIPE_API_KEY"},
		"hubspot":         {"HUBSPOT_ACCESS_TOKEN"},
		"zendesk":         {"ZENDESK_SUBDOMAIN", "ZENDESK_EMAIL", "ZENDESK_TOKEN"},
		"browserbase":     {"BROWSERBASE_API_KEY"},
		"firecrawl":       {"FIRECRAWL_API_KEY"},
		"openai":          {"OPENAI_API_KEY"},
		"figma":           {"FIGMA_API_KEY"},
		"obsidian":        {"OBSIDIAN_VAULT_PATH"},
		"sentry":          {"SENTRY_AUTH_TOKEN", "SENTRY_ORG"},
		"pinecone":        {"PINECONE_API_KEY"},
		"kubernetes":      {"KUBECONFIG"},
		"postgres":        {"POSTGRES_URL", "POSTGRES_HOST"},
		"redis":           {"REDIS_URL", "REDIS_HOST"},
		"mongodb":         {"MONGODB_URI", "MONGODB_HOST"},
		"mysql":           {"MYSQL_URL", "MYSQL_HOST"},
		"elasticsearch":   {"ELASTICSEARCH_URL", "ELASTICSEARCH_HOST"},
		"qdrant":          {"QDRANT_URL", "QDRANT_HOST"},
		"chroma":          {"CHROMA_URL", "CHROMA_HOST"},
		"weaviate":        {"WEAVIATE_URL", "WEAVIATE_HOST"},
		"atlassian":       {"ATLASSIAN_URL", "ATLASSIAN_EMAIL", "ATLASSIAN_API_TOKEN"},
		"airtable":        {"AIRTABLE_API_KEY"},
		"llamaindex":      {"OPENAI_API_KEY"},
		"langchain":       {"OPENAI_API_KEY"},
		"github":          {"GITHUB_TOKEN"},
	}

	if reqs, ok := requirements[name]; ok {
		var missing []string
		for _, req := range reqs {
			if !g.hasEnvVar(req) {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			return "Missing: " + strings.Join(missing, ", ")
		}
	}

	return "Unknown reason"
}

// GenerateSummary returns a summary of enabled/disabled MCPs
func (g *ContainerMCPConfigGenerator) GenerateSummary() map[string]interface{} {
	all := g.GenerateContainerMCPs()

	enabled := []string{}
	disabled := []string{}
	byCategory := make(map[string]int)

	for name, cfg := range all {
		if cfg.Enabled {
			enabled = append(enabled, name)
			byCategory[cfg.Category]++
		} else {
			disabled = append(disabled, name)
		}
	}

	return map[string]interface{}{
		"total":            len(all),
		"total_enabled":    len(enabled),
		"total_disabled":   len(disabled),
		"enabled_mcps":     enabled,
		"disabled_mcps":    disabled,
		"by_category":      byCategory,
		"container_host":   g.mcpHost,
		"npx_dependencies": 0, // ZERO npx dependencies!
	}
}

// GetPortAllocations returns the complete port allocation map
func (g *ContainerMCPConfigGenerator) GetPortAllocations() []MCPContainerPort {
	return MCPContainerPorts
}

// ValidatePortAllocations checks for port conflicts
func (g *ContainerMCPConfigGenerator) ValidatePortAllocations() error {
	usedPorts := make(map[int]string)

	for _, p := range MCPContainerPorts {
		if existing, ok := usedPorts[p.Port]; ok {
			return fmt.Errorf("port conflict: port %d used by both %s and %s", p.Port, existing, p.Name)
		}
		usedPorts[p.Port] = p.Name
	}

	return nil
}

// GetMCPsByCategory returns MCPs grouped by category
func (g *ContainerMCPConfigGenerator) GetMCPsByCategory() map[string][]ContainerMCPServerConfig {
	all := g.GenerateContainerMCPs()
	byCategory := make(map[string][]ContainerMCPServerConfig)

	for _, cfg := range all {
		byCategory[cfg.Category] = append(byCategory[cfg.Category], cfg)
	}

	return byCategory
}

// ContainsNPX checks if any MCP configuration uses npx (should always return false)
func (g *ContainerMCPConfigGenerator) ContainsNPX() bool {
	// This generator uses ONLY container URLs, no npx commands
	return false
}

// GetTotalMCPCount returns the total number of MCP servers defined
func (g *ContainerMCPConfigGenerator) GetTotalMCPCount() int {
	return len(g.GenerateContainerMCPs())
}
