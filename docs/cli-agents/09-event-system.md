# Event System

Real-time event streaming documentation for CLI agent plugins.

## Overview

HelixAgent provides three event transport mechanisms:

| Transport | Use Case | Latency | Reliability |
|-----------|----------|---------|-------------|
| **SSE** | Default for web clients | Low | Auto-reconnect |
| **WebSocket** | Bidirectional communication | Lowest | Persistent |
| **Webhooks** | Server-to-server | Variable | At-least-once |

## Event Types

### Task Events (14 types)

| Event | Description | Payload |
|-------|-------------|---------|
| `task.created` | Task created | `{taskId, command, priority}` |
| `task.started` | Execution started | `{taskId, workerId, startTime}` |
| `task.progress` | Progress update | `{taskId, percent, message}` |
| `task.heartbeat` | Keepalive signal | `{taskId, timestamp}` |
| `task.paused` | Task paused | `{taskId, reason}` |
| `task.resumed` | Task resumed | `{taskId}` |
| `task.completed` | Success | `{taskId, result, duration}` |
| `task.failed` | Failure | `{taskId, error, retryCount}` |
| `task.stuck` | Stuck detected | `{taskId, lastActivity, threshold}` |
| `task.cancelled` | Cancelled | `{taskId, reason}` |
| `task.retrying` | Retry attempt | `{taskId, attempt, maxRetries}` |
| `task.deadletter` | DLQ'd | `{taskId, error, attempts}` |
| `task.log` | Log output | `{taskId, level, message}` |
| `task.resource` | Resource usage | `{taskId, cpu, memory, io}` |

### Debate Events (8 types)

| Event | Description | Payload |
|-------|-------------|---------|
| `debate.started` | Debate initiated | `{debateId, topic, participants}` |
| `debate.round_started` | New round | `{debateId, round, totalRounds}` |
| `debate.position_submitted` | Position from LLM | `{debateId, position, participant, confidence}` |
| `debate.validation_phase` | Entering validation | `{debateId, round, positions}` |
| `debate.polish_phase` | Entering polish | `{debateId, round, validatedPositions}` |
| `debate.consensus` | Consensus reached | `{debateId, consensus, confidence, votes}` |
| `debate.completed` | Debate finished | `{debateId, result, duration, rounds}` |
| `debate.failed` | Debate failed | `{debateId, error, lastRound}` |

---

## SSE (Server-Sent Events)

### Overview

SSE provides unidirectional real-time streaming over HTTP with automatic reconnection.

### Endpoint

```
GET /v1/tasks/:taskId/events
GET /v1/debates/:debateId/events
```

### Headers

```
Accept: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

### Wire Format

```
event: task.progress
id: 12345
data: {"taskId":"abc","percent":45,"message":"Processing..."}

event: task.completed
id: 12346
data: {"taskId":"abc","result":"success","duration":5200}
```

### Client Implementation (TypeScript)

```typescript
// packages/events/src/sse_client.ts

export interface SSEOptions {
  endpoint: string;
  taskId?: string;
  debateId?: string;
  reconnectInterval?: number;
  maxRetries?: number;
}

export class SSEClient {
  private eventSource: EventSource | null = null;
  private endpoint: string;
  private reconnectInterval: number;
  private retries = 0;
  private maxRetries: number;
  private lastEventId: string | null = null;
  private handlers: Map<string, Set<(data: any) => void>> = new Map();

  constructor(options: SSEOptions) {
    this.endpoint = options.endpoint;
    this.reconnectInterval = options.reconnectInterval ?? 5000;
    this.maxRetries = options.maxRetries ?? 10;
  }

  connect(): void {
    const url = new URL(this.endpoint);
    if (this.lastEventId) {
      url.searchParams.set('lastEventId', this.lastEventId);
    }

    this.eventSource = new EventSource(url.toString());

    this.eventSource.onopen = () => {
      this.retries = 0;
      this.emit('connected', {});
    };

    this.eventSource.onerror = (error) => {
      this.handleError(error);
    };

    this.eventSource.onmessage = (event) => {
      this.lastEventId = event.lastEventId;
      this.handleMessage(event);
    };

    // Register specific event types
    const eventTypes = [
      'task.created', 'task.started', 'task.progress', 'task.completed',
      'task.failed', 'debate.started', 'debate.consensus', 'debate.completed'
    ];

    for (const type of eventTypes) {
      this.eventSource.addEventListener(type, (event: MessageEvent) => {
        this.lastEventId = event.lastEventId;
        this.emit(type, JSON.parse(event.data));
      });
    }
  }

