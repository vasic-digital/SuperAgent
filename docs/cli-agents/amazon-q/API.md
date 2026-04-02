# Amazon Q CLI - API Reference

## Command Line Interface

### Global Commands

```bash
q [SUBCOMMAND] [options]
```

### Subcommands

| Subcommand | Description | Example |
|------------|-------------|---------|
| `chat` | Start interactive chat session | `q chat` |
| `settings` | View/modify settings | `q settings <key> <value>` |
| `login` | Authenticate with AWS | `q login` |
| `logout` | Sign out | `q logout` |
| `status` | Check authentication status | `q status` |

### Chat Options

| Option | Description | Example |
|--------|-------------|---------|
| `--agent <name>` | Use specific agent | `q chat --agent aws-expert` |
| `--verbose` | Enable verbose output | `q chat --verbose` |
| `-h, --help` | Show help | `q chat --help` |

---

## Slash Commands

### Built-in Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/help` | Show available commands | `/help` |
| `/clear` | Clear conversation history | `/clear` |
| `/exit` | Exit Q CLI | `/exit` or `Ctrl+C` |
| `/model` | Change active model | `/model` |
| `/agent` | Agent management | `/agent list`, `/agent generate` |
| `/settings` | View/edit settings | `/settings` |
| `/context` | Manage conversation context | `/context` |
| `/issue` | Report an issue | `/issue` |

### Knowledge Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/knowledge show` | Display knowledge base entries | `/knowledge show` |
| `/knowledge add` | Add to knowledge base | `/knowledge add -n "docs" -p /path` |
| `/knowledge remove` | Remove knowledge entry | `/knowledge remove "docs"` |
| `/knowledge update` | Update knowledge entry | `/knowledge update /path` |
| `/knowledge clear` | Clear all knowledge | `/knowledge clear` |
| `/knowledge cancel` | Cancel background operations | `/knowledge cancel` |

### TODO Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `/todos view` | View TODO lists | `/todos view` |
| `/todos resume` | Resume a TODO list | `/todos resume` |
| `/todos delete` | Delete TODO lists | `/todos delete --all` |
| `/clear-finished` | Remove completed lists | `/clear-finished` |

---

## Settings Reference

### Chat Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `chat.defaultAgent` | string | - | Default agent name |
| `chat.enableKnowledge` | boolean | false | Enable knowledge base |
| `chat.theme` | string | "dark" | UI theme |

### Knowledge Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `knowledge.indexType` | string | "Best" | Default index type (Fast/Best) |
| `knowledge.maxFiles` | number | 10000 | Maximum files per knowledge base |
| `knowledge.chunkSize` | number | 1024 | Text chunk size |
| `knowledge.chunkOverlap` | number | 256 | Overlap between chunks |
| `knowledge.defaultIncludePatterns` | array | [] | Default include patterns |
| `knowledge.defaultExcludePatterns` | array | [] | Default exclude patterns |

### Tool Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `tools.execute_bash.allowedCommands` | array | [] | Allowed bash commands |
| `tools.execute_bash.deniedCommands` | array | [] | Denied bash commands |
| `tools.execute_bash.autoAllowReadonly` | boolean | false | Auto-allow read-only commands |
| `tools.execute_bash.denyByDefault` | boolean | false | Deny by default |
| `tools.fs_read.allowedPaths` | array | [] | Allowed read paths |
| `tools.fs_read.deniedPaths` | array | [] | Denied read paths |
| `tools.fs_write.allowedPaths` | array | [] | Allowed write paths |
| `tools.fs_write.deniedPaths` | array | [] | Denied write paths |
| `tools.use_aws.allowedServices` | array | [] | Allowed AWS services |
| `tools.use_aws.deniedServices` | array | [] | Denied AWS services |
| `tools.use_aws.autoAllowReadonly` | boolean | false | Auto-allow read-only AWS calls |

### Settings Commands

```bash
# Get a setting
q settings <key>

# Set a setting
q settings <key> <value>

# Delete a setting
q settings --delete <key>

# List all settings
q settings --list

# Export settings
q settings --export > settings.json

# Import settings
q settings --import < settings.json
```

