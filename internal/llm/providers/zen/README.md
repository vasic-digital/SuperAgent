# Zen Provider (OpenCode)

This package implements the LLM provider interface for Zen (OpenCode) models.

## Overview

The Zen provider enables HelixAgent to communicate with OpenCode.ai's free API, providing access to AI models without requiring API keys.

## Status

**FREE PROVIDER**: Zen is a free provider with a verification score range of 6.0-7.0. It's used when other providers are unavailable or as a cost-effective alternative.

## Supported Models

| Model | Context | Description |
|-------|---------|-------------|
| opencode | 32K | Default OpenCode model |

## Authentication

No API key required. Uses anonymous access with optional device ID for tracking.

```bash
# Optional: Set device ID for consistent sessions
export OPENCODE_DEVICE_ID="your-device-id"
```

## Configuration

```yaml
providers:
  zen:
    enabled: true
    base_url: "https://api.opencode.ai/v1"
    default_model: "opencode"
    timeout_seconds: 120
```

## Features

- **Free Access**: No API key or payment required
- **Anonymous**: No account registration needed
- **OpenAI Compatible**: Uses OpenAI-compatible API format
- **Streaming**: Real-time response streaming

## Usage

```go
import "dev.helix.agent/internal/llm/providers/zen"

provider := zen.NewZenProvider(config)
response, err := provider.Complete(ctx, request)
```

## Limitations

- Lower rate limits than paid providers
- Variable response times
- Reduced verification score (6.0-7.0)
- May have availability limitations

## Rate Limits

Being a free service, rate limits are conservative:
- Requests per minute: Limited
- Daily usage caps may apply

## Error Handling

The provider handles:
- Rate limits (with backoff)
- Service unavailability
- Network errors

## Testing

```bash
go test -v ./internal/llm/providers/zen/...
```

## Files

- `zen.go` - Main provider implementation
- `zen_test.go` - Unit tests
