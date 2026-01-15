# TOON Protocol

Token-Optimized Object Notation (TOON) is HelixAgent's custom serialization format designed to minimize token consumption when communicating with AI systems via MCP.

## Overview

TOON reduces the token footprint of JSON data through intelligent key compression, value abbreviation, and optional binary compression. This is critical for:
- Reducing API costs when communicating with LLMs
- Maximizing context window utilization
- Improving response latency through smaller payloads
- Efficient MCP (Model Context Protocol) communication

## Compression Levels

| Level | Description | Savings | Use Case |
|-------|-------------|---------|----------|
| `None` | No compression | 0% | Debug/development |
| `Minimal` | Key compression only | 20-30% | Light optimization |
| `Standard` | Key + value compression | 40-50% | **Default** |
| `Aggressive` | Full compression + gzip | 60-70% | Bandwidth constrained |

## Key Compression

Common JSON keys are compressed to single characters:

| Original Key | Compressed | Savings |
|--------------|------------|---------|
| `id` | `i` | 1 byte |
| `name` | `n` | 3 bytes |
| `type` | `t` | 3 bytes |
| `status` | `s` | 5 bytes |
| `message` | `m` | 6 bytes |
| `content` | `c` | 6 bytes |
| `created_at` | `ca` | 8 bytes |
| `updated_at` | `ua` | 8 bytes |
| `description` | `d` | 10 bytes |
| `error` | `e` | 4 bytes |
| `data` | `da` | 2 bytes |
| `result` | `r` | 5 bytes |
| `value` | `v` | 4 bytes |
| `timestamp` | `ts` | 7 bytes |
| `priority` | `p` | 7 bytes |
| `provider` | `pr` | 6 bytes |
| `model` | `mo` | 4 bytes |
| `response` | `re` | 7 bytes |
| `request` | `rq` | 6 bytes |

### Custom Key Mappings

Register domain-specific key mappings:

```go
encoder := toon.NewEncoder(toon.CompressionStandard)
encoder.AddKeyMapping("verification_score", "vs")
encoder.AddKeyMapping("debate_round", "dr")
encoder.AddKeyMapping("llm_provider", "lp")
```

## Value Abbreviations

Common string values are abbreviated:

| Original Value | Abbreviated | Context |
|----------------|-------------|---------|
| `healthy` | `H` | Status |
| `unhealthy` | `U` | Status |
| `pending` | `P` | State |
| `running` | `R` | State |
| `completed` | `C` | State |
| `failed` | `F` | State |
| `success` | `S` | Result |
| `error` | `E` | Result |
| `true` | `T` | Boolean |
| `false` | `X` | Boolean |
| `null` | `N` | Null |

## Encoder API

### Basic Usage

```go
import "dev.helix.agent/internal/toon"

// Create encoder with compression level
encoder := toon.NewEncoder(toon.CompressionStandard)

// Encode struct to TOON
data := map[string]interface{}{
    "id":      "123",
    "name":    "test",
    "status":  "healthy",
}

encoded, err := encoder.Encode(data)
if err != nil {
    log.Fatal(err)
}

// Result: {"i":"123","n":"test","s":"H"}
fmt.Println(string(encoded))
```

### Encode to String

```go
str, err := encoder.EncodeToString(data)
// Result: {"i":"123","n":"test","s":"H"}
```

### With Custom Options

```go
encoder := toon.NewEncoder(toon.CompressionStandard,
    toon.WithKeyMapping(customKeyMap),
    toon.WithValueAbbreviations(customValueMap),
    toon.WithGzipThreshold(1024),
)
```

## Decoder API

### Basic Usage

```go
decoder := toon.NewDecoder()

// Decode TOON back to original format
var result map[string]interface{}
err := decoder.Decode(encoded, &result)
if err != nil {
    log.Fatal(err)
}

// Result: {"id":"123","name":"test","status":"healthy"}
```

### Decode String

```go
err := decoder.DecodeString(str, &result)
```

## Token Counting

TOON includes token estimation for cost calculations:

```go
// Estimate token count
original := []byte(`{"id":"123","name":"test","status":"healthy"}`)
encoded, _ := encoder.Encode(data)

originalTokens := toon.TokenCount(original)  // ~12 tokens
encodedTokens := toon.TokenCount(encoded)    // ~8 tokens
savings := originalTokens - encodedTokens     // 4 tokens saved

// Calculate compression ratio
ratio := encoder.CompressionRatio(original, encoded)
// Result: 0.33 (33% reduction)
```

## Transport Layer

TOON provides an HTTP transport for API communication:

### Transport Configuration

```go
transport := toon.NewTransport(&toon.TransportConfig{
    BaseURL:          "https://api.example.com",
    CompressionLevel: toon.CompressionStandard,
    HTTPClient:       &http.Client{Timeout: 30 * time.Second},
})
```

### HTTP Methods

```go
// GET request
resp, err := transport.Get(ctx, "/providers")

// POST request with TOON body
resp, err := transport.Post(ctx, "/debates", requestData)

// Custom request
req := &toon.Request{
    Method:  "PUT",
    Path:    "/providers/123",
    Body:    updateData,
    Headers: map[string]string{"X-Custom": "value"},
}
resp, err := transport.Do(ctx, req)
```

### Content Type

TOON uses a custom content type for protocol identification:

```
Content-Type: application/toon+json
```

Servers can negotiate format:
```
Accept: application/toon+json, application/json;q=0.9
```

## Middleware

