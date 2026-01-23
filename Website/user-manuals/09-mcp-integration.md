# User Manual 09: MCP Integration Guide

## Introduction

The Model Context Protocol (MCP) enables HelixAgent to integrate with external services, databases, and tools. This guide covers how to configure, use, and extend MCP adapters.

## Overview

MCP provides a standardized way to:
- Connect AI to external data sources
- Execute actions in third-party services
- Build context from multiple sources
- Automate workflows across platforms

## Quick Start

### 1. Enable MCP

```yaml
# config.yaml
mcp:
  enabled: true
  adapters:
    slack:
      enabled: true
      bot_token: ${SLACK_BOT_TOKEN}
    github:
      enabled: true
      token: ${GITHUB_TOKEN}
```

### 2. Use MCP Tools

```bash
# Via API
curl -X POST http://localhost:8080/v1/mcp/execute \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "adapter": "slack",
    "tool": "slack_send_message",
    "parameters": {
      "channel": "#general",
      "text": "Hello from HelixAgent!"
    }
  }'
```

### 3. List Available Tools

```bash
curl http://localhost:8080/v1/mcp/tools \
  -H "Authorization: Bearer $API_KEY"
```

## Available Adapters

### Productivity

| Adapter | Description | Setup Guide |
|---------|-------------|-------------|
| `linear` | Issue tracking | [Linear Setup](#linear-setup) |
| `asana` | Project management | [Asana Setup](#asana-setup) |
| `jira` | Issue tracking | [Jira Setup](#jira-setup) |
| `notion` | Workspace | [Notion Setup](#notion-setup) |
| `todoist` | Task management | [Todoist Setup](#todoist-setup) |
| `trello` | Kanban boards | [Trello Setup](#trello-setup) |

### Communication

| Adapter | Description | Setup Guide |
|---------|-------------|-------------|
| `slack` | Team messaging | [Slack Setup](#slack-setup) |
| `discord` | Community chat | [Discord Setup](#discord-setup) |
| `gmail` | Email | [Gmail Setup](#gmail-setup) |
| `teams` | Microsoft Teams | [Teams Setup](#teams-setup) |

### Development

| Adapter | Description | Setup Guide |
|---------|-------------|-------------|
| `github` | Code hosting | [GitHub Setup](#github-setup) |
| `gitlab` | DevOps platform | [GitLab Setup](#gitlab-setup) |
| `sentry` | Error tracking | [Sentry Setup](#sentry-setup) |

### Data

| Adapter | Description | Setup Guide |
|---------|-------------|-------------|
| `postgres` | PostgreSQL | [PostgreSQL Setup](#postgresql-setup) |
| `qdrant` | Vector search | [Qdrant Setup](#qdrant-setup) |
| `google_drive` | File storage | [Drive Setup](#google-drive-setup) |

## Adapter Setup

### Linear Setup

1. **Get API Key**
   - Go to Linear Settings → API → Create new key
   - Copy the API key

2. **Configure**
   ```yaml
   mcp:
     adapters:
       linear:
         enabled: true
         api_key: ${LINEAR_API_KEY}
         workspace_id: your-workspace-id
   ```

3. **Available Tools**
   - `linear_list_issues` - List issues with filters
   - `linear_create_issue` - Create new issue
   - `linear_update_issue` - Update issue
   - `linear_list_projects` - List projects
   - `linear_search` - Search issues

4. **Example Usage**
   ```bash
   # Create an issue
   curl -X POST http://localhost:8080/v1/mcp/execute \
     -H "Authorization: Bearer $API_KEY" \
     -d '{
       "adapter": "linear",
       "tool": "linear_create_issue",
       "parameters": {
         "title": "Bug: Login page broken",
         "description": "Users cannot log in",
         "team_id": "team-123",
         "priority": 1
       }
     }'
   ```

### Slack Setup

1. **Create Slack App**
   - Go to https://api.slack.com/apps
   - Click "Create New App"
   - Choose "From scratch"
   - Name your app and select workspace

2. **Configure Permissions**
   - Go to "OAuth & Permissions"
   - Add scopes:
     - `chat:write`
     - `channels:read`
     - `channels:history`
     - `users:read`

3. **Install App**
   - Install to workspace
   - Copy Bot User OAuth Token

4. **Configure HelixAgent**
   ```yaml
   mcp:
     adapters:
       slack:
         enabled: true
         bot_token: xoxb-your-token
         app_token: xapp-your-token  # For socket mode
   ```

5. **Available Tools**
   - `slack_send_message` - Send message
   - `slack_list_channels` - List channels
   - `slack_get_messages` - Get channel history
   - `slack_upload_file` - Upload file

### GitHub Setup

1. **Create Personal Access Token**
   - Go to GitHub Settings → Developer Settings → Personal Access Tokens
   - Create new token (classic or fine-grained)
   - Required scopes: `repo`, `read:org`

2. **Configure**
   ```yaml
   mcp:
     adapters:
       github:
         enabled: true
         token: ${GITHUB_TOKEN}
   ```

3. **Available Tools**
   - `github_list_repos` - List repositories
   - `github_create_issue` - Create issue
   - `github_create_pr` - Create pull request
   - `github_get_file` - Get file content
   - `github_search_code` - Search code

### PostgreSQL Setup

1. **Configure Connection**
   ```yaml
   mcp:
     adapters:
       postgres:
         enabled: true
         connection_string: postgresql://user:pass@host:5432/db
         # Or individual settings
         host: localhost
         port: 5432
         user: postgres
         password: ${POSTGRES_PASSWORD}
         database: mydb
         ssl_mode: require
   ```

2. **Available Tools**
   - `postgres_query` - Execute SELECT
   - `postgres_execute` - Execute INSERT/UPDATE/DELETE
   - `postgres_list_tables` - List tables
   - `postgres_describe_table` - Get schema

3. **Security**
   - Use read-only user for queries
   - Limit accessible schemas
   - Enable query logging

### Qdrant Setup

1. **Configure**
   ```yaml
   mcp:
     adapters:
       qdrant:
         enabled: true
         host: localhost
         port: 6333
         api_key: ${QDRANT_API_KEY}  # If using cloud
   ```

2. **Available Tools**
   - `qdrant_search` - Similarity search
   - `qdrant_upsert` - Insert/update vectors
   - `qdrant_list_collections` - List collections
   - `qdrant_delete_points` - Delete vectors

## Using MCP in Conversations

### Natural Language

HelixAgent can automatically use MCP tools based on your requests:

```
User: Create a GitHub issue for the login bug we discussed

HelixAgent: I'll create that issue for you.
[Uses github_create_issue tool]
Created issue #123: "Login page not loading"
```

### Explicit Tool Calls

You can also explicitly request tool usage:

```
User: Use the slack adapter to send "Deployment complete" to #releases

HelixAgent: I'll send that message now.
[Uses slack_send_message tool]
Message sent to #releases
```

## Tool Chaining

MCP tools can be chained together for complex workflows:

```yaml
# Example: Create issue from Sentry error
workflow:
  - adapter: sentry
    tool: sentry_get_issue
    parameters:
      issue_id: ${error_id}
    output: error_details

  - adapter: github
    tool: github_create_issue
    parameters:
      title: "Bug: ${error_details.title}"
      body: |
        ## Error Details
        ${error_details.message}

        ## Stack Trace
        ${error_details.stacktrace}
      labels: ["bug", "sentry"]
    output: github_issue

  - adapter: slack
    tool: slack_send_message
    parameters:
      channel: "#bugs"
      text: "New bug from Sentry: ${github_issue.url}"
```

## API Reference

### List Adapters

```bash
GET /v1/mcp/adapters

Response:
{
  "adapters": [
    {
      "name": "slack",
      "enabled": true,
      "status": "connected",
      "tools_count": 18
    },
    {
      "name": "github",
      "enabled": true,
      "status": "connected",
      "tools_count": 25
    }
  ]
}
```

### List Tools

```bash
GET /v1/mcp/tools?adapter=slack

Response:
{
  "tools": [
    {
      "name": "slack_send_message",
      "description": "Send a message to a Slack channel",
      "parameters": {
        "channel": {
          "type": "string",
          "required": true,
          "description": "Channel name or ID"
        },
        "text": {
          "type": "string",
          "required": true,
          "description": "Message text"
        }
      }
    }
  ]
}
```

### Execute Tool

```bash
POST /v1/mcp/execute
Content-Type: application/json

{
  "adapter": "slack",
  "tool": "slack_send_message",
  "parameters": {
    "channel": "#general",
    "text": "Hello!"
  }
}

Response:
{
  "success": true,
  "result": {
    "ts": "1234567890.123456",
    "channel": "C0123456"
  }
}
```

### Check Adapter Health

```bash
GET /v1/mcp/adapters/slack/health

Response:
{
  "adapter": "slack",
  "status": "healthy",
  "latency_ms": 45,
  "last_check": "2026-01-23T10:00:00Z"
}
```

## Troubleshooting

### Connection Issues

**Problem**: Adapter shows "disconnected"

**Solution**:
1. Verify credentials are correct
2. Check network connectivity
3. Ensure API endpoint is accessible
4. Review rate limits

```bash
# Test connection
curl http://localhost:8080/v1/mcp/adapters/slack/health
```

### Permission Errors

**Problem**: "Insufficient permissions" error

**Solution**:
1. Review required scopes for the tool
2. Update token with correct permissions
3. Reinstall app if using OAuth

### Rate Limiting

**Problem**: "Rate limit exceeded" error

**Solution**:
1. Check adapter rate limit settings
2. Implement request queuing
3. Use batch operations where available

```yaml
mcp:
  adapters:
    github:
      rate_limit:
        requests_per_minute: 30
        burst_size: 10
```

## Security Best Practices

### 1. Credential Management

```bash
# Use environment variables
export SLACK_BOT_TOKEN=xoxb-...

# Or use secrets manager
mcp:
  adapters:
    slack:
      bot_token: ${vault:secret/slack/bot_token}
```

### 2. Least Privilege

- Only enable required adapters
- Use minimum required scopes
- Create dedicated service accounts

### 3. Audit Logging

```yaml
mcp:
  audit:
    enabled: true
    log_parameters: false  # Don't log sensitive params
    log_results: false
```

### 4. Input Validation

All tool parameters are validated before execution. Custom validation can be added:

```yaml
mcp:
  adapters:
    postgres:
      validation:
        allowed_tables: ["users", "orders"]
        blocked_operations: ["DROP", "TRUNCATE"]
```

## Next Steps

- Review [MCP Adapters Registry](/docs/mcp/adapters-registry.md) for all available adapters
- Learn about [Agentic Workflows](/docs/guides/agentic-workflows.md) for tool chaining
- See [Security Guide](10-security-hardening.md) for securing MCP access

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
