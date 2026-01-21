# Chapter 3: Provider Configuration

Learn how to configure and manage LLM providers in HelixAgent.

## Supported Providers

HelixAgent supports 18+ LLM providers:

| Provider | API Key Variable | Models |
|----------|------------------|--------|
| Anthropic Claude | `CLAUDE_API_KEY` | claude-3.5-sonnet, claude-3-opus |
| DeepSeek | `DEEPSEEK_API_KEY` | deepseek-chat, deepseek-coder |
| Google Gemini | `GEMINI_API_KEY` | gemini-2.0-flash, gemini-pro |
| Alibaba Qwen | `QWEN_API_KEY` | qwen-max, qwen-plus, qwen-turbo |
| OpenRouter | `OPENROUTER_API_KEY` | Multiple models via routing |
| Mistral | `MISTRAL_API_KEY` | mistral-large, mixtral-8x7b |
| Cerebras | `CEREBRAS_API_KEY` | llama3.1-70b |
| Groq | `GROQ_API_KEY` | llama-3.1-70b-versatile |
| ZAI | `ZAI_API_KEY` | Various models |

## Configuration Methods

### Environment Variables

The simplest method is using environment variables:

```bash
# Required: At least one provider
export CLAUDE_API_KEY="your-claude-key"
export DEEPSEEK_API_KEY="your-deepseek-key"
export GEMINI_API_KEY="your-gemini-key"

# Optional: Additional providers
export QWEN_API_KEY="your-qwen-key"
export OPENROUTER_API_KEY="your-openrouter-key"
export MISTRAL_API_KEY="your-mistral-key"
```

### Configuration File

Create `configs/providers.yaml`:

```yaml
providers:
  claude:
    enabled: true
    api_key: ${CLAUDE_API_KEY}
    models:
      - claude-3.5-sonnet-20241022
      - claude-3-opus-20240229
    weight: 1.0
    timeout: 30s

  deepseek:
    enabled: true
    api_key: ${DEEPSEEK_API_KEY}
    models:
      - deepseek-chat
      - deepseek-coder
    weight: 0.8
    timeout: 30s

  gemini:
    enabled: true
    api_key: ${GEMINI_API_KEY}
    models:
      - gemini-2.0-flash
      - gemini-pro
    weight: 0.9
    timeout: 30s
```

### Docker Compose

In `docker-compose.yml`:

```yaml
services:
  helixagent:
    environment:
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
```

## Provider Verification

HelixAgent uses LLMsVerifier to validate providers:

### Automatic Verification

Providers are verified on startup:

```bash
# View verification status
curl http://localhost:7061/v1/providers/status
```

Response:
```json
{
  "providers": [
    {
      "name": "claude",
      "status": "verified",
      "score": 9.5,
      "models": ["claude-3.5-sonnet"]
    },
    {
      "name": "deepseek",
      "status": "verified",
      "score": 8.5,
      "models": ["deepseek-chat"]
    }
  ]
}
```

### Manual Verification

```bash
# Verify a specific provider
curl -X POST http://localhost:7061/v1/providers/verify \
  -d '{"provider": "claude"}'
```

## Provider Selection

### Dynamic Selection

HelixAgent automatically selects providers based on:
1. Verification status
2. Health status
3. LLMsVerifier scores
4. Response time

### Weighted Selection

Configure provider weights:

```yaml
providers:
  claude:
    weight: 1.0  # Highest priority

  deepseek:
    weight: 0.8  # Second priority

  gemini:
    weight: 0.6  # Third priority
```

### Fallback Chain

When a provider fails, HelixAgent automatically falls back:

```
Claude (primary) -> DeepSeek -> Gemini -> Qwen
```

## AI Debate Team

The AI Debate Ensemble uses 5 positions with fallbacks:

### Team Configuration

```yaml
debate:
  positions:
    analyst:
      primary: claude-3.5-sonnet
      fallbacks:
        - deepseek-chat
        - gemini-pro

    proposer:
      primary: deepseek-chat
      fallbacks:
        - claude-3.5-sonnet
        - qwen-max

    critic:
      primary: gemini-pro
      fallbacks:
        - claude-3.5-sonnet
        - deepseek-chat

    synthesizer:
      primary: qwen-max
      fallbacks:
        - claude-3.5-sonnet
        - deepseek-chat

    mediator:
      primary: claude-3-opus
      fallbacks:
        - claude-3.5-sonnet
        - gemini-pro
```

### Dynamic Team Selection

Teams are dynamically formed based on:
1. OAuth2 providers verified first (Claude, Qwen)
2. LLMsVerifier scored providers fill remaining slots
3. Same LLM can fill multiple positions if needed

## Provider Health Monitoring

### Health Endpoints

```bash
# Get all provider health
curl http://localhost:7061/v1/providers/health

# Get specific provider health
curl http://localhost:7061/v1/providers/claude/health
```

### Health Metrics

HelixAgent tracks:
- Response time
- Error rate
- Token usage
- Cost per request

### Prometheus Metrics

