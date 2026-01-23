# Tier 1 Plugin Development

Complete guide for developing plugins for Tier 1 CLI agents: Claude Code, OpenCode, Cline, and Kilo-Code.

## Claude Code Plugin

Claude Code supports plugins through its marketplace and hook system.

### Plugin Structure

```
~/.claude/plugins/helixagent-integration/
├── .claude-plugin/
│   └── plugin.json
├── hooks/
│   ├── hooks.json
│   ├── session_start.js
│   ├── session_end.js
│   ├── pre_tool.js
│   └── post_tool.js
├── lib/
│   ├── transport.js
│   ├── events.js
│   ├── debate_renderer.js
│   └── progress_renderer.js
└── MANIFEST.md
```

### Plugin Manifest (plugin.json)

```json
{
  "name": "helixagent-integration",
  "version": "1.0.0",
  "description": "HelixAgent integration with HTTP/3, TOON, and AI debate visualization",
  "author": "HelixAgent",
  "hooks": ["SessionStart", "SessionEnd", "PreToolUse", "PostToolUse"],
  "permissions": ["network", "filesystem"],
  "config": {
    "endpoint": "http://localhost:7061",
    "enableDebateVisualization": true,
    "enableTOON": true
  }
}
```

### Hook Definitions (hooks.json)

```json
{
  "hooks": [
    {
      "name": "SessionStart",
      "script": "./hooks/session_start.js",
      "description": "Initialize HelixAgent connection"
    },
    {
      "name": "SessionEnd",
      "script": "./hooks/session_end.js",
      "description": "Cleanup HelixAgent session"
    },
    {
      "name": "PreToolUse",
      "script": "./hooks/pre_tool.js",
      "description": "Transform requests for TOON encoding"
    },
    {
      "name": "PostToolUse",
      "script": "./hooks/post_tool.js",
      "description": "Render debate results"
    }
  ]
}
```

### Session Start Hook

```javascript
// hooks/session_start.js
const { HelixTransport } = require('../lib/transport');
const { EventClient } = require('../lib/events');

module.exports = async function sessionStart(context) {
  // Initialize transport
  const transport = new HelixTransport({
    endpoint: context.config.endpoint,
    enableTOON: context.config.enableTOON,
    enableBrotli: true
  });

  await transport.connect();

  // Initialize event subscription
  const events = new EventClient(transport);
  await events.subscribe(['debate.*', 'task.*']);

  // Store in context for other hooks
  context.helix = { transport, events };

  return {
    success: true,
    message: 'HelixAgent connected'
  };
};
```

### Pre-Tool Hook (TOON Encoding)

```javascript
// hooks/pre_tool.js
const { TOONEncoder } = require('../lib/transport');

module.exports = async function preToolUse(context, toolUse) {
  // Only transform helix_* tools
  if (!toolUse.toolName.startsWith('helix_')) {
    return { modified: false };
  }

  // Encode parameters with TOON
  const encoder = new TOONEncoder();
  const encoded = encoder.encode(toolUse.parameters);

  return {
    modified: true,
    toolUse: {
      ...toolUse,
      parameters: encoded,
      headers: {
        'Content-Type': 'application/toon+json',
        'Accept-Encoding': 'br'
      }
    }
  };
};
```

### Post-Tool Hook (Debate Rendering)

```javascript
// hooks/post_tool.js
const { DebateRenderer } = require('../lib/debate_renderer');

module.exports = async function postToolUse(context, toolResult) {
  // Check if this is a debate result
  if (!toolResult.debate) {
    return { modified: false };
  }

  // Render debate visualization
  const renderer = new DebateRenderer({
    style: context.config.debateStyle || 'theater'
  });

  const visualization = renderer.render(toolResult.debate);

  return {
    modified: true,
    contextModification: visualization,
    notification: {
      title: 'AI Debate Complete',
      body: `Consensus reached with ${toolResult.debate.confidence * 100}% confidence`
    }
  };
};
```

### Transport Library

```javascript
// lib/transport.js
const { createQuicConnection } = require('@helixagent/quic');

class HelixTransport {
  constructor(options) {
    this.endpoint = options.endpoint;
    this.enableTOON = options.enableTOON;
    this.enableBrotli = options.enableBrotli;
    this.connection = null;
  }

  async connect() {
    try {
      // Try HTTP/3 first
      this.connection = await createQuicConnection(this.endpoint);
      this.protocol = 'h3';
    } catch {
      // Fallback to HTTP/2
      this.connection = await createHttp2Connection(this.endpoint);
      this.protocol = 'h2';
    }
  }

  async request(path, options) {
    const headers = {
      'Accept': this.enableTOON ? 'application/toon+json' : 'application/json',
      'Accept-Encoding': this.enableBrotli ? 'br, gzip' : 'gzip',
      ...options.headers
    };

    const response = await this.connection.request({
      path,
      method: options.method || 'POST',
      headers,
      body: options.body
    });

    return this.decodeResponse(response);
  }

  decodeResponse(response) {
    // Handle TOON decoding
    if (response.headers['content-type']?.includes('toon')) {
      return TOONDecoder.decode(response.body);
    }
    return JSON.parse(response.body);
  }
}

module.exports = { HelixTransport };
```

