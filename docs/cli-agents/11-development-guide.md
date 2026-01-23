# Plugin Development Guide

Step-by-step guide for developing CLI agent plugins.

## Prerequisites

- Node.js 18+ or Go 1.24+
- HelixAgent running locally (`http://localhost:7061`)
- CLI agent installed and configured

## Development Setup

### 1. Clone Plugin Template

```bash
# TypeScript template
git clone https://github.com/helixagent/plugin-template-ts
cd plugin-template-ts
npm install

# Go template
git clone https://github.com/helixagent/plugin-template-go
cd plugin-template-go
go mod download
```

### 2. Configure Environment

```bash
# Create .env file
cat > .env << EOF
HELIXAGENT_ENDPOINT=http://localhost:7061
HELIXAGENT_API_KEY=your-key
DEBUG=true
EOF
```

### 3. Project Structure

**TypeScript:**
```
plugin-name/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts           # Entry point
│   ├── transport/
│   │   ├── client.ts      # HTTP/3 + TOON client
│   │   └── toon.ts        # TOON codec
│   ├── events/
│   │   ├── sse.ts         # SSE client
│   │   └── websocket.ts   # WebSocket client
│   ├── ui/
│   │   ├── debate.ts      # Debate renderer
│   │   └── progress.ts    # Progress bar
│   └── hooks/             # Agent-specific hooks
│       ├── session_start.ts
│       └── post_tool.ts
├── tests/
│   ├── transport.test.ts
│   ├── events.test.ts
│   └── ui.test.ts
└── README.md
```

**Go:**
```
plugin-name/
├── go.mod
├── go.sum
├── main.go                # Entry point
├── transport/
│   ├── client.go          # HTTP/3 + TOON client
│   └── toon.go            # TOON codec
├── events/
│   ├── sse.go             # SSE client
│   └── websocket.go       # WebSocket client
├── ui/
│   ├── debate.go          # Debate renderer
│   └── progress.go        # Progress bar
├── handlers/
│   └── tools.go           # MCP tool handlers
├── tests/
│   ├── transport_test.go
│   ├── events_test.go
│   └── ui_test.go
└── README.md
```

---

## Implementation Steps

### Step 1: Transport Layer

Create the HTTP client with TOON support.

**TypeScript:**
```typescript
// src/transport/client.ts
import { TOONCodec } from './toon';

export interface ClientOptions {
  endpoint: string;
  apiKey?: string;
  enableTOON?: boolean;
  enableCompression?: boolean;
  timeout?: number;
}

export class HelixClient {
  private endpoint: string;
  private apiKey: string;
  private toon: TOONCodec;
  private enableTOON: boolean;

  constructor(options: ClientOptions) {
    this.endpoint = options.endpoint;
    this.apiKey = options.apiKey || process.env.HELIXAGENT_API_KEY || '';
    this.toon = new TOONCodec({ level: 'standard' });
    this.enableTOON = options.enableTOON ?? true;
  }

  async chat(message: string, options: ChatOptions = {}): Promise<ChatResponse> {
    const body = {
      model: 'helixagent-debate',
      messages: [{ role: 'user', content: message }],
      ...options,
    };

    return this.request('/v1/chat/completions', {
      method: 'POST',
      body,
    });
  }

  async embed(texts: string[]): Promise<EmbeddingResponse> {
    return this.request('/v1/embeddings', {
      method: 'POST',
      body: { input: texts },
    });
  }

  private async request<T>(path: string, options: RequestOptions): Promise<T> {
    const url = `${this.endpoint}${path}`;

    // Prepare body
    let body: string;
    let contentType: string;

    if (this.enableTOON) {
      body = this.toon.encode(options.body);
      contentType = 'application/toon+json';
    } else {
      body = JSON.stringify(options.body);
      contentType = 'application/json';
    }

    // Make request
    const response = await fetch(url, {
      method: options.method || 'POST',
      headers: {
        'Content-Type': contentType,
        'Accept': this.enableTOON ? 'application/toon+json' : 'application/json',
        'Accept-Encoding': 'br, gzip',
        'Authorization': `Bearer ${this.apiKey}`,
      },
      body,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${await response.text()}`);
    }

    // Decode response
    const responseType = response.headers.get('Content-Type');
    const responseBody = await response.text();

    if (responseType?.includes('toon')) {
      return this.toon.decode(responseBody) as T;
    }
    return JSON.parse(responseBody);
  }
}
```

**Go:**
```go
// transport/client.go
package transport

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type ClientOptions struct {
    Endpoint          string
    APIKey            string
    EnableTOON        bool
    EnableCompression bool
    Timeout           time.Duration
}

