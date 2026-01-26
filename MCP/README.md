# HelixAgent MCP Servers

This directory contains **48 MCP (Model Context Protocol) server submodules** that can be used with HelixAgent CLI agents (OpenCode, Crush, etc.) and other MCP-compatible tools.

## Quick Start

### Using Docker Compose

```bash
# Start all MCP servers
docker-compose -f MCP/docker-compose.yml up -d

# Start specific services
docker-compose -f MCP/docker-compose.yml up -d mcp-github mcp-slack mcp-notion

# View logs
docker-compose -f MCP/docker-compose.yml logs -f

# Stop all services
docker-compose -f MCP/docker-compose.yml down
```

### Using Podman Compose

```bash
# Start all MCP servers
podman-compose -f MCP/docker-compose.yml up -d

# Start specific services
podman-compose -f MCP/docker-compose.yml up -d mcp-github mcp-slack

# Stop all services
podman-compose -f MCP/docker-compose.yml down
```

## MCP Server Categories

### Core Servers (No Auth Required) - 7 Servers

| Server | Port | Description |
|--------|------|-------------|
| mcp-everything | 3001 | Reference server with all MCP features |
| mcp-filesystem | 3002 | Secure file system operations |
| mcp-memory | 3003 | Knowledge graph persistent memory |
| mcp-sequential-thinking | 3004 | Step-by-step reasoning |
| mcp-fetch | 3005 | HTTP requests and web fetching |
| mcp-git | 3006 | Git repository operations |
| mcp-time | 3007 | Time and timezone utilities |

### Browser Automation - 2 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-playwright | 3010 | (none) |
| mcp-browserbase | 3011 | BROWSERBASE_API_KEY, BROWSERBASE_PROJECT_ID |

### Databases & Storage - 5 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-redis | 3020 | REDIS_URL |
| mcp-mongodb | 3021 | MONGODB_URI |
| mcp-qdrant | 3022 | QDRANT_URL |
| mcp-elasticsearch | 3023 | ELASTICSEARCH_URL |
| mcp-supabase | 3024 | SUPABASE_URL, SUPABASE_KEY |

### Version Control & DevOps - 4 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-github | 3030 | GITHUB_TOKEN |
| mcp-sentry | 3031 | SENTRY_AUTH_TOKEN, SENTRY_ORG |
| mcp-heroku | 3032 | HEROKU_API_KEY |
| mcp-cloudflare | 3033 | CLOUDFLARE_API_TOKEN, CLOUDFLARE_ACCOUNT_ID |

### Cloud Platforms - 2 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-aws | 3040 | AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION |
| mcp-kubernetes | 3041 | ~/.kube/config (mounted) |

### Productivity & Communication - 7 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-slack | 3050 | SLACK_BOT_TOKEN, SLACK_TEAM_ID |
| mcp-telegram | 3051 | TELEGRAM_API_ID, TELEGRAM_API_HASH |
| mcp-notion | 3052 | NOTION_API_KEY |
| mcp-airtable | 3053 | AIRTABLE_API_KEY |
| mcp-trello | 3054 | TRELLO_API_KEY, TRELLO_TOKEN |
| mcp-atlassian | 3055 | JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN |
| mcp-obsidian | 3056 | OBSIDIAN_VAULT_PATH (mounted) |

### Search & AI - 5 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-brave-search | 3060 | BRAVE_API_KEY |
| mcp-perplexity | 3061 | PERPLEXITY_API_KEY |
| mcp-context7 | 3062 | CONTEXT7_API_KEY |
| mcp-firecrawl | 3063 | FIRECRAWL_API_KEY |
| mcp-omnisearch | 3064 | TAVILY_API_KEY, BRAVE_API_KEY, PERPLEXITY_API_KEY |

### AI Framework Integrations - 3 Servers

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-langchain | 3070 | OPENAI_API_KEY, ANTHROPIC_API_KEY |
| mcp-llamaindex | 3071 | OPENAI_API_KEY |
| mcp-docs | 3072 | (none) |

### Microsoft/Azure - 1 Server

| Server | Port | Required Env Vars |
|--------|------|-------------------|
| mcp-microsoft | 3080 | AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET |

