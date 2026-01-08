# Provider Configuration Guide

## Introduction

HelixAgent integrates with 7 major LLM providers, allowing you to leverage the strengths of each model while maintaining a unified API interface. This guide provides comprehensive configuration instructions for each provider, including authentication, model selection, performance tuning, and best practices.

---

## Table of Contents

1. [Provider Overview](#provider-overview)
2. [Claude (Anthropic)](#claude-anthropic)
3. [DeepSeek](#deepseek)
4. [Gemini (Google)](#gemini-google)
5. [Qwen (Alibaba Cloud)](#qwen-alibaba-cloud)
6. [ZAI](#zai)
7. [Ollama](#ollama)
8. [OpenRouter](#openrouter)
9. [Multi-Provider Configuration](#multi-provider-configuration)
10. [Provider Health Monitoring](#provider-health-monitoring)
11. [Fallback Strategies](#fallback-strategies)
12. [Performance Optimization](#performance-optimization)

---

## Provider Overview

### Supported Providers Comparison

| Provider | Type | Key Strengths | Rate Limits | Pricing Model |
|----------|------|---------------|-------------|---------------|
| Claude | Cloud API | Reasoning, safety, long context | 60 RPM | Per token |
| DeepSeek | Cloud API | Cost-effective, coding | 100 RPM | Per token |
| Gemini | Cloud API | Multimodal, reasoning | 60 RPM | Per token |
| Qwen | Cloud API | Multilingual, math | 100 RPM | Per token |
| ZAI | Cloud API | Fast inference | 120 RPM | Per token |
| Ollama | Local | Privacy, no limits | Unlimited | Free (self-hosted) |
| OpenRouter | Aggregator | Model variety | Varies | Per token |

### Provider Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                      Provider Registry                            │
├──────────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
│  │ Claude  │  │DeepSeek │  │ Gemini  │  │  Qwen   │  │   ZAI   │ │
│  │Provider │  │Provider │  │Provider │  │Provider │  │Provider │ │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘ │
│       │            │            │            │            │       │
│  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐ │
│  │ Ollama  │  │OpenRouter│  │Circuit  │  │ Health  │  │ Metrics │ │
│  │Provider │  │Provider │  │Breaker  │  │ Monitor │  │Collector│ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

---

## Claude (Anthropic)

### Overview

Claude is Anthropic's advanced AI assistant, known for exceptional reasoning capabilities, safety features, and support for extremely long context windows (up to 200K tokens).

### Configuration

#### Environment Variables

```bash
# Required
CLAUDE_API_KEY=sk-ant-api03-your-key-here

# Optional
CLAUDE_MODEL=claude-3-sonnet-20240229
CLAUDE_MAX_TOKENS=4096
CLAUDE_TEMPERATURE=0.7
CLAUDE_TIMEOUT=60s
```

#### YAML Configuration

```yaml
llm_providers:
  claude:
    enabled: true
    api_key: "${CLAUDE_API_KEY}"
    base_url: "https://api.anthropic.com"
    model: "claude-3-sonnet-20240229"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.9

    # Connection settings
    timeout: "60s"
    retry_attempts: 3
    retry_delay: "1s"

    # Rate limiting
    rate_limit:
      requests_per_minute: 60
      tokens_per_minute: 100000

    # Ensemble weight (for multi-provider setups)
    weight: 1.0
    priority: 1
```

### Available Models

| Model | Context Window | Best For |
|-------|----------------|----------|
| claude-3-opus-20240229 | 200K | Complex reasoning, research |
| claude-3-sonnet-20240229 | 200K | Balanced performance (recommended) |
| claude-3-haiku-20240307 | 200K | Fast responses, simple tasks |
| claude-3-5-sonnet-20241022 | 200K | Latest improvements |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "messages": [
      {"role": "user", "content": "Explain quantum entanglement."}
    ],
    "max_tokens": 1000
  }'
```

### Best Practices

1. **Use system prompts**: Claude responds well to detailed system instructions
2. **Leverage long context**: Great for document analysis and multi-turn conversations
3. **Handle rate limits**: Implement exponential backoff for retries
4. **Monitor usage**: Track token consumption for cost management

---

## DeepSeek

### Overview

DeepSeek offers cost-effective, high-quality models with excellent coding capabilities and strong reasoning performance.

### Configuration

#### Environment Variables

```bash
# Required
DEEPSEEK_API_KEY=sk-your-deepseek-key

# Optional
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_MAX_TOKENS=4096
DEEPSEEK_BASE_URL=https://api.deepseek.com
```

#### YAML Configuration

```yaml
llm_providers:
  deepseek:
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com"
    model: "deepseek-chat"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.95
    frequency_penalty: 0.0
    presence_penalty: 0.0

    # Connection settings
    timeout: "60s"
    retry_attempts: 3
    retry_delay: "1s"

    # Rate limiting
    rate_limit:
      requests_per_minute: 100
      tokens_per_minute: 200000

    # Ensemble configuration
    weight: 0.9
    priority: 2
```

### Available Models

| Model | Context Window | Best For |
|-------|----------------|----------|
| deepseek-chat | 32K | General conversation |
| deepseek-coder | 16K | Code generation |
| deepseek-reasoner | 32K | Complex reasoning |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-coder",
    "messages": [
      {
        "role": "system",
        "content": "You are an expert programmer. Write clean, efficient code."
      },
      {
        "role": "user",
        "content": "Write a Python function to find prime numbers using the Sieve of Eratosthenes."
      }
    ],
    "max_tokens": 1500
  }'
```

### Best Practices

1. **Code tasks**: Use `deepseek-coder` for programming tasks
2. **Temperature tuning**: Lower temperature (0.3-0.5) for code, higher (0.7-0.9) for creative tasks
3. **Cost monitoring**: DeepSeek offers competitive pricing; monitor to optimize
4. **Batch requests**: Group similar requests to improve throughput

---

## Gemini (Google)

### Overview

Google's Gemini models offer strong multimodal capabilities, excellent reasoning, and tight integration with Google Cloud services.

### Configuration

#### Environment Variables

```bash
# Required
GEMINI_API_KEY=your-gemini-api-key

# Optional
GEMINI_MODEL=gemini-pro
GEMINI_MAX_TOKENS=4096
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
```

#### YAML Configuration

```yaml
llm_providers:
  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    base_url: "https://generativelanguage.googleapis.com/v1beta"
    model: "gemini-pro"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.95
    top_k: 40

    # Safety settings
    safety_settings:
      harm_block_threshold: "BLOCK_MEDIUM_AND_ABOVE"
      categories:
        - HARM_CATEGORY_HARASSMENT
        - HARM_CATEGORY_HATE_SPEECH
        - HARM_CATEGORY_SEXUALLY_EXPLICIT
        - HARM_CATEGORY_DANGEROUS_CONTENT

    # Connection settings
    timeout: "60s"
    retry_attempts: 3
    retry_delay: "1s"

    # Rate limiting
    rate_limit:
      requests_per_minute: 60
      tokens_per_minute: 120000

    # Ensemble configuration
    weight: 0.85
    priority: 3
```

### Available Models

| Model | Context Window | Best For |
|-------|----------------|----------|
| gemini-pro | 32K | General text tasks |
| gemini-pro-vision | 16K | Image + text understanding |
| gemini-1.5-pro | 1M | Very long context tasks |
| gemini-1.5-flash | 1M | Fast responses, long context |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-pro",
    "messages": [
      {"role": "user", "content": "Compare and contrast neural networks and decision trees."}
    ],
    "max_tokens": 1000
  }'
```

### Best Practices

1. **Safety settings**: Configure appropriate safety thresholds for your use case
2. **Multimodal tasks**: Use gemini-pro-vision for image analysis
3. **Long context**: Leverage 1M context for extensive document analysis
4. **Regional availability**: Check model availability in your region

---

## Qwen (Alibaba Cloud)

### Overview

Qwen (Tongyi Qianwen) is Alibaba Cloud's large language model, offering strong multilingual support, especially for Chinese, and excellent mathematical reasoning.

### Configuration

#### Environment Variables

```bash
# Required
QWEN_API_KEY=your-qwen-api-key

# Optional
QWEN_MODEL=qwen-turbo
QWEN_MAX_TOKENS=4096
QWEN_BASE_URL=https://dashscope.aliyuncs.com/api/v1
```

#### YAML Configuration

```yaml
llm_providers:
  qwen:
    enabled: true
    api_key: "${QWEN_API_KEY}"
    base_url: "https://dashscope.aliyuncs.com/api/v1"
    model: "qwen-turbo"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.8
    top_k: 50
    repetition_penalty: 1.1

    # Connection settings
    timeout: "60s"
    retry_attempts: 3
    retry_delay: "1s"

    # Rate limiting
    rate_limit:
      requests_per_minute: 100
      tokens_per_minute: 150000

    # Ensemble configuration
    weight: 0.8
    priority: 4
```

### Available Models

| Model | Context Window | Best For |
|-------|----------------|----------|
| qwen-turbo | 8K | Fast, general tasks |
| qwen-plus | 32K | Balanced performance |
| qwen-max | 32K | Complex reasoning |
| qwen-math | 8K | Mathematical problems |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen-max",
    "messages": [
      {
        "role": "system",
        "content": "You are a multilingual assistant proficient in Chinese and English."
      },
      {
        "role": "user",
        "content": "Translate and explain this Chinese proverb: 学如逆水行舟，不进则退"
      }
    ],
    "max_tokens": 500
  }'
```

### Best Practices

1. **Multilingual tasks**: Excellent for Chinese-English translation and content
2. **Math problems**: Use qwen-math for mathematical reasoning
3. **Repetition control**: Adjust repetition_penalty to prevent loops
4. **Regional latency**: Consider latency if deploying outside Asia

---

## ZAI

### Overview

ZAI provides fast inference speeds with competitive model performance, optimized for production workloads requiring low latency.

### Configuration

#### Environment Variables

```bash
# Required
ZAI_API_KEY=your-zai-api-key

# Optional
ZAI_MODEL=zai-chat
ZAI_MAX_TOKENS=4096
ZAI_BASE_URL=https://api.zai.com/v1
```

#### YAML Configuration

```yaml
llm_providers:
  zai:
    enabled: true
    api_key: "${ZAI_API_KEY}"
    base_url: "https://api.zai.com/v1"
    model: "zai-chat"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.9

    # Connection settings
    timeout: "30s"
    retry_attempts: 3
    retry_delay: "500ms"

    # Rate limiting
    rate_limit:
      requests_per_minute: 120
      tokens_per_minute: 200000

    # Ensemble configuration
    weight: 0.75
    priority: 5
```

### Available Models

| Model | Context Window | Best For |
|-------|----------------|----------|
| zai-chat | 8K | General conversation |
| zai-fast | 4K | Ultra-low latency |
| zai-pro | 16K | Complex tasks |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "zai-fast",
    "messages": [
      {"role": "user", "content": "Quick summary of climate change effects."}
    ],
    "max_tokens": 200
  }'
```

### Best Practices

1. **Low latency needs**: Use zai-fast for real-time applications
2. **Timeout settings**: Shorter timeouts suitable due to fast inference
3. **Batch processing**: Efficient for high-volume workloads
4. **Fallback provider**: Good secondary provider for ensemble setups

---

## Ollama

### Overview

Ollama enables running open-source LLMs locally, providing complete privacy, no API costs, and unlimited usage. Ideal for development, testing, and privacy-sensitive deployments.

### Configuration

#### Environment Variables

```bash
# Required
OLLAMA_ENABLED=true
OLLAMA_BASE_URL=http://localhost:11434

# Optional
OLLAMA_MODEL=llama2
OLLAMA_NUM_GPU=1
OLLAMA_NUM_THREAD=8
```

#### YAML Configuration

```yaml
llm_providers:
  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    model: "llama2"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.9
    top_k: 40

    # Model-specific options
    options:
      num_gpu: 1
      num_thread: 8
      num_ctx: 4096
      repeat_penalty: 1.1

    # Connection settings
    timeout: "120s"  # Longer timeout for local inference
    retry_attempts: 2
    retry_delay: "2s"

    # Ensemble configuration
    weight: 0.5  # Lower weight due to smaller model capabilities
    priority: 10  # Lower priority, used as fallback
```

### Installing and Managing Models

```bash
# Pull a model
docker exec -it helixagent-ollama ollama pull llama2

# List installed models
docker exec -it helixagent-ollama ollama list

# Pull additional models
docker exec -it helixagent-ollama ollama pull codellama
docker exec -it helixagent-ollama ollama pull mistral
docker exec -it helixagent-ollama ollama pull llama2:70b
```

### Available Models

| Model | Size | Best For |
|-------|------|----------|
| llama2 | 7B | General tasks, lightweight |
| llama2:13b | 13B | Balanced performance |
| llama2:70b | 70B | High quality (requires GPU) |
| codellama | 7B-34B | Code generation |
| mistral | 7B | Fast, efficient |
| mixtral | 8x7B | MoE architecture |

### Usage Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "ollama:llama2",
    "messages": [
      {"role": "user", "content": "Write a haiku about programming."}
    ],
    "max_tokens": 100
  }'
```

### Best Practices

1. **GPU acceleration**: Enable GPU for faster inference
2. **Model selection**: Match model size to available resources
3. **Context length**: Adjust num_ctx based on needs and memory
4. **Development use**: Ideal for testing without API costs
5. **Privacy**: Use for sensitive data that cannot leave premises

---

## OpenRouter

### Overview

OpenRouter is a unified API gateway that provides access to 100+ models from various providers, offering flexibility and easy model switching without managing multiple API keys.

### Configuration

#### Environment Variables

```bash
# Required
OPENROUTER_API_KEY=sk-or-v1-your-key-here

# Optional
OPENROUTER_MODEL=openai/gpt-4
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_SITE_URL=https://your-site.com
OPENROUTER_APP_NAME=HelixAgent
```

#### YAML Configuration

```yaml
llm_providers:
  openrouter:
    enabled: true
    api_key: "${OPENROUTER_API_KEY}"
    base_url: "https://openrouter.ai/api/v1"
    model: "anthropic/claude-3-sonnet"

    # Request settings
    temperature: 0.7
    max_tokens: 4096
    top_p: 0.9

    # OpenRouter-specific headers
    http_headers:
      HTTP-Referer: "${OPENROUTER_SITE_URL}"
      X-Title: "${OPENROUTER_APP_NAME}"

    # Connection settings
    timeout: "60s"
    retry_attempts: 3
    retry_delay: "1s"

    # Rate limiting
    rate_limit:
      requests_per_minute: 100
      tokens_per_minute: 200000

    # Ensemble configuration
    weight: 0.7
    priority: 6

    # Model routing preferences
    model_preferences:
      - "anthropic/claude-3-sonnet"
      - "openai/gpt-4-turbo"
      - "google/gemini-pro"
```

### Available Models (Sample)

| Model ID | Provider | Best For |
|----------|----------|----------|
| anthropic/claude-3-opus | Anthropic | Complex reasoning |
| anthropic/claude-3-sonnet | Anthropic | Balanced |
| openai/gpt-4-turbo | OpenAI | General tasks |
| openai/gpt-4 | OpenAI | Complex reasoning |
| google/gemini-pro | Google | General tasks |
| mistralai/mistral-large | Mistral | European deployment |
| meta-llama/llama-3-70b | Meta | Open source |

### Usage Example

```bash
# Access Claude via OpenRouter
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openrouter:anthropic/claude-3-sonnet",
    "messages": [
      {"role": "user", "content": "Explain the benefits of microservices architecture."}
    ],
    "max_tokens": 1000
  }'
```

### Best Practices

1. **Model diversity**: Access models not directly available
2. **Cost comparison**: Compare pricing across providers
3. **Fallback routing**: Use for automatic provider failover
4. **Site attribution**: Set HTTP-Referer for proper attribution
5. **Rate limit pooling**: OpenRouter aggregates rate limits

---

## Multi-Provider Configuration

### Ensemble Configuration

Configure multiple providers to work together:

```yaml
# configs/multi-provider.yaml
ensemble:
  enabled: true
  strategy: "confidence_weighted"  # or "majority_vote", "first_response", "best_quality"

  # Provider weights
  provider_weights:
    claude: 1.0
    deepseek: 0.9
    gemini: 0.85
    qwen: 0.8
    zai: 0.75
    ollama: 0.5
    openrouter: 0.7

  # Voting configuration
  voting:
    minimum_providers: 2
    maximum_providers: 5
    timeout: "30s"
    confidence_threshold: 0.7

  # Fallback configuration
  fallback:
    enabled: true
    max_retries: 3
    providers_order:
      - claude
      - deepseek
      - gemini
      - ollama
```

### Load Balancing

```yaml
load_balancing:
  enabled: true
  strategy: "weighted_round_robin"  # or "least_connections", "random"

  health_check:
    interval: "30s"
    timeout: "5s"
    unhealthy_threshold: 3
    healthy_threshold: 2

  circuit_breaker:
    enabled: true
    failure_threshold: 5
    success_threshold: 3
    timeout: "60s"
    half_open_requests: 3
```

---

## Provider Health Monitoring

### Health Check Endpoints

```bash
# Check all providers
curl http://localhost:8080/v1/providers

# Check specific provider
curl http://localhost:8080/v1/providers/deepseek/health

# Get provider metrics
curl http://localhost:8080/v1/providers/claude/metrics
```

### Health Check Response

```json
{
  "provider": "deepseek",
  "status": "healthy",
  "latency_ms": 245,
  "last_check": "2024-01-15T10:30:00Z",
  "uptime_percentage": 99.9,
  "requests_today": 1523,
  "errors_today": 2,
  "rate_limit_remaining": 85
}
```

### Monitoring Configuration

```yaml
monitoring:
  health_check:
    enabled: true
    interval: "30s"
    timeout: "10s"

  metrics:
    enabled: true
    prometheus:
      enabled: true
      endpoint: "/metrics"

  alerts:
    enabled: true
    channels:
      - type: "slack"
        webhook: "${SLACK_WEBHOOK_URL}"
      - type: "email"
        recipients:
          - "ops@company.com"

    thresholds:
      error_rate: 5  # Alert if > 5% errors
      latency_p99: 5000  # Alert if p99 > 5 seconds
      availability: 99  # Alert if < 99% uptime
```

---

## Fallback Strategies

### Configuration

```yaml
fallback:
  enabled: true
  strategies:

    # Primary → Secondary fallback
    primary_secondary:
      primary: "claude"
      secondary: "deepseek"
      conditions:
        - error_code: "rate_limit_exceeded"
        - error_code: "provider_unavailable"
        - latency_threshold: "10s"

    # Round-robin fallback
    round_robin:
      providers:
        - claude
        - deepseek
        - gemini
      on_failure: "next"

    # Quality-based fallback
    quality_based:
      preferred: "claude"
      fallback_order:
        - name: "deepseek"
          min_confidence: 0.8
        - name: "gemini"
          min_confidence: 0.7
        - name: "ollama"
          min_confidence: 0.5
```

### Usage

```bash
# Request with fallback
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-ensemble",
    "messages": [{"role": "user", "content": "Hello"}],
    "options": {
      "fallback_enabled": true,
      "fallback_providers": ["claude", "deepseek", "ollama"]
    }
  }'
```

---

## Performance Optimization

### Connection Pooling

```yaml
connection_pool:
  max_connections_per_host: 100
  max_idle_connections: 50
  idle_connection_timeout: "90s"

  # Per-provider settings
  providers:
    claude:
      max_connections: 50
    deepseek:
      max_connections: 80
    ollama:
      max_connections: 20
```

### Caching

```yaml
caching:
  enabled: true

  # Response caching
  response_cache:
    enabled: true
    ttl: "1h"
    max_entries: 10000

  # Semantic caching
  semantic_cache:
    enabled: true
    similarity_threshold: 0.92
    embedding_model: "text-embedding-3-small"
```

### Request Optimization

```yaml
optimization:
  # Batch requests
  batching:
    enabled: true
    max_batch_size: 10
    max_wait_time: "100ms"

  # Request deduplication
  deduplication:
    enabled: true
    window: "5s"

  # Streaming optimization
  streaming:
    buffer_size: 4096
    flush_interval: "50ms"
```

---

## Troubleshooting Provider Issues

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `authentication_failed` | Invalid API key | Verify API key is correct and active |
| `rate_limit_exceeded` | Too many requests | Implement backoff, increase limits |
| `model_not_found` | Invalid model name | Check model availability |
| `timeout` | Slow response | Increase timeout, check provider status |
| `context_length_exceeded` | Too many tokens | Reduce input length |

### Diagnostic Commands

```bash
# Test provider connection
curl http://localhost:8080/v1/providers/deepseek/test

# View provider logs
docker-compose logs helixagent | grep "deepseek"

# Check rate limit status
curl http://localhost:8080/v1/providers/claude/rate-limit
```

---

## Summary

This guide covered the configuration of all 7 LLM providers supported by HelixAgent. Key takeaways:

1. **Choose providers based on strengths**: Claude for reasoning, DeepSeek for code, Gemini for multimodal
2. **Configure fallbacks**: Ensure high availability with proper fallback chains
3. **Monitor health**: Set up alerts for provider issues
4. **Optimize performance**: Use caching, pooling, and batching
5. **Balance cost and quality**: Mix providers based on use case requirements

Continue to the [AI Debate System Guide](03-ai-debate-system.md) to learn about HelixAgent's unique multi-AI debate capabilities.
