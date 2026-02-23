// Package config provides comprehensive MCP configuration generation for CLI agents
package config

import (
	"fmt"
	"os"
	"strings"
)

// MCPServerConfigFull represents a full MCP server configuration with all fields
type MCPServerConfigFull struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Env         map[string]string `json:"env,omitempty"` // Crush uses "env" instead of "environment"
	Enabled     bool              `json:"enabled"`
}

// MCPCategory represents a category of MCPs
type MCPCategory struct {
	Name        string
	Description string
	MCPs        []string
}

// FullMCPConfigGenerator generates comprehensive MCP configurations
type FullMCPConfigGenerator struct {
	homeDir   string
	helixHome string
	baseURL   string
	envVars   map[string]string
}

// NewFullMCPConfigGenerator creates a new full MCP config generator
func NewFullMCPConfigGenerator(baseURL string) *FullMCPConfigGenerator {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}

	helixHome := os.Getenv("HELIXAGENT_HOME")
	if helixHome == "" {
		helixHome = homeDir + "/.helixagent"
	}

	g := &FullMCPConfigGenerator{
		homeDir:   homeDir,
		helixHome: helixHome,
		baseURL:   baseURL,
		envVars:   make(map[string]string),
	}

	g.loadEnvVars()
	return g
}

// loadEnvVars loads environment variables from .env files and environment
func (g *FullMCPConfigGenerator) loadEnvVars() {
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
func (g *FullMCPConfigGenerator) hasEnvVar(name string) bool {
	val, ok := g.envVars[name]
	return ok && val != ""
}

// hasAnyEnvVar checks if any of the given environment variables is set
func (g *FullMCPConfigGenerator) hasAnyEnvVar(names ...string) bool {
	for _, name := range names {
		if g.hasEnvVar(name) {
			return true
		}
	}
	return false
}

// hasAllEnvVars checks if all of the given environment variables are set
func (g *FullMCPConfigGenerator) hasAllEnvVars(names ...string) bool {
	for _, name := range names {
		if !g.hasEnvVar(name) {
			return false
		}
	}
	return true
}

// GenerateAllMCPs generates ALL MCP configurations, marking which are enabled
func (g *FullMCPConfigGenerator) GenerateAllMCPs() map[string]MCPServerConfigFull {
	mcps := make(map[string]MCPServerConfigFull)

	// ==========================================================================
	// CATEGORY 1: HelixAgent Core (Always enabled)
	// ==========================================================================
	mcps["helixagent"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"node", g.helixHome + "/plugins/mcp-server/dist/index.js", "--endpoint", g.baseURL},
		Enabled: true,
	}

	// ==========================================================================
	// CATEGORY 2: Anthropic Official MCPs (No API keys, always work)
	// ==========================================================================
	mcps["filesystem"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", g.homeDir},
		Enabled: true,
	}
	mcps["fetch"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-fetch-server"},
		Enabled: true,
	}
	mcps["memory"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-memory"},
		Enabled: true,
	}
	mcps["time"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@theo.foobar/mcp-time"},
		Enabled: true,
	}
	mcps["git"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-git"},
		Enabled: true,
	}
	mcps["sequential-thinking"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-sequential-thinking"},
		Enabled: true,
	}
	mcps["everything"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@anthropic-ai/mcp-server-everything"},
		Enabled: true,
	}
	mcps["sqlite"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-sqlite", "--db-path", "/tmp/helixagent.db"},
		Enabled: true,
	}
	mcps["puppeteer"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-puppeteer"},
		Enabled: true,
	}

	// ==========================================================================
	// CATEGORY 3: Database MCPs (Local services - enabled if env vars set)
	// ==========================================================================
	mcps["postgres"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-postgres"},
		Environment: map[string]string{
			"POSTGRES_URL": g.getEnvOrDefault("POSTGRES_URL", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
				g.getEnvOrDefault("POSTGRES_USER", "helixagent"),
				os.Getenv("POSTGRES_PASSWORD"),
				g.getEnvOrDefault("POSTGRES_HOST", "localhost"),
				g.getEnvOrDefault("POSTGRES_PORT", "15432"),
				g.getEnvOrDefault("POSTGRES_DB", "helixagent_db"),
			)),
		},
		Enabled: g.hasAnyEnvVar("POSTGRES_URL", "POSTGRES_HOST"),
	}
	mcps["redis"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-redis"},
		Environment: map[string]string{
			"REDIS_URL": g.getEnvOrDefault("REDIS_URL", fmt.Sprintf("redis://:%s@%s:%s",
				os.Getenv("REDIS_PASSWORD"),
				g.getEnvOrDefault("REDIS_HOST", "localhost"),
				g.getEnvOrDefault("REDIS_PORT", "16379"),
			)),
		},
		Enabled: g.hasAnyEnvVar("REDIS_URL", "REDIS_HOST"),
	}
	mcps["mongodb"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-mongodb"},
		Environment: map[string]string{
			"MONGODB_URI": g.getEnvOrDefault("MONGODB_URI", fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
				g.getEnvOrDefault("MONGODB_USER", "helixagent"),
				os.Getenv("MONGODB_PASSWORD"),
				g.getEnvOrDefault("MONGODB_HOST", "localhost"),
				g.getEnvOrDefault("MONGODB_PORT", "27017"),
				g.getEnvOrDefault("MONGODB_DB", "helixagent"),
			)),
		},
		Enabled: g.hasAnyEnvVar("MONGODB_URI", "MONGODB_HOST"),
	}
	mcps["mysql"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-mysql"},
		Environment: map[string]string{
			"MYSQL_URL": g.getEnvOrDefault("MYSQL_URL", fmt.Sprintf("mysql://%s:%s@%s:%s/%s",
				g.getEnvOrDefault("MYSQL_USER", "helixagent"),
				os.Getenv("MYSQL_PASSWORD"),
				g.getEnvOrDefault("MYSQL_HOST", "localhost"),
				g.getEnvOrDefault("MYSQL_PORT", "3306"),
				g.getEnvOrDefault("MYSQL_DB", "helixagent"),
			)),
		},
		Enabled: g.hasAnyEnvVar("MYSQL_URL", "MYSQL_HOST"),
	}
	mcps["elasticsearch"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-elasticsearch"},
		Environment: map[string]string{
			"ELASTICSEARCH_URL": g.getEnvOrDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
		},
		Enabled: g.hasAnyEnvVar("ELASTICSEARCH_URL", "ELASTICSEARCH_HOST"),
	}

	// ==========================================================================
	// CATEGORY 4: Vector Databases (Local services)
	// ==========================================================================
	mcps["qdrant"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-qdrant"},
		Environment: map[string]string{
			"QDRANT_URL": g.getEnvOrDefault("QDRANT_URL", "http://localhost:6333"),
		},
		Enabled: g.hasAnyEnvVar("QDRANT_URL", "QDRANT_HOST"),
	}
	mcps["chroma"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-chroma"},
		Environment: map[string]string{
			"CHROMA_URL": g.getEnvOrDefault("CHROMA_URL", "http://localhost:8000"),
		},
		Enabled: g.hasAnyEnvVar("CHROMA_URL", "CHROMA_HOST"),
	}
	mcps["pinecone"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-pinecone"},
		Environment: map[string]string{
			"PINECONE_API_KEY": "{env:PINECONE_API_KEY}",
		},
		Enabled: g.hasEnvVar("PINECONE_API_KEY"),
	}
	mcps["weaviate"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-weaviate"},
		Environment: map[string]string{
			"WEAVIATE_URL": g.getEnvOrDefault("WEAVIATE_URL", "http://localhost:8080"),
		},
		Enabled: g.hasAnyEnvVar("WEAVIATE_URL", "WEAVIATE_HOST"),
	}

	// ==========================================================================
	// CATEGORY 5: DevOps & Infrastructure
	// ==========================================================================
	mcps["docker"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-docker"},
		Enabled: true, // Always enabled - uses local docker/podman
	}
	mcps["kubernetes"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-kubernetes"},
		Environment: map[string]string{
			"KUBECONFIG": g.getEnvOrDefault("KUBECONFIG", g.homeDir+"/.kube/config"),
		},
		Enabled: g.hasEnvVar("KUBECONFIG"),
	}
	mcps["ansible"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-ansible"},
		Enabled: true, // Works with local ansible
	}

	// ==========================================================================
	// CATEGORY 6: Development Platforms
	// ==========================================================================
	mcps["github"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-github"},
		Environment: map[string]string{
			"GITHUB_PERSONAL_ACCESS_TOKEN": "{env:GITHUB_TOKEN}",
		},
		Enabled: g.hasEnvVar("GITHUB_TOKEN"),
	}
	mcps["gitlab"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-gitlab"},
		Environment: map[string]string{
			"GITLAB_PERSONAL_ACCESS_TOKEN": "{env:GITLAB_TOKEN}",
		},
		Enabled: g.hasEnvVar("GITLAB_TOKEN"),
	}
	mcps["sentry"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-sentry"},
		Environment: map[string]string{
			"SENTRY_AUTH_TOKEN": "{env:SENTRY_AUTH_TOKEN}",
			"SENTRY_ORG":        "{env:SENTRY_ORG}",
		},
		Enabled: g.hasAllEnvVars("SENTRY_AUTH_TOKEN", "SENTRY_ORG"),
	}
	mcps["jetbrains"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-jetbrains"},
		Enabled: true, // Works locally
	}

	// ==========================================================================
	// CATEGORY 7: Communication & Collaboration
	// ==========================================================================
	mcps["slack"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-slack"},
		Environment: map[string]string{
			"SLACK_BOT_TOKEN": "{env:SLACK_BOT_TOKEN}",
			"SLACK_TEAM_ID":   "{env:SLACK_TEAM_ID}",
		},
		Enabled: g.hasAllEnvVars("SLACK_BOT_TOKEN", "SLACK_TEAM_ID"),
	}
	mcps["discord"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-discord"},
		Environment: map[string]string{
			"DISCORD_TOKEN": "{env:DISCORD_TOKEN}",
		},
		Enabled: g.hasEnvVar("DISCORD_TOKEN"),
	}
	mcps["telegram"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-telegram"},
		Environment: map[string]string{
			"TELEGRAM_BOT_TOKEN": "{env:TELEGRAM_BOT_TOKEN}",
		},
		Enabled: g.hasEnvVar("TELEGRAM_BOT_TOKEN"),
	}

	// ==========================================================================
	// CATEGORY 8: Productivity & Project Management
	// ==========================================================================
	mcps["notion"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@notionhq/notion-mcp-server"},
		Environment: map[string]string{
			"NOTION_API_KEY": "{env:NOTION_API_KEY}",
		},
		Enabled: g.hasEnvVar("NOTION_API_KEY"),
	}
	mcps["linear"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-linear"},
		Environment: map[string]string{
			"LINEAR_API_KEY": "{env:LINEAR_API_KEY}",
		},
		Enabled: g.hasEnvVar("LINEAR_API_KEY"),
	}
	mcps["jira"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-jira"},
		Environment: map[string]string{
			"JIRA_URL":       "{env:JIRA_URL}",
			"JIRA_EMAIL":     "{env:JIRA_EMAIL}",
			"JIRA_API_TOKEN": "{env:JIRA_API_TOKEN}",
		},
		Enabled: g.hasAllEnvVars("JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"),
	}
	mcps["asana"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-asana"},
		Environment: map[string]string{
			"ASANA_ACCESS_TOKEN": "{env:ASANA_ACCESS_TOKEN}",
		},
		Enabled: g.hasEnvVar("ASANA_ACCESS_TOKEN"),
	}
	mcps["trello"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-trello"},
		Environment: map[string]string{
			"TRELLO_API_KEY":   "{env:TRELLO_API_KEY}",
			"TRELLO_API_TOKEN": "{env:TRELLO_API_TOKEN}",
		},
		Enabled: g.hasAllEnvVars("TRELLO_API_KEY", "TRELLO_API_TOKEN"),
	}
	mcps["todoist"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-todoist"},
		Environment: map[string]string{
			"TODOIST_API_TOKEN": "{env:TODOIST_API_TOKEN}",
		},
		Enabled: g.hasEnvVar("TODOIST_API_TOKEN"),
	}
	mcps["monday"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-monday"},
		Environment: map[string]string{
			"MONDAY_API_TOKEN": "{env:MONDAY_API_TOKEN}",
		},
		Enabled: g.hasEnvVar("MONDAY_API_TOKEN"),
	}

	// ==========================================================================
	// CATEGORY 9: Search & AI
	// ==========================================================================
	mcps["brave-search"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-brave-search"},
		Environment: map[string]string{
			"BRAVE_API_KEY": "{env:BRAVE_API_KEY}",
		},
		Enabled: g.hasEnvVar("BRAVE_API_KEY"),
	}
	mcps["exa"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "exa-mcp-server"},
		Environment: map[string]string{
			"EXA_API_KEY": "{env:EXA_API_KEY}",
		},
		Enabled: g.hasEnvVar("EXA_API_KEY"),
	}
	mcps["tavily"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-tavily"},
		Environment: map[string]string{
			"TAVILY_API_KEY": "{env:TAVILY_API_KEY}",
		},
		Enabled: g.hasEnvVar("TAVILY_API_KEY"),
	}
	mcps["perplexity"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-perplexity"},
		Environment: map[string]string{
			"PERPLEXITY_API_KEY": "{env:PERPLEXITY_API_KEY}",
		},
		Enabled: g.hasEnvVar("PERPLEXITY_API_KEY"),
	}
	mcps["kagi"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-kagi"},
		Environment: map[string]string{
			"KAGI_API_KEY": "{env:KAGI_API_KEY}",
		},
		Enabled: g.hasEnvVar("KAGI_API_KEY"),
	}

	// ==========================================================================
	// CATEGORY 10: Cloud Providers
	// ==========================================================================
	mcps["cloudflare"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@cloudflare/mcp-server-cloudflare"},
		Environment: map[string]string{
			"CLOUDFLARE_API_TOKEN": "{env:CLOUDFLARE_API_TOKEN}",
		},
		Enabled: g.hasEnvVar("CLOUDFLARE_API_TOKEN"),
	}
	mcps["vercel"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-vercel"},
		Environment: map[string]string{
			"VERCEL_TOKEN": "{env:VERCEL_TOKEN}",
		},
		Enabled: g.hasEnvVar("VERCEL_TOKEN"),
	}
	mcps["heroku"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-heroku"},
		Environment: map[string]string{
			"HEROKU_API_KEY": "{env:HEROKU_API_KEY}",
		},
		Enabled: g.hasEnvVar("HEROKU_API_KEY"),
	}
	mcps["aws"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-aws"},
		Environment: map[string]string{
			"AWS_ACCESS_KEY_ID":     "{env:AWS_ACCESS_KEY_ID}",
			"AWS_SECRET_ACCESS_KEY": "{env:AWS_SECRET_ACCESS_KEY}",
			"AWS_REGION":            g.getEnvOrDefault("AWS_REGION", "us-east-1"),
		},
		Enabled: g.hasAllEnvVars("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"),
	}
	mcps["gcp"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-gcp"},
		Environment: map[string]string{
			"GOOGLE_APPLICATION_CREDENTIALS": "{env:GOOGLE_APPLICATION_CREDENTIALS}",
		},
		Enabled: g.hasEnvVar("GOOGLE_APPLICATION_CREDENTIALS"),
	}
	mcps["supabase"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-supabase"},
		Environment: map[string]string{
			"SUPABASE_URL": "{env:SUPABASE_URL}",
			"SUPABASE_KEY": "{env:SUPABASE_KEY}",
		},
		Enabled: g.hasAllEnvVars("SUPABASE_URL", "SUPABASE_KEY"),
	}

	// ==========================================================================
	// CATEGORY 11: Google Services
	// ==========================================================================
	mcps["google-drive"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-gdrive"},
		Environment: map[string]string{
			"GOOGLE_CLIENT_ID":     "{env:GOOGLE_CLIENT_ID}",
			"GOOGLE_CLIENT_SECRET": "{env:GOOGLE_CLIENT_SECRET}",
		},
		Enabled: g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
	}
	mcps["google-calendar"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-google-calendar"},
		Environment: map[string]string{
			"GOOGLE_CLIENT_ID":     "{env:GOOGLE_CLIENT_ID}",
			"GOOGLE_CLIENT_SECRET": "{env:GOOGLE_CLIENT_SECRET}",
		},
		Enabled: g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
	}
	mcps["google-maps"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-google-maps"},
		Environment: map[string]string{
			"GOOGLE_MAPS_API_KEY": "{env:GOOGLE_MAPS_API_KEY}",
		},
		Enabled: g.hasEnvVar("GOOGLE_MAPS_API_KEY"),
	}
	mcps["youtube"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-youtube"},
		Environment: map[string]string{
			"YOUTUBE_API_KEY": "{env:YOUTUBE_API_KEY}",
		},
		Enabled: g.hasEnvVar("YOUTUBE_API_KEY"),
	}
	mcps["gmail"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-gmail"},
		Environment: map[string]string{
			"GOOGLE_CLIENT_ID":     "{env:GOOGLE_CLIENT_ID}",
			"GOOGLE_CLIENT_SECRET": "{env:GOOGLE_CLIENT_SECRET}",
		},
		Enabled: g.hasAllEnvVars("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"),
	}

	// ==========================================================================
	// CATEGORY 12: Monitoring & Observability
	// ==========================================================================
	mcps["datadog"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-datadog"},
		Environment: map[string]string{
			"DD_API_KEY": "{env:DD_API_KEY}",
			"DD_APP_KEY": "{env:DD_APP_KEY}",
		},
		Enabled: g.hasAllEnvVars("DD_API_KEY", "DD_APP_KEY"),
	}
	mcps["grafana"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-grafana"},
		Environment: map[string]string{
			"GRAFANA_URL":   "{env:GRAFANA_URL}",
			"GRAFANA_TOKEN": "{env:GRAFANA_TOKEN}",
		},
		Enabled: g.hasAllEnvVars("GRAFANA_URL", "GRAFANA_TOKEN"),
	}
	mcps["prometheus"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-prometheus"},
		Environment: map[string]string{
			"PROMETHEUS_URL": g.getEnvOrDefault("PROMETHEUS_URL", "http://localhost:9090"),
		},
		Enabled: g.hasEnvVar("PROMETHEUS_URL"),
	}

	// ==========================================================================
	// CATEGORY 13: Finance & Business
	// ==========================================================================
	mcps["stripe"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-stripe"},
		Environment: map[string]string{
			"STRIPE_API_KEY": "{env:STRIPE_API_KEY}",
		},
		Enabled: g.hasEnvVar("STRIPE_API_KEY"),
	}
	mcps["hubspot"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-hubspot"},
		Environment: map[string]string{
			"HUBSPOT_ACCESS_TOKEN": "{env:HUBSPOT_ACCESS_TOKEN}",
		},
		Enabled: g.hasEnvVar("HUBSPOT_ACCESS_TOKEN"),
	}
	mcps["zendesk"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-zendesk"},
		Environment: map[string]string{
			"ZENDESK_SUBDOMAIN": "{env:ZENDESK_SUBDOMAIN}",
			"ZENDESK_EMAIL":     "{env:ZENDESK_EMAIL}",
			"ZENDESK_TOKEN":     "{env:ZENDESK_TOKEN}",
		},
		Enabled: g.hasAllEnvVars("ZENDESK_SUBDOMAIN", "ZENDESK_EMAIL", "ZENDESK_TOKEN"),
	}

	// ==========================================================================
	// CATEGORY 14: Browser & Web
	// ==========================================================================
	mcps["browserbase"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-browserbase"},
		Environment: map[string]string{
			"BROWSERBASE_API_KEY": "{env:BROWSERBASE_API_KEY}",
		},
		Enabled: g.hasEnvVar("BROWSERBASE_API_KEY"),
	}
	mcps["playwright"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-playwright"},
		Enabled: true, // Works locally
	}
	mcps["crawl4ai"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-crawl4ai"},
		Enabled: true, // Works locally
	}

	// ==========================================================================
	// CATEGORY 15: AI & OpenAI
	// ==========================================================================
	mcps["openai"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-openai"},
		Environment: map[string]string{
			"OPENAI_API_KEY": "{env:OPENAI_API_KEY}",
		},
		Enabled: g.hasEnvVar("OPENAI_API_KEY"),
	}

	// ==========================================================================
	// CATEGORY 16: Design
	// ==========================================================================
	mcps["figma"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "figma-developer-mcp"},
		Environment: map[string]string{
			"FIGMA_API_KEY": "{env:FIGMA_API_KEY}",
		},
		Enabled: g.hasEnvVar("FIGMA_API_KEY"),
	}

	// ==========================================================================
	// CATEGORY 17: Notes & Knowledge
	// ==========================================================================
	mcps["obsidian"] = MCPServerConfigFull{
		Type:    "local",
		Command: []string{"npx", "-y", "mcp-server-obsidian"},
		Environment: map[string]string{
			"OBSIDIAN_VAULT_PATH": g.getEnvOrDefault("OBSIDIAN_VAULT_PATH", g.homeDir+"/Documents/Obsidian"),
		},
		Enabled: g.hasEnvVar("OBSIDIAN_VAULT_PATH"),
	}

	return mcps
}