type HelixClient struct {
    endpoint   string
    apiKey     string
    toon       *TOONCodec
    httpClient *http.Client
    enableTOON bool
}

func NewClient(opts ClientOptions) *HelixClient {
    apiKey := opts.APIKey
    if apiKey == "" {
        apiKey = os.Getenv("HELIXAGENT_API_KEY")
    }

    timeout := opts.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second
    }

    return &HelixClient{
        endpoint:   opts.Endpoint,
        apiKey:     apiKey,
        toon:       NewTOONCodec(LevelStandard),
        httpClient: &http.Client{Timeout: timeout},
        enableTOON: opts.EnableTOON,
    }
}

func (c *HelixClient) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResponse, error) {
    body := map[string]interface{}{
        "model":    "helixagent-debate",
        "messages": []map[string]string{{"role": "user", "content": message}},
    }

    if opts != nil {
        if opts.EnableDebate {
            body["enable_debate"] = true
        }
    }

    var response ChatResponse
    err := c.request(ctx, "/v1/chat/completions", body, &response)
    return &response, err
}

func (c *HelixClient) Embed(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
    body := map[string]interface{}{
        "input": texts,
    }

    var response EmbeddingResponse
    err := c.request(ctx, "/v1/embeddings", body, &response)
    return &response, err
}

func (c *HelixClient) request(ctx context.Context, path string, body interface{}, result interface{}) error {
    url := c.endpoint + path

    // Encode body
    var reqBody []byte
    var contentType string
    var err error

    if c.enableTOON {
        reqBody, err = c.toon.Encode(body)
        contentType = "application/toon+json"
    } else {
        reqBody, err = json.Marshal(body)
        contentType = "application/json"
    }
    if err != nil {
        return fmt.Errorf("encode body: %w", err)
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", contentType)
    req.Header.Set("Accept", contentType)
    req.Header.Set("Accept-Encoding", "br, gzip")
    if c.apiKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.apiKey)
    }

    // Send request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
    }

    // Decode response
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("read response: %w", err)
    }

    respType := resp.Header.Get("Content-Type")
    if c.enableTOON && respType == "application/toon+json" {
        return c.toon.Decode(respBody, result)
    }
    return json.Unmarshal(respBody, result)
}
```

### Step 2: Event Subscription

Add real-time event handling.

**TypeScript:**
```typescript
// src/events/sse.ts
export class EventSubscriber {
  private client: HelixClient;
  private eventSource: EventSource | null = null;
  private handlers: Map<string, ((data: any) => void)[]> = new Map();

  constructor(client: HelixClient) {
    this.client = client;
  }

  subscribe(events: string[]): void {
    const url = `${this.client.endpoint}/v1/events?subscribe=${events.join(',')}`;

    this.eventSource = new EventSource(url);

    this.eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.dispatch(data.type, data);
    };

    this.eventSource.onerror = () => {
      // Reconnect after delay
      setTimeout(() => this.subscribe(events), 5000);
    };
  }

  on(eventType: string, handler: (data: any) => void): void {
    const handlers = this.handlers.get(eventType) || [];
    handlers.push(handler);
    this.handlers.set(eventType, handlers);
  }

  private dispatch(type: string, data: any): void {
    // Exact match handlers
    const handlers = this.handlers.get(type) || [];
    handlers.forEach((h) => h(data));

    // Wildcard handlers
    const wildcardHandlers = this.handlers.get('*') || [];
    wildcardHandlers.forEach((h) => h(data));
  }

  close(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }
}
```

### Step 3: UI Components

Implement debate visualization.

**TypeScript:**
```typescript
// src/ui/debate.ts
export class DebateUI {
  private renderer: DebateRenderer;
  private events: EventSubscriber;