## Submodule Structure

All MCP servers are organized as git submodules under `MCP/submodules/`:

```
MCP/
├── docker-compose.yml          # Main Docker Compose configuration
├── dockerfiles/               # Dockerfile templates for each server
│   ├── Dockerfile.base-node   # Base Node.js image
│   ├── Dockerfile.base-python # Base Python image
│   ├── Dockerfile.github      # GitHub MCP server
│   ├── Dockerfile.slack       # Slack MCP server
│   └── ...                    # Other Dockerfiles
├── submodules/                # Git submodules (44 repositories)
│   ├── github-mcp-server/     # GitHub's official MCP server
│   ├── slack-mcp/             # Slack integration
│   ├── notion-mcp-server/     # Notion integration
│   ├── atlassian-mcp/         # Jira/Confluence
│   ├── redis-mcp/             # Redis (official)
│   ├── mongodb-mcp/           # MongoDB
│   ├── qdrant-mcp/            # Qdrant vector search
│   ├── supabase-mcp/          # Supabase
│   ├── aws-mcp/               # AWS Labs MCP
│   ├── kubernetes-mcp/        # Kubernetes
│   ├── playwright-mcp/        # Microsoft Playwright
│   ├── browserbase-mcp/       # Browserbase cloud browser
│   ├── sentry-mcp/            # Sentry error tracking
│   ├── heroku-mcp/            # Heroku Platform
│   ├── cloudflare-mcp/        # Cloudflare Workers/KV/R2/D1
│   ├── brave-search/          # Brave Search
│   ├── perplexity-mcp/        # Perplexity AI
│   ├── context7-mcp/          # Context7 documentation
│   ├── firecrawl-mcp/         # Firecrawl web scraping
│   ├── omnisearch-mcp/        # Unified search
│   ├── langchain-mcp/         # LangChain adapters
│   ├── llamaindex-mcp/        # LlamaIndex integration
│   ├── docs-mcp/              # Documentation retrieval
│   ├── telegram-mcp/          # Telegram
│   ├── airtable-mcp/          # Airtable
│   ├── trello-mcp/            # Trello
│   ├── obsidian-mcp/          # Obsidian knowledge management
│   ├── microsoft-mcp/         # Microsoft MCP catalog
│   ├── python-sdk/            # MCP Python SDK
│   ├── typescript-sdk/        # MCP TypeScript SDK
│   ├── registry/              # MCP Registry
│   ├── inspector/             # MCP Inspector
│   ├── create-python-server/  # Python server template
│   ├── create-typescript-server/ # TypeScript server template
│   ├── awesome-mcp-servers/   # Curated MCP servers list
│   ├── awesome-devops-mcp/    # DevOps MCP servers
│   └── ...                    # More submodules
└── servers/                   # Additional server configs
```

## Environment Variables

Create a `.env.mcp` file with your API keys:

```bash
# Version Control
export GITHUB_TOKEN="ghp_xxx"
export GITLAB_TOKEN="glpat-xxx"

# Communication
export SLACK_BOT_TOKEN="xoxb-xxx"
export SLACK_TEAM_ID="Txxx"
export TELEGRAM_API_ID="xxx"
export TELEGRAM_API_HASH="xxx"
export NOTION_API_KEY="secret_xxx"

# Databases
export REDIS_URL="redis://localhost:6379"
export MONGODB_URI="mongodb://localhost:27017/helixagent"
export QDRANT_URL="http://localhost:6333"
export ELASTICSEARCH_URL="http://localhost:9200"
export SUPABASE_URL="https://xxx.supabase.co"
export SUPABASE_KEY="xxx"

# Cloud
export AWS_ACCESS_KEY_ID="xxx"
export AWS_SECRET_ACCESS_KEY="xxx"
export AWS_REGION="us-east-1"
export CLOUDFLARE_API_TOKEN="xxx"
export CLOUDFLARE_ACCOUNT_ID="xxx"
export HEROKU_API_KEY="xxx"

# Search & AI
export BRAVE_API_KEY="BSA_xxx"
export PERPLEXITY_API_KEY="pplx-xxx"
export TAVILY_API_KEY="tvly-xxx"
export FIRECRAWL_API_KEY="fc-xxx"
export CONTEXT7_API_KEY="xxx"
export OPENAI_API_KEY="sk-xxx"
export ANTHROPIC_API_KEY="sk-ant-xxx"

# DevOps
export SENTRY_AUTH_TOKEN="xxx"
export SENTRY_ORG="xxx"

# Productivity
export AIRTABLE_API_KEY="xxx"
export TRELLO_API_KEY="xxx"
export TRELLO_TOKEN="xxx"
export JIRA_URL="https://xxx.atlassian.net"
export JIRA_EMAIL="xxx@example.com"
export JIRA_API_TOKEN="xxx"

# Browser Automation
export BROWSERBASE_API_KEY="xxx"
export BROWSERBASE_PROJECT_ID="xxx"
```

