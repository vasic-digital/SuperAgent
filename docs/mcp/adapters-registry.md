# MCP Adapters Registry

## Overview

HelixAgent provides 45+ MCP (Model Context Protocol) adapters for integrating with external services. This document catalogs all available adapters, their capabilities, and configuration.

## Adapter Categories

| Category | Count | Description |
|----------|-------|-------------|
| Productivity | 8 | Project management and issue tracking |
| Communication | 5 | Messaging and notification services |
| Development | 8 | Code management and developer tools |
| Data | 10 | Databases, storage, and search |
| Cloud | 6 | Cloud infrastructure services |
| Specialized | 8+ | Domain-specific integrations |

## Productivity Adapters

### Linear (`linear`)

Issue tracking and project management for modern teams.

**Configuration:**
```yaml
mcp:
  adapters:
    linear:
      enabled: true
      api_key: ${LINEAR_API_KEY}
      workspace_id: ${LINEAR_WORKSPACE_ID}
```

**Tools (14):**
| Tool | Description |
|------|-------------|
| `linear_list_issues` | List issues with filters |
| `linear_create_issue` | Create a new issue |
| `linear_update_issue` | Update issue properties |
| `linear_get_issue` | Get issue details |
| `linear_delete_issue` | Delete an issue |
| `linear_list_projects` | List all projects |
| `linear_create_project` | Create a new project |
| `linear_list_teams` | List all teams |
| `linear_list_cycles` | List sprint cycles |
| `linear_create_cycle` | Create a new cycle |
| `linear_list_labels` | List available labels |
| `linear_create_label` | Create a new label |
| `linear_search` | Search issues and projects |
| `linear_get_user` | Get user information |

**Key File:** `internal/mcp/adapters/linear.go`

---

### Asana (`asana`)

Work management platform for teams.

**Configuration:**
```yaml
mcp:
  adapters:
    asana:
      enabled: true
      access_token: ${ASANA_ACCESS_TOKEN}
      default_workspace: ${ASANA_WORKSPACE_GID}
```

**Tools (20):**
| Tool | Description |
|------|-------------|
| `asana_list_workspaces` | List all workspaces |
| `asana_list_projects` | List projects in workspace |
| `asana_create_project` | Create a new project |
| `asana_get_project` | Get project details |
| `asana_list_tasks` | List tasks in project |
| `asana_create_task` | Create a new task |
| `asana_update_task` | Update task properties |
| `asana_complete_task` | Mark task as complete |
| `asana_delete_task` | Delete a task |
| `asana_list_sections` | List project sections |
| `asana_create_section` | Create a new section |
| `asana_move_task` | Move task to section |
| `asana_list_tags` | List available tags |
| `asana_create_tag` | Create a new tag |
| `asana_add_tag` | Add tag to task |
| `asana_list_users` | List workspace users |
| `asana_assign_task` | Assign task to user |
| `asana_add_comment` | Add comment to task |
| `asana_list_comments` | List task comments |
| `asana_search` | Search tasks and projects |

**Key File:** `internal/mcp/adapters/asana.go`

---

### Jira (`jira`)

Issue and project tracking for software teams.

**Configuration:**
```yaml
mcp:
  adapters:
    jira:
      enabled: true
      domain: ${JIRA_DOMAIN}  # e.g., company.atlassian.net
      email: ${JIRA_EMAIL}
      api_token: ${JIRA_API_TOKEN}
```

**Tools (20):**
| Tool | Description |
|------|-------------|
| `jira_list_projects` | List all projects |
| `jira_get_project` | Get project details |
| `jira_list_issues` | List issues with JQL |
| `jira_create_issue` | Create a new issue |
| `jira_update_issue` | Update issue fields |
| `jira_delete_issue` | Delete an issue |
| `jira_transition_issue` | Change issue status |
| `jira_assign_issue` | Assign issue to user |
| `jira_add_comment` | Add comment to issue |
| `jira_list_comments` | List issue comments |
| `jira_list_sprints` | List sprints in board |
| `jira_create_sprint` | Create a new sprint |
| `jira_move_to_sprint` | Move issue to sprint |
| `jira_list_boards` | List all boards |
| `jira_get_board` | Get board details |
| `jira_list_epics` | List all epics |
| `jira_create_epic` | Create a new epic |
| `jira_link_issues` | Link two issues |
| `jira_search` | Search with JQL |
| `jira_get_user` | Get user information |

