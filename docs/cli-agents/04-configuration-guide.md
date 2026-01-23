# Configuration Guide

Detailed guide for configuring CLI agents to work with HelixAgent.

## Configuration Formats

HelixAgent supports multiple configuration formats depending on the agent:

| Format | Agents |
|--------|--------|
| JSON | OpenCode, Cline, Codex, most agents |
| YAML | Aider, Goose, Bridle, TaskWeaver |
| TOML | GPT Engineer, GPTME, OpenHands |
| XML | IntelliJ AI (for IDE integration) |

## Universal Configuration Structure

All generated configurations follow a common structure:

```json
{
  "version": "1.0",
  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "http://localhost:7061/v1",
    "api_key": "",
    "api_key_env": "HELIXAGENT_API_KEY"
  },
  "models": [
    {
      "id": "helixagent-debate",
      "name": "HelixAgent AI Debate Ensemble",
      "max_tokens": 128000,
      "capabilities": ["vision", "streaming", "function_calls", "embeddings", "mcp", "acp", "lsp"]
    }
  ],
  "mcp": {
    // MCP server configurations
  },
  "settings": {
    // Agent-specific settings
  }
}
```

## Provider Configuration

### OpenAI-Compatible Endpoint

HelixAgent exposes an OpenAI-compatible API:

```json
{
  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "http://localhost:7061/v1",
    "api_key_env": "HELIXAGENT_API_KEY"
  }
}
```

### Available Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/chat/completions` | Chat completions (AI Debate) |
| `/v1/completions` | Text completions |
| `/v1/embeddings` | Vector embeddings |
| `/v1/models` | List available models |
| `/v1/mcp` | MCP protocol |
| `/v1/acp` | Agent Communication Protocol |
| `/v1/lsp` | Language Server Protocol |
| `/v1/vision` | Vision/image analysis |
| `/v1/cognee` | Knowledge graph & RAG |

## Model Configuration

### The HelixAgent Debate Model

HelixAgent presents as a single model that internally uses the AI Debate Ensemble:

```json
{
  "models": [
    {
      "id": "helixagent-debate",
      "name": "HelixAgent AI Debate Ensemble",
      "max_tokens": 128000,
      "capabilities": [
        "vision",
        "streaming",
        "function_calls",
        "embeddings",
        "mcp",
        "acp",
        "lsp"
      ]
    }
  ]
}
```

### Model Capabilities

| Capability | Description |
|------------|-------------|
| `vision` | Image analysis and OCR |
| `streaming` | Server-sent events streaming |
| `function_calls` | Tool use / function calling |
| `embeddings` | Vector embeddings generation |
| `mcp` | Model Context Protocol support |
| `acp` | Agent Communication Protocol |
| `lsp` | Language Server Protocol |

## MCP Server Configuration

### Default MCP Servers

HelixAgent configures 12 MCP servers by default:

#### HelixAgent Remote Servers (6)

```json
{
  "mcp": {
    "helixagent-mcp": {
      "type": "remote",
      "url": "http://localhost:7061/v1/mcp"
    },
    "helixagent-acp": {
      "type": "remote",
      "url": "http://localhost:7061/v1/acp"
    },
    "helixagent-lsp": {
      "type": "remote",
      "url": "http://localhost:7061/v1/lsp"
    },
    "helixagent-embeddings": {
      "type": "remote",
      "url": "http://localhost:7061/v1/embeddings"
    },
    "helixagent-vision": {
      "type": "remote",
      "url": "http://localhost:7061/v1/vision"
    },
    "helixagent-cognee": {
      "type": "remote",
      "url": "http://localhost:7061/v1/cognee"
    }
  }
}
```

#### Standard Local Servers (6)

```json
{
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"]
    },
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"]
    },
    "memory": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-memory"]
    },
    "fetch": {
      "type": "local",
      "command": ["npx", "-y", "mcp-fetch"]
    },
    "puppeteer": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-puppeteer"]
    },
    "sqlite": {
      "type": "local",
      "command": ["npx", "-y", "mcp-server-sqlite"]
    }
  }
}
```

### Adding Custom MCP Servers

Add your own MCP servers:

```json
{
  "mcp": {
    "my-custom-server": {
      "type": "local",
      "command": ["node", "/path/to/server.js"],
      "env": {
        "API_KEY": "secret"
      }
    }
  }
}
```

## Agent-Specific Settings

Each agent has specific default settings. Here are examples:

### Aider Settings

```json
{
  "settings": {
    "auto_commits": true,
    "edit_format": "diff",
    "stream": true,
    "show_diffs": true
  }
}
```

### Continue Settings

```json
{
  "settings": {
    "allowAnonymousTelemetry": false,
    "tabAutocomplete": true,
    "systemMessage": "You are a helpful AI coding assistant powered by HelixAgent."
  }
}
```

