# MCP Configuration Requirements

This document lists all 43 MCPs configured for HelixAgent CLI agents (OpenCode, Crush) and their prerequisites.

## Quick Reference

| Status | Count | Description |
|--------|-------|-------------|
| No Config Required | 14 | Works out of the box |
| Local Service Required | 6 | Requires local service running |
| API Key Required | 14 | Requires API key/token |
| HelixAgent Remote | 9 | Requires HelixAgent running |

---

## MCPs That Work Without Configuration (14)

These MCPs work immediately without any additional setup:

| MCP | Package | Description |
|-----|---------|-------------|
| **filesystem** | `@modelcontextprotocol/server-filesystem` | File system operations |
| **memory** | `@modelcontextprotocol/server-memory` | In-memory key-value storage |
| **sequential-thinking** | `@modelcontextprotocol/server-sequential-thinking` | Chain-of-thought reasoning |
| **everything** | `@modelcontextprotocol/server-everything` | Demo server with all capabilities |
| **puppeteer** | `@modelcontextprotocol/server-puppeteer` | Browser automation |
| **docker** | `mcp-server-docker` | Docker container management |
| **kubernetes** | `mcp-server-kubernetes` | K8s cluster management (uses ~/.kube/config) |
| **git** | `git-mcp-server` | Git repository operations |
| **time** | `time-mcp-server` | Current time and timezone info |
| **sqlite** | `mcp-server-sqlite` | SQLite database operations |
| **qdrant** | `mcp-server-qdrant` | Qdrant vector database (default: localhost:6333) |
| **chroma** | `mcp-server-chroma` | ChromaDB vector database (default: localhost:8000) |
| **youtube** | `youtube-mcp-server` | YouTube video information |
| **google** | `google-mcp-server` | Google services (requires one-time OAuth setup) |

---

## MCPs Requiring Local Services (6)

These MCPs require a local service to be running:

### postgres
**Package:** `@modelcontextprotocol/server-postgres`

**Preconditions:**
- PostgreSQL server running
- Database credentials

**Configuration:**
```bash
# Connection string passed as argument
npx -y @modelcontextprotocol/server-postgres "postgresql://user:password@localhost:5432/database"
```

**Current Config:** Uses HelixAgent's PostgreSQL: `postgresql://helixagent:helixagent123@localhost:5432/helixagent`

---

### mongodb
**Package:** `mongodb-mcp-server`

**Preconditions:**
- MongoDB server running (local or Atlas)
- Connection string

**Environment Variable:**
```bash
export MDB_MCP_CONNECTION_STRING="mongodb://localhost:27017/mydb"
# OR for Atlas:
export MDB_MCP_CONNECTION_STRING="mongodb+srv://user:pass@cluster.mongodb.net/mydb"
```

**Atlas API Access (optional for cloud management):**
```bash
export MDB_MCP_ATLAS_CLIENT_ID="your-atlas-client-id"
export MDB_MCP_ATLAS_CLIENT_SECRET="your-atlas-secret"
```

---

### mysql
**Package:** `mcp-server-mysql`

**Preconditions:**
- MySQL server running
- Node.js 20+ required
- Database credentials

**Environment Variables:**
```bash
export MYSQL_HOST="localhost"
export MYSQL_PORT="3306"
export MYSQL_USER="root"
export MYSQL_PASSWORD="your-password"
export MYSQL_DATABASE="your-database"
```

---

### elasticsearch
**Package:** `mcp-server-elasticsearch`

**Preconditions:**
- Elasticsearch server running
- API key with appropriate permissions

**Environment Variables:**
```bash
export ES_URL="http://localhost:9200"
export ES_API_KEY="your-elasticsearch-api-key"
```

---

### qdrant
**Package:** `mcp-server-qdrant`

**Preconditions:**
- Qdrant server running (default: localhost:6333)

**Environment Variables (optional):**
```bash
export QDRANT_URL="http://localhost:6333"
export QDRANT_API_KEY="your-api-key"  # If authentication enabled
```