**Key File:** `internal/mcp/adapters/jira.go`

---

### Notion (`notion`)

All-in-one workspace for notes and collaboration.

**Configuration:**
```yaml
mcp:
  adapters:
    notion:
      enabled: true
      api_key: ${NOTION_API_KEY}
```

**Tools (12):**
| Tool | Description |
|------|-------------|
| `notion_list_databases` | List all databases |
| `notion_query_database` | Query database with filters |
| `notion_create_page` | Create a new page |
| `notion_update_page` | Update page properties |
| `notion_get_page` | Get page content |
| `notion_delete_page` | Archive a page |
| `notion_append_blocks` | Add blocks to page |
| `notion_get_blocks` | Get page blocks |
| `notion_delete_block` | Delete a block |
| `notion_search` | Search pages and databases |
| `notion_create_database` | Create a new database |
| `notion_get_user` | Get user information |

---

### Todoist (`todoist`)

Task management for personal productivity.

**Configuration:**
```yaml
mcp:
  adapters:
    todoist:
      enabled: true
      api_token: ${TODOIST_API_TOKEN}
```

**Tools (10):**
| Tool | Description |
|------|-------------|
| `todoist_list_projects` | List all projects |
| `todoist_create_project` | Create a new project |
| `todoist_list_tasks` | List tasks with filters |
| `todoist_create_task` | Create a new task |
| `todoist_update_task` | Update task properties |
| `todoist_complete_task` | Mark task complete |
| `todoist_delete_task` | Delete a task |
| `todoist_list_labels` | List available labels |
| `todoist_create_label` | Create a new label |
| `todoist_add_comment` | Add comment to task |

---

### Trello (`trello`)

Visual project management with boards.

**Configuration:**
```yaml
mcp:
  adapters:
    trello:
      enabled: true
      api_key: ${TRELLO_API_KEY}
      token: ${TRELLO_TOKEN}
```

**Tools (15):**
| Tool | Description |
|------|-------------|
| `trello_list_boards` | List all boards |
| `trello_create_board` | Create a new board |
| `trello_list_lists` | List board lists |
| `trello_create_list` | Create a new list |
| `trello_list_cards` | List cards in list |
| `trello_create_card` | Create a new card |
| `trello_update_card` | Update card properties |
| `trello_move_card` | Move card to list |
| `trello_delete_card` | Delete a card |
| `trello_add_comment` | Add comment to card |
| `trello_list_members` | List board members |
| `trello_assign_member` | Assign member to card |
| `trello_add_label` | Add label to card |
| `trello_add_checklist` | Add checklist to card |
| `trello_search` | Search cards and boards |

---

### ClickUp (`clickup`)

All-in-one productivity platform.

**Tools (16):** Similar to Asana/Jira with spaces, folders, lists, tasks.

---

### Monday.com (`monday`)

Work operating system.

**Tools (14):** Boards, groups, items, updates, status columns.

---

## Communication Adapters

### Slack (`slack`)

Team messaging and collaboration.

**Configuration:**
```yaml
mcp:
  adapters:
    slack:
      enabled: true
      bot_token: ${SLACK_BOT_TOKEN}
      app_token: ${SLACK_APP_TOKEN}
```

