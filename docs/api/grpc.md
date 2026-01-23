# gRPC API Documentation

## Overview

HelixAgent provides a gRPC API for high-performance, low-latency communication. The gRPC API offers the same functionality as the REST API but with better performance for service-to-service communication.

## Service Definition

The LLMFacade service provides all core HelixAgent functionality:

```protobuf
service LLMFacade {
  // Completion Methods
  rpc Complete(CompletionRequest) returns (CompletionResponse);
  rpc CompleteStream(CompletionRequest) returns (stream CompletionResponse);

  // Chat Methods
  rpc Chat(ChatRequest) returns (stream ChatResponse);

  // Provider Management
  rpc ListProviders(ListProvidersRequest) returns (ListProvidersResponse);
  rpc AddProvider(AddProviderRequest) returns (ProviderResponse);
  rpc UpdateProvider(UpdateProviderRequest) returns (ProviderResponse);
  rpc RemoveProvider(RemoveProviderRequest) returns (ProviderResponse);

  // Health & Metrics
  rpc HealthCheck(HealthRequest) returns (HealthResponse);
  rpc GetMetrics(MetricsRequest) returns (MetricsResponse);

  // Session Management
  rpc CreateSession(CreateSessionRequest) returns (SessionResponse);
  rpc GetSession(GetSessionRequest) returns (SessionResponse);
  rpc TerminateSession(TerminateSessionRequest) returns (SessionResponse);
}
```

## Connection

### Server Configuration

```yaml
grpc:
  enabled: true
  host: 0.0.0.0
  port: 9090
  tls:
    enabled: true
    cert_file: /etc/ssl/certs/helixagent.crt
    key_file: /etc/ssl/private/helixagent.key
  max_message_size: 100MB
  keepalive:
    time: 30s
    timeout: 10s
```

### Client Connection

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    pb "dev.helix.agent/pkg/api"
)

// With TLS
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
if err != nil {
    log.Fatal(err)
}
conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(creds))

// Without TLS (development only)
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())

if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewLLMFacadeClient(conn)
```

## Authentication

### API Key Authentication

Include API key in metadata:

```go
import "google.golang.org/grpc/metadata"

ctx := metadata.AppendToOutgoingContext(context.Background(),
    "x-api-key", "your-api-key",
)
response, err := client.Complete(ctx, request)
```

### JWT Authentication

Include JWT token in metadata:

```go
ctx := metadata.AppendToOutgoingContext(context.Background(),
    "authorization", "Bearer your-jwt-token",
)
response, err := client.Complete(ctx, request)
```

## Methods

### Complete

Performs a single LLM completion.

**Request:**
```protobuf
message CompletionRequest {
  string prompt = 1;
  string model = 2;
  string provider = 3;
  int32 max_tokens = 4;
  float temperature = 5;
  float top_p = 6;
  repeated string stop = 7;
  bool stream = 8;
  map<string, string> metadata = 9;
}
```

**Response:**
```protobuf
message CompletionResponse {
  string id = 1;
  string content = 2;
  string model = 3;
  string provider = 4;
  UsageInfo usage = 5;
  int64 latency_ms = 6;
  bool cached = 7;
  string error = 8;
}

message UsageInfo {
  int32 prompt_tokens = 1;
  int32 completion_tokens = 2;
  int32 total_tokens = 3;
}
```

**Example:**
```go
request := &pb.CompletionRequest{
    Prompt:      "What is the capital of France?",
    Model:       "claude-3-5-sonnet",
    MaxTokens:   100,
    Temperature: 0.7,
}

response, err := client.Complete(ctx, request)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Response:", response.Content)
fmt.Printf("Tokens used: %d\n", response.Usage.TotalTokens)
```

### CompleteStream

Streams completion responses for real-time output.

**Example:**
```go
request := &pb.CompletionRequest{
    Prompt:    "Write a short story about...",
    MaxTokens: 500,
}

stream, err := client.CompleteStream(ctx, request)
if err != nil {
    log.Fatal(err)
}

for {
    response, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(response.Content)
}
```

### Chat

Streams chat responses with conversation context.

**Request:**
```protobuf
message ChatRequest {
  repeated ChatMessage messages = 1;
  string model = 2;
  string provider = 3;
  int32 max_tokens = 4;
  float temperature = 5;
  string system_prompt = 6;
}

message ChatMessage {
  string role = 1;    // "user", "assistant", "system"
  string content = 2;
}
```

**Example:**
```go
request := &pb.ChatRequest{
    Messages: []*pb.ChatMessage{
        {Role: "user", Content: "Hello!"},
        {Role: "assistant", Content: "Hi! How can I help you?"},
        {Role: "user", Content: "What's the weather like?"},
    },
    MaxTokens: 100,
}

stream, err := client.Chat(ctx, request)
if err != nil {
    log.Fatal(err)
}

for {
    response, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(response.Content)
}
```

### ListProviders

Lists all available LLM providers.

**Request:**
```protobuf
message ListProvidersRequest {
  bool include_disabled = 1;
  string filter = 2;
}
```

**Response:**
```protobuf
message ListProvidersResponse {
  repeated ProviderInfo providers = 1;
}

message ProviderInfo {
  string name = 1;
  string type = 2;
  bool enabled = 3;
  float score = 4;
  string health_status = 5;
  repeated string models = 6;
  Capabilities capabilities = 7;
}

message Capabilities {
  bool streaming = 1;
  bool tools = 2;
  bool vision = 3;
  int32 max_tokens = 4;
}
```

**Example:**
```go
request := &pb.ListProvidersRequest{
    IncludeDisabled: false,
}