---

### chroma
**Package:** `mcp-server-chroma`

**Preconditions:**
- ChromaDB server running (default: localhost:8000)

**Environment Variables (optional):**
```bash
export CHROMA_URL="http://localhost:8000"
```

---

## MCPs Requiring API Keys/Tokens (14)

### github
**Package:** `github-mcp-server`

**Preconditions:**
- GitHub account
- Personal Access Token (PAT) or GitHub App

**Get Token:**
1. Go to https://github.com/settings/tokens
2. Generate new token (classic)
3. Select scopes: `repo`, `read:org`, `read:user`

**Environment Variable:**
```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
# OR
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
```

---

### gitlab
**Package:** `gitlab-mcp-server`

**Preconditions:**
- GitLab account
- Personal Access Token

**Get Token:**
1. Go to GitLab → Settings → Access Tokens
2. Create token with `api`, `read_repository` scopes

**Environment Variable:**
```bash
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
# OR
export GITLAB_PERSONAL_ACCESS_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

---

### slack
**Package:** `slack-mcp-server`

**Preconditions:**
- Slack workspace admin access
- Slack App with Bot Token

**Get Token:**
1. Go to https://api.slack.com/apps
2. Create New App → From scratch
3. Add Bot Token Scopes: `channels:history`, `channels:read`, `chat:write`, `users:read`
4. Install to Workspace
5. Copy Bot User OAuth Token

**Environment Variable:**
```bash
export SLACK_BOT_TOKEN="xoxb-xxxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx"
```

**Optional - Enable message posting:**
```bash
export SLACK_ENABLE_SEND_MESSAGE="true"
export SLACK_ALLOWED_CHANNELS="C12345678,C87654321"  # Restrict to specific channels
```

---

### discord
**Package:** `mcp-server-discord`

**Preconditions:**
- Discord server with admin access
- Discord Bot Token

**Get Token:**
1. Go to https://discord.com/developers/applications
2. Create New Application
3. Go to Bot → Add Bot
4. Copy Token
5. Enable Privileged Gateway Intents (MESSAGE CONTENT INTENT)
6. Invite bot to server with Administrator permissions

**Environment Variable:**
```bash
export DISCORD_BOT_TOKEN="MTxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

---

### telegram
**Package:** `telegram-mcp-server`

**Preconditions:**
- Telegram account
- Telegram API credentials

**Get Credentials:**
1. Go to https://my.telegram.org/apps
2. Create new application
3. Get `api_id` and `api_hash`

**Environment Variables:**
```bash
export TELEGRAM_API_ID="12345678"
export TELEGRAM_API_HASH="abcdef1234567890abcdef1234567890"
```

---

### linear
**Package:** `mcp-server-linear`

**Preconditions:**
- Linear account
- API Token

**Get Token:**
1. Go to Linear → Settings → API
2. Create Personal API Key or Developer Token

**Environment Variable:**
```bash
export LINEAR_ACCESS_TOKEN="lin_api_xxxxxxxxxxxxxxxxxxxx"
# OR
export LINEAR_API_KEY="lin_api_xxxxxxxxxxxxxxxxxxxx"
```

---

### notion
**Package:** `notion-mcp-server`

**Preconditions:**
- Notion account
- Integration token

**Get Token:**
1. Go to https://www.notion.so/my-integrations
2. Create New Integration
3. Select workspace
4. Copy Internal Integration Token
5. Share pages/databases with the integration

**Environment Variable:**
```bash
export NOTION_API_KEY="secret_xxxxxxxxxxxxxxxxxxxx"
# OR
export NOTION_TOKEN="secret_xxxxxxxxxxxxxxxxxxxx"
```

---

### jira
**Package:** `jira-mcp-server`

**Preconditions:**
- Jira Cloud or Server account
- API Token (Cloud) or Password (Server)

**Get Token (Jira Cloud):**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Create API Token

**Environment Variables:**
```bash
export JIRA_HOST="https://your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="ATATT3xxxxxxxxxxxxxxxxxxx"
```