**Tools (18):**
| Tool | Description |
|------|-------------|
| `slack_send_message` | Send message to channel |
| `slack_reply_thread` | Reply in thread |
| `slack_list_channels` | List all channels |
| `slack_create_channel` | Create a new channel |
| `slack_join_channel` | Join a channel |
| `slack_leave_channel` | Leave a channel |
| `slack_get_channel_info` | Get channel details |
| `slack_list_users` | List workspace users |
| `slack_get_user_info` | Get user details |
| `slack_send_dm` | Send direct message |
| `slack_upload_file` | Upload file to channel |
| `slack_get_messages` | Get channel messages |
| `slack_add_reaction` | Add emoji reaction |
| `slack_search_messages` | Search messages |
| `slack_set_topic` | Set channel topic |
| `slack_invite_user` | Invite user to channel |
| `slack_kick_user` | Remove user from channel |
| `slack_update_message` | Edit a message |

**Key File:** `internal/mcp/adapters/slack.go`

---

### Discord (`discord`)

Community and gaming platform.

**Tools (16):** Similar to Slack with guilds, channels, roles, reactions.

---

### Gmail (`gmail`)

Email service integration.

**Configuration:**
```yaml
mcp:
  adapters:
    gmail:
      enabled: true
      credentials_file: ${GMAIL_CREDENTIALS_FILE}
      token_file: ${GMAIL_TOKEN_FILE}
```

**Tools (12):**
| Tool | Description |
|------|-------------|
| `gmail_list_messages` | List messages with query |
| `gmail_get_message` | Get message content |
| `gmail_send_email` | Send a new email |
| `gmail_reply_email` | Reply to email |
| `gmail_forward_email` | Forward an email |
| `gmail_delete_message` | Delete message |
| `gmail_trash_message` | Move to trash |
| `gmail_list_labels` | List all labels |
| `gmail_add_label` | Add label to message |
| `gmail_remove_label` | Remove label |
| `gmail_mark_read` | Mark as read |
| `gmail_mark_unread` | Mark as unread |

---

### Microsoft Teams (`teams`)

Microsoft collaboration platform.

**Tools (14):** Teams, channels, messages, meetings, calls.

---

### Twilio (`twilio`)

SMS and communication platform.

**Tools (8):** Send SMS, make calls, verify phone numbers.

---

## Development Adapters

### GitHub (`github`)

Code hosting and collaboration.

**Configuration:**
```yaml
mcp:
  adapters:
    github:
      enabled: true
      token: ${GITHUB_TOKEN}
```

**Tools (25):**
| Tool | Description |
|------|-------------|
| `github_list_repos` | List repositories |
| `github_get_repo` | Get repo details |
| `github_create_repo` | Create repository |
| `github_list_issues` | List repo issues |
| `github_create_issue` | Create an issue |
| `github_update_issue` | Update issue |
| `github_close_issue` | Close an issue |
| `github_list_prs` | List pull requests |
| `github_create_pr` | Create pull request |
| `github_merge_pr` | Merge pull request |
| `github_get_file` | Get file content |
| `github_create_file` | Create/update file |
| `github_list_branches` | List branches |
| `github_create_branch` | Create branch |
| `github_list_commits` | List commits |
| `github_get_commit` | Get commit details |
| `github_list_releases` | List releases |
| `github_create_release` | Create release |
| `github_list_actions` | List workflow runs |
| `github_trigger_workflow` | Trigger workflow |
| `github_list_comments` | List PR comments |
| `github_add_comment` | Add PR comment |
| `github_review_pr` | Review pull request |
| `github_search_code` | Search code |
| `github_search_issues` | Search issues |

**Key File:** `internal/mcp/adapters/github.go`

---

### GitLab (`gitlab`)

DevOps platform.

**Tools (22):** Similar to GitHub with CI/CD pipelines.

---

### Sentry (`sentry`)

Error tracking and monitoring.

**Configuration:**
```yaml
mcp:
  adapters:
    sentry:
      enabled: true
      auth_token: ${SENTRY_AUTH_TOKEN}
      organization: ${SENTRY_ORG}
```

**Tools (12):**
| Tool | Description |
|------|-------------|
| `sentry_list_projects` | List projects |
| `sentry_list_issues` | List error issues |
| `sentry_get_issue` | Get issue details |
| `sentry_resolve_issue` | Resolve an issue |
| `sentry_ignore_issue` | Ignore an issue |
| `sentry_assign_issue` | Assign to user |
| `sentry_list_events` | List error events |
| `sentry_get_event` | Get event details |
| `sentry_list_releases` | List releases |
| `sentry_create_release` | Create release |
| `sentry_list_deploys` | List deployments |
| `sentry_create_deploy` | Record deployment |