// GetEnabledMCPs returns only the MCPs that are enabled
func (g *FullMCPConfigGenerator) GetEnabledMCPs() map[string]MCPServerConfigFull {
	all := g.GenerateAllMCPs()
	enabled := make(map[string]MCPServerConfigFull)

	for name, cfg := range all {
		if cfg.Enabled {
			enabled[name] = cfg
		}
	}

	return enabled
}

// GetDisabledMCPs returns MCPs that are disabled with reasons
func (g *FullMCPConfigGenerator) GetDisabledMCPs() map[string]string {
	all := g.GenerateAllMCPs()
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
func (g *FullMCPConfigGenerator) getDisableReason(name string) string {
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
		"openai":          {"OPENAI_API_KEY"},
		"figma":           {"FIGMA_API_KEY"},
		"obsidian":        {"OBSIDIAN_VAULT_PATH"},
		"sentry":          {"SENTRY_AUTH_TOKEN", "SENTRY_ORG"},
		"pinecone":        {"PINECONE_API_KEY"},
		"kubernetes":      {"KUBECONFIG"},
		"postgres":        {"POSTGRES_URL"},
		"redis":           {"REDIS_URL"},
		"mongodb":         {"MONGODB_URI"},
		"mysql":           {"MYSQL_URL"},
		"elasticsearch":   {"ELASTICSEARCH_URL"},
		"qdrant":          {"QDRANT_URL"},
		"chroma":          {"CHROMA_URL"},
		"weaviate":        {"WEAVIATE_URL"},
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

// getEnvOrDefault returns the environment variable value or default
func (g *FullMCPConfigGenerator) getEnvOrDefault(name, defaultVal string) string {
	if val, ok := g.envVars[name]; ok && val != "" {
		return val
	}
	return defaultVal
}

// GenerateSummary returns a summary of enabled/disabled MCPs
func (g *FullMCPConfigGenerator) GenerateSummary() map[string]interface{} {
	all := g.GenerateAllMCPs()

	enabled := []string{}
	disabled := []string{}

	for name, cfg := range all {
		if cfg.Enabled {
			enabled = append(enabled, name)
		} else {
			disabled = append(disabled, name)
		}
	}

	return map[string]interface{}{
		"total":          len(all),
		"total_enabled":  len(enabled),
		"total_disabled": len(disabled),
		"enabled_mcps":   enabled,
		"disabled_mcps":  disabled,
	}
}