  on(eventType: string, handler: (data: any) => void): void {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, new Set());
    }
    this.handlers.get(eventType)!.add(handler);
  }

  off(eventType: string, handler: (data: any) => void): void {
    this.handlers.get(eventType)?.delete(handler);
  }

  private emit(eventType: string, data: any): void {
    const handlers = this.handlers.get(eventType);
    if (handlers) {
      for (const handler of handlers) {
        handler(data);
      }
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const data = JSON.parse(event.data);
      this.emit('message', data);
    } catch (error) {
      console.error('Failed to parse SSE message:', error);
    }
  }

  private handleError(error: Event): void {
    this.emit('error', error);

    if (this.retries < this.maxRetries) {
      this.retries++;
      setTimeout(() => this.reconnect(), this.reconnectInterval);
    } else {
      this.emit('maxRetriesReached', { retries: this.retries });
    }
  }

  private reconnect(): void {
    this.close();
    this.connect();
  }

  close(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }
}
```

### Client Implementation (Go)

```go
// packages/events/sse_client.go
package events

import (
    "bufio"
    "context"
    "encoding/json"
    "net/http"
    "strings"
    "time"
)

type SSEClient struct {
    endpoint          string
    httpClient        *http.Client
    reconnectInterval time.Duration
    maxRetries        int
    lastEventID       string
    handlers          map[string][]func(interface{})
}

type SSEEvent struct {
    Type string
    ID   string
    Data json.RawMessage
}

func NewSSEClient(endpoint string, opts ...SSEOption) *SSEClient {
    client := &SSEClient{
        endpoint:          endpoint,
        httpClient:        &http.Client{Timeout: 0}, // No timeout for SSE
        reconnectInterval: 5 * time.Second,
        maxRetries:        10,
        handlers:          make(map[string][]func(interface{})),
    }

    for _, opt := range opts {
        opt(client)
    }

    return client
}

func (c *SSEClient) Connect(ctx context.Context) (<-chan SSEEvent, error) {
    events := make(chan SSEEvent, 100)

    go func() {
        defer close(events)

        retries := 0
        for {
            err := c.stream(ctx, events)
            if err == nil || ctx.Err() != nil {
                return
            }

            retries++
            if retries > c.maxRetries {
                return
            }

            select {
            case <-ctx.Done():
                return
            case <-time.After(c.reconnectInterval):
                // Retry
            }
        }
    }()

    return events, nil
}

