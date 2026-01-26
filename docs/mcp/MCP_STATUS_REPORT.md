# MCP Status Report

**Generated**: 2026-01-26
**Total MCPs Configured**: 72
**Working MCPs**: ~20 (without additional API keys)

---

## Overview

HelixAgent supports 72 MCP (Model Context Protocol) servers. This report shows which MCPs are working, which require additional setup, and how to enable them.

---

## Section 1: Always Working MCPs (13)

These MCPs require no external API keys or services and work out of the box:

| MCP | Package | Description |
|-----|---------|-------------|
| `helixagent` | Local plugin | HelixAgent MCP integration |
| `filesystem` | `@modelcontextprotocol/server-filesystem` | File system access |
| `fetch` | `@modelcontextprotocol/server-fetch` | HTTP requests |
| `memory` | `@modelcontextprotocol/server-memory` | In-memory storage |
| `time` | `@modelcontextprotocol/server-time` | Time utilities |
| `git` | `@modelcontextprotocol/server-git` | Git operations |
| `sequential-thinking` | `@modelcontextprotocol/server-sequential-thinking` | Reasoning |
| `everything` | `@anthropic-ai/mcp-server-everything` | Test MCP |
| `sqlite` | `@modelcontextprotocol/server-sqlite` | Local SQLite database |
| `docker` | `mcp-server-docker` | Container management |
| `puppeteer` | `@modelcontextprotocol/server-puppeteer` | Browser automation |
| `playwright` | `mcp-server-playwright` | Browser automation |
| `ansible` | `mcp-server-ansible` | Configuration management |

---

## Section 2: Local Service MCPs (6)

These MCPs require local services started via Docker/Podman:

| MCP | Service | Port | Status | Start Command |
|-----|---------|------|--------|---------------|
| `postgres` | PostgreSQL | 15432 | ✅ Running | `./scripts/mcp/start-mcp-services.sh --databases` |
| `redis` | Redis | 16379 | ✅ Running | `./scripts/mcp/start-mcp-services.sh --databases` |
| `mongodb` | MongoDB | 27017 | ❌ Not Running | `./scripts/mcp/start-mcp-services.sh --databases` |
| `elasticsearch` | Elasticsearch | 9200 | ❌ Not Running | `./scripts/mcp/start-mcp-services.sh --databases` |
| `qdrant` | Qdrant | 6333 | ✅ Running | `./scripts/mcp/start-mcp-services.sh --vectors` |
| `chroma` | Chroma | 8000 | ✅ Running | `./scripts/mcp/start-mcp-services.sh --vectors` |

### How to Start All Services

```bash
# Start all MCP backend services
./scripts/mcp/start-mcp-services.sh --all

# Or start specific categories
./scripts/mcp/start-mcp-services.sh --databases  # Redis, MongoDB, PostgreSQL, MySQL, Elasticsearch
./scripts/mcp/start-mcp-services.sh --vectors    # Qdrant, Chroma
./scripts/mcp/start-mcp-services.sh --minimal    # Redis, PostgreSQL, Qdrant
```

---

## Section 3: API Key Required MCPs (53)

These MCPs require external API keys. Add them to your `.env` file:

### Development Platforms

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `github` | `GITHUB_TOKEN` | ✅ Set | https://github.com/settings/tokens |
| `gitlab` | `GITLAB_TOKEN` | ❌ Not Set | https://gitlab.com/-/profile/personal_access_tokens |
| `sentry` | `SENTRY_AUTH_TOKEN`, `SENTRY_ORG` | ❌ Not Set | https://sentry.io/settings/account/api/auth-tokens/ |

### Communication & Collaboration

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `slack` | `SLACK_BOT_TOKEN`, `SLACK_TEAM_ID` | ❌ Not Set | https://api.slack.com/apps |
| `discord` | `DISCORD_TOKEN` | ❌ Not Set | https://discord.com/developers/applications |
| `microsoft-teams` | `TEAMS_TOKEN` | ❌ Not Set | https://portal.azure.com |
| `zoom` | `ZOOM_CLIENT_ID`, `ZOOM_CLIENT_SECRET` | ❌ Not Set | https://marketplace.zoom.us |

### Project Management

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `notion` | `NOTION_API_KEY` | ❌ Not Set | https://www.notion.so/my-integrations |
| `linear` | `LINEAR_API_KEY` | ❌ Not Set | Linear Settings → API |
| `jira` | `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN` | ❌ Not Set | https://id.atlassian.com/manage-profile/security/api-tokens |
| `asana` | `ASANA_ACCESS_TOKEN` | ❌ Not Set | https://app.asana.com/0/developer-console |
| `trello` | `TRELLO_API_KEY`, `TRELLO_API_TOKEN` | ❌ Not Set | https://trello.com/power-ups/admin |
| `monday` | `MONDAY_API_TOKEN` | ❌ Not Set | https://monday.com/developers/apps |
| `todoist` | `TODOIST_API_TOKEN` | ❌ Not Set | https://todoist.com/app/settings/integrations |
| `clickup` | `CLICKUP_API_KEY` | ❌ Not Set | https://clickup.com/settings |

### Search & AI

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `brave-search` | `BRAVE_API_KEY` | ❌ Not Set | https://brave.com/search/api/ |
| `exa` | `EXA_API_KEY` | ❌ Not Set | https://exa.ai/ |
| `openai` | `OPENAI_API_KEY` | ❌ Not Set | https://platform.openai.com/api-keys |
| `huggingface` | `HF_TOKEN` | ❌ Not Set | https://huggingface.co/settings/tokens |
| `replicate` | `REPLICATE_API_TOKEN` | ❌ Not Set | https://replicate.com/account/api-tokens |