---

### trello
**Package:** `mcp-server-trello`

**Preconditions:**
- Trello account
- API Key and Token

**Get Credentials:**
1. Go to https://trello.com/power-ups/admin
2. Create new Power-Up or use existing
3. Get API Key
4. Generate Token with API Key

**Environment Variables:**
```bash
export TRELLO_API_KEY="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TRELLO_TOKEN="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TRELLO_BOARD_ID="xxxxxxxx"  # Optional default board
```

---

### brave-search
**Package:** `mcp-server-brave-search`

**Preconditions:**
- Brave Search API account

**Get API Key:**
1. Go to https://brave.com/search/api/
2. Sign up for API access
3. Get API Key from dashboard

**Environment Variable:**
```bash
export BRAVE_API_KEY="BSAxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

---

### openai
**Package:** `openai-mcp-server`

**Preconditions:**
- OpenAI account with API access

**Get API Key:**
1. Go to https://platform.openai.com/api-keys
2. Create new secret key

**Configuration:**
```bash
# Pass as argument:
npx openai-mcp-server "sk-xxxxxxxxxxxxxxxxxxxx"

# OR environment variable:
export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxx"
```

---

### twitter
**Package:** `twitter-mcp-server`

**Preconditions:**
- Twitter/X Developer account
- API credentials (v2)

**Get Credentials:**
1. Go to https://developer.twitter.com/en/portal/dashboard
2. Create Project and App
3. Generate API Key, API Secret, Bearer Token

**Environment Variables:**
```bash
export TWITTER_API_KEY="xxxxxxxxxxxxxxxxxxxx"
export TWITTER_API_SECRET="xxxxxxxxxxxxxxxxxxxx"
export TWITTER_BEARER_TOKEN="xxxxxxxxxxxxxxxxxxxx"
export TWITTER_ACCESS_TOKEN="xxxxxxxxxxxxxxxxxxxx"
export TWITTER_ACCESS_SECRET="xxxxxxxxxxxxxxxxxxxx"
```

---

### vercel
**Package:** `vercel-mcp-server`

**Preconditions:**
- Vercel account
- Access Token

**Get Token:**
1. Go to https://vercel.com/account/tokens
2. Create new token

**Environment Variable:**
```bash
export VERCEL_TOKEN="xxxxxxxxxxxxxxxxxxxx"
# OR
export VERCEL_ACCESS_TOKEN="xxxxxxxxxxxxxxxxxxxx"
```

---

### cloudflare
**Package:** `mcp-server-cloudflare`

**Preconditions:**
- Cloudflare account
- API Token

**Get Token:**
1. Go to https://dash.cloudflare.com/profile/api-tokens
2. Create Token with appropriate permissions

**Environment Variables:**
```bash
export CLOUDFLARE_API_TOKEN="xxxxxxxxxxxxxxxxxxxx"
# OR for Global API Key:
export CLOUDFLARE_API_KEY="xxxxxxxxxxxxxxxxxxxx"
export CLOUDFLARE_EMAIL="your-email@example.com"
```

---

## Cloud Provider MCPs (3)

### aws
**Package:** `mcp-server-aws`

**Preconditions:**
- AWS account
- IAM credentials or AWS CLI configured

**Configuration Options:**

**Option 1: Environment Variables**
```bash
export AWS_ACCESS_KEY_ID="AKIAXXXXXXXXXXXXXXXX"
export AWS_SECRET_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export AWS_REGION="us-east-1"
```

**Option 2: AWS CLI Profile**
```bash
# Uses default profile from ~/.aws/credentials
aws configure
```

**Option 3: IAM Role (for EC2/Lambda)**
- Attach appropriate IAM role to instance/function

---

### gcp
**Package:** `mcp-server-gcp`

**Preconditions:**
- Google Cloud account
- Service account or user credentials

**Configuration Options:**

**Option 1: Service Account Key**
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

**Option 2: gcloud CLI**
```bash
gcloud auth application-default login
```

---

### google
**Package:** `google-mcp-server`

**Preconditions:**
- Google Cloud project
- OAuth2 credentials

**Setup:**
```bash
# Run one-time setup
npx google-mcp-server --setup
```

This will:
1. Open browser for OAuth consent
2. Save credentials to local config
3. Enable required APIs

---

## HelixAgent Remote MCPs (9)

These MCPs connect to HelixAgent's API endpoints:

| MCP | Endpoint | Description |
|-----|----------|-------------|
| **helixagent-mcp** | `/v1/mcp` | Model Context Protocol interface |
| **helixagent-acp** | `/v1/acp` | Agent Communication Protocol |
| **helixagent-lsp** | `/v1/lsp` | Language Server Protocol |
| **helixagent-embeddings** | `/v1/embeddings` | Vector embeddings API |
| **helixagent-vision** | `/v1/vision` | Image analysis API |
| **helixagent-cognee** | `/v1/cognee` | Knowledge graph & RAG |
| **helixagent-tools-search** | `/v1/mcp/tools/search` | MCP Tool Search |
| **helixagent-adapters-search** | `/v1/mcp/adapters/search` | MCP Adapter Search |
| **helixagent-tools-suggestions** | `/v1/mcp/tools/suggestions` | Tool suggestions |

**Preconditions:**
- HelixAgent running on localhost:7061
- HELIXAGENT_API_KEY environment variable set (optional for local dev)

**Environment Variable:**
```bash
export HELIXAGENT_API_KEY="your-api-key"  # Optional for local development
```

---

## Environment Variables Summary

Create a `.env` file or add to your shell profile:

```bash
# === REQUIRED FOR FULL FUNCTIONALITY ===

