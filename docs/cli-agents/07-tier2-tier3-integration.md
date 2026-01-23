# Tier 2-3 Agent Integration

Guide for integrating HelixAgent with Tier 2 and Tier 3 CLI agents using the generic MCP server approach.

## Overview

For agents without rich plugin systems (44 of 48 agents), HelixAgent provides:

1. **Configuration Generation** - Auto-generated configs for each agent
2. **Generic MCP Server** - Universal MCP server for tool integration
3. **OpenAI-Compatible Endpoint** - Drop-in replacement for most agents

## Tier Classification

### Tier 2 - Moderate Extensibility (8 agents)

| Agent | Language | Extension Mechanism | Integration Strategy |
|-------|----------|---------------------|---------------------|
| **Aider** | Python | Coder classes, litellm | Custom model provider |
| **Codename Goose** | TypeScript/Rust | MCP transport | MCP server |
| **Forge** | TypeScript | Templates | Config + MCP |
| **Amazon Q** | TypeScript | AWS extensions | Config only |
| **Kiro** | TypeScript | AWS integration | Config + MCP |
| **GPT Engineer** | Python | TOML config | Config only |
| **Gemini CLI** | Go | Config files | Config only |
| **DeepSeek CLI** | Python | Config files | Config only |

### Tier 3 - Limited Extensibility (36 agents)

Config-only integration using OpenAI-compatible endpoint.

---

## Generic MCP Server

### Installation

```bash
# Install globally
npm install -g @helixagent/mcp-server

# Or run with npx
npx @helixagent/mcp-server
```

### Server Structure

```
@helixagent/mcp-server/
├── package.json
├── src/
│   ├── index.ts              # MCP server entry point
│   ├── tools/
│   │   ├── chat.ts           # Chat completion tool
│   │   ├── debate.ts         # AI debate tool
│   │   ├── embeddings.ts     # Embeddings tool
│   │   ├── vision.ts         # Vision analysis tool
│   │   └── tasks.ts          # Background task tool
│   ├── transport/
│   │   ├── http_client.ts    # HTTP/3 with fallback
│   │   └── toon_codec.ts     # TOON encoding
│   └── events/
│       └── sse_listener.ts   # Event streaming
└── bin/
    └── helixagent-mcp        # CLI entry point
```

### Usage

```bash
# Start MCP server
helixagent-mcp --endpoint http://localhost:7061

# With custom port
helixagent-mcp --endpoint http://localhost:7061 --port 7062

# Enable debug logging
helixagent-mcp --endpoint http://localhost:7061 --debug
```

### Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `helix_chat` | Chat with AI Debate Ensemble | `message`, `enableDebate` |
| `helix_debate` | Start multi-agent debate | `topic`, `positions`, `rounds` |
| `helix_embeddings` | Generate embeddings | `texts`, `model` |
| `helix_vision` | Analyze images | `image`, `prompt` |
| `helix_task` | Create background task | `command`, `timeout` |
| `helix_task_status` | Check task status | `taskId` |

---

## Tier 2 Integration Guides

### Aider Integration

Aider uses litellm for model routing. Configure a custom provider.

**~/.aider.conf.yml:**
```yaml
model: helixagent/helix-debate-ensemble
openai-api-base: http://localhost:7061/v1
openai-api-key: ${HELIXAGENT_API_KEY}

auto-commits: true
edit-format: diff
stream: true
show-diffs: true

# Enable code interpreter
code-theme: monokai
pretty: true
```

**Environment:**
```bash
export HELIXAGENT_API_KEY="your-key"
export AIDER_MODEL="openai/helixagent-debate"
```

**Usage:**
```bash
aider --model helixagent/helix-debate-ensemble
```

### Codename Goose Integration

Goose supports MCP servers natively.

**~/.config/goose/config.yaml:**
```yaml
mcp:
  servers:
    helixagent:
      command: ["npx", "@helixagent/mcp-server"]
      env:
        HELIXAGENT_ENDPOINT: "http://localhost:7061"

provider:
  name: openai
  base_url: http://localhost:7061/v1
  model: helixagent-debate
```

### Forge Integration

