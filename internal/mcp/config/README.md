# MCP Configuration Generator

This package provides configuration generators for MCP (Model Context Protocol) servers, supporting both traditional npx-based execution and containerized deployments. It generates validated configurations for CLI agents like Claude Code, OpenCode, and Crush.

## Overview

The MCP config generator solves the challenge of:
- **Dynamic Configuration**: Generate MCP configs based on available environment variables
- **Dependency Validation**: Only enable MCPs when all requirements are met
- **Port Management**: Organized port allocation scheme for containerized MCPs
- **Multi-Format Output**: Support for various CLI agent configuration formats

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MCP Configuration Generator                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐     ┌─────────────────────┐                        │
│  │  MCPConfigGenerator │     │ ContainerMCPConfig  │                        │
│  │   (npx-based)       │     │    Generator        │                        │
│  └──────────┬──────────┘     └──────────┬──────────┘                        │
│             │                            │                                   │
│             │    ┌───────────────────────┤                                   │
│             │    │                       │                                   │
│             ▼    ▼                       ▼                                   │
│  ┌─────────────────────┐     ┌─────────────────────┐                        │
│  │ Environment Loader  │     │   Port Allocator    │                        │
│  │ (.env, env vars)    │     │   (9101-9999)       │                        │
│  └──────────┬──────────┘     └──────────┬──────────┘                        │
│             │                            │                                   │
│             └────────────┬───────────────┘                                   │
│                          │                                                   │
│                          ▼                                                   │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Configuration Output                                  ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    ││
│  │  │   Claude    │  │  OpenCode   │  │   Crush     │  │   Docker    │    ││
│  │  │   Config    │  │   Config    │  │   Config    │  │   Compose   │    ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
```

## Docker Compose Generation

### Port Management Scheme

The containerized MCP generator uses a structured port allocation:

```
TIER 1: Core Official MCP Servers (9101-9120)
├── fetch:              9101
├── git:                9102
├── time:               9103
├── filesystem:         9104
├── memory:             9105
├── everything:         9106
├── sequential-thinking: 9107
├── sqlite:             9108
├── puppeteer:          9109
└── postgres:           9110

TIER 2: Database MCP Servers (9201-9220)
├── mongodb:            9201
├── redis:              9202
├── mysql:              9203
├── elasticsearch:      9204
└── supabase:           9205

TIER 3: Vector Database MCP Servers (9301-9320)
├── qdrant:             9301
├── chroma:             9302
├── pinecone:           9303
└── weaviate:           9304

TIER 4: DevOps & Infrastructure (9401-9440)
├── github:             9401
├── gitlab:             9402
├── sentry:             9403
├── kubernetes:         9404
├── docker:             9405
├── ansible:            9406
├── aws:                9407
├── gcp:                9408
├── heroku:             9409
├── cloudflare:         9410
├── vercel:             9411
├── workers:            9412
└── jetbrains:          9413

TIER 5: Browser & Web Automation (9501-9520)
├── playwright:         9501
├── browserbase:        9502
├── firecrawl:          9503
└── crawl4ai:           9504

TIER 6: Communication (9601-9620)
├── slack:              9601
├── discord:            9602
└── telegram:           9603

TIER 7: Productivity (9701-9740)
├── notion:             9701
├── linear:             9702
├── jira:               9703
├── asana:              9704
├── trello:             9705
├── todoist:            9706
├── monday:             9707
├── airtable:           9708
├── obsidian:           9709
└── atlassian:          9710

TIER 8: Search & AI (9801-9840)
├── brave-search:       9801
├── exa:                9802
├── tavily:             9803
├── perplexity:         9804
├── kagi:               9805
├── omnisearch:         9806
├── context7:           9807
├── llamaindex:         9808
├── langchain:          9809
└── openai:             9810

TIER 9: Google Services (9901-9920)
├── google-drive:       9901
├── google-calendar:    9902
├── google-maps:        9903
├── youtube:            9904
└── gmail:              9905

TIER 10: Monitoring (9921-9940)
├── datadog:            9921
├── grafana:            9922
└── prometheus:         9923

TIER 11: Finance & Business (9941-9960)
├── stripe:             9941
├── hubspot:            9942
└── zendesk:            9943

TIER 12: Design (9961-9970)
└── figma:              9961
```

### Generating Docker Compose

```go
import "dev.helix.agent/internal/mcp/config"

gen := config.NewContainerMCPConfigGenerator("http://localhost:7061")

// Get all enabled MCPs
mcps := gen.GetEnabledContainerMCPs()

// Generate docker-compose.yml content
for name, cfg := range mcps {
    fmt.Printf("  %s:\n", name)
    fmt.Printf("    image: mcp-server-%s:latest\n", name)
    fmt.Printf("    ports:\n")
    fmt.Printf("      - \"%d:8080\"\n", cfg.Port)
    fmt.Printf("    environment:\n")
    for k, v := range cfg.Environment {
        fmt.Printf("      %s: %s\n", k, v)
    }
}
```

### Port Validation

```go
gen := config.NewContainerMCPConfigGenerator("http://localhost:7061")

// Validate no port conflicts
if err := gen.ValidatePortAllocations(); err != nil {
    log.Fatalf("Port conflict detected: %v", err)
}