Source before running:

```bash
source .env.mcp
```

## Initializing Submodules

```bash
# Initialize all submodules
git submodule update --init --recursive

# Update all submodules to latest
git submodule update --remote --merge
```

## Building MCP Servers

### Build all servers
```bash
cd MCP/submodules/github-mcp-server && npm install && npm run build
cd MCP/submodules/slack-mcp && npm install && npm run build
# ... repeat for other Node.js servers

# Python servers
cd MCP/submodules/qdrant-mcp && pip install -e .
cd MCP/submodules/atlassian-mcp && pip install -e .
```

### Build using Docker
```bash
docker-compose -f MCP/docker-compose.yml build
```

## Testing

Run MCP challenge tests:

```bash
# All MCP tests
./challenges/scripts/cli_agent_mcp_challenge.sh

# Specific CLI agent tests
./challenges/scripts/opencode_mcp_challenge.sh
./challenges/scripts/crush_mcp_challenge.sh
```

## CLI Agent Configuration

### OpenCode

Copy to `~/.config/opencode/opencode.json`:

```json
{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["docker", "exec", "-i", "mcp-github", "node", "dist/index.js"]
    },
    "slack": {
      "type": "local",
      "command": ["docker", "exec", "-i", "mcp-slack", "node", "dist/index.js"]
    }
  }
}
```

### Crush

Copy to `~/.config/crush/crush.json`:

```json
{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["docker", "exec", "-i", "mcp-github", "node", "dist/index.js"],
      "enabled": true
    }
  }
}
```

## Sources & Attribution

MCP servers are sourced from these official repositories:

- [Model Context Protocol Servers](https://github.com/modelcontextprotocol/servers)
- [GitHub MCP Server](https://github.com/github/github-mcp-server)
- [Microsoft MCP Catalog](https://github.com/microsoft/mcp)
- [AWS Labs MCP](https://github.com/awslabs/mcp)
- [Cloudflare MCP](https://github.com/cloudflare/mcp-server-cloudflare)
- [Heroku MCP](https://github.com/heroku/heroku-mcp-server)
- [Redis MCP](https://github.com/redis/mcp-redis)
- [MongoDB MCP](https://github.com/mongodb-js/mongodb-mcp-server)
- [Qdrant MCP](https://github.com/qdrant/mcp-server-qdrant)
- [Supabase MCP](https://github.com/supabase-community/supabase-mcp)
- [Sentry MCP](https://github.com/getsentry/sentry-mcp)
- [Elasticsearch MCP](https://github.com/elastic/mcp-server-elasticsearch)
- [Playwright MCP](https://github.com/microsoft/playwright-mcp)
- [Browserbase MCP](https://github.com/browserbase/mcp-server-browserbase)
- [Notion MCP](https://github.com/makenotion/notion-mcp-server)
- [Brave Search MCP](https://github.com/brave/brave-search-mcp-server)
- [Perplexity MCP](https://github.com/perplexityai/modelcontextprotocol)
- [Context7 MCP](https://github.com/upstash/context7)
- [Firecrawl MCP](https://github.com/mendableai/firecrawl-mcp-server)
- [LangChain MCP Adapters](https://github.com/langchain-ai/langchain-mcp-adapters)
- [LlamaIndex MCP](https://github.com/run-llama/mcp-llamaindex-ai)
