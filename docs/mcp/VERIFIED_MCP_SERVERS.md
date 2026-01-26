# Verified MCP Servers for HelixAgent

This document lists all **92 verified MCP servers** available for HelixAgent CLI agents (OpenCode, Crush, etc.).

## MCP Server Sources

### Git Submodules (48 repositories)

HelixAgent includes 48 MCP server repositories as git submodules under `MCP/submodules/`:

| Category | Submodules |
|----------|------------|
| **Official** | github-mcp-server, notion-mcp-server, sentry-mcp, heroku-mcp |
| **Databases** | redis-mcp, mongodb-mcp, qdrant-mcp, elasticsearch-mcp, supabase-mcp |
| **Cloud** | aws-mcp, kubernetes-mcp, k8s-mcp-server, cloudflare-mcp, workers-mcp |
| **Browser** | playwright-mcp, mcp-playwright-ea, browserbase-mcp |
| **Productivity** | slack-mcp, telegram-mcp, atlassian-mcp, airtable-mcp, trello-mcp, obsidian-mcp, all-in-one-mcp |
| **Search & AI** | brave-search, perplexity-mcp, context7-mcp, firecrawl-mcp, omnisearch-mcp |
| **AI Frameworks** | langchain-mcp, llamaindex-mcp, docs-mcp |
| **Microsoft** | microsoft-mcp |
| **SDKs** | python-sdk, typescript-sdk, create-python-server, create-typescript-server |
| **Registry** | registry, inspector |
| **Awesome Lists** | awesome-mcp-servers, appcypher-awesome-mcp, habito-awesome-mcp, ever-works-awesome-mcp, awesome-devops-mcp |

### npm Packages (57 packages)

In addition to submodules, these npm packages are configured:

#### Core (No Auth Required) - 16 Servers

| Server | npm Package | Description |
|--------|-------------|-------------|
| helixagent | Built-in plugin | HelixAgent protocol tools (ACP, LSP, Embeddings, Vision, Cognee) |
| everything | @modelcontextprotocol/server-everything | Reference server with all MCP features |
| filesystem | @modelcontextprotocol/server-filesystem | File system operations |
| memory | @modelcontextprotocol/server-memory | Persistent memory storage |
| sequential-thinking | @modelcontextprotocol/server-sequential-thinking | Step-by-step reasoning |
| fetch | mcp-server-fetch (pip) | HTTP requests and web fetching |
| git | mcp-server-git (pip) | Git operations |
| time | mcp-server-time (pip) | Time and timezone utilities |
| puppeteer | @modelcontextprotocol/server-puppeteer | Browser automation |
| docker | mcp-server-docker | Docker container management |
| kubernetes | mcp-server-kubernetes | Kubernetes cluster management |
| hackernews | mcp-server-hackernews | Hacker News API |
| wikipedia | wikipedia-mcp-server | Wikipedia search |
| context7 | @upstash/context7-mcp | Context management |
| chrome-devtools | chrome-devtools-mcp | Chrome DevTools integration |
| playwright | @playwright/mcp | Cross-browser automation |

#### Database & Storage - 8 Servers

| Server | npm Package | Required Env Vars |
|--------|-------------|-------------------|
| sqlite | @modelcontextprotocol/server-sqlite | (file path) |
| postgres | @modelcontextprotocol/server-postgres | DATABASE_URL |
| mongodb | mongodb-mcp-server | MONGODB_URI |
| redis | @hamaster/redis-mcp-server | REDIS_URL |
| elasticsearch | mcp-server-elasticsearch | ELASTICSEARCH_URL |
| qdrant | mcp-server-qdrant | QDRANT_URL |
| neon | mcp-server-neon | NEON_API_KEY |
| aws-s3 | aws-s3-mcp | AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY |

#### Version Control & DevOps - 9 Servers

| Server | npm Package | Required Env Vars |
|--------|-------------|-------------------|
| github | @modelcontextprotocol/server-github | GITHUB_TOKEN |
| gitlab | @modelcontextprotocol/server-gitlab | GITLAB_TOKEN |
| bitbucket | mcp-server-bitbucket | BITBUCKET_USERNAME, BITBUCKET_APP_PASSWORD |
| vercel | mcp-server-vercel | VERCEL_TOKEN |
| netlify | mcp-server-netlify | NETLIFY_AUTH_TOKEN |
| heroku | @heroku/mcp-server | HEROKU_API_KEY |
| sentry | @sentry/mcp-server | SENTRY_AUTH_TOKEN |
| datadog | datadog-mcp-server | DD_API_KEY, DD_APP_KEY |
| dynatrace | @dynatrace-oss/dynatrace-mcp-server | DYNATRACE_URL, DYNATRACE_API_TOKEN |

#### Productivity & Communication - 11 Servers

| Server | npm Package | Required Env Vars |
|--------|-------------|-------------------|
| slack | @modelcontextprotocol/server-slack | SLACK_BOT_TOKEN |
| discord | mcp-server-discord | DISCORD_TOKEN |
| notion | @notionhq/notion-mcp-server | NOTION_API_KEY |
| linear | mcp-linear | LINEAR_API_KEY |
| jira | @rokealvo/jira-mcp | JIRA_HOST, JIRA_EMAIL, JIRA_API_TOKEN |
| todoist | todoist-mcp | TODOIST_API_KEY |
| trello | mcp-server-trello | TRELLO_API_KEY, TRELLO_TOKEN |
| airtable | mcp-server-airtable | AIRTABLE_API_KEY |
| google-drive | @piotr-agier/google-drive-mcp | GOOGLE_CREDENTIALS |
| obsidian | @jianruidutong/obsidian-mcp | OBSIDIAN_VAULT_PATH |
| figma | figma-mcp | FIGMA_ACCESS_TOKEN |