---

### Brave Search (`brave_search`)

Privacy-focused web search.

**Tools (3):**
| Tool | Description |
|------|-------------|
| `brave_web_search` | Search the web |
| `brave_local_search` | Search local businesses |
| `brave_news_search` | Search news articles |

---

### Browserbase (`browserbase`)

Headless browser service.

**Tools (8):** Create session, navigate, screenshot, evaluate JS, close session.

---

### CircleCI (`circleci`)

Continuous integration platform.

**Tools (10):** Pipelines, workflows, jobs, artifacts.

---

### Jenkins (`jenkins`)

Automation server.

**Tools (10):** Jobs, builds, pipelines, nodes.

---

### Vercel (`vercel`)

Frontend cloud platform.

**Tools (10):** Deployments, domains, environment variables.

---

## Data Adapters

### PostgreSQL (`postgres`)

Relational database.

**Configuration:**
```yaml
mcp:
  adapters:
    postgres:
      enabled: true
      connection_string: ${POSTGRES_URL}
```

**Tools (8):**
| Tool | Description |
|------|-------------|
| `postgres_query` | Execute SELECT query |
| `postgres_execute` | Execute INSERT/UPDATE/DELETE |
| `postgres_list_tables` | List all tables |
| `postgres_describe_table` | Get table schema |
| `postgres_list_schemas` | List schemas |
| `postgres_list_indexes` | List indexes |
| `postgres_explain` | Explain query plan |
| `postgres_transaction` | Execute in transaction |

**Key File:** `internal/mcp/adapters/postgres.go`

---

### Qdrant (`qdrant`)

Vector similarity search.

**Configuration:**
```yaml
mcp:
  adapters:
    qdrant:
      enabled: true
      host: ${QDRANT_HOST}
      port: ${QDRANT_PORT}
      api_key: ${QDRANT_API_KEY}
```

**Tools (10):**
| Tool | Description |
|------|-------------|
| `qdrant_list_collections` | List collections |
| `qdrant_create_collection` | Create collection |
| `qdrant_delete_collection` | Delete collection |
| `qdrant_upsert` | Insert/update vectors |
| `qdrant_search` | Similarity search |
| `qdrant_get_points` | Get points by ID |
| `qdrant_delete_points` | Delete points |
| `qdrant_count` | Count points |
| `qdrant_scroll` | Scroll through points |
| `qdrant_recommend` | Get recommendations |

---

### Google Drive (`google_drive`)

Cloud file storage.

**Tools (12):** List, upload, download, share, create folders.

---

### AWS S3 (`s3`)

Object storage.

**Tools (10):** List buckets, list objects, get, put, delete, presign URLs.

---

### MongoDB (`mongodb`)

Document database.

**Tools (10):** Find, insert, update, delete, aggregate, indexes.

---

### Redis (`redis`)

In-memory data store.

**Tools (12):** GET, SET, DELETE, HGET, HSET, LPUSH, RPOP, ZADD, etc.

---

### Elasticsearch (`elasticsearch`)

Search and analytics.

**Tools (10):** Search, index, delete, bulk, aggregations.

---

### Pinecone (`pinecone`)

Vector database.

**Tools (8):** Upsert, query, delete, describe index, list indexes.

---

### Snowflake (`snowflake`)

Cloud data warehouse.

**Tools (8):** Query, list schemas, list tables, describe table.

---

### BigQuery (`bigquery`)

Google analytics data warehouse.

**Tools (8):** Query, list datasets, list tables, get schema.

---

## Cloud Adapters

### AWS (`aws`)

Amazon Web Services.

**Tools (20+):** EC2, S3, Lambda, RDS, DynamoDB, SQS, SNS, CloudWatch.

---

### GCP (`gcp`)

Google Cloud Platform.