response, err := client.ListProviders(ctx, request)
if err != nil {
    log.Fatal(err)
}

for _, provider := range response.Providers {
    fmt.Printf("Provider: %s (score: %.2f)\n", provider.Name, provider.Score)
}
```

### AddProvider

Adds a new LLM provider dynamically.

**Request:**
```protobuf
message AddProviderRequest {
  string name = 1;
  string type = 2;
  string api_key = 3;
  map<string, string> config = 4;
  int32 priority = 5;
}
```

**Example:**
```go
request := &pb.AddProviderRequest{
    Name:   "my-openai",
    Type:   "openai",
    ApiKey: "sk-...",
    Config: map[string]string{
        "base_url": "https://api.openai.com/v1",
    },
    Priority: 5,
}

response, err := client.AddProvider(ctx, request)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Provider added:", response.Name)
```

### HealthCheck

Checks service health.

**Request:**
```protobuf
message HealthRequest {
  bool check_providers = 1;
  bool check_database = 2;
  bool check_cache = 3;
}
```

**Response:**
```protobuf
message HealthResponse {
  string status = 1;  // "healthy", "degraded", "unhealthy"
  map<string, ComponentHealth> components = 2;
  int64 uptime_seconds = 3;
  string version = 4;
}

message ComponentHealth {
  string status = 1;
  string message = 2;
  int64 latency_ms = 3;
}
```

**Example:**
```go
request := &pb.HealthRequest{
    CheckProviders: true,
    CheckDatabase:  true,
    CheckCache:     true,
}

response, err := client.HealthCheck(ctx, request)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Status: %s (uptime: %ds)\n", response.Status, response.UptimeSeconds)
```

### GetMetrics

Retrieves service metrics.

**Response:**
```protobuf
message MetricsResponse {
  int64 total_requests = 1;
  int64 total_tokens = 2;
  float avg_latency_ms = 3;
  float cache_hit_rate = 4;
  map<string, ProviderMetrics> provider_metrics = 5;
}

message ProviderMetrics {
  int64 requests = 1;
  int64 errors = 2;
  float avg_latency_ms = 3;
  int64 tokens = 4;
}
```

### Session Management

#### CreateSession

Creates a new conversation session.

```go
request := &pb.CreateSessionRequest{
    UserId:    "user-123",
    Provider:  "claude",
    ExpiresIn: 3600, // seconds
}

response, err := client.CreateSession(ctx, request)
fmt.Println("Session ID:", response.SessionId)
```

#### GetSession

Retrieves session information.

```go
request := &pb.GetSessionRequest{
    SessionId: "session-123",
}

response, err := client.GetSession(ctx, request)
fmt.Printf("Messages: %d, Last activity: %s\n",
    response.MessageCount, response.LastActivity)
```

#### TerminateSession

Ends a session and cleans up resources.

```go
request := &pb.TerminateSessionRequest{
    SessionId: "session-123",
}

response, err := client.TerminateSession(ctx, request)
fmt.Println("Session terminated:", response.Success)
```

## Error Handling

gRPC errors use standard status codes:

| Code | Description |
|------|-------------|
| `OK` | Success |
| `INVALID_ARGUMENT` | Bad request parameters |
| `UNAUTHENTICATED` | Missing or invalid credentials |
| `PERMISSION_DENIED` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `RESOURCE_EXHAUSTED` | Rate limit exceeded |
| `INTERNAL` | Server error |
| `UNAVAILABLE` | Service unavailable |
| `DEADLINE_EXCEEDED` | Request timeout |

**Example error handling:**
```go
import "google.golang.org/grpc/status"

response, err := client.Complete(ctx, request)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.InvalidArgument:
            log.Printf("Invalid request: %s", st.Message())
        case codes.ResourceExhausted:
            log.Printf("Rate limited: %s", st.Message())
        case codes.Unavailable:
            log.Printf("Service unavailable: %s", st.Message())
        default:
            log.Printf("Error: %s", st.Message())
        }
    }
}
```

## Interceptors

### Client Interceptors

```go
// Logging interceptor
func loggingInterceptor(ctx context.Context, method string, req, reply interface{},
    cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

    start := time.Now()
    err := invoker(ctx, method, req, reply, cc, opts...)
    log.Printf("Method: %s, Duration: %v, Error: %v", method, time.Since(start), err)
    return err
}

conn, err := grpc.Dial("localhost:9090",
    grpc.WithUnaryInterceptor(loggingInterceptor),
)
```

### Retry Interceptor

```go
import "github.com/grpc-ecosystem/go-grpc-middleware/retry"

retryOpts := []grpc_retry.CallOption{
    grpc_retry.WithMax(3),
    grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
    grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
}

conn, err := grpc.Dial("localhost:9090",
    grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
    grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
)
```

## Performance Tips

### Connection Pooling

```go
// Use a connection pool for high-throughput applications
pool := make([]*grpc.ClientConn, 10)
for i := range pool {
    pool[i], _ = grpc.Dial("localhost:9090", grpc.WithInsecure())
}

// Round-robin selection
func getConnection() *grpc.ClientConn {
    return pool[atomic.AddUint64(&counter, 1) % uint64(len(pool))]
}
```

### Compression

```go
import "google.golang.org/grpc/encoding/gzip"

// Enable compression
conn, err := grpc.Dial("localhost:9090",
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
)
```

### Keepalive

```go
import "google.golang.org/grpc/keepalive"

conn, err := grpc.Dial("localhost:9090",
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
)
```

## Protobuf Files

The protobuf definitions are located at:
- `proto/llm-facade.proto` - Main service definition
- `proto/messages.proto` - Message types

Generate Go code:
```bash
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