// Get complete port map
ports := gen.GetPortAllocations()
for _, p := range ports {
    fmt.Printf("%s: %d (%s)\n", p.Name, p.Port, p.Category)
}
```

## Environment Variables

### Loading Environment Variables

The generator loads environment variables from multiple sources:

1. `.env` file (lowest priority)
2. `.env.local` file
3. `.env.mcp` file
4. `.env.mcp.generated` file
5. System environment (highest priority)

### Required Variables by MCP

| MCP | Required Variables |
|-----|-------------------|
| github | `GITHUB_TOKEN` |
| gitlab | `GITLAB_TOKEN` |
| slack | `SLACK_BOT_TOKEN`, `SLACK_TEAM_ID` |
| notion | `NOTION_API_KEY` |
| brave-search | `BRAVE_API_KEY` |
| sentry | `SENTRY_AUTH_TOKEN`, `SENTRY_ORG` |
| kubernetes | `KUBECONFIG` |
| aws | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| google-drive | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` |
| pinecone | `PINECONE_API_KEY` |
| postgres | `POSTGRES_URL` or `POSTGRES_HOST` |
| redis | `REDIS_URL` or `REDIS_HOST` |
| mongodb | `MONGODB_URI` or `MONGODB_HOST` |

### Example .env File

```bash
# HelixAgent MCP Configuration
HELIXAGENT_HOME=/home/user/.helixagent

# Core services
POSTGRES_URL=postgresql://user:pass@localhost:5432/db
REDIS_URL=redis://localhost:6379

# Development tools
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
GITLAB_TOKEN=glpat-xxxxxxxxxxxx

# Cloud services
AWS_ACCESS_KEY_ID=AKIAXXXXXXXX
AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxx
PINECONE_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

# Productivity
NOTION_API_KEY=secret_xxxxxxxxxxxx
LINEAR_API_KEY=lin_api_xxxxxxxxxxxx
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxx
SLACK_TEAM_ID=T0XXXXXXX

# Search
BRAVE_API_KEY=BSAxxxxxxxxxxxxxxxx
```

## Usage

### Traditional NPX-based Config

```go
import "dev.helix.agent/internal/mcp/config"

gen := config.NewMCPConfigGenerator("http://localhost:7061")

// Generate OpenCode MCP configurations
mcps := gen.GenerateOpenCodeMCPs()

for name, cfg := range mcps {
    if cfg.Enabled {
        fmt.Printf("%s: %v\n", name, cfg.Command)
    }
}

// Get summary
summary := gen.GetMCPSummary()
fmt.Printf("Enabled: %d MCPs\n", summary["total_enabled"])
```

### Container-based Config

```go
gen := config.NewContainerMCPConfigGenerator("http://localhost:7061")

// Get all MCPs (enabled and disabled)
allMCPs := gen.GenerateContainerMCPs()

// Get only enabled MCPs
enabledMCPs := gen.GetEnabledContainerMCPs()

// Get disabled MCPs with reasons
disabledMCPs := gen.GetDisabledContainerMCPs()
for name, reason := range disabledMCPs {
    fmt.Printf("%s: %s\n", name, reason)
}

// Get MCPs grouped by category
byCategory := gen.GetMCPsByCategory()
for category, mcps := range byCategory {
    fmt.Printf("%s: %d MCPs\n", category, len(mcps))
}
```

### Configuration Output

```go
// MCPServerConfig (npx-based)
type MCPServerConfig struct {
    Type        string            // "local" for npx
    Command     []string          // ["npx", "-y", "@modelcontextprotocol/server-fetch"]
    URL         string            // For remote MCPs
    Headers     map[string]string // HTTP headers
    Environment map[string]string // Environment variables
    Enabled     bool              // Whether MCP is enabled
}

// ContainerMCPServerConfig
type ContainerMCPServerConfig struct {
    Type        string            // "remote"
    URL         string            // "http://localhost:9101/sse"
    Headers     map[string]string // HTTP headers
    Environment map[string]string // Environment variables
    Env         map[string]string // Crush uses "env" instead
    Enabled     bool              // Whether MCP is enabled
    Port        int               // Container port
    Category    string            // MCP category
}
```

## Integration with CLI Agents

### Claude Code Integration

```go
// Generate claude_desktop_config.json
gen := config.NewMCPConfigGenerator("http://localhost:7061")
mcps := gen.GenerateOpenCodeMCPs()

claudeConfig := map[string]interface{}{
    "mcpServers": mcps,
}

data, _ := json.MarshalIndent(claudeConfig, "", "  ")
os.WriteFile("claude_desktop_config.json", data, 0644)
```

### OpenCode Integration

```go
// Generate .opencode/config.json
gen := config.NewContainerMCPConfigGenerator("http://localhost:7061")
mcps := gen.GetEnabledContainerMCPs()

openCodeConfig := map[string]interface{}{
    "mcp": mcps,
}

data, _ := json.MarshalIndent(openCodeConfig, "", "  ")
os.WriteFile(".opencode/config.json", data, 0644)
```

## Summary Generation

```go
gen := config.NewContainerMCPConfigGenerator("http://localhost:7061")
summary := gen.GenerateSummary()

fmt.Printf("Total MCPs: %d\n", summary["total"])
fmt.Printf("Enabled: %d\n", summary["total_enabled"])
fmt.Printf("Disabled: %d\n", summary["total_disabled"])
fmt.Printf("Container Host: %s\n", summary["container_host"])
fmt.Printf("NPX Dependencies: %d\n", summary["npx_dependencies"]) // Always 0 for container

byCategory := summary["by_category"].(map[string]int)
for category, count := range byCategory {
    fmt.Printf("  %s: %d\n", category, count)
}
```

## Testing

```bash
# Unit tests
go test -v ./internal/mcp/config/...

# Test with environment
GITHUB_TOKEN=test POSTGRES_URL=test \
go test -v ./internal/mcp/config/...

# Container config tests
go test -v ./internal/mcp/config/... -run Container
```

## Related Files

- `generator.go` - NPX-based MCP config generator
- `generator_container.go` - Container-based MCP config generator
- `generator_full.go` - Extended configuration options
- `generator_container_test.go` - Container config tests