**Tools (15+):** Compute, Storage, BigQuery, Pub/Sub, Cloud Functions.

---

### Azure (`azure`)

Microsoft Azure.

**Tools (15+):** VMs, Storage, Cosmos DB, Functions, Service Bus.

---

### Cloudflare (`cloudflare`)

Edge computing and CDN.

**Tools (10):** DNS, Workers, KV, R2, WAF rules.

---

### DigitalOcean (`digitalocean`)

Cloud infrastructure.

**Tools (10):** Droplets, volumes, domains, load balancers.

---

### Heroku (`heroku`)

Platform as a service.

**Tools (8):** Apps, dynos, addons, config vars.

---

## Specialized Adapters

### OpenAI (`openai`)

OpenAI API integration.

**Tools (6):** Completions, embeddings, images, audio, moderation.

---

### Anthropic (`anthropic`)

Claude API integration.

**Tools (4):** Messages, embeddings.

---

### HuggingFace (`huggingface`)

ML models hub.

**Tools (6):** Inference, models, datasets, spaces.

---

### LangChain (`langchain`)

LLM application framework.

**Tools (8):** Chains, agents, memory, tools.

---

### Stripe (`stripe`)

Payment processing.

**Tools (15):** Customers, payments, subscriptions, invoices.

---

### Shopify (`shopify`)

E-commerce platform.

**Tools (12):** Products, orders, customers, inventory.

---

### Salesforce (`salesforce`)

CRM platform.

**Tools (15):** Accounts, contacts, opportunities, leads, cases.

---

### Zapier (`zapier`)

Automation platform.

**Tools (5):** Trigger zap, list zaps, get zap status.

---

## Creating Custom Adapters

### Adapter Interface

```go
// Adapter defines the MCP adapter interface.
type Adapter interface {
    // Name returns the adapter name.
    Name() string

    // Description returns a human-readable description.
    Description() string

    // Tools returns the list of available tools.
    Tools() []Tool

    // Execute runs a tool with the given parameters.
    Execute(ctx context.Context, tool string, params map[string]interface{}) (interface{}, error)

    // Configure sets up the adapter with configuration.
    Configure(config map[string]interface{}) error

    // HealthCheck checks adapter connectivity.
    HealthCheck(ctx context.Context) error
}
```

### Example Custom Adapter

```go
package adapters

import (
    "context"
    "dev.helix.agent/internal/mcp"
)

type CustomAdapter struct {
    apiKey string
    client *CustomClient
}

func NewCustomAdapter(config map[string]interface{}) (*CustomAdapter, error) {
    apiKey, _ := config["api_key"].(string)
    return &CustomAdapter{
        apiKey: apiKey,
        client: NewCustomClient(apiKey),
    }, nil
}

func (a *CustomAdapter) Name() string {
    return "custom"
}

func (a *CustomAdapter) Description() string {
    return "Custom service integration"
}

func (a *CustomAdapter) Tools() []mcp.Tool {
    return []mcp.Tool{
        {
            Name:        "custom_action",
            Description: "Perform custom action",
            Parameters: map[string]mcp.Parameter{
                "input": {Type: "string", Required: true},
            },
        },
    }
}

func (a *CustomAdapter) Execute(ctx context.Context, tool string, params map[string]interface{}) (interface{}, error) {
    switch tool {
    case "custom_action":
        input := params["input"].(string)
        return a.client.DoAction(ctx, input)
    default:
        return nil, fmt.Errorf("unknown tool: %s", tool)
    }
}

func (a *CustomAdapter) Configure(config map[string]interface{}) error {
    if key, ok := config["api_key"].(string); ok {
        a.apiKey = key
        a.client = NewCustomClient(key)
    }
    return nil
}

func (a *CustomAdapter) HealthCheck(ctx context.Context) error {
    return a.client.Ping(ctx)
}
```

### Registration

```go
// In internal/mcp/adapters/registry.go
func init() {
    Register("custom", NewCustomAdapter)
}
```

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
**Author**: Generated by Claude Code