#### Search & AI - 8 Servers

| Server | npm Package | Required Env Vars |
|--------|-------------|-------------------|
| brave-search | @brave/brave-search-mcp-server | BRAVE_API_KEY |
| exa | exa-mcp-server | EXA_API_KEY |
| tavily | tavily-mcp | TAVILY_API_KEY |
| openai | mcp-server-openai | OPENAI_API_KEY |
| apify | @apify/actors-mcp-server | APIFY_TOKEN |
| axiom | mcp-server-axiom | AXIOM_TOKEN |
| mapbox | @mapbox/mcp-server | MAPBOX_ACCESS_TOKEN |
| weather | mcp-server-weather | OPENWEATHER_API_KEY |

#### Social & Commerce - 5 Servers

| Server | npm Package | Required Env Vars |
|--------|-------------|-------------------|
| youtube | mcp-server-youtube | YOUTUBE_API_KEY |
| twitter | mcp-server-twitter | TWITTER_BEARER_TOKEN |
| reddit | mcp-server-reddit | REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET |
| stripe | mcp-server-stripe | STRIPE_SECRET_KEY |
| shopify | mcp-server-shopify | SHOPIFY_ACCESS_TOKEN, SHOPIFY_STORE_URL |

## Installation

### Using Git Submodules (Recommended)

```bash
# Initialize all MCP submodules
cd /path/to/HelixAgent
git submodule update --init --recursive

# Update submodules to latest
git submodule update --remote --merge
```

### Using Docker/Podman Compose

```bash
# Start all MCP servers via Docker Compose
docker-compose -f MCP/docker-compose.yml up -d

# Or using Podman
podman-compose -f MCP/docker-compose.yml up -d

# View logs
docker-compose -f MCP/docker-compose.yml logs -f

# Stop services
docker-compose -f MCP/docker-compose.yml down
```

### Install npm MCP Servers

```bash
# For OpenCode/Crush via npx (automatic)
npx -y @modelcontextprotocol/server-github

# Or install globally
npm install -g @modelcontextprotocol/server-github
```

### Install Python MCP Servers

```bash
pip3 install mcp-server-fetch mcp-server-git mcp-server-time
```

### Install HelixAgent MCP Plugin

```bash
cd plugins/mcp-server
npm install && npm run build
```

## Configuration

### OpenCode

Copy the template and customize:

```bash
cp scripts/cli-agents/configs/opencode.template.json ~/.config/opencode/opencode.json
# Edit paths and add your API keys
```

### Crush

```bash
cp scripts/cli-agents/configs/crush.template.json ~/.config/crush/crush.json
# Edit paths and add your API keys
```

## Environment Variables

Create a `.env.mcp` file with your API keys:

```bash
# Version Control
export GITHUB_TOKEN="ghp_xxx"
export GITLAB_TOKEN="glpat-xxx"

# Communication
export SLACK_BOT_TOKEN="xoxb-xxx"
export NOTION_API_KEY="secret_xxx"

# Search & AI
export BRAVE_API_KEY="BSA_xxx"
export OPENAI_API_KEY="sk-xxx"
export ANTHROPIC_API_KEY="sk-ant-xxx"
export PERPLEXITY_API_KEY="pplx-xxx"

# Databases
export REDIS_URL="redis://localhost:6379"
export MONGODB_URI="mongodb://localhost:27017/helixagent"
export QDRANT_URL="http://localhost:6333"

# Cloud
export AWS_ACCESS_KEY_ID="xxx"
export AWS_SECRET_ACCESS_KEY="xxx"
export CLOUDFLARE_API_TOKEN="xxx"

# ... add other keys as needed
```

Source before running CLI agents:

```bash
source .env.mcp
```

## Verification

Run the MCP verification challenges:

```bash
# MCP submodules challenge (48 tests)
./challenges/scripts/mcp_submodules_challenge.sh

# CLI agent MCP challenge (26 tests)
./challenges/scripts/cli_agent_mcp_challenge.sh
```

## Package Verification Status

All 57 npm/pip packages have been verified to exist as of 2026-01-26:
- npm packages verified using `npm view <package> version`
- Python packages verified using `pip3 show <package>`

### Packages NOT Available (Removed from Config)

The following packages were found to NOT exist and have been removed:
- @anthropic-ai/brave-search-mcp (use @brave/brave-search-mcp-server)
- @anthropic-ai/sentry-mcp (use @sentry/mcp-server)
- @anthropic-ai/google-maps-mcp (no alternative found)
- @anthropic-ai/everart-mcp (no alternative found)
- @anthropic-ai/aws-kb-retrieval-mcp (no alternative found)
- cloudflare-mcp-server (no verified package)
- mcp-server-twilio, sendgrid, mailgun (no packages)
- mcp-server-replicate, huggingface, e2b (no packages)
- mcp-server-confluence, miro, circleci, jenkins (no packages)
- mcp-server-planetscale, supabase, dropbox (no packages)
- mcp-server-newrelic, pagerduty (no packages)

## Sources & Attribution

MCP servers are sourced from:

- [Model Context Protocol Official](https://github.com/modelcontextprotocol)
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
- [Notion MCP](https://github.com/makenotion/notion-mcp-server)
- [Brave Search MCP](https://github.com/brave/brave-search-mcp-server)
- [Perplexity MCP](https://github.com/perplexityai/modelcontextprotocol)
- [Context7 MCP](https://github.com/upstash/context7)
- [LangChain MCP](https://github.com/langchain-ai/langchain-mcp-adapters)
- [LlamaIndex MCP](https://github.com/run-llama/mcp-llamaindex-ai)
