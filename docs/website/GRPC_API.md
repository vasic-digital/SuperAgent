# HelixAgent gRPC API

High-performance binary protocol for service-to-service communication with HelixAgent.

---

## Overview

HelixAgent provides a gRPC API alongside its REST API for scenarios requiring:

- **Lower latency** - Binary protocol reduces serialization overhead
- **Streaming** - Efficient bidirectional streaming for real-time responses
- **Strong typing** - Protocol buffer contracts for compile-time safety
- **Service mesh integration** - Native gRPC support in Kubernetes/Istio

---

## Quick Start

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

### Start Server

```bash
helixagent serve --grpc-port=9090
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

---

## Service Definition

The `LLMFacade` service provides all core HelixAgent functionality:

```protobuf
syntax = "proto3";

package helixagent.v1;

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

---

## Authentication

### API Key Authentication

Include API key in gRPC metadata:

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

---

## Methods

### Complete

Performs a single LLM completion using the AI debate ensemble.

**Request:**

```protobuf
message CompletionRequest {
  string prompt = 1;
  string model = 2;           // Default: "helixagent-debate"
  string provider = 3;        // Optional: force specific provider
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
    Model:       "helixagent-debate",
    MaxTokens:   100,
    Temperature: 0.7,
}

response, err := client.Complete(ctx, request)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Response:", response.Content)
fmt.Printf("Tokens used: %d\n", response.Usage.TotalTokens)
fmt.Printf("Latency: %dms\n", response.LatencyMs)
```

---

### CompleteStream

Streams completion responses for real-time output.

**Example:**

```go
request := &pb.CompletionRequest{
    Prompt:    "Write a short story about a robot learning to paint.",
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
    fmt.Print(response.Content)  // Print tokens as they arrive
}
fmt.Println()
```

---

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

**Response:**

```protobuf
message ChatResponse {
  string id = 1;
  string content = 2;        // Token or full response
  bool is_final = 3;         // True for last chunk
  UsageInfo usage = 4;       // Only on final chunk
  int64 latency_ms = 5;
}
```

**Example:**

```go
request := &pb.ChatRequest{
    Messages: []*pb.ChatMessage{
        {Role: "system", Content: "You are a helpful coding assistant."},
        {Role: "user", Content: "How do I reverse a string in Go?"},
    },
    MaxTokens:   500,
    Temperature: 0.7,
}

stream, err := client.Chat(ctx, request)
if err != nil {
    log.Fatal(err)
}

var fullResponse strings.Builder
for {
    response, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    fullResponse.WriteString(response.Content)
    fmt.Print(response.Content)

    if response.IsFinal {
        fmt.Printf("\n\nTokens: %d\n", response.Usage.TotalTokens)
    }
}
```

---

### ListProviders

Lists all available LLM providers and their status.

**Request:**

```protobuf
message ListProvidersRequest {
  bool include_disabled = 1;
  string filter = 2;          // Filter by name
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
  string health_status = 5;   // "healthy", "degraded", "unhealthy"
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
    fmt.Printf("Provider: %s (score: %.2f, status: %s)\n",
        provider.Name, provider.Score, provider.HealthStatus)
    fmt.Printf("  Models: %v\n", provider.Models)
    fmt.Printf("  Capabilities: streaming=%v, tools=%v, vision=%v\n",
        provider.Capabilities.Streaming,
        provider.Capabilities.Tools,
        provider.Capabilities.Vision)
}
```

---

### AddProvider

Dynamically adds a new LLM provider.

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

fmt.Printf("Provider added: %s (score: %.2f)\n", response.Name, response.Score)
```

---

### HealthCheck

Checks service and component health.

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
  string status = 1;                          // "healthy", "degraded", "unhealthy"
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

fmt.Printf("Status: %s (uptime: %ds, version: %s)\n",
    response.Status, response.UptimeSeconds, response.Version)

for name, health := range response.Components {
    fmt.Printf("  %s: %s (%dms)\n", name, health.Status, health.LatencyMs)
}
```

---

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

**Example:**

```go
response, err := client.GetMetrics(ctx, &pb.MetricsRequest{})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total requests: %d\n", response.TotalRequests)
fmt.Printf("Total tokens: %d\n", response.TotalTokens)
fmt.Printf("Avg latency: %.2fms\n", response.AvgLatencyMs)
fmt.Printf("Cache hit rate: %.2f%%\n", response.CacheHitRate*100)

for name, metrics := range response.ProviderMetrics {
    fmt.Printf("  %s: %d requests, %.2fms avg\n",
        name, metrics.Requests, metrics.AvgLatencyMs)
}
```