**~/.forge/config.json:**
```json
{
  "provider": {
    "type": "openai-compatible",
    "baseUrl": "http://localhost:7061/v1",
    "model": "helixagent-debate"
  },
  "mcp": {
    "helixagent": {
      "command": ["npx", "@helixagent/mcp-server"],
      "env": {
        "HELIXAGENT_ENDPOINT": "http://localhost:7061"
      }
    }
  }
}
```

### GPT Engineer Integration

**gpt-engineer.toml:**
```toml
[model]
provider = "openai"
model = "helixagent-debate"
api_base = "http://localhost:7061/v1"

[settings]
streaming = true
```

### Gemini CLI Integration

**~/.config/gemini-cli/config.json:**
```json
{
  "provider": {
    "type": "openai-compatible",
    "baseUrl": "http://localhost:7061/v1",
    "apiKey": "${HELIXAGENT_API_KEY}",
    "model": "helixagent-debate"
  }
}
```

### DeepSeek CLI Integration

**~/.config/deepseek/config.json:**
```json
{
  "provider": {
    "type": "openai-compatible",
    "baseUrl": "http://localhost:7061/v1",
    "model": "helixagent-debate"
  }
}
```

---

## Tier 3 Integration (Config Only)

For Tier 3 agents, generate configuration and point to HelixAgent's OpenAI-compatible endpoint.

### Generate Configuration

```bash
# Generate for any agent
./bin/helixagent --generate-agent-config=agentdeck --agent-config-output=config.json

# Generate all 48 configs at once
./bin/helixagent --generate-all-agents --all-agents-output-dir=~/helix-configs/
```

### Installation Per Agent

| Agent | Config Location | Command |
|-------|-----------------|---------|
| Agent-Deck | `~/.config/agent-deck/config.json` | `cp config.json ~/.config/agent-deck/` |
| Bridle | `~/.config/bridle/config.yaml` | `cp config.yaml ~/.config/bridle/` |
| Codai | `~/.config/codai/config.json` | `cp config.json ~/.config/codai/` |
| Conduit | `~/.config/conduit/config.json` | `cp config.json ~/.config/conduit/` |
| Emdash | `~/.config/emdash/config.json` | `cp config.json ~/.config/emdash/` |
| FauxPilot | `~/.config/fauxpilot/config.yaml` | `cp config.yaml ~/.config/fauxpilot/` |
| GetShitDone | `~/.config/gsd/config.json` | `cp config.json ~/.config/gsd/` |
| GPTME | `~/.config/gptme/config.toml` | `cp config.toml ~/.config/gptme/` |
| Nanocoder | `~/.config/nanocoder/config.json` | `cp config.json ~/.config/nanocoder/` |
| Noi | `~/.config/noi/config.json` | `cp config.json ~/.config/noi/` |
| Octogen | `~/.config/octogen/config.yaml` | `cp config.yaml ~/.config/octogen/` |
| Shai | `~/.config/shai/config.json` | `cp config.json ~/.config/shai/` |
| SnowCLI | `~/.config/snowcli/config.yaml` | `cp config.yaml ~/.config/snowcli/` |
| VTCode | `~/.config/vtcode/config.json` | `cp config.json ~/.config/vtcode/` |
| Warp | `~/.config/warp/config.yaml` | `cp config.yaml ~/.config/warp/` |

### Common Configuration Pattern

All Tier 3 agents use this basic structure:

```json
{
  "provider": {
    "type": "openai-compatible",
    "baseUrl": "http://localhost:7061/v1",
    "apiKeyEnv": "HELIXAGENT_API_KEY",
    "model": "helixagent-debate"
  },
  "settings": {
    "streaming": true
  }
}
```

---

## MCP Server Implementation

### TypeScript Implementation

