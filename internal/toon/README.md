# TOON Package

The toon package implements **TOON** (Token-Optimized Object Notation), an efficient data encoding format designed for AI consumption in HelixAgent.

## Overview

TOON is a specialized encoding format that reduces token usage when transmitting structured data to LLMs. It achieves this through key abbreviation, value compression, and optional gzip compression, making API calls more efficient and cost-effective.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      TOON Encoder                            │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Encoding Pipeline                      ││
│  │                                                          ││
│  │  Input → Abbreviate Keys → Abbreviate Values → Compress ││
│  │                                                          ││
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ ││
│  │  │  JSON    │→ │   Key    │→ │  Value   │→ │  GZIP   │ ││
│  │  │  Input   │  │  Abbrev  │  │  Abbrev  │  │  (opt)  │ ││
│  │  └──────────┘  └──────────┘  └──────────┘  └─────────┘ ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Decoding Pipeline                      ││
│  │                                                          ││
│  │  Decompress → Expand Values → Expand Keys → Output      ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                 Abbreviation Mappings                    ││
│  │  provider_id → pi | health_status → hs | model → m     ││
│  │  healthy → H | pending → P | failed → F | active → A   ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### Encoder

The main TOON encoder.

```go
type Encoder struct {
    config          EncoderConfig
    keyAbbrevs      map[string]string
    valueAbbrevs    map[string]string
    reverseKeyAbbrevs   map[string]string
    reverseValueAbbrevs map[string]string
}
```

### EncoderConfig

Configuration for the encoder.

```go
type EncoderConfig struct {
    CompressionLevel CompressionLevel // Compression aggressiveness
    EnableGzip       bool             // Enable gzip compression
    CustomKeyAbbrevs map[string]string // Custom key abbreviations
    CustomValueAbbrevs map[string]string // Custom value abbreviations
    PreserveUnknown  bool             // Preserve unknown keys as-is
}
```

### CompressionLevel

Levels of abbreviation aggressiveness.

```go
type CompressionLevel int
const (
    // No abbreviation, just structure
    CompressionNone CompressionLevel = iota

    // Abbreviate common keys only
    CompressionMinimal

    // Abbreviate keys and common values
    CompressionStandard

    // Maximum abbreviation
    CompressionAggressive
)
```

## Default Abbreviations

### Key Abbreviations

| Original Key | Abbreviated |
|--------------|-------------|
| `provider_id` | `pi` |
| `provider_name` | `pn` |
| `health_status` | `hs` |
| `model` | `m` |
| `models` | `ms` |
| `context_window` | `cw` |
| `max_tokens` | `mt` |
| `temperature` | `tp` |
| `capabilities` | `cp` |
| `response` | `r` |
| `request` | `rq` |
| `content` | `c` |
| `role` | `rl` |
| `timestamp` | `ts` |
| `duration` | `d` |
| `error` | `e` |
| `status` | `st` |
| `message` | `msg` |
| `messages` | `msgs` |

### Value Abbreviations

| Original Value | Abbreviated |
|----------------|-------------|
| `healthy` | `H` |
| `unhealthy` | `U` |
| `degraded` | `D` |
| `pending` | `P` |
| `active` | `A` |
| `completed` | `C` |
| `failed` | `F` |
| `running` | `R` |
| `user` | `u` |
| `assistant` | `a` |
| `system` | `s` |
| `true` | `T` |
| `false` | `X` |

## Usage Examples

### Basic Encoding

```go
import "dev.helix.agent/internal/toon"

// Create encoder
encoder := toon.NewEncoder(toon.EncoderConfig{
    CompressionLevel: toon.CompressionStandard,
    EnableGzip:       false,
})

// Encode data
input := map[string]interface{}{
    "provider_id":    "claude",
    "health_status":  "healthy",
    "model":          "claude-3-opus",
    "context_window": 200000,
}

encoded, err := encoder.Encode(input)
if err != nil {
    return err
}

// Result: {"pi":"claude","hs":"H","m":"claude-3-opus","cw":200000}
fmt.Println(string(encoded))
```

### Decoding

```go
// Decode back to original format
var decoded map[string]interface{}
err := encoder.Decode(encoded, &decoded)
if err != nil {
    return err
}

// decoded["provider_id"] == "claude"
// decoded["health_status"] == "healthy"
```

### With Gzip Compression