func (c *SSEClient) stream(ctx context.Context, events chan<- SSEEvent) error {
    req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Accept", "text/event-stream")
    req.Header.Set("Cache-Control", "no-cache")
    if c.lastEventID != "" {
        req.Header.Set("Last-Event-ID", c.lastEventID)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    var event SSEEvent

    for scanner.Scan() {
        line := scanner.Text()

        if line == "" {
            // End of event
            if event.Type != "" {
                c.lastEventID = event.ID
                events <- event
                event = SSEEvent{}
            }
            continue
        }

        if strings.HasPrefix(line, "event:") {
            event.Type = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
        } else if strings.HasPrefix(line, "id:") {
            event.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
        } else if strings.HasPrefix(line, "data:") {
            event.Data = json.RawMessage(strings.TrimPrefix(line, "data:"))
        }
    }

    return scanner.Err()
}

func (c *SSEClient) On(eventType string, handler func(interface{})) {
    c.handlers[eventType] = append(c.handlers[eventType], handler)
}
```

---

## WebSocket

### Overview

WebSocket provides full-duplex communication with the lowest latency.

### Endpoint

```
GET /v1/ws/tasks/:taskId
GET /v1/ws/debates/:debateId
```

### Wire Format

```json
// Client → Server (subscription)
{"type": "subscribe", "events": ["task.*", "debate.consensus"]}

// Server → Client (event)
{"type": "task.progress", "taskId": "abc", "data": {"percent": 45}}

// Client → Server (ping)
{"type": "ping"}

// Server → Client (pong)
{"type": "pong", "timestamp": 1705312200000}
```

### Client Implementation (TypeScript)

```typescript
// packages/events/src/websocket_client.ts

export interface WebSocketOptions {
  endpoint: string;
  reconnectInterval?: number;
  maxRetries?: number;
  pingInterval?: number;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private endpoint: string;
  private reconnectInterval: number;
  private pingInterval: number;
  private pingTimer: NodeJS.Timer | null = null;
  private retries = 0;
  private maxRetries: number;
  private handlers: Map<string, Set<(data: any) => void>> = new Map();
  private subscriptions: string[] = [];

  constructor(options: WebSocketOptions) {
    this.endpoint = options.endpoint;
    this.reconnectInterval = options.reconnectInterval ?? 5000;
    this.pingInterval = options.pingInterval ?? 30000;
    this.maxRetries = options.maxRetries ?? 10;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.endpoint);

      this.ws.onopen = () => {
        this.retries = 0;
        this.startPing();
        this.resubscribe();
        this.emit('connected', {});
        resolve();
      };

      this.ws.onerror = (error) => {
        this.emit('error', error);
        reject(error);
      };

      this.ws.onclose = (event) => {
        this.stopPing();
        this.emit('disconnected', { code: event.code, reason: event.reason });
        this.attemptReconnect();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event);
      };
    });
  }

  subscribe(events: string[]): void {
    this.subscriptions = [...new Set([...this.subscriptions, ...events])];
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.send({ type: 'subscribe', events: this.subscriptions });
    }
  }

  unsubscribe(events: string[]): void {
    this.subscriptions = this.subscriptions.filter((e) => !events.includes(e));
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.send({ type: 'unsubscribe', events });
    }
  }

  on(eventType: string, handler: (data: any) => void): void {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, new Set());
    }
    this.handlers.get(eventType)!.add(handler);
  }

  off(eventType: string, handler: (data: any) => void): void {
    this.handlers.get(eventType)?.delete(handler);
  }

  private emit(eventType: string, data: any): void {
    // Emit specific event
    const handlers = this.handlers.get(eventType);
    if (handlers) {
      for (const handler of handlers) {
        handler(data);
      }
    }

    // Emit wildcard handlers
    const wildcardHandlers = this.handlers.get('*');
    if (wildcardHandlers) {
      for (const handler of wildcardHandlers) {
        handler({ type: eventType, ...data });
      }
    }
  }

  private send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message = JSON.parse(event.data);

      if (message.type === 'pong') {
        return; // Ignore pong responses
      }

      this.emit(message.type, message.data || message);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private startPing(): void {
    this.pingTimer = setInterval(() => {
      this.send({ type: 'ping' });
    }, this.pingInterval);
  }

  private stopPing(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  private resubscribe(): void {
    if (this.subscriptions.length > 0) {
      this.send({ type: 'subscribe', events: this.subscriptions });
    }
  }

  private attemptReconnect(): void {
    if (this.retries < this.maxRetries) {
      this.retries++;
      setTimeout(() => this.connect(), this.reconnectInterval);
    } else {
      this.emit('maxRetriesReached', { retries: this.retries });
    }
  }

  close(): void {
    this.stopPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
```

### Client Implementation (Go)

```go
// packages/events/websocket_client.go
package events

import (
    "context"
    "encoding/json"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type WebSocketClient struct {
    endpoint          string
    conn              *websocket.Conn
    mu                sync.Mutex
    reconnectInterval time.Duration
    pingInterval      time.Duration
    maxRetries        int
    handlers          map[string][]func(json.RawMessage)
    subscriptions     []string
    done              chan struct{}
}

type WSMessage struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data,omitempty"`
}

func NewWebSocketClient(endpoint string, opts ...WSOption) *WebSocketClient {
    client := &WebSocketClient{
        endpoint:          endpoint,
        reconnectInterval: 5 * time.Second,
        pingInterval:      30 * time.Second,
        maxRetries:        10,
        handlers:          make(map[string][]func(json.RawMessage)),
        done:              make(chan struct{}),
    }

    for _, opt := range opts {
        opt(client)
    }

    return client
}

func (c *WebSocketClient) Connect(ctx context.Context) error {
    conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.endpoint, nil)
    if err != nil {
        return err
    }

    c.mu.Lock()
    c.conn = conn
    c.mu.Unlock()

    // Start ping/pong
    go c.pingLoop(ctx)

    // Start message handler
    go c.readLoop(ctx)

    // Resubscribe
    if len(c.subscriptions) > 0 {
        c.Subscribe(c.subscriptions)
    }

    return nil
}

func (c *WebSocketClient) Subscribe(events []string) error {
    c.subscriptions = append(c.subscriptions, events...)
    return c.send(WSMessage{
        Type: "subscribe",
        Data: mustMarshal(map[string][]string{"events": events}),
    })
}

func (c *WebSocketClient) On(eventType string, handler func(json.RawMessage)) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.handlers[eventType] = append(c.handlers[eventType], handler)
}

func (c *WebSocketClient) send(msg WSMessage) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn == nil {
        return ErrNotConnected
    }

    return c.conn.WriteJSON(msg)
}

func (c *WebSocketClient) readLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-c.done:
            return
        default:
            var msg WSMessage
            err := c.conn.ReadJSON(&msg)
            if err != nil {
                c.handleDisconnect(ctx)
                return
            }

            c.dispatch(msg)
        }
    }
}

func (c *WebSocketClient) dispatch(msg WSMessage) {
    c.mu.Lock()
    handlers := c.handlers[msg.Type]
    wildcardHandlers := c.handlers["*"]
    c.mu.Unlock()

    for _, handler := range handlers {
        go handler(msg.Data)
    }
    for _, handler := range wildcardHandlers {
        go handler(mustMarshal(msg))
    }
}