# GitHub (for github MCP)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# Slack (for slack MCP)
export SLACK_BOT_TOKEN="xoxb-xxxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx"

# Linear (for linear MCP)
export LINEAR_ACCESS_TOKEN="lin_api_xxxxxxxxxxxxxxxxxxxx"

# Notion (for notion MCP)
export NOTION_API_KEY="secret_xxxxxxxxxxxxxxxxxxxx"

# Brave Search (for brave-search MCP)
export BRAVE_API_KEY="BSAxxxxxxxxxxxxxxxxxxxxxxxxxx"

# OpenAI (for openai MCP)
export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxx"

# === OPTIONAL - CLOUD PROVIDERS ===

# AWS
export AWS_ACCESS_KEY_ID="AKIAXXXXXXXXXXXXXXXX"
export AWS_SECRET_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export AWS_REGION="us-east-1"

# GCP
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# === OPTIONAL - ADDITIONAL SERVICES ===

# GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"

# Discord
export DISCORD_BOT_TOKEN="MTxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Telegram
export TELEGRAM_API_ID="12345678"
export TELEGRAM_API_HASH="abcdef1234567890abcdef1234567890"

# Jira
export JIRA_HOST="https://your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="ATATT3xxxxxxxxxxxxxxxxxxx"

# Trello
export TRELLO_API_KEY="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TRELLO_TOKEN="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Twitter
export TWITTER_BEARER_TOKEN="xxxxxxxxxxxxxxxxxxxx"

# Vercel
export VERCEL_TOKEN="xxxxxxxxxxxxxxxxxxxx"

# Cloudflare
export CLOUDFLARE_API_TOKEN="xxxxxxxxxxxxxxxxxxxx"

# Elasticsearch
export ES_URL="http://localhost:9200"
export ES_API_KEY="your-elasticsearch-api-key"

# MongoDB
export MDB_MCP_CONNECTION_STRING="mongodb://localhost:27017/mydb"

# MySQL
export MYSQL_HOST="localhost"
export MYSQL_USER="root"
export MYSQL_PASSWORD="your-password"
export MYSQL_DATABASE="your-database"

# YouTube
export YOUTUBE_API_KEY="AIzaxxxxxxxxxxxxxxxxxxxxxxxxxx"