### Cline Settings

```json
{
  "settings": {
    "autoApprove": false,
    "streamResponse": true,
    "showDiff": true
  }
}
```

### Codex Settings

```json
{
  "settings": {
    "permissions": ["read", "write", "execute"],
    "streaming": true
  }
}
```

### OpenHands Settings

```json
{
  "settings": {
    "runtime": "docker",
    "workspace": "~/openhands-workspace",
    "streaming": true
  }
}
```

### TaskWeaver Settings

```json
{
  "settings": {
    "codeInterpreter": true,
    "plugins": []
  }
}
```

## Environment Variables

### Required Variables

```bash
# API key for HelixAgent authentication
export HELIXAGENT_API_KEY="your-api-key"
```

### Optional Variables

```bash
# Override HelixAgent host
export HELIXAGENT_HOST="localhost"
export HELIXAGENT_PORT="7061"

# Provider API keys (if using providers directly)
export CLAUDE_API_KEY="..."
export OPENAI_API_KEY="..."
export DEEPSEEK_API_KEY="..."
export GEMINI_API_KEY="..."
```

### Agent-Specific Variables

Some agents use environment variables for their config directory:

| Agent | Environment Variable |
|-------|---------------------|
| Continue | `CONTINUE_CONFIG_DIR` |
| Cline | `CLINE_CONFIG_DIR` |
| Claude Code | `CLAUDE_CONFIG_DIR` |
| Qwen Code | `QWEN_CODE_CONFIG_DIR` |

## Configuration Validation

### Validate Single Configuration

```bash
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json
```

### Validation Output

**Valid configuration:**
```
✓ Config file is valid for opencode

Warnings:
  - $schema field is recommended for validation
```

**Invalid configuration:**
```
✗ Config file is invalid for opencode

Errors:
  - provider.base_url is required
  - mcp.helixagent: type is required
```

### Schema Validation Rules

Each agent has specific validation rules:

**OpenCode:**
- Required: `provider.options.baseURL`
- Remote MCP servers must have `url`
- Local MCP servers must have `command`

**Crush:**
- Required: `providers` array
- At least one provider must be configured

## Configuration Examples

### Minimal Configuration

```json
{
  "provider": {
    "base_url": "http://localhost:7061/v1"
  }
}
```

### Full Configuration with All Options

```json
{
  "version": "1.0",
  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "http://localhost:7061/v1",
    "api_key_env": "HELIXAGENT_API_KEY"
  },
  "models": [
    {
      "id": "helixagent-debate",
      "name": "HelixAgent AI Debate Ensemble",
      "max_tokens": 128000,
      "capabilities": ["vision", "streaming", "function_calls", "embeddings", "mcp", "acp", "lsp"]
    }
  ],
  "mcp": {
    "helixagent-mcp": {"type": "remote", "url": "http://localhost:7061/v1/mcp"},
    "helixagent-acp": {"type": "remote", "url": "http://localhost:7061/v1/acp"},
    "helixagent-lsp": {"type": "remote", "url": "http://localhost:7061/v1/lsp"},
    "helixagent-embeddings": {"type": "remote", "url": "http://localhost:7061/v1/embeddings"},
    "helixagent-vision": {"type": "remote", "url": "http://localhost:7061/v1/vision"},
    "helixagent-cognee": {"type": "remote", "url": "http://localhost:7061/v1/cognee"},
    "filesystem": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"]},
    "github": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-github"]},
    "memory": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-memory"]},
    "fetch": {"type": "local", "command": ["npx", "-y", "mcp-fetch"]},
    "puppeteer": {"type": "local", "command": ["npx", "-y", "@modelcontextprotocol/server-puppeteer"]},
    "sqlite": {"type": "local", "command": ["npx", "-y", "mcp-server-sqlite"]}
  },
  "settings": {
    "streaming": true,
    "autoApprove": false
  }
}
```

### Remote HelixAgent Configuration

For connecting to a remote HelixAgent instance:

```json
{
  "provider": {
    "type": "openai-compatible",
    "name": "helixagent",
    "base_url": "https://helix.example.com/v1",
    "api_key_env": "HELIXAGENT_API_KEY"
  },
  "mcp": {
    "helixagent-mcp": {"type": "remote", "url": "https://helix.example.com/v1/mcp"}
  }
}
```

## Troubleshooting Configuration

### Common Issues

1. **"provider.base_url is required"**
   - Ensure the `provider` section has a `base_url` field

2. **"mcp.X: type is required"**
   - Each MCP server needs a `type` field ("local" or "remote")

3. **"mcp.X: url is required for remote type"**
   - Remote MCP servers need a `url` field

4. **"Invalid JSON syntax"**
   - Check for trailing commas, missing quotes, etc.

### Debug Mode

Enable debug logging:

```bash
GIN_MODE=debug ./bin/helixagent
```