  constructor(client: HelixClient, events: EventSubscriber) {
    this.renderer = new DebateRenderer({ style: 'theater' });
    this.events = events;

    this.setupEventHandlers();
  }

  private setupEventHandlers(): void {
    this.events.on('debate.started', (data) => {
      console.clear();
      console.log(this.renderer.render({
        round: 1,
        totalRounds: data.totalRounds,
        positions: [],
        phase: 'initial',
        votes: { for: 0, total: 15 },
      }));
    });

    this.events.on('debate.position_submitted', (data) => {
      console.clear();
      console.log(this.renderer.render(data));
    });

    this.events.on('debate.consensus', (data) => {
      console.clear();
      console.log(this.renderer.render({
        ...data,
        phase: 'conclusion',
      }));
      console.log('\n✓ Consensus reached!');
    });
  }
}
```

### Step 4: Agent Integration

Create hooks for your target agent.

**Claude Code Hook Example:**
```typescript
// src/hooks/post_tool.ts
import { HelixClient } from '../transport/client';
import { DebateRenderer } from '../ui/debate';

interface PostToolContext {
  toolName: string;
  toolResult: any;
  config: any;
}

export async function postToolUse(context: PostToolContext): Promise<any> {
  // Only process HelixAgent results
  if (!context.toolResult?.debate) {
    return { modified: false };
  }

  // Render debate visualization
  const renderer = new DebateRenderer({
    style: context.config.renderStyle || 'theater',
  });

  const visualization = renderer.render(context.toolResult.debate);

  return {
    modified: true,
    contextModification: visualization,
    notification: {
      title: 'AI Debate Complete',
      body: `Consensus: ${context.toolResult.debate.consensus}`,
    },
  };
}
```

**OpenCode MCP Tool Example:**
```go
// handlers/tools.go
package handlers

import (
    "context"

    "github.com/mark3labs/mcp-go/mcp"
    "plugin-name/transport"
)

func HandleChat(client *transport.HelixClient) mcp.ToolHandler {
    return func(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
        message := args["message"].(string)
        enableDebate := true
        if v, ok := args["enableDebate"].(bool); ok {
            enableDebate = v
        }

        resp, err := client.Chat(ctx, message, &transport.ChatOptions{
            EnableDebate: enableDebate,
        })
        if err != nil {
            return nil, err
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{
                {Type: "text", Text: resp.Content},
            },
        }, nil
    }
}
```

---

## Testing

### Unit Tests

```typescript
// tests/transport.test.ts
import { HelixClient } from '../src/transport/client';
import { TOONCodec } from '../src/transport/toon';

describe('HelixClient', () => {
  let client: HelixClient;

  beforeEach(() => {
    client = new HelixClient({
      endpoint: 'http://localhost:7061',
      enableTOON: true,
    });
  });

  it('should encode requests with TOON', async () => {
    // Mock fetch
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/toon+json' }),
      text: () => Promise.resolve('r=a;c=Hello'),
    });

    const response = await client.chat('Hello');

    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          'Content-Type': 'application/toon+json',
        }),
      })
    );
  });
});

describe('TOONCodec', () => {
  const codec = new TOONCodec({ level: 'standard' });

  it('should encode objects', () => {
    const input = { message: 'hello', role: 'user' };
    const encoded = codec.encode(input);
    expect(encoded).toBe('m=hello;r=u');
  });

  it('should decode TOON strings', () => {
    const input = 'm=hello;r=u';
    const decoded = codec.decode(input);
    expect(decoded).toEqual({ message: 'hello', role: 'user' });
  });
});
```

### Integration Tests

```typescript
// tests/integration.test.ts
import { HelixClient } from '../src/transport/client';
import { EventSubscriber } from '../src/events/sse';

describe('Integration', () => {
  let client: HelixClient;

  beforeAll(() => {
    client = new HelixClient({
      endpoint: process.env.HELIXAGENT_ENDPOINT || 'http://localhost:7061',
    });
  });

  it('should complete chat request', async () => {
    const response = await client.chat('What is 2+2?');
    expect(response.content).toBeDefined();
  });

  it('should receive events', (done) => {
    const events = new EventSubscriber(client);

    events.on('task.created', (data) => {
      expect(data.taskId).toBeDefined();
      events.close();
      done();
    });

    events.subscribe(['task.*']);

    // Trigger a task
    client.chat('Count to 10');
  }, 30000);
});
```

### E2E Tests

```bash
#!/bin/bash
# tests/e2e.sh