---

### Session Management

#### CreateSession

```go
request := &pb.CreateSessionRequest{
    UserId:    "user-123",
    Provider:  "helixagent-debate",
    ExpiresIn: 3600, // seconds
}

response, err := client.CreateSession(ctx, request)
fmt.Printf("Session created: %s\n", response.SessionId)
```

#### GetSession

```go
request := &pb.GetSessionRequest{
    SessionId: "session-123",
}

response, err := client.GetSession(ctx, request)
fmt.Printf("Messages: %d, Last activity: %s\n",
    response.MessageCount, response.LastActivity)
```

#### TerminateSession

```go
request := &pb.TerminateSessionRequest{
    SessionId: "session-123",
}

response, err := client.TerminateSession(ctx, request)
fmt.Printf("Session terminated: %v\n", response.Success)
```

---

## Error Handling

gRPC uses standard status codes:

| Code | Description | When |
|------|-------------|------|
| `OK` | Success | Request completed |
| `INVALID_ARGUMENT` | Bad request parameters | Validation failed |
| `UNAUTHENTICATED` | Missing or invalid credentials | Auth required |
| `PERMISSION_DENIED` | Insufficient permissions | RBAC violation |
| `NOT_FOUND` | Resource not found | Invalid ID |
| `RESOURCE_EXHAUSTED` | Rate limit exceeded | Too many requests |
| `INTERNAL` | Server error | Unexpected error |
| `UNAVAILABLE` | Service unavailable | Maintenance/overload |
| `DEADLINE_EXCEEDED` | Request timeout | Slow response |

**Example error handling:**

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

response, err := client.Complete(ctx, request)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.InvalidArgument:
            log.Printf("Invalid request: %s", st.Message())
        case codes.ResourceExhausted:
            log.Printf("Rate limited, retry after backoff")
        case codes.Unavailable:
            log.Printf("Service unavailable, retry with backoff")
        case codes.DeadlineExceeded:
            log.Printf("Request timed out")
        default:
            log.Printf("Error: %s", st.Message())
        }
    }
    return
}
```

---

## Interceptors

### Client Logging

```go
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

### Tracing Interceptor

```go
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

conn, err := grpc.Dial("localhost:9090",
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
    grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
)
```

---

## Performance Optimization

### Connection Pooling

```go
// Create a pool of connections
pool := make([]*grpc.ClientConn, 10)
for i := range pool {
    pool[i], _ = grpc.Dial("localhost:9090", grpc.WithInsecure())
}

var counter uint64

func getConnection() *grpc.ClientConn {
    return pool[atomic.AddUint64(&counter, 1) % uint64(len(pool))]
}
```

### Compression

```go
import "google.golang.org/grpc/encoding/gzip"

conn, err := grpc.Dial("localhost:9090",
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
)
```

### Keepalive

```go
import "google.golang.org/grpc/keepalive"

conn, err := grpc.Dial("localhost:9090",
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,  // Ping every 10s
        Timeout:             3 * time.Second,   // Wait 3s for pong
        PermitWithoutStream: true,
    }),
)
```

### Load Balancing

```go
import (
    "google.golang.org/grpc/balancer/roundrobin"
    "google.golang.org/grpc/resolver"
)

// DNS-based load balancing
conn, err := grpc.Dial(
    "dns:///helixagent.service.local:9090",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

---

## Protobuf Files

Generate Go code from proto files:

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

**Proto file locations:**
- `proto/llm-facade.proto` - Main service definition
- `proto/messages.proto` - Message types

---

## Challenges

Validate gRPC functionality:

```bash
# Run gRPC challenge
./challenges/scripts/grpc_service_challenge.sh

# Expected: 9 tests
# - Server startup
# - Complete method
# - CompleteStream method
# - Chat method
# - ListProviders method
# - HealthCheck method
# - GetMetrics method
# - Authentication
# - Error handling
```

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System architecture
- [API Reference](/docs/api/README.md) - Full API documentation
- [Getting Started](./GETTING_STARTED.md) - Quick start guide

---

**Last Updated**: February 2026
**Version**: 1.0.0