# HelixAgent (optional for local)
export HELIXAGENT_API_KEY="your-api-key"
```

---

## Troubleshooting

### MCP shows "Connection closed" error
- Package may not be installed yet (first run downloads it)
- Check if required environment variables are set
- Verify the package name is correct

### MCP shows "Connected" but doesn't work
- API key may be invalid or expired
- Service may be unreachable
- Insufficient permissions

### MCP shows "Authentication failed"
- Double-check credential format
- Regenerate tokens if expired
- Verify scopes/permissions

### Local service MCPs don't work
- Ensure service is running (docker, postgres, etc.)
- Check connection parameters (host, port)
- Verify network access

---

## MCP Categories by Use Case

### Development
- filesystem, git, github, gitlab, docker, kubernetes

### Databases
- postgres, mongodb, mysql, sqlite, elasticsearch, qdrant, chroma

### Communication
- slack, discord, telegram

### Project Management
- linear, notion, jira, trello

### AI & Search
- openai, brave-search, google, youtube

### Cloud & DevOps
- aws, gcp, vercel, cloudflare

### HelixAgent Integration
- helixagent-mcp, helixagent-acp, helixagent-lsp, helixagent-embeddings
- helixagent-vision, helixagent-cognee, helixagent-tools-search
- helixagent-adapters-search, helixagent-tools-suggestions

---

## Version Information

| MCP | Package Version | Last Verified |
|-----|-----------------|---------------|
| @modelcontextprotocol/server-filesystem | 2026.1.14 | 2026-01-24 |
| @modelcontextprotocol/server-memory | 2025.11.25 | 2026-01-24 |
| @modelcontextprotocol/server-postgres | 0.6.2 | 2026-01-24 |
| @modelcontextprotocol/server-puppeteer | 2025.5.12 | 2026-01-24 |
| @modelcontextprotocol/server-sequential-thinking | 2025.12.18 | 2026-01-24 |
| @modelcontextprotocol/server-everything | 2026.1.14 | 2026-01-24 |
| mcp-server-sqlite | 0.0.2 | 2026-01-24 |
| mcp-server-docker | 1.0.0 | 2026-01-24 |
| mcp-server-kubernetes | 3.2.0 | 2026-01-24 |
| git-mcp-server | 1.0.0 | 2026-01-24 |
| mcp-server-qdrant | 0.0.1 | 2026-01-24 |
| mcp-server-chroma | 0.0.1 | 2026-01-24 |
| mcp-server-elasticsearch | 0.2.0 | 2026-01-24 |
| time-mcp-server | 1.0.0 | 2026-01-24 |
| github-mcp-server | 1.8.7 | 2026-01-24 |
| gitlab-mcp-server | 0.0.1 | 2026-01-24 |
| slack-mcp-server | 1.1.28 | 2026-01-24 |
| mcp-server-discord | 1.2.8 | 2026-01-24 |
| telegram-mcp-server | 1.0.0 | 2026-01-24 |
| mcp-server-linear | 1.6.0 | 2026-01-24 |
| notion-mcp-server | 1.0.1 | 2026-01-24 |
| jira-mcp-server | 0.0.1 | 2026-01-24 |
| mcp-server-trello | 1.0.4 | 2026-01-24 |
| youtube-mcp-server | 1.0.0 | 2026-01-24 |
| twitter-mcp-server | 0.1.1 | 2026-01-24 |
| google-mcp-server | 1.0.0 | 2026-01-24 |
| mcp-server-brave-search | 1.0.0 | 2026-01-24 |
| openai-mcp-server | 0.1.6 | 2026-01-24 |
| mongodb-mcp-server | 1.5.0 | 2026-01-24 |
| mcp-server-mysql | 1.0.42 | 2026-01-24 |
| mcp-server-aws | 0.0.1 | 2026-01-24 |
| mcp-server-gcp | 0.0.1 | 2026-01-24 |
| vercel-mcp-server | 1.0.0 | 2026-01-24 |
| mcp-server-cloudflare | 0.0.1 | 2026-01-24 |
