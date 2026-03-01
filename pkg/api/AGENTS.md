# AGENTS.md - pkg/api

## Module Overview for AI Agents

This module provides the **public gRPC API** for HelixAgent. It's a generated code package that defines the contract between HelixAgent server and external clients.

## Agent Guidelines

### DO

✅ **Use this package** when building external clients or integrations  
✅ **Reference the types** for API request/response structures  
✅ **Follow proto field naming** (snake_case in proto, CamelCase in Go)  
✅ **Handle all error cases** defined in the API  
✅ **Respect deprecation notices** in the code  

### DON'T

❌ **Modify generated files** (.pb.go files are auto-generated)  
❌ **Add business logic** here - this is just API definitions  
❌ **Import internal packages** from this module  
❌ **Remove or rename fields** without updating the proto definition  
❌ **Add tests** to this package (test the server implementation instead)  

## Common Tasks

### Adding a New API Field

1. Update the Protocol Buffer definition (in docs/ or proto/ directory)
2. Regenerate Go code: `protoc --go_out=. --go-grpc_out=. llm-facade.proto`
3. Update server implementation to handle new field
4. Update client examples and documentation
5. Run integration tests

### Using the API Client

```go
import api "dev.helix.agent/pkg/api"

// Create client
conn, _ := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
client := api.NewLLMFacadeClient(conn)

// Make request
req := &api.CompletionRequest{
    Prompt: "Hello, world!",
    ModelParams: &api.ModelParameters{
        MaxTokens: 100,
    },
}
resp, err := client.Complete(ctx, req)
```

### Handling Streaming Responses

```go
stream, err := client.CompleteStream(ctx, req)
if err != nil {
    return err
}

for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break // Stream complete
    }
    if err != nil {
        return err // Handle error
    }
    
    // Process response chunk
    if chunk := resp.GetChunk(); chunk != nil {
        fmt.Print(chunk.Content)
    }
}
```

## Type Reference

### Request Types

Use these when making API requests:

- `CompletionRequest` - Main request type
- `ModelParameters` - Model configuration (temp, max_tokens, etc.)
- `EnsembleConfig` - Multi-provider configuration
- `Message` - Chat message for conversational requests

### Response Types

Handle these in API responses:

- `CompletionResponse` - Main response type
- `ProviderResponse` - Individual provider output
- `EnsembleResult` - Aggregation metadata
- `UsageStats` - Token usage and latency
- `StreamingResponse` - Stream wrapper type

### Enum Types

Use appropriate enums for configuration:

- `EnsembleStrategy` - SINGLE_BEST, CONFIDENCE_WEIGHTED, VOTING, etc.
- `VotingMethod` - MAJORITY, WEIGHTED, BORDA_COUNT, etc.
- `RequestType` - code_generation, reasoning, tool_use, chat

## Integration Patterns

### Simple Completion

```go
client := api.NewLLMFacadeClient(conn)
resp, err := client.Complete(ctx, &api.CompletionRequest{
    Prompt: "Say hello",
})
```

### With Ensemble

```go
resp, err := client.Complete(ctx, &api.CompletionRequest{
    Prompt: "Complex question",
    EnsembleConfig: &api.EnsembleConfig{
        Strategy: api.EnsembleStrategy_CONFIDENCE_WEIGHTED,
        Providers: []string{"openai", "anthropic"},
    },
})
```

### With Memory

```go
resp, err := client.Complete(ctx, &api.CompletionRequest{
    SessionId:      "session-123",  // Links to memory context
    Prompt:         "What did we discuss?",
    MemoryEnhanced: true,
})
```

## Error Handling

Always handle these error cases:

```go
resp, err := client.Complete(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.InvalidArgument:
            // Bad request - fix request parameters
        case codes.Unavailable:
            // Service down - retry with backoff
        case codes.ResourceExhausted:
            // Rate limit - slow down requests
        case codes.DeadlineExceeded:
            // Timeout - increase timeout or reduce complexity
        }
    }
}
```

## Testing

When testing code that uses this API:

```go
// Use test server
srv := grpc.NewServer()
api.RegisterLLMFacadeServer(srv, mockServer)

// Connect to test server
conn, _ := grpc.Dial("passthrough:///test", grpc.WithContextDialer(...))
client := api.NewLLMFacadeClient(conn)
```

## Constraints

- **No Direct Instantiation**: Don't create `CompletionRequest` with invalid state
- **Required Fields**: Some fields are required (Prompt or Messages)
- **Enum Values**: Use only valid enum values
- **Timeouts**: gRPC calls need context with timeout

## Dependencies

This module only depends on:
- Standard library
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers

Keep it lightweight - no internal HelixAgent dependencies.

## Versioning

Follow semantic versioning:
- **Major**: Breaking changes to proto (removed fields)
- **Minor**: New features (added fields)
- **Patch**: Bug fixes, documentation updates

Always update version when modifying proto definition.

## Examples

See [README.md](README.md) for complete usage examples.

## Questions?

- Check [docs/API_REFERENCE.md](../../docs/API_REFERENCE.md) for detailed API docs
- Look at `tests/e2e/` for integration test examples
- Review server implementation in `internal/handlers/grpc_server.go`
