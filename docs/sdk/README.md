# SDK and Client Libraries

This directory contains documentation for HelixAgent client SDKs across multiple programming languages.

## Overview

HelixAgent provides official SDKs for Python, JavaScript/TypeScript, Go, and mobile platforms (iOS/Android). All SDKs offer OpenAI-compatible APIs, making it easy to migrate existing applications or integrate with HelixAgent's advanced ensemble and debate features.

## Documentation Index

| Document | Description |
|----------|-------------|
| [Python SDK](./python-sdk.md) | Python SDK documentation and examples |
| [JavaScript SDK](./javascript-sdk.md) | JavaScript/TypeScript SDK with full type support |
| [Go SDK](./go-sdk.md) | Idiomatic Go SDK documentation |
| [Mobile SDKs](./mobile-sdks.md) | iOS (Swift) and Android (Kotlin) SDKs |

## SDK Availability

| SDK | Status | Installation | Source |
|-----|--------|--------------|--------|
| Python | Available | `pip install helixagent-sdk` | `/sdk/python/` |
| JavaScript | Available | `npm install helixagent-sdk` | `/sdk/web/` |
| Go | Available | `go get dev.helix.agent-go` | `/sdk/go/` |
| iOS | Available | CocoaPods/SPM | `/sdk/ios/` |
| Android | Available | Maven Central | `/sdk/android/` |

## Quick Start Examples

### Python

```python
from helixagent import HelixAgent

client = HelixAgent(
    api_key="your-api-key",
    base_url="https://api.helixagent.ai"
)

# Simple chat completion
response = client.chat.completions.create(
    model="helixagent-ensemble",
    messages=[
        {"role": "user", "content": "Explain quantum computing"}
    ]
)

print(response.choices[0].message.content)
```

### JavaScript/TypeScript

```typescript
import { HelixAgent } from '@helixagent/sdk';

const client = new HelixAgent({
  apiKey: 'your-api-key',
  baseURL: 'https://api.helixagent.ai'
});

// Simple chat completion
const response = await client.chat.completions.create({
  model: 'helixagent-ensemble',
  messages: [
    { role: 'user', content: 'Explain quantum computing' }
  ]
});

console.log(response.choices[0].message.content);
```

### Go

```go
package main

import (
    "context"
    "fmt"
    "log"

    "dev.helix.agent-go"
)

func main() {
    client := helixagent.NewClient(&helixagent.Config{
        APIKey:  "your-api-key",
        BaseURL: "https://api.helixagent.ai",
    })

    resp, err := client.Chat.Completions.Create(context.Background(), &helixagent.ChatCompletionRequest{
        Model: "helixagent-ensemble",
        Messages: []helixagent.ChatMessage{
            {Role: "user", Content: "Explain quantum computing"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## API Client Patterns

### Authentication

All SDKs support multiple authentication methods:

```python
# API Key (recommended)
client = HelixAgent(api_key="your-api-key")

# JWT Token
client = HelixAgent(token="your-jwt-token")

# Environment variable (automatic)
# Set HELIXAGENT_API_KEY environment variable
client = HelixAgent()
```

### Streaming Responses

All SDKs support streaming for real-time output:

```python
# Python streaming
stream = client.chat.completions.create(
    model="helixagent-ensemble",
    messages=[{"role": "user", "content": "Write a story"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

```javascript
// JavaScript streaming
const stream = await client.chat.completions.create({
  model: 'helixagent-ensemble',
  messages: [{ role: 'user', content: 'Write a story' }],
  stream: true
});

for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

### Ensemble Configuration

Configure ensemble behavior for multi-provider responses:

```python
response = client.chat.completions.create(
    model="helixagent-ensemble",
    messages=[{"role": "user", "content": "Your question"}],
    extra_body={
        "ensemble_config": {
            "strategy": "confidence_weighted",
            "providers": ["claude", "deepseek", "gemini"],
            "min_responses": 2
        }
    }
)
```

### AI Debate Mode

Trigger AI debate for complex questions:

```python
response = client.chat.completions.create(
    model="helixagent-debate",
    messages=[{"role": "user", "content": "Your complex question"}],
    extra_body={
        "debate_config": {
            "rounds": 3,
            "participants": 5,
            "strategy": "structured",
            "voting": "confidence_weighted"
        }
    }
)
```

### Error Handling

All SDKs provide consistent error handling:

```python
from helixagent import HelixAgentError, RateLimitError, APIError

try:
    response = client.chat.completions.create(...)
except RateLimitError as e:
    print(f"Rate limited. Retry after {e.retry_after} seconds")
except APIError as e:
    print(f"API error: {e.message}")
except HelixAgentError as e:
    print(f"General error: {e}")
```

### Retry Configuration

Configure automatic retries:

```python
client = HelixAgent(
    api_key="your-api-key",
    max_retries=3,
    timeout=30.0,
    retry_on_status_codes=[429, 500, 502, 503]
)
```

## SDK Features

| Feature | Python | JavaScript | Go | Mobile |
|---------|--------|------------|-----|--------|
| Chat completions | Yes | Yes | Yes | Yes |
| Streaming | Yes | Yes | Yes | Yes |
| Ensemble mode | Yes | Yes | Yes | Yes |
| AI Debate | Yes | Yes | Yes | Yes |
| Embeddings | Yes | Yes | Yes | Yes |
| Tool calling | Yes | Yes | Yes | Limited |
| Async support | Yes | Yes | Yes (goroutines) | Yes |
| Type safety | Optional | Full | Full | Full |
| Retry logic | Yes | Yes | Yes | Yes |

## OpenAI Compatibility

All HelixAgent SDKs are drop-in replacements for OpenAI SDKs:

```python
# Simply change the import and base URL
# from openai import OpenAI
from helixagent import HelixAgent

# client = OpenAI(api_key="sk-...")
client = HelixAgent(
    api_key="your-api-key",
    base_url="https://api.helixagent.ai"
)

# Rest of your code works unchanged
response = client.chat.completions.create(
    model="helixagent-ensemble",  # or use "gpt-4" model mapping
    messages=[{"role": "user", "content": "Hello"}]
)
```

## Local Development

For local development, point the SDK to your local server:

```python
client = HelixAgent(
    api_key="test-key",
    base_url="http://localhost:7061"
)
```

## Related Documentation

- [API Reference](../api/README.md)
- [Authentication Guide](../security/AUTHENTICATION.md)
- [Rate Limiting](../operations/RATE_LIMITING.md)