set -e

echo "=== E2E Plugin Tests ==="

# Start plugin
npm run start &
PLUGIN_PID=$!
sleep 2

# Test 1: Health check
echo "Test 1: Health check"
curl -s http://localhost:7062/health | grep -q "ok"
echo "✓ Health check passed"

# Test 2: Chat request
echo "Test 2: Chat request"
RESPONSE=$(curl -s -X POST http://localhost:7062/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}')
echo "$RESPONSE" | grep -q "content"
echo "✓ Chat request passed"

# Test 3: MCP tools list
echo "Test 3: MCP tools"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  node dist/index.js | grep -q "helix_chat"
echo "✓ MCP tools passed"

# Cleanup
kill $PLUGIN_PID

echo "=== All tests passed ==="
```

---

## Debugging

### Enable Debug Logging

```typescript
// src/index.ts
import debug from 'debug';

const log = debug('helixagent:plugin');
const logTransport = debug('helixagent:transport');
const logEvents = debug('helixagent:events');

// Usage
logTransport('Request: %O', { url, method, body });
logEvents('Event received: %s', eventType);
```

```bash
# Run with debug logging
DEBUG=helixagent:* npm start
```

### Inspect Network Traffic

```bash
# Use mitmproxy to inspect traffic
mitmproxy --mode reverse:http://localhost:7061

# Or use Charles Proxy / Wireshark
```

### Test TOON Encoding

```bash
# Compare JSON vs TOON sizes
echo '{"message":"hello","role":"user"}' | node -e "
const { TOONCodec } = require('./dist/transport/toon');
const codec = new TOONCodec();
let json = '';
process.stdin.on('data', d => json += d);
process.stdin.on('end', () => {
  const obj = JSON.parse(json);
  const toon = codec.encode(obj);
  console.log('JSON:', json.length, 'bytes');
  console.log('TOON:', toon.length, 'bytes');
  console.log('Savings:', Math.round((1 - toon.length/json.length) * 100) + '%');
});
"
```

---

## Deployment

### Package for Distribution

```bash
# TypeScript
npm run build
npm pack

# Go
go build -o plugin-name ./...
```

### Installation Instructions

**Claude Code:**
```bash
# Copy to plugins directory
cp -r dist/ ~/.claude/plugins/helixagent-integration/

# Or install from npm
npm install -g @helixagent/claude-code-plugin
```

**OpenCode:**
```bash
# Add to opencode.json
{
  "mcpServers": {
    "helixagent": {
      "command": ["helixagent-mcp-go"],
      "env": {
        "HELIXAGENT_ENDPOINT": "http://localhost:7061"
      }
    }
  }
}
```

### Version Management

Follow semantic versioning:
- **MAJOR**: Breaking API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes

```json
{
  "name": "@helixagent/plugin-name",
  "version": "1.2.3"
}
```

---

## Best Practices

### Error Handling

```typescript
// Always wrap external calls in try/catch
try {
  const response = await client.chat(message);
  return response;
} catch (error) {
  if (error instanceof NetworkError) {
    // Retry with exponential backoff
    return retry(() => client.chat(message), { maxAttempts: 3 });
  }
  throw error;
}
```

### Resource Cleanup

```typescript
// Implement cleanup handlers
process.on('SIGINT', async () => {
  await events.close();
  await client.close();
  process.exit(0);
});
```

### Configuration Validation

```typescript
// Validate required config
function validateConfig(config: PluginConfig): void {
  if (!config.endpoint) {
    throw new Error('Missing required config: endpoint');
  }
  if (!config.endpoint.startsWith('http')) {
    throw new Error('Invalid endpoint URL');
  }
}
```

### Performance Optimization

```typescript
// Use connection pooling
const agent = new http.Agent({
  keepAlive: true,
  maxSockets: 10,
});

// Cache frequently used data
const cache = new Map<string, CacheEntry>();

// Batch requests when possible
const embeddings = await client.embedBatch(texts, { batchSize: 100 });
```