```
helixagent_provider_requests_total{provider="claude"} 1000
helixagent_provider_errors_total{provider="claude"} 5
helixagent_provider_latency_seconds{provider="claude"} 0.5
```

## Cost Management

### Cost Tracking

```bash
curl http://localhost:7061/v1/usage
```

Response:
```json
{
  "period": "current_month",
  "providers": {
    "claude": {
      "requests": 1000,
      "tokens_input": 500000,
      "tokens_output": 200000,
      "cost_usd": 15.50
    }
  }
}
```

### Cost Limits

Set cost limits in configuration:

```yaml
cost_management:
  monthly_limit_usd: 100.00
  daily_limit_usd: 10.00
  alert_threshold: 0.8  # Alert at 80%
```

## Cloud Provider Integration

### AWS Bedrock

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
```

### GCP Vertex AI

```bash
export GCP_PROJECT_ID="your-project"
export GCP_LOCATION="us-central1"
export GOOGLE_ACCESS_TOKEN="your-token"
```

### Azure OpenAI

```bash
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="your-key"
export AZURE_OPENAI_API_VERSION="2024-02-01"
```

## Timeout Configuration

HelixAgent provides comprehensive timeout configuration at multiple levels.

### Provider Timeouts

Configure timeouts per provider in `configs/development.yaml` or `configs/production.yaml`:

```yaml
providers:
  claude:
    timeout: "120s"        # Development: generous timeout
  deepseek:
    timeout: "120s"
  gemini:
    timeout: "120s"
  qwen:
    timeout: "120s"
  openrouter:
    timeout: "120s"
```

**Recommended timeout values:**

| Environment | Provider Timeout | Notes |
|-------------|------------------|-------|
| Development | 120s | Generous for debugging |
| Production | 60s | Balance between reliability and speed |
| Testing | 30s | Quick failure detection |

### Server Timeouts

Configure HTTP server timeouts:

```yaml
server:
  graceful_shutdown_timeout: "15s"   # Time to finish in-flight requests
  request_timeout: "180s"            # Overall request timeout
  idle_timeout: "300s"               # Keep-alive connection timeout
  read_timeout: "180s"               # Time to read full request
  write_timeout: "180s"              # Time to write response
```

### AI Debate Timeouts

Configure debate-specific timeouts in `configs/ai-debate-example.yaml`:

```yaml
debate_timeout: 300000              # 5 minutes total debate timeout (ms)

positions:
  analyst:
    response_timeout: 30000         # 30 seconds per position response (ms)
    providers:
      - name: claude
        timeout: 30000              # Individual provider timeout (ms)
        request_timeout: 35000      # Request with overhead (ms)
```

### Ensemble Timeouts

Configure ensemble decision timeouts:

```yaml
ensemble:
  timeout: 60s                      # Maximum time for ensemble decision
```

### Verification Timeouts

Configure LLMsVerifier timeouts:

```yaml
verifier:
  verification_timeout: 60s         # Time for provider verification
  circuit_breaker:
    timeout: 10s                    # Circuit breaker timeout
    half_open_timeout: 60s          # Time before retry after failure
```

### Challenge/Test Timeouts

For challenge scripts, timeouts are set via environment variables:

```bash
# Set timeout for RAGS challenge (default: 60s)
TIMEOUT=60 ./challenges/scripts/rags_challenge.sh

# Set timeout for MCPS challenge (default: 30s)
TIMEOUT=30 ./challenges/scripts/mcps_challenge.sh
```

**Important**: The RAGS challenge timeout was increased from 30s to 60s to accommodate RAG pipeline initialization delays.

### Timeout Best Practices

1. **Development**: Use generous timeouts (120s) to avoid false failures during debugging
2. **Production**: Use moderate timeouts (30-60s) with circuit breakers for resilience
3. **Testing**: Use shorter timeouts (5-30s) for quick feedback
4. **AI Debate**: Allow 30s per position, 5 minutes total debate time
5. **RAG Operations**: Allow 60s+ for complex retrieval operations

## Troubleshooting

### Common Issues

1. **Provider Not Found**
   - Check API key is set correctly
   - Verify provider is enabled in config

2. **Authentication Failed**
   - Regenerate API key
   - Check key permissions

3. **Rate Limited**
   - Implement backoff
   - Use multiple providers

4. **Timeout**
   - Increase timeout setting in config
   - Check network connectivity
   - For RAG operations, use 60s+ timeout
   - For AI Debate, ensure debate_timeout is sufficient

5. **Circuit Breaker Open**
   - Provider has failed too many times
   - Wait for half_open_timeout before retry
   - Check provider health: `curl http://localhost:7061/v1/providers/health`

### Debug Mode

```bash
# Enable provider debug logging
export HELIXAGENT_DEBUG=true
./bin/helixagent
```

### Timeout Debugging

```bash
# Check provider response times
curl http://localhost:7061/v1/providers/health

# Test specific provider with timeout
curl --max-time 30 "http://localhost:7061/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{"model": "helixagent-debate", "messages": [{"role": "user", "content": "test"}]}'
```