TOON middleware for Gin framework:

```go
import "dev.helix.agent/internal/toon"

// Create middleware
middleware := toon.NewMiddleware(&toon.MiddlewareConfig{
    CompressionLevel: toon.CompressionStandard,
    EnableMetrics:    true,
})

// Apply to router
router := gin.Default()
router.Use(middleware.Handler())

// Middleware automatically:
// 1. Decodes TOON requests to JSON
// 2. Encodes JSON responses to TOON
// 3. Tracks metrics
```

### Selective Routes

```go
// Only apply to specific routes
api := router.Group("/v1")
api.Use(middleware.Handler())

// Bypass for specific endpoints
router.GET("/health", healthHandler)  // No TOON
```

## Metrics

TOON tracks compression metrics:

```go
type TransportMetrics struct {
    RequestCount     int64   // Total requests
    BytesSent        int64   // Original bytes sent
    BytesReceived    int64   // Original bytes received
    BytesSaved       int64   // Bytes saved by compression
    TokensSaved      int64   // Estimated tokens saved
    CompressionRatio float64 // Average compression ratio
    EncodeLatency    int64   // Total encode time (ns)
    DecodeLatency    int64   // Total decode time (ns)
}

// Get metrics
metrics := transport.Metrics()
fmt.Printf("Tokens saved: %d\n", metrics.TokensSaved)
fmt.Printf("Compression ratio: %.2f%%\n", metrics.CompressionRatio*100)
```

## MCP Integration

TOON is optimized for MCP (Model Context Protocol):

```go
// MCP handler with TOON
func (h *MCPHandler) HandleRequest(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
    // Decode TOON request if present
    if req.ContentType == "application/toon+json" {
        decoded, err := h.toonDecoder.Decode(req.Body, &request)
        if err != nil {
            return nil, err
        }
        req.Body = decoded
    }

    // Process request...
    response := processRequest(req)

    // Encode response as TOON
    encoded, err := h.toonEncoder.Encode(response)
    if err != nil {
        return nil, err
    }

    return &mcp.Response{
        ContentType: "application/toon+json",
        Body:        encoded,
    }, nil
}
```

## Compression Examples

### Before/After Comparison

**Original JSON (45 tokens, 312 bytes):**
```json
{
  "id": "debate-123",
  "name": "AI Ethics Discussion",
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "participants": [
    {"id": "p1", "name": "Claude", "status": "healthy"},
    {"id": "p2", "name": "GPT-4", "status": "healthy"}
  ],
  "result": {
    "winner": "consensus",
    "confidence": 0.95,
    "message": "All participants agree"
  }
}
```

**TOON Compressed (28 tokens, 156 bytes):**
```json
{
  "i": "debate-123",
  "n": "AI Ethics Discussion",
  "s": "C",
  "ca": "2024-01-15T10:30:00Z",
  "participants": [
    {"i": "p1", "n": "Claude", "s": "H"},
    {"i": "p2", "n": "GPT-4", "s": "H"}
  ],
  "r": {
    "winner": "consensus",
    "confidence": 0.95,
    "m": "All participants agree"
  }
}
```

**Savings: 38% tokens, 50% bytes**

### Aggressive Compression

With `CompressionAggressive` + gzip, the same payload compresses to ~80 bytes (74% reduction).

## Best Practices

### When to Use TOON

| Scenario | Recommendation |
|----------|----------------|
| LLM API requests | **Standard** compression |
| MCP communication | **Standard** compression |
| Real-time streaming | **Minimal** or None |
| Audit logs | **Aggressive** compression |
| Debug/development | **None** |

### Performance Considerations

1. **Cache encoded data** - Avoid re-encoding static data
2. **Batch requests** - Compress multiple items together
3. **Measure impact** - Monitor token savings vs encode latency
4. **Test edge cases** - Ensure decoder handles all value types

### Security

1. **Validate input** - Decode can fail on malformed data
2. **Size limits** - Set max payload size before decode
3. **Schema validation** - Validate decoded data structure

## Testing

```bash
# Run TOON unit tests
go test ./internal/toon/... -v

# Run TOON challenge
./challenges/scripts/toon_protocol_challenge.sh

# Benchmark compression
go test ./internal/toon/... -bench=. -benchmem
```

## Configuration

### Environment Variables

```bash
TOON_COMPRESSION_LEVEL=standard
TOON_GZIP_THRESHOLD=1024
TOON_ENABLE_METRICS=true
```

### Configuration File

```yaml
toon:
  compression_level: standard  # none, minimal, standard, aggressive
  gzip_threshold: 1024         # bytes before gzip kicks in (aggressive mode)
  enable_metrics: true
  custom_key_mappings:
    verification_score: vs
    debate_round: dr
```

## Error Handling

```go
// Decode errors
var result map[string]interface{}
err := decoder.Decode(data, &result)
if err != nil {
    switch {
    case errors.Is(err, toon.ErrInvalidFormat):
        // Not valid TOON/JSON
    case errors.Is(err, toon.ErrDecompressionFailed):
        // Gzip decompression failed
    case errors.Is(err, toon.ErrUnknownKey):
        // Unknown compressed key (strict mode)
    default:
        // Other error
    }
}
```

## Future Enhancements

- **Schema Registry** - Pre-registered schemas for maximum compression
- **Binary Mode** - Pure binary encoding for non-text transports
- **Streaming Encoder** - Memory-efficient streaming compression
- **Protocol Buffers** - Optional protobuf backend