---

## OpenCode MCP Server

OpenCode uses MCP servers for extensibility. HelixAgent provides a native Go MCP server.

### MCP Server Structure

```
cmd/helixagent-mcp-go/
├── main.go
├── handler.go
├── transport.go
├── tools.go
└── ui.go
```

### Main Entry Point

```go
// main.go
package main

import (
    "log"
    "os"

    "github.com/mark3labs/mcp-go/server"
)

func main() {
    // Create MCP server
    s := server.NewMCPServer(
        "helixagent-mcp",
        "1.0.0",
        server.WithToolCapabilities(true),
        server.WithResourceCapabilities(true),
    )

    // Register tools
    registerTools(s)

    // Start server
    if err := server.ServeStdio(s); err != nil {
        log.Fatal(err)
    }
}
```

### Tool Registration

```go
// tools.go
package main

import (
    "context"

    "github.com/mark3labs/mcp-go/mcp"
)

func registerTools(s *server.MCPServer) {
    // Chat completion tool
    s.AddTool(mcp.Tool{
        Name:        "helix_chat",
        Description: "Send a message to HelixAgent AI Debate Ensemble",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]mcp.Property{
                "message": {
                    Type:        "string",
                    Description: "The message to send",
                },
                "enableDebate": {
                    Type:        "boolean",
                    Description: "Enable AI Debate for this request",
                    Default:     true,
                },
            },
            Required: []string{"message"},
        },
    }, handleChat)

    // Embeddings tool
    s.AddTool(mcp.Tool{
        Name:        "helix_embeddings",
        Description: "Generate embeddings using HelixAgent",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]mcp.Property{
                "texts": {
                    Type:        "array",
                    Items:       &mcp.Property{Type: "string"},
                    Description: "Texts to embed",
                },
            },
            Required: []string{"texts"},
        },
    }, handleEmbeddings)

    // Vision tool
    s.AddTool(mcp.Tool{
        Name:        "helix_vision",
        Description: "Analyze images using HelixAgent",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]mcp.Property{
                "image": {
                    Type:        "string",
                    Description: "Base64 encoded image or URL",
                },
                "prompt": {
                    Type:        "string",
                    Description: "Analysis prompt",
                },
            },
            Required: []string{"image"},
        },
    }, handleVision)
}
```

### Tool Handler

```go
// handler.go
package main

import (
    "context"
    "encoding/json"

    "github.com/mark3labs/mcp-go/mcp"
)

func handleChat(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
    message := args["message"].(string)
    enableDebate := true
    if v, ok := args["enableDebate"].(bool); ok {
        enableDebate = v
    }

    // Create HelixAgent request
    req := &ChatRequest{
        Messages: []Message{
            {Role: "user", Content: message},
        },
        Model:        "helixagent-debate",
        EnableDebate: enableDebate,
    }

    // Send request using transport
    transport := GetTransport()
    resp, err := transport.Chat(ctx, req)
    if err != nil {
        return nil, err
    }

    // Format response with debate info
    result := formatChatResponse(resp)

    return &mcp.CallToolResult{
        Content: []mcp.Content{
            {
                Type: "text",
                Text: result,
            },
        },
    }, nil
}
```

### OpenCode Configuration

```json
{
  "mcpServers": {
    "helixagent": {
      "type": "stdio",
      "command": ["helixagent-mcp-go"],
      "env": {
        "HELIXAGENT_ENDPOINT": "http://localhost:7061"
      }
    }
  }
}
```

---

## Cline Extension

Cline provides 8 lifecycle hooks for deep integration.

### Hook Implementation

```javascript
// .clinerules/hooks/pre_tool_use.js

/**
 * PreToolUse hook - intercepts tool calls before execution
 *
 * Input (stdin):
 * {
 *   "hookName": "PreToolUse",
 *   "taskId": "...",
 *   "preToolUse": {
 *     "toolName": "helix_debate",
 *     "parameters": {...}
 *   }
 * }
 *
 * Output (stdout):
 * {
 *   "cancel": false,
 *   "modifiedParameters": {...},
 *   "contextModification": "..."
 * }
 */

const readline = require('readline');
const { HelixClient } = require('./lib/helix_client');

async function main() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false
  });

  for await (const line of rl) {
    const input = JSON.parse(line);

    if (input.preToolUse?.toolName?.startsWith('helix_')) {
      const client = new HelixClient();

      // Transform parameters for HelixAgent
      const transformed = await client.transformRequest(input.preToolUse);

      console.log(JSON.stringify({
        cancel: false,
        modifiedParameters: transformed.parameters,
        contextModification: transformed.context
      }));
    } else {
      console.log(JSON.stringify({ cancel: false }));
    }
  }
}

main();
```

