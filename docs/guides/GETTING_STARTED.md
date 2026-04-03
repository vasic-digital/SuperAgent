# Getting Started with HelixAgent

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone git@github.com:vasic-digital/HelixAgent.git
cd HelixAgent

# Build the main binary
make build

# Or build all applications
make build-all
```

### 2. Configuration

Copy the example environment file and configure your API keys:

```bash
cp .env.example .env
```

Edit `.env` and add at least one provider API key:

```bash
# Required: At least one provider
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
DEEPSEEK_API_KEY=sk-...
GROQ_API_KEY=gsk_...

# Optional: Additional providers
MISTRAL_API_KEY=...
COHERE_API_KEY=...
PERPLEXITY_API_KEY=...
GEMINI_API_KEY=...
```

### 3. Run HelixAgent

```bash
# Start with auto-container management
./bin/helixagent

# The service will be available at http://localhost:7061
```

## First API Call

Test the installation:

```bash
# Health check
curl http://localhost:7061/health

# List available models
curl http://localhost:7061/v1/models

# Simple completion
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Examples

### Basic Chat

```bash
cd examples
go run basic_chat.go
```

### Streaming

```bash
go run streaming.go
```

### Tool Calling

```bash
go run tool_calling.go
```

## Running Tests

```bash
# Unit tests only (fast)
make test-unit

# Integration tests (requires API keys)
make test-integration

# All tests
make test

# Test specific provider
export OPENAI_API_KEY=sk-...
go test -v ./tests/providers/openai_test.go
```

## Next Steps

- [API Reference](API_REFERENCE.md)
- [Provider Configuration](PROVIDERS.md)
- [Examples](../examples/)
- [Architecture](../ARCHITECTURE.md)