---

## Tool Reference

### execute_bash Tool

Execute shell commands with configurable permissions.

**Parameters:**
```json
{
  "command": "string",      // Command to execute
  "description": "string"   // Human-readable description
}
```

**Configuration:**
```json
{
  "toolsSettings": {
    "execute_bash": {
      "allowedCommands": ["git status", "git fetch"],
      "deniedCommands": ["git push .*"],
      "autoAllowReadonly": true,
      "denyByDefault": false
    }
  }
}
```

### fs_read Tool

Read files, directories, and images.

**Parameters:**
```json
{
  "operations": [
    {
      "mode": "Line",
      "path": "/path/to/file"
    }
  ]
}
```

**Configuration:**
```json
{
  "toolsSettings": {
    "fs_read": {
      "allowedPaths": ["~/projects", "./src/**"],
      "deniedPaths": ["/etc/**"]
    }
  }
}
```

### fs_write Tool

Create and edit files.

**Parameters:**
```json
{
  "operations": [
    {
      "mode": "Write",
      "path": "/path/to/file",
      "content": "file content"
    }
  ]
}
```

**Configuration:**
```json
{
  "toolsSettings": {
    "fs_write": {
      "allowedPaths": ["~/projects/**"],
      "deniedPaths": ["~/.ssh/**"]
    }
  }
}
```

### use_aws Tool

Make AWS CLI API calls.

**Parameters:**
```json
{
  "service": "s3",          // AWS service name
  "operation": "list-buckets", // Operation name
  "parameters": {}          // Operation parameters
}
```

**Configuration:**
```json
{
  "toolsSettings": {
    "use_aws": {
      "allowedServices": ["s3", "lambda"],
      "deniedServices": ["iam"],
      "autoAllowReadonly": true
    }
  }
}
```

### knowledge Tool

Store and retrieve information from knowledge base.

**Note:** This is an experimental feature that must be enabled with `chat.enableKnowledge`.

### todo_list Tool

Create and manage TODO lists for tracking multi-step tasks.

**Storage:** Lists are saved in `.amazonq/cli-todo-lists/`

### introspect Tool

Provide information about Q CLI capabilities and documentation.

Automatically used when asking questions about Q CLI itself.

### report_issue Tool

Open browser to pre-filled GitHub issue template.

---

## Agent Configuration

### Agent File Format

Agent configuration files are JSON files stored in:
- Global: `~/.aws/amazonq/cli-agents/<name>.json`
- Local: `.amazonq/cli-agents/<name>.json`

### Complete Example

```json
{
  "name": "aws-rust-agent",
  "description": "Specialized agent for AWS and Rust development",
  "prompt": "You are an expert in AWS services and Rust programming...",
  "model": "claude-sonnet-4",
  
  "mcpServers": {
    "fetch": {
      "command": "fetch3.1",
      "args": []
    },
    "git": {
      "command": "git-mcp",
      "args": [],
      "env": {
        "GIT_CONFIG_GLOBAL": "/dev/null"
      },
      "timeout": 120000
    }
  },
  
  "tools": [
    "fs_read",
    "fs_write",
    "execute_bash",
    "use_aws",
    "@git",
    "@fetch/fetch_url"
  ],
  
  "toolAliases": {
    "@git/git_status": "status",
    "@fetch/fetch_url": "get"
  },
  
  "allowedTools": [
    "fs_read",
    "fs_*",
    "@git/git_status",
    "@git"
  ],
  
  "toolsSettings": {
    "fs_write": {
      "allowedPaths": ["src/**", "tests/**", "Cargo.toml"]
    },
    "use_aws": {
      "allowedServices": ["s3", "lambda", "dynamodb"]
    }
  },
  
  "resources": [
    "file://README.md",
    "file://docs/**/*.md"
  ],
  
  "hooks": {
    "agentSpawn": [
      {
        "command": "git status"
      }
    ],
    "userPromptSubmit": [
      {
        "command": "echo 'Processing...'"
      }
    ],
    "preToolUse": [
      {
        "matcher": "execute_bash",
        "command": "echo '$(date) - Bash used' >> /tmp/audit.log"
      }
    ],
    "postToolUse": [
      {
        "matcher": "fs_write",
        "command": "cargo fmt --all"
      }
    ]
  },
  
  "useLegacyMcpJson": true
}
```

### Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Agent name (derived from filename if not specified) |
| `description` | string | No | Human-readable description |
| `prompt` | string | No | System prompt (supports `file://` URIs) |
| `model` | string | No | Model ID to use |
| `mcpServers` | object | No | MCP server definitions |
| `tools` | array | No | Available tools (`*`, `@builtin`, `@server`) |
| `toolAliases` | object | No | Tool name remapping |
| `allowedTools` | array | No | Tools that don't require permission |
| `toolsSettings` | object | No | Per-tool configuration |
| `resources` | array | No | File resources (use `file://` URIs) |
| `hooks` | object | No | Lifecycle hooks |
| `useLegacyMcpJson` | boolean | No | Include legacy MCP config |

### Wildcard Patterns

**Tools Field:**
- `"*"` - All available tools
- `"@builtin"` - All built-in tools
- `"@server_name"` - All tools from MCP server
- `"@server_name/tool_name"` - Specific MCP tool

**AllowedTools Field:**
- `"fs_*"` - All tools starting with "fs_"
- `"*_read"` - All tools ending with "_read"
- `"@server/*"` - All tools from server
- `"@git-*/status"` - Status tool from any git-* server

---

## MCP Server Configuration

### Global MCP Config

File: `~/.aws/amazonq/mcp.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "${DATABASE_URL}"
      },
      "timeout": 120000
    }
  }
}
```

### MCP Server Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `command` | string | Yes | Executable command |
| `args` | array | No | Command arguments |
| `env` | object | No | Environment variables |
| `timeout` | number | No | Request timeout in ms (default: 120000) |

---

## Hook Reference

### Hook Types

| Hook | Trigger | Can Block |
|------|---------|-----------|
| `agentSpawn` | Agent initialization | No |
| `userPromptSubmit` | User sends message | No |
| `preToolUse` | Before tool execution | Yes (exit 2) |
| `postToolUse` | After tool execution | No |
| `stop` | Response complete | No |

### Hook Event Format

**agentSpawn:**
```json
{
  "hook_event_name": "agentSpawn",
  "cwd": "/current/working/directory"
}
```

**userPromptSubmit:**
```json
{
  "hook_event_name": "userPromptSubmit",
  "cwd": "/current/working/directory",
  "prompt": "user's input prompt"
}
```

**preToolUse:**
```json
{
  "hook_event_name": "preToolUse",
  "cwd": "/current/working/directory",
  "tool_name": "fs_read",
  "tool_input": {
    "operations": [...]
  }
}
```

**postToolUse:**
```json
{
  "hook_event_name": "postToolUse",
  "cwd": "/current/working/directory",
  "tool_name": "fs_read",
  "tool_input": {...},
  "tool_response": {
    "success": true,
    "result": [...]
  }
}
```

---

## Knowledge Base API

### Index Types

| Type | Algorithm | Best For |
|------|-----------|----------|
| `Fast` | BM25 | Logs, configs, large codebases |
| `Best` | Embeddings (all-MiniLM-L6-v2) | Documentation, natural language |

### Commands

```bash
# Add knowledge
/knowledge add -n "name" -p /path [--index-type Fast|Best]

# Add with patterns
/knowledge add -n "rust-code" -p /project --include "*.rs" --exclude "target/**"

# Show all entries
/knowledge show

# Update entry
/knowledge update /path

# Remove entry
/knowledge remove "name"

# Clear all
/knowledge clear

# Cancel operations
/knowledge cancel [operation_id|all]
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `130` | Interrupted (Ctrl+C) |

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System design
- [Usage Guide](./USAGE.md) - Practical examples
- [External References](./REFERENCES.md) - Links and resources