```typescript
// src/index.ts
import { Server } from "@modelcontextprotocol/sdk/server";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio";
import { HelixClient } from "./client";

const server = new Server(
  { name: "helixagent-mcp", version: "1.0.0" },
  { capabilities: { tools: {} } }
);

const client = new HelixClient({
  endpoint: process.env.HELIXAGENT_ENDPOINT || "http://localhost:7061"
});

// Register tools
server.setRequestHandler("tools/list", async () => ({
  tools: [
    {
      name: "helix_chat",
      description: "Chat with HelixAgent AI Debate Ensemble",
      inputSchema: {
        type: "object",
        properties: {
          message: { type: "string", description: "Message to send" },
          enableDebate: { type: "boolean", default: true }
        },
        required: ["message"]
      }
    },
    {
      name: "helix_embeddings",
      description: "Generate embeddings",
      inputSchema: {
        type: "object",
        properties: {
          texts: { type: "array", items: { type: "string" } }
        },
        required: ["texts"]
      }
    }
  ]
}));

server.setRequestHandler("tools/call", async (request) => {
  const { name, arguments: args } = request.params;

  switch (name) {
    case "helix_chat":
      const response = await client.chat(args.message, {
        enableDebate: args.enableDebate ?? true
      });
      return { content: [{ type: "text", text: response.content }] };

    case "helix_embeddings":
      const embeddings = await client.embed(args.texts);
      return { content: [{ type: "text", text: JSON.stringify(embeddings) }] };

    default:
      throw new Error(`Unknown tool: ${name}`);
  }
});

// Start server
const transport = new StdioServerTransport();
server.connect(transport);
```

### Go Implementation

```go
// cmd/helixagent-mcp-go/main.go
package main

import (
    "log"
    "os"

    "github.com/mark3labs/mcp-go/server"
)

func main() {
    endpoint := os.Getenv("HELIXAGENT_ENDPOINT")
    if endpoint == "" {
        endpoint = "http://localhost:7061"
    }

    s := server.NewMCPServer(
        "helixagent-mcp",
        "1.0.0",
        server.WithToolCapabilities(true),
    )

    // Register tools
    registerChatTool(s, endpoint)
    registerEmbeddingsTool(s, endpoint)
    registerDebateTool(s, endpoint)

    // Start server on stdio
    if err := server.ServeStdio(s); err != nil {
        log.Fatal(err)
    }
}
```

---

## Adding MCP to Any Agent

For agents that support MCP but don't have pre-built integration:

### Step 1: Check MCP Support

```bash
# Look for MCP configuration options
grep -r "mcp" ~/.config/<agent>/
```

### Step 2: Add MCP Server Configuration

Most agents use one of these formats:

**JSON format:**
```json
{
  "mcpServers": {
    "helixagent": {
      "command": ["npx", "@helixagent/mcp-server"],
      "args": [],
      "env": {
        "HELIXAGENT_ENDPOINT": "http://localhost:7061"
      }
    }
  }
}
```

**YAML format:**
```yaml
mcp:
  servers:
    helixagent:
      command: ["npx", "@helixagent/mcp-server"]
      env:
        HELIXAGENT_ENDPOINT: "http://localhost:7061"
```

### Step 3: Verify Integration

```bash
# Test the MCP server directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx @helixagent/mcp-server
```

---

## Fallback Strategy

For agents that don't support MCP or custom providers:

### Option 1: Environment Variable Override

```bash
export OPENAI_API_BASE="http://localhost:7061/v1"
export OPENAI_API_KEY="$HELIXAGENT_API_KEY"
export OPENAI_MODEL="helixagent-debate"
```

### Option 2: Proxy Configuration

Run a local proxy that intercepts OpenAI API calls:

```bash
# Using mitmproxy or similar
mitmproxy --mode reverse:http://localhost:7061/v1
```

### Option 3: Wrapper Script

Create a wrapper that sets up the environment:

```bash
#!/bin/bash
# helix-<agent>

export OPENAI_API_BASE="http://localhost:7061/v1"
export OPENAI_API_KEY="${HELIXAGENT_API_KEY}"

exec <agent> "$@"
```

---

## Verification

### Test Configuration

```bash
# Validate generated config
./bin/helixagent --validate-agent-config=<agent>:<config-path>
```

### Test Connectivity

```bash
# Test HelixAgent endpoint
curl http://localhost:7061/health

# Test MCP server
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx @helixagent/mcp-server
```

### Test Agent Integration

```bash
# Run agent and verify it uses HelixAgent
<agent> --debug
# Look for requests to localhost:7061
```