### All 8 Hooks

```javascript
// Task lifecycle
// hooks/task_start.js
module.exports = async (input) => {
  // Initialize HelixAgent session
  const session = await helix.createSession();
  return { sessionId: session.id };
};

// hooks/task_resume.js
module.exports = async (input) => {
  // Resume existing session
  await helix.resumeSession(input.sessionId);
  return { restored: true };
};

// hooks/task_cancel.js
module.exports = async (input) => {
  // Cancel active requests
  await helix.cancelSession(input.sessionId);
  return { cancelled: true };
};

// hooks/task_complete.js
module.exports = async (input) => {
  // Render final summary
  const summary = await helix.getSessionSummary(input.sessionId);
  return { contextModification: formatSummary(summary) };
};

// User interaction
// hooks/user_prompt_submit.js
module.exports = async (input) => {
  // Transform user prompt for TOON
  const encoded = toon.encode(input.prompt);
  return { modifiedPrompt: encoded };
};

// Tool lifecycle
// hooks/pre_tool_use.js - shown above

// hooks/post_tool_use.js
module.exports = async (input) => {
  // Render debate results
  if (input.toolResult?.debate) {
    const rendered = debateRenderer.render(input.toolResult.debate);
    return { contextModification: rendered };
  }
  return {};
};

// Context management
// hooks/pre_compact.js
module.exports = async (input) => {
  // Save context before compaction
  await helix.saveContext(input.sessionId, input.context);
  return {};
};
```

---

## Kilo-Code Packages

Kilo-Code uses a monorepo with platform-specific packages.

### Package Structure

```
packages/helixagent/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts
│   ├── transport/
│   │   ├── index.ts
│   │   ├── quic.ts
│   │   ├── toon.ts
│   │   └── brotli.ts
│   ├── events/
│   │   ├── index.ts
│   │   ├── sse.ts
│   │   └── websocket.ts
│   └── ui/
│       ├── index.ts
│       ├── debate.ts
│       └── progress.ts
├── cli/
│   └── index.ts
├── vscode/
│   └── extension.ts
└── jetbrains/
    └── plugin.kt
```

### Core Package

```typescript
// src/index.ts
export { HelixTransport } from './transport';
export { EventClient } from './events';
export { DebateRenderer, ProgressBar } from './ui';
export { HelixClient } from './client';

// Main client class
export class HelixClient {
  private transport: HelixTransport;
  private events: EventClient;
  private ui: DebateRenderer;

  constructor(options: HelixOptions) {
    this.transport = new HelixTransport(options);
    this.events = new EventClient(this.transport);
    this.ui = new DebateRenderer(options.uiOptions);
  }

  async connect(): Promise<void> {
    await this.transport.connect();
    await this.events.subscribe(['debate.*', 'task.*']);
  }

  async chat(message: string, options?: ChatOptions): Promise<ChatResponse> {
    const response = await this.transport.request('/v1/chat/completions', {
      method: 'POST',
      body: {
        model: 'helixagent-debate',
        messages: [{ role: 'user', content: message }],
        ...options
      }
    });

    // Render debate if available
    if (response.debate) {
      this.ui.render(response.debate);
    }

    return response;
  }

  async embed(texts: string[]): Promise<EmbeddingResponse> {
    return this.transport.request('/v1/embeddings', {
      method: 'POST',
      body: { input: texts }
    });
  }

  disconnect(): void {
    this.events.unsubscribe();
    this.transport.close();
  }
}
```

### VS Code Extension

```typescript
// vscode/extension.ts
import * as vscode from 'vscode';
import { HelixClient } from '@helixagent/client';

export function activate(context: vscode.ExtensionContext) {
  const client = new HelixClient({
    endpoint: vscode.workspace.getConfiguration('helixagent').get('endpoint'),
  });

  // Register chat command
  const chatCommand = vscode.commands.registerCommand(
    'helixagent.chat',
    async () => {
      const input = await vscode.window.showInputBox({
        prompt: 'Ask HelixAgent AI Debate Ensemble'
      });

      if (input) {
        const response = await client.chat(input);

        // Show in output channel
        const channel = vscode.window.createOutputChannel('HelixAgent');
        channel.appendLine(response.content);
        channel.show();
      }
    }
  );

  context.subscriptions.push(chatCommand);
}
```

### CLI Integration

```typescript
// cli/index.ts
import { Command } from 'commander';
import { HelixClient } from '../src';

const program = new Command();

program
  .command('chat <message>')
  .description('Chat with HelixAgent')
  .option('--no-debate', 'Disable AI debate')
  .action(async (message, options) => {
    const client = new HelixClient({
      endpoint: process.env.HELIXAGENT_ENDPOINT || 'http://localhost:7061'
    });

    await client.connect();

    const response = await client.chat(message, {
      enableDebate: options.debate
    });

    console.log(response.content);

    client.disconnect();
  });

program.parse();
```