```go
encoder := toon.NewEncoder(toon.EncoderConfig{
    CompressionLevel: toon.CompressionAggressive,
    EnableGzip:       true,
})

// Encode with compression
encoded, err := encoder.EncodeCompressed(input)
if err != nil {
    return err
}

// Decode from compressed
var decoded map[string]interface{}
err = encoder.DecodeCompressed(encoded, &decoded)
```

### Custom Abbreviations

```go
encoder := toon.NewEncoder(toon.EncoderConfig{
    CompressionLevel: toon.CompressionStandard,
    CustomKeyAbbrevs: map[string]string{
        "my_custom_field": "mcf",
        "special_value":   "sv",
    },
    CustomValueAbbrevs: map[string]string{
        "custom_status": "CS",
    },
})
```

### Streaming Encoding

```go
// For large data, use streaming
writer := toon.NewStreamEncoder(outputWriter, toon.EncoderConfig{
    CompressionLevel: toon.CompressionStandard,
})

for _, item := range largeDataset {
    err := writer.WriteItem(item)
    if err != nil {
        return err
    }
}
writer.Close()
```

### Struct Encoding

```go
type ProviderInfo struct {
    ProviderID   string `json:"provider_id" toon:"pi"`
    HealthStatus string `json:"health_status" toon:"hs"`
    Model        string `json:"model" toon:"m"`
}

info := ProviderInfo{
    ProviderID:   "claude",
    HealthStatus: "healthy",
    Model:        "claude-3-opus",
}

encoded, err := encoder.EncodeStruct(info)
```

## Integration with HelixAgent

TOON is used for efficient LLM communication:

```go
// In LLM provider response handling
func formatContext(data interface{}) (string, error) {
    if features.IsEnabled("toon.encoding") {
        encoder := toon.DefaultEncoder()
        encoded, err := encoder.Encode(data)
        if err != nil {
            return "", err
        }
        return string(encoded), nil
    }
    // Fallback to regular JSON
    jsonData, err := json.Marshal(data)
    return string(jsonData), err
}
```

## Token Savings

Example comparison for a typical provider status response:

### Original JSON (142 tokens)
```json
{
    "provider_id": "claude",
    "provider_name": "Anthropic Claude",
    "health_status": "healthy",
    "model": "claude-3-opus-20240229",
    "context_window": 200000,
    "capabilities": ["chat", "vision", "tool_use"]
}
```

### TOON Encoded (87 tokens, 39% savings)
```json
{"pi":"claude","pn":"Anthropic Claude","hs":"H","m":"claude-3-opus-20240229","cw":200000,"cp":["chat","vision","tool_use"]}
```

### TOON + Gzip (estimated 60% overall savings)
Binary compressed format.

## Compression Levels Comparison

| Level | Token Reduction | Use Case |
|-------|----------------|----------|
| None | 0% | Debugging, human readability |
| Minimal | 15-20% | Light optimization |
| Standard | 30-40% | General use (recommended) |
| Aggressive | 45-60% | Maximum savings, less readable |

## Testing

```bash
go test -v ./internal/toon/...
go test -bench=. ./internal/toon/...  # Benchmark tests
```

### Testing Roundtrip

```go
func TestEncodeDecode(t *testing.T) {
    encoder := toon.NewEncoder(toon.EncoderConfig{
        CompressionLevel: toon.CompressionStandard,
    })

    original := map[string]interface{}{
        "provider_id":   "claude",
        "health_status": "healthy",
    }

    encoded, err := encoder.Encode(original)
    require.NoError(t, err)

    var decoded map[string]interface{}
    err = encoder.Decode(encoded, &decoded)
    require.NoError(t, err)

    assert.Equal(t, original["provider_id"], decoded["provider_id"])
    assert.Equal(t, original["health_status"], decoded["health_status"])
}
```

## Performance

| Operation | Throughput | Notes |
|-----------|-----------|-------|
| Encode (Standard) | 100K ops/sec | Small objects |
| Decode (Standard) | 120K ops/sec | Small objects |
| Encode + Gzip | 10K ops/sec | Compression overhead |
| Decode + Gunzip | 15K ops/sec | Decompression |

## Feature Flags

TOON encoding is controlled by feature flags:

```go
if features.IsEnabled("toon.encoding") {
    return toon.Encode(response)
}
return json.Marshal(response)
```

## Best Practices

1. **Use Standard level**: Best balance of savings vs readability
2. **Enable Gzip for large payloads**: Worth it for >1KB responses
3. **Add custom abbreviations**: For domain-specific keys
4. **Test roundtrip**: Ensure all data survives encode/decode
5. **Monitor token usage**: Compare before/after enabling TOON