### Cloud Providers

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `cloudflare` | `CLOUDFLARE_API_TOKEN` | ❌ Not Set | https://dash.cloudflare.com/profile/api-tokens |
| `aws-s3`, `aws-lambda`, `aws-kb-retrieval` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | ❌ Not Set | https://console.aws.amazon.com/iam/ |
| `azure` | `AZURE_SUBSCRIPTION_ID` | ❌ Not Set | https://portal.azure.com |
| `gcp` | `GOOGLE_APPLICATION_CREDENTIALS` | ❌ Not Set | https://console.cloud.google.com |
| `supabase` | `SUPABASE_URL`, `SUPABASE_KEY` | ❌ Not Set | https://supabase.com |
| `neon` | `NEON_API_KEY` | ❌ Not Set | https://neon.tech |

### Google Services

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `google-maps` | `GOOGLE_MAPS_API_KEY` | ❌ Not Set | https://console.cloud.google.com |
| `gdrive`, `gmail`, `calendar` | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | ❌ Not Set | https://console.cloud.google.com |

### Monitoring & Observability

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `datadog` | `DD_API_KEY`, `DD_APP_KEY` | ❌ Not Set | https://app.datadoghq.com |
| `grafana` | `GRAFANA_URL`, `GRAFANA_TOKEN` | ❌ Not Set | Grafana Admin |
| `prometheus` | `PROMETHEUS_URL` | ❌ Not Set | (local service) |

### Vector Databases (Cloud)

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `pinecone` | `PINECONE_API_KEY` | ❌ Not Set | https://app.pinecone.io/ |
| `weaviate` | `WEAVIATE_URL` | ❌ Not Set | (local or cloud) |
| `milvus` | `MILVUS_URL` | ❌ Not Set | (local or cloud) |

### Design

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `figma` | `FIGMA_API_KEY` | ❌ Not Set | https://www.figma.com/developers/api |

### Other

| MCP | Environment Variable | Status | Get API Key From |
|-----|---------------------|--------|------------------|
| `obsidian` | `OBSIDIAN_VAULT_PATH` | ❌ Not Set | (local path to vault) |
| `kubernetes` | `KUBECONFIG` | ❌ Not Set | (local kubeconfig path) |
| `terraform` | `TF_TOKEN` | ❌ Not Set | https://app.terraform.io |
| `circleci` | `CIRCLECI_TOKEN` | ❌ Not Set | https://circleci.com/account/api |

---

## Section 4: Summary

| Category | Count | Notes |
|----------|-------|-------|
| Always Working | 13 | No setup required |
| Local Services Running | 4/6 | Start with `./scripts/mcp/start-mcp-services.sh` |
| API Keys Configured | 1 | Add more to `.env` file |
| **TOTAL WORKING** | **~18** | Without additional API keys |
| **TOTAL CONFIGURED** | **72** | All MCPs in config |

---

## How to Enable More MCPs

### Step 1: Start Local Services

```bash
# Start all MCP backend services
./scripts/mcp/start-mcp-services.sh --all
```

### Step 2: Add API Keys

Copy `.env.mcp.example` to `.env.mcp` and fill in your API keys:

```bash
cp .env.mcp.example .env.mcp
nano .env.mcp
```

Then source the file:

```bash
source .env.mcp
```

### Step 3: Regenerate Configuration

```bash
./bin/helixagent --generate-opencode-config --opencode-output ~/.config/opencode/.opencode.json
./bin/helixagent --generate-crush-config --crush-output ~/.config/crush/crush.json
```

### Step 4: Restart OpenCode

```bash
opencode
```

---

## MCPs That CANNOT Work Without External Accounts

The following MCPs **require external API keys or accounts** that cannot be generated locally. You must sign up for these services:

| Category | MCPs | Why |
|----------|------|-----|
| **Cloud Services** | aws-*, azure, gcp, cloudflare, vercel, heroku, supabase, neon | Require cloud provider accounts |
| **SaaS Platforms** | slack, discord, notion, linear, jira, asana, trello, monday, todoist, clickup, zoom, microsoft-teams | Require SaaS subscriptions/free accounts |
| **Search APIs** | brave-search, exa | Require API subscriptions |
| **AI Services** | openai, huggingface, replicate | Require AI platform accounts |
| **Design Tools** | figma | Require Figma account |
| **Monitoring** | datadog, sentry | Require monitoring platform accounts |
| **Vector DBs (Cloud)** | pinecone | Require cloud vector DB accounts |

**Total MCPs requiring external accounts: ~40**

---

## Quick Reference: Environment Variables

Add these to your `.env` file to enable more MCPs:

```bash
# Most commonly needed
GITHUB_TOKEN=ghp_xxxxxxxxxxxx          # GitHub API
OPENAI_API_KEY=sk-xxxxxxxxxxxx         # OpenAI API
BRAVE_API_KEY=BSAxxxxxxxxxxxx          # Brave Search
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxx      # Slack
NOTION_API_KEY=secret_xxxxxxxxxxxx     # Notion

# Local services (auto-configured if using start script)
REDIS_URL=redis://:helixagent123@localhost:16379
POSTGRES_URL=postgresql://helixagent:helixagent123@localhost:15432/helixagent_db
QDRANT_URL=http://localhost:6333
CHROMA_URL=http://localhost:8000
```

---

## Related Documentation

- [MCP Validation Guide](MCP_VALIDATION_GUIDE.md)
- [MCP Configuration Requirements](MCP_CONFIGURATION_REQUIREMENTS.md)
- [CLI Agent Configuration](../cli-agents/README.md)