func (c *WebSocketClient) pingLoop(ctx context.Context) {
    ticker := time.NewTicker(c.pingInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-c.done:
            return
        case <-ticker.C:
            c.send(WSMessage{Type: "ping"})
        }
    }
}

func (c *WebSocketClient) Close() error {
    close(c.done)
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}
```

---

## Webhooks

### Overview

Webhooks provide server-to-server event delivery with at-least-once semantics.

### Configuration

```json
{
  "webhooks": {
    "url": "https://your-server.com/helix-webhook",
    "secret": "your-webhook-secret",
    "events": ["task.completed", "task.failed", "debate.completed"],
    "retries": 3,
    "retryInterval": 5000
  }
}
```

### Webhook Payload

```json
POST /helix-webhook
Content-Type: application/json
X-HelixAgent-Signature: sha256=<hmac-signature>
X-HelixAgent-Event: task.completed
X-HelixAgent-Delivery: <delivery-id>

{
  "id": "evt_12345",
  "type": "task.completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "taskId": "task_abc",
    "result": "success",
    "duration": 5200
  }
}
```

### Signature Verification

```typescript
// packages/events/src/webhook_handler.ts
import * as crypto from 'crypto';

export function verifyWebhookSignature(
  payload: string,
  signature: string,
  secret: string
): boolean {
  const expectedSignature = `sha256=${crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex')}`;

  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

// Express middleware example
export function webhookMiddleware(secret: string) {
  return (req: Request, res: Response, next: NextFunction) => {
    const signature = req.headers['x-helixagent-signature'] as string;
    const payload = JSON.stringify(req.body);

    if (!verifyWebhookSignature(payload, signature, secret)) {
      return res.status(401).json({ error: 'Invalid signature' });
    }

    next();
  };
}
```

### Webhook Handler (Go)

```go
// packages/events/webhook_handler.go
package events

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
)

type WebhookHandler struct {
    secret   string
    handlers map[string][]func(WebhookEvent)
}

type WebhookEvent struct {
    ID        string          `json:"id"`
    Type      string          `json:"type"`
    Timestamp string          `json:"timestamp"`
    Data      json.RawMessage `json:"data"`
}

func NewWebhookHandler(secret string) *WebhookHandler {
    return &WebhookHandler{
        secret:   secret,
        handlers: make(map[string][]func(WebhookEvent)),
    }
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Verify signature
    signature := r.Header.Get("X-HelixAgent-Signature")
    body, _ := io.ReadAll(r.Body)

    if !h.verifySignature(body, signature) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Parse event
    var event WebhookEvent
    if err := json.Unmarshal(body, &event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    // Dispatch to handlers
    h.dispatch(event)

    w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(h.secret))
    mac.Write(payload)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expected))
}

func (h *WebhookHandler) On(eventType string, handler func(WebhookEvent)) {
    h.handlers[eventType] = append(h.handlers[eventType], handler)
}

func (h *WebhookHandler) dispatch(event WebhookEvent) {
    for _, handler := range h.handlers[event.Type] {
        go handler(event)
    }
    for _, handler := range h.handlers["*"] {
        go handler(event)
    }
}
```

---

## Event Subscription Patterns

### Subscribe to All Task Events

```typescript
const client = new SSEClient({
  endpoint: 'http://localhost:7061/v1/tasks/*/events'
});

client.on('task.*', (event) => {
  console.log('Task event:', event);
});
```

### Subscribe to Specific Debate

```typescript
const client = new WebSocketClient({
  endpoint: 'ws://localhost:7061/v1/ws/debates/debate_123'
});

client.subscribe(['debate.round_started', 'debate.consensus', 'debate.completed']);

client.on('debate.consensus', (data) => {
  console.log('Consensus reached:', data.consensus);
  console.log('Confidence:', data.confidence);
});
```

### Wildcard Subscriptions

```typescript
// Subscribe to all events
client.subscribe(['*']);

// Subscribe to all task events
client.subscribe(['task.*']);

// Subscribe to all debate events
client.subscribe(['debate.*']);
```

---

## Configuration Reference

### Full Event Configuration

```json
{
  "events": {
    "transport": "sse",
    "sse": {
      "reconnectInterval": 5000,
      "maxRetries": 10
    },
    "websocket": {
      "reconnectInterval": 5000,
      "maxRetries": 10,
      "pingInterval": 30000
    },
    "webhook": {
      "url": "https://your-server.com/webhook",
      "secret": "your-secret",
      "retries": 3,
      "retryInterval": 5000
    },
    "subscriptions": {
      "task": ["task.started", "task.completed", "task.failed"],
      "debate": ["debate.started", "debate.consensus", "debate.completed"]
    }
  }
}
```
